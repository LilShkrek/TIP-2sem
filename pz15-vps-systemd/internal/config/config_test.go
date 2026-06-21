package config

import "testing"

func TestLoadFromEnvReadsConfig(t *testing.T) {
	env := map[string]string{
		"TASKS_HOST": "0.0.0.0",
		"TASKS_PORT": "9090",
		"LOG_LEVEL":  "debug",
	}

	cfg, err := LoadFromEnv(func(key string) string {
		return env[key]
	})
	if err != nil {
		t.Fatalf("LoadFromEnv returned error: %v", err)
	}

	if cfg.Host != "0.0.0.0" {
		t.Fatalf("Host = %q, want %q", cfg.Host, "0.0.0.0")
	}
	if cfg.Port != "9090" {
		t.Fatalf("Port = %q, want %q", cfg.Port, "9090")
	}
	if cfg.LogLevel != "debug" {
		t.Fatalf("LogLevel = %q, want %q", cfg.LogLevel, "debug")
	}
	if cfg.Address() != "0.0.0.0:9090" {
		t.Fatalf("Address() = %q, want %q", cfg.Address(), "0.0.0.0:9090")
	}
}

func TestLoadFromEnvUsesDefaultPort(t *testing.T) {
	cfg, err := LoadFromEnv(func(string) string {
		return ""
	})
	if err != nil {
		t.Fatalf("LoadFromEnv returned error: %v", err)
	}

	if cfg.Port != DefaultPort {
		t.Fatalf("Port = %q, want default %q", cfg.Port, DefaultPort)
	}
}
