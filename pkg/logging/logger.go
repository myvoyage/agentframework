// Agent Framework - Structured Logging Package
// Copyright (C) 2025  Agent Framework Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package logging

import (
	"context"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is the interface for structured logging
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)

	With(fields ...Field) Logger
	WithErr(err error) Logger
	Named(name string) Logger
	Sync() error
}

// Field represents a log field (alias for zap.Field)
type Field = zap.Field

// Field constructors (aliases for zap field constructors)
var (
	String   = zap.String
	Int      = zap.Int
	Int64    = zap.Int64
	Float64  = zap.Float64
	Bool     = zap.Bool
	Any      = zap.Any
	Err      = zap.Error
	Duration = zap.Duration
	Time     = zap.Time
	Uint64   = zap.Uint64
)

// logger implements the Logger interface using zap
type logger struct {
	*zap.Logger
}

// New creates a new logger with the specified configuration
func New(level string, development bool) (Logger, error) {
	var config zap.Config
	if development {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		config = zap.NewProductionConfig()
	}

	// Set log level
	switch level {
	case "debug":
		config.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	case "info":
		config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	case "warn":
		config.Level = zap.NewAtomicLevelAt(zapcore.WarnLevel)
	case "error":
		config.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	default:
		config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	}

	// Build logger
	zapLogger, err := config.Build()
	if err != nil {
		return nil, err
	}

	return &logger{Logger: zapLogger}, nil
}

// NewFromZap creates a logger from an existing zap logger
func NewFromZap(zapLogger *zap.Logger) Logger {
	return &logger{Logger: zapLogger}
}

// With returns a new logger with the given fields
func (l *logger) With(fields ...Field) Logger {
	return &logger{Logger: l.Logger.With(fields...)}
}

// WithErr returns a new logger with an error field
func (l *logger) WithErr(err error) Logger {
	return &logger{Logger: l.Logger.With(zap.Error(err))}
}

// Named returns a new logger with the given name
func (l *logger) Named(name string) Logger {
	return &logger{Logger: l.Logger.Named(name)}
}

// Debug logs a debug message
func (l *logger) Debug(msg string, fields ...Field) {
	l.Logger.Debug(msg, fields...)
}

// Info logs an info message
func (l *logger) Info(msg string, fields ...Field) {
	l.Logger.Info(msg, fields...)
}

// Warn logs a warning message
func (l *logger) Warn(msg string, fields ...Field) {
	l.Logger.Warn(msg, fields...)
}

// Error logs an error message
func (l *logger) Error(msg string, fields ...Field) {
	l.Logger.Error(msg, fields...)
}

// Fatal logs a fatal message and exits
func (l *logger) Fatal(msg string, fields ...Field) {
	l.Logger.Fatal(msg, fields...)
}

// Sync flushes any buffered log entries
func (l *logger) Sync() error {
	return l.Logger.Sync()
}

// Global logger instance
var globalLogger Logger

// Init initializes the global logger
func Init(level string, development bool) error {
	logger, err := New(level, development)
	if err != nil {
		return err
	}
	globalLogger = logger
	return nil
}

// L returns the global logger instance
func L() Logger {
	if globalLogger == nil {
		// Initialize with default settings if not initialized
		development := os.Getenv("ENV") == "development" || os.Getenv("ENV") == "dev"
		level := os.Getenv("LOG_LEVEL")
		if level == "" {
			level = "info"
		}
		globalLogger, _ = New(level, development)
	}
	return globalLogger
}

// SetLogger sets the global logger instance
func SetLogger(l Logger) {
	globalLogger = l
}

// ContextLogger extracts logger from context or returns global logger
func ContextLogger(ctx context.Context) Logger {
	if logger, ok := ctx.Value(loggerKey).(Logger); ok {
		return logger
	}
	return L()
}

// WithLogger adds logger to context
func WithLogger(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

type contextKey string

const loggerKey contextKey = "logger"
