package config

import (
	"flag"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// Config represents the configuration of the application
type Config struct {
	App struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		LogLevel string `yaml:"log_level"`
	} `yaml:"app"`

	IPSet struct {
		Names []string `yaml:"names"`
	} `yaml:"ipset"`
}

// Load reads the configuration from the given file path
func Load() (*Config, error) {
	var configPath string
	flag.StringVar(&configPath, "config", getEnv("CONFIG_FILE", "config.yml"), "Configuration file path")
	flag.Parse()

	var cfg Config
	absPath, err := filepath.Abs(configPath)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(absPath)
	if err != nil {
		return nil, err
	}

	defer func() {
		if f != nil {
			if err = f.Close(); err != nil {
				log.WithError(err).Error("Error closing config file")
			}
		}
	}()
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&cfg)
	return &cfg, err
}

// getEnv retrieves environment variables or returns a default value.
func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}
