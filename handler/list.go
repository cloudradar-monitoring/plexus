package handler

import (
	"fmt"
	"net/http"

	"github.com/cloudradar-monitoring/plexus/api"
)

// ListSessions godoc
// @Summary Lists all active session.
// @Tags session
// @Produce  application/json
// @Success 200 {array} api.Session
// @Failure 400 {object} api.Error
// @Failure 401 {object} api.Error
// @Failure 500 {object} api.Error
// @Router /session [get]
func (h *Handler) ListSessions(rw http.ResponseWriter, r *http.Request) {
	h.lock.Lock()
	defer h.lock.Unlock()
	sessions := make([]*api.ListSessionsItem, 0, len(h.sessions))
	for _, s := range h.sessions {
		sessions = append(sessions, &api.ListSessionsItem{
			ID:              s.ID,
			SessionURL:      fmt.Sprintf("https://%s/session/%s", r.Host, s.ID),
			AgentMSH:        fmt.Sprintf("https://%s/config/%s:%s", r.Host, s.ID, s.Token),
			SessionUsername: s.Username,
			SessionPassword: s.Password,
			ExpiresAt:       s.ExpiresAt,
		})
	}
	api.WriteJSONResponse(rw, http.StatusOK, sessions)
}
