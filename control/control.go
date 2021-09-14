package control

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/rs/zerolog/log"

	"github.com/cloudradar-monitoring/plexus/config"
)

const GetTimeout = 5 * time.Second

func Connect(cfg *config.Config) (*MeshCentral, error) {
	mc := &MeshCentral{
		pendingActions: make(map[string]Payload),
		waitFor:        make(map[string]chan<- Payload),
		cfg:            cfg,
	}

	err := mc.connect()
	return mc, err
}

type MeshCentral struct {
	mutex          sync.Mutex
	pendingActions map[string]Payload
	waitFor        map[string]chan<- Payload
	cfg            *config.Config
	conn           net.Conn
}

func (m *MeshCentral) Close() {
	log.Debug().Msg("MeshControl: Disconnect")
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if m.conn != nil {
		m.conn.Close()
	}
}

func (m *MeshCentral) connect() error {
	user := base64.StdEncoding.EncodeToString([]byte(m.cfg.MeshCentralUser))
	pass := base64.StdEncoding.EncodeToString([]byte(m.cfg.MeshCentralPass))
	auth := user + "," + pass
	conn, _, _, err := ws.Dialer{
		Header: ws.HandshakeHeaderHTTP(http.Header{
			"x-meshauth": []string{auth},
		}),
	}.Dial(context.Background(), m.cfg.MeshCentralControlURL())
	if err != nil {
		log.Error().Err(err).Msg("MeshControl: Connect")
		return fmt.Errorf("could not connect to control server: %s", err)
	}
	log.Debug().Msg("MeshControl: Connected")
	m.conn = conn

	go func() {
		for {
			msg, err := wsutil.ReadServerText(conn)
			if err == io.EOF {
				return
			}
			if err != nil {
				if err == io.EOF || strings.Contains(err.Error(), "use of closed network connection") {
					log.Trace().Err(err).Str("body", string(msg)).Msg("Control Read")
					return
				}
				log.Warn().Err(err).Str("body", string(msg)).Msg("Control Read")
				return
			}

			payload := Payload{}
			if err := json.Unmarshal(msg, &payload); err != nil {
				conn.Close()
				log.Warn().Err(err).Msg("Control Invalid Json")
				return
			}

			action := payload.Action()
			log.Debug().Interface("payload", payload).Msg("Control Read")

			func() {
				m.mutex.Lock()
				defer m.mutex.Unlock()
				if callback, ok := m.waitFor[action]; ok {
					callback <- payload
					delete(m.waitFor, action)
				} else {
					m.pendingActions[action] = payload
				}
			}()
		}
	}()

	return nil
}

func (m *MeshCentral) Get(action string) (Payload, error) {
	m.mutex.Lock()
	if pending, ok := m.pendingActions[action]; ok {
		delete(m.pendingActions, action)
		m.mutex.Unlock()
		return pending, nil
	}

	if _, ok := m.waitFor[action]; ok {
		panic("canont wait multiple times")
	}

	callback := make(chan Payload)
	m.waitFor[action] = callback

	m.mutex.Unlock()

	select {
	case payload := <-callback:
		return payload, payload.Error()
	case <-time.After(GetTimeout):
		m.mutex.Lock()
		defer m.mutex.Unlock()
		select {
		case payload := <-callback:
			return payload, payload.Error()
		default:
			delete(m.waitFor, action)
			return nil, fmt.Errorf("get control timeouted: %s", action)
		}
	}
}

func (m *MeshCentral) Send(payload map[string]interface{}) error {
	log.Debug().Interface("payload", payload).Msg("Control Write")
	payloadBytes, err := json.Marshal(&payload)
	if err != nil {
		return err
	}
	return wsutil.WriteClientText(m.conn, payloadBytes)
}
