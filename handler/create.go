package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"strconv"
	"time"

	"github.com/cloudradar-monitoring/plexus/api"
	"github.com/cloudradar-monitoring/plexus/control"
	"github.com/cloudradar-monitoring/plexus/token"
)

// CreateSession godoc
// @Summary Create a session.
// @Description Create a plexus session where meshagents can connect to.
// @Tags session
// @Accept application/x-www-form-urlencoded
// @Produce  application/json
// @Param id formData string true "session id"
// @Param ttl formData int true "the time to live for the session"
// @Param username formData string true "the credentials to open the remote control interface & delete the session"
// @Param password formData string true "the credentials to open the remote control interface & delete the session"
// @Success 200 {object} api.Session
// @Failure 400 {object} api.Error
// @Failure 500 {object} api.Error
// @Failure 502 {object} api.Error
// @Router /session [post]
func (h *Handler) CreateSession(rw http.ResponseWriter, r *http.Request) {
	if !h.auth(rw, r) {
		return
	}
	id := r.FormValue("id")
	ttlStr := r.FormValue("ttl")
	user := r.FormValue("username")
	pass := r.FormValue("password")
	ttl, err := strconv.ParseInt(ttlStr, 10, 64)
	if err != nil {
		api.WriteBadRequestJSON(rw, fmt.Sprintf("invalid ttl %s: %s", id, err))
		return
	}

	if !h.sessionCredentials && (user != "" || pass != "") {
		api.WriteBadRequestJSON(rw, "session credentials are not allowed")
		return
	}

	h.lock.Lock()
	defer h.lock.Unlock()

	if _, ok := h.sessions[id]; ok {
		api.WriteBadRequestJSON(rw, fmt.Sprintf("session with id %s does already exist", id))
		return
	}

	mc, err := control.Connect(h.cfg, h.log)
	defer mc.Close()
	if err != nil {
		api.WriteBadGatewayJSON(rw, fmt.Sprintf("could not connect to mesh control: %s", err))
		return
	}
	mesh, err := mc.CreateMesh(h.cfg.MeshCentralGroupPrefix + "/" + id + "/" + token.New(5))
	if err != nil {
		api.WriteBadGatewayJSON(rw, fmt.Sprintf("could not create mesh: %s", err))
		return
	}
	serverInfo, err := mc.ServerInfo()
	if err != nil {
		api.WriteBadGatewayJSON(rw, fmt.Sprintf("could not get server id: %s", err))
		return
	}
	sessionToken := token.NewAuth()

	session := &Session{
		Token:     sessionToken,
		ID:        id,
		Username:  user,
		Password:  pass,
		ExpiresAt: time.Now().Add(time.Duration(ttl) * time.Second),
		AgentConfig: api.AgentConfig{
			MeshName:   mesh.Name,
			MeshIDHex:  mesh.IDHex,
			MeshID:     mesh.ID,
			ServerID:   serverInfo.AgentHash,
			MeshServer: fmt.Sprintf("wss://%s%s", r.Host, path.Join(h.prefix, "agent", fmt.Sprintf("%s:%s", id, sessionToken))),
			MeshType:   2,
		},
	}
	h.sessions[id] = session

	go func() {
		<-time.After(time.Duration(ttl) * time.Second)
		h.lock.Lock()
		defer h.lock.Unlock()
		if s, ok := h.sessions[id]; ok {
			h.log.Infof("Session %s expired", id)
			err := h.deleteInternal(s)
			if err != nil {
				h.log.Errorf("Could not clean session %s: %s", id, err)
			}
		}
	}()

	rw.Header().Add("content-type", "application/json")
	rw.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(rw).Encode(&api.Session{
		ID:          id,
		AgentMSH:    fmt.Sprintf("https://%s%s", r.Host, path.Join(h.prefix, "config", fmt.Sprintf("%s:%s", id, sessionToken))),
		SessionURL:  fmt.Sprintf("https://%s%s", r.Host, path.Join(h.prefix, "session", id)),
		AgentConfig: session.AgentConfig,
		ExpiresAt:   session.ExpiresAt,
	})
}
