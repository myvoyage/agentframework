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
	"fmt"
	"runtime"
	"sync"
	"time"
)

// MemoryStats represents memory usage statistics
type MemoryStats struct {
	Time         time.Time     `json:"time"`           // Timestamp of the measurement
	Alloc        uint64        `json:"alloc"`          // Current allocated memory in bytes
	TotalAlloc   uint64        `json:"total_alloc"`    // Total allocated memory in bytes (cumulative)
	Sys          uint64        `json:"sys"`            // System memory used in bytes
	Mallocs      uint64        `json:"mallocs"`        // Total number of mallocs
	Frees        uint64        `json:"frees"`          // Total number of frees
	HeapAlloc    uint64        `json:"heap_alloc"`     // Heap memory allocated
	HeapSys      uint64        `json:"heap_sys"`       // System memory for heap
	HeapObjects  uint64        `json:"heap_objects"`   // Number of heap objects
	StackInuse   uint64        `json:"stack_inuse"`    // Stack memory in use
	StackSys     uint64        `json:"stack_sys"`      // System memory for stack
	GCSys        uint64        `json:"gc_sys"`         // System memory for GC
	NextGC       uint64        `json:"next_gc"`        // Next GC target in bytes
	NumGC        uint32        `json:"num_gc"`         // Number of GC cycles
	GCPauseTotal time.Duration `json:"gc_pause_total"` // Total GC pause time
	GCPause      time.Duration `json:"gc_pause"`       // Last GC pause time
	// Enhanced memory statistics
	HeapIdle          uint64                          `json:"heap_idle"`            // Heap memory idle
	HeapInuse         uint64                          `json:"heap_inuse"`           // Heap memory in use
	HeapReleased      uint64                          `json:"heap_released"`        // Heap memory released
	HeapObjectsBySize map[string]uint64               `json:"heap_objects_by_size"` // Heap objects by size range
	ComponentStats    map[string]ComponentMemoryStats `json:"component_stats"`      // Component-specific memory statistics
}

// ComponentMemoryStats represents memory usage statistics for a specific component
type ComponentMemoryStats struct {
	Alloc          uint64    `json:"alloc"`           // Current allocated memory in bytes
	TotalAlloc     uint64    `json:"total_alloc"`     // Total allocated memory in bytes (cumulative)
	NumObjects     uint64    `json:"num_objects"`     // Number of objects
	LastUpdate     time.Time `json:"last_update"`     // Last update time
	AllocationRate float64   `json:"allocation_rate"` // Allocation rate in bytes per second
}

// ComponentMemoryTracker tracks memory usage for a specific component
type ComponentMemoryTracker struct {
	ComponentName string
	alloc         uint64
	totalAlloc    uint64
	numObjects    uint64
	lastAlloc     uint64
	lastTime      time.Time
	mu            sync.Mutex
}

// LeakDetectionConfig contains configuration for memory leak detection
type LeakDetectionConfig struct {
	Enabled            bool          // Whether leak detection is enabled
	CheckInterval      time.Duration // Interval between leak checks
	LeakThreshold      float64       // Threshold for leak detection (in percentage increase)
	LeakDuration       time.Duration // Duration of continuous increase to trigger leak alert
	SampleSize         int           // Number of samples to analyze
	ReportInterval     time.Duration // Interval between leak reports
	HeapProfileEnabled bool          // Whether to enable heap profiling
	SampleSize         int           `json:"sample_size"`
	SamplingRate       float64       `json:"sampling_rate"` // Sampling rate (0.0-1.0) to reduce overhead
	MaxSamples         int           `json:"max_samples"` // Maximum samples to keep in history
}

// LeakReport represents a memory leak report
type LeakReport struct {
	Time                time.Time     `json:"time"`                 // Timestamp of the report
	Detected            bool          `json:"detected"`             // Whether a leak was detected
	MemoryGrowth        float64       `json:"memory_growth"`        // Memory growth percentage
	Duration            time.Duration `json:"duration"`             // Duration of growth
	StartAlloc          uint64        `json:"start_alloc"`          // Starting allocation size
	EndAlloc            uint64        `json:"end_alloc"`            // Ending allocation size
	GrowthRate          float64       `json:"growth_rate"`          // Growth rate in bytes per second
	SuspectedComponents []string      `json:"suspected_components"` // Components suspected of leaking
	MemoryStatsHistory  []MemoryStats `json:"memory_stats_history"` // Memory stats history
	HeapProfileURL      string        `json:"heap_profile_url"`     // URL to heap profile (if enabled)
}

// LeakDetectionResult represents the result of a leak detection check
type LeakDetectionResult struct {
	IsLeak              bool          `json:"is_leak"`              // Whether a leak was detected
	GrowthPercentage    float64       `json:"growth_percentage"`    // Memory growth percentage
	Duration            time.Duration `json:"duration"`             // Duration of growth
	SuspectedComponents []string      `json:"suspected_components"` // Suspected leaking components
}

// alertHandlerEntry represents an entry in the alert handler list
type alertHandlerEntry struct {
	id      AlertHandlerID
	handler AlertHandler
}

// MemoryMonitorConfig contains configuration for the memory monitor
type MemoryMonitorConfig struct {
	Enabled        bool                // Whether memory monitoring is enabled
	Interval       time.Duration       // Interval between memory measurements
	HistorySize    int                 // Number of historical measurements to keep
	AlertRules     []AlertRule         // List of alert rules
	AlertInterval  time.Duration       // Minimum interval between alerts
	AlertHandlers  []alertHandlerEntry // List of alert handlers
	Storage        MonitorStorage      // Storage for monitor data
	EnabledMetrics []string            // List of enabled metrics (if empty, all metrics are enabled)
	LeakDetection  LeakDetectionConfig // Configuration for memory leak detection
}

// DefaultLeakDetectionConfig returns default leak detection configuration
func DefaultLeakDetectionConfig() LeakDetectionConfig {
	return LeakDetectionConfig{
		Enabled:            false,
		CheckInterval:      1 * time.Minute,
		LeakThreshold:      10.0, // 10% growth
		LeakDuration:       5 * time.Minute,
		SampleSize:         10,
		ReportInterval:     30 * time.Minute,
		HeapProfileEnabled: false,
		SamplingRate:       0.1,  // 10% sampling rate to reduce overhead
		MaxSamples:         100,  // Keep at most 100 samples
	}
}

// RuleViolation tracks how long a rule has been violated
type RuleViolation struct {
	StartTime time.Time
	Count     int
	LastSeen  time.Time
}

// MemoryMonitor monitors memory usage in real-time
type MemoryMonitor struct {
	config         MemoryMonitorConfig
	stats          []MemoryStats
	mu             sync.RWMutex
	stopChan       chan struct{}
	isRunning      bool
	lastAlertTime  time.Time
	ruleViolations map[string]*RuleViolation
	activeAlerts   map[string]Alert
	alertMu        sync.Mutex
	ctx            context.Context
	cancel         context.CancelFunc
	// Enhanced monitoring fields
	componentTrackers map[string]*ComponentMemoryTracker
	componentMu       sync.RWMutex
	// Leak detection fields
	leakCheckTicker      *time.Ticker          // Ticker for leak checks
	reportTicker         *time.Ticker          // Ticker for leak reports
	lastLeakCheck        time.Time             // Time of last leak check
	lastLeakReport       time.Time             // Time of last leak report
	leakDetectionResults []LeakDetectionResult // Historical leak detection results
	leakReports          []LeakReport          // Historical leak reports
	lastLeakAlert        time.Time             // Time of last leak alert
	leakChecks           int                   // Number of leak checks performed
	muLeak               sync.RWMutex          // Mutex for leak detection
}

// NewComponentMemoryTracker creates a new ComponentMemoryTracker
func NewComponentMemoryTracker(componentName string) *ComponentMemoryTracker {
	return &ComponentMemoryTracker{
		ComponentName: componentName,
		lastTime:      time.Now(),
	}
}

// NewMemoryMonitor creates a new MemoryMonitor instance
func NewMemoryMonitor(config MemoryMonitorConfig) *MemoryMonitor {
	// Set default values if not provided
	if config.Interval <= 0 {
		config.Interval = 5 * time.Second
	}
	if config.HistorySize <= 0 {
		config.HistorySize = 100
	}
	if config.AlertInterval <= 0 {
		config.AlertInterval = 1 * time.Minute
	}
	if len(config.AlertRules) == 0 {
		// Default alert rule: warn when heap alloc exceeds 512MB
		config.AlertRules = []AlertRule{
			{
				ID:          "default-heap-512mb",
				Name:        "Default Heap Usage Warning",
				Description: "Alert when heap memory usage exceeds 512MB",
				Severity:    AlertSeverityWarning,
				Threshold:   512 * 1024 * 1024, // 512MB
				Operator:    ">",
				Duration:    30 * time.Second,
				Enabled:     true,
			},
			{
				ID:          "default-heap-1gb",
				Name:        "Default Heap Usage Error",
				Description: "Alert when heap memory usage exceeds 1GB",
				Severity:    AlertSeverityError,
				Threshold:   1024 * 1024 * 1024, // 1GB
				Operator:    ">",
				Duration:    15 * time.Second,
				Enabled:     true,
			},
		}
	}

	// Set default leak detection configuration if not provided
	if config.LeakDetection.CheckInterval <= 0 {
		config.LeakDetection = DefaultLeakDetectionConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &MemoryMonitor{
		config:            config,
		stats:             make([]MemoryStats, 0, config.HistorySize),
		stopChan:          make(chan struct{}),
		isRunning:         false,
		lastAlertTime:     time.Time{},
		ruleViolations:    make(map[string]*RuleViolation),
		activeAlerts:      make(map[string]Alert),
		ctx:               ctx,
		cancel:            cancel,
		componentTrackers: make(map[string]*ComponentMemoryTracker),
		// Leak detection initialization
		leakCheckTicker:      nil,
		reportTicker:         nil,
		lastLeakCheck:        time.Time{},
		lastLeakReport:       time.Time{},
		leakDetectionResults: make([]LeakDetectionResult, 0, 100),
		leakReports:          make([]LeakReport, 0, 100),
		lastLeakAlert:        time.Time{},
		leakChecks:           0,
	}
}

// RegisterComponent registers a component for memory tracking
func (m *MemoryMonitor) RegisterComponent(componentName string) *ComponentMemoryTracker {
	m.componentMu.Lock()
	defer m.componentMu.Unlock()

	if tracker, exists := m.componentTrackers[componentName]; exists {
		return tracker
	}

	tracker := NewComponentMemoryTracker(componentName)
	m.componentTrackers[componentName] = tracker
	return tracker
}

// UpdateComponentMemory updates memory usage for a component
func (m *MemoryMonitor) UpdateComponentMemory(componentName string, alloc, totalAlloc, numObjects uint64) {
	m.componentMu.Lock()
	defer m.componentMu.Unlock()

	tracker, exists := m.componentTrackers[componentName]
	if !exists {
		// Create a new tracker if it doesn't exist
		tracker = NewComponentMemoryTracker(componentName)
		m.componentTrackers[componentName] = tracker
	}

	tracker.mu.Lock()
	defer tracker.mu.Unlock()

	// Update tracker values
	tracker.alloc = alloc
	tracker.totalAlloc = totalAlloc
	tracker.numObjects = numObjects
	tracker.lastTime = time.Now()
}

// Start starts the memory monitor
func (m *MemoryMonitor) Start() {
	if m.isRunning {
		return
	}
	m.isRunning = true

	// Start the main monitoring loop
	go m.monitorLoop()

	// Start leak detection if enabled
	if m.config.LeakDetection.Enabled {
		go m.leakDetectionLoop()
	}
}

// Stop stops the memory monitor
func (m *MemoryMonitor) Stop() {
	if !m.isRunning {
		return
	}
	m.isRunning = false

	close(m.stopChan)
	m.cancel()

	// Stop leak detection tickers if they exist
	if m.leakCheckTicker != nil {
		m.leakCheckTicker.Stop()
	}
	if m.reportTicker != nil {
		m.reportTicker.Stop()
	}

	// Close storage if provided
	if m.config.Storage != nil {
		m.config.Storage.Close()
	}
}

// leakDetectionLoop performs periodic leak detection checks
func (m *MemoryMonitor) leakDetectionLoop() {
	// Initialize tickers
	m.leakCheckTicker = time.NewTicker(m.config.LeakDetection.CheckInterval)
	defer m.leakCheckTicker.Stop()

	m.reportTicker = time.NewTicker(m.config.LeakDetection.ReportInterval)
	defer m.reportTicker.Stop()

	for {
		select {
		case <-m.leakCheckTicker.C:
			m.checkForLeaks()
		case <-m.reportTicker.C:
			m.generateLeakReport()
		case <-m.stopChan:
			return
		}
	}
}

// checkForLeaks checks for memory leaks by analyzing memory usage trends
func (m *MemoryMonitor) checkForLeaks() {
	m.mu.RLock()
	statsCount := len(m.stats)
	if statsCount < m.config.LeakDetection.SampleSize {
		m.mu.RUnlock()
		return
	}

	// Get the most recent samples
	samples := m.stats[statsCount-m.config.LeakDetection.SampleSize:]
	m.mu.RUnlock()

	if len(samples) < 2 {
		return
	}

	// Calculate memory growth
	startAlloc := samples[0].HeapAlloc
	endAlloc := samples[len(samples)-1].HeapAlloc
	growth := endAlloc - startAlloc
	growthPercentage := (float64(growth) / float64(startAlloc)) * 100
	duration := samples[len(samples)-1].Time.Sub(samples[0].Time)

	// Determine if this is a leak
	isLeak := growthPercentage > m.config.LeakDetection.LeakThreshold &&
		duration > m.config.LeakDetection.LeakDuration

	// Identify suspected components
	suspectedComponents := m.identifySuspectedComponents(samples)

	result := LeakDetectionResult{
		IsLeak:              isLeak,
		GrowthPercentage:    growthPercentage,
		Duration:            duration,
		SuspectedComponents: suspectedComponents,
	}

	// Store the result
	m.muLeak.Lock()
	m.leakDetectionResults = append(m.leakDetectionResults, result)
	// Keep only the last 100 results
	if len(m.leakDetectionResults) > 100 {
		m.leakDetectionResults = m.leakDetectionResults[len(m.leakDetectionResults)-100:]
	}
	m.muLeak.Unlock()

	// Update last check time
	m.lastLeakCheck = time.Now()

	// If a leak is detected, trigger an alert
	if isLeak {
		m.triggerLeakAlert(result)
	}
}

// identifySuspectedComponents identifies components that may be leaking
func (m *MemoryMonitor) identifySuspectedComponents(samples []MemoryStats) []string {
	// This is a simplified implementation
	// In a real implementation, this would analyze component-specific memory growth
	var suspectedComponents []string

	// For demonstration, we'll just return all components
	m.componentMu.RLock()
	for componentName := range m.componentTrackers {
		suspectedComponents = append(suspectedComponents, componentName)
	}
	m.componentMu.RUnlock()

	return suspectedComponents
}

// triggerLeakAlert triggers an alert for a detected leak
func (m *MemoryMonitor) triggerLeakAlert(result LeakDetectionResult) {
	// Rate limit alerts
	if time.Since(m.lastLeakAlert) < m.config.AlertInterval {
		return
	}

	// Create a leak alert
	alert := Alert{
		ID:          fmt.Sprintf("leak-detected-%d", time.Now().UnixNano()),
		RuleID:      "leak-detection-rule",
		Name:        "Memory Leak Detected",
		Description: fmt.Sprintf("Memory leak detected: %.2f%% growth over %v", result.GrowthPercentage, result.Duration),
		Severity:    AlertSeverityError,
		MetricName:  "heap_alloc",
		MetricValue: 0, // Not applicable for leak detection
		Threshold:   uint64(m.config.LeakDetection.LeakThreshold),
		Operator:    ">",
		Timestamp:   time.Now(),
		IsActive:    true,
	}

	// Handle the alert
	m.handleAlert(alert)
	m.lastLeakAlert = time.Now()
}

// generateLeakReport generates a comprehensive leak report
func (m *MemoryMonitor) generateLeakReport() {
	m.mu.RLock()
	statsCount := len(m.stats)
	if statsCount < m.config.LeakDetection.SampleSize {
		m.mu.RUnlock()
		return
	}

	// Get the most recent samples
	samples := m.stats[statsCount-m.config.LeakDetection.SampleSize:]
	m.mu.RUnlock()

	if len(samples) < 2 {
		return
	}

	// Calculate memory growth
	startAlloc := samples[0].HeapAlloc
	endAlloc := samples[len(samples)-1].HeapAlloc
	growth := endAlloc - startAlloc
	growthPercentage := (float64(growth) / float64(startAlloc)) * 100
	duration := samples[len(samples)-1].Time.Sub(samples[0].Time)
	growthRate := float64(growth) / duration.Seconds()

	// Identify suspected components
	suspectedComponents := m.identifySuspectedComponents(samples)

	// Create the report
	report := LeakReport{
		Time:                time.Now(),
		Detected:            growthPercentage > m.config.LeakDetection.LeakThreshold,
		MemoryGrowth:        growthPercentage,
		Duration:            duration,
		StartAlloc:          startAlloc,
		EndAlloc:            endAlloc,
		GrowthRate:          growthRate,
		SuspectedComponents: suspectedComponents,
		MemoryStatsHistory:  samples,
		HeapProfileURL:      "", // Not implemented yet
	}

	// Store the report
	m.muLeak.Lock()
	m.leakReports = append(m.leakReports, report)
	// Keep only the last 100 reports
	if len(m.leakReports) > 100 {
		m.leakReports = m.leakReports[len(m.leakReports)-100:]
	}
	m.muLeak.Unlock()

	// Update last report time
	m.lastLeakReport = time.Now()

	// Log the report
	Info(m.ctx, "Memory Leak Report Generated",
		StringField("component", "memory_monitor"),
		BoolField("leak_detected", report.Detected),
		Field{Key: "memory_growth", Value: report.MemoryGrowth},
		Field{Key: "duration", Value: report.Duration},
		IntField("suspected_components_count", len(report.SuspectedComponents)))
}

// GetLeakReports returns the historical leak reports
func (m *MemoryMonitor) GetLeakReports() []LeakReport {
	m.muLeak.RLock()
	defer m.muLeak.RUnlock()

	// Return a copy to avoid external modification
	reportsCopy := make([]LeakReport, len(m.leakReports))
	copy(reportsCopy, m.leakReports)
	return reportsCopy
}

// GetLeakDetectionResults returns the historical leak detection results
func (m *MemoryMonitor) GetLeakDetectionResults() []LeakDetectionResult {
	m.muLeak.RLock()
	defer m.muLeak.RUnlock()

	// Return a copy to avoid external modification
	resultsCopy := make([]LeakDetectionResult, len(m.leakDetectionResults))
	copy(resultsCopy, m.leakDetectionResults)
	return resultsCopy
}

// GetLatestLeakReport returns the most recent leak report
func (m *MemoryMonitor) GetLatestLeakReport() (LeakReport, bool) {
	m.muLeak.RLock()
	defer m.muLeak.RUnlock()

	if len(m.leakReports) == 0 {
		return LeakReport{}, false
	}
	return m.leakReports[len(m.leakReports)-1], true
}

// IsLeakDetected returns whether a leak is currently detected
func (m *MemoryMonitor) IsLeakDetected() bool {
	m.muLeak.RLock()
	defer m.muLeak.RUnlock()

	if len(m.leakDetectionResults) == 0 {
		return false
	}
	return m.leakDetectionResults[len(m.leakDetectionResults)-1].IsLeak
}

// monitorLoop continuously monitors memory usage
func (m *MemoryMonitor) monitorLoop() {
	ticker := time.NewTicker(m.config.Interval)
	defer ticker.Stop()

	// Initialize random source for sampling
	import "math/rand"

	for {
		select {
		case <-ticker.C:
			// Apply sampling rate to reduce monitoring overhead
			// Default to 1.0 (100%) if not set (0.0 means no sampling)
			samplingRate := m.config.LeakDetection.SamplingRate
			if samplingRate <= 0 {
				samplingRate = 1.0 // Default to full sampling
			}
			if samplingRate < 1.0 && rand.Float64() > samplingRate {
				// Skip this sample based on sampling rate
				continue
			}

			stats := m.collectMemoryStats()
			m.addStats(stats)
			m.saveStats(stats)
			m.checkAlertRules(stats)
		case <-m.stopChan:
			return
		}
	}
}

// collectMemoryStats collects memory usage statistics
func (m *MemoryMonitor) collectMemoryStats() MemoryStats {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// Calculate heap objects by size range
	heapObjectsBySize := map[string]uint64{
		"<1KB":     memStats.HeapObjects - memStats.MCacheInuse/memStats.MCacheSys,
		"1KB-4KB":  0, // This would require more detailed heap profiling
		"4KB-16KB": 0,
		">16KB":    0,
	}

	// Collect component-specific memory stats
	componentStats := make(map[string]ComponentMemoryStats)

	m.componentMu.RLock()
	for componentName, tracker := range m.componentTrackers {
		tracker.mu.Lock()
		// Calculate allocation rate
		now := time.Now()
		timeElapsed := now.Sub(tracker.lastTime).Seconds()
		allocationRate := float64(tracker.alloc-tracker.lastAlloc) / timeElapsed
		if timeElapsed == 0 {
			allocationRate = 0
		}

		componentStats[componentName] = ComponentMemoryStats{
			Alloc:          tracker.alloc,
			TotalAlloc:     tracker.totalAlloc,
			NumObjects:     tracker.numObjects,
			LastUpdate:     now,
			AllocationRate: allocationRate,
		}
		tracker.mu.Unlock()
	}
	m.componentMu.RUnlock()

	return MemoryStats{
		Time:         time.Now(),
		Alloc:        memStats.Alloc,
		TotalAlloc:   memStats.TotalAlloc,
		Sys:          memStats.Sys,
		Mallocs:      memStats.Mallocs,
		Frees:        memStats.Frees,
		HeapAlloc:    memStats.HeapAlloc,
		HeapSys:      memStats.HeapSys,
		HeapObjects:  memStats.HeapObjects,
		StackInuse:   memStats.StackInuse,
		StackSys:     memStats.StackSys,
		GCSys:        memStats.GCSys,
		NextGC:       memStats.NextGC,
		NumGC:        memStats.NumGC,
		GCPauseTotal: time.Duration(memStats.PauseTotalNs),
		GCPause:      time.Duration(getLastGCPause(&memStats)),
		// Enhanced memory statistics
		HeapIdle:          memStats.HeapIdle,
		HeapInuse:         memStats.HeapInuse,
		HeapReleased:      memStats.HeapReleased,
		HeapObjectsBySize: heapObjectsBySize,
		ComponentStats:    componentStats,
	}
}

// getLastGCPause returns the last GC pause time in nanoseconds
func getLastGCPause(memStats *runtime.MemStats) uint64 {
	if memStats.NumGC == 0 {
		return 0
	}
	// Get the last GC pause time from the PauseNs array
	idx := (memStats.NumGC - 1) % 256
	return memStats.PauseNs[idx]
}

// addStats adds a new set of memory statistics to the history
func (m *MemoryMonitor) addStats(stats MemoryStats) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Add the new stats
	m.stats = append(m.stats, stats)

	// Keep only the most recent stats
	if len(m.stats) > m.config.HistorySize {
		m.stats = m.stats[len(m.stats)-m.config.HistorySize:]
	}
}

// saveStats saves memory statistics to storage if provided
func (m *MemoryMonitor) saveStats(stats MemoryStats) {
	if m.config.Storage != nil {
		go func() {
			// Convert MemoryStats to Metric format
			metrics := []Metric{
				{
					Name:        "memory_alloc",
					Type:        MetricTypeGauge,
					Value:       stats.Alloc,
					Labels:      map[string]string{"unit": "bytes"},
					Timestamp:   stats.Time,
					Description: "Current allocated memory",
				},
				{
					Name:        "memory_heap_alloc",
					Type:        MetricTypeGauge,
					Value:       stats.HeapAlloc,
					Labels:      map[string]string{"unit": "bytes"},
					Timestamp:   stats.Time,
					Description: "Current heap memory allocation",
				},
			}

			// Save using SaveMetrics method from MonitorStorage interface
			if err := m.config.Storage.SaveMetrics(m.ctx, metrics); err != nil {
				// Log the error using the global logger
				Error(m.ctx, "Failed to save memory statistics to storage",
					ErrorField("error", err),
					StringField("component", "memory_monitor"))
			}
		}()
	}
}

// checkAlertRules checks all alert rules against current stats
func (m *MemoryMonitor) checkAlertRules(stats MemoryStats) {
	now := time.Now()

	// Iterate through all alert rules
	for _, rule := range m.config.AlertRules {
		if !rule.Enabled {
			continue
		}

		// Check if the rule condition is met for heap allocation
		// For now, we're only checking heap allocation, but could extend to other metrics
		isViolated := false

		// Convert Threshold to uint64 for comparison
		threshold, ok := rule.Threshold.(uint64)
		if !ok {
			// If threshold is not a uint64, skip this rule
			continue
		}

		switch rule.Operator {
		case ">":
			isViolated = stats.HeapAlloc > threshold
		case ">=":
			isViolated = stats.HeapAlloc >= threshold
		case "<":
			isViolated = stats.HeapAlloc < threshold
		case "<=":
			isViolated = stats.HeapAlloc <= threshold
		}

		m.alertMu.Lock()
		violation, exists := m.ruleViolations[rule.ID]

		if isViolated {
			// Rule is violated
			if !exists {
				// First violation, start tracking
				m.ruleViolations[rule.ID] = &RuleViolation{
					StartTime: now,
					Count:     1,
					LastSeen:  now,
				}
			} else {
				// Update existing violation
				violation.Count++
				violation.LastSeen = now

				// Check if violation duration meets the rule's duration requirement
				if now.Sub(violation.StartTime) >= rule.Duration {
					// Check if this alert is already active
					if _, active := m.activeAlerts[rule.ID]; !active {
						// Create and trigger alert
						alert := Alert{
							ID:          fmt.Sprintf("%s-%d", rule.ID, now.UnixNano()),
							RuleID:      rule.ID,
							Name:        rule.Name,
							Description: rule.Description,
							Severity:    rule.Severity,
							MetricName:  "heap_alloc",
							MetricValue: stats.HeapAlloc,
							Threshold:   rule.Threshold,
							Operator:    rule.Operator,
							Timestamp:   now,
							IsActive:    true,
						}

						// Add to active alerts
						m.activeAlerts[rule.ID] = alert

						// Handle the alert
						m.handleAlert(alert)
					}
				}
			}
		} else {
			// Rule is not violated
			if exists {
				// Check if there was an active alert for this rule
				if alert, active := m.activeAlerts[rule.ID]; active {
					// Deactivate the alert
					alert.IsActive = false
					alert.Timestamp = now

					// Update the alert
					m.activeAlerts[rule.ID] = alert
					m.handleAlert(alert)

					// Remove from active alerts after handling
					delete(m.activeAlerts, rule.ID)
				}

				// Remove the violation
				delete(m.ruleViolations, rule.ID)
			}
		}

		m.alertMu.Unlock()
	}
}

// handleAlert handles an alert by calling all registered alert handlers
func (m *MemoryMonitor) handleAlert(alert Alert) {
	// Log the alert
	logLevel := LogLevelInfo
	switch alert.Severity {
	case AlertSeverityWarning:
		logLevel = LogLevelWarning
	case AlertSeverityError, AlertSeverityCritical:
		logLevel = LogLevelError
	}

	// Log the alert details
	logger := GetLogger()
	message := fmt.Sprintf("Memory alert triggered: %s", alert.Name)

	// Convert metric values to appropriate types for logging
	var metricValue int64
	var threshold int64

	if mv, ok := alert.MetricValue.(uint64); ok {
		metricValue = int64(mv)
	} else if mv, ok := alert.MetricValue.(int64); ok {
		metricValue = mv
	} else if mv, ok := alert.MetricValue.(float64); ok {
		metricValue = int64(mv)
	}

	if t, ok := alert.Threshold.(uint64); ok {
		threshold = int64(t)
	} else if t, ok := alert.Threshold.(int64); ok {
		threshold = t
	} else if t, ok := alert.Threshold.(float64); ok {
		threshold = int64(t)
	}

	fields := []Field{
		StringField("alert_id", alert.ID),
		StringField("rule_id", alert.RuleID),
		StringField("severity", string(alert.Severity)),
		Int64Field("metric_value", metricValue),
		Int64Field("threshold", threshold),
		StringField("operator", alert.Operator),
		BoolField("is_active", alert.IsActive),
		StringField("component", "memory_monitor"),
	}

	switch logLevel {
	case LogLevelTrace:
		logger.Trace(m.ctx, message, fields...)
	case LogLevelDebug:
		logger.Debug(m.ctx, message, fields...)
	case LogLevelInfo:
		logger.Info(m.ctx, message, fields...)
	case LogLevelWarning:
		logger.Warning(m.ctx, message, fields...)
	case LogLevelError:
		logger.Error(m.ctx, message, fields...)
	case LogLevelFatal:
		logger.Fatal(m.ctx, message, fields...)
	}

	// Save alert to storage if provided
	if m.config.Storage != nil {
		go func() {
			if err := m.config.Storage.SaveAlert(m.ctx, alert); err != nil {
				// Log the error using the global logger
				Error(m.ctx, "Failed to save memory alert to storage",
					ErrorField("error", err),
					StringField("alert_id", alert.ID),
					StringField("component", "memory_monitor"))
			}
		}()
	}

	// Call all alert handlers
	m.mu.RLock()
	handlers := make([]alertHandlerEntry, len(m.config.AlertHandlers))
	copy(handlers, m.config.AlertHandlers)
	m.mu.RUnlock()

	for _, entry := range handlers {
		go func(handler AlertHandler) {
			defer func() {
				if r := recover(); r != nil {
					Error(m.ctx, "Alert handler panicked",
						StringField("panic", fmt.Sprintf("%v", r)),
						StringField("component", "memory_monitor"))
				}
			}()
			handler(alert)
		}(entry.handler)
	}
}

// GetCurrentStats returns the current memory statistics
func (m *MemoryMonitor) GetCurrentStats() MemoryStats {
	return m.collectMemoryStats()
}

// GetHistory returns the historical memory statistics
func (m *MemoryMonitor) GetHistory() []MemoryStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy to avoid external modification
	statsCopy := make([]MemoryStats, len(m.stats))
	copy(statsCopy, m.stats)
	return statsCopy
}

// AddAlertHandler adds an alert handler to the memory monitor
// It returns a unique ID that can be used to remove the handler later
func (m *MemoryMonitor) AddAlertHandler(handler AlertHandler) AlertHandlerID {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Generate a unique ID
	id := AlertHandlerID(fmt.Sprintf("handler-%d-%d", time.Now().UnixNano(), len(m.config.AlertHandlers)))

	// Add the handler entry
	entry := alertHandlerEntry{
		id:      id,
		handler: handler,
	}
	m.config.AlertHandlers = append(m.config.AlertHandlers, entry)

	return id
}

// RemoveAlertHandler removes an alert handler from the memory monitor by its ID
func (m *MemoryMonitor) RemoveAlertHandler(id AlertHandlerID) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, entry := range m.config.AlertHandlers {
		if entry.id == id {
			m.config.AlertHandlers = append(m.config.AlertHandlers[:i], m.config.AlertHandlers[i+1:]...)
			break
		}
	}
}

// GetAlertRules returns the list of alert rules
func (m *MemoryMonitor) GetAlertRules() []AlertRule {
	m.mu.RLock()
	defer m.mu.RUnlock()
	// Return a copy to avoid external modification
	rulesCopy := make([]AlertRule, len(m.config.AlertRules))
	copy(rulesCopy, m.config.AlertRules)
	return rulesCopy
}

// GetActiveAlerts returns the list of active alerts
func (m *MemoryMonitor) GetActiveAlerts() []Alert {
	m.alertMu.Lock()
	defer m.alertMu.Unlock()
	// Return a copy to avoid external modification
	alerts := make([]Alert, 0, len(m.activeAlerts))
	for _, alert := range m.activeAlerts {
		alerts = append(alerts, alert)
	}
	return alerts
}

// AddAlertRule adds a new alert rule to the memory monitor
func (m *MemoryMonitor) AddAlertRule(rule AlertRule) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if rule with the same ID already exists
	for _, existingRule := range m.config.AlertRules {
		if existingRule.ID == rule.ID {
			return fmt.Errorf("alert rule with ID %s already exists", rule.ID)
		}
	}

	// Add the new rule
	m.config.AlertRules = append(m.config.AlertRules, rule)
	return nil
}

// RemoveAlertRule removes an alert rule from the memory monitor by ID
func (m *MemoryMonitor) RemoveAlertRule(ruleID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Find and remove the rule
	for i, rule := range m.config.AlertRules {
		if rule.ID == ruleID {
			m.config.AlertRules = append(m.config.AlertRules[:i], m.config.AlertRules[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("alert rule with ID %s not found", ruleID)
}

// UpdateAlertRule updates an existing alert rule in the memory monitor
func (m *MemoryMonitor) UpdateAlertRule(rule AlertRule) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Find and update the rule
	for i, existingRule := range m.config.AlertRules {
		if existingRule.ID == rule.ID {
			m.config.AlertRules[i] = rule
			return nil
		}
	}

	return fmt.Errorf("alert rule with ID %s not found", rule.ID)
}

// IsRunning returns whether the memory monitor is running
func (m *MemoryMonitor) IsRunning() bool {
	return m.isRunning
}

// PrintStats prints the current memory statistics in a human-readable format
func (m *MemoryMonitor) PrintStats() {
	stats := m.GetCurrentStats()
	m.printStats(stats)
}

// printStats prints memory statistics in a human-readable format
func (m *MemoryMonitor) printStats(stats MemoryStats) {
	fmt.Printf("Memory Usage at %s:\n", stats.Time.Format("2006-01-02 15:04:05"))
	fmt.Printf("  Allocated: %s (%.2f MB)\n", formatBytes(stats.Alloc), float64(stats.Alloc)/(1024*1024))
	fmt.Printf("  Total Allocated: %s (%.2f MB)\n", formatBytes(stats.TotalAlloc), float64(stats.TotalAlloc)/(1024*1024))
	fmt.Printf("  System: %s (%.2f MB)\n", formatBytes(stats.Sys), float64(stats.Sys)/(1024*1024))
	fmt.Printf("  Heap Allocated: %s (%.2f MB)\n", formatBytes(stats.HeapAlloc), float64(stats.HeapAlloc)/(1024*1024))
	fmt.Printf("  Heap Objects: %d\n", stats.HeapObjects)
	fmt.Printf("  Stack Inuse: %s (%.2f MB)\n", formatBytes(stats.StackInuse), float64(stats.StackInuse)/(1024*1024))
	fmt.Printf("  Next GC: %s (%.2f MB)\n", formatBytes(stats.NextGC), float64(stats.NextGC)/(1024*1024))
	fmt.Printf("  Number of GC cycles: %d\n", stats.NumGC)
	fmt.Printf("  Total GC pause: %v\n", stats.GCPauseTotal)
	fmt.Printf("  Last GC pause: %v\n", stats.GCPause)
}

// formatBytes formats bytes into a human-readable string
func formatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	prefixes := []string{"KB", "MB", "GB", "TB"}
	return fmt.Sprintf("%.1f %s", float64(b)/float64(div), prefixes[exp])
}

// Global memory monitor instance
var globalMemoryMonitor *MemoryMonitor
var globalMonitorOnce sync.Once

// InitGlobalMemoryMonitor initializes the global memory monitor
func InitGlobalMemoryMonitor(config MemoryMonitorConfig) {
	globalMonitorOnce.Do(func() {
		globalMemoryMonitor = NewMemoryMonitor(config)
		globalMemoryMonitor.Start()
	})
}

// GetGlobalMemoryMonitor returns the global memory monitor instance
func GetGlobalMemoryMonitor() *MemoryMonitor {
	if globalMemoryMonitor == nil {
		// Initialize with default config if not already initialized
		InitGlobalMemoryMonitor(MemoryMonitorConfig{
			Enabled: true,
		})
	}
	return globalMemoryMonitor
}

// ShutdownGlobalMemoryMonitor shuts down the global memory monitor
func ShutdownGlobalMemoryMonitor() {
	if globalMemoryMonitor != nil {
		globalMemoryMonitor.Stop()
	}
}
