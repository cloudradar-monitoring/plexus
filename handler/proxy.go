package handler

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"

	"github.com/cloudradar-monitoring/plexus/proxy"
)

func (h *Handler) ProxyMeshCentralURL() string {
	return fmt.Sprintf("/%s/*", h.cfg.MeshCentralDomain)
}
func (h *Handler) ProxyMeshCentral() http.Handler {
	director := func(req *http.Request) {
		if strings.HasPrefix(req.URL.Scheme, "ws") {
			req.URL.Scheme = "ws"
		} else {
			req.URL.Scheme = "http"
		}
		req.URL.Host = h.cfg.MeshCentralURLParsed.Host
		req.Header.Set("User-Agent", "plexus")
	}
	return &httputil.ReverseProxy{
		Director: director,
	}
}

// ProxyRelay godoc
// @Summary Forwards meshagent relay requests to the meshcentral server.
// @Tags websocket
// @Success 200 {object} string
// @Router /meshrelay.ashx [get]
func (h *Handler) ProxyRelay(rw http.ResponseWriter, r *http.Request) {
	log.Info().Interface("headers", r.Header).Msg("Proxy Relay")

	proxyURL := fmt.Sprintf("%s?%s", h.cfg.MeshRelayURL(), r.URL.RawQuery)
	_, ok := proxy.Proxy(rw, r, proxyURL)
	if ok {
		log.Debug().Str("url", r.URL.String()).Msg("MeshRelay Connected")
	}
}

// ProxyAgent godoc
// @Summary Forwards the agent control requests to the meshcentral server
// @Tags websocket
// @Param id path string true "session id"
// @Param token path string true "the authentication token"
// @Success 200 {object} string
// @Router /agent/{id}:{token} [get]
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

	agentClose, ok := proxy.Proxy(rw, r, h.cfg.MeshCentralAgentURL())
	if ok {
		session.ProxyClose = agentClose
	}
}
