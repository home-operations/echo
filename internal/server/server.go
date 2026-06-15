// Package server wires echo's HTTP, HTTPS, and metrics listeners together and
// manages their lifecycle.
package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/home-operations/echo/internal/config"
	"github.com/home-operations/echo/internal/echo"
	"golang.org/x/sync/errgroup"
)

// Connection timeouts applied to every listener, bounding slow-client
// (Slowloris) and idle keep-alive resource exhaustion. echo's requests and
// responses are small and fast, so these limits are generous.
const (
	readHeaderTimeout = 10 * time.Second
	readTimeout       = 30 * time.Second
	writeTimeout      = 30 * time.Second
	idleTimeout       = 120 * time.Second
)

// newHTTPServer builds an http.Server with echo's standard connection timeouts.
func newHTTPServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}
}

// Server owns the configured listeners and shared request state.
type Server struct {
	cfg      *config.Config
	log      *slog.Logger
	hostname string
	k8s      *echo.KubernetesInfo
}

// New constructs a Server, resolving the OS hostname once so request handlers
// don't re-syscall on every request.
func New(cfg *config.Config, log *slog.Logger) *Server {
	host, err := os.Hostname()
	if err != nil {
		host = "unknown"
		log.Warn("could not resolve hostname", "error", err)
	}

	var k8s *echo.KubernetesInfo
	if cfg.Kubernetes {
		k8s = &echo.KubernetesInfo{
			PodName:      cfg.PodName,
			PodNamespace: cfg.PodNamespace,
			PodIP:        cfg.PodIP,
			NodeName:     cfg.NodeName,
		}
	}

	return &Server{cfg: cfg, log: log, hostname: host, k8s: k8s}
}

// Run starts every enabled listener and blocks until ctx is cancelled or a
// listener fails, then drains them within the configured shutdown timeout.
func (s *Server) Run(ctx context.Context) error {
	g, gctx := errgroup.WithContext(ctx)

	handler := s.handler()

	httpSrv := newHTTPServer(fmt.Sprintf(":%d", s.cfg.HTTPPort), handler)
	g.Go(func() error { return serve(httpSrv, "http", s.log) })

	var metricsSrv *http.Server
	if s.cfg.MetricsEnabled {
		metricsSrv = newHTTPServer(s.cfg.MetricsAddr, metricsHandler())
		g.Go(func() error { return serve(metricsSrv, "metrics", s.log) })
	}

	g.Go(func() error {
		<-gctx.Done()
		s.log.Info("shutting down")
		sctx, cancel := context.WithTimeout(context.Background(), s.cfg.ShutdownTimeout)
		defer cancel()
		shutdown(sctx, httpSrv)
		shutdown(sctx, metricsSrv)
		return nil
	})

	if err := g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		return err
	}
	return nil
}

func serve(srv *http.Server, name string, log *slog.Logger) error {
	log.Info("listening", "server", name, "addr", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("%s server: %w", name, err)
	}
	return nil
}

func shutdown(ctx context.Context, srv *http.Server) {
	if srv == nil {
		return
	}
	_ = srv.Shutdown(ctx)
}
