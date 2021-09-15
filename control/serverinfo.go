package control

import (
	"encoding/json"
	"fmt"
)

type ServerInfo struct {
	AgentHash string `json:"agentCertHash"`
	TLSHash   string `json:"tlshash"`
}

func (m *MeshCentral) ServerInfo() (*ServerInfo, error) {
	payload, err := m.Get("serverinfo")
	if err != nil {
		return nil, fmt.Errorf("could not get server info: %s", err)
	}

	info := ServerInfo{}
	if err := json.Unmarshal(payload["serverinfo"], &info); err != nil {
		return nil, fmt.Errorf("could not parse server info: %s", err)
	}

	info.AgentHash, err = base64IDToHex(info.AgentHash)
	return &info, err
}
