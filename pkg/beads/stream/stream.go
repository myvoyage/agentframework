// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later


// Package stream provides high-performance real-time data processing capabilities.
package stream

import (
	"context"
	"errors"
	"sync"
	"time"
)

// DataProcessor defines the interface for data processors.
type DataProcessor interface {
	Process(ctx context.Context, data interface{}) (interface{}, error)
}

// DataPipeline represents a data processing pipeline.
type DataPipeline struct {
	input          chan interface{}
	output         chan interface{}
	processors     []DataProcessor
	workers        int
	bufferSize     int
	done           chan struct{}
	wg             sync.WaitGroup
	errorHandler   func(error)
	metrics        *PipelineMetrics
	metricsMutex   sync.RWMutex
}

// PipelineMetrics contains pipeline performance metrics.
type PipelineMetrics struct {
	ProcessedCount int64
	ErrorCount     int64
	AverageLatency time.Duration
	StartTime      time.Time
	LastUpdateTime time.Time
}

// PipelineOption defines a function that configures a pipeline.
type PipelineOption func(*DataPipeline)

// WithWorkers sets the number of worker goroutines.
func WithWorkers(workers int) PipelineOption {
	return func(p *DataPipeline) {
		p.workers = workers
	}
}

// WithBufferSize sets the buffer size for channels.
func WithBufferSize(size int) PipelineOption {
	return func(p *DataPipeline) {
		p.bufferSize = size
	}
}

// WithErrorHandler sets the error handler for the pipeline.
func WithErrorHandler(handler func(error)) PipelineOption {
	return func(p *DataPipeline) {
		p.errorHandler = handler
	}
}

// NewDataPipeline creates a new data processing pipeline.
func NewDataPipeline(processors []DataProcessor, opts ...PipelineOption) *DataPipeline {
	p := &DataPipeline{
		processors:   processors,
		workers:      1,
		bufferSize:   100,
		done:         make(chan struct{}),
		errorHandler: func(err error) {},
		metrics: &PipelineMetrics{
			StartTime: time.Now(),
		},
	}

	for _, opt := range opts {
		opt(p)
	}

	p.input = make(chan interface{}, p.bufferSize)
	p.output = make(chan interface{}, p.bufferSize)

	return p
}

// Start starts the data processing pipeline.
func (p *DataPipeline) Start(ctx context.Context) error {
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.worker(ctx, i)
	}

	return nil
}

// Stop stops the data processing pipeline.
func (p *DataPipeline) Stop() {
	close(p.done)
	p.wg.Wait()
	close(p.input)
	close(p.output)
}

// Process adds data to the pipeline for processing.
func (p *DataPipeline) Process(data interface{}) error {
	select {
	case p.input <- data:
		p.metricsMutex.Lock()
		p.metrics.ProcessedCount++
		p.metrics.LastUpdateTime = time.Now()
		p.metricsMutex.Unlock()
		return nil
	case <-p.done:
		return ErrPipelineClosed
	case <-time.After(100 * time.Millisecond):
		return ErrPipelineTimeout
	}
}

// Output returns the output channel for processed data.
func (p *DataPipeline) Output() <-chan interface{} {
	return p.output
}

// worker processes data from the input channel.
func (p *DataPipeline) worker(ctx context.Context, workerID int) {
	defer p.wg.Done()

	for {
		select {
		case <-p.done:
			return
		case <-ctx.Done():
			return
		case data, ok := <-p.input:
			if !ok {
				return
			}

			startTime := time.Now()
			result, err := p.processData(ctx, data)
			latency := time.Since(startTime)

			if err != nil {
				p.metricsMutex.Lock()
				p.metrics.ErrorCount++
				p.metricsMutex.Unlock()

				p.errorHandler(err)
				continue
			}

			// Update average latency
			p.metricsMutex.Lock()
			p.metrics.AverageLatency = (p.metrics.AverageLatency*time.Duration(p.metrics.ProcessedCount-1) + latency) / time.Duration(p.metrics.ProcessedCount)
			p.metricsMutex.Unlock()

			select {
			case p.output <- result:
			case <-p.done:
				return
			}
		}
	}
}

// processData processes data through all processors.
func (p *DataPipeline) processData(ctx context.Context, data interface{}) (interface{}, error) {
	var err error
	result := data

	for _, processor := range p.processors {
		result, err = processor.Process(ctx, result)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

// GetMetrics returns the current pipeline metrics.
func (p *DataPipeline) GetMetrics() PipelineMetrics {
	p.metricsMutex.RLock()
	defer p.metricsMutex.RUnlock()

	return *p.metrics
}

// ResetMetrics resets the pipeline metrics.
func (p *DataPipeline) ResetMetrics() {
	p.metricsMutex.Lock()
	defer p.metricsMutex.Unlock()

	p.metrics = &PipelineMetrics{
		StartTime: time.Now(),
	}
}

// FilterProcessor filters data based on a predicate function.
type FilterProcessor struct {
	predicate func(interface{}) bool
}

// NewFilterProcessor creates a new filter processor.
func NewFilterProcessor(predicate func(interface{}) bool) *FilterProcessor {
	return &FilterProcessor{
		predicate: predicate,
	}
}

// Process filters the data based on the predicate.
func (p *FilterProcessor) Process(ctx context.Context, data interface{}) (interface{}, error) {
	if p.predicate(data) {
		return data, nil
	}
	return nil, nil // Signal to drop the data
}

// MapProcessor transforms data using a mapping function.
type MapProcessor struct {
	mapper func(interface{}) (interface{}, error)
}

// NewMapProcessor creates a new map processor.
func NewMapProcessor(mapper func(interface{}) (interface{}, error)) *MapProcessor {
	return &MapProcessor{
		mapper: mapper,
	}
}

// Process transforms the data using the mapper function.
func (p *MapProcessor) Process(ctx context.Context, data interface{}) (interface{}, error) {
	return p.mapper(data)
}

// ReduceProcessor reduces data using a reduction function.
type ReduceProcessor struct {
	reducer func(acc, val interface{}) (interface{}, error)
	acc     interface{}
}

// NewReduceProcessor creates a new reduce processor.
func NewReduceProcessor(initial interface{}, reducer func(acc, val interface{}) (interface{}, error)) *ReduceProcessor {
	return &ReduceProcessor{
		reducer: reducer,
		acc:     initial,
	}
}

// Process reduces the data using the reducer function.
func (p *ReduceProcessor) Process(ctx context.Context, data interface{}) (interface{}, error) {
	result, err := p.reducer(p.acc, data)
	if err != nil {
		return nil, err
	}
	p.acc = result
	return result, nil
}

// BatchProcessor batches data for efficient processing.
type BatchProcessor struct {
	batchSize int
	timeout   time.Duration
	batch     []interface{}
	timer     *time.Timer
	mu        sync.Mutex
}

// NewBatchProcessor creates a new batch processor.
func NewBatchProcessor(batchSize int, timeout time.Duration) *BatchProcessor {
	return &BatchProcessor{
		batchSize: batchSize,
		timeout:   timeout,
		batch:     make([]interface{}, 0, batchSize),
	}
}

// Process adds data to the batch.
func (p *BatchProcessor) Process(ctx context.Context, data interface{}) (interface{}, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.batch = append(p.batch, data)

	if len(p.batch) >= p.batchSize {
		batch := p.batch
		p.batch = make([]interface{}, 0, p.batchSize)
		if p.timer != nil {
			p.timer.Stop()
			p.timer = nil
		}
		return batch, nil
	}

	if p.timer == nil {
		p.timer = time.AfterFunc(p.timeout, func() {
			p.mu.Lock()
			defer p.mu.Unlock()
			if len(p.batch) > 0 {
				_ = p.batch // Consume the batch
				p.batch = make([]interface{}, 0, p.batchSize)
				p.timer = nil
				// In a real implementation, we'd send the batch to the output
			}
		})
	}

	return nil, nil
}

// DebounceProcessor debounces data using a time window.
type DebounceProcessor struct {
	duration time.Duration
	lastData interface{}
	timer    *time.Timer
	mu       sync.Mutex
	output   chan interface{}
}

// NewDebounceProcessor creates a new debounce processor.
func NewDebounceProcessor(duration time.Duration) *DebounceProcessor {
	return &DebounceProcessor{
		duration: duration,
		output:   make(chan interface{}, 1),
	}
}

// Process debounces the data.
func (p *DebounceProcessor) Process(ctx context.Context, data interface{}) (interface{}, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.lastData = data

	if p.timer != nil {
		p.timer.Stop()
	}

	p.timer = time.AfterFunc(p.duration, func() {
		p.mu.Lock()
		defer p.mu.Unlock()

		if p.lastData != nil {
			p.output <- p.lastData
			p.lastData = nil
		}
	})

	return nil, nil
}

// Output returns the output channel for debounced data.
func (p *DebounceProcessor) Output() <-chan interface{} {
	return p.output
}

// ThrottleProcessor throttles data to a maximum rate.
type ThrottleProcessor struct {
	interval time.Duration
	lastSend time.Time
	mu       sync.Mutex
}

// NewThrottleProcessor creates a new throttle processor.
func NewThrottleProcessor(interval time.Duration) *ThrottleProcessor {
	return &ThrottleProcessor{
		interval: interval,
	}
}

// Process throttles the data.
func (p *ThrottleProcessor) Process(ctx context.Context, data interface{}) (interface{}, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now()
	if now.Sub(p.lastSend) < p.interval {
		return nil, nil
	}

	p.lastSend = now
	return data, nil
}

// Errors
var (
	ErrPipelineClosed = errors.New("pipeline is closed")
	ErrPipelineTimeout = errors.New("pipeline timeout")
)