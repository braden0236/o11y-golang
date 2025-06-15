package config

import (
	"strings"

	"github.com/spf13/viper"
)

type Options struct {
	Prefix string
}

type Option func(*Options)

func WithPrefix(prefix string) Option {
	return func(o *Options) {
		o.Prefix = prefix
	}
}

type Config struct {
	Server  ServerConfig
	Metrics MetricsConfig
}

type ServerConfig struct {
	Port int
}

type MetricsConfig struct {
	Username string
	Password string
}

var Cfg *Config

const (
	DefaultServerPort = 8080
)

func InitConfig(opts ...Option) (*Config,error) {

	options := &Options{}
	for _, opt := range opts {
		opt(options)
	}

	if options.Prefix != "" {
		viper.SetEnvPrefix(options.Prefix)
		viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	}

	viper.AutomaticEnv()

	Cfg = &Config{}
	Cfg.Server.Port = DefaultServerPort

	if port := viper.GetInt("SERVER_PORT"); port >= 1 && port <= 65535 {
		Cfg.Server.Port = port
	}

	Cfg.Metrics.Username = viper.GetString("METRICS_USERNAME")
	Cfg.Metrics.Password = viper.GetString("METRICS_PASSWORD")

	return Cfg, nil
}
