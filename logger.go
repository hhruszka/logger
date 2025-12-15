package logger

import (
	"errors"
	"fmt"
	"os"
	"syscall"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// InitLogger initializes and configures the global Zap logger.
// It writes logs to the specified file and optionally to stdout.
// Use logToStdout=true to enable dual output (both file and console).
func InitLogger(logLevel string, logFilePath string, logToStdout bool) (*zap.Logger, error) {
	// Create a default configuration (Default is Debug level, Console encoding)
	cfg := zap.NewDevelopmentConfig()

	// Customize the encoder
	cfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")
	cfg.DisableStacktrace = true

	// Set the Log Level
	if level, err := zapcore.ParseLevel(logLevel); err == nil {
		cfg.Level = zap.NewAtomicLevelAt(level)
	} else {
		// Debug in NewDevelopmentConfig, just log the warning
		fmt.Printf("Invalid log level '%s', defaulting to DEBUG\n", logLevel)
	}

	// Configure Output Paths
	// zap.Config handles opening files and "stdout" logic automatically
	cfg.OutputPaths = []string{}
	if logFilePath != "" {
		cfg.OutputPaths = append(cfg.OutputPaths, logFilePath)
	}
	if logToStdout {
		cfg.OutputPaths = append(cfg.OutputPaths, "stdout")
	}

	// Build the logger
	logger, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	zap.ReplaceGlobals(logger)
	return logger, nil
}

// CloseLogger flushes any buffered log entries.
// Should be called using defer in your main function.
func CloseLogger() {
	err := zap.L().Sync()
	if err != nil {
		// Define the "whitelist" of errors to ignore.
		// These are safe to ignore because they simply mean the
		// underlying output stream (like stdout) doesn't support syncing.
		ignoredErrors := []error{
			syscall.ENOTTY, // "inappropriate ioctl for device"
			syscall.EINVAL, // "invalid argument"
		}

		for _, ignored := range ignoredErrors {
			// errors.Is unwraps the error chain to find if the underlying
			// cause matches.
			if errors.Is(err, ignored) {
				return
			}
		}

		// If it wasn't a whitelisted error, print it.
		fmt.Fprintf(os.Stderr, "Error flushing logger: %v\n", err)
	}
}
