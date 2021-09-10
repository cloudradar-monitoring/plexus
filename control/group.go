package control

import (
	"encoding/json"
	"fmt"

	"github.com/rs/zerolog/log"
)

type CreatedMesh struct {
	Name  string
	IDHex string
	ID    string
}

func (m *MeshCentral) CreateMesh(name string) (*CreatedMesh, error) {
	if err := m.Send(map[string]interface{}{
		"action":     "createmesh",
		"meshname":   name,
		"meshtype":   2,
		"desc":       "used by https://github.com/cloudradar-monitoring/plexus",
		"responseid": "meshctrl"}); err != nil {
		return nil, fmt.Errorf("could not send create group: %s", err)
	}

	payload, err := m.Get("createmesh")
	if err != nil {
		return nil, fmt.Errorf("could not get create mesh response: %s", err)
	}

	meshid := ""
	if err := json.Unmarshal(payload["meshid"], &meshid); err != nil {
		return nil, fmt.Errorf("could not parse create mesh id: %s", err)
	}
	hexID, err := base64MeshIDToHexMeshID(meshid)
	if err != nil {
		return nil, err
	}
	log.Debug().Str("name", name).Msg("MeshControl: Created Mesh")
	return &CreatedMesh{
		Name:  name,
		IDHex: hexID,
		ID:    meshid,
	}, nil
}
