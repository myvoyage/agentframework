// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
// Object pool for memory optimization
// SPDX-License-Identifier: AGPL-3.0-or-later

package pool

import (
	"sync"
)

// MessagePool manages a pool of Message objects to reduce memory allocation
// PoolMetrics tracks pool usage metrics
type PoolMetrics struct {
	Hits   int64
	Misses int64
	Puts   int64
	Gets   int64
}

type ChunkPool struct {
	pool    sync.Pool
	chunkSize int
	metrics  *PoolMetrics
}

// NewChunkPool creates a new chunk pool
func NewChunkPool(chunkSize int) *ChunkPool {
	return &ChunkPool{
		pool: sync.Pool{
			New: func() interface{} {
				return make([]byte, chunkSize)
			},
		},
		chunkSize: chunkSize,
		metrics:  &PoolMetrics{},
	}
}

// Acquire gets a chunk from the pool
func (p *ChunkPool) Acquire() []byte {
	p.metrics.Gets++
	chunk := p.pool.Get().([]byte)
	if cap(chunk) == p.chunkSize && len(chunk) == 0 {
		p.metrics.Misses++
	} else {
		p.metrics.Hits++
	}
	return chunk
}

// Release returns a chunk to the pool
func (p *ChunkPool) Release(chunk []byte) {
	if chunk == nil {
		return
	}
	// Reset the chunk for reuse
	for i := range chunk {
		chunk[i] = 0
	}
	p.pool.Put(chunk[:p.chunkSize])
	p.metrics.Puts++
}

// Metrics returns the pool metrics
func (p *ChunkPool) Metrics() *PoolMetrics {
	return p.metrics
}

// WorkerPool manages a pool of worker goroutines
type WorkerPool struct {
	workers   chan chan Task
	taskQueue chan Task
	quit      chan bool
	wg        sync.WaitGroup
	metrics   *WorkerPoolMetrics
}

// WorkerPoolMetrics tracks worker pool metrics
type WorkerPoolMetrics struct {
	TotalWorkers      int32
	ActiveWorkers     int32
	TotalTasks        int64
	CompletedTasks    int64
	FailedTasks       int64
	PendingTasks      int32
}

// Task represents a task to be executed by a worker
type Task struct {
	ID       string
	Handler func() error
	Result   chan error
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(numWorkers int) *WorkerPool {
	pool := &WorkerPool{
		workers:   make(chan chan Task, numWorkers),
		taskQueue: make(chan Task, 100),
		quit:      make(chan bool),
		metrics:   &WorkerPoolMetrics{},
	}

	// Start workers
	pool.wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go pool.worker(i)
	}
	pool.metrics.TotalWorkers = int32(numWorkers)

	// Start dispatcher
	go pool.dispatcher()

	return pool
}

// worker represents a worker goroutine
func (p *WorkerPool) worker(id int) {
	defer p.wg.Done()

	taskChannel := make(chan Task)
	p.workers <- taskChannel

	for {
		select {
		case task := <-taskChannel:
			p.metrics.ActiveWorkers++
			p.metrics.PendingTasks--
			err := task.Handler()
			p.metrics.ActiveWorkers--
			p.metrics.TotalTasks++
			if err != nil {
				p.metrics.FailedTasks++
			} else {
				p.metrics.CompletedTasks++
			}
			if task.Result != nil {
				task.Result <- err
			}

		case <-p.quit:
			return
		}
	}
}

// dispatcher dispatches tasks to workers
func (p *WorkerPool) dispatcher() {
	for {
		select {
		case task := <-p.taskQueue:
			p.metrics.PendingTasks++
			go func() {
				worker := <-p.workers
				worker <- task
			}()

		case <-p.quit:
			// Close all worker channels
			close(p.workers)
			return
		}
	}
}

// Submit submits a task to the pool
func (p *WorkerPool) Submit(handler func() error) error {
	result := make(chan error, 1)
	task := Task{
		Handler: handler,
		Result:   result,
	}

	select {
	case p.taskQueue <- task:
		return nil
	default:
		return ErrPoolFull
	}
}

// SubmitAndWait submits a task and waits for completion
func (p *WorkerPool) SubmitAndWait(handler func() error) error {
	result := make(chan error, 1)
	task := Task{
		Handler: handler,
		Result:   result,
	}

	p.taskQueue <- task
	return <-result
}

// Stop stops the worker pool
func (p *WorkerPool) Stop() {
	close(p.quit)
	p.wg.Wait()
}

// Metrics returns the pool metrics
func (p *WorkerPool) Metrics() *WorkerPoolMetrics {
	return p.metrics
}

// Errors
var (
	ErrPoolFull = NewError("POOL_FULL", "task queue is full")
)

// NewError creates a new pool error
func NewError(code, message string) error {
	return &PoolError{
		Code:    code,
		Message: message,
	}
}

// PoolError represents a pool error
type PoolError struct {
	Code    string
	Message string
}

func (e *PoolError) Error() string {
	return e.Code + ": " + e.Message
}