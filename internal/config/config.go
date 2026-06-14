// Package config provides configuration management for the application.
// It loads configuration from command-line flags, environment variables,
// and an optional JSON configuration file, with support for server settings,
// database connections, storage options, JWT authentication, logging, and
// audit configuration.
//
// Priority order (highest to lowest):
//  1. Environment variables
//  2. Command-line flags
//  3. Configuration file values
//  4. Default values
package config

import (
	"encoding/json"
	"flag"
	"os"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	ServerAddress  string `env:"SERVER_ADDRESS"    json:"server_address"`
	BaseURL        string `env:"BASE_URL"          json:"base_url"`
	FileStorageURL string `env:"FILE_STORAGE_PATH" json:"file_storage_path"`
	DatabaseDSN    string `env:"DATABASE_DSN"      json:"database_dsn"`
	JWTSecret      string `env:"JWT_SECRET"        json:"jwt_secret"       envDefault:"secret"`
	JWTExpiration  int    `env:"JWT_EXT"           json:"jwt_expiration"   envDefault:"36000"`
	LogLevel       string `env:"LOG_LEVEL"         json:"log_level"        envDefault:"info"`
	AuditFile      string `env:"AUDIT_FILE"        json:"audit_file"`
	AuditURL       string `env:"AUDIT_URL"         json:"audit_url"`
	EnableHTTPS    bool   `env:"ENABLE_HTTPS"      json:"enable_https"`
	TLSCertFile    string `env:"TLS_CERT_FILE"     json:"tls_cert_file"    envDefault:"certs/cert.pem"`
	TLSKeyFile     string `env:"TLS_KEY_FILE"      json:"tls_key_file"     envDefault:"certs/key.pem"`
}

var (
	Cfg        Config
	configFile string
)

func init() {
	flag.StringVar(&configFile, "config", "", "Path to JSON configuration file")
	flag.StringVar(&configFile, "c", "", "Path to JSON configuration file (alias for -config)")
	flag.StringVar(&Cfg.ServerAddress, "a", "localhost:8080", "HTTP server address")
	flag.StringVar(&Cfg.BaseURL, "b", "http://localhost:8080", "Base URL for shortened links")
	flag.StringVar(&Cfg.FileStorageURL, "f", "", "Path to data file")
	flag.StringVar(&Cfg.DatabaseDSN, "d", "", "Database connection path")
	flag.StringVar(&Cfg.JWTSecret, "j", "", "JWT secret")
	flag.IntVar(&Cfg.JWTExpiration, "e", 0, "JWT lifetime in seconds")
	flag.StringVar(&Cfg.AuditFile, "audit-file", "", "Path to audit log file")
	flag.StringVar(&Cfg.AuditURL, "audit-url", "", "URL of remote audit server")
	flag.BoolVar(&Cfg.EnableHTTPS, "s", false, "Enable HTTPS (true/false)")
	flag.StringVar(&Cfg.TLSCertFile, "tls-cert", "", "Path to TLS certificate file")
	flag.StringVar(&Cfg.TLSKeyFile, "k", "", "Path to TLS private key file")
}

func loadConfigFile(path string) (Config, error) {
	var fileCfg Config
	data, err := os.ReadFile(path)
	if err != nil {
		return fileCfg, err
	}
	err = json.Unmarshal(data, &fileCfg)
	return fileCfg, err
}

func Load() {
	if !flag.Parsed() {
		flag.Parse()
	}

	setFlags := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) {
		setFlags[f.Name] = true
	})

	cfgPath := configFile
	if cfgPath == "" {
		cfgPath = os.Getenv("CONFIG")
	}

	var fileCfg Config
	if cfgPath != "" {
		if loaded, err := loadConfigFile(cfgPath); err == nil {
			fileCfg = loaded
		}
	}

	if !setFlags["a"] && os.Getenv("SERVER_ADDRESS") == "" && fileCfg.ServerAddress != "" {
		Cfg.ServerAddress = fileCfg.ServerAddress
	}

	if !setFlags["b"] && os.Getenv("BASE_URL") == "" && fileCfg.BaseURL != "" {
		Cfg.BaseURL = fileCfg.BaseURL
	}

	if !setFlags["f"] && os.Getenv("FILE_STORAGE_PATH") == "" && fileCfg.FileStorageURL != "" {
		Cfg.FileStorageURL = fileCfg.FileStorageURL
	}

	if !setFlags["d"] && os.Getenv("DATABASE_DSN") == "" && fileCfg.DatabaseDSN != "" {
		Cfg.DatabaseDSN = fileCfg.DatabaseDSN
	}

	if !setFlags["j"] && os.Getenv("JWT_SECRET") == "" && fileCfg.JWTSecret != "" {
		Cfg.JWTSecret = fileCfg.JWTSecret
	}

	if !setFlags["e"] && os.Getenv("JWT_EXT") == "" && fileCfg.JWTExpiration != 0 {
		Cfg.JWTExpiration = fileCfg.JWTExpiration
	}

	if !setFlags["audit-file"] && os.Getenv("AUDIT_FILE") == "" && fileCfg.AuditFile != "" {
		Cfg.AuditFile = fileCfg.AuditFile
	}

	if !setFlags["audit-url"] && os.Getenv("AUDIT_URL") == "" && fileCfg.AuditURL != "" {
		Cfg.AuditURL = fileCfg.AuditURL
	}

	if !setFlags["s"] && os.Getenv("ENABLE_HTTPS") == "" && fileCfg.EnableHTTPS {
		Cfg.EnableHTTPS = fileCfg.EnableHTTPS
	}

	if !setFlags["tls-cert"] && os.Getenv("TLS_CERT_FILE") == "" && fileCfg.TLSCertFile != "" {
		Cfg.TLSCertFile = fileCfg.TLSCertFile
	}

	if !setFlags["k"] && os.Getenv("TLS_KEY_FILE") == "" && fileCfg.TLSKeyFile != "" {
		Cfg.TLSKeyFile = fileCfg.TLSKeyFile
	}

	_ = env.Parse(&Cfg)
}
