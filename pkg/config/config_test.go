package config

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yml")
	if err := os.WriteFile(path, []byte("app:\n  host: ::1\n  port: 8080\nipset:\n  names: [one, two]\n"), 0600); err != nil {
		t.Fatal(err)
	}
	for range 2 {
		cfg, err := Load(path)
		if err != nil {
			t.Fatal(err)
		}
		if cfg.App.Host != "::1" || cfg.App.Port != 8080 || cfg.IPSet.UpdateInterval != 15 || len(cfg.IPSet.Names) != 2 {
			t.Fatalf("unexpected configuration: %+v", cfg)
		}
	}
	if _, err := Load(filepath.Join(t.TempDir(), "missing.yml")); err == nil {
		t.Fatal("missing file must fail")
	}
}

func TestDecode(t *testing.T) {
	tests := []struct {
		name, yaml, wantError string
	}{
		{"defaults", "{}", ""},
		{"empty", "", "decode YAML"},
		{"malformed", "app: [", "decode YAML"},
		{"unknown field", "app:\n  prt: 42", "field prt"},
		{"duplicate field", "app:\n  port: 42\n  port: 43", "already defined"},
		{"multiple documents", "{}\n---\n{}", "exactly one"},
		{"zero interval", "ipset:\n  update_interval: 0", "update_interval"},
		{"negative interval", "ipset:\n  update_interval: -1", "update_interval"},
		{"low port", "app:\n  port: 0", "app.port"},
		{"high port", "app:\n  port: 65536", "app.port"},
		{"log level", "app:\n  log_level: noisy", "log_level"},
		{"no sets", "ipset:\n  names: []", "ipset.names"},
		{"empty name", "ipset:\n  names: ['']", "invalid ipset name"},
		{"long name", "ipset:\n  names: ['" + strings.Repeat("a", 32) + "']", "invalid ipset name"},
		{"whitespace name", "ipset:\n  names: [' a']", "invalid ipset name"},
	}
	if strconv.IntSize == 64 {
		tests = append(tests, struct{ name, yaml, wantError string }{"overflow interval", "ipset:\n  update_interval: 9223372036854775807", "update_interval"})
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := decode(strings.NewReader(tt.yaml))
			if tt.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantError) {
					t.Fatalf("expected %q, got %v", tt.wantError, err)
				}
				return
			}
			if err != nil || cfg.App.Port != 4613 || cfg.IPSet.Names[0] != "all" {
				t.Fatalf("unexpected defaults: %+v, %v", cfg, err)
			}
		})
	}
}
