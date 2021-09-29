package handler

import (
	"sync"
	"time"

	"github.com/cloudradar-monitoring/plexus/api"
	"github.com/cloudradar-monitoring/plexus/config"
)

func New(cfg *config.Config) *Handler {
	return &Handler{
		cfg:      cfg,
		sessions: make(map[string]*Session),
	}
}

type Handler struct {
	cfg      *config.Config
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

// SetAgentName sets Session ID as agentName in Session.AgentConfig
func (s *Session) SetAgentName() {
	s.AgentConfig.SetAgentName(s.ID)
}
