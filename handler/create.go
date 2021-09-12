package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/cloudradar-monitoring/plexus/api"
	"github.com/cloudradar-monitoring/plexus/control"
	"github.com/cloudradar-monitoring/plexus/token"
	"github.com/rs/zerolog/log"
)

func (h *Handler) CreateSession(rw http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	ttlStr := r.FormValue("ttl")
	user := r.FormValue("username")
	pass := r.FormValue("password")
	ttl, err := strconv.ParseInt(ttlStr, 10, 64)
	if err != nil {
		api.WriteBadRequest(rw, fmt.Sprintf("invalid ttl %s: %s", id, err))
		return
	}

	h.lock.Lock()
	defer h.lock.Unlock()

	if _, ok := h.sessions[id]; ok {
		api.WriteBadRequest(rw, fmt.Sprintf("session with id %s does already exist", id))
		return
	}

	mc, err := control.Connect(h.cfg)
	defer mc.Close()
	if err != nil {
		api.WriteBadGateway(rw, fmt.Sprintf("could not connect to mesh control: %s", err))
		return
	}
	mesh, err := mc.CreateMesh(h.cfg.MeshCentralGroupPrefix + "/" + id + "/" + token.New(5))
	if err != nil {
		api.WriteBadGateway(rw, fmt.Sprintf("could not create mesh: %s", err))
		return
	}
	serverID, err := mc.ServerID()
	if err != nil {
		api.WriteBadGateway(rw, fmt.Sprintf("could not get server id: %s", err))
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
			ServerID:   serverID,
			MeshServer: fmt.Sprintf("wss://%s/agent/%s:%s", h.cfg.Host(r), id, sessionToken),
			MeshType:   2,
		},
	}
	h.sessions[id] = session

	go func() {
		<-time.After(time.Duration(ttl) * time.Second)
		h.lock.Lock()
		defer h.lock.Unlock()
		if s, ok := h.sessions[id]; ok {
			log.Info().Str("id", id).Msg("Session Expired")
			h.deleteInternal(s)
		}
	}()

	rw.Header().Add("content-type", "application/json")
	rw.WriteHeader(http.StatusCreated)
	json.NewEncoder(rw).Encode(&api.Session{
		ID:          id,
		AgentMSH:    fmt.Sprintf("https://%s/config/%s:%s", h.cfg.Host(r), id, session.Token),
		SessionURL:  fmt.Sprintf("https://%s/session/%s", h.cfg.Host(r), id),
		AgentConfig: session.AgentConfig,
		ExpiresAt:   session.ExpiresAt,
	})
}
