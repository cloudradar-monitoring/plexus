package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/cloudradar-monitoring/plexus/api"
)

// GetAgentMsh godoc
// @Summary Gets the meshagent.msh for the given session.
// @Tags session
// @Produce  text/plain
// @Param id path string true "session id"
// @Param token path string true "auth token"
// @Success 200 {object} string
// @Failure 400 {object} string
// @Failure 401 {object} string
// @Failure 404 {object} string
// @Failure 500 {object} string
// @Router /config/{id}:{token} [get]
func (h *Handler) GetAgentMsh(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	token := vars["token"]
	h.lock.Lock()
	defer h.lock.Unlock()

	session, ok := h.sessions[id]
	if !ok {
		api.WriteTextError(rw, http.StatusNotFound, fmt.Sprintf("session with id %s does not exist", id))
		return
	}

	if session.Token != token {
		api.WriteTextError(rw, http.StatusUnauthorized, "invalid token")
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
