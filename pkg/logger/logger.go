package logger

import (
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

// Init initializes the logger
func Init(level string) {
	log.SetFormatter(&log.TextFormatter{
		TimestampFormat:  "2006-01-02 15:04:05",
		DisableTimestamp: false,
		ForceQuote:       true,
		FullTimestamp:    true,
	})
	log.Infoln("Setup Logger")

	// Parse and set log level
	parsedLevel, err := log.ParseLevel(strings.ToLower(level))
	if err != nil {
		log.Fatalf("Invalid log level '%s': %v", level, err)
	}
	log.SetLevel(parsedLevel)

	// Output to stdout
	log.SetOutput(os.Stdout)
}
