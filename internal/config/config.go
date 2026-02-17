package config

import (
	"flag"
	"strings"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	ServerAddress string `env:"SERVER_ADDRESS"`
	BaseURL       string `env:"BASE_URL"`
}

var Cfg Config

func Load() {
	env.Parse(&Cfg)

	flagsRegistered := false

	if strings.TrimSpace(Cfg.ServerAddress) == "" {
		flag.StringVar(&Cfg.ServerAddress, "a", "localhost:8080", "HTTP server address")
		flagsRegistered = true
	}

	if strings.TrimSpace(Cfg.BaseURL) == "" {
		flag.StringVar(&Cfg.BaseURL, "b", "http://localhost:8080", "Base URL for shortened links")
		flagsRegistered = true
	}

	if flagsRegistered {
		flag.Parse()
	}
}
