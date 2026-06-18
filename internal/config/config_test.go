package config

import (
	"log/slog"
	"testing"
	"time"
)

func TestLoadDefaults(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if cfg.HTTPPort != 8080 {
		t.Errorf("HTTPPort = %d, want 8080", cfg.HTTPPort)
	}
	// Monitoring (metrics + the /healthz probe) listens on 8081, separate from
	// the public echo port.
	if cfg.MetricsAddr != ":8081" {
		t.Errorf("MetricsAddr = %q, want :8081", cfg.MetricsAddr)
	}
	if !cfg.EchoBackToClient {
		t.Error("EchoBackToClient = false, want true")
	}
	if cfg.LogLevel != slog.LevelInfo {
		t.Errorf("LogLevel = %v, want info", cfg.LogLevel)
	}
	if cfg.MaxBodyBytes != 1<<20 {
		t.Errorf("MaxBodyBytes = %d, want %d", cfg.MaxBodyBytes, 1<<20)
	}
	if cfg.ShutdownTimeout != 15*time.Second {
		t.Errorf("ShutdownTimeout = %v, want 15s", cfg.ShutdownTimeout)
	}
	if len(cfg.TrustedProxies) != 0 {
		t.Errorf("TrustedProxies = %v, want empty", cfg.TrustedProxies)
	}
}

func TestLoadOverrides(t *testing.T) {
	t.Setenv("ECHO_HTTP_PORT", "9000")
	t.Setenv("ECHO_LOG_LEVEL", "debug")
	t.Setenv("ECHO_LOG_FORMAT", "text")
	t.Setenv("ECHO_TRUSTED_PROXIES", "10.0.0.0/8, 192.168.0.0/16")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if cfg.HTTPPort != 9000 {
		t.Errorf("HTTPPort = %d, want 9000", cfg.HTTPPort)
	}
	if cfg.LogLevel != slog.LevelDebug {
		t.Errorf("LogLevel = %v, want debug", cfg.LogLevel)
	}
	if cfg.LogFormat != "text" {
		t.Errorf("LogFormat = %q, want text", cfg.LogFormat)
	}
	if len(cfg.TrustedProxies) != 2 {
		t.Fatalf("TrustedProxies = %v, want 2 entries", cfg.TrustedProxies)
	}
}

func TestLoadInvalid(t *testing.T) {
	tests := map[string]map[string]string{
		"bad http port":     {"ECHO_HTTP_PORT": "70000"},
		"zero http port":    {"ECHO_HTTP_PORT": "0"},
		"bad log level":     {"ECHO_LOG_LEVEL": "loud"},
		"bad log format":    {"ECHO_LOG_FORMAT": "xml"},
		"bad trusted proxy": {"ECHO_TRUSTED_PROXIES": "not-a-cidr"},
		"negative max body": {"ECHO_MAX_BODY_BYTES": "-1"},
	}

	for name, env := range tests {
		t.Run(name, func(t *testing.T) {
			for k, v := range env {
				t.Setenv(k, v)
			}
			if _, err := Load(); err == nil {
				t.Fatalf("Load() = nil error, want error for %s", name)
			}
		})
	}
}
