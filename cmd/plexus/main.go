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
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "config",
				Aliases:  []string{"c"},
				Usage:    "path to the config",
				Required: true,
			},
		},
		Action: func(c *cli.Context) error {
			cfg, errs := config.Get(c.String("config"))
			logger.Init(cfg.LogLevel.AsZeroLogLevel())

			exit := false
			for _, err := range errs {
				log.WithLevel(err.Level).Msg(err.Msg)
				exit = exit || err.Level == zerolog.FatalLevel || err.Level == zerolog.PanicLevel
			}
			if exit {
				os.Exit(1)
			}
			log.Debug().Interface("config", cfg).Msg("Using")

			mc, err := control.Connect(&cfg)
			if err != nil {
				return fmt.Errorf("could not connect to meshcentral: %s", err)
			}
			serverID, err := mc.ServerID()
			if err != nil {
				mc.Close()
				return fmt.Errorf("could not get server id: %s", err)
			}
			log.Info().Str("serverID", serverID).Msg("MeshControl: Authenticated")
			if err := mc.DeleteMeshes(); err != nil {
				return fmt.Errorf("could not clean old meshes: %s", err)
			}
			mc.Close()

			h := handler.New(&cfg)
			r := router.New(h)
			log.Info().Str("addr", cfg.ServerAddress).Msg("Start listening")
			return server.Start(r, cfg.ServerAddress, cfg.TLSCertFile, cfg.TLSKeyFile)
		},
	}
	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
