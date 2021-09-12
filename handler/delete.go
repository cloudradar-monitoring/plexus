package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/cloudradar-monitoring/plexus/api"
	"github.com/cloudradar-monitoring/plexus/control"
)

// DeleteSession godoc
// @Summary Delete a session.
// @Tags session
// @Produce  application/json
// @Param id path string true "session id"
// @Security BasicAuth
// @Success 200 {object} api.Result
// @Failure 400 {object} api.Error
// @Failure 500 {object} api.Error
// @Failure 502 {object} api.Error
// @Router /session/{id} [delete]
func (h *Handler) DeleteSession(rw http.ResponseWriter, r *http.Request) {
	session, ok := h.basicAuth(rw, r, chi.URLParam(r, "id"))
	if !ok {
		return
	}
	if err := h.deleteInternal(session); err != nil {
		api.WriteBadGatewayJSON(rw, err.Error())
		return
	}
	api.WriteResult(rw, http.StatusOK, "ok")
}

func (h *Handler) deleteInternal(s *Session) error {
	delete(h.sessions, s.ID)
	if s.ProxyClose != nil {
		s.ProxyClose()
	}

	mc, err := control.Connect(h.cfg)
	if err != nil {
		return err
	}
	defer mc.Close()
	if err := mc.DeleteMesh(s.AgentConfig.MeshID); err != nil {
		return err
	}
	return nil
}
