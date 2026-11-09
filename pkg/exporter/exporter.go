// Package exporter serves cached ipset metrics collected from Linux netlink.
package exporter

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/hatamiarash7/ipset-exporter/internal/buildinfo"
	"github.com/hatamiarash7/ipset-exporter/pkg/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

type App struct {
	address   string
	interval  time.Duration
	collector *collector
	handler   http.Handler
}

func NewExporter(cfg *config.Config) (*App, error) {
	if cfg == nil {
		return nil, errors.New("configuration is required")
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	a := &App{
		address:   net.JoinHostPort(cfg.App.Host, strconv.Itoa(cfg.App.Port)),
		interval:  time.Duration(cfg.IPSet.UpdateInterval) * time.Second,
		collector: newCollector(cfg.IPSet.Names, listIPSets),
	}
	registry := prometheus.NewRegistry()
	registry.MustRegister(a.collector)
	info := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ipset_exporter_build_info", Help: "Exporter build information.",
	}, []string{"version", "revision", "build_date"})
	info.WithLabelValues(buildinfo.Version, buildinfo.Revision, buildinfo.BuildDate).Set(1)
	registry.MustRegister(info)
	mux := http.NewServeMux()
	mux.Handle("GET /metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprintln(w, "ok")
	})
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, _ *http.Request) {
		if a.collector.state.Load().up != 1 {
			http.Error(w, "ipset collection unavailable", http.StatusServiceUnavailable)
			return
		}
		_, _ = fmt.Fprintln(w, "ok")
	})
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = fmt.Fprint(w, "<!doctype html><title>ipset exporter</title><h1>ipset exporter</h1><a href=\"/metrics\">Metrics</a>")
	})
	a.handler = mux
	return a, nil
}

// Run binds synchronously and returns on cancellation or a server error.
func (a *App) Run(ctx context.Context) error {
	listener, err := net.Listen("tcp", a.address)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", a.address, err)
	}
	return a.serve(ctx, listener)
}

func (a *App) serve(ctx context.Context, listener net.Listener) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	server := &http.Server{
		Handler: a.handler, ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout: 10 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second,
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		a.poll(ctx)
	}()
	serverErr := make(chan error, 1)
	go func() { serverErr <- server.Serve(listener) }()
	log.WithField("address", listener.Addr()).Info("Serving ipset metrics")
	var err error
	select {
	case <-ctx.Done():
	case err = <-serverErr:
		if errors.Is(err, http.ErrServerClosed) {
			err = nil
		}
	}
	cancel()
	shutdownCtx, stop := context.WithTimeout(context.Background(), 5*time.Second)
	defer stop()
	if shutdownErr := server.Shutdown(shutdownCtx); shutdownErr != nil {
		_ = server.Close()
		err = errors.Join(err, shutdownErr)
	}
	select {
	case <-done:
	case <-shutdownCtx.Done():
		err = errors.Join(err, fmt.Errorf("stop collector: %w", shutdownCtx.Err()))
	}
	return err
}

func (a *App) poll(ctx context.Context) {
	ticker := time.NewTicker(a.interval)
	defer ticker.Stop()
	for {
		if ctx.Err() != nil {
			return
		}
		a.collector.update()
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
