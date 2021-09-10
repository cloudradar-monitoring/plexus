package router

import (
	"net/http"
	"time"

	"github.com/cloudradar-monitoring/plexus/handler"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"
)

func New(h *handler.Handler) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(accessLog)
	r.Post("/session", h.CreateSession)
	r.Get("/session/{id}", h.ShareSession)
	r.Get("/check/{id}", h.ShareSessionCheck)
	r.Get("/config/{id}:{token}", h.GetAgentMsh)
	r.Get("/agent/{id}:{token}", h.ProxyAgent)
	r.Get("/meshrelay.ashx", h.ProxyRelay)
	return r
}
func accessLog(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		start := time.Now()
		h.ServeHTTP(rw, r)
		duration := time.Now().Sub(start)

		log.Info().
			Str("host", r.Host).
			Str("ip", r.RemoteAddr).
			Str("path", r.URL.Path).
			Str("duration", duration.String()).
			Msg("HTTP")
	})
}
