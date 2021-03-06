package config

import "github.com/rs/zerolog"

// FutureLog is an intermediate type for log messages. It is used before the config was loaded because without loaded
// config we do not know the log level, so we log these messages once the config was initialized.
type futureLog struct {
	Level zerolog.Level
	Msg   string
}

func futureFatal(msg string) futureLog {
	return futureLog{
		Level: zerolog.FatalLevel,
		Msg:   msg,
	}
}
