package config

import (
	"os"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	unsetEnv(t, "TASKS_HOST", "TASKS_PORT", "AUTH_BASE_URL", "LOG_LEVEL")

	cfg := Load()

	if cfg.Host != "0.0.0.0" {
		t.Fatalf("Host = %q, want %q", cfg.Host, "0.0.0.0")
	}
	if cfg.Port != "8082" {
		t.Fatalf("Port = %q, want %q", cfg.Port, "8082")
	}
	if cfg.AuthBaseURL != "" {
		t.Fatalf("AuthBaseURL = %q, want empty value", cfg.AuthBaseURL)
	}
	if cfg.LogLevel != "info" {
		t.Fatalf("LogLevel = %q, want %q", cfg.LogLevel, "info")
	}
	if cfg.Address() != "0.0.0.0:8082" {
		t.Fatalf("Address() = %q, want %q", cfg.Address(), "0.0.0.0:8082")
	}
}

func TestLoadFromEnvironment(t *testing.T) {
	t.Setenv("TASKS_HOST", "127.0.0.1")
	t.Setenv("TASKS_PORT", "9090")
	t.Setenv("AUTH_BASE_URL", "http://auth.local:8081")
	t.Setenv("LOG_LEVEL", "debug")

	cfg := Load()

	if cfg.Host != "127.0.0.1" {
		t.Fatalf("Host = %q, want %q", cfg.Host, "127.0.0.1")
	}
	if cfg.Port != "9090" {
		t.Fatalf("Port = %q, want %q", cfg.Port, "9090")
	}
	if cfg.AuthBaseURL != "http://auth.local:8081" {
		t.Fatalf("AuthBaseURL = %q, want %q", cfg.AuthBaseURL, "http://auth.local:8081")
	}
	if cfg.LogLevel != "debug" {
		t.Fatalf("LogLevel = %q, want %q", cfg.LogLevel, "debug")
	}
	if cfg.Address() != "127.0.0.1:9090" {
		t.Fatalf("Address() = %q, want %q", cfg.Address(), "127.0.0.1:9090")
	}
}

func unsetEnv(t *testing.T, keys ...string) {
	t.Helper()

	previous := make(map[string]string, len(keys))
	present := make(map[string]bool, len(keys))

	for _, key := range keys {
		value, ok := os.LookupEnv(key)
		previous[key] = value
		present[key] = ok
		if err := os.Unsetenv(key); err != nil {
			t.Fatalf("unset %s: %v", key, err)
		}
	}

	t.Cleanup(func() {
		for _, key := range keys {
			if present[key] {
				if err := os.Setenv(key, previous[key]); err != nil {
					t.Fatalf("restore %s: %v", key, err)
				}
				continue
			}

			if err := os.Unsetenv(key); err != nil {
				t.Fatalf("restore unset %s: %v", key, err)
			}
		}
	})
}
