package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/cloudradar-monitoring/plexus/api"
	"github.com/go-chi/chi/v5"
)

func (h *Handler) GetAgentMsh(rw http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	token := chi.URLParam(r, "token")
	h.lock.Lock()
	defer h.lock.Unlock()

	session, ok := h.sessions[id]
	if !ok {
		api.WriteBadRequest(rw, fmt.Sprintf("session with id %s does not exist", id))
		return
	}

	if session.Token != token {
		api.WriteError(rw, http.StatusUnauthorized, "invalid token")
		return
	}
	rw.Header().Add("content-type", "text/plain")
	rw.WriteHeader(http.StatusOK)
	fmt.Fprintln(rw, "ServerID="+session.AgentConfig.ServerID)
	fmt.Fprintln(rw, "MeshName="+session.AgentConfig.MeshName)
	fmt.Fprintln(rw, "MeshType="+strconv.Itoa(session.AgentConfig.MeshType))
	fmt.Fprintln(rw, "MeshID="+session.AgentConfig.MeshIDHex)
	fmt.Fprintln(rw, "MeshServer="+session.AgentConfig.MeshServer)
}
