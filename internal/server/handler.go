package server

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/home-operations/echo/internal/echo"
)

// handler builds the request-serving chain for the HTTP and HTTPS listeners.
// Order (outermost first): recover panics, set security headers, observe
// (metrics + access log), then route.
func (s *Server) handler() http.Handler {
	mux := http.NewServeMux()
	// The org pair standard: /healthz = liveness (cheap, no dependencies),
	// /readyz = readiness. echo has no serving condition beyond being up, so
	// readyz aliases healthz — the endpoint exists so probes and external
	// configs never need to change if that ever stops being true.
	mux.HandleFunc("GET /healthz", handleHealth)
	mux.HandleFunc("GET /readyz", handleHealth)
	if s.cfg.WSEnabled {
		mux.HandleFunc("GET /ws", s.handleWebSocket)
	}
	mux.HandleFunc("/", s.handleEcho)
	return recoverer(s.log)(securityHeaders(s.observe(mux)))
}

func handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleEcho(w http.ResponseWriter, r *http.Request) {
	body, truncated, err := echo.ReadBody(r, s.cfg.MaxBodyBytes)
	if err != nil {
		s.log.Warn("reading request body", "error", err)
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "failed to read request body"})
		return
	}

	cmd := echo.ParseCommands(r, echo.CommandOptions{
		Enabled:       s.cfg.CommandsEnabled,
		MaxDelay:      s.cfg.MaxDelay,
		DefaultPretty: s.cfg.PrettyPrint,
	})

	if !delay(r.Context(), cmd.Delay) {
		return // client disconnected or shutdown began during the delay
	}
	setCustomHeaders(w, cmd.Headers)
	for _, ck := range cmd.Cookies {
		http.SetCookie(w, ck)
	}

	status := http.StatusOK
	if !s.cfg.EchoBackToClient {
		status = http.StatusNoContent
	}
	if cmd.Status != 0 {
		status = cmd.Status
	}

	if !s.cfg.EchoBackToClient {
		w.WriteHeader(status)
		return
	}

	doc := echo.Build(r, body, truncated, echo.Options{
		Now:            time.Now(),
		Hostname:       s.hostname,
		TrustedProxies: s.cfg.TrustedProxies,
		Kubernetes:     s.k8s,
	})
	doc.Applied = cmd.Applied()

	writeJSONIndent(w, status, doc, cmd.Pretty)
}

// delay waits d before responding, returning early if the request context is
// cancelled (client gone or shutdown). It reports whether to proceed with the
// response. A non-positive d returns immediately.
func delay(ctx context.Context, d time.Duration) bool {
	if d <= 0 {
		return true
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-t.C:
		return true
	case <-ctx.Done():
		return false
	}
}

// setCustomHeaders applies caller-requested response headers, replacing any
// existing value (including echo's own defaults) so the caller gets exactly the
// header they asked for. writeJSON still enforces the JSON Content-Type after.
func setCustomHeaders(w http.ResponseWriter, h http.Header) {
	for name, values := range h {
		w.Header().Del(name)
		for _, v := range values {
			w.Header().Add(name, v)
		}
	}
}

// writeJSON writes v as a compact application/json response.
func writeJSON(w http.ResponseWriter, status int, v any) {
	writeJSONIndent(w, status, v, false)
}

// writeJSONIndent writes v as an application/json response, optionally indented.
// HTML escaping stays on so the JSON is inert if a browser is ever pointed at it.
func writeJSONIndent(w http.ResponseWriter, status int, v any, pretty bool) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(true)
	if pretty {
		enc.SetIndent("", "  ")
	}
	_ = enc.Encode(v)
}
