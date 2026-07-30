package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"
	"time"
)

// noKeepAliveClient avoids lingering connections that race with other tests
// mutating process-global state (e.g. time.Local).
var noKeepAliveClient = &http.Client{
	Transport: &http.Transport{DisableKeepAlives: true},
	Timeout:   2 * time.Second,
}

func newListener(t *testing.T) net.Listener {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	return ln
}

func waitReady(t *testing.T, url string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for {
		resp, err := noKeepAliveClient.Get(url)
		if err == nil {
			resp.Body.Close()
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("server did not become ready: %v", err)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestGracefulServe_CompletesInFlightRequest(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})

	mux := http.NewServeMux()
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "pong")
	})
	mux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		close(started)
		<-release
		_, _ = io.WriteString(w, "done")
	})

	ln := newListener(t)
	addr := ln.Addr().String()
	srv := &http.Server{Handler: mux}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- gracefulServe(ctx, srv, ln, 2*time.Second)
	}()
	t.Cleanup(cancel)

	waitReady(t, "http://"+addr+"/ping")

	reqDone := make(chan string, 1)
	go func() {
		resp, err := noKeepAliveClient.Get("http://" + addr + "/slow")
		if err != nil {
			reqDone <- fmt.Sprintf("err: %v", err)
			return
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		reqDone <- string(body)
	}()

	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("in-flight request never started")
	}

	cancel() // trigger graceful shutdown while request is in flight

	// ctx.Done() is closed synchronously with cancel(); release the handler
	// only after that, so the slow handler finishes inside the drain window.
	<-ctx.Done()
	close(release)

	select {
	case body := <-reqDone:
		if body != "done" {
			t.Fatalf("in-flight request result = %q, want done", body)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("in-flight request did not complete during shutdown")
	}

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("gracefulServe returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("gracefulServe did not return")
	}
}

func TestGracefulServe_RejectsNewConnectionsAfterShutdown(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "ok")
	})

	ln := newListener(t)
	addr := ln.Addr().String()
	srv := &http.Server{Handler: mux}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- gracefulServe(ctx, srv, ln, time.Second)
	}()

	waitReady(t, "http://"+addr+"/")
	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("gracefulServe returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("gracefulServe did not return")
	}

	_, err := noKeepAliveClient.Get("http://" + addr + "/")
	if err == nil {
		t.Fatal("expected connection error after shutdown, got nil")
	}
}
