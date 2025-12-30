package config

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

// LoggerConfig holds logger configuration
type LoggerConfig struct {
	Level      string // debug, info, warn, error
	Format     string // json or text
	Output     string // stdout or file path
	FilePath   string // log file path if output is file
}

// InitializeLogger initializes and configures the global logger
func InitializeLogger(cfg *LoggerConfig) (*logrus.Logger, error) {
	logger := logrus.New()

	// Set log level
	level, err := logrus.ParseLevel(cfg.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	// Set formatter
	if cfg.Format == "json" {
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
			PrettyPrint:     false,
		})
	} else {
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
			PadLevelText:    true,
		})
	}

	// Set output
	var output io.Writer
	if cfg.Output == "file" && cfg.FilePath != "" {
		file, err := os.OpenFile(cfg.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, err
		}
		output = file
	} else {
		output = os.Stdout
	}

	logger.SetOutput(output)

	return logger, nil
}

// GetLoggerConfig returns logger config based on environment
func GetLoggerConfig(env string) *LoggerConfig {
	switch env {
	case "production":
		return &LoggerConfig{
			Level:  "warn",
			Format: "json",
			Output: "stdout",
		}
	case "staging":
		return &LoggerConfig{
			Level:  "info",
			Format: "text",
			Output: "stdout",
		}
	default: // development
		return &LoggerConfig{
			Level:  "debug",
			Format: "text",
			Output: "stdout",
		}
	}
}
