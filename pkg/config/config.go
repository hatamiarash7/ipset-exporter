// Package config loads and validates exporter configuration.
package config

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"go.yaml.in/yaml/v3"
)

type App struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	LogLevel string `yaml:"log_level"`
}

type IPSet struct {
	UpdateInterval int      `yaml:"update_interval"`
	Names          []string `yaml:"names"`
}

type Config struct {
	App   App   `yaml:"app"`
	IPSet IPSet `yaml:"ipset"`
}

// Default returns an independent configuration with safe defaults.
func Default() Config {
	return Config{
		App:   App{Host: "127.0.0.1", Port: 4613, LogLevel: "info"},
		IPSet: IPSet{UpdateInterval: 15, Names: []string{"all"}},
	}
}

// Load reads one strict YAML document. Flag and environment handling belong to the CLI.
func Load(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open configuration: %w", err)
	}
	defer func() { _ = f.Close() }()
	cfg, err := decode(f)
	if err != nil {
		return nil, fmt.Errorf("configuration %q: %w", path, err)
	}
	return cfg, nil
}

func decode(r io.Reader) (*Config, error) {
	cfg := Default()
	decoder := yaml.NewDecoder(r)
	decoder.KnownFields(true)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("decode YAML: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return nil, errors.New("expected exactly one YAML document")
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Validate also protects callers that construct Config without loading YAML.
func (c Config) Validate() error {
	if c.App.Port < 1 || c.App.Port > 65535 {
		return errors.New("app.port must be between 1 and 65535")
	}
	if strings.TrimSpace(c.App.Host) != c.App.Host {
		return errors.New("app.host must not contain surrounding whitespace")
	}
	if _, err := log.ParseLevel(c.App.LogLevel); err != nil {
		return fmt.Errorf("app.log_level: %w", err)
	}
	if c.IPSet.UpdateInterval <= 0 || int64(c.IPSet.UpdateInterval) > int64((1<<63-1)/time.Second) {
		return errors.New("ipset.update_interval must be a positive number of seconds within time.Duration range")
	}
	if len(c.IPSet.Names) == 0 {
		return errors.New("ipset.names must contain at least one name or 'all'")
	}
	for _, name := range c.IPSet.Names {
		if name == "" || len(name) > 31 || strings.ContainsAny(name, " \t\r\n\x00") {
			return fmt.Errorf("invalid ipset name %q: use 1–31 bytes without whitespace or NUL", name)
		}
	}
	return nil
}
