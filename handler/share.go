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
	id := chi.URLParam(r, "id")

	user, password, _ := r.BasicAuth()

	h.lock.Lock()
	defer h.lock.Unlock()
	session, ok := h.sessions[id]
	if !ok {
		api.WriteBadRequest(rw, fmt.Sprintf("session with id %s does not exist", id))
		return
	}

	if session.Username != user || session.Password != password {
		rw.Header().Add("WWW-Authenticate", `Basic realm="Plexus Session", charset="UTF-8"`)
		api.WriteError(rw, http.StatusUnauthorized, fmt.Sprintf("invalid username / password"))
		return
	}

	rw.Header().Add("content-type", "text/html")
	rw.WriteHeader(http.StatusOK)
	asset.ShareTemplate.Execute(rw, map[string]string{
		"ID": session.ID,
	})
}

func (h *Handler) ShareSessionURL(rw http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	user, password, _ := r.BasicAuth()

	h.lock.Lock()
	defer h.lock.Unlock()
	session, ok := h.sessions[id]
	if !ok {
		api.WriteBadRequest(rw, fmt.Sprintf("session with id %s does not exist", id))
		return
	}

	if session.Username != user || session.Password != password {
		rw.Header().Add("WWW-Authenticate", `Basic realm="Plexus Session", charset="UTF-8"`)
		api.WriteError(rw, http.StatusUnauthorized, fmt.Sprintf("invalid username / password"))
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
