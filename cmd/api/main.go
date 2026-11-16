package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"domain-scout/internal/api"
	"domain-scout/internal/config"
	"domain-scout/internal/discovery"
	"domain-scout/internal/inspectors"
	"domain-scout/internal/inspectors/apache"
	"domain-scout/internal/inspectors/caddy"
	"domain-scout/internal/inspectors/docker"
	"domain-scout/internal/inspectors/nginx"
	"domain-scout/internal/inspectors/system"
	"domain-scout/internal/server"
	"domain-scout/pkg/logger"
)

func main() {
	cfg := config.Load()
	logr := logger.New()
	registry := []inspectors.Inspector{
		nginx.New(), apache.New(), caddy.New(), docker.New(), system.New(),
	}
	discoveryService := discovery.NewService(registry)
	serverService := server.NewService()
	router := api.NewRouter(cfg, discoveryService, serverService, logr)

	srv := &http.Server{
		Addr:              cfg.HTTPAddress,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      2 * time.Minute,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Printf("DomainScout API listening on %s", cfg.HTTPAddress)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server: %v", err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP shutdown: %v", err)
	}
}
