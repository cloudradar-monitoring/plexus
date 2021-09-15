package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/cloudradar-monitoring/plexus/api"
	"github.com/cloudradar-monitoring/plexus/asset"
	"github.com/cloudradar-monitoring/plexus/control"
)

// ShareSession godoc
// @Summary Start the remote control on the session via a browser.
// @Tags session
// @Produce text/html
// @Param id path string true "session id"
// @Security BasicAuth
// @Success 200 {object} string
// @Failure 400 {object} api.Error
// @Failure 401 {object} api.Error
// @Failure 500 {object} api.Error
// @Router /session/{id} [get]
func (h *Handler) ShareSession(rw http.ResponseWriter, r *http.Request) {
	h.lock.Lock()
	defer h.lock.Unlock()

	session, ok := h.basicAuth(rw, r, chi.URLParam(r, "id"))
	if !ok {
		return
	}

	rw.Header().Add("content-type", "text/html")
	rw.WriteHeader(http.StatusOK)
	_ = asset.ShareTemplate.Execute(rw, map[string]string{
		"ID": session.ID,
	})
}

// ShareSessionURL godoc
// @Summary Gets the meshcentral share session for the session id.
// @Tags session
// @Produce application/json
// @Param id path string true "session id"
// @Security BasicAuth
// @Success 200 {object} api.URLResponse
// @Failure 400 {object} api.Error
// @Failure 401 {object} api.Error
// @Failure 500 {object} api.Error
// @Failure 502 {object} api.Error
// @Router /session/{id}/url [get]
func (h *Handler) ShareSessionURL(rw http.ResponseWriter, r *http.Request) {
	h.lock.Lock()
	defer h.lock.Unlock()

	session, ok := h.basicAuth(rw, r, chi.URLParam(r, "id"))
	if !ok {
		return
	}

	if session.ShareURL == "" {
		url, exit := h.tryGetURL(rw, r.Host, session)
		if exit {
			return
		}
		session.ShareURL = url
	}

	rw.Header().Add("content-type", "application/json")
	rw.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(rw).Encode(&api.URLResponse{
		URL: session.ShareURL,
	})
}

func (h *Handler) tryGetURL(rw http.ResponseWriter, host string, session *Session) (string, bool) {
	mc, err := control.Connect(h.cfg)
	if err != nil {
		api.WriteBadGatewayJSON(rw, fmt.Sprintf("could not connect: %s", err))
		return "", true
	}
	defer mc.Close()
	share, err := mc.Share(session.AgentConfig.MeshID, session.ID, session.ExpiresAt)
	if err != nil && err != control.ErrAgentNotConnected {
		api.WriteBadGatewayJSON(rw, fmt.Sprintf("could not create share: %s", err))
		return "", true
	}
	return "https://" + host + share, false
}
