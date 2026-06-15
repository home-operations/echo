// Command echo runs an HTTP server that echoes each request back to the caller
// as JSON.
package main

import (
	"context"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"

	"github.com/KimMachineGun/automemlimit/memlimit"
	"github.com/home-operations/echo/internal/config"
	"github.com/home-operations/echo/internal/server"
)

// Build metadata, stamped via -ldflags at release time.
var (
	version = "dev"
	commit  = "none"
)

func main() {
	if err := run(); err != nil {
		slog.Error("fatal error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	setMemLimit()

	// The first signal triggers a graceful shutdown; re-arm the default handler so
	// a second signal force-quits instead of being swallowed during a slow drain.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	go func() {
		<-ctx.Done()
		stop()
	}()

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	logger := newLogger(cfg)
	slog.SetDefault(logger)
	logger.Info("starting echo",
		"version", version,
		"commit", commit,
		"http_port", cfg.HTTPPort,
		"gomaxprocs", runtime.GOMAXPROCS(0),
	)

	return server.New(cfg, logger).Run(ctx)
}

// newLogger builds the root logger: JSON by default (the container-friendly
// format), text on request, always to stdout.
func newLogger(cfg *config.Config) *slog.Logger {
	opts := &slog.HandlerOptions{Level: cfg.LogLevel}
	if strings.EqualFold(cfg.LogFormat, "text") {
		return slog.New(slog.NewTextHandler(os.Stdout, opts))
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, opts))
}

// setMemLimit caps the Go heap (GOMEMLIMIT) at 90% of the cgroup memory limit
// when one is set, so the GC reclaims before the container is OOM-killed. It is
// a silent no-op outside a memory-limited cgroup.
func setMemLimit() {
	_, _ = memlimit.SetGoMemLimitWithOpts(
		memlimit.WithRatio(0.9),
		memlimit.WithProvider(memlimit.FromCgroup),
		memlimit.WithLogger(slog.New(slog.NewTextHandler(io.Discard, nil))),
	)
}
