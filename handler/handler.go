package handler

import (
	"sync"
	"time"

	"github.com/cloudradar-monitoring/plexus/api"
	"github.com/cloudradar-monitoring/plexus/config"
	"github.com/cloudradar-monitoring/plexus/logger"
)

func New(cfg *config.Config, log logger.Logger) *Handler {
	return &Handler{
		log:      log,
		cfg:      cfg,
		sessions: make(map[string]*Session),
	}
}

type Handler struct {
	log      logger.Logger
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
