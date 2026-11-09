package exporter

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hatamiarash7/ipset-exporter/pkg/config"
)

func TestEndpoints(t *testing.T) {
	cfg := config.Default()
	a, err := NewExporter(&cfg)
	if err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct {
		path string
		code int
	}{{"/metrics", 200}, {"/healthz", 200}, {"/readyz", 503}, {"/", 200}, {"/missing", 404}} {
		r := httptest.NewRecorder()
		a.handler.ServeHTTP(r, httptest.NewRequest(http.MethodGet, tt.path, nil))
		if r.Code != tt.code {
			t.Errorf("%s: got %d, want %d", tt.path, r.Code, tt.code)
		}
	}
	a.collector.list = func() (map[string]uint32, error) { return map[string]uint32{}, nil }
	a.collector.update()
	r := httptest.NewRecorder()
	a.handler.ServeHTTP(r, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if r.Code != 200 {
		t.Fatal("successful empty collection should be ready")
	}
	a.collector.list = func() (map[string]uint32, error) { return nil, errors.New("failed") }
	a.collector.update()
	r = httptest.NewRecorder()
	a.handler.ServeHTTP(r, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if r.Code != 503 {
		t.Fatal("failed collection should not be ready")
	}
}

func TestRunListenerFailure(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = listener.Close() }()
	cfg := config.Default()
	cfg.App.Port = listener.Addr().(*net.TCPAddr).Port
	a, err := NewExporter(&cfg)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := a.Run(ctx); err == nil {
		t.Fatal("occupied port must fail")
	}
}

func TestShutdownStopsPolling(t *testing.T) {
	cfg := config.Default()
	a, err := NewExporter(&cfg)
	if err != nil {
		t.Fatal(err)
	}
	collected := make(chan struct{}, 1)
	a.collector.list = func() (map[string]uint32, error) {
		select {
		case collected <- struct{}{}:
		default:
		}
		return map[string]uint32{}, nil
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- a.serve(ctx, listener) }()
	select {
	case <-collected:
	case <-time.After(2 * time.Second):
		t.Fatal("initial collection did not run")
	}
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("shutdown should return promptly")
	}
}
