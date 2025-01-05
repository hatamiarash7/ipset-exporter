package exporter

import (
	"context"
	"net/http"

	"github.com/hatamiarash7/ipset-exporter/pkg/config"
)

// App represents the application
type App struct {
	config        *config.Config
	metricsServer *http.Server
}

// NewExporter creates a new instance of the exporter
func NewExporter(config *config.Config) *App {
	app := &App{config: config}
	return app
}

// Boot will boots the application
func (a *App) Boot() error {
	a.registerMetrics()
	a.serveMetrics(a.config.IPSet.Names)
	return nil
}

// Shutdown will shutdown the application
func (a *App) Shutdown(ctx context.Context) error {
	if err := a.metricsServer.Shutdown(ctx); err != nil {
		return err
	}

	return nil
}
