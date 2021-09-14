package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog"
)

var (
	prefix = "plexus"
)

// Config represents the application configuration.
type Config struct {
	LogLevel LogLevel `default:"info" split_words:"true"`
	LogFile  string   `split_words:"true"`

	ServerAddress string `default:":5050" split_words:"true"`

	TLSCertFile string `split_words:"true" required:"true"`
	TLSKeyFile  string `split_words:"true" required:"true"`

	MeshCentralURL         string `split_words:"true" required:"true"`
	MeshCentralGroupPrefix string `split_words:"true" default:"plexus"`
	MeshCentralUser        string `split_words:"true" required:"true"`
	MeshCentralPass        string `split_words:"true" required:"true"`

	ExternalHost string `split_words:"true"`
}

func (s *Config) MeshCentralControlURL() string {
	return s.MeshCentralURL + "/control.ashx"
}
func (s *Config) MeshCentralAgentURL() string {
	return s.MeshCentralURL + "/agent.ashx"
}
func (s *Config) MeshRelayURL() string {
	return s.MeshCentralURL + "/meshrelay.ashx"
}

// Get loads the application config.
func Get(file string) (Config, []futureLog) {
	var logs []futureLog

	_, fileErr := os.Stat(file)
	if fileErr == nil {
		if err := godotenv.Load(file); err != nil {
			logs = append(logs, futureFatal(fmt.Sprintf("cannot load file %s: %s", file, err)))
		} else {
			logs = append(logs, futureLog{
				Level: zerolog.DebugLevel,
				Msg:   fmt.Sprintf("Loading file %s", file)})
		}
	} else if os.IsNotExist(fileErr) {
		logs = append(logs, futureFatal(fmt.Sprintf("file %s does not exist", file)))
	} else {
		logs = append(logs, futureFatal(fmt.Sprintf("cannot read file %s because %s", file, fileErr)))
	}

	config := Config{
		LogLevel: LogLevel(zerolog.InfoLevel),
	}
	err := envconfig.Process(prefix, &config)
	if err != nil {
		logs = append(logs, futureFatal(fmt.Sprintf("cannot parse env params: %s", err)))
	}

	if strings.HasPrefix(config.MeshCentralURL, "http") {
		config.MeshCentralURL = strings.Replace(config.MeshCentralURL, "http", "ws", 1)
	}

	return config, logs
}
