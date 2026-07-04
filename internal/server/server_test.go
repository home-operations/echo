package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
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
		CommandsEnabled:    true,
		MaxDelay:           10 * time.Second,
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

	// Probe paths drop to debug — silent when the level is info.
	for _, p := range []string{"/healthz", "/readyz"} {
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

// Health lives on the main HTTP listener (the chart's probes target the http
// port), NOT the optional metrics listener — so disabling metrics can never
// break the probes.
func TestHealthOnMainPortNotMetricsPort(t *testing.T) {
	rec := httptest.NewRecorder()
	newTestServer(t, baseConfig()).handler().
		ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("main handler /healthz status = %d, want 200", rec.Code)
	}
	var doc map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &doc); err != nil {
		t.Fatalf("health is not JSON: %v", err)
	}
	if doc["status"] != "ok" {
		t.Errorf("status = %q, want ok", doc["status"])
	}

	// /readyz aliases /healthz (the pair standard) on the same main listener.
	rec = httptest.NewRecorder()
	newTestServer(t, baseConfig()).handler().
		ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if rec.Code != http.StatusOK {
		t.Errorf("main handler /readyz status = %d, want 200", rec.Code)
	}

	// The metrics handler is metrics-only.
	rec = httptest.NewRecorder()
	metricsHandler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if rec.Code != http.StatusNotFound {
		t.Errorf("metrics handler /healthz status = %d, want 404 (metrics-only listener)", rec.Code)
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

func TestWebSocketIdleTimeout(t *testing.T) {
	cfg := baseConfig()
	cfg.WSIdleTimeout = 100 * time.Millisecond
	srv := httptest.NewServer(newTestServer(t, cfg).handler())
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(srv.URL, "http")+"/ws", nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer func() { _ = conn.CloseNow() }()

	// Send nothing: the server must close the connection once the idle window
	// lapses, so this read returns an error well before the test context expires.
	if _, _, err := conn.Read(ctx); err == nil {
		t.Fatal("read succeeded, want connection closed by server idle timeout")
	}
	if ctx.Err() != nil {
		t.Fatal("test context expired first: server did not enforce the idle timeout")
	}
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

func TestEchoCommandStatus(t *testing.T) {
	s := newTestServer(t, baseConfig())

	t.Run("via query", func(t *testing.T) {
		rec := httptest.NewRecorder()
		s.handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/?echo-code=503", nil))
		if rec.Code != http.StatusServiceUnavailable {
			t.Fatalf("status = %d, want 503", rec.Code)
		}
		var doc map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &doc); err != nil {
			t.Fatalf("response is not JSON: %v", err)
		}
		applied, ok := doc["applied"].(map[string]any)
		if !ok || applied["status"] != float64(503) {
			t.Errorf("applied = %v, want status 503", doc["applied"])
		}
	})

	t.Run("via header", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Echo-Code", "418")
		s.handler().ServeHTTP(rec, req)
		if rec.Code != http.StatusTeapot {
			t.Errorf("status = %d, want 418", rec.Code)
		}
	})

	t.Run("query wins over header", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/?echo-code=503", nil)
		req.Header.Set("X-Echo-Code", "418")
		s.handler().ServeHTTP(rec, req)
		if rec.Code != http.StatusServiceUnavailable {
			t.Errorf("status = %d, want 503 (query wins)", rec.Code)
		}
	})

	t.Run("out of range is ignored", func(t *testing.T) {
		rec := httptest.NewRecorder()
		s.handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/?echo-code=999", nil))
		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want 200 (invalid code ignored)", rec.Code)
		}
		var doc map[string]any
		_ = json.Unmarshal(rec.Body.Bytes(), &doc)
		if _, ok := doc["applied"]; ok {
			t.Errorf("applied present for an ignored directive: %v", doc["applied"])
		}
	})
}

func TestEchoCommandDelay(t *testing.T) {
	s := newTestServer(t, baseConfig())
	rec := httptest.NewRecorder()

	start := time.Now()
	s.handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/?echo-delay=20ms", nil))
	if elapsed := time.Since(start); elapsed < 20*time.Millisecond {
		t.Errorf("elapsed = %v, want >= 20ms", elapsed)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var doc map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &doc); err != nil {
		t.Fatalf("response is not JSON: %v", err)
	}
	applied, ok := doc["applied"].(map[string]any)
	if !ok || applied["delay"] != "20ms" {
		t.Errorf("applied = %v, want delay 20ms", doc["applied"])
	}
}

func TestEchoCommandDelayClamped(t *testing.T) {
	cfg := baseConfig()
	cfg.MaxDelay = 5 * time.Millisecond
	s := newTestServer(t, cfg)
	rec := httptest.NewRecorder()

	// Request far more delay than the cap; it must clamp and stay fast.
	start := time.Now()
	s.handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/?echo-delay=10s", nil))
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Errorf("elapsed = %v, want clamped well under 1s", elapsed)
	}
	var doc map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &doc)
	applied, _ := doc["applied"].(map[string]any)
	if applied["delay"] != "5ms" {
		t.Errorf("applied delay = %v, want clamped to 5ms", applied["delay"])
	}
}

func TestEchoCommandHeaders(t *testing.T) {
	s := newTestServer(t, baseConfig())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/?echo-header=X-Test:1", nil)
	req.Header.Add("X-Echo-Header", "X-Cache: HIT")
	s.handler().ServeHTTP(rec, req)

	if got := rec.Header().Get("X-Test"); got != "1" {
		t.Errorf("X-Test = %q, want 1", got)
	}
	if got := rec.Header().Get("X-Cache"); got != "HIT" {
		t.Errorf("X-Cache = %q, want HIT", got)
	}
	var doc map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &doc)
	applied, ok := doc["applied"].(map[string]any)
	if !ok {
		t.Fatalf("applied missing; got %v", doc["applied"])
	}
	if hdrs, _ := applied["headers"].([]any); len(hdrs) != 2 {
		t.Errorf("applied headers = %v, want 2 entries", applied["headers"])
	}
}

func TestEchoCommandCookies(t *testing.T) {
	s := newTestServer(t, baseConfig())

	t.Run("sets response cookies and reports them", func(t *testing.T) {
		rec := httptest.NewRecorder()
		s.handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/?echo-cookie=session:abc&echo-cookie=theme:dark", nil))

		got := map[string]string{}
		for _, c := range rec.Result().Cookies() {
			got[c.Name] = c.Value
		}
		if got["session"] != "abc" || got["theme"] != "dark" {
			t.Errorf("Set-Cookie = %v, want session=abc and theme=dark", got)
		}
		var doc map[string]any
		_ = json.Unmarshal(rec.Body.Bytes(), &doc)
		applied, _ := doc["applied"].(map[string]any)
		if ck, _ := applied["cookies"].([]any); len(ck) != 2 {
			t.Errorf("applied.cookies = %v, want 2 entries", applied["cookies"])
		}
	})

	t.Run("cookies are set even in no-body mode", func(t *testing.T) {
		cfg := baseConfig()
		cfg.EchoBackToClient = false
		rec := httptest.NewRecorder()
		newTestServer(t, cfg).handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/?echo-cookie=id:1", nil))
		if c := rec.Result().Cookies(); len(c) != 1 || c[0].Name != "id" {
			t.Errorf("cookies = %v, want id set on the no-body response", c)
		}
	})
}

func TestEchoReservedHeadersProtected(t *testing.T) {
	s := newTestServer(t, baseConfig())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet,
		"/?echo-header=Cache-Control:public&echo-header=X-Content-Type-Options:off&echo-header=Content-Type:text/html&echo-header=X-Allowed:yes", nil)
	s.handler().ServeHTTP(rec, req)

	// echo's own contract/safety headers survive the override attempt...
	if got := rec.Header().Get("Cache-Control"); got != "no-store" {
		t.Errorf("Cache-Control = %q, want no-store (reserved)", got)
	}
	if got := rec.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Errorf("X-Content-Type-Options = %q, want nosniff (reserved)", got)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Errorf("Content-Type = %q, want application/json (reserved)", ct)
	}
	// ...but a non-reserved header still goes through, and is the only one in applied.
	if got := rec.Header().Get("X-Allowed"); got != "yes" {
		t.Errorf("X-Allowed = %q, want yes", got)
	}
	var doc map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &doc)
	applied, _ := doc["applied"].(map[string]any)
	hdrs, _ := applied["headers"].([]any)
	if len(hdrs) != 1 || hdrs[0] != "X-Allowed" {
		t.Errorf("applied.headers = %v, want [X-Allowed] only", applied["headers"])
	}
}

func TestEchoAppliedOmittedWhenNoDirectives(t *testing.T) {
	// Commands enabled but the caller asks for nothing: no applied block at all.
	rec := httptest.NewRecorder()
	newTestServer(t, baseConfig()).handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	var doc map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &doc); err != nil {
		t.Fatalf("response is not JSON: %v", err)
	}
	if _, ok := doc["applied"]; ok {
		t.Errorf("applied present for a plain request: %v", doc["applied"])
	}
}

func TestEchoCommandsDisabled(t *testing.T) {
	cfg := baseConfig()
	cfg.CommandsEnabled = false
	s := newTestServer(t, cfg)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/?echo-code=503&echo-header=X-Test:1", nil)
	s.handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200 (commands disabled)", rec.Code)
	}
	if got := rec.Header().Get("X-Test"); got != "" {
		t.Errorf("X-Test = %q, want unset (commands disabled)", got)
	}
	var doc map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &doc)
	if _, ok := doc["applied"]; ok {
		t.Errorf("applied present while commands disabled: %v", doc["applied"])
	}
}

func TestEchoCommandStatusWithBackDisabled(t *testing.T) {
	cfg := baseConfig()
	cfg.EchoBackToClient = false
	s := newTestServer(t, cfg)

	// No directive: still 204.
	rec := httptest.NewRecorder()
	s.handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusNoContent {
		t.Errorf("status = %d, want 204", rec.Code)
	}

	// echo-code overrides the 204 even with the body suppressed, and a caller
	// still cannot override Content-Type (reserved) on the no-body path — while
	// the security headers remain in place.
	rec = httptest.NewRecorder()
	s.handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/?echo-code=503&echo-header=Content-Type:text/html", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want 503", rec.Code)
	}
	if rec.Body.Len() != 0 {
		t.Errorf("body = %q, want empty in no-body mode", rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); ct == "text/html" {
		t.Errorf("Content-Type = %q, want it not overridable to text/html", ct)
	}
	if got := rec.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Errorf("X-Content-Type-Options = %q, want nosniff even in no-body mode", got)
	}
}

func TestEchoPrettyPrint(t *testing.T) {
	indented := func(body string) bool { return strings.Contains(body, "\n  ") }

	t.Run("compact by default", func(t *testing.T) {
		rec := httptest.NewRecorder()
		newTestServer(t, baseConfig()).handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
		if indented(rec.Body.String()) {
			t.Errorf("response is indented by default: %s", rec.Body.String())
		}
	})

	t.Run("query flag indents", func(t *testing.T) {
		rec := httptest.NewRecorder()
		newTestServer(t, baseConfig()).handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/?echo-pretty-print", nil))
		if !indented(rec.Body.String()) {
			t.Errorf("response is not indented with echo-pretty-print: %s", rec.Body.String())
		}
	})

	t.Run("header flag indents", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Echo-Pretty-Print", "true")
		newTestServer(t, baseConfig()).handler().ServeHTTP(rec, req)
		if !indented(rec.Body.String()) {
			t.Errorf("response is not indented with X-Echo-Pretty-Print: %s", rec.Body.String())
		}
	})

	t.Run("env default indents and is always available", func(t *testing.T) {
		cfg := baseConfig()
		cfg.PrettyPrint = true
		cfg.CommandsEnabled = false // pretty-print must work regardless of the commands gate
		rec := httptest.NewRecorder()
		newTestServer(t, cfg).handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
		if !indented(rec.Body.String()) {
			t.Errorf("response is not indented with ECHO_PRETTY_PRINT: %s", rec.Body.String())
		}
	})

	t.Run("query can override the env default off", func(t *testing.T) {
		cfg := baseConfig()
		cfg.PrettyPrint = true
		rec := httptest.NewRecorder()
		newTestServer(t, cfg).handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/?echo-pretty-print=false", nil))
		if indented(rec.Body.String()) {
			t.Errorf("echo-pretty-print=false did not turn off indentation: %s", rec.Body.String())
		}
	})
}

// freePort reserves an ephemeral port and releases it for the server to bind.
func freePort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("reserve port: %v", err)
	}
	defer func() { _ = l.Close() }()
	return l.Addr().(*net.TCPAddr).Port
}

// TestRunLifecycle exercises the full listener lifecycle: both listeners come
// up, serve their endpoints, and drain cleanly on context cancellation.
func TestRunLifecycle(t *testing.T) {
	cfg := baseConfig()
	cfg.HTTPPort = freePort(t)
	cfg.MetricsEnabled = true
	cfg.MetricsPort = freePort(t)
	cfg.ShutdownTimeout = 5 * time.Second

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- newTestServer(t, cfg).Run(ctx) }()

	get := func(port int, path string) *http.Response {
		t.Helper()
		var resp *http.Response
		var err error
		for range 100 {
			resp, err = http.Get(fmt.Sprintf("http://127.0.0.1:%d%s", port, path))
			if err == nil {
				return resp
			}
			time.Sleep(20 * time.Millisecond)
		}
		t.Fatalf("GET :%d%s never succeeded: %v", port, path, err)
		return nil
	}

	resp := get(cfg.HTTPPort, "/healthz")
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("/healthz status = %d, want 200", resp.StatusCode)
	}

	resp = get(cfg.MetricsPort, "/metrics")
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("/metrics status = %d, want 200", resp.StatusCode)
	}

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Run returned %v, want nil after graceful shutdown", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("Run did not return after cancellation")
	}
}

func TestRunReportsBindFailure(t *testing.T) {
	// Occupy the wildcard port so the listener fails to bind and Run surfaces
	// the error (a loopback-only listener would not conflict on every OS).
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer func() { _ = l.Close() }()

	cfg := baseConfig()
	cfg.HTTPPort = l.Addr().(*net.TCPAddr).Port
	cfg.MetricsEnabled = false

	if err := newTestServer(t, cfg).Run(context.Background()); err == nil {
		t.Error("Run = nil, want bind error for an occupied port")
	}
}

func TestRecovererTurnsPanicInto500(t *testing.T) {
	h := recoverer(slog.New(slog.NewTextHandler(io.Discard, nil)))(
		http.HandlerFunc(func(http.ResponseWriter, *http.Request) { panic("boom") }))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "internal server error") {
		t.Errorf("body = %q, want the generic error document", rec.Body.String())
	}
}

func TestNewClampsMaxDelayUnderWriteTimeout(t *testing.T) {
	cfg := baseConfig()
	cfg.MaxDelay = writeTimeout + time.Minute
	newTestServer(t, cfg)
	if cfg.MaxDelay >= writeTimeout {
		t.Errorf("MaxDelay = %v, want clamped under the %v write timeout", cfg.MaxDelay, writeTimeout)
	}
}

// errBody always fails, exercising the request-body read error path.
type errBody struct{}

func (errBody) Read([]byte) (int, error) { return 0, errors.New("broken body") }
func (errBody) Close() error             { return nil }

func TestEchoHandlerBodyReadError(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Body = errBody{}
	newTestServer(t, baseConfig()).handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400 for an unreadable body", rec.Code)
	}
}

func TestEchoDelayAbortsWhenClientGone(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // the client is already gone when the delay starts

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/?echo-delay=5s", nil).WithContext(ctx)

	start := time.Now()
	newTestServer(t, baseConfig()).handler().ServeHTTP(rec, req)
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Errorf("handler took %v, want immediate return for a cancelled context", elapsed)
	}
	if rec.Body.Len() != 0 {
		t.Errorf("body = %q, want none when the delay is aborted", rec.Body.String())
	}
}
