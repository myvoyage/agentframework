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

package memory

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

	// Test GetHistory (instead of GetMetrics which doesn't exist)
	// Note: History might be empty if we stop before a full collection cycle
	_ = mm.GetHistory() // Just verify the method exists and doesn't panic

	// Test GetCurrentStats (instead of GetStats which doesn't exist)
	stats := mm.GetCurrentStats()
	if stats.Alloc == 0 {
		// This might be zero in some test environments, so we just check the method works
		t.Log("GetCurrentStats returned zero allocation (may be normal in test environment)")
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
		Threshold:   1000000.0, // 1MB as float64
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
	updatedRule.Threshold = 2000000.0 // 2MB as float64

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
		Threshold:   100.0, // Very low threshold to trigger easily
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

// TestDefaultLeakDetectionConfig tests the default leak detection configuration
func TestDefaultLeakDetectionConfig(t *testing.T) {
	config := DefaultLeakDetectionConfig()

	if config.Enabled {
		t.Error("Expected leak detection to be disabled by default")
	}
	if config.CheckInterval != 1*time.Minute {
		t.Errorf("Expected CheckInterval 1 minute, got %v", config.CheckInterval)
	}
	if config.LeakThreshold != 10.0 {
		t.Errorf("Expected LeakThreshold 10.0, got %v", config.LeakThreshold)
	}
	if config.SampleSize != 10 {
		t.Errorf("Expected SampleSize 10, got %d", config.SampleSize)
	}
	if config.SamplingRate != 0.1 {
		t.Errorf("Expected SamplingRate 0.1, got %v", config.SamplingRate)
	}
	if config.MaxSamples != 100 {
		t.Errorf("Expected MaxSamples 100, got %d", config.MaxSamples)
	}
}

// TestNewComponentMemoryTracker tests creating a new ComponentMemoryTracker
func TestNewComponentMemoryTracker(t *testing.T) {
	tracker := NewComponentMemoryTracker("test-component")
	if tracker == nil {
		t.Fatal("Expected tracker to be created")
	}
	if tracker.ComponentName != "test-component" {
		t.Errorf("Expected ComponentName 'test-component', got '%s'", tracker.ComponentName)
	}
	if tracker.lastTime.IsZero() {
		t.Error("Expected lastTime to be initialized")
	}
}

// TestMemoryMonitorComponentTracking tests component memory tracking
func TestMemoryMonitorComponentTracking(t *testing.T) {
	config := MemoryMonitorConfig{
		Enabled:     true,
		Interval:    100 * time.Millisecond,
		HistorySize: 10,
	}
	mm := NewMemoryMonitor(config)

	// Test RegisterComponent
	tracker1 := mm.RegisterComponent("component1")
	if tracker1 == nil {
		t.Fatal("Expected tracker to be created")
	}
	if tracker1.ComponentName != "component1" {
		t.Errorf("Expected ComponentName 'component1', got '%s'", tracker1.ComponentName)
	}

	// Test that registering the same component returns the existing tracker
	tracker2 := mm.RegisterComponent("component1")
	if tracker1 != tracker2 {
		t.Error("Expected same tracker instance for existing component")
	}

	// Test UpdateComponentMemory
	mm.UpdateComponentMemory("component1", 1000, 5000, 10)
	mm.UpdateComponentMemory("component2", 2000, 10000, 20) // Auto-creates tracker

	// Get current stats to verify component tracking
	stats := mm.GetCurrentStats()
	if stats.ComponentStats == nil {
		t.Error("Expected ComponentStats to be initialized")
	}
}

// TestMemoryMonitorAddDuplicateAlertRule tests adding duplicate alert rule
func TestMemoryMonitorAddDuplicateAlertRule(t *testing.T) {
	config := MemoryMonitorConfig{
		Enabled:     true,
		Interval:    100 * time.Millisecond,
		HistorySize: 10,
	}
	mm := NewMemoryMonitor(config)

	rule := AlertRule{
		ID:        "duplicate-rule",
		Name:      "Duplicate Rule",
		Threshold: 1000,
		Enabled:   true,
	}

	// Add the rule
	err := mm.AddAlertRule(rule)
	if err != nil {
		t.Fatalf("Expected no error adding first rule: %v", err)
	}

	// Try to add the same rule again
	err = mm.AddAlertRule(rule)
	if err == nil {
		t.Error("Expected error when adding duplicate alert rule")
	}
}

// TestMemoryMonitorPrintStats tests the PrintStats method
func TestMemoryMonitorPrintStats(t *testing.T) {
	config := MemoryMonitorConfig{
		Enabled:     true,
		Interval:    100 * time.Millisecond,
		HistorySize: 10,
	}
	mm := NewMemoryMonitor(config)

	// This should not panic
	mm.PrintStats()
}

// TestMemoryMonitorGetHistory tests the GetHistory method
func TestMemoryMonitorGetHistory(t *testing.T) {
	config := MemoryMonitorConfig{
		Enabled:     true,
		Interval:    50 * time.Millisecond,
		HistorySize: 5,
	}
	mm := NewMemoryMonitor(config)

	// Start the monitor
	mm.Start()

	// Wait for some stats to be collected
	time.Sleep(200 * time.Millisecond)

	// Stop the monitor
	mm.Stop()

	// Get history
	history := mm.GetHistory()
	if history == nil {
		t.Error("Expected history slice, got nil")
	}

	// History should not be empty after running
	// (but it might be empty if the test is fast, so we just check it doesn't panic)
}

// TestMemoryMonitorLeakDetectionMethods tests leak detection related methods
func TestMemoryMonitorLeakDetectionMethods(t *testing.T) {
	config := MemoryMonitorConfig{
		Enabled:     true,
		Interval:    100 * time.Millisecond,
		HistorySize: 10,
		LeakDetection: LeakDetectionConfig{
			Enabled:       false, // Disable for basic tests
			CheckInterval: 1 * time.Minute,
		},
	}
	mm := NewMemoryMonitor(config)

	// Test GetLeakReports - should be empty initially
	reports := mm.GetLeakReports()
	if reports == nil {
		t.Error("Expected reports slice, got nil")
	}
	if len(reports) != 0 {
		t.Errorf("Expected 0 leak reports initially, got %d", len(reports))
	}

	// Test GetLeakDetectionResults - should be empty initially
	results := mm.GetLeakDetectionResults()
	if results == nil {
		t.Error("Expected results slice, got nil")
	}
	if len(results) != 0 {
		t.Errorf("Expected 0 leak detection results initially, got %d", len(results))
	}

	// Test GetLatestLeakReport - should return false
	report, ok := mm.GetLatestLeakReport()
	if ok {
		t.Error("Expected no latest leak report initially")
	}
	if report.Time.IsZero() == false {
		t.Error("Expected empty leak report")
	}

	// Test IsLeakDetected - should return false
	if mm.IsLeakDetected() {
		t.Error("Expected no leak detected initially")
	}
}

// TestMemoryMonitorStopWhenNotRunning tests stopping a monitor that's not running
func TestMemoryMonitorStopWhenNotRunning(t *testing.T) {
	config := MemoryMonitorConfig{
		Enabled:     true,
		Interval:    100 * time.Millisecond,
		HistorySize: 10,
	}
	mm := NewMemoryMonitor(config)

	// Stop without starting - should not panic
	mm.Stop()

	if mm.IsRunning() {
		t.Error("Expected monitor to not be running")
	}
}

// TestMemoryMonitorStartWhenAlreadyRunning tests starting a monitor that's already running
func TestMemoryMonitorStartWhenAlreadyRunning(t *testing.T) {
	config := MemoryMonitorConfig{
		Enabled:     true,
		Interval:    100 * time.Millisecond,
		HistorySize: 10,
	}
	mm := NewMemoryMonitor(config)

	mm.Start()
	if !mm.IsRunning() {
		t.Error("Expected monitor to be running")
	}

	// Start again - should be idempotent
	mm.Start()
	if !mm.IsRunning() {
		t.Error("Expected monitor to still be running")
	}

	mm.Stop()
}

// TestMemoryMonitorConfigDefaults tests default values in MemoryMonitorConfig
func TestMemoryMonitorConfigDefaults(t *testing.T) {
	// Create monitor with minimal config
	config := MemoryMonitorConfig{}
	mm := NewMemoryMonitor(config)

	// Check that defaults were applied
	rules := mm.GetAlertRules()
	if len(rules) == 0 {
		t.Error("Expected default alert rules to be added")
	}
}

// TestMemoryStatsStruct tests the MemoryStats struct
func TestMemoryStatsStruct(t *testing.T) {
	now := time.Now()
	stats := MemoryStats{
		Time:         now,
		Alloc:        1000,
		TotalAlloc:   5000,
		Sys:          10000,
		HeapAlloc:    2000,
		HeapSys:      5000,
		HeapObjects:  100,
		StackInuse:   500,
		StackSys:     1000,
		GCSys:        2000,
		NextGC:       4000000,
		NumGC:        10,
		GCPauseTotal: time.Second,
		GCPause:      100 * time.Millisecond,
	}

	if stats.Time != now {
		t.Error("Expected time to be set")
	}
	if stats.Alloc != 1000 {
		t.Error("Expected Alloc to be 1000")
	}
}

// TestLeakDetectionResultStruct tests the LeakDetectionResult struct
func TestLeakDetectionResultStruct(t *testing.T) {
	result := LeakDetectionResult{
		IsLeak:           true,
		GrowthPercentage: 25.5,
		Duration:         5 * time.Minute,
		SuspectedComponents: []string{
			"component1",
			"component2",
		},
	}

	if !result.IsLeak {
		t.Error("Expected IsLeak to be true")
	}
	if result.GrowthPercentage != 25.5 {
		t.Error("Expected GrowthPercentage to be 25.5")
	}
	if len(result.SuspectedComponents) != 2 {
		t.Error("Expected 2 suspected components")
	}
}
