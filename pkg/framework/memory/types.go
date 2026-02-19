// Agent Framework - Memory Package Types
// SPDX-License-Identifier: AGPL-3.0-or-later

package memory

import (
	"context"
	"time"
)

// WorkflowState represents the state of a workflow execution
type WorkflowState struct {
	ID        string                 `json:"id"`
	WorkflowID string                `json:"workflow_id"`
	Status    string                 `json:"status"`
	Data      map[string]interface{} `json:"data"`
	CreatedAt int64                  `json:"created_at"`
	UpdatedAt int64                  `json:"updated_at"`
}

// Thread represents a conversation thread
type Thread struct {
	ID        string                 `json:"id"`
	Messages  []Message              `json:"messages"`
	Metadata  map[string]interface{} `json:"metadata"`
	CreatedAt int64                  `json:"created_at"`
	UpdatedAt int64                  `json:"updated_at"`
}

// Message represents a message in a thread
type Message struct {
	ID        string                 `json:"id"`
	Role      string                 `json:"role"` // user, assistant, system
	Content   string                 `json:"content"`
	Metadata  map[string]interface{} `json:"metadata"`
	Timestamp int64                  `json:"timestamp"`
}

// AlertHandlerID is a unique identifier for alert handlers
type AlertHandlerID string

// AlertHandler is a function that handles memory alerts
type AlertHandler func(alert Alert)

// AlertSeverity represents the severity level of an alert
type AlertSeverity string

const (
	AlertSeverityInfo    AlertSeverity = "info"
	AlertSeverityWarning AlertSeverity = "warning"
	AlertSeverityError   AlertSeverity = "error"
	AlertSeverityCritical AlertSeverity = "critical"
)

// AlertOperator represents the comparison operator for alert rules
type AlertOperator string

const (
	AlertOperatorGreaterThan AlertOperator = ">"
	AlertOperatorLessThan    AlertOperator = "<"
	AlertOperatorEquals      AlertOperator = "=="
)

// AlertRule defines when an alert should be triggered
type AlertRule struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Threshold   float64       `json:"threshold"`
	Duration    time.Duration `json:"duration"`
	Enabled     bool          `json:"enabled"`
	Severity    AlertSeverity `json:"severity"`
	Operator    AlertOperator `json:"operator"`
}

// MonitorStorage defines the storage interface for monitoring data
type MonitorStorage interface {
	SaveMetrics(ctx context.Context, metrics []Metric) error
	GetMetrics(startTime, endTime int64) ([]*MemoryStats, error)
	SaveAlert(alert Alert) error
	GetAlerts(limit int) ([]*Alert, error)
	Close() error
}

// Alert represents a memory-related alert
type Alert struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"` // leak, threshold, system
	Severity    string                 `json:"severity"` // info, warning, critical
	Message     string                 `json:"message"`
	Details     map[string]interface{} `json:"details"`
	Timestamp   int64                  `json:"timestamp"`
	Acknowledged bool                  `json:"acknowledged"`
	RuleID      string                 `json:"rule_id,omitempty"`
	Name        string                 `json:"name,omitempty"`
	Description string                 `json:"description,omitempty"`
	MetricName  string                 `json:"metric_name,omitempty"`
	MetricValue uint64                 `json:"metric_value,omitempty"`
	Threshold   uint64                 `json:"threshold,omitempty"`
	Operator    string                 `json:"operator,omitempty"`
	IsActive    bool                   `json:"is_active,omitempty"`
}

// MetricType represents the type of a metric
type MetricType string

const (
	MetricTypeGauge     MetricType = "gauge"
	MetricTypeCounter   MetricType = "counter"
	MetricTypeHistogram MetricType = "histogram"
	MetricTypeSummary   MetricType = "summary"
)

// Metric represents a single metric data point
type Metric struct {
	Name        string                 `json:"name"`
	Type        MetricType             `json:"type"`
	Value       uint64                 `json:"value"`
	Labels      map[string]string      `json:"labels"`
	Timestamp   time.Time              `json:"timestamp"`
	Description string                 `json:"description"`
}

// ErrorField creates an error field for logging
func ErrorField(key string, err error) Field {
	return Field{Key: key, Value: err.Error()}
}
