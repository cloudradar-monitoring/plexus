package handler

import (
	"sync"
	"time"

	"github.com/cloudradar-monitoring/plexus/api"
	"github.com/cloudradar-monitoring/plexus/config"
)

func New(cfg *config.Server) *Handler {
	return &Handler{
		cfg:      cfg,
		sessions: make(map[string]*Session),
	}
}

type Handler struct {
	cfg      *config.Server
	lock     sync.RWMutex
	sessions map[string]*Session
}
type Session struct {
	ID                 string
	Username, Password string
	ExpiresAt          time.Time
	Token              string
	AgentConfig        api.AgentConfig
	ShareURL           string
	ProxyClose         func()
}
