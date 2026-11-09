package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLI(t *testing.T) {
	t.Setenv("CONFIG_FILE", filepath.Join(t.TempDir(), "missing.yml"))
	for _, arg := range []string{"-version", "--version", "-help"} {
		var out, stderr bytes.Buffer
		if err := run(context.Background(), []string{arg}, &out, &stderr); err != nil {
			t.Fatalf("%s must work without configuration: %v", arg, err)
		}
		if arg != "-help" && !strings.Contains(out.String(), "ipset-exporter") {
			t.Fatal("version output missing")
		}
	}
	path := filepath.Join(t.TempDir(), "valid.yml")
	if err := os.WriteFile(path, []byte("{}"), 0600); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := run(context.Background(), []string{"-config", path, "-check-config"}, &out, &out); err != nil {
		t.Fatalf("explicit flag must override invalid environment: %v", err)
	}
	t.Setenv("CONFIG_FILE", path)
	if err := run(context.Background(), []string{"-check-config"}, &out, &out); err != nil {
		t.Fatalf("environment configuration: %v", err)
	}
	for _, args := range [][]string{{"-unknown"}, {"unexpected"}} {
		if err := run(context.Background(), args, &out, &out); err == nil {
			t.Fatalf("invalid arguments accepted: %v", args)
		}
	}
}
