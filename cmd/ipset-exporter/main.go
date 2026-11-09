package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/hatamiarash7/ipset-exporter/internal/buildinfo"
	"github.com/hatamiarash7/ipset-exporter/pkg/config"
	"github.com/hatamiarash7/ipset-exporter/pkg/exporter"
	"github.com/hatamiarash7/ipset-exporter/pkg/logger"
	log "github.com/sirupsen/logrus"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()
	if err := run(ctx, os.Args[1:], os.Stdout, os.Stderr); err != nil {
		log.WithError(err).Error("Exporter stopped")
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	path := os.Getenv("CONFIG_FILE")
	if path == "" {
		path = "config.yml"
	}
	flags := flag.NewFlagSet("ipset-exporter", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.StringVar(&path, "config", path, "Configuration file path (or CONFIG_FILE)")
	version := flags.Bool("version", false, "Print build information and exit")
	check := flags.Bool("check-config", false, "Validate configuration and exit")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments: %v", flags.Args())
	}
	if *version {
		_, err := fmt.Fprintf(stdout, "ipset-exporter %s (revision=%s, build_date=%s)\n", buildinfo.Version, buildinfo.Revision, buildinfo.BuildDate)
		return err
	}
	cfg, err := config.Load(path)
	if err != nil {
		return err
	}
	if *check {
		_, err := fmt.Fprintln(stdout, "Configuration is valid")
		return err
	}
	if err := logger.Init(cfg.App.LogLevel); err != nil {
		return err
	}
	app, err := exporter.NewExporter(cfg)
	if err != nil {
		return err
	}
	log.WithField("version", buildinfo.Version).Info("Starting ipset exporter")
	return app.Run(ctx)
}
