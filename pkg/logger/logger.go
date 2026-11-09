package logger

import (
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

// Init initializes the logger
func Init(level string) error {
	log.SetFormatter(&log.TextFormatter{
		TimestampFormat:  "2006-01-02 15:04:05",
		DisableTimestamp: false,
		ForceQuote:       true,
		FullTimestamp:    true,
	})

	// Parse and set log level
	parsedLevel, err := log.ParseLevel(strings.ToLower(level))
	if err != nil {
		return err
	}
	log.SetLevel(parsedLevel)

	// Output to stdout
	log.SetOutput(os.Stdout)
	return nil
}
