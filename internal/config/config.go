package config

import (
	"flag"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	ServerAddress string `env:"SERVER_ADDRESS"`
	BaseURL       string `env:"BASE_URL"`
}

var Cfg Config

func Load() {
	env.Parse(&Cfg)

	if Cfg.ServerAddress == "" {
		flag.StringVar(&Cfg.ServerAddress, "a", "localhost:8080", "HTTP server address")
	}

	if Cfg.BaseURL == "" {
		flag.StringVar(&Cfg.BaseURL, "b", "http://localhost:8080", "Base URL for shortened links")
	}
}
