package router

import (
	"net/http"

	"github.com/gorilla/mux"

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

func Register(r *mux.Router, h *handler.Handler) {
	r.HandleFunc("/session", h.SessionCreationAuth(h.CreateSession)).Methods(http.MethodPost)
	r.HandleFunc("/session", h.SessionCreationAuth(h.ListSessions)).Methods(http.MethodGet)
	r.HandleFunc("/session/{id}", h.ShareSession).Methods(http.MethodGet)
	r.HandleFunc("/session/{id}/url", h.ShareSessionURL).Methods(http.MethodGet)
	r.HandleFunc("/session/{id}", h.DeleteSession).Methods(http.MethodDelete)
	r.HandleFunc("/config/{id}:{token}", h.GetAgentMsh).Methods(http.MethodGet)
	r.HandleFunc("/agent/{id}:{token}", h.ProxyAgent).Methods(http.MethodGet)
	r.HandleFunc("/meshrelay.ashx", h.ProxyRelay).Methods(http.MethodGet)
	r.PathPrefix(h.ProxyMeshCentralURL()).Handler(h.ProxyMeshCentral())
}
