package handshake

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"net"
	"strings"
)

func parseAgentInfo(handshake *Handshake, payload *bytes.Buffer, conn net.Conn) error {
	if err := binary.Read(payload, binary.BigEndian, &handshake.AgentInfo.InfoVersion); err != nil {
		return fmt.Errorf("could not read info version: %s", err)
	}
	if err := binary.Read(payload, binary.BigEndian, &handshake.AgentInfo.AgentID); err != nil {
		return fmt.Errorf("could not read agent id: %s", err)
	}
	if err := binary.Read(payload, binary.BigEndian, &handshake.AgentInfo.AgentVersion); err != nil {
		return fmt.Errorf("could not read agent version: %s", err)
	}
	if err := binary.Read(payload, binary.BigEndian, &handshake.AgentInfo.PlatformType); err != nil {
		return fmt.Errorf("could not read agent version: %s", err)
	}
	if handshake.AgentInfo.PlatformType > 8 || handshake.AgentInfo.PlatformType < 1 {
		handshake.AgentInfo.PlatformType = 1
	}
	meshIDBytes := make([]byte, 48)
	if _, err := payload.Read(meshIDBytes); err != nil {
		return fmt.Errorf("could not read mesh id: %s", err)
	}

	meshID := base64.StdEncoding.EncodeToString(meshIDBytes)
	meshID = strings.ReplaceAll(meshID, "+", "@")
	meshID = strings.ReplaceAll(meshID, "/", "$")
	handshake.AgentInfo.MeshID = meshID

	if err := binary.Read(payload, binary.BigEndian, &handshake.AgentInfo.Capabilities); err != nil {
		return fmt.Errorf("could not read capatibilities: %s", err)
	}
	if payload.Len() > 0 {
		var nameLen int16
		if err := binary.Read(payload, binary.BigEndian, &nameLen); err != nil {
			return fmt.Errorf("could not read agent name len: %s", err)
		}
		nameBytes := make([]byte, nameLen)
		if _, err := payload.Read(nameBytes); err != nil {
			return fmt.Errorf("could not read agent name: %s", err)
		}
		handshake.AgentInfo.AgentName = string(nameBytes)
	}
	handshake.HasAgentInfo = true
	if handshake.Authenticated {
		handshake.OnSuccess(&handshake.AgentInfo)
	}
	return nil
}
