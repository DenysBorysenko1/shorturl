package config

import "flag"

type Config struct {
	ServerAddress string
	BaseURL       string
}

var Cfg *Config

func Load() {
	Cfg = &Config{}
	flag.StringVar(&Cfg.ServerAddress, "a", "localhost:8080", "HTTP server address")
	flag.StringVar(&Cfg.BaseURL, "b", "http://localhost:8080", "Base URL for shortened links")
	flag.Parse()
}
