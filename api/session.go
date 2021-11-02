package api

import (
	"time"
)

type Session struct {
	ID          string
	SessionURL  string
	AgentMSH    string
	PairingCode string `json:"PairingCode,omitempty"`
	PairingUrl  string `json:"PairingUrl,omitempty"`
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

type ListSessionsItem struct {
	ID              string
	SessionURL      string
	AgentMSH        string
	SessionUsername string `json:",omitempty"`
	SessionPassword string `json:",omitempty"`
	ExpiresAt       time.Time
}

type PairedSession struct {
	AgentMSH        string
	SupporterName   string
	SupporterAvatar string
	CompanyName     string
	CompanyLogo     string
	ExpiresAt       time.Time
}
