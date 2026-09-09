package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"projectOzonBank/internal/api"
	"projectOzonBank/internal/app"
	"projectOzonBank/internal/domain"
	"projectOzonBank/internal/storage/memory"
	"projectOzonBank/internal/storage/postgres"
)

func main() {
	storageType := flag.String(
		"storage",
		"memory",
		"storage backend: memory or postgres",
	)

	addr := flag.String(
		"addr",
		":8080",
		"http server address",
	)

	dsn := flag.String(
		"dsn",
		os.Getenv("DATABASE_URL"),
		"PostgreSQL connection string",
	)

	flag.Parse()

	ctx := context.Background()

	var storage domain.Storage

	switch *storageType {
	case "memory":
		storage = memory.New()

	case "postgres":
		if *dsn == "" {
			log.Fatal("DATABASE_URL or -dsn is required for postgres storage")
		}

		pgStorage, err := postgres.New(ctx, *dsn)
		if err != nil {
			log.Fatalf("failed to create postgres storage: %v", err)
		}

		storage = pgStorage

		defer pgStorage.Close()

	default:
		log.Fatalf("unknown storage type: %s", *storageType)
	}

	service := app.New(storage)
	handler := api.NewHandler(service)
	router := api.NewRouter(handler)

	server := &http.Server{
		Addr:         *addr,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("server started on %s", *addr)

		if err := server.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {
			log.Printf("server failed: %v", err)
		}
	}()

	shutdownCtx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	<-shutdownCtx.Done()

	log.Println("shutdown signal received")

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}

	log.Println("server stopped")
}
