package control

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"

	"github.com/cloudradar-monitoring/plexus/config"
)

const GetTimeout = 5 * time.Second

func Connect(cfg *config.Config) (*MeshCentral, error) {
	mc := &MeshCentral{
		pendingActions: make(map[string]Payload),
		waitFor:        make(map[string]chan<- Payload),
		dead:           make(chan error, 1),
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
	conn           *websocket.Conn
	dead           chan error
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
	conn, _, err := websocket.DefaultDialer.Dial(m.cfg.MeshCentralControlURL(), http.Header{
		"x-meshauth": []string{auth},
	})
	if err != nil {
		log.Error().Err(err).Msg("MeshControl: Connect")
		return fmt.Errorf("could not connect to control server: %s", err)
	}
	log.Debug().Msg("MeshControl: Connected")
	m.conn = conn

	go func() {
		for {
			payload := Payload{}
			if err := conn.ReadJSON(&payload); err != nil {
				if err == io.EOF {
					return
				}
				if errors.Is(err, net.ErrClosed) {
					log.Trace().Err(err).Interface("body", &payload).Msg("Control Read")
				} else {
					log.Warn().Err(err).Interface("body", &payload).Msg("Control Read")
				}
				m.dead <- err
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
	select {
	case err := <-m.dead:
		return nil, err
	default:
	}
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
	case err := <-m.dead:
		return nil, err
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
	select {
	case err := <-m.dead:
		return err
	default:
	}
	log.Debug().Interface("payload", payload).Msg("Control Write")
	return m.conn.WriteJSON(&payload)
}
