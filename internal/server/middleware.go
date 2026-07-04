package server

import (
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"
)

// statusRecorder captures the response status code for the access log and
// metrics, defaulting to 200 if the handler writes a body without an explicit
// WriteHeader.
type statusRecorder struct {
	http.ResponseWriter
	status int
	wrote  bool
}

func (r *statusRecorder) WriteHeader(code int) {
	if !r.wrote {
		r.status = code
		r.wrote = true
	}
	r.ResponseWriter.WriteHeader(code)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	if !r.wrote {
		r.status = http.StatusOK
		r.wrote = true
	}
	return r.ResponseWriter.Write(b)
}

// Unwrap exposes the wrapped ResponseWriter. http.ResponseController and
// coder/websocket both walk Unwrap to reach the real Flusher/Hijacker, so no
// explicit forwarding methods are needed here.
func (r *statusRecorder) Unwrap() http.ResponseWriter { return r.ResponseWriter }

// securityHeaders sets headers that make the echoed JSON inert in a browser:
// nosniff stops content-type guessing, and responses are never cached.
func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

// recoverer turns a handler panic into a 500 instead of crashing the process.
func recoverer(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					log.Error("recovered panic", "error", rec, "path", r.URL.Path, "stack", string(debug.Stack()))
					writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// quietPaths are the health-probe endpoints whose access log drops to debug:
// liveness/readiness probes are frequent and routine, so logging them at info
// would drown the real request log. (/metrics needs no entry — it lives on the
// separate metrics listener, outside this middleware.)
var quietPaths = map[string]struct{}{
	"/healthz": {},
	"/readyz":  {},
}

// observe records Prometheus metrics and emits the per-request access log.
func (s *Server) observe(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(rec, r)

		duration := time.Since(start)
		method := methodLabel(r.Method)
		httpRequests.WithLabelValues(method, statusClass(rec.status)).Inc()
		// A 101 response means the connection was hijacked for a WebSocket:
		// duration is the whole session, not a request latency, and one
		// long-lived session would bury every real sample in the histogram.
		if rec.status != http.StatusSwitchingProtocols {
			httpDuration.WithLabelValues(method).Observe(duration.Seconds())
		}

		if !s.cfg.DisableRequestLogs {
			level := slog.LevelInfo
			if _, quiet := quietPaths[r.URL.Path]; quiet {
				level = slog.LevelDebug
			}
			s.log.Log(r.Context(), level, "request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", rec.status,
				"remote", r.RemoteAddr,
				"duration", duration.String(),
			)
		}
	})
}
