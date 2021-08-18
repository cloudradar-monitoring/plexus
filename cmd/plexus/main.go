package main

import (
	"encoding/hex"
	"os"

	"github.com/cloudradar-monitoring/plexus/config"
	"github.com/cloudradar-monitoring/plexus/logger"
	"github.com/cloudradar-monitoring/plexus/router"
	"github.com/cloudradar-monitoring/plexus/server"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

func main() {
	logger.Init(zerolog.DebugLevel)
	app := cli.App{
		Commands: []*cli.Command{
			{
				Name: "serve",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "web-cert",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "web-key",
						Required: true,
					},
					&cli.StringFlag{
						Name:  "addr",
						Value: ":8080",
					},
				},
				Action: func(c *cli.Context) error {
					cert := c.String("web-cert")
					key := c.String("web-key")
					addr := c.String("addr")
					cfg, err := config.NewServer(cert, key)
					if err != nil {
						return err
					}

					log.Debug().Str("ServerID", cfg.ServerID).Msg("Config")
					log.Debug().Str("Cert", hex.EncodeToString(cfg.CertificateHash[:])).Msg("Config")

					r := router.New(cfg)

					log.Info().Str("addr", addr).Msg("Start listening")
					return server.Start(r, addr, cert, key)
				},
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Error().Err(err).Msg("App")
		os.Exit(1)
	}
}
