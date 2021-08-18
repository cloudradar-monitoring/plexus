package handshake

import (
	"bytes"
	"crypto/x509"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"net"

	"github.com/cloudradar-monitoring/plexus/config"
	"github.com/rs/zerolog/log"
)

const nonceSize = 48

type Handshake struct {
	Server           *config.Server
	ServerNonce      []byte
	AgentNonce       []byte
	AgentNonceSet    bool
	AgentCertificate *x509.Certificate
	HasAgentInfo     bool
	AgentInfo        Info
	OnSuccess        func(*Info)
	Authenticated    bool
}

func New(cfg *config.Server, onSuccess func(*Info)) *Handshake {
	return &Handshake{
		Server:      cfg,
		ServerNonce: make([]byte, nonceSize),
		AgentNonce:  make([]byte, nonceSize),
		OnSuccess:   onSuccess,
	}
}

type Info struct {
	InfoVersion  int32
	AgentID      int32
	AgentVersion int32
	PlatformType int32
	MeshID       string
	Capabilities int32
	AgentName    string
}

func (i Info) Platform() string {
	switch i.PlatformType {
	case 1:
		return "desktop"
	case 2:
		return "laptop"
	case 3:
		return "mobile"
	case 4:
		return "server"
	case 5:
		return "disk"
	case 6:
		return "router"
	case 7:
		return "pi"
	case 8:
		return "virtual"
	default:
		return "unknown"
	}
}

func (h *Handshake) Handle(payload *bytes.Buffer, conn net.Conn) error {
	var cmd int16
	if err := binary.Read(payload, binary.BigEndian, &cmd); err != nil {
		return fmt.Errorf("cannot read cmd: %s", err)
	}
	switch cmd {
	case 30:
		log.Debug().Str("commit_date", payload.String()).Int16("cmd", cmd).Msg("Agent Handshake")
		return nil
	case 5:
		log.Debug().Str("server_id", hex.EncodeToString(payload.Bytes())).Int16("cmd", cmd).Msg("Agent Handshake")
		return nil
	case 4:
		log.Debug().Int16("cmd", cmd).Msg("Agent Handshake: Skip Certificate")
		return nil
	case 1:
		log.Debug().Int16("cmd", cmd).Msg("Agent Handshake: Auth Request")
		return agentAuthRequest(h, payload, conn)
	case 2:
		log.Debug().Int16("cmd", cmd).Msg("Agent Handshake: Validate Agent Cert")
		return validateAgentCert(h, payload, conn)
	case 3:
		log.Debug().Int16("cmd", cmd).Msg("Agent Handshake: Agent Info")
		return parseAgentInfo(h, payload, conn)
	default:
		log.Warn().Int16("cmd", cmd).Err(errors.New("unhandled command")).Msg("Agent Handshake")
		return nil
	}
}
