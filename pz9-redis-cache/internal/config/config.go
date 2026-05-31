package config

import (
	"os"
	"time"
)

type Config struct {
	HTTPAddr          string
	RedisAddr         string
	RedisPassword     string
	CacheTTL          time.Duration
	CacheTTLJitter    time.Duration
	RedisDialTimeout  time.Duration
	RedisReadTimeout  time.Duration
	RedisWriteTimeout time.Duration
}

func New() Config {
	return Config{
		HTTPAddr:          envString("HTTP_ADDR", ":8082"),
		RedisAddr:         envString("REDIS_ADDR", "localhost:6379"),
		RedisPassword:     envString("REDIS_PASSWORD", ""),
		CacheTTL:          120 * time.Second,
		CacheTTLJitter:    30 * time.Second,
		RedisDialTimeout:  2 * time.Second,
		RedisReadTimeout:  2 * time.Second,
		RedisWriteTimeout: 2 * time.Second,
	}
}

func envString(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
