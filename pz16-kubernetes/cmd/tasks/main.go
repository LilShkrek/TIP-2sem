package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"pz16-kubernetes/internal/config"
	"pz16-kubernetes/internal/httpserver"
)

func main() {
	cfg := config.Load()
	logger := log.New(os.Stdout, "tasks: ", log.LstdFlags)

	server := &http.Server{
		Addr:              cfg.Address(),
		Handler:           httpserver.NewHandler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Printf("starting service on %s", cfg.Address())
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-errCh:
		logger.Printf("server error: %v", err)
		os.Exit(1)
	case <-ctx.Done():
		logger.Println("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Printf("graceful shutdown failed: %v", err)
		os.Exit(1)
	}

	logger.Println("service stopped")
}
