package zerologger

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/cloudradar-monitoring/plexus/logger"
)

var _ logger.Logger = (*zeroLogger)(nil)

type zeroLogger struct{ *zerolog.Logger }

func (z *zeroLogger) Errorf(f string, args ...interface{}) {
	z.Debug().Msgf(f, args...)
}
func (z *zeroLogger) Infof(f string, args ...interface{}) {
	z.Info().Msgf(f, args...)
}
func (z *zeroLogger) Debugf(f string, args ...interface{}) {
	z.Debug().Msgf(f, args...)
}

func Get() logger.Logger {
	return &zeroLogger{Logger: &log.Logger}
}

type c struct {
}

func (*c) Close() error {
	return nil
}

// Init initializes the logger
func Init(lvl zerolog.Level, file string) (io.Closer, error) {
	if file == "" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}).Level(lvl)
		return &c{}, nil
	}
	f, err := os.OpenFile(file, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("could not open logfile %s: %s", file, err)
	}
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: io.MultiWriter(os.Stdout, f), NoColor: true, TimeFormat: time.RFC3339}).Level(lvl)
	return f, nil
}
