package proxy

import (
	"io"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

var Upgrader = websocket.Upgrader{
	HandshakeTimeout: 10 * time.Second,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func Hold(rw http.ResponseWriter, r *http.Request) {
	agentConn, err := Upgrader.Upgrade(rw, r, nil)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(rw, "upgrade failed "+err.Error())
		log.Info().Err(err).Msg("Proxy: upgrade failed")
		return
	}
	go func() {
		defer agentConn.Close()
		for {
			// ignore messages
			_, _, err := agentConn.ReadMessage()
			if err != nil {
				break
			}
		}
	}()
}

func Proxy(rw http.ResponseWriter, r *http.Request, target string) (func(), bool) {
	agentConn, err := Upgrader.Upgrade(rw, r, nil)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(rw, "upgrade failed "+err.Error())
		log.Info().Err(err).Msg("Proxy: upgrade failed")
		return nil, false
	}
	serverConn, _, err := websocket.DefaultDialer.Dial(target, nil)
	if err != nil {
		rw.WriteHeader(http.StatusBadGateway)
		_, _ = io.WriteString(rw, "could not reach meshcentral server "+err.Error())
		log.Error().Err(err).Msg("Proxy: meshcentral unavailable")
		return nil, false
	}

	closeAll := func() {
		_ = agentConn.Close()
		_ = serverConn.Close()
	}

	go func() {
		defer closeAll()
		for {
			t, msg, err := agentConn.ReadMessage()
			if err != nil {
				break
			}
			err = serverConn.WriteMessage(t, msg)
			if err != nil {
				break
			}
		}
	}()
	go func() {
		defer closeAll()
		for {
			t, msg, err := serverConn.ReadMessage()
			if err != nil {
				break
			}
			err = agentConn.WriteMessage(t, msg)
			if err != nil {
				break
			}
		}
	}()
	return closeAll, true
}
