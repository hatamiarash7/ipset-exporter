package exporter

import (
	"context"
	"net/http"
	"time"

	"github.com/hatamiarash7/ipset-exporter/pkg/config"
)

// App represents the application
type App struct {
	config        *config.Config
	metricsServer *http.Server
	stopChan      chan struct{}
}

// NewExporter creates a new instance of the exporter
func NewExporter(config *config.Config) *App {
	app := &App{
		config:   config,
		stopChan: make(chan struct{}),
	}
	return app
}

// Boot will boots the application
func (a *App) Boot() error {
	a.registerMetrics()

	// Start background metrics updater
	go a.updateMetricsPeriodically(a.config.IPSet.Names, time.Duration(a.config.IPSet.UpdateInterval)*time.Second)

	// Serve the metrics endpoint
	a.serveMetrics(a.config.IPSet.Names)
	return nil
}

// Shutdown will shutdown the application
func (a *App) Shutdown(ctx context.Context) error {
	// Stop the background updater
	close(a.stopChan)

	// Shutdown the metrics server
	if err := a.metricsServer.Shutdown(ctx); err != nil {
		return err
	}

	return nil
}
