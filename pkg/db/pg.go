package db

import (
	"context"
	"log"
	"math"
	"strconv"
	"time"

	"github.com/gener8ads/lib-go/pkg/env"
	"github.com/go-pg/pg/v9"
)

var conn *pg.DB
var isInit bool

type dbLogger struct{}

func (d dbLogger) BeforeQuery(c context.Context, q *pg.QueryEvent) (context.Context, error) {
	return c, nil
}

func (d dbLogger) AfterQuery(c context.Context, q *pg.QueryEvent) error {
	query, err := q.FormattedQuery()
	log.Println(query)

	if err != nil {
		log.Printf("ERROR: %s", err)
	}

	return nil
}

// Connection to a DB for go-pg
func Connection() *pg.DB {
	if isInit {
		return conn
	}

	var err error
	const maxRetries = 3
	attempts := 0

	for attempts <= maxRetries {
		conn, err = connect()

		if err != nil {
			attempts++
			backoff := time.Duration(math.Pow(2, float64(attempts)))
			log.Printf("DB connecton error: %s \nRetying in %ds…", err, backoff)
			time.Sleep(backoff * time.Second)
		} else {
			break
		}
	}

	if err != nil {
		log.Fatalf("Unable to establish DB connection:\n%s", err)
	}

	queryLogEnabled, _ := strconv.ParseBool(env.Get("DB_QUERY_LOG", "false"))
	if queryLogEnabled {
		conn.AddQueryHook(dbLogger{})
		go func() {
			for {
				time.Sleep(time.Second * 10)
				PoolStats()
			}
		}()
	}

	isInit = true
	return conn
}

// GetHandle for a single connection in the pool for go-pg
func GetHandle() *pg.Conn {
	if isInit {
		return conn.Conn()
	}
	return Connection().Conn()
}

// PoolStats print the Connection Pool stats to STDOUT
func PoolStats() {
	if isInit {
		stats := conn.PoolStats()
		log.Printf("Pool stats:\n\tHits: %d\n\tMisses: %d\n\tTimeouts: %d\n\tTotalConns: %d\n\tIdleConns: %d\n\tStaleConns: %v\n",
			stats.Hits, stats.Misses, stats.Timeouts, stats.TotalConns, stats.IdleConns, stats.StaleConns)
	}
}

func connect() (*pg.DB, error) {
	poolSize, err := strconv.Atoi(env.Get("DB_POOL_SIZE", "10"))
	if err != nil {
		poolSize = 10
	}
	minIdleConns, err := strconv.Atoi(env.Get("DB_MIN_IDLE", "5"))
	if err != nil {
		minIdleConns = 5
	}
	var maxConnAge time.Duration
	iMaxConnAge, err := strconv.Atoi(env.Get("DB_MAX_CONN_AGE", "5"))
	if err != nil {
		maxConnAge = 5 * time.Minute
	} else {
		maxConnAge = time.Duration(iMaxConnAge) * time.Minute
	}
	var idleTimeout time.Duration
	iIdleTimeout, err := strconv.Atoi(env.Get("DB_IDLE_TIMEOUT", "2"))
	if err != nil {
		idleTimeout = 2 * time.Minute
	} else {
		idleTimeout = time.Duration(iIdleTimeout) * time.Minute
	}

	db := pg.Connect(&pg.Options{
		Addr:         env.Get("DB_HOST", "localhost:5432"),
		Database:     env.Get("DB_NAME", "postgres"),
		User:         env.Get("DB_USER", "postgres"),
		Password:     env.Get("DB_PASS", "postgres"),
		PoolSize:     poolSize,
		MinIdleConns: minIdleConns,
		MaxConnAge:   maxConnAge,
		IdleTimeout:  idleTimeout,
	})

	var n int
	_, err = db.QueryOne(pg.Scan(&n), "SELECT 1")

	return db, err
}
