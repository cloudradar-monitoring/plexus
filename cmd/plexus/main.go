package main

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"

	"github.com/cloudradar-monitoring/plexus/config"
	"github.com/cloudradar-monitoring/plexus/control"
	"github.com/cloudradar-monitoring/plexus/handler"
	"github.com/cloudradar-monitoring/plexus/logger"
	"github.com/cloudradar-monitoring/plexus/router"
	"github.com/cloudradar-monitoring/plexus/server"
)

func main() {
	app := cli.App{
		Commands: []*cli.Command{
			{
				Name: "serve",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "log-level",
						Value: "info",
					},
					&cli.StringFlag{
						Name:     "tls-cert",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "tls-key",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "meshcentral-url",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "meshcentral-user",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "meshcentral-pass",
						Required: true,
					},
					&cli.StringFlag{
						Name:  "meshcentral-group-prefix",
						Value: "plexus",
					},
					&cli.BoolFlag{
						Name: "meshcentral-insecure",
					},
					&cli.StringFlag{
						Name: "external-host",
					},
					&cli.StringFlag{
						Name:  "addr",
						Value: ":8080",
					},
				},
				Action: func(c *cli.Context) error {
					lvl, err := zerolog.ParseLevel(c.String("log-level"))
					if err != nil {
						return fmt.Errorf("could not parse log level: %s", err)
					}
					logger.Init(lvl)
					addr := c.String("addr")
					cfg := &config.Server{
						TLSPrivateKey:          c.String("tls-key"),
						TLSCertificate:         c.String("tls-cert"),
						MeshCentralURL:         c.String("meshcentral-url"),
						MeshCentralUsername:    c.String("meshcentral-user"),
						MeshCentralPassword:    c.String("meshcentral-pass"),
						MeshCentralInsecure:    c.Bool("meshcentral-insecure"),
						MeshCentralGroupPrefix: c.String("meshcentral-group-prefix"),
						ExternalHost:           c.String("external-host"),
					}

					mc, err := control.Connect(cfg)
					if err != nil {
						return fmt.Errorf("could not connect to meshcentral: %s", err)
					}
					serverID, err := mc.ServerID()
					if err != nil {
						mc.Close()
						return fmt.Errorf("could not get server id: %s", err)
					}
					log.Info().Str("user", cfg.MeshCentralUsername).Str("serverID", serverID).Msg("MeshControl: Authenticated")
					if err := mc.DeleteMeshes(); err != nil {
						return fmt.Errorf("could not clean old meshes: %s", err)
					}
					mc.Close()

					h := handler.New(cfg)
					r := router.New(h)
					log.Info().Str("addr", addr).Msg("Start listening")
					return server.Start(r, addr, cfg.TLSCertificate, cfg.TLSPrivateKey)
				},
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
