package api

import (
	"strings"
	"time"
)

type Session struct {
	ID          string
	SessionURL  string
	AgentMSH    string
	AgentConfig AgentConfig
	ExpiresAt   time.Time
}

type AgentConfig struct {
	ServerID   string
	MeshName   string
	MeshType   int
	MeshID     string
	MeshIDHex  string
	MeshServer string
	AgentName  string `json:"agentName"`
}

// SetAgentName sets the sessionID as agentName in AgentConfig
func (ac *AgentConfig) SetAgentName(sessionID string) {
	ac.AgentName = strings.ReplaceAll(sessionID, " ", "_")
}
