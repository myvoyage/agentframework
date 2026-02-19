// Agent Framework - Logging Tests
// SPDX-License-Identifier: AGPL-3.0-or-later

package memory

import (
	"context"
	"testing"
	"time"
)

func TestLoggingFunctions(t *testing.T) {
	// These tests just verify the logging functions don't panic
	ctx := context.Background()

	// Test Info
	Info(ctx, "test message", StringField("key", "value"))

	// Test Error
	Error(ctx, "error message", StringField("key", "value"))

	// Test Warn
	Warn(ctx, "warning message", StringField("key", "value"))

	// Test Debug
	Debug(ctx, "debug message", StringField("key", "value"))
}

func TestAllFieldTypes(t *testing.T) {
	ctx := context.Background()

	// Test all field types with Info
	Info(ctx, "test with all field types",
		StringField("string", "value"),
		IntField("int", 42),
		Int64Field("int64", int64(1234567890)),
		BoolField("bool", true),
		Float64Field("float64", 3.14),
		DurationField("duration", 5*time.Second),
		Uint64Field("uint64", uint64(1024)),
	)
}

func TestMultipleFields(t *testing.T) {
	ctx := context.Background()

	// Test with multiple fields
	Info(ctx, "test message",
		StringField("key1", "value1"),
		StringField("key2", "value2"),
		StringField("key3", "value3"),
		IntField("count", 100),
	)
}

func TestNilContext(t *testing.T) {
	// Test with nil context - should not panic
	var ctx context.Context
	Info(ctx, "test message", StringField("key", "value"))
	Error(ctx, "error message", StringField("key", "value"))
	Warn(ctx, "warning message", StringField("key", "value"))
	Debug(ctx, "debug message", StringField("key", "value"))
}
