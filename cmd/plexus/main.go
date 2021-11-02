package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"

	"github.com/cloudradar-monitoring/plexus/api"
	"github.com/cloudradar-monitoring/plexus/config"
	"github.com/cloudradar-monitoring/plexus/control"
	"github.com/cloudradar-monitoring/plexus/handler"
	"github.com/cloudradar-monitoring/plexus/logger/zerologger"
	"github.com/cloudradar-monitoring/plexus/server"
	"github.com/cloudradar-monitoring/plexus/verify"
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
		Commands: []*cli.Command{serve, verifyConfig},
	}
	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var verifyConfig = &cli.Command{
	Name:        "verify-config",
	Description: "Verifies the plexus config and checks if everything should work.",
	Action: func(c *cli.Context) error {
		if !verify.Verify(c.String("config")) {
			os.Exit(1)
		}
		return nil
	},
}

var serve = &cli.Command{
	Name:        "serve",
	Description: "Serves plexus",
	Action: func(c *cli.Context) error {
		cfg, errs := config.Get(c.String("config"))
		handle, err := zerologger.Init(cfg.LogLevel.AsZeroLogLevel(), cfg.LogFile)
		if err != nil {
			return err
		}
		defer handle.Close()
		exit := false
		for _, err := range errs {
			log.WithLevel(err.Level).Msg(err.Msg)
			exit = exit || err.Level == zerolog.FatalLevel || err.Level == zerolog.PanicLevel
		}
		if exit {
			os.Exit(1)
		}
		log.Debug().Interface("config", cfg).Msg("Using")

		mc, err := control.Connect(cfg.AsControlConfig(), zerologger.Get())
		if err != nil {
			return fmt.Errorf("could not connect to meshcentral: %s", err)
		}
		serverInfo, err := mc.ServerInfo()
		if err != nil {
			mc.Close()
			return fmt.Errorf("could not get server id: %s", err)
		}
		log.Info().Str("serverID", serverInfo.AgentHash).Msg("MeshControl: Authenticated")
		if err := mc.DeleteMeshes(); err != nil {
			return fmt.Errorf("could not clean old meshes: %s", err)
		}
		mc.Close()

		var auth handler.AuthChecker = func(rw http.ResponseWriter, r *http.Request) bool {
			user, password, _ := r.BasicAuth()
			if cfg.AuthUser != "" || cfg.AuthPass != "" {
				if cfg.AuthUser != user || cfg.AuthPass != password {
					rw.Header().Add("WWW-Authenticate", `Basic realm="Plexus Session", charset="UTF-8"`)
					api.WriteJSONError(rw, http.StatusUnauthorized, "invalid username / password")
					return false
				}
			}
			return true
		}

		opt := &handler.Options{
			ControlConfig:           cfg.AsControlConfig(),
			PairingConfig:           cfg.AsPairingConfig(),
			Auth:                    auth,
			Log:                     zerologger.Get(),
			Prefix:                  cfg.PathPrefix,
			AllowSessionCredentials: true}

		r := mux.NewRouter()
		r.Use(accessLog)
		handler.Register(r, opt)

		log.Info().Str("addr", cfg.ServerAddress).Msg("Start listening")
		return server.Start(r, cfg.ServerAddress, cfg.TLSCertFile, cfg.TLSKeyFile)
	},
}

func accessLog(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		start := time.Now()
		h.ServeHTTP(rw, r)
		duration := time.Since(start)

		log.Debug().
			Str("host", r.Host).
			Str("ip", r.RemoteAddr).
			Str("path", r.URL.Path).
			Str("duration", duration.String()).
			Msg("HTTP")
	})
}
