package server

import (
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
	mux.HandleFunc("GET /healthz", handleHealth)
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

	doc := echo.Build(r, body, truncated, echo.Options{
		Now:            time.Now(),
		Hostname:       s.hostname,
		TrustedProxies: s.cfg.TrustedProxies,
		Kubernetes:     s.k8s,
	})

	if !s.cfg.EchoBackToClient {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	writeJSON(w, http.StatusOK, doc)
}

// writeJSON writes v as an application/json response. HTML escaping stays on so
// the JSON is inert if a browser is ever pointed at it.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(true)
	_ = enc.Encode(v)
}
