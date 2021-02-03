package grpc

import (
	"context"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	ctx_zap "github.com/grpc-ecosystem/go-grpc-middleware/tags/zap"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	grpc_datadogtrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/google.golang.org/grpc"
)

var (
	zapLogger *zap.Logger
)

// NewServer returns an instrumented gRPC server
func NewServer(serviceName string) *grpc.Server {
	zapLogger = zap.NewExample()

	// Shared options for the logger, with a custom gRPC code to log level function.
	opts := []grpc_zap.Option{
		grpc_zap.WithLevels(grpc_zap.DefaultCodeToLevel),
	}
	// Make sure that log statements internal to gRPC library are logged using the zapLogger as well.
	grpc_zap.ReplaceGrpcLogger(zapLogger)

	server := grpc.NewServer(
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			grpc_ctxtags.StreamServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
			grpc_opentracing.StreamServerInterceptor(),
			grpc_prometheus.StreamServerInterceptor,
			grpc_zap.StreamServerInterceptor(zapLogger, opts...),
			RequestIDLoggingStreamInterceptor(),
			grpc_recovery.StreamServerInterceptor(),
		)),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_ctxtags.UnaryServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
			grpc_opentracing.UnaryServerInterceptor(),
			grpc_prometheus.UnaryServerInterceptor,
			grpc_datadogtrace.UnaryServerInterceptor(grpc_datadogtrace.WithServiceName(serviceName)),
			grpc_zap.UnaryServerInterceptor(zapLogger, opts...),
			RequestIDLoggingUnaryInterceptor(),
			grpc_recovery.UnaryServerInterceptor(),
		)),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: 5 * time.Minute,
		}),
	)
	// Add healthz to this server based on the v1 standardised health contract
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(server, healthServer)

	reflection.Register(server)

	return server
}

// RequestIDLoggingUnaryInterceptor hauls the requestId (if available) out of the requests context & adds it to the zap.Logger's context.
func RequestIDLoggingUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		logRequestID(ctx)
		return handler(ctx, req)
	}
}

// RequestIDLoggingStreamInterceptor hauls the requestId (if available) out of the requests context & adds it to the zap.Logger's context.
func RequestIDLoggingStreamInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {

		ctx := stream.Context()
		logRequestID(ctx)
		wrapped := grpc_middleware.WrapServerStream(stream)
		wrapped.WrappedContext = ctx

		return handler(srv, wrapped)
	}
}

func logRequestID(ctx context.Context) {
	if requestID := getRequestID(ctx); requestID != "" {
		ctx_zap.AddFields(ctx, zap.String("requestId", requestID))
	}
}

func getRequestID(ctx context.Context) string {
	// Anything linked to this variable will fetch request headers.
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if xrid := md["x-request-id"]; len(xrid) > 0 && xrid[0] != "" {
			return xrid[0]
		}
	}
	return "not provided"
}
