// Package config loads echo's runtime configuration from ECHO_*-prefixed
// environment variables.
package config

import (
	"fmt"
	"log/slog"
	"net/netip"
	"os"
	"strings"
	"time"

	"github.com/caarlos0/env/v11"
)

// Config is the fully resolved runtime configuration. The simple fields are
// populated by env.Parse; LogLevel and TrustedProxies are derived in Load so
// their parsing fails fast with a clear message.
type Config struct {
	HTTPPort int `env:"ECHO_HTTP_PORT" envDefault:"8080"`

	// MetricsEnabled exposes Prometheus metrics at /metrics on MetricsAddr. The
	// /healthz probe endpoint is served on this address too, so probes target the
	// monitoring port rather than the public echo (HTTP) port.
	MetricsEnabled bool   `env:"ECHO_METRICS_ENABLED" envDefault:"true"`
	MetricsAddr    string `env:"ECHO_METRICS_ADDR" envDefault:":8081"`

	// LogFormat selects the slog handler: "json" (default) or "text".
	LogFormat string `env:"ECHO_LOG_FORMAT" envDefault:"json"`
	// DisableRequestLogs silences the per-request access log; the echo response
	// to the client is unaffected.
	DisableRequestLogs bool `env:"ECHO_DISABLE_REQUEST_LOGS" envDefault:"false"`

	// EchoBackToClient controls whether the JSON document is returned to the
	// client. When false the server still observes the request but replies 204.
	EchoBackToClient bool `env:"ECHO_BACK_TO_CLIENT" envDefault:"true"`
	// MaxBodyBytes caps how much request body is read and echoed; larger bodies
	// are clipped and flagged with bodyTruncated in the response.
	MaxBodyBytes int64 `env:"ECHO_MAX_BODY_BYTES" envDefault:"1048576"`

	// CommandsEnabled lets callers shape the response — status code, artificial
	// delay, and extra response headers — via echo-* query parameters and
	// X-Echo-* headers, turning the reflector into a test target for ingress,
	// proxies, and clients. On by default (echo is a test tool); set false to
	// make echo a pure reflector. Pretty-printing is unaffected by this gate.
	CommandsEnabled bool `env:"ECHO_COMMANDS_ENABLED" envDefault:"true"`
	// MaxDelay caps the artificial response delay a caller may request via
	// echo-delay; larger requests are clamped to it. Bounds a caller's ability
	// to tie up a connection; keep it below the server write timeout.
	MaxDelay time.Duration `env:"ECHO_MAX_DELAY" envDefault:"10s"`
	// PrettyPrint indents the JSON response by default. Callers can override per
	// request with echo-pretty-print; it is always available, even when commands
	// are disabled.
	PrettyPrint bool `env:"ECHO_PRETTY_PRINT" envDefault:"false"`

	// WSEnabled serves a WebSocket echo at /ws (non-upgrade requests to that path
	// fall through to the normal HTTP echo).
	WSEnabled bool `env:"ECHO_WS_ENABLED" envDefault:"true"`
	// WSAllowedOrigins restricts which Origins may open a WebSocket (host
	// patterns). Empty allows any origin — safe here because the endpoint only
	// echoes the caller's own frames.
	WSAllowedOrigins []string `env:"ECHO_WS_ALLOWED_ORIGINS" envSeparator:","`

	// Kubernetes adds a kubernetes object (pod/node identity from the Downward
	// API env below) to the response. Off by default.
	Kubernetes   bool   `env:"ECHO_KUBERNETES" envDefault:"false"`
	PodName      string `env:"ECHO_POD_NAME"`
	PodNamespace string `env:"ECHO_POD_NAMESPACE"`
	PodIP        string `env:"ECHO_POD_IP"`
	NodeName     string `env:"ECHO_NODE_NAME"`

	// ShutdownTimeout bounds the graceful drain on SIGINT/SIGTERM.
	ShutdownTimeout time.Duration `env:"ECHO_SHUTDOWN_TIMEOUT" envDefault:"15s"`

	// LogLevel is parsed from ECHO_LOG_LEVEL (debug|info|warn|error) in Load.
	LogLevel slog.Level `env:"-"`
	// TrustedProxies are CIDRs whose X-Forwarded-For header is trusted when
	// resolving the client IP. Parsed from ECHO_TRUSTED_PROXIES (comma-separated)
	// in Load; empty means the immediate peer is treated as the client.
	TrustedProxies []netip.Prefix `env:"-"`
}

// Load reads the configuration from the environment, applies defaults, derives
// the computed fields, and validates the result.
func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}

	level := os.Getenv("ECHO_LOG_LEVEL")
	if level == "" {
		level = "info"
	}
	if err := cfg.LogLevel.UnmarshalText([]byte(strings.TrimSpace(level))); err != nil {
		return nil, fmt.Errorf("config: invalid ECHO_LOG_LEVEL %q: %w", level, err)
	}

	proxies, err := parsePrefixes(os.Getenv("ECHO_TRUSTED_PROXIES"))
	if err != nil {
		return nil, err
	}
	cfg.TrustedProxies = proxies

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) validate() error {
	if err := validatePort(c.HTTPPort, "ECHO_HTTP_PORT"); err != nil {
		return err
	}
	switch strings.ToLower(c.LogFormat) {
	case "json", "text":
	default:
		return fmt.Errorf("config: invalid ECHO_LOG_FORMAT %q (want json or text)", c.LogFormat)
	}
	if c.MaxBodyBytes < 0 {
		return fmt.Errorf("config: ECHO_MAX_BODY_BYTES must be >= 0, got %d", c.MaxBodyBytes)
	}
	if c.MaxDelay < 0 {
		return fmt.Errorf("config: ECHO_MAX_DELAY must be >= 0, got %s", c.MaxDelay)
	}
	return nil
}

func parsePrefixes(raw string) ([]netip.Prefix, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	parts := strings.Split(raw, ",")
	out := make([]netip.Prefix, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		prefix, err := netip.ParsePrefix(p)
		if err != nil {
			return nil, fmt.Errorf("config: invalid ECHO_TRUSTED_PROXIES entry %q: %w", p, err)
		}
		out = append(out, prefix)
	}
	return out, nil
}

func validatePort(p int, name string) error {
	if p < 1 || p > 65535 {
		return fmt.Errorf("config: %s must be between 1 and 65535, got %d", name, p)
	}
	return nil
}
