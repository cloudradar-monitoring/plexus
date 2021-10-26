package handler

import (
	"net/http"
	"sync"
	"time"

	"github.com/cloudradar-monitoring/plexus/api"
	"github.com/cloudradar-monitoring/plexus/control"
	"github.com/cloudradar-monitoring/plexus/logger"
)

type AuthChecker func(rw http.ResponseWriter, r *http.Request) bool

func New(cfg *control.Config, log logger.Logger, auth AuthChecker) *Handler {
	return &Handler{
		log:      log,
		auth:     auth,
		cfg:      cfg,
		sessions: make(map[string]*Session),
	}
}

type Handler struct {
	log      logger.Logger
	cfg      *control.Config
	auth     AuthChecker
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
