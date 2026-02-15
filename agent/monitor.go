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
	"sync"
	"time"
)

// MetricType defines the type of metric being monitored
// It categorizes metrics based on how they should be interpreted and aggregated

type MetricType string

// Supported metric types:
// - Counter: A cumulative metric that increases over time (e.g., request count)
// - Gauge: A metric that can go up and down (e.g., memory usage)
// - Timer: A metric that measures the duration of events (e.g., request latency)
// - Histogram: A metric that samples observations and counts them in buckets
const (
	MetricTypeCounter   MetricType = "counter"   // Cumulative metric that only increases
	MetricTypeGauge     MetricType = "gauge"     // Metric that can increase or decrease
	MetricTypeTimer     MetricType = "timer"     // Metric that measures event durations
	MetricTypeHistogram MetricType = "histogram" // Metric that samples observations into buckets
)

// Metric represents a single metric value with metadata
// It provides a standardized format for all types of metrics from different monitors

type Metric struct {
	Name        string            `json:"name"`        // Name of the metric
	Type        MetricType        `json:"type"`        // Type of the metric
	Value       interface{}       `json:"value"`       // Value of the metric
	Labels      map[string]string `json:"labels"`      // Labels for filtering and aggregation
	Timestamp   time.Time         `json:"timestamp"`   // Time the metric was recorded
	Description string            `json:"description"` // Description of what the metric measures
}

// AlertSeverity defines the severity level of an alert
// It categorizes alerts based on their importance and urgency

type AlertSeverity string

// Supported alert severity levels:
// - Info: Informational alert, no action required
// - Warning: Warning alert, action may be required
// - Error: Error alert, immediate action required
// - Critical: Critical alert, system is at risk
const (
	AlertSeverityInfo     AlertSeverity = "info"
	AlertSeverityWarning  AlertSeverity = "warning"
	AlertSeverityError    AlertSeverity = "error"
	AlertSeverityCritical AlertSeverity = "critical"
)

// Alert represents a generic alert from any monitor
// It provides a standardized format for alerts from different monitors

type Alert struct {
	ID          string            `json:"id"`               // Unique identifier for the alert
	RuleID      string            `json:"rule_id"`          // ID of the rule that triggered the alert
	Name        string            `json:"name"`             // Name of the alert
	Description string            `json:"description"`      // Description of the alert
	Severity    AlertSeverity     `json:"severity"`         // Severity of the alert
	MetricName  string            `json:"metric_name"`      // Name of the metric that triggered the alert
	MetricValue interface{}       `json:"metric_value"`     // Value of the metric that triggered the alert
	Threshold   interface{}       `json:"threshold"`        // Threshold value from the rule
	Operator    string            `json:"operator"`         // Comparison operator from the rule
	Timestamp   time.Time         `json:"timestamp"`        // Time the alert was triggered
	IsActive    bool              `json:"is_active"`        // Whether the alert is still active
	MonitorName string            `json:"monitor_name"`     // Name of the monitor that generated the alert
	Labels      map[string]string `json:"labels,omitempty"` // Additional labels for the alert
}

// AlertRule defines a generic alert rule that can be applied to any metric
// It allows for flexible configuration of alert conditions

type AlertRule struct {
	ID          string            `json:"id"`               // Unique identifier for the rule
	Name        string            `json:"name"`             // Human-readable name of the rule
	Description string            `json:"description"`      // Description of the rule
	MetricName  string            `json:"metric_name"`      // Name of the metric to monitor
	Severity    AlertSeverity     `json:"severity"`         // Severity of the alert
	Threshold   interface{}       `json:"threshold"`        // Threshold value
	Operator    string            `json:"operator"`         // Comparison operator: ">", "<", ">=", "<=", "==", "!="
	Duration    time.Duration     `json:"duration"`         // Duration condition must be true before alerting
	Enabled     bool              `json:"enabled"`          // Whether the rule is enabled
	Labels      map[string]string `json:"labels,omitempty"` // Labels for filtering and grouping alerts
	// For histogram metrics, we can add additional fields
	// HistogramQuantile float64 `json:"histogram_quantile,omitempty"` // Quantile to monitor (e.g., 0.95)
}

// AlertHandlerID is a unique identifier for an alert handler

type AlertHandlerID string

// AlertHandler is a function that handles alerts
// It's called when an alert is triggered

type AlertHandler func(alert Alert)

// MonitorStorage defines an interface for persisting monitor data
// It provides methods for saving and retrieving metrics, alerts, and other monitor data

type MonitorStorage interface {
	// SaveMetrics saves multiple metrics
	SaveMetrics(ctx context.Context, metrics []Metric) error

	// GetMetrics returns metrics within a time range
	GetMetrics(ctx context.Context, startTime, endTime time.Time, filters map[string]string) ([]Metric, error)

	// SaveAlert saves an alert
	SaveAlert(ctx context.Context, alert Alert) error

	// GetAlerts returns alerts within a time range
	GetAlerts(ctx context.Context, startTime, endTime time.Time, filters map[string]string) ([]Alert, error)

	// Close closes the storage backend
	Close() error
}

// Monitor defines a generic interface for all types of monitors
// It provides a unified way to manage different monitoring components

type Monitor interface {
	// Name returns the unique name of the monitor (e.g., "memory", "cpu", "network")
	Name() string

	// Start starts the monitor, beginning to collect metrics
	Start()

	// Stop stops the monitor, ceasing metric collection
	Stop()

	// IsRunning returns whether the monitor is currently running
	IsRunning() bool

	// GetMetrics returns the current metrics from the monitor in a standardized format
	GetMetrics() []Metric

	// GetStats returns monitor-specific statistics in its native format
	GetStats() interface{}

	// AddAlertHandler adds an alert handler to be called when alerts are triggered
	AddAlertHandler(handler AlertHandler) AlertHandlerID

	// RemoveAlertHandler removes an alert handler using its unique ID
	RemoveAlertHandler(id AlertHandlerID)

	// GetAlertRules returns the list of alert rules configured for this monitor
	GetAlertRules() []AlertRule

	// AddAlertRule adds a new alert rule to the monitor
	AddAlertRule(rule AlertRule) error

	// RemoveAlertRule removes an alert rule from the monitor by ID
	RemoveAlertRule(ruleID string) error

	// UpdateAlertRule updates an existing alert rule in the monitor
	UpdateAlertRule(rule AlertRule) error
}

// MonitorManagerConfig contains configuration for the MonitorManager
// It defines global monitoring settings and initial monitors

type MonitorManagerConfig struct {
	Enabled       bool           // Whether monitoring is enabled globally
	Monitors      []Monitor      // List of monitors to manage initially
	Storage       MonitorStorage // Storage for persisting monitor data
	AlertHandlers []AlertHandler // Global alert handlers to be added to all monitors
}

// MonitorManager manages multiple monitors and provides a unified interface
// It simplifies the management of multiple monitors by providing a single entry point

type MonitorManager struct {
	config        MonitorManagerConfig            // The manager configuration
	monitors      map[string]Monitor              // Map of monitor name to monitor instance
	alertHandlers map[AlertHandlerID]AlertHandler // Global alert handlers
	nextHandlerID AlertHandlerID                  // Next available handler ID
	mu            sync.RWMutex                    // Mutex for thread safety
	storage       MonitorStorage                  // Storage for monitor data
	isRunning     bool                            // Whether the manager is running
	ctx           context.Context                 // Context for cancellation
	cancel        context.CancelFunc              // Cancellation function
}

// NewMonitorManager creates a new MonitorManager instance with the given configuration
// It initializes all provided monitors and alert handlers
func NewMonitorManager(config MonitorManagerConfig) *MonitorManager {
	ctx, cancel := context.WithCancel(context.Background())

	// Initialize map from monitors slice
	monitors := make(map[string]Monitor)
	for _, monitor := range config.Monitors {
		monitors[monitor.Name()] = monitor
	}

	// Initialize alert handlers map
	handlers := make(map[AlertHandlerID]AlertHandler)
	for i, handler := range config.AlertHandlers {
		handlers[AlertHandlerID(fmt.Sprintf("global-handler-%d", i))] = handler
	}

	return &MonitorManager{
		config:        config,
		monitors:      monitors,
		alertHandlers: handlers,
		nextHandlerID: AlertHandlerID("handler-0"),
		storage:       config.Storage,
		isRunning:     false,
		ctx:           ctx,
		cancel:        cancel,
	}
}

// Start starts all monitors managed by the MonitorManager
// It sets the running state and starts each individual monitor
func (mm *MonitorManager) Start() {
	if mm.isRunning {
		return // Already running, do nothing
	}
	mm.isRunning = true

	// Start all monitors
	for _, monitor := range mm.monitors {
		monitor.Start()
	}
}

// Stop stops all monitors managed by the MonitorManager
// It sets the running state to false, stops each monitor, and closes the storage
func (mm *MonitorManager) Stop() {
	if !mm.isRunning {
		return // Not running, do nothing
	}
	mm.isRunning = false

	// Stop all monitors
	for _, monitor := range mm.monitors {
		monitor.Stop()
	}

	// Close storage if provided
	if mm.storage != nil {
		mm.storage.Close()
	}

	// Cancel the context to stop any background operations
	mm.cancel()
}

// AddMonitor adds a new monitor to the manager
// If the manager is already running, the new monitor is started immediately
func (mm *MonitorManager) AddMonitor(monitor Monitor) {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	// Add the monitor to the map
	mm.monitors[monitor.Name()] = monitor

	// Start the monitor if manager is running
	if mm.isRunning {
		monitor.Start()
	}

	// Add all global alert handlers to the new monitor
	for _, handler := range mm.alertHandlers {
		monitor.AddAlertHandler(handler)
	}
}

// RemoveMonitor removes a monitor from the manager by name
// It stops the monitor if it's running and removes it from the map
func (mm *MonitorManager) RemoveMonitor(name string) {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	if monitor, exists := mm.monitors[name]; exists {
		monitor.Stop()
		delete(mm.monitors, name)
	}
}

// GetMonitor returns a monitor by name
// It returns the monitor and a boolean indicating if it was found
func (mm *MonitorManager) GetMonitor(name string) (Monitor, bool) {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	monitor, exists := mm.monitors[name]
	return monitor, exists
}

// GetMonitors returns all monitors managed by the MonitorManager
// It returns a copy of the monitors map to avoid external modification
func (mm *MonitorManager) GetMonitors() map[string]Monitor {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	// Return a copy to avoid external modification
	copy := make(map[string]Monitor)
	for name, monitor := range mm.monitors {
		copy[name] = monitor
	}
	return copy
}

// GetMetrics returns all metrics from all monitors
// It aggregates metrics from all monitors into a single slice
func (mm *MonitorManager) GetMetrics() []Metric {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	var allMetrics []Metric
	for _, monitor := range mm.monitors {
		metrics := monitor.GetMetrics()
		allMetrics = append(allMetrics, metrics...)
	}
	return allMetrics
}

// AddAlertHandler adds a global alert handler to all monitors
// It returns a unique ID that can be used to remove the handler later
func (mm *MonitorManager) AddAlertHandler(handler AlertHandler) AlertHandlerID {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	// Assign a unique ID to the handler
	id := mm.nextHandlerID
	mm.alertHandlers[id] = handler

	// Increment next handler ID for future use
	mm.nextHandlerID = AlertHandlerID(fmt.Sprintf("handler-%d", len(mm.alertHandlers)))

	// Add the handler to all existing monitors
	for _, monitor := range mm.monitors {
		monitor.AddAlertHandler(handler)
	}

	return id
}

// RemoveAlertHandler removes a global alert handler by ID
// It removes the handler from all monitors and from the global map
func (mm *MonitorManager) RemoveAlertHandler(id AlertHandlerID) {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	if _, exists := mm.alertHandlers[id]; exists {
		// Remove the handler from all existing monitors
		for _, monitor := range mm.monitors {
			monitor.RemoveAlertHandler(id)
		}

		// Remove the handler from the global map
		delete(mm.alertHandlers, id)
	}
}

// IsRunning returns whether the MonitorManager is running
// It indicates if the monitors are actively collecting metrics
func (mm *MonitorManager) IsRunning() bool {
	return mm.isRunning
}

// GetStorage returns the monitor storage
// It provides access to the storage backend used for persisting metrics and alerts
func (mm *MonitorManager) GetStorage() MonitorStorage {
	return mm.storage
}

// Name returns the name of the MemoryMonitor
// It implements the Monitor interface's Name method
func (m *MemoryMonitor) Name() string {
	return "memory"
}

// GetStats returns the current memory statistics in their native format
// It implements the Monitor interface's GetStats method
func (m *MemoryMonitor) GetStats() interface{} {
	return m.GetCurrentStats()
}

// GetMetrics returns memory metrics as standardized Metric objects
// It implements the Monitor interface's GetMetrics method
func (m *MemoryMonitor) GetMetrics() []Metric {
	stats := m.GetCurrentStats()

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
			Name:        "memory_total_alloc",
			Type:        MetricTypeCounter,
			Value:       stats.TotalAlloc,
			Labels:      map[string]string{"unit": "bytes"},
			Timestamp:   stats.Time,
			Description: "Total allocated memory (cumulative)",
		},
		{
			Name:        "memory_heap_alloc",
			Type:        MetricTypeGauge,
			Value:       stats.HeapAlloc,
			Labels:      map[string]string{"unit": "bytes"},
			Timestamp:   stats.Time,
			Description: "Heap memory allocated",
		},
		{
			Name:        "memory_heap_objects",
			Type:        MetricTypeGauge,
			Value:       stats.HeapObjects,
			Labels:      map[string]string{"unit": "objects"},
			Timestamp:   stats.Time,
			Description: "Number of heap objects",
		},
		{
			Name:        "memory_num_gc",
			Type:        MetricTypeCounter,
			Value:       stats.NumGC,
			Labels:      map[string]string{"unit": "cycles"},
			Timestamp:   stats.Time,
			Description: "Number of GC cycles",
		},
		{
			Name:        "memory_gc_pause",
			Type:        MetricTypeTimer,
			Value:       stats.GCPause,
			Labels:      map[string]string{"unit": "nanoseconds"},
			Timestamp:   stats.Time,
			Description: "Last GC pause time",
		},
	}

	return metrics
}

// OptimizationLevel defines the level of optimization suggestions
// It determines how aggressive or conservative the optimization suggestions should be

type OptimizationLevel string

// Supported optimization levels:
// - Conservative: Only safe, low-impact suggestions
// - Moderate: Balanced mix of safe and moderate suggestions
// - Aggressive: All possible suggestions, including high-impact ones
const (
	OptimizationLevelConservative OptimizationLevel = "conservative"
	OptimizationLevelModerate     OptimizationLevel = "moderate"
	OptimizationLevelAggressive   OptimizationLevel = "aggressive"
)

// OptimizationSuggestion represents a single optimization suggestion
// It provides detailed information about what to optimize and why

type OptimizationSuggestion struct {
	ID                    string             `json:"id"`                               // Unique identifier for the suggestion
	Category              string             `json:"category"`                         // Category of the suggestion (e.g., "memory", "cache", "performance")
	Description           string             `json:"description"`                      // Human-readable description of the suggestion
	Severity              string             `json:"severity"`                         // Severity level ("low", "medium", "high")
	Level                 OptimizationLevel  `json:"level"`                            // Optimization level required to trigger this suggestion
	Impact                string             `json:"impact"`                           // Expected impact ("low", "medium", "high")
	Effort                string             `json:"effort"`                           // Implementation effort ("low", "medium", "high")
	Metrics               map[string]float64 `json:"metrics"`                          // Metrics that triggered this suggestion
	Recommendations       []string           `json:"recommendations"`                  // Specific recommendations for implementing the suggestion
	Timestamp             time.Time          `json:"timestamp"`                        // Time the suggestion was generated
	IsImplemented         bool               `json:"is_implemented"`                   // Whether the suggestion has been implemented
	ImplementationDetails string             `json:"implementation_details,omitempty"` // Details about implementation
}

// OptimizerConfig contains configuration for the optimizer
// It defines how the optimizer should behave and what rules to apply

type OptimizerConfig struct {
	Enabled         bool              // Whether optimization suggestions are enabled
	Level           OptimizationLevel // Default optimization level
	CheckInterval   time.Duration     // Interval between optimization checks
	SuggestionLimit int               // Maximum number of suggestions to keep
	MonitorManager  *MonitorManager   // Reference to the monitor manager for accessing metrics
	Storage         MonitorStorage    // Storage for persisting suggestions
}

// Optimizer analyzes monitoring data and generates optimization suggestions
// It uses various rules to identify potential improvements

type Optimizer struct {
	config      OptimizerConfig          // Optimizer configuration
	suggestions []OptimizationSuggestion // Historical optimization suggestions
	lastCheck   time.Time                // Time of last optimization check
	mu          sync.RWMutex             // Mutex for thread safety
	ctx         context.Context          // Context for cancellation
	cancel      context.CancelFunc       // Cancellation function
}

// NewOptimizer creates a new Optimizer instance with the given configuration
// It initializes all necessary fields and returns the optimizer
func NewOptimizer(config OptimizerConfig) *Optimizer {
	ctx, cancel := context.WithCancel(context.Background())

	// Set default values if not provided
	if config.CheckInterval <= 0 {
		config.CheckInterval = 5 * time.Minute
	}
	if config.SuggestionLimit <= 0 {
		config.SuggestionLimit = 100
	}
	if config.Level == "" {
		config.Level = OptimizationLevelModerate
	}

	return &Optimizer{
		config:      config,
		suggestions: make([]OptimizationSuggestion, 0, config.SuggestionLimit),
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Start starts the optimizer's background check loop
// It begins periodically analyzing metrics and generating suggestions
func (o *Optimizer) Start() {
	if !o.config.Enabled {
		return
	}

	go o.optimizationLoop()
}

// Stop stops the optimizer's background check loop
// It cancels the context and stops any ongoing checks
func (o *Optimizer) Stop() {
	o.cancel()
}

// optimizationLoop periodically analyzes metrics and generates optimization suggestions
// It runs at the configured check interval
func (o *Optimizer) optimizationLoop() {
	ticker := time.NewTicker(o.config.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			o.analyzeMetrics()
		case <-o.ctx.Done():
			return
		}
	}
}

// analyzeMetrics analyzes all metrics from the monitor manager and generates suggestions
// It applies various rules to identify potential optimizations
func (o *Optimizer) analyzeMetrics() {
	if o.config.MonitorManager == nil {
		return
	}

	// Get all metrics from the monitor manager
	metrics := o.config.MonitorManager.GetMetrics()
	if len(metrics) == 0 {
		return
	}

	// Analyze metrics and generate suggestions
	suggestions := o.generateSuggestions(metrics)

	// Add new suggestions to history
	o.mu.Lock()
	defer o.mu.Unlock()

	for _, suggestion := range suggestions {
		o.suggestions = append(o.suggestions, suggestion)
	}

	// Keep only the most recent suggestions
	if len(o.suggestions) > o.config.SuggestionLimit {
		o.suggestions = o.suggestions[len(o.suggestions)-o.config.SuggestionLimit:]
	}

	// Save suggestions to storage if provided
	if o.config.Storage != nil {
		// Implementation would go here to save suggestions to storage
	}

	o.lastCheck = time.Now()
}

// generateSuggestions generates optimization suggestions based on metrics
// It applies various rules to identify potential improvements
func (o *Optimizer) generateSuggestions(metrics []Metric) []OptimizationSuggestion {
	var suggestions []OptimizationSuggestion

	// Group metrics by name for easier analysis
	metricsByName := make(map[string][]Metric)
	for _, metric := range metrics {
		metricsByName[metric.Name] = append(metricsByName[metric.Name], metric)
	}

	// Analyze memory metrics
	suggestions = append(suggestions, o.analyzeMemoryMetrics(metricsByName)...)

	// Analyze cache metrics if available
	suggestions = append(suggestions, o.analyzeCacheMetrics(metricsByName)...)

	// Return the generated suggestions
	return suggestions
}

// analyzeMemoryMetrics analyzes memory-related metrics and generates suggestions
// It looks for memory usage patterns that could be optimized
func (o *Optimizer) analyzeMemoryMetrics(metricsByName map[string][]Metric) []OptimizationSuggestion {
	var suggestions []OptimizationSuggestion

	// Check memory allocation patterns
	if metrics, ok := metricsByName["memory_heap_alloc"]; ok && len(metrics) > 0 {
		latestMetric := metrics[len(metrics)-1]
		heapAlloc, ok := latestMetric.Value.(uint64)
		if ok {
			// Convert bytes to MB for easier analysis
			heapAllocMB := float64(heapAlloc) / (1024 * 1024)

			// Generate suggestion if memory usage is high
			if heapAllocMB > 1000 { // > 1GB
				suggestions = append(suggestions, OptimizationSuggestion{
					ID:          fmt.Sprintf("mem-%d", time.Now().UnixNano()),
					Category:    "memory",
					Description: "High heap memory usage detected",
					Severity:    "high",
					Level:       OptimizationLevelModerate,
					Impact:      "high",
					Effort:      "medium",
					Metrics: map[string]float64{
						"heap_alloc_mb": heapAllocMB,
					},
					Recommendations: []string{
						"Review memory-intensive components for potential leaks",
						"Consider increasing garbage collection frequency",
						"Optimize data structures to reduce memory footprint",
						"Implement object pooling for frequently created objects",
					},
					Timestamp: time.Now(),
				})
			}
		}
	}

	return suggestions
}

// analyzeCacheMetrics analyzes cache-related metrics and generates suggestions
// It looks for cache usage patterns that could be optimized
func (o *Optimizer) analyzeCacheMetrics(metricsByName map[string][]Metric) []OptimizationSuggestion {
	var suggestions []OptimizationSuggestion

	// This would be implemented to analyze cache metrics when they become available
	// For now, we'll return an empty slice

	return suggestions
}

// GetSuggestions returns all optimization suggestions
// It returns a copy to avoid external modification
func (o *Optimizer) GetSuggestions() []OptimizationSuggestion {
	o.mu.RLock()
	defer o.mu.RUnlock()

	// Return a copy to avoid external modification
	suggestionsCopy := make([]OptimizationSuggestion, len(o.suggestions))
	copy(suggestionsCopy, o.suggestions)
	return suggestionsCopy
}

// GetSuggestionsByLevel returns suggestions filtered by optimization level
// It allows retrieving only suggestions appropriate for a specific optimization level
func (o *Optimizer) GetSuggestionsByLevel(level OptimizationLevel) []OptimizationSuggestion {
	o.mu.RLock()
	defer o.mu.RUnlock()

	var filtered []OptimizationSuggestion
	for _, suggestion := range o.suggestions {
		if suggestion.Level <= level {
			filtered = append(filtered, suggestion)
		}
	}
	return filtered
}

// MarkAsImplemented marks a suggestion as implemented with optional details
// It updates the suggestion's status and records implementation details
func (o *Optimizer) MarkAsImplemented(suggestionID string, details string) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	for i, suggestion := range o.suggestions {
		if suggestion.ID == suggestionID {
			o.suggestions[i].IsImplemented = true
			o.suggestions[i].ImplementationDetails = details
			o.suggestions[i].Timestamp = time.Now()
			return nil
		}
	}

	return fmt.Errorf("suggestion not found: %s", suggestionID)
}

// GetLatestSuggestions returns the most recent optimization suggestions
// It limits the results to the specified number of suggestions
func (o *Optimizer) GetLatestSuggestions(limit int) []OptimizationSuggestion {
	o.mu.RLock()
	defer o.mu.RUnlock()

	if limit <= 0 || limit > len(o.suggestions) {
		limit = len(o.suggestions)
	}

	return o.suggestions[len(o.suggestions)-limit:]
}

// AnalyzeAndSuggest performs an immediate analysis and returns suggestions
// It bypasses the background loop and provides on-demand suggestions
func (o *Optimizer) AnalyzeAndSuggest() []OptimizationSuggestion {
	if o.config.MonitorManager == nil {
		return nil
	}

	metrics := o.config.MonitorManager.GetMetrics()
	return o.generateSuggestions(metrics)
}

// AddSuggestion adds a manually created suggestion
// It allows users or other components to add custom suggestions
func (o *Optimizer) AddSuggestion(suggestion OptimizationSuggestion) {
	o.mu.Lock()
	defer o.mu.Unlock()

	o.suggestions = append(o.suggestions, suggestion)

	// Keep only the most recent suggestions
	if len(o.suggestions) > o.config.SuggestionLimit {
		o.suggestions = o.suggestions[len(o.suggestions)-o.config.SuggestionLimit:]
	}
}

// GetSuggestionByID returns a suggestion by its ID
// It returns the suggestion and a boolean indicating if it was found
func (o *Optimizer) GetSuggestionByID(id string) (OptimizationSuggestion, bool) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	for _, suggestion := range o.suggestions {
		if suggestion.ID == id {
			return suggestion, true
		}
	}

	return OptimizationSuggestion{}, false
}
