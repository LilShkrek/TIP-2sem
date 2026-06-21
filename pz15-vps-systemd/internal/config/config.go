package config

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultHost            = "127.0.0.1"
	DefaultPort            = "8082"
	DefaultLogLevel        = "info"
	DefaultShutdownTimeout = 5 * time.Second
)

type Config struct {
	Host            string
	Port            string
	LogLevel        string
	ShutdownTimeout time.Duration
}

func Load() (Config, error) {
	return LoadFromEnv(os.Getenv)
}

func LoadFromEnv(getenv func(string) string) (Config, error) {
	cfg := Config{
		Host:            envOrDefault(getenv("TASKS_HOST"), DefaultHost),
		Port:            envOrDefault(getenv("TASKS_PORT"), DefaultPort),
		LogLevel:        envOrDefault(getenv("LOG_LEVEL"), DefaultLogLevel),
		ShutdownTimeout: DefaultShutdownTimeout,
	}

	if err := validatePort(cfg.Port); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c Config) Address() string {
	return net.JoinHostPort(c.Host, c.Port)
}

func envOrDefault(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

func validatePort(port string) error {
	number, err := strconv.Atoi(port)
	if err != nil || number < 1 || number > 65535 {
		return fmt.Errorf("invalid TASKS_PORT %q: expected TCP port from 1 to 65535", port)
	}

	return nil
}
