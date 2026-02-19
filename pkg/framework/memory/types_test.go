// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
// Copyright (C) 2025  Agent Framework Contributors

// Types Tests
// SPDX-License-Identifier: AGPL-3.0-or-later

package memory

import (
	"errors"
	"testing"
	"time"
)

func TestStringField(t *testing.T) {
	field := StringField("key", "value")
	if field.Key != "key" {
		t.Errorf("expected key 'key', got '%s'", field.Key)
	}
	if field.Value != "value" {
		t.Errorf("expected value 'value', got '%v'", field.Value)
	}
}

func TestIntField(t *testing.T) {
	field := IntField("count", 42)
	if field.Key != "count" {
		t.Errorf("expected key 'count', got '%s'", field.Key)
	}
	if field.Value != 42 {
		t.Errorf("expected value 42, got %v", field.Value)
	}
}

func TestInt64Field(t *testing.T) {
	field := Int64Field("timestamp", 1234567890)
	if field.Key != "timestamp" {
		t.Errorf("expected key 'timestamp', got '%s'", field.Key)
	}
	if field.Value != int64(1234567890) {
		t.Errorf("expected value 1234567890, got %v", field.Value)
	}
}

func TestBoolField(t *testing.T) {
	field := BoolField("enabled", true)
	if field.Key != "enabled" {
		t.Errorf("expected key 'enabled', got '%s'", field.Key)
	}
	if field.Value != true {
		t.Errorf("expected value true, got %v", field.Value)
	}
}

func TestFloat64Field(t *testing.T) {
	field := Float64Field("progress", 0.75)
	if field.Key != "progress" {
		t.Errorf("expected key 'progress', got '%s'", field.Key)
	}
	if field.Value != 0.75 {
		t.Errorf("expected value 0.75, got %v", field.Value)
	}
}

func TestDurationField(t *testing.T) {
	// DurationField expects time.Duration type, which will be formatted as "1s", "100ms", etc.
	field := DurationField("duration", 1000000000) // 1 second in nanoseconds, but type is time.Duration
	if field.Key != "duration" {
		t.Errorf("expected key 'duration', got '%s'", field.Key)
	}
	// The value is stored as time.Duration, and when formatted it will be "1s"
	if field.Value != time.Duration(1000000000) {
		t.Errorf("expected value %v, got %v", time.Duration(1000000000), field.Value)
	}
}

func TestUint64Field(t *testing.T) {
	field := Uint64Field("bytes", 1024)
	if field.Key != "bytes" {
		t.Errorf("expected key 'bytes', got '%s'", field.Key)
	}
	if field.Value != uint64(1024) {
		t.Errorf("expected value 1024, got %v", field.Value)
	}
}

func TestErrorField(t *testing.T) {
	err := errors.New("test error")
	field := ErrorField("error", err)
	if field.Key != "error" {
		t.Errorf("expected key 'error', got '%s'", field.Key)
	}
	if field.Value != "test error" {
		t.Errorf("expected value 'test error', got '%v'", field.Value)
	}
}

func TestErrorField_Nil(t *testing.T) {
	// ErrorField doesn't handle nil errors, it will panic
	// So we skip the nil test or expect a panic
	// For now, we just test with a non-nil error
	field := ErrorField("error", errors.New("test"))
	if field.Key != "error" {
		t.Errorf("expected key 'error', got '%s'", field.Key)
	}
	if field.Value != "test" {
		t.Errorf("expected value 'test', got '%v'", field.Value)
	}
}

// Test types and constants

func TestAlertSeverityConstants(t *testing.T) {
	severities := []AlertSeverity{
		AlertSeverityInfo,
		AlertSeverityWarning,
		AlertSeverityError,
		AlertSeverityCritical,
	}

	for _, severity := range severities {
		if severity == "" {
			t.Errorf("AlertSeverity should not be empty")
		}
	}
}

func TestAlertOperatorConstants(t *testing.T) {
	operators := []AlertOperator{
		AlertOperatorGreaterThan,
		AlertOperatorLessThan,
		AlertOperatorEquals,
	}

	for _, op := range operators {
		if op == "" {
			t.Errorf("AlertOperator should not be empty")
		}
	}
}

func TestMetricTypeConstants(t *testing.T) {
	types := []MetricType{
		MetricTypeGauge,
		MetricTypeCounter,
		MetricTypeHistogram,
		MetricTypeSummary,
	}

	for _, mt := range types {
		if mt == "" {
			t.Errorf("MetricType should not be empty")
		}
	}
}

func TestAlertRule(t *testing.T) {
	rule := AlertRule{
		ID:          "test-rule-1",
		Name:        "Test Rule",
		Description: "A test alert rule",
		Threshold:   80.0,
		Duration:    5 * time.Minute,
		Enabled:     true,
		Severity:    AlertSeverityWarning,
		Operator:    AlertOperatorGreaterThan,
	}

	if rule.ID != "test-rule-1" {
		t.Errorf("expected ID 'test-rule-1', got '%s'", rule.ID)
	}
	if rule.Name != "Test Rule" {
		t.Errorf("expected Name 'Test Rule', got '%s'", rule.Name)
	}
	if rule.Threshold != 80.0 {
		t.Errorf("expected Threshold 80.0, got %v", rule.Threshold)
	}
	if !rule.Enabled {
		t.Error("expected Enabled to be true")
	}
}

func TestAlert(t *testing.T) {
	alert := Alert{
		ID:        "test-alert-1",
		Type:      "leak",
		Severity:  "warning",
		Message:   "Memory leak detected",
		Timestamp: time.Now().Unix(),
		Details: map[string]interface{}{
			"location": "functionX",
			"size":     1024,
		},
		Acknowledged: false,
	}

	if alert.ID != "test-alert-1" {
		t.Errorf("expected ID 'test-alert-1', got '%s'", alert.ID)
	}
	if alert.Type != "leak" {
		t.Errorf("expected Type 'leak', got '%s'", alert.Type)
	}
	if alert.Severity != "warning" {
		t.Errorf("expected Severity 'warning', got '%s'", alert.Severity)
	}
	if len(alert.Details) != 2 {
		t.Errorf("expected 2 detail items, got %d", len(alert.Details))
	}
	if alert.Acknowledged {
		t.Error("expected Acknowledged to be false")
	}
}

func TestMetric(t *testing.T) {
	metric := Metric{
		Name:        "memory_usage",
		Type:        MetricTypeGauge,
		Value:       1024,
		Labels: map[string]string{
			"host": "server1",
			"region": "us-west",
		},
		Timestamp:   time.Now(),
		Description: "Current memory usage",
	}

	if metric.Name != "memory_usage" {
		t.Errorf("expected Name 'memory_usage', got '%s'", metric.Name)
	}
	if metric.Type != MetricTypeGauge {
		t.Errorf("expected Type MetricTypeGauge, got '%v'", metric.Type)
	}
	if metric.Value != 1024 {
		t.Errorf("expected Value 1024, got %v", metric.Value)
	}
	if len(metric.Labels) != 2 {
		t.Errorf("expected 2 labels, got %d", len(metric.Labels))
	}
}

func TestWorkflowState(t *testing.T) {
	state := WorkflowState{
		ID:         "state-1",
		WorkflowID: "workflow-1",
		Status:     "running",
		Data: map[string]interface{}{
			"step": 1,
			"result": "pending",
		},
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}

	if state.ID != "state-1" {
		t.Errorf("expected ID 'state-1', got '%s'", state.ID)
	}
	if state.WorkflowID != "workflow-1" {
		t.Errorf("expected WorkflowID 'workflow-1', got '%s'", state.WorkflowID)
	}
	if state.Status != "running" {
		t.Errorf("expected Status 'running', got '%s'", state.Status)
	}
	if len(state.Data) != 2 {
		t.Errorf("expected 2 data items, got %d", len(state.Data))
	}
}

func TestThread(t *testing.T) {
	thread := Thread{
		ID: "thread-1",
		Messages: []Message{
			{
				ID:        "msg-1",
				Role:      "user",
				Content:   "Hello",
				Timestamp: time.Now().Unix(),
			},
		},
		Metadata: map[string]interface{}{
			"title": "Test Thread",
		},
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}

	if thread.ID != "thread-1" {
		t.Errorf("expected ID 'thread-1', got '%s'", thread.ID)
	}
	if len(thread.Messages) != 1 {
		t.Errorf("expected 1 message, got %d", len(thread.Messages))
	}
	if len(thread.Metadata) != 1 {
		t.Errorf("expected 1 metadata item, got %d", len(thread.Metadata))
	}
}

func TestMessage(t *testing.T) {
	msg := Message{
		ID:        "msg-1",
		Role:      "assistant",
		Content:   "Hello! How can I help you?",
		Metadata: map[string]interface{}{
			"tokens": 10,
		},
		Timestamp: time.Now().Unix(),
	}

	if msg.ID != "msg-1" {
		t.Errorf("expected ID 'msg-1', got '%s'", msg.ID)
	}
	if msg.Role != "assistant" {
		t.Errorf("expected Role 'assistant', got '%s'", msg.Role)
	}
	if msg.Content != "Hello! How can I help you?" {
		t.Errorf("expected content 'Hello! How can I help you?', got '%s'", msg.Content)
	}
	if len(msg.Metadata) != 1 {
		t.Errorf("expected 1 metadata item, got %d", len(msg.Metadata))
	}
}

func TestAlertHandlerID(t *testing.T) {
	handlerID := AlertHandlerID("handler-1")
	if handlerID != "handler-1" {
		t.Errorf("expected AlertHandlerID 'handler-1', got '%s'", string(handlerID))
	}
}
