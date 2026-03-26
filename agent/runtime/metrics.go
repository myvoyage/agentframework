// Runtime Metrics - Metrics collection for agent runtime
// SPDX-License-Identifier: AGPL-3.0-or-later

package runtime

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// RuntimeMetrics collects and tracks runtime metrics
type RuntimeMetrics struct {
	// Request metrics
	requestCount     int64
	requestSuccess   int64
	requestError     int64
	totalLatency     time.Duration
	latencySamples   int64
	
	// Instance metrics
	instancesCreated int64
	instancesDestroyed int64
	instancesPooled   int64
	
	// Time series data (simplified)
	byMinute map[int64]*TimeSeriesPoint
	mu       sync.RWMutex
	
	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
}

// TimeSeriesPoint represents a time series data point
type TimeSeriesPoint struct {
	Timestamp    int64
	RequestCount int64
	ErrorCount   int64
	AvgLatency   time.Duration
}

// NewRuntimeMetrics creates a new metrics collector
func NewRuntimeMetrics() *RuntimeMetrics {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &RuntimeMetrics{
		byMinute: make(map[int64]*TimeSeriesPoint),
		ctx:      ctx,
		cancel:   cancel,
	}
}

// Start starts the metrics collector
func (m *RuntimeMetrics) Start(ctx context.Context) {
	// Start cleanup goroutine for old time series data
	go m.cleanupLoop()
}

// Stop stops the metrics collector
func (m *RuntimeMetrics) Stop() {
	m.cancel()
}

// RecordRequest records a request metric
func (m *RuntimeMetrics) RecordRequest(success bool, latency time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.requestCount++
	if success {
		m.requestSuccess++
	} else {
		m.requestError++
	}
	
	m.totalLatency += latency
	m.latencySamples++
	
	// Update time series
	minute := time.Now().Unix() / 60
	point := m.byMinute[minute]
	if point == nil {
		point = &TimeSeriesPoint{Timestamp: minute}
		m.byMinute[minute] = point
	}
	point.RequestCount++
	if !success {
		point.ErrorCount++
	}
}

// RecordInstanceCreated records an instance creation
func (m *RuntimeMetrics) RecordInstanceCreated() {
	atomic.AddInt64(&m.instancesCreated, 1)
}

// RecordInstanceDestroyed records an instance destruction
func (m *RuntimeMetrics) RecordInstanceDestroyed() {
	atomic.AddInt64(&m.instancesDestroyed, 1)
}

// RecordInstancePooled records an instance being pooled
func (m *RuntimeMetrics) RecordInstancePooled() {
	atomic.AddInt64(&m.instancesPooled, 1)
}

// GetSummary returns metrics summary
func (m *RuntimeMetrics) GetSummary() *MetricsSummary {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	avgLatency := time.Duration(0)
	if m.latencySamples > 0 {
		avgLatency = m.totalLatency / time.Duration(m.latencySamples)
	}
	
	return &MetricsSummary{
		RequestCount:        m.requestCount,
		RequestSuccess:      m.requestSuccess,
		RequestError:        m.requestError,
		AvgLatency:          avgLatency,
		InstancesCreated:    atomic.LoadInt64(&m.instancesCreated),
		InstancesDestroyed:  atomic.LoadInt64(&m.instancesDestroyed),
		InstancesPooled:     atomic.LoadInt64(&m.instancesPooled),
		SuccessRate:         calculateSuccessRate(m.requestSuccess, m.requestCount),
		ErrorRate:           calculateSuccessRate(m.requestError, m.requestCount),
	}
}

// GetTimeSeries returns time series data for a time range
func (m *RuntimeMetrics) GetTimeSeries(minutes int) []*TimeSeriesPoint {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	now := time.Now()
	start := now.Add(-time.Duration(minutes) * time.Minute)
	startMinute := start.Unix() / 60
	
	points := make([]*TimeSeriesPoint, 0)
	for minute := startMinute; minute <= now.Unix()/60; minute++ {
		if point := m.byMinute[minute]; point != nil {
			points = append(points, point)
		}
	}
	
	return points
}

// cleanupLoop removes old time series data
func (m *RuntimeMetrics) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.cleanup()
		}
	}
}

// cleanup removes time series data older than 24 hours
func (m *RuntimeMetrics) cleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	cutoff := time.Now().Add(-24 * time.Hour).Unix() / 60
	
	for minute := range m.byMinute {
		if minute < cutoff {
			delete(m.byMinute, minute)
		}
	}
}

// MetricsSummary holds metrics summary
type MetricsSummary struct {
	RequestCount       int64
	RequestSuccess     int64
	RequestError       int64
	AvgLatency         time.Duration
	InstancesCreated   int64
	InstancesDestroyed int64
	InstancesPooled    int64
	SuccessRate        float64
	ErrorRate          float64
}

// Helper functions for atomic operations (using sync/atomic)

func calculateSuccessRate(count, total int64) float64 {
	if total == 0 {
		return 0
	}
	return float64(count) / float64(total) * 100
}
