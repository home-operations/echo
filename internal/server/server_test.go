package server

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/home-operations/echo/internal/config"
)

func newTestServer(t *testing.T, cfg *config.Config) *Server {
	t.Helper()
	return New(cfg, slog.New(slog.NewTextHandler(io.Discard, nil)))
}

func baseConfig() *config.Config {
	return &config.Config{
		HTTPPort:           8080,
		MaxBodyBytes:       1 << 20,
		EchoBackToClient:   true,
		DisableRequestLogs: true,
		LogFormat:          "json",
		WSEnabled:          true,
	}
}

func TestEchoHandlerGET(t *testing.T) {
	s := newTestServer(t, baseConfig())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/foo?x=1", nil)

	s.handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
	if got := rec.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Errorf("X-Content-Type-Options = %q, want nosniff", got)
	}

	var doc map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &doc); err != nil {
		t.Fatalf("response is not JSON: %v", err)
	}
	if doc["method"] != http.MethodGet {
		t.Errorf("method = %v, want GET", doc["method"])
	}
	if doc["path"] != "/foo" {
		t.Errorf("path = %v, want /foo", doc["path"])
	}
}

func TestEchoHandlerPOSTBody(t *testing.T) {
	s := newTestServer(t, baseConfig())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"hello":"world"}`))
	req.Header.Set("Content-Type", "application/json")

	s.handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var doc map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &doc); err != nil {
		t.Fatalf("response is not JSON: %v", err)
	}
	if doc["body"] != `{"hello":"world"}` {
		t.Errorf("body = %v, want the raw JSON string", doc["body"])
	}
	if _, ok := doc["json"]; !ok {
		t.Error("response missing parsed json field for a JSON body")
	}
}

func TestEchoBackDisabled(t *testing.T) {
	cfg := baseConfig()
	cfg.EchoBackToClient = false
	s := newTestServer(t, cfg)
	rec := httptest.NewRecorder()

	s.handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", rec.Code)
	}
	if rec.Body.Len() != 0 {
		t.Errorf("body = %q, want empty", rec.Body.String())
	}
}

func TestHealthEndpoint(t *testing.T) {
	s := newTestServer(t, baseConfig())
	rec := httptest.NewRecorder()

	s.handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var doc map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &doc); err != nil {
		t.Fatalf("health is not JSON: %v", err)
	}
	if doc["status"] != "ok" {
		t.Errorf("status = %q, want ok", doc["status"])
	}
}

func TestAccessLogLevels(t *testing.T) {
	run := func(level slog.Level, path string) string {
		var buf bytes.Buffer
		cfg := baseConfig()
		cfg.DisableRequestLogs = false
		s := New(cfg, slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: level})))
		s.handler().ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, path, nil))
		return buf.String()
	}

	// A normal request logs at info.
	if out := run(slog.LevelInfo, "/foo"); !strings.Contains(out, `"level":"INFO"`) || !strings.Contains(out, `"path":"/foo"`) {
		t.Errorf("a normal request should log at info, got: %s", out)
	}

	// Probe/scrape paths drop to debug — silent when the level is info.
	for _, p := range []string{"/healthz", "/ping", "/metrics"} {
		if out := run(slog.LevelInfo, p); strings.Contains(out, `"msg":"request"`) {
			t.Errorf("%s should be debug-level (silent at info), got: %s", p, out)
		}
	}

	// At debug, the probe paths are logged (at debug level).
	if out := run(slog.LevelDebug, "/healthz"); !strings.Contains(out, `"level":"DEBUG"`) || !strings.Contains(out, `"path":"/healthz"`) {
		t.Errorf("/healthz should log at debug, got: %s", out)
	}
}

func TestMetricsHandler(t *testing.T) {
	// Drive a request through the chain so the counter has a series, then scrape.
	s := newTestServer(t, baseConfig())
	s.handler().ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	rec := httptest.NewRecorder()
	metricsHandler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "echo_http_requests_total") {
		t.Error("metrics output missing echo_http_requests_total")
	}
}

// The monitoring listener also answers the health/readiness probe, so the
// chart's probes can target the metrics port instead of the public echo port.
func TestMetricsHandlerServesHealth(t *testing.T) {
	rec := httptest.NewRecorder()
	metricsHandler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var doc map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &doc); err != nil {
		t.Fatalf("health is not JSON: %v", err)
	}
	if doc["status"] != "ok" {
		t.Errorf("status = %q, want ok", doc["status"])
	}
}

func TestMethodLabelBoundsCardinality(t *testing.T) {
	if got := methodLabel(http.MethodGet); got != http.MethodGet {
		t.Errorf("methodLabel(GET) = %q, want GET", got)
	}
	// An arbitrary client-supplied method must collapse to a fixed label.
	if got := methodLabel("WEIRD-METHOD"); got != "other" {
		t.Errorf("methodLabel(WEIRD-METHOD) = %q, want other", got)
	}
}

func TestNewHTTPServerSetsTimeouts(t *testing.T) {
	srv := newHTTPServer(":0", nil)
	if srv.ReadHeaderTimeout == 0 || srv.ReadTimeout == 0 || srv.WriteTimeout == 0 || srv.IdleTimeout == 0 {
		t.Errorf("a connection timeout is unset: %+v", srv)
	}
}

func TestWebSocketEcho(t *testing.T) {
	srv := httptest.NewServer(newTestServer(t, baseConfig()).handler())
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(srv.URL, "http")+"/ws", nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer func() { _ = conn.CloseNow() }()

	if err := conn.Write(ctx, websocket.MessageText, []byte("ping")); err != nil {
		t.Fatalf("write: %v", err)
	}
	typ, data, err := conn.Read(ctx)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if typ != websocket.MessageText || string(data) != "ping" {
		t.Errorf("echo = (%v, %q), want (text, ping)", typ, data)
	}
	_ = conn.Close(websocket.StatusNormalClosure, "")
}

func TestWebSocketDisabledFallsThroughToEcho(t *testing.T) {
	cfg := baseConfig()
	cfg.WSEnabled = false
	srv := httptest.NewServer(newTestServer(t, cfg).handler())
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(srv.URL, "http")+"/ws", nil)
	if err == nil {
		_ = conn.CloseNow()
		t.Fatal("dial succeeded, but with WS disabled /ws should respond as a plain HTTP echo")
	}
}

func TestEchoKubernetesBlock(t *testing.T) {
	cfg := baseConfig()
	cfg.Kubernetes = true
	cfg.PodName = "echo-xyz"
	cfg.NodeName = "node-7"
	rec := httptest.NewRecorder()
	newTestServer(t, cfg).handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	var doc map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &doc); err != nil {
		t.Fatalf("response is not JSON: %v", err)
	}
	k8s, ok := doc["kubernetes"].(map[string]any)
	if !ok {
		t.Fatalf("kubernetes block missing; got %v", doc["kubernetes"])
	}
	if k8s["podName"] != "echo-xyz" || k8s["nodeName"] != "node-7" {
		t.Errorf("kubernetes = %v", k8s)
	}
}

func TestEchoOmitsKubernetesByDefault(t *testing.T) {
	rec := httptest.NewRecorder()
	newTestServer(t, baseConfig()).handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	var doc map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &doc); err != nil {
		t.Fatalf("response is not JSON: %v", err)
	}
	if _, ok := doc["kubernetes"]; ok {
		t.Error("kubernetes block present by default; want omitted")
	}
}
