package logger

import (
	"testing"
)

func TestSetup(t *testing.T) {
	Setup("debug", "json")
	if defaultLogger == nil {
		t.Fatal("expected logger to be initialized")
	}

	Setup("info", "text")
	if defaultLogger == nil {
		t.Fatal("expected logger to be initialized")
	}
}

func TestLogLevels(t *testing.T) {
	Setup("debug", "text")

	// These should not panic
	Debug("test debug", "key", "value")
	Info("test info", "key", "value")
	Warn("test warn", "key", "value")
	Error("test error", "key", "value")
}

func TestWith(t *testing.T) {
	child := With("component", "test")
	if child == nil {
		t.Fatal("expected child logger")
	}
	child.Info("child logger message")
}
