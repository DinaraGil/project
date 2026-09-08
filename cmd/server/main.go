package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"projectOzonBank/internal/api"
	"projectOzonBank/internal/app"
	"projectOzonBank/internal/domain"
	"projectOzonBank/internal/storage/memory"
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

	flag.Parse()

	var storage domain.Storage

	switch *storageType {
	case "memory":
		storage = memory.New()

	case "postgres":
		log.Fatal("postgres storage not implemented yet")

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
			log.Fatalf("server failed: %v", err)
		}
	}()

	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	<-ctx.Done()

	log.Println("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}

	log.Println("server stopped")
}
