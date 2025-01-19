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
	os.Setenv("CONFIG_FILE", "test_config.yml")
	result := getEnv("CONFIG_FILE", "default.yml")
	assert.Equal(t, "test_config.yml", result, "Expected value from environment variable")

	// Test when the environment variable is not set
	os.Unsetenv("CONFIG_FILE")
	result = getEnv("CONFIG_FILE", "default.yml")
	assert.Equal(t, "default.yml", result, "Expected default value when environment variable is not set")
}

func TestLoadConfigFile(t *testing.T) {
	// Reset the flag set to avoid flag redefinition errors and handle Go test flags
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	// Set environment variable for config file path
	os.Setenv("CONFIG_FILE", "test_config.yml")

	// Mocking the file open and decode process
	mockFile := new(MockFile)
	defer os.Unsetenv("CONFIG_FILE")

	// Assume we have a test YAML file and mock the file open method
	mockFile.On("Open", "test_config.yml").Return(nil, nil)

	// Here we are testing a valid scenario
	cfg, err := Load()

	// Validate the function returns a configuration struct and no error
	assert.NoError(t, err, "Expected no error loading config")
	assert.NotNil(t, cfg, "Expected a non-nil configuration")
	assert.Equal(t, "localhost", cfg.App.Host, "Expected default host")
	assert.Equal(t, 8080, cfg.App.Port, "Expected default port")
}

func TestLoadConfigFileError(t *testing.T) {
	// Reset the flag set to avoid flag redefinition errors and handle Go test flags
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	// Test the error when the config file does not exist
	_, err := Load()
	assert.Error(t, err, "Expected error loading config due to non-existing file")
}

func TestLoadConfigWithInvalidYaml(t *testing.T) {
	// Reset the flag set to avoid flag redefinition errors and handle Go test flags
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	// Test loading a file with invalid YAML
	invalidYaml := []byte("invalid: yaml: syntax")
	mockFile := new(MockFile)
	mockFile.On("Open", "test_config.yml").Return(invalidYaml, nil)

	// Simulate the file being opened
	_, err := Load()
	assert.Error(t, err, "Expected error due to invalid YAML syntax")
}
