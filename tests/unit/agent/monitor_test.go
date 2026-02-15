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

func TestMonitorManager(t *testing.T) {
	// Create a memory monitor
	memoryMonitorConfig := MemoryMonitorConfig{
		Enabled:       true,
		Interval:      1 * time.Second,
		HistorySize:   10,
		AlertInterval: 5 * time.Second,
	}
	memoryMonitor := NewMemoryMonitor(memoryMonitorConfig)

	// Create a monitor manager
	monitorManager := NewMonitorManager(MonitorManagerConfig{
		Enabled:       true,
		Monitors:      []Monitor{memoryMonitor},
		Storage:       nil,
		AlertHandlers: []AlertHandler{},
	})

	// Test IsRunning
	if monitorManager.IsRunning() {
		t.Error("Expected monitor manager to be not running initially")
	}

	// Test Start
	monitorManager.Start()
	if !monitorManager.IsRunning() {
		t.Error("Expected monitor manager to be running after Start()")
	}

	// Test GetMonitors
	monitors := monitorManager.GetMonitors()
	if len(monitors) != 1 {
		t.Errorf("Expected 1 monitor, got %d", len(monitors))
	}

	// Test GetMonitor
	memMonitor, exists := monitorManager.GetMonitor("memory")
	if !exists {
		t.Error("Expected to find 'memory' monitor")
	}
	if memMonitor.Name() != "memory" {
		t.Errorf("Expected monitor name 'memory', got '%s'", memMonitor.Name())
	}

	// Test GetMetrics
	metrics := monitorManager.GetMetrics()
	if len(metrics) == 0 {
		t.Error("Expected at least 1 metric, got 0")
	}

	// Test AddAlertHandler and RemoveAlertHandler
	alertCount := 0
	alertHandler := func(alert Alert) {
		alertCount++
	}

	handlerID := monitorManager.AddAlertHandler(alertHandler)
	if handlerID == "" {
		t.Error("Expected non-empty handler ID")
	}

	monitorManager.RemoveAlertHandler(handlerID)

	// Test Stop
	monitorManager.Stop()
	if monitorManager.IsRunning() {
		t.Error("Expected monitor manager to be not running after Stop()")
	}
}

func TestMonitorInterface(t *testing.T) {
	// Create a memory monitor
	memoryMonitorConfig := MemoryMonitorConfig{
		Enabled:       true,
		Interval:      1 * time.Second,
		HistorySize:   10,
		AlertInterval: 5 * time.Second,
	}
	memoryMonitor := NewMemoryMonitor(memoryMonitorConfig)

	// Test Name()
	if memoryMonitor.Name() != "memory" {
		t.Errorf("Expected monitor name 'memory', got '%s'", memoryMonitor.Name())
	}

	// Test IsRunning()
	if memoryMonitor.IsRunning() {
		t.Error("Expected monitor to be not running initially")
	}

	// Test Start()
	memoryMonitor.Start()
	if !memoryMonitor.IsRunning() {
		t.Error("Expected monitor to be running after Start()")
	}

	// Test GetMetrics()
	metrics := memoryMonitor.GetMetrics()
	if len(metrics) == 0 {
		t.Error("Expected at least 1 metric, got 0")
	}

	// Test GetStats()
	stats := memoryMonitor.GetStats()
	if stats == nil {
		t.Error("Expected non-nil stats")
	}

	// Test GetAlertRules()
	rules := memoryMonitor.GetAlertRules()
	if len(rules) == 0 {
		t.Error("Expected at least 1 alert rule, got 0")
	}

	// Test AddAlertHandler and RemoveAlertHandler
	alertCount := 0
	alertHandler := func(alert Alert) {
		alertCount++
	}

	handlerID := memoryMonitor.AddAlertHandler(alertHandler)
	if handlerID == "" {
		t.Error("Expected non-empty handler ID")
	}

	memoryMonitor.RemoveAlertHandler(handlerID)

	// Test Stop()
	memoryMonitor.Stop()
	if memoryMonitor.IsRunning() {
		t.Error("Expected monitor to be not running after Stop()")
	}
}
