// Agent Framework - Memory Package Logging Utilities
// SPDX-License-Identifier: AGPL-3.0-or-later

package memory

import (
	"context"
	"fmt"
	"log"
	"time"
)

// Field represents a log field
type Field struct {
	Key   string
	Value interface{}
}

// StringField creates a string field
func StringField(key, value string) Field {
	return Field{Key: key, Value: value}
}

// IntField creates an int field
func IntField(key string, value int) Field {
	return Field{Key: key, Value: value}
}

// Int64Field creates an int64 field
func Int64Field(key string, value int64) Field {
	return Field{Key: key, Value: value}
}

// BoolField creates a bool field
func BoolField(key string, value bool) Field {
	return Field{Key: key, Value: value}
}

// Float64Field creates a float64 field
func Float64Field(key string, value float64) Field {
	return Field{Key: key, Value: value}
}

// DurationField creates a duration field
func DurationField(key string, value time.Duration) Field {
	return Field{Key: key, Value: value}
}

// Uint64Field creates a uint64 field
func Uint64Field(key string, value uint64) Field {
	return Field{Key: key, Value: value}
}

// Info logs an info message
func Info(_ context.Context, msg string, fields ...Field) {
	logMessage("INFO", msg, fields...)
}

// Error logs an error message
func Error(_ context.Context, msg string, fields ...Field) {
	logMessage("ERROR", msg, fields...)
}

// Warn logs a warning message
func Warn(_ context.Context, msg string, fields ...Field) {
	logMessage("WARN", msg, fields...)
}

// Debug logs a debug message
func Debug(_ context.Context, msg string, fields ...Field) {
	logMessage("DEBUG", msg, fields...)
}

// logMessage is the internal logging function
func logMessage(level, msg string, fields ...Field) {
	logStr := fmt.Sprintf("[%s] %s", level, msg)
	for _, field := range fields {
		logStr += fmt.Sprintf(" %s=%v", field.Key, field.Value)
	}
	log.Println(logStr)
}
