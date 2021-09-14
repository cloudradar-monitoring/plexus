package logger

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

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
