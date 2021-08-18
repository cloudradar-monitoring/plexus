package router

import (
	"net/http"

	"github.com/cloudradar-monitoring/plexus/agent"
	"github.com/cloudradar-monitoring/plexus/config"
	"github.com/go-chi/chi/v5"
)

func New(cfg *config.Server) http.Handler {
	r := chi.NewRouter()
	r.Get("/agent", agent.Handler(cfg))
	return r
}
