package handler

import (
	"fmt"
	"net/http"

	"github.com/cloudradar-monitoring/plexus/api"
)

func (h *Handler) checkSessionAuthentication(rw http.ResponseWriter, r *http.Request, id string) (*Session, bool) {
	user, password, _ := r.BasicAuth()
	session, ok := h.sessions[id]
	if !ok {
		api.WriteBadRequestJSON(rw, fmt.Sprintf("session with id %s does not exist", id))
		return nil, false
	}

	if h.sessionCredentials {
		if session.Username == user && session.Password == password {
			return session, true
		}
	}

	return session, h.auth(rw, r)
}
