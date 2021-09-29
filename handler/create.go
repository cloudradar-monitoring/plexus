package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"

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
	id := r.FormValue("id")
	ttlStr := r.FormValue("ttl")
	user := r.FormValue("username")
	pass := r.FormValue("password")
	ttl, err := strconv.ParseInt(ttlStr, 10, 64)
	if err != nil {
		api.WriteBadRequestJSON(rw, fmt.Sprintf("invalid ttl %s: %s", id, err))
		return
	}

	h.lock.Lock()
	defer h.lock.Unlock()

	if _, ok := h.sessions[id]; ok {
		api.WriteBadRequestJSON(rw, fmt.Sprintf("session with id %s does already exist", id))
		return
	}

	mc, err := control.Connect(h.cfg)
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
			MeshServer: fmt.Sprintf("wss://%s/agent/%s:%s", r.Host, id, sessionToken),
			MeshType:   2,
		},
	}

	session.SetAgentName()
	h.sessions[id] = session

	go func() {
		<-time.After(time.Duration(ttl) * time.Second)
		h.lock.Lock()
		defer h.lock.Unlock()
		if s, ok := h.sessions[id]; ok {
			log.Info().Str("id", id).Msg("Session Expired")
			err := h.deleteInternal(s)
			if err != nil {
				log.Err(err).Str("id", id).Msg("Could not clean session")
			}
		}
	}()

	rw.Header().Add("content-type", "application/json")
	rw.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(rw).Encode(&api.Session{
		ID:          id,
		AgentMSH:    fmt.Sprintf("https://%s/config/%s:%s", r.Host, id, session.Token),
		SessionURL:  fmt.Sprintf("https://%s/session/%s", r.Host, id),
		AgentConfig: session.AgentConfig,
		ExpiresAt:   session.ExpiresAt,
	})
}
