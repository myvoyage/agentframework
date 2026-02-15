// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
// Copyright (C) 2025  Agent Framework Contributors

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.

// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

// Additional permission under GNU Affero General Public License version 3 section 7
// If you modify this Program, or any covered work, by linking or combining it
// with other code, such other code is not for that reason alone subject to any
// of the requirements of the GNU Affero GPL version 3 as long as you maintain
// the separation between the Program and the other code.

// For network interaction purposes, when this Program is used over a network,
// the source code of the Program must be made available to users of the network.
// You can comply with this requirement by providing a link to the source code
// repository in your user interface or documentation.

// SPDX-License-Identifier: AGPL-3.0-or-later

package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

// LogLevel defines the log level type
type LogLevel string

// Log levels for the logger
const (
	LogLevelTrace   LogLevel = "TRACE" // 用于结构化推理追踪
	LogLevelDebug   LogLevel = "DEBUG"
	LogLevelInfo    LogLevel = "INFO"
	LogLevelWarning LogLevel = "WARNING"
	LogLevelError   LogLevel = "ERROR"
	LogLevelFatal   LogLevel = "FATAL"
)

// LoggerConfig contains configuration options for the logger
type LoggerConfig struct {
	Level      LogLevel  `json:"level"`       // Minimum log level to log
	Format     string    `json:"format"`      // Log format: "json" or "text"
	Output     io.Writer `json:"-"`           // Output destination
	WithCaller bool      `json:"with_caller"` // Include caller information
}

// DefaultLoggerConfig returns a default LoggerConfig
func DefaultLoggerConfig() LoggerConfig {
	return LoggerConfig{
		Level:      LogLevelInfo,
		Format:     "text",
		Output:     os.Stdout,
		WithCaller: false,
	}
}

// Logger represents a structured logger
// It provides methods for logging at different levels with context

type Logger interface {
	// Debug logs a debug message
	Debug(ctx context.Context, msg string, fields ...Field)
	// Info logs an info message
	Info(ctx context.Context, msg string, fields ...Field)
	// Warning logs a warning message
	Warning(ctx context.Context, msg string, fields ...Field)
	// Error logs an error message
	Error(ctx context.Context, msg string, fields ...Field)
	// Fatal logs a fatal message and exits
	Fatal(ctx context.Context, msg string, fields ...Field)

	// Trace logs a trace message for structured reasoning tracking
	Trace(ctx context.Context, msg string, fields ...Field)
	// Reason logs a reasoning step message
	Reason(ctx context.Context, msg string, fields ...Field)
	// Decision logs a decision message
	Decision(ctx context.Context, msg string, fields ...Field)
	// Action logs an action message
	Action(ctx context.Context, msg string, fields ...Field)
	// Result logs a result message
	Result(ctx context.Context, msg string, fields ...Field)

	// WithFields returns a new logger with the given fields
	WithFields(fields ...Field) Logger
	// WithContext returns a new logger with the given context
	WithContext(ctx context.Context) Logger
	// SetLevel sets the minimum log level
	SetLevel(level LogLevel)
	// GetLevel returns the current log level
	GetLevel() LogLevel
}

// Field represents a key-value pair for logging context
type Field struct {
	Key   string
	Value interface{}
}

// Fields creates multiple Field instances from a map
func Fields(fields map[string]interface{}) []Field {
	result := make([]Field, 0, len(fields))
	for k, v := range fields {
		result = append(result, Field{Key: k, Value: v})
	}
	return result
}

// StringField creates a Field with a string value
func StringField(key, value string) Field {
	return Field{Key: key, Value: value}
}

// IntField creates a Field with an int value
func IntField(key string, value int) Field {
	return Field{Key: key, Value: value}
}

// Int64Field creates a Field with an int64 value
func Int64Field(key string, value int64) Field {
	return Field{Key: key, Value: value}
}

// BoolField creates a Field with a bool value
func BoolField(key string, value bool) Field {
	return Field{Key: key, Value: value}
}

// ErrorField creates a Field with an error value
func ErrorField(key string, value error) Field {
	return Field{Key: key, Value: value.Error()}
}

// TimeField creates a Field with a time value
func TimeField(key string, value time.Time) Field {
	return Field{Key: key, Value: value.Format(time.RFC3339)}
}

// logger implements the Logger interface
type logger struct {
	config     LoggerConfig
	baseFields []Field
	ctx        context.Context
}

// NewLogger creates a new Logger instance with the given configuration
func NewLogger(config LoggerConfig) Logger {
	if config.Output == nil {
		config.Output = os.Stdout
	}
	if config.Format == "" {
		config.Format = "text"
	}
	if config.Level == "" {
		config.Level = LogLevelInfo
	}
	return &logger{
		config: config,
	}
}

// Global logger instance
var globalLogger Logger

// init initializes the global logger
func init() {
	globalLogger = NewLogger(DefaultLoggerConfig())
}

// GetLogger returns the global logger instance
func GetLogger() Logger {
	return globalLogger
}

// SetLogger sets the global logger instance
func SetLogger(logger Logger) {
	globalLogger = logger
}

// logEntry represents a single log entry
type logEntry struct {
	Time    string                 `json:"time"`
	Level   string                 `json:"level"`
	Message string                 `json:"message"`
	Fields  map[string]interface{} `json:"fields,omitempty"`
}

// log writes a log message at the given level
func (l *logger) log(level LogLevel, ctx context.Context, msg string, fields []Field) {
	// Check if log level is enabled
	if !l.isLevelEnabled(level) {
		return
	}

	// Create log entry
	entry := logEntry{
		Time:    time.Now().Format(time.RFC3339),
		Level:   string(level),
		Message: msg,
		Fields:  make(map[string]interface{}),
	}

	// Add base fields
	for _, field := range l.baseFields {
		entry.Fields[field.Key] = field.Value
	}

	// Add context fields
	if ctx != nil {
		// Add request ID if present
		if reqID, ok := ctx.Value("request_id").(string); ok {
			entry.Fields["request_id"] = reqID
		}
		// Add user ID if present
		if userID, ok := ctx.Value("user_id").(string); ok {
			entry.Fields["user_id"] = userID
		}
	}

	// Add log fields
	for _, field := range fields {
		entry.Fields[field.Key] = field.Value
	}

	// Write log entry
	l.writeEntry(entry)

	// Exit if fatal level
	if level == LogLevelFatal {
		os.Exit(1)
	}
}

// writeEntry writes a log entry to the output
func (l *logger) writeEntry(entry logEntry) {
	var line string
	var err error

	if l.config.Format == "json" {
		// JSON format
		jsonBytes, err := json.Marshal(entry)
		if err != nil {
			line = fmt.Sprintf("{\"time\":\"%s\",\"level\":\"ERROR\",\"message\":\"Failed to marshal log entry\",\"error\":\"%v\"}", entry.Time, err)
		} else {
			line = string(jsonBytes)
		}
	} else {
		// Text format
		line = fmt.Sprintf("%s [%s] %s", entry.Time, entry.Level, entry.Message)
		if len(entry.Fields) > 0 {
			line += " ("
			first := true
			for k, v := range entry.Fields {
				if !first {
					line += ", "
				}
				line += fmt.Sprintf("%s=%v", k, v)
				first = false
			}
			line += ")"
		}
	}

	// Write to output
	_, err = fmt.Fprintln(l.config.Output, line)
	if err != nil {
		// Fallback to standard error
		fmt.Fprintf(os.Stderr, "Failed to write log entry: %v\n", err)
	}
}

// isLevelEnabled checks if the given log level is enabled
func (l *logger) isLevelEnabled(level LogLevel) bool {
	// Define log level hierarchy
	levelOrder := map[LogLevel]int{
		LogLevelTrace:   0, // 最低级别，用于结构化推理追踪
		LogLevelDebug:   1,
		LogLevelInfo:    2,
		LogLevelWarning: 3,
		LogLevelError:   4,
		LogLevelFatal:   5,
	}

	// Check if level is enabled
	return levelOrder[level] >= levelOrder[l.config.Level]
}

// Debug logs a debug message
func (l *logger) Debug(ctx context.Context, msg string, fields ...Field) {
	l.log(LogLevelDebug, ctx, msg, fields)
}

// Info logs an info message
func (l *logger) Info(ctx context.Context, msg string, fields ...Field) {
	l.log(LogLevelInfo, ctx, msg, fields)
}

// Warning logs a warning message
func (l *logger) Warning(ctx context.Context, msg string, fields ...Field) {
	l.log(LogLevelWarning, ctx, msg, fields)
}

// Error logs an error message
func (l *logger) Error(ctx context.Context, msg string, fields ...Field) {
	l.log(LogLevelError, ctx, msg, fields)
}

// Fatal logs a fatal message and exits
func (l *logger) Fatal(ctx context.Context, msg string, fields ...Field) {
	l.log(LogLevelFatal, ctx, msg, fields)
}

// WithFields returns a new logger with the given fields
func (l *logger) WithFields(fields ...Field) Logger {
	newLogger := *l
	newLogger.baseFields = append(newLogger.baseFields, fields...)
	return &newLogger
}

// WithContext returns a new logger with the given context
func (l *logger) WithContext(ctx context.Context) Logger {
	newLogger := *l
	newLogger.ctx = ctx
	return &newLogger
}

// SetLevel sets the minimum log level
func (l *logger) SetLevel(level LogLevel) {
	l.config.Level = level
}

// GetLevel returns the current log level
func (l *logger) GetLevel() LogLevel {
	return l.config.Level
}

// Trace logs a trace message for structured reasoning tracking
func (l *logger) Trace(ctx context.Context, msg string, fields ...Field) {
	l.log(LogLevelTrace, ctx, msg, append(fields, StringField("phase", "trace")))
}

// Reason logs a reasoning step message
func (l *logger) Reason(ctx context.Context, msg string, fields ...Field) {
	l.log(LogLevelTrace, ctx, msg, append(fields, StringField("phase", "reasoning")))
}

// Decision logs a decision message
func (l *logger) Decision(ctx context.Context, msg string, fields ...Field) {
	l.log(LogLevelTrace, ctx, msg, append(fields, StringField("phase", "decision")))
}

// Action logs an action message
func (l *logger) Action(ctx context.Context, msg string, fields ...Field) {
	l.log(LogLevelTrace, ctx, msg, append(fields, StringField("phase", "action")))
}

// Result logs a result message
func (l *logger) Result(ctx context.Context, msg string, fields ...Field) {
	l.log(LogLevelTrace, ctx, msg, append(fields, StringField("phase", "result")))
}

// Helper functions for logging

// Debug logs a debug message using the global logger
func Debug(ctx context.Context, msg string, fields ...Field) {
	globalLogger.Debug(ctx, msg, fields...)
}

// Info logs an info message using the global logger
func Info(ctx context.Context, msg string, fields ...Field) {
	globalLogger.Info(ctx, msg, fields...)
}

// Warning logs a warning message using the global logger
func Warning(ctx context.Context, msg string, fields ...Field) {
	globalLogger.Warning(ctx, msg, fields...)
}

// Error logs an error message using the global logger
func Error(ctx context.Context, msg string, fields ...Field) {
	globalLogger.Error(ctx, msg, fields...)
}

// Fatal logs a fatal message using the global logger and exits
func Fatal(ctx context.Context, msg string, fields ...Field) {
	globalLogger.Fatal(ctx, msg, fields...)
}

// Trace logs a trace message for structured reasoning tracking using the global logger
func Trace(ctx context.Context, msg string, fields ...Field) {
	globalLogger.Trace(ctx, msg, fields...)
}

// Reason logs a reasoning step message using the global logger
func Reason(ctx context.Context, msg string, fields ...Field) {
	globalLogger.Reason(ctx, msg, fields...)
}

// Decision logs a decision message using the global logger
func Decision(ctx context.Context, msg string, fields ...Field) {
	globalLogger.Decision(ctx, msg, fields...)
}

// Action logs an action message using the global logger
func Action(ctx context.Context, msg string, fields ...Field) {
	globalLogger.Action(ctx, msg, fields...)
}

// Result logs a result message using the global logger
func Result(ctx context.Context, msg string, fields ...Field) {
	globalLogger.Result(ctx, msg, fields...)
}
