package pubsub

import (
	"context"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"cloud.google.com/go/pubsub"
)

const (
	ErrorKey  = "pubsub.error"
	ResultKey = "pubsub.result"
)

type options struct {
	messageFunc MessageProducer
}

// HandlerFunc is a pubsub handler function which can be used in pubsub subscription receive.
type HandlerFunc func(ctx context.Context, m *pubsub.Message)

func defaultOptions() *options {
	return &options{
		messageFunc: defaultMessageProducer,
	}
}

// Option describes a logging option.
type Option func(*options)

// WithMessageProducer overrides the default message producer for producing log messages.
func WithMessageProducer(messageFunc MessageProducer) Option {
	return func(o *options) {
		o.messageFunc = messageFunc
	}
}

// MessageProducer produces log messages.
type MessageProducer func(ctx context.Context, msg string, logger *zap.Logger, fields []zapcore.Field)

// defaultMessageProducer writes default log messages.
func defaultMessageProducer(ctx context.Context, msg string, logger *zap.Logger, fields []zapcore.Field) {
	logger.Info(msg, fields...)
}

// NewLoggingHandler returns a new logging handler middleware which uses Zap as the logger.
func NewLoggingHandler(next HandlerFunc, subscriptionName string, opts ...Option) HandlerFunc {
	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}

	var logger *zap.Logger
	// First, define our level-handling logic.
	highPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.ErrorLevel
	})
	lowPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl < zapcore.ErrorLevel
	})

	config := zapcore.EncoderConfig{
		MessageKey:     string(ResultKey),
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.MillisDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// High-priority output should also go to standard error, and low-priority
	// output should also go to standard out.
	jsonDebugging := zapcore.Lock(os.Stdout)
	jsonErrors := zapcore.Lock(os.Stderr)
	jsonEncoder := zapcore.NewJSONEncoder(config)

	// Join the outputs, encoders, and level-handling functions into
	// zapcore.Cores, then tee the four cores together.
	core := zapcore.NewTee(
		zapcore.NewCore(jsonEncoder, jsonErrors, highPriority),
		zapcore.NewCore(jsonEncoder, jsonDebugging, lowPriority),
	)

	logger = zap.New(core)

	return func(ctx context.Context, msg *pubsub.Message) {
		defer logger.Sync()
		startTime := time.Now()

		if msg.Attributes == nil {
			msg.Attributes = make(map[string]string, 0)
		}

		next(ctx, msg)

		fields := []zapcore.Field{
			zap.String("pubsub.subscription", subscriptionName),
			zap.String("pubsub.msg.id", msg.ID),
			zap.Time("pubsub.msg.publishTime", msg.PublishTime),
			zap.Duration("pubsub.latency", time.Since(startTime)),
		}

		if msg.DeliveryAttempt != nil {
			fields = append(fields, zap.Int("pubsub.msg.deliveryAttempt", *msg.DeliveryAttempt))
		}

		if val, ok := msg.Attributes[string(ErrorKey)]; ok {
			fields = append(fields, zap.String("pubsub.error", val))
		}

		result := "not provided"
		if val, ok := msg.Attributes[string(ResultKey)]; ok {
			result = val
		}

		if o.messageFunc != nil {
			o.messageFunc(ctx, result, logger, fields)
		}
	}
}
