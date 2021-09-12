package proxy

import (
	"context"
	"crypto/tls"
	"io"
	"net/http"

	"github.com/gobwas/ws"
	"github.com/rs/zerolog/log"
)

func Hold(rw http.ResponseWriter, r *http.Request) {
	agentConn, _, _, err := ws.UpgradeHTTP(r, rw)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(rw, "upgrade failed "+err.Error())
		log.Info().Err(err).Msg("Proxy: upgrade failed")
		return
	}
	go func() {
		defer agentConn.Close()
		_, _ = io.Copy(io.Discard, agentConn)
	}()
}

func Proxy(rw http.ResponseWriter, r *http.Request, target string, insecure bool) (func(), bool) {
	agentConn, _, _, err := ws.UpgradeHTTP(r, rw)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(rw, "upgrade failed "+err.Error())
		log.Info().Err(err).Msg("Proxy: upgrade failed")
		return nil, false
	}
	/* #nosec */
	serverConn, _, _, err := ws.Dialer{
		TLSConfig: &tls.Config{
			InsecureSkipVerify: insecure,
		},
	}.Dial(context.Background(), target)
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
		_, _ = io.Copy(agentConn, serverConn)
	}()
	go func() {
		defer closeAll()
		_, _ = io.Copy(serverConn, agentConn)
	}()
	return closeAll, true
}
