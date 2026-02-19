// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later


package batch

import (
	"context"
	"sync"
	"time"
)

// Processor defines the interface for batch processing
type Processor interface {
	Process(ctx context.Context, items []interface{}) error
}

// BatchProcessor manages batch processing of items
type BatchProcessor struct {
	processor   Processor
	batchSize   int
	timeout     time.Duration
	buffer      []interface{}
	mu          sync.Mutex
	flushTimer  *time.Timer
	stopChan    chan struct{}
}

// NewBatchProcessor creates a new batch processor
func NewBatchProcessor(processor Processor, batchSize int, timeout time.Duration) *BatchProcessor {
	bp := &BatchProcessor{
		processor:  processor,
		batchSize:  batchSize,
		timeout:    timeout,
		buffer:     make([]interface{}, 0, batchSize),
		stopChan:   make(chan struct{}),
		flushTimer: time.NewTimer(timeout),
	}

	// Start background flush goroutine
	go bp.backgroundFlush()

	return bp
}

// Add adds an item to the batch
func (bp *BatchProcessor) Add(ctx context.Context, item interface{}) error {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	bp.buffer = append(bp.buffer, item)

	// Flush immediately if batch is full
	if len(bp.buffer) >= bp.batchSize {
		return bp.flush(ctx)
	}

	return nil
}

// Flush flushes the current batch
func (bp *BatchProcessor) Flush(ctx context.Context) error {
	bp.mu.Lock()
	defer bp.mu.Unlock()
	return bp.flush(ctx)
}

// flush performs the actual flush (must be called with lock held)
func (bp *BatchProcessor) flush(ctx context.Context) error {
	if len(bp.buffer) == 0 {
		return nil
	}

	// Copy buffer to avoid holding lock during processing
	batch := make([]interface{}, len(bp.buffer))
	copy(batch, bp.buffer)

	// Clear buffer
	bp.buffer = bp.buffer[:0]

	// Reset timer
	if !bp.flushTimer.Stop() {
		select {
		case <-bp.flushTimer.C:
		default:
		}
	}
	bp.flushTimer.Reset(bp.timeout)

	// Process batch (release lock during processing)
	bp.mu.Unlock()
	err := bp.processor.Process(ctx, batch)
	bp.mu.Lock()

	return err
}

// backgroundFlush periodically flushes the batch
func (bp *BatchProcessor) backgroundFlush() {
	ticker := time.NewTicker(bp.timeout / 2) // Check twice per timeout
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), bp.timeout)
			bp.Flush(ctx)
			cancel()
		case <-bp.stopChan:
			return
		}
	}
}

// Close stops the batch processor
func (bp *BatchProcessor) Close(ctx context.Context) error {
	close(bp.stopChan)
	bp.flushTimer.Stop()
	return bp.Flush(ctx)
}

// BatchWriter provides batch writing functionality
type BatchWriter struct {
	writeFunc func(ctx context.Context, items []interface{}) error
	batchSize int
	timeout   time.Duration
}

// NewBatchWriter creates a new batch writer
func NewBatchWriter(writeFunc func(ctx context.Context, items []interface{}) error, batchSize int, timeout time.Duration) *BatchWriter {
	return &BatchWriter{
		writeFunc: writeFunc,
		batchSize: batchSize,
		timeout:   timeout,
	}
}

// WriteBatch writes a batch of items
func (bw *BatchWriter) WriteBatch(ctx context.Context, items []interface{}) error {
	// Split into smaller batches if needed
	for i := 0; i < len(items); i += bw.batchSize {
		end := i + bw.batchSize
		if end > len(items) {
			end = len(items)
		}

		batch := items[i:end]
		if err := bw.writeFunc(ctx, batch); err != nil {
			return err
		}
	}

	return nil
}

// DatabaseBatchWriter provides batch database operations
type DatabaseBatchWriter struct {
	db         interface{} // database connection
	batchSize  int
	tableName  string
}

// NewDatabaseBatchWriter creates a new database batch writer
func NewDatabaseBatchWriter(db interface{}, batchSize int, tableName string) *DatabaseBatchWriter {
	return &DatabaseBatchWriter{
		db:        db,
		batchSize: batchSize,
		tableName: tableName,
	}
}

// BatchInsert performs batch insert
func (dw *DatabaseBatchWriter) BatchInsert(ctx context.Context, items []map[string]interface{}) error {
	// Split into batches
	for i := 0; i < len(items); i += dw.batchSize {
		end := i + dw.batchSize
		if end > len(items) {
			end = len(items)
		}

		batch := items[i:end]
		if err := dw.executeBatchInsert(ctx, batch); err != nil {
			return err
		}
	}

	return nil
}

// executeBatchInsert executes a single batch insert
func (dw *DatabaseBatchWriter) executeBatchInsert(ctx context.Context, batch []map[string]interface{}) error {
	// Implementation depends on database type
	// This is a placeholder for the actual implementation
	_ = ctx
	_ = batch
	return nil
}

// CacheBatchWriter provides batch cache operations
type CacheBatchWriter struct {
	cache      interface{}
	batchSize  int
}

// NewCacheBatchWriter creates a new cache batch writer
func NewCacheBatchWriter(cache interface{}, batchSize int) *CacheBatchWriter {
	return &CacheBatchWriter{
		cache:     cache,
		batchSize: batchSize,
	}
}

// BatchSet sets multiple keys in the cache
func (cw *CacheBatchWriter) BatchSet(ctx context.Context, items map[string]interface{}) error {
	// Split into batches
	keys := make([]string, 0, len(items))
	for k := range items {
		keys = append(keys, k)
	}

	for i := 0; i < len(keys); i += cw.batchSize {
		end := i + cw.batchSize
		if end > len(keys) {
			end = len(keys)
		}

		batch := make(map[string]interface{})
		for _, k := range keys[i:end] {
			batch[k] = items[k]
		}

		if err := cw.executeBatchSet(ctx, batch); err != nil {
			return err
		}
	}

	return nil
}

// executeBatchSet executes a single batch set
func (cw *CacheBatchWriter) executeBatchSet(ctx context.Context, batch map[string]interface{}) error {
	// Implementation depends on cache type
	_ = ctx
	_ = batch
	return nil
}

// BatchExecutor provides concurrent batch execution
type BatchExecutor struct {
	concurrency int
	timeout     time.Duration
}

// NewBatchExecutor creates a new batch executor
func NewBatchExecutor(concurrency int, timeout time.Duration) *BatchExecutor {
	return &BatchExecutor{
		concurrency: concurrency,
		timeout:     timeout,
	}
}

// Execute executes multiple functions concurrently
func (be *BatchExecutor) Execute(ctx context.Context, funcs []func(context.Context) error) error {
	// Create semaphore for concurrency control
	sem := make(chan struct{}, be.concurrency)
	errChan := make(chan error, len(funcs))

	// Execute functions concurrently
	for _, fn := range funcs {
		go func(f func(context.Context) error) {
			sem <- struct{}{}        // Acquire semaphore
			defer func() { <-sem }() // Release semaphore

			// Add timeout
			ctx, cancel := context.WithTimeout(ctx, be.timeout)
			defer cancel()

			errChan <- f(ctx)
		}(fn)
	}

	// Collect results
	var errors []error
	for i := 0; i < len(funcs); i++ {
		if err := <-errChan; err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return errors[0] // Return first error
	}

	return nil
}

// BatchAggregator aggregates items for batch processing
type BatchAggregator struct {
	aggregatorFunc func(interface{}, interface{}) (interface{}, error)
}

// NewBatchAggregator creates a new batch aggregator
func NewBatchAggregator(aggregatorFunc func(interface{}, interface{}) (interface{}, error)) *BatchAggregator {
	return &BatchAggregator{
		aggregatorFunc: aggregatorFunc,
	}
}

// Aggregate aggregates multiple items into one
func (ba *BatchAggregator) Aggregate(items []interface{}) (interface{}, error) {
	if len(items) == 0 {
		return nil, nil
	}

	result := items[0]
	for i := 1; i < len(items); i++ {
		aggregated, err := ba.aggregatorFunc(result, items[i])
		if err != nil {
			return nil, err
		}
		result = aggregated
	}

	return result, nil
}

// MetricsAggregator aggregates metrics
type MetricsAggregator struct {
	sumAggregator *BatchAggregator
}

// NewMetricsAggregator creates a new metrics aggregator
func NewMetricsAggregator() *MetricsAggregator {
	return &MetricsAggregator{
		sumAggregator: NewBatchAggregator(func(a, b interface{}) (interface{}, error) {
			sumA := a.(float64)
			sumB := b.(float64)
			return sumA + sumB, nil
		}),
	}
}

// AggregateSum aggregates values by summing them
func (ma *MetricsAggregator) AggregateSum(values []float64) (float64, error) {
	items := make([]interface{}, len(values))
	for i, v := range values {
		items[i] = v
	}

	result, err := ma.sumAggregator.Aggregate(items)
	if err != nil {
		return 0, err
	}

	return result.(float64), nil
}

// AggregateAverage calculates the average of values
func (ma *MetricsAggregator) AggregateAverage(values []float64) (float64, error) {
	sum, err := ma.AggregateSum(values)
	if err != nil {
		return 0, err
	}

	return sum / float64(len(values)), nil
}
