package api

import (
	"encoding/json"
	"net/http"
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
}

type Error struct {
	Error       string
	Description string
}

func WriteBadRequest(rw http.ResponseWriter, message string) {
	WriteError(rw, http.StatusBadRequest, message)
}
func WriteBadGateway(rw http.ResponseWriter, message string) {
	WriteError(rw, http.StatusBadGateway, message)
}

func WriteError(rw http.ResponseWriter, code int, message string) {
	rw.WriteHeader(code)
	_ = json.NewEncoder(rw).Encode(err(code, message))
}

func err(code int, message string) *Error {
	return &Error{
		Error:       http.StatusText(code),
		Description: message,
	}
}
