package logger

import (
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

// Fields type for structured logging
type Fields map[string]interface{}

// Init initializes the logger with configuration
func Init(level, format, output string) error {
	log = logrus.New()

	// Set log level
	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		return fmt.Errorf("invalid log level: %s", level)
	}
	log.SetLevel(logLevel)

	// Set log format
	switch format {
	case "json":
		log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
		})
	case "text":
		log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: time.RFC3339,
		})
	default:
		return fmt.Errorf("unsupported log format: %s", format)
	}

	// Set output
	switch output {
	case "stdout":
		log.SetOutput(os.Stdout)
	case "stderr":
		log.SetOutput(os.Stderr)
	default:
		file, err := os.OpenFile(output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return fmt.Errorf("failed to open log file: %w", err)
		}
		log.SetOutput(file)
	}

	return nil
}

// GetLogger returns the logger instance
func GetLogger() *logrus.Logger {
	if log == nil {
		// Initialize with defaults if not initialized
		Init("info", "text", "stdout")
	}
	return log
}

// Info logs info level message with fields
func Info(message string, fields Fields) {
	GetLogger().WithFields(logrus.Fields(fields)).Info(message)
}

// Warn logs warning level message with fields
func Warn(message string, fields Fields) {
	GetLogger().WithFields(logrus.Fields(fields)).Warn(message)
}

// Error logs error level message with fields
func Error(message string, fields Fields) {
	GetLogger().WithFields(logrus.Fields(fields)).Error(message)
}

// Debug logs debug level message with fields
func Debug(message string, fields Fields) {
	GetLogger().WithFields(logrus.Fields(fields)).Debug(message)
}

// Fatal logs fatal level message with fields and exits
func Fatal(message string, fields Fields) {
	GetLogger().WithFields(logrus.Fields(fields)).Fatal(message)
}

// WithFields creates a logger with predefined fields
func WithFields(fields Fields) *logrus.Entry {
	return GetLogger().WithFields(logrus.Fields(fields))
}

// LogOperation logs operation start/end with timing
func LogOperation(operation string, fields Fields) func() {
	start := time.Now()
	fields["operation"] = operation
	fields["start_time"] = start.Format(time.RFC3339)

	Info(fmt.Sprintf("Starting %s", operation), fields)

	return func() {
		duration := time.Since(start)
		fields["duration"] = duration.String()
		fields["end_time"] = time.Now().Format(time.RFC3339)

		Info(fmt.Sprintf("Completed %s", operation), fields)
	}
}

// LogError logs error with context
func LogError(err error, context string, fields Fields) {
	if fields == nil {
		fields = Fields{}
	}
	fields["error"] = err.Error()
	fields["context"] = context

	Error(fmt.Sprintf("Error in %s: %v", context, err), fields)
}

// LogValidationError logs validation errors
func LogValidationError(field, message string, fields Fields) {
	if fields == nil {
		fields = Fields{}
	}
	fields["validation_field"] = field
	fields["validation_message"] = message

	Warn("Validation error", fields)
}

// LogResourceOperation logs resource-specific operations
func LogResourceOperation(operation, resourceType, resourceName string, fields Fields) {
	if fields == nil {
		fields = Fields{}
	}
	fields["resource_type"] = resourceType
	fields["resource_name"] = resourceName
	fields["operation"] = operation

	Info(fmt.Sprintf("%s %s '%s'", operation, resourceType, resourceName), fields)
}

// LogProviderOperation logs provider-specific operations
func LogProviderOperation(operation, provider string, fields Fields) {
	if fields == nil {
		fields = Fields{}
	}
	fields["provider"] = provider
	fields["operation"] = operation

	Info(fmt.Sprintf("%s for provider '%s'", operation, provider), fields)
}
