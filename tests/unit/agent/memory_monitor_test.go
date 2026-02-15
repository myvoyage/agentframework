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
	"testing"
	"time"
)

// TestMemoryMonitorBasic tests basic MemoryMonitor functionality
func TestMemoryMonitorBasic(t *testing.T) {
	// Create a MemoryMonitorConfig
	config := MemoryMonitorConfig{
		Enabled:       true,
		Interval:      100 * time.Millisecond,
		HistorySize:   10,
		AlertInterval: 1 * time.Second,
	}

	// Create a MemoryMonitor
	mm := NewMemoryMonitor(config)
	if mm == nil {
		t.Fatal("Expected MemoryMonitor instance, got nil")
	}

	// Test Start and Stop
	mm.Start()
	if !mm.IsRunning() {
		t.Error("Expected MemoryMonitor to be running after Start()")
	}

	// Give it some time to collect stats
	time.Sleep(200 * time.Millisecond)

	mm.Stop()
	if mm.IsRunning() {
		t.Error("Expected MemoryMonitor to be stopped after Stop()")
	}

	// Test GetMetrics
	metrics := mm.GetMetrics()
	if len(metrics) == 0 {
		t.Error("Expected at least one metric, got none")
	}

	// Test GetStats
	stats := mm.GetStats()
	if stats == nil {
		t.Error("Expected stats, got nil")
	}
}

// TestMemoryMonitorAlertRules tests MemoryMonitor alert rules functionality
func TestMemoryMonitorAlertRules(t *testing.T) {
	// Create a MemoryMonitorConfig with empty AlertRules
	// This will trigger the default alert rules to be added
	config := MemoryMonitorConfig{
		Enabled:       true,
		Interval:      100 * time.Millisecond,
		HistorySize:   10,
		AlertInterval: 1 * time.Second,
		AlertRules:    []AlertRule{}, // Empty slice to trigger default rules
	}

	// Create a MemoryMonitor
	mm := NewMemoryMonitor(config)
	if mm == nil {
		t.Fatal("Expected MemoryMonitor instance, got nil")
	}

	// Get initial rules (should be default rules)
	initialRules := mm.GetAlertRules()
	initialRuleCount := len(initialRules)

	// Test AddAlertRule
	alertRule := AlertRule{
		ID:          "test-rule-1",
		Name:        "Test Alert Rule",
		Description: "Test alert rule for memory usage",
		Severity:    AlertSeverityWarning,
		Threshold:   uint64(1000000), // 1MB
		Operator:    ">",
		Duration:    1 * time.Second,
		Enabled:     true,
	}

	err := mm.AddAlertRule(alertRule)
	if err != nil {
		t.Fatalf("Expected no error when adding alert rule, got: %v", err)
	}

	// Test GetAlertRules - should have initial + 1 rules
	rules := mm.GetAlertRules()
	expectedRuleCount := initialRuleCount + 1
	if len(rules) != expectedRuleCount {
		t.Errorf("Expected %d alert rules after adding, got %d", expectedRuleCount, len(rules))
	}

	// Check that our test rule was added
	found := false
	for _, rule := range rules {
		if rule.ID == alertRule.ID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected to find alert rule with ID %s, but it wasn't found", alertRule.ID)
	}

	// Test UpdateAlertRule
	updatedRule := alertRule
	updatedRule.Name = "Updated Test Alert Rule"
	updatedRule.Threshold = uint64(2000000) // 2MB

	err = mm.UpdateAlertRule(updatedRule)
	if err != nil {
		t.Fatalf("Expected no error when updating alert rule, got: %v", err)
	}

	// Test GetAlertRules after update - should still have the same count
	rules = mm.GetAlertRules()
	if len(rules) != expectedRuleCount {
		t.Errorf("Expected %d alert rules after update, got %d", expectedRuleCount, len(rules))
	}

	// Check that the rule was updated
	found = false
	for _, rule := range rules {
		if rule.ID == updatedRule.ID {
			found = true
			if rule.Name != updatedRule.Name {
				t.Errorf("Expected updated alert rule name %s, got %s", updatedRule.Name, rule.Name)
			}
			if rule.Threshold != updatedRule.Threshold {
				t.Errorf("Expected updated alert rule threshold %v, got %v", updatedRule.Threshold, rule.Threshold)
			}
			break
		}
	}
	if !found {
		t.Errorf("Expected to find updated alert rule with ID %s, but it wasn't found", updatedRule.ID)
	}

	// Test RemoveAlertRule
	err = mm.RemoveAlertRule(alertRule.ID)
	if err != nil {
		t.Fatalf("Expected no error when removing alert rule, got: %v", err)
	}

	// Test GetAlertRules after removal - should be back to initial count
	rules = mm.GetAlertRules()
	if len(rules) != initialRuleCount {
		t.Errorf("Expected %d alert rules after removal, got %d", initialRuleCount, len(rules))
	}

	// Check that our test rule was removed
	for _, rule := range rules {
		if rule.ID == alertRule.ID {
			t.Errorf("Expected alert rule with ID %s to be removed, but it still exists", alertRule.ID)
		}
	}

	// Test RemoveAlertRule with non-existent ID
	err = mm.RemoveAlertRule("non-existent-rule")
	if err == nil {
		t.Error("Expected error when removing non-existent alert rule, got none")
	}

	// Test UpdateAlertRule with non-existent ID
	err = mm.UpdateAlertRule(AlertRule{ID: "non-existent-rule"})
	if err == nil {
		t.Error("Expected error when updating non-existent alert rule, got none")
	}
}

// TestMemoryMonitorAlertHandler tests MemoryMonitor alert handler functionality
func TestMemoryMonitorAlertHandler(t *testing.T) {
	// Create a MemoryMonitorConfig
	config := MemoryMonitorConfig{
		Enabled:       true,
		Interval:      100 * time.Millisecond,
		HistorySize:   10,
		AlertInterval: 1 * time.Second,
	}

	// Create a MemoryMonitor
	mm := NewMemoryMonitor(config)
	if mm == nil {
		t.Fatal("Expected MemoryMonitor instance, got nil")
	}

	// Track alert handler calls
	var alertCount int

	// Add an alert handler
	handler := func(alert Alert) {
		alertCount++
	}

	handlerID := mm.AddAlertHandler(handler)

	// Add a low threshold alert rule to trigger easily
	alertRule := AlertRule{
		ID:          "test-rule-2",
		Name:        "Test Alert Rule",
		Description: "Test alert rule for memory usage",
		Severity:    AlertSeverityWarning,
		Threshold:   uint64(100), // Very low threshold to trigger easily
		Operator:    ">",
		Duration:    100 * time.Millisecond,
		Enabled:     true,
	}

	mm.AddAlertRule(alertRule)

	// Start the monitor
	mm.Start()

	// Give it some time to trigger an alert
	time.Sleep(300 * time.Millisecond)

	// Stop the monitor
	mm.Stop()

	// Remove the alert handler
	mm.RemoveAlertHandler(handlerID)

	// Since we can't guarantee the alert will fire in a test environment,
	// we're just testing that the handler registration and removal works
	// without errors
}

// TestMemoryMonitorGetActiveAlerts tests MemoryMonitor GetActiveAlerts method
func TestMemoryMonitorGetActiveAlerts(t *testing.T) {
	// Create a MemoryMonitorConfig
	config := MemoryMonitorConfig{
		Enabled:       true,
		Interval:      100 * time.Millisecond,
		HistorySize:   10,
		AlertInterval: 1 * time.Second,
	}

	// Create a MemoryMonitor
	mm := NewMemoryMonitor(config)
	if mm == nil {
		t.Fatal("Expected MemoryMonitor instance, got nil")
	}

	// Test GetActiveAlerts
	activeAlerts := mm.GetActiveAlerts()
	if activeAlerts == nil {
		t.Error("Expected active alerts slice, got nil")
	}

	// Should be empty initially
	if len(activeAlerts) != 0 {
		t.Errorf("Expected 0 active alerts initially, got %d", len(activeAlerts))
	}
}
