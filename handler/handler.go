package handler

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"

	"github.com/cloudradar-monitoring/plexus/api"
	"github.com/cloudradar-monitoring/plexus/control"
	"github.com/cloudradar-monitoring/plexus/logger"
	"github.com/cloudradar-monitoring/plexus/pairing"
)

type AuthChecker func(rw http.ResponseWriter, r *http.Request) bool

type Options struct {
	ControlConfig           *control.Config
	PairingConfig           *pairing.Config
	Log                     logger.Logger
	Auth                    AuthChecker
	Prefix                  string
	AllowSessionCredentials bool
}

func Register(r *mux.Router, opt *Options) {
	opt.Prefix = strings.TrimSuffix(opt.Prefix, "/")
	if !strings.HasPrefix(opt.Prefix, "/") {
		opt.Prefix = "/" + opt.Prefix
	}
	h := &Handler{
		log:                opt.Log,
		auth:               opt.Auth,
		ccfg:               opt.ControlConfig,
		pcfg:               opt.PairingConfig,
		prefix:             opt.Prefix,
		sessionCredentials: opt.AllowSessionCredentials,
		sessions:           make(map[string]*Session),
	}
	plexus := r.PathPrefix(opt.Prefix).Subrouter()
	plexus.HandleFunc("/session", h.CreateSession).Methods(http.MethodPost)
	plexus.HandleFunc("/session", h.ListSessions).Methods(http.MethodGet)
	plexus.HandleFunc("/session/{id}", h.ShareSession).Methods(http.MethodGet)
	plexus.HandleFunc("/session/{id}/url", h.ShareSessionURL).Methods(http.MethodGet)
	plexus.HandleFunc("/session/{id}", h.DeleteSession).Methods(http.MethodDelete)
	plexus.HandleFunc("/config/{id}:{token}", h.GetAgentMsh).Methods(http.MethodGet)
	plexus.HandleFunc("/agent/{id}:{token}", h.ProxyAgent).Methods(http.MethodGet)
	plexus.HandleFunc("/meshrelay.ashx", h.ProxyRelay).Methods(http.MethodGet)
	plexus.HandleFunc("/pairing/{code}", h.Pair).Methods(http.MethodGet)
	r.PathPrefix(h.ProxyMeshCentralURL()).Handler(h.ProxyMeshCentral())
}

type Handler struct {
	log                logger.Logger
	ccfg               *control.Config
	pcfg               *pairing.Config
	auth               AuthChecker
	lock               sync.RWMutex
	sessionCredentials bool
	sessions           map[string]*Session
	prefix             string
}

type Session struct {
	ID                 string
	Username, Password string
	ExpiresAt          time.Time
	Token              string
	AgentConfig        api.AgentConfig
	ShareURL           string
	SupporterName      string
	SupporterAvatar    string
	PairingCode        string
	PairingURL         string
	ProxyClose         func()
}

func (h *Handler) getSessionIDByPairingCode(code string) (*Session, bool) {
	for _, session := range h.sessions {
		if session.PairingCode == code {
			return session, true
		}
	}

	return nil, false
}
