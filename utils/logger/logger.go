package logger

import (
	"context"

	"github.com/sirupsen/logrus"
)

var globalLogger *logrus.Logger

// Initialize sets the global logger
func Initialize(logger *logrus.Logger) {
	globalLogger = logger
}

// Get returns the global logger
func Get() *logrus.Logger {
	if globalLogger == nil {
		return logrus.New()
	}
	return globalLogger
}

// Info logs an info message
func Info(msg string, fields ...map[string]interface{}) {
	if len(fields) > 0 {
		Get().WithFields(logrus.Fields(fields[0])).Info(msg)
	} else {
		Get().Info(msg)
	}
}

// Error logs an error message
func Error(msg string, err error, fields ...map[string]interface{}) {
	entry := Get().WithError(err)
	if len(fields) > 0 {
		entry = entry.WithFields(logrus.Fields(fields[0]))
	}
	entry.Error(msg)
}

// Debug logs a debug message
func Debug(msg string, fields ...map[string]interface{}) {
	if len(fields) > 0 {
		Get().WithFields(logrus.Fields(fields[0])).Debug(msg)
	} else {
		Get().Debug(msg)
	}
}

// Warn logs a warning message
func Warn(msg string, fields ...map[string]interface{}) {
	if len(fields) > 0 {
		Get().WithFields(logrus.Fields(fields[0])).Warn(msg)
	} else {
		Get().Warn(msg)
	}
}

// WithContext returns logger with context fields
func WithContext(ctx context.Context) *logrus.Entry {
	entry := Get().WithContext(ctx)

	// Add correlation ID if present
	if correlationID, ok := ctx.Value("correlation_id").(string); ok {
		entry = entry.WithField("correlation_id", correlationID)
	}

	// Add user ID if present
	if userID, ok := ctx.Value("user_id").(int); ok {
		entry = entry.WithField("user_id", userID)
	}

	return entry
}
