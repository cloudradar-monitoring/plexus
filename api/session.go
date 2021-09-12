package api

import "time"

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
}
