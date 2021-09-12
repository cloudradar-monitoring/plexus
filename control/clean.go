package control

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
)

type Mesh struct {
	Name string `json:"name"`
	ID   string `json:"_id"`
}

func (m *MeshCentral) DeleteMeshes() error {
	request := map[string]interface{}{
		"action":     "meshes",
		"responseid": "meshctrl",
	}
	if err := m.Send(request); err != nil {
		return fmt.Errorf("could not send get meshes: %s", err)
	}
	payload, err := m.Get("meshes")
	if err != nil {
		return fmt.Errorf("could not get meshes: %s", err)
	}
	meshes := []Mesh{}
	if err := json.Unmarshal(payload["meshes"], &meshes); err != nil {
		return fmt.Errorf("could not parse meshes: %s", err)
	}

	for _, mesh := range meshes {
		if strings.HasPrefix(mesh.Name, m.cfg.MeshCentralGroupPrefix+"/") {
			log.Info().Str("name", mesh.Name).Msg("Remove Mesh")
			if err := m.DeleteMesh(mesh.ID); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *MeshCentral) DeleteMesh(id string) error {
	request := map[string]interface{}{
		"action":     "deletemesh",
		"meshid":     id,
		"responseid": "meshctrl",
	}
	if err := m.Send(request); err != nil {
		return fmt.Errorf("could not send delete meshes: %s", err)
	}
	if _, err := m.Get("deletemesh"); err != nil {
		return fmt.Errorf("could not delete meshes: %s", err)
	}
	return nil
}
