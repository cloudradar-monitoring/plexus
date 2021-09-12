package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/cloudradar-monitoring/plexus/api"
	"github.com/cloudradar-monitoring/plexus/asset"
	"github.com/cloudradar-monitoring/plexus/control"
	"github.com/go-chi/chi/v5"
)

func (h *Handler) ShareSession(rw http.ResponseWriter, r *http.Request) {
	h.lock.Lock()
	defer h.lock.Unlock()

	session, ok := h.basicAuth(rw, r, chi.URLParam(r, "id"))
	if !ok {
		return
	}

	rw.Header().Add("content-type", "text/html")
	rw.WriteHeader(http.StatusOK)
	asset.ShareTemplate.Execute(rw, map[string]string{
		"ID": session.ID,
	})
}

func (h *Handler) ShareSessionURL(rw http.ResponseWriter, r *http.Request) {
	h.lock.Lock()
	defer h.lock.Unlock()

	session, ok := h.basicAuth(rw, r, chi.URLParam(r, "id"))
	if !ok {
		return
	}

	if session.ShareURL == "" {
		url, exit := h.tryGetURL(rw, session)
		if exit {
			return
		}
		session.ShareURL = url
	}

	rw.Header().Add("content-type", "application/json")
	rw.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(rw).Encode(&map[string]interface{}{
		"url": session.ShareURL,
	})
}

func (h *Handler) tryGetURL(rw http.ResponseWriter, session *Session) (string, bool) {
	mc, err := control.Connect(h.cfg)
	defer mc.Close()
	share, err := mc.Share(session.AgentConfig.MeshID, session.ID, session.ExpiresAt)
	if err != nil && err != control.ErrAgentNotConnected {
		api.WriteBadGateway(rw, fmt.Sprintf("could not create share: %s", err))
		return "", true
	}
	return share, false
}
