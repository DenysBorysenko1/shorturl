package config

import (
	"flag"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	ServerAddress  string `env:"SERVER_ADDRESS"`
	BaseURL        string `env:"BASE_URL"`
	FileStorageURL string `env:"FILE_STORAGE_PATH"`
	DatabaseDSN    string `env:"DATABASE_DSN"`
	LogLevel       string `env:"LOG_LEVEL" envDefault:"info"`
}

var Cfg Config

func init() {
	flag.StringVar(&Cfg.ServerAddress, "a", "localhost:8080", "HTTP server address")
	flag.StringVar(&Cfg.BaseURL, "b", "http://localhost:8080", "Base URL for shortened links")
	flag.StringVar(&Cfg.FileStorageURL, "f", "", "Path to data file")
	flag.StringVar(&Cfg.DatabaseDSN, "d", "", "Database connection path")
}

func Load() {
	if !flag.Parsed() {
		flag.Parse()
	}

	_ = env.Parse(&Cfg)
}
