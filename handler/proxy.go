package handler

import (
	"fmt"
	"net/http"

	"github.com/cloudradar-monitoring/plexus/proxy"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

func (h *Handler) ProxyRelay(rw http.ResponseWriter, r *http.Request) {
	log.Info().Interface("headers", r.Header).Msg("Proxy reley")

	proxyURL := fmt.Sprintf("%s?%s", h.cfg.MeshRelayURL(), r.URL.RawQuery)
	_, ok := proxy.Proxy(rw, r, proxyURL, h.cfg.MeshCentralInsecure)
	if ok {
		log.Debug().Str("url", r.URL.String()).Msg("MeshRelay Connected")
	}
}

func (h *Handler) ProxyAgent(rw http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	token := chi.URLParam(r, "token")

	h.lock.Lock()
	defer h.lock.Unlock()
	session, ok := h.sessions[id]
	if !ok || session.Token != token {
		proxy.Hold(rw, r)
		return
	}

	agentClose, ok := proxy.Proxy(rw, r, h.cfg.MeshCentralAgentURL(), h.cfg.MeshCentralInsecure)
	if ok {
		session.ProxyClose = agentClose
	}
}
