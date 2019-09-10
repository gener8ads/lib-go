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
	}

	isInit = true
	return conn
}

func connect() (*pg.DB, error) {
	db := pg.Connect(&pg.Options{
		Addr:     env.Get("DB_HOST", "localhost:5432"),
		Database: env.Get("DB_NAME", "postgres"),
		User:     env.Get("DB_USER", "postgres"),
		Password: env.Get("DB_PASS", "postgres"),
	})

	var n int
	_, err := db.QueryOne(pg.Scan(&n), "SELECT 1")

	return db, err
}
