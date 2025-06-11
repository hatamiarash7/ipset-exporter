package config

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock for os.Open
type MockFile struct {
	mock.Mock
}

func (m *MockFile) Open(name string) (*os.File, error) {
	args := m.Called(name)
	return args.Get(0).(*os.File), args.Error(1)
}

func TestGetEnv(t *testing.T) {
	// Test when the environment variable is set
	_ = os.Setenv("CONFIG_FILE", "test_config.yml")
	result := getEnv("CONFIG_FILE", "default.yml")
	assert.Equal(t, "test_config.yml", result, "Expected value from environment variable")

	// Test when the environment variable is not set
	_ = os.Unsetenv("CONFIG_FILE")
	result = getEnv("CONFIG_FILE", "default.yml")
	assert.Equal(t, "default.yml", result, "Expected default value when environment variable is not set")
}

func TestLoadConfigFile(t *testing.T) {
	// Reset the flag set to avoid flag redefinition errors and handle Go test flags
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	// Set environment variable for config file path
	validConfigPath := "./test_config_valid.yml"
	_ = os.Setenv("CONFIG_FILE", validConfigPath)
	defer func() { _ = os.Unsetenv("CONFIG_FILE") }()

	// Create a dummy valid config file
	validYAML := `
app:
  host: "testhost"
  port: 1234
  log_level: "debug"
ipset:
  update_interval: 30
  names:
    - "test_set_1"
    - "test_set_2"
`
	err := os.WriteFile(validConfigPath, []byte(validYAML), 0644)
	assert.NoError(t, err, "Failed to write valid test config file")
	defer func() { _ = os.Remove(validConfigPath) }()

	cfg, err := Load()

	assert.NoError(t, err, "Expected no error loading config")
	assert.NotNil(t, cfg, "Expected a non-nil configuration")
	if cfg != nil { // Guard against nil pointer dereference if Load() failed differently
		assert.Equal(t, "testhost", cfg.App.Host, "Expected host from test config")
		assert.Equal(t, 1234, cfg.App.Port, "Expected port from test config")
		assert.Equal(t, "debug", cfg.App.LogLevel, "Expected log_level from test config")
		assert.Equal(t, 30, cfg.IPSet.UpdateInterval, "Expected update_interval from test config")
		assert.Equal(t, []string{"test_set_1", "test_set_2"}, cfg.IPSet.Names, "Expected names from test config")
	}
}

func TestLoadConfigFileError(t *testing.T) {
	// Reset the flag set to avoid flag redefinition errors and handle Go test flags
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	// Point to a non-existent file
	nonExistentConfigPath := "./test_config_non_existent.yml"
	_ = os.Setenv("CONFIG_FILE", nonExistentConfigPath)
	defer func() { _ = os.Unsetenv("CONFIG_FILE") }()

	// Test the error when the config file does not exist
	_, err := Load()
	assert.Error(t, err, "Expected error loading config due to non-existing file")
	assert.True(t, os.IsNotExist(err), "Expected a file not exist error")
}

func TestLoadConfigWithInvalidYaml(t *testing.T) {
	// Reset the flag set to avoid flag redefinition errors and handle Go test flags
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	invalidConfigPath := "./test_config_invalid.yml"
	_ = os.Setenv("CONFIG_FILE", invalidConfigPath)
	defer func() { _ = os.Unsetenv("CONFIG_FILE") }()

	// Create a dummy invalid config file
	// This YAML is structurally invalid due to the unclosed quote and incorrect indentation.
	invalidYAML := "app:\n  host: \"testhost\n  port: 1234"
	err := os.WriteFile(invalidConfigPath, []byte(invalidYAML), 0644)
	assert.NoError(t, err, "Failed to write invalid test config file")
	defer func() { _ = os.Remove(invalidConfigPath) }()

	_, err = Load()
	assert.Error(t, err, "Expected error due to invalid YAML syntax")
	// Optionally, check for a specific YAML error type if the library provides one
	// For now, just checking for any error is fine.
}
