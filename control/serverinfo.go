package control

import (
	"encoding/json"
	"fmt"
)

type serverInfo struct {
	AgentHash string `json:"agentCertHash"`
}

func (m *MeshCentral) ServerID() (string, error) {
	payload, err := m.Get("serverinfo")
	if err != nil {
		return "", fmt.Errorf("no event: %s", err)
	}

	info := serverInfo{}
	if err := json.Unmarshal(payload["serverinfo"], &info); err != nil {
		return "", fmt.Errorf("could not parse agent hash: %s", err)
	}

	return base64IDToHex(info.AgentHash)
}
