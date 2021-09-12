package config

import (
	"net/http"
)

type Server struct {
	ExternalHost           string
	TLSPrivateKey          string
	TLSCertificate         string
	MeshCentralURL         string
	MeshCentralUsername    string
	MeshCentralPassword    string
	MeshCentralInsecure    bool
	MeshCentralGroupPrefix string
	ServerID               string
}

func (s *Server) MeshCentralControlURL() string {
	return s.MeshCentralURL + "/control.ashx"
}
func (s *Server) MeshCentralAgentURL() string {
	return s.MeshCentralURL + "/agent.ashx"
}
func (s *Server) MeshRelayURL() string {
	return s.MeshCentralURL + "/meshrelay.ashx"
}

func (s *Server) Host(r *http.Request) string {
	if s.ExternalHost != "" {
		return s.ExternalHost
	}
	return r.Host
}
