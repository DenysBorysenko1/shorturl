// Package config provides configuration management for the application.
// It loads configuration from command-line flags and environment variables,
// with support for server settings, database connections, storage options,
// JWT authentication, logging, and audit configuration.
package config

import (
	"flag"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	ServerAddress  string `env:"SERVER_ADDRESS"`
	BaseURL        string `env:"BASE_URL"`
	FileStorageURL string `env:"FILE_STORAGE_PATH"`
	DatabaseDSN    string `env:"DATABASE_DSN"`
	JWTSecret      string `env:"JWT_SECRET" envDefault:"secret"`
	JWTExpiration  int    `env:"JWT_EXT" envDefault:"36000"`
	LogLevel       string `env:"LOG_LEVEL" envDefault:"info"`
	AuditFile      string `env:"AUDIT_FILE"`
	AuditURL       string `env:"AUDIT_URL"`
}

var Cfg Config

func init() {
	flag.StringVar(&Cfg.ServerAddress, "a", "localhost:8080", "HTTP server address")
	flag.StringVar(&Cfg.BaseURL, "b", "http://localhost:8080", "Base URL for shortened links")
	flag.StringVar(&Cfg.FileStorageURL, "f", "", "Path to data file")
	flag.StringVar(&Cfg.DatabaseDSN, "d", "", "Database connection path")
	flag.StringVar(&Cfg.JWTSecret, "j", "", "JWT secret")
	flag.IntVar(&Cfg.JWTExpiration, "e", 0, "JWT lifetime in seconds")
	flag.StringVar(&Cfg.AuditFile, "audit-file", "", "Path to audit log file")
	flag.StringVar(&Cfg.AuditURL, "audit-url", "", "URL of remote audit server")
}

func Load() {
	if !flag.Parsed() {
		flag.Parse()
	}

	_ = env.Parse(&Cfg)
}
