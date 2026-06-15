package server

import (
	"net/http"
	"strings"
	"time"

	"github.com/coder/websocket"
)

// handleWebSocket upgrades the connection and echoes every message back. A
// request to /ws without the WebSocket upgrade headers falls through to the
// normal HTTP echo, so the path behaves like any other for plain requests.
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	if !isWebSocketUpgrade(r) {
		s.handleEcho(w, r)
		return
	}

	// A WebSocket is long-lived, so clear the per-connection Read/Write deadlines
	// the server applies from ReadTimeout/WriteTimeout; the library manages
	// deadlines per message via the request context instead.
	rc := http.NewResponseController(w)
	_ = rc.SetReadDeadline(time.Time{})
	_ = rc.SetWriteDeadline(time.Time{})

	origins := s.cfg.WSAllowedOrigins
	if len(origins) == 0 {
		origins = []string{"*"}
	}

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{OriginPatterns: origins})
	if err != nil {
		s.log.Warn("websocket accept failed", "error", err)
		return
	}
	defer func() { _ = conn.CloseNow() }()

	ctx := r.Context()
	for {
		typ, data, err := conn.Read(ctx)
		if err != nil {
			return // normal close, context cancellation, or client gone
		}
		if err := conn.Write(ctx, typ, data); err != nil {
			return
		}
	}
}

func isWebSocketUpgrade(r *http.Request) bool {
	return strings.EqualFold(r.Header.Get("Upgrade"), "websocket") &&
		strings.Contains(strings.ToLower(r.Header.Get("Connection")), "upgrade")
}
