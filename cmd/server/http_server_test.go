package main

import (
	"net/http"
	"testing"
	"time"

	"github.com/prefeitura-rio/app-go-api/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestNewHTTPServer_AppliesConfiguredTimeouts(t *testing.T) {
	handler := http.NewServeMux()
	cfg := config.ServerSettings{
		ReadHeaderTimeout: 10,
		ReadTimeout:       30,
		IdleTimeout:       120,
	}

	srv := newHTTPServer(handler, cfg)

	assert.Equal(t, 10*time.Second, srv.ReadHeaderTimeout)
	assert.Equal(t, 30*time.Second, srv.ReadTimeout)
	assert.Equal(t, 120*time.Second, srv.IdleTimeout)
	assert.Same(t, handler, srv.Handler)
}

func TestNewHTTPServer_UsesCustomTimeouts(t *testing.T) {
	handler := http.NewServeMux()
	cfg := config.ServerSettings{
		ReadHeaderTimeout: 5,
		ReadTimeout:       15,
		IdleTimeout:       60,
	}

	srv := newHTTPServer(handler, cfg)

	assert.Equal(t, 5*time.Second, srv.ReadHeaderTimeout)
	assert.Equal(t, 15*time.Second, srv.ReadTimeout)
	assert.Equal(t, 60*time.Second, srv.IdleTimeout)
}
