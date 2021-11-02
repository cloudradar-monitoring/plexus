package handler

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/gorilla/mux"

	"github.com/cloudradar-monitoring/plexus/proxy"
)

func (h *Handler) ProxyMeshCentralURL() string {
	return fmt.Sprintf("/%s/", h.ccfg.MeshCentralDomain)
}
func (h *Handler) ProxyMeshCentral() http.Handler {
	director := func(req *http.Request) {
		if strings.HasPrefix(req.URL.Scheme, "ws") {
			req.URL.Scheme = "ws"
		} else {
			req.URL.Scheme = "http"
		}
		req.URL.Host = h.ccfg.MeshCentralURL.Host
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
	proxyURL := fmt.Sprintf("%s?%s", h.ccfg.MeshRelayURL(), r.URL.RawQuery)
	_, ok := proxy.Proxy(h.log, rw, r, proxyURL)
	if ok {
		h.log.Debugf("MeshRelay Connected: %s", r.URL.String())
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
	vars := mux.Vars(r)
	id := vars["id"]
	token := vars["token"]

	h.lock.Lock()
	defer h.lock.Unlock()
	session, ok := h.sessions[id]
	if !ok || session.Token != token {
		proxy.Hold(h.log, rw, r)
		return
	}

	agentClose, ok := proxy.Proxy(h.log, rw, r, h.ccfg.MeshCentralAgentURL())
	if ok {
		session.ProxyClose = agentClose
	}
}
