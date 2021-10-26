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

type Options struct {
	Config                  *control.Config
	Log                     logger.Logger
	Auth                    AuthChecker
	AllowSessionCredentials bool
}

func New(opt *Options) *Handler {
	return &Handler{
		log:                opt.Log,
		auth:               opt.Auth,
		cfg:                opt.Config,
		sessionCredentials: opt.AllowSessionCredentials,
		sessions:           make(map[string]*Session),
	}
}

type Handler struct {
	log                logger.Logger
	cfg                *control.Config
	auth               AuthChecker
	lock               sync.RWMutex
	sessionCredentials bool
	sessions           map[string]*Session
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
