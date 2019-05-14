package db

import (
	"fmt"
	"log"
	"math"
	"strconv"
	"time"

	"github.com/gener8ads/lib-go/pkg/env"
	"github.com/jinzhu/gorm"

	// postgres driver for gorm
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

// Connect returns an open database connection
func Connect() *gorm.DB {
	host := env.Get("DB_HOST", "localhost")
	name := env.Get("DB_NAME", "postgres")
	user := env.Get("DB_USER", "postgres")
	pass := env.Get("DB_PASS", "postgres")
	port := env.Get("DB_PORT", "5432")

	config := fmt.Sprintf("host=%s dbname=%s user=%s password=%s port=%s sslmode=disable", host, name, user, pass, port)

	var db *gorm.DB
	var err error

	attempts := 0
	maxRetries := 3

	for attempts <= maxRetries {
		db, err = gorm.Open("postgres", config)

		if err != nil {
			attempts++
			backoff := time.Duration(math.Pow(2, float64(attempts)))
			log.Printf("DB connecton error, retying in %ds...", backoff)
			time.Sleep(backoff * time.Second)
		} else {
			break
		}
	}

	if err != nil {
		log.Fatalf("Unable to establish DB connection:\n%s", err.Error())
	}

	logMode, _ := strconv.ParseBool(env.Get("DB_QUERY_LOG", "false"))
	db.LogMode(logMode)

	return db
}
