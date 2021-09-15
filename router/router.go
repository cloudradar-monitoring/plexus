package router

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"

	"github.com/cloudradar-monitoring/plexus/handler"
)

// @title Plexus API
// @version 1.0
// @description Simple remote-control interface

// @contact.name CloudRadar
// @contact.url https://www.cloudradar.io/

// @license.name MIT
// @license.url https://github.com/cloudradar-monitoring/plexus/blob/main/LICENSE

// @host localhost:8080
// @BasePath /

// @securityDefinitions.basic BasicAuth
func New(h *handler.Handler) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(accessLog)

	r.Post("/session", h.CreateSession)
	r.Get("/session/{id}", h.ShareSession)
	r.Get("/session/{id}/url", h.ShareSessionURL)
	r.Delete("/session/{id}", h.DeleteSession)
	r.Get("/config/{id}:{token}", h.GetAgentMsh)
	r.Get("/agent/{id}:{token}", h.ProxyAgent)
	r.Get("/meshrelay.ashx", h.ProxyRelay)
	r.Handle(h.ProxyMeshCentralURL(), h.ProxyMeshCentral())
	return r
}
func accessLog(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		start := time.Now()
		h.ServeHTTP(rw, r)
		duration := time.Since(start)

		log.Info().
			Str("host", r.Host).
			Str("ip", r.RemoteAddr).
			Str("path", r.URL.Path).
			Str("duration", duration.String()).
			Msg("HTTP")
	})
}
