package handler

import (
	"io"
	"net/http"

	"github.com/cloudradar-monitoring/plexus/api"
	"github.com/cloudradar-monitoring/plexus/control"
	"github.com/go-chi/chi/v5"
)

func (h *Handler) DeleteSession(rw http.ResponseWriter, r *http.Request) {
	session, ok := h.basicAuth(rw, r, chi.URLParam(r, "id"))
	if !ok {
		return
	}
	if err := h.deleteInternal(session); err != nil {
		api.WriteBadGateway(rw, err.Error())
		return
	}
	rw.Header().Add("content-type", "text/plain")
	rw.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(rw, "ok")
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
