package main

import (
	"log/slog"
	"testing"

	"github.com/home-operations/echo/internal/config"
)

func TestNewLoggerHandlerSelection(t *testing.T) {
	if h := newLogger(&config.Config{LogFormat: "text"}).Handler(); func() bool {
		_, ok := h.(*slog.TextHandler)
		return !ok
	}() {
		t.Errorf("LogFormat=text produced %T, want *slog.TextHandler", h)
	}
	if h := newLogger(&config.Config{LogFormat: "json"}).Handler(); func() bool {
		_, ok := h.(*slog.JSONHandler)
		return !ok
	}() {
		t.Errorf("LogFormat=json produced %T, want *slog.JSONHandler", h)
	}
}
