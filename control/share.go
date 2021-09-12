package control

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

var (
	ErrAgentNotConnected = errors.New("agent not connected")
)

type Share struct {
	URL string
}

type Node struct {
	ID   string `json:"_id"`
	Name string `json:"name"`
	PWR  int    `json:"pwr"`
	Conn int    `json:"conn"`
}

type MeshDeviceShare struct {
	URL string `json:"url"`
}

func (m *MeshCentral) Share(meshID, name string, expires time.Time) (string, error) {
	request := map[string]interface{}{
		"action":     "nodes",
		"meshid":     meshID,
		"responseid": "meshctrl",
	}
	if err := m.Send(request); err != nil {
		return "", fmt.Errorf("could not send get nodes: %s", err)
	}
	payload, err := m.Get("nodes")
	if err != nil {
		return "", fmt.Errorf("could not get create nodes response: %s", err)
	}
	if _, ok := payload["nodes"]; !ok {
		return "", ErrAgentNotConnected
	}
	meshToNodes := map[string][]*Node{}
	if err := json.Unmarshal(payload["nodes"], &meshToNodes); err != nil {
		return "", fmt.Errorf("could not unmarshal nodes response: %s", err)
	}
	nodes, ok := meshToNodes[meshID]
	if !ok {
		return "", ErrAgentNotConnected
	}

	var found *Node
	for _, node := range nodes {
		if node.PWR == 1 && node.Conn == 1 {
			found = node
			break
		}
	}

	if found == nil {
		return "", ErrAgentNotConnected
	}

	request = map[string]interface{}{
		"action":     "createDeviceShareLink",
		"nodeid":     found.ID,
		"p":          2, // desktop share
		"consent":    1, // desktop notify
		"guestname":  name,
		"responseid": "meshctrl",
		"start":      time.Now().Unix(),
		"end":        expires.Unix(),
	}

	if err := m.Send(request); err != nil {
		return "", fmt.Errorf("could not send create device sharing: %s", err)
	}
	payload, err = m.Get("createDeviceShareLink")
	if err != nil {
		return "", fmt.Errorf("could not get create response: %s", err)
	}

	url, err := payload.String("url")
	if err != nil {
		return "", fmt.Errorf("could not unmarshal url: %s", err)
	}

	return url, nil
}
