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
	if cfg.HTTPSEnabled() {
		t.Errorf("HTTPSEnabled() = true with HTTPSPort %d, want disabled by default", cfg.HTTPSPort)
	}
	// Metrics listen on 8081, separate from the public echo port (which also
	// serves the /healthz probe).
	if cfg.MetricsPort != 8081 {
		t.Errorf("MetricsPort = %d, want 8081", cfg.MetricsPort)
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
	if !cfg.CommandsEnabled {
		t.Error("CommandsEnabled = false, want true")
	}
	if cfg.MaxDelay != 10*time.Second {
		t.Errorf("MaxDelay = %v, want 10s", cfg.MaxDelay)
	}
	if cfg.WSIdleTimeout != 5*time.Minute {
		t.Errorf("WSIdleTimeout = %v, want 5m", cfg.WSIdleTimeout)
	}
	if cfg.PrettyPrint {
		t.Error("PrettyPrint = true, want false")
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

func TestLoadHTTPS(t *testing.T) {
	t.Setenv("ECHO_HTTPS_PORT", "8443")
	t.Setenv("ECHO_HTTPS_CERT", "/tls/tls.crt")
	t.Setenv("ECHO_HTTPS_KEY", "/tls/tls.key")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if !cfg.HTTPSEnabled() {
		t.Fatal("HTTPSEnabled() = false, want true")
	}
	if cfg.HTTPSPort != 8443 {
		t.Errorf("HTTPSPort = %d, want 8443", cfg.HTTPSPort)
	}
	if cfg.HTTPSCert != "/tls/tls.crt" || cfg.HTTPSKey != "/tls/tls.key" {
		t.Errorf("HTTPSCert/HTTPSKey = %q/%q, want /tls/tls.crt and /tls/tls.key", cfg.HTTPSCert, cfg.HTTPSKey)
	}
}

func TestLoadInvalid(t *testing.T) {
	tests := map[string]map[string]string{
		"bad http port":                         {"ECHO_HTTP_PORT": "70000"},
		"zero http port":                        {"ECHO_HTTP_PORT": "0"},
		"bad log level":                         {"ECHO_LOG_LEVEL": "loud"},
		"bad log format":                        {"ECHO_LOG_FORMAT": "xml"},
		"bad trusted proxy":                     {"ECHO_TRUSTED_PROXIES": "not-a-cidr"},
		"negative max body":                     {"ECHO_MAX_BODY_BYTES": "-1"},
		"negative max delay":                    {"ECHO_MAX_DELAY": "-1s"},
		"bad max delay":                         {"ECHO_MAX_DELAY": "soon"},
		"negative ws idle":                      {"ECHO_WS_IDLE_TIMEOUT": "-1s"},
		"metrics port collides with http port":  {"ECHO_METRICS_PORT": "8080"},
		"bad https port":                        {"ECHO_HTTPS_PORT": "70000", "ECHO_HTTPS_CERT": "c", "ECHO_HTTPS_KEY": "k"},
		"https port without cert and key":       {"ECHO_HTTPS_PORT": "8443"},
		"https port without key":                {"ECHO_HTTPS_PORT": "8443", "ECHO_HTTPS_CERT": "c"},
		"https cert without port":               {"ECHO_HTTPS_CERT": "c", "ECHO_HTTPS_KEY": "k"},
		"https port collides with http port":    {"ECHO_HTTPS_PORT": "8080", "ECHO_HTTPS_CERT": "c", "ECHO_HTTPS_KEY": "k"},
		"https port collides with metrics port": {"ECHO_HTTPS_PORT": "8081", "ECHO_HTTPS_CERT": "c", "ECHO_HTTPS_KEY": "k"},
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
