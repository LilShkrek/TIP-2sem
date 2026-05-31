package config

import "os"

type Config struct {
	Addr     string
	CertFile string
	KeyFile  string
	DSN      string
}

func New() Config {
	return Config{
		Addr:     valueOrDefault("SERVER_ADDR", ":8443"),
		CertFile: valueOrDefault("TLS_CERT_FILE", "certs/server.crt"),
		KeyFile:  valueOrDefault("TLS_KEY_FILE", "certs/server.key"),
		DSN:      valueOrDefault("POSTGRES_DSN", "postgres://postgres:postgres@localhost:5432/study_security?sslmode=disable"),
	}
}

func valueOrDefault(name, fallback string) string {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}
	return value
}
