package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"threat-service/api"
	"threat-service/internal/handler"
	"threat-service/internal/repository"
	"threat-service/internal/service"
	"threat-service/internal/validation"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return errors.New("DATABASE_URL environment variable is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	repo, err := repository.NewPostgresThreatRepository(ctx, databaseURL)
	cancel()

	if err != nil {
		return fmt.Errorf("failed to initialize PostgreSQL repository: %w", err)
	}
	defer repo.Close()

	svc := service.NewThreatService(repo)
	h := handler.NewThreatHandler(svc)

	router := api.Handler(h)

	spec, err := validation.LoadSpec("api/openapi.yaml")
	if err != nil {
		return fmt.Errorf("failed to load OpenAPI spec: %w", err)
	}

	middleware, err := validation.NewMiddleware(spec)
	if err != nil {
		return fmt.Errorf("failed to initialize validation middleware: %w", err)
	}

	server := &http.Server{
		Addr:              ":8080",
		Handler:           middleware(router),
		ReadHeaderTimeout: 5 * time.Second,
	}

	signalCtx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	serverErrors := make(chan error, 1)

	go func() {
		log.Println("Server starting on :8080")
		serverErrors <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("HTTP server failed: %w", err)
		}

	case <-signalCtx.Done():
		log.Println("Shutting down server...")
	}

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	log.Println("Server stopped")
	return nil
}
