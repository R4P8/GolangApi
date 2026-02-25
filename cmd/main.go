package main

import (
	"ApiSimple/internal/config"
	"ApiSimple/internal/routes"
	"ApiSimple/pkg/logger"
	"ApiSimple/pkg/tracing"
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment")
	}

	ctx := context.Background()
	otlpEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")

	// Init Tracer
	shutdown := tracing.InitTracer(ctx, "ExampleApi", otlpEndpoint)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := shutdown(ctx); err != nil {
			log.Printf("Tracer shutdown error: %v", err)
		}
	}()

	// Init Logger
	shutdownLogger, err := logger.InitLogger(ctx, "ExampleApi", otlpEndpoint)
	if err != nil {
		log.Fatal("failed to init logger: ", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		shutdownLogger(ctx)
	}()

	// Database Connection
	config.DatabaseConnection(ctx)
	if config.DB == nil {
		log.Fatal("Database connection failed!")
	}
	var db *sql.DB = config.DB

	// Router
	router := mux.NewRouter()
	routes.RegisterRoutes(router, db)

	log.Println("Server running on port 8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal(err)
	}
}
