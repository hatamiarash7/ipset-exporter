package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hatamiarash7/ipset-exporter/pkg/config"
	"github.com/hatamiarash7/ipset-exporter/pkg/exporter"
	"github.com/hatamiarash7/ipset-exporter/pkg/logger"
	log "github.com/sirupsen/logrus"
)

var configs *config.Config

func init() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.WithError(err).Fatal("failed to load configs")
	}
	configs = cfg

	// Initialize logger
	logger.Init(cfg.App.LogLevel)
}

func main() {
	app := exporter.NewExporter(configs)

	if err := app.Boot(); err != nil {
		log.WithError(err).Fatal("could not boot application")
	}

	closeSignal := make(chan os.Signal, 1)
	signal.Notify(closeSignal, syscall.SIGTERM, syscall.SIGINT, os.Interrupt)

	<-closeSignal

	log.Info("Shutdown...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := app.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}

	<-ctx.Done()

}
