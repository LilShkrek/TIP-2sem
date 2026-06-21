package config

import (
	"net"
	"os"
)

const (
	defaultHost     = "0.0.0.0"
	defaultPort     = "8082"
	defaultLogLevel = "info"
)

type Config struct {
	Host        string
	Port        string
	AuthBaseURL string
	LogLevel    string
}

func Load() Config {
	return Config{
		Host:        envOrDefault("TASKS_HOST", defaultHost),
		Port:        envOrDefault("TASKS_PORT", defaultPort),
		AuthBaseURL: os.Getenv("AUTH_BASE_URL"),
		LogLevel:    envOrDefault("LOG_LEVEL", defaultLogLevel),
	}
}

func (c Config) Address() string {
	return net.JoinHostPort(c.Host, c.Port)
}

func envOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	return value
}
