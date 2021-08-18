package agent

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/cloudradar-monitoring/plexus/agent/handshake"
	"github.com/cloudradar-monitoring/plexus/config"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/rs/zerolog/log"
)

func Handler(cfg *config.Server) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		log.Debug().Str("status", "connecting").Msg("Agent")
		conn, _, _, err := ws.UpgradeHTTP(r, rw)
		if err != nil {
			rw.WriteHeader(400)
			log.Warn().Err(err).Msg("Could not Upgrade")
			return
		}
		log.Debug().Str("status", "upgraded").Msg("Agent")
		go func() {
			defer conn.Close()

			h := handshake.New(cfg, func(info *handshake.Info) {
				log.Info().Str("name", info.AgentName).
					Int32("capatibilies", info.Capabilities).
					Str("platform", info.Platform()).
					Int32("version", info.AgentVersion).
					Msg("Agent Handshake: Success")
			})
			for {
				msg, op, err := wsutil.ReadClientData(conn)
				if err != nil {
					log.Warn().Err(err).Msg("Agent Read")
					return
				}
				if op == ws.OpBinary {
					buffer := bytes.NewBuffer(msg)
					if err := h.Handle(buffer, conn); err != nil {
						log.Error().Err(err).Msg("Agent Handshake")
					}
				} else {

					fmt.Printf("Text %#v\n", string(msg))
				}
			}
		}()
	}
}
