package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"cloudprep-backend/internal/config"
	"cloudprep-backend/internal/db"
	"cloudprep-backend/internal/router"
)

func main() {
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("[main] security configuration validation failed: %v", err)
	}

	database, dbDriver, err := db.ConnectAndMigrate(cfg)
	if err != nil {
		log.Fatalf("[main] database initialization failed: %v", err)
	}

	r := router.SetupRouter(cfg, database, dbDriver)

	addr := ":" + cfg.Port
	srv := &http.Server{
		Addr:              addr,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1 MB
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("[main] CloudPrep Golang/Gin hardened API listening on %s (db driver: %s)", addr, dbDriver)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("[main] server stopped with error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("[main] shutting down server gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("[main] forced server shutdown: %v", err)
	}
}
