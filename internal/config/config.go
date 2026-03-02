package config

import (
	"flag"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	ServerAddress  string `env:"SERVER_ADDRESS"`
	BaseURL        string `env:"BASE_URL"`
	FileStorageURL string `env:"FILE_STORAGE_PATH"`
	LogLevel       string `env:"LOG_LEVEL" envDefault:"info"`
}

var Cfg Config

func Load() {
	flag.StringVar(&Cfg.ServerAddress, "a", "localhost:8080", "HTTP server address")
	flag.StringVar(&Cfg.BaseURL, "b", "http://localhost:8080", "Base URL for shortened links")
	flag.StringVar(&Cfg.FileStorageURL, "f", "./links.json", "Path to data file")

	flag.Parse()
	env.Parse(&Cfg)
}
