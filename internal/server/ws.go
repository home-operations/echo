package server

import (
	"context"
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
		typ, data, err := readWithIdleTimeout(ctx, conn, s.cfg.WSIdleTimeout)
		if err != nil {
			return // normal close, idle timeout, context cancellation, or client gone
		}
		if err := conn.Write(ctx, typ, data); err != nil {
			return
		}
	}
}

// readWithIdleTimeout reads the next message, giving up after idle so a silent
// or half-open client can't hold the connection (and its goroutine) forever.
// The library replies to pings internally during the wait, but a ping is
// client-initiated liveness, not activity — only a real message resets the
// clock. idle <= 0 waits indefinitely.
func readWithIdleTimeout(ctx context.Context, conn *websocket.Conn, idle time.Duration) (websocket.MessageType, []byte, error) {
	if idle > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, idle)
		defer cancel()
	}
	typ, data, err := conn.Read(ctx)
	if err != nil && ctx.Err() != nil {
		_ = conn.Close(websocket.StatusGoingAway, "idle timeout")
	}
	return typ, data, err
}

func isWebSocketUpgrade(r *http.Request) bool {
	return strings.EqualFold(r.Header.Get("Upgrade"), "websocket") &&
		strings.Contains(strings.ToLower(r.Header.Get("Connection")), "upgrade")
}
