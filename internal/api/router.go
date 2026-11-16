package api

import (
	"net/http"
	"time"

	"domain-scout/internal/api/handler"
	"domain-scout/internal/config"
	"domain-scout/internal/discovery"
	"domain-scout/internal/server"
	"domain-scout/pkg/logger"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
)

func NewRouter(cfg config.Config, ds *discovery.Service, ss *server.Service, logr *logger.Logger) http.Handler {
	r := chi.NewRouter()
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Recoverer)
	r.Use(chimw.Timeout(time.Duration(cfg.SSHTimeoutSeconds*4) * time.Second))
	r.Use(chimw.RequestLogger(&chimw.DefaultLogFormatter{Logger: logr.Std(), NoColor: true}))
	r.Use(chimw.SetHeader("Content-Type", "application/json"))

	r.Get("/api/v1/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok","service":"domain-scout"}`))
	})
	r.Route("/api/v1", func(api chi.Router) {
		api.Post("/servers/test-connection", handler.NewServerHandler(ss).TestConnection)
		api.Post("/discoveries", handler.NewDiscoveryHandler(ds).Discover)
	})
	return r
}
