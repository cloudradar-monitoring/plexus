package handler

import (
	"fmt"
	"net/http"

	"github.com/cloudradar-monitoring/plexus/api"
)

func (h *Handler) basicAuth(rw http.ResponseWriter, r *http.Request, id string) (*Session, bool) {
	user, password, _ := r.BasicAuth()
	session, ok := h.sessions[id]
	if !ok {
		api.WriteBadRequestJSON(rw, fmt.Sprintf("session with id %s does not exist", id))
		return nil, false
	}

	if session.Username != "" || session.Password != "" {
		if session.Username != user || session.Password != password {
			rw.Header().Add("WWW-Authenticate", `Basic realm="Plexus Session", charset="UTF-8"`)
			api.WriteJSONError(rw, http.StatusUnauthorized, "invalid username / password")
			return nil, false
		}
	}
	return session, true
}

// SessionCreationAuth is middleware authenticating create session requests
func (h *Handler) SessionCreationAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		user, password, _ := r.BasicAuth()
		if h.cfg.AuthUser != "" || h.cfg.AuthPass != "" {
			if h.cfg.AuthUser != user || h.cfg.AuthPass != password {
				rw.Header().Add("WWW-Authenticate", `Basic realm="Plexus Session", charset="UTF-8"`)
				api.WriteJSONError(rw, http.StatusUnauthorized, "invalid username / password")
				return
			}
		}
		next.ServeHTTP(rw, r)
	}
}
