package config

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/XSAM/otelsql"
	_ "github.com/lib/pq"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

var DB *sql.DB

func DatabaseConnection(ctx context.Context) {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	driverName, err := otelsql.Register("postgres", otelsql.WithAttributes(semconv.DBSystemPostgreSQL))
	if err != nil {
		log.Fatalf("Register otelsql driver failed: %v", err)
	}

	DB, err = sql.Open(driverName, dsn)
	if err != nil {
		log.Fatalf("DB connection error: %v", err)
	}

	if err = DB.PingContext(ctx); err != nil {
		log.Fatalf("Ping failed: %v", err)
	}

	log.Println("Connected to DB")
}