package logger

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// InitLogger initializes and configures the global Zap logger.
// It writes logs to the specified file and optionally to stdout.
// Use logToStdout=true to enable dual output (both file and console).
func InitLogger(logLevel string, logFilePath string, logToStdout bool) (*zap.Logger, error) {
	// Parse log level with fallback to Debug
	var level zapcore.Level
	if err := level.UnmarshalText([]byte(logLevel)); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid log level '%s', using debug\n", logLevel)
		level = zap.InfoLevel
	}

	// Configure console-style encoder
	encoderCfg := zap.NewDevelopmentEncoderConfig()
	encoderCfg.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")
	encoderCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	encoderCfg.EncodeCaller = zapcore.ShortCallerEncoder
	encoderCfg.StacktraceKey = "" // Disable stack traces

	// Open log file
	logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	// Configure output syncer (file only or file + stdout)
	var writeSyncer zapcore.WriteSyncer
	if logToStdout {
		writeSyncer = zapcore.NewMultiWriteSyncer(
			zapcore.AddSync(os.Stdout),
			zapcore.AddSync(logFile),
		)
	} else {
		writeSyncer = zapcore.AddSync(logFile)
	}

	// Create core and logger
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderCfg),
		writeSyncer,
		level,
	)

	logger := zap.New(core, zap.AddCaller())

	// Replace global logger so zap.L() and zap.S() work correctly
	zap.ReplaceGlobals(logger)

	return logger, nil
}

// CloseLogger flushes any buffered log entries.
// Should be called using defer in your main function.
func CloseLogger() {
	if err := zap.L().Sync(); err != nil {
		fmt.Fprintf(os.Stderr, "Error flushing logger: %v\n", err)
	}
}
