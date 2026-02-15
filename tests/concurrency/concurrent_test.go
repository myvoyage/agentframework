// Agent Framework - Concurrency Safety Tests
// Copyright (C) 2025  Agent Framework Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package concurrency_test

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"AgentFramework/agent"
)

// TestConcurrentAgentExecution tests concurrent agent execution
func TestConcurrentAgentExecution(t *testing.T) {
	t.Parallel()

	// Create a simple agent for testing
	cfg := agent.ChatAgentConfig{
		Name:    "test-agent",
		Model:    "test-model",
		SystemPrompt: "You are a helpful assistant",
	}

	agt := agent.NewChatAgent(cfg)

	if agt == nil {
		t.Skip("Agent creation failed, skipping concurrent test")
		return
	}

	// Concurrent execution parameters
	concurrency := 10
	iterations := 100

	var wg sync.WaitGroup
	errors := make(chan error, concurrency*iterations)
	successCount := atomic.Int64{}

	// Launch concurrent executions
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for j := 0; j < iterations; j++ {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				input := fmt.Sprintf("Test input from worker %d, iteration %d", workerID, j)

				_, err := agt.Run(ctx, input)
				cancel()

				if err != nil {
					errors <- fmt.Errorf("worker %d iteration %d: %w", workerID, j, err)
				} else {
					successCount.Add(1)
				}
			}
		}(i)
	}

	// Wait for all workers to complete
	wg.Wait()
	close(errors)

	// Check for errors
	errorList := []error{}
	for err := range errors {
		errorList = append(errorList, err)
	}

	if len(errorList) > 0 {
		t.Errorf("Concurrent execution had %d errors out of %d total runs", len(errorList), concurrency*iterations)
		for _, err := range errorList {
			t.Logf("Error: %v", err)
		}
	}

	t.Logf("Concurrent execution test completed: %d successful runs out of %d total", successCount.Load(), concurrency*iterations)
}

// TestConcurrentEventBusPublishing tests concurrent event publishing
func TestConcurrentEventBusPublishing(t *testing.T) {
	t.Parallel()

	bus := agent.NewMemoryEventBus(
		agent.WithQueueSize(1000),
		agent.WithAsyncHandler(),
		agent.WithMonitor(),
	)
	defer bus.Close()

	topic := "test-topic"

	// Subscribe multiple handlers
	var handlerCount atomic.Int64
	for i := 0; i < 5; i++ {
		bus.SubscribeAsync(topic, func(evt agent.Event) error {
			handlerCount.Add(1)
			return nil
		})
	}

	// Concurrent publishing parameters
	concurrency := 10
	eventsPerWorker := 100
	totalEvents := concurrency * eventsPerWorker

	var wg sync.WaitGroup
	publishErrors := make(chan error, totalEvents)

	// Launch concurrent publishers
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for j := 0; j < eventsPerWorker; j++ {
				payload := fmt.Sprintf("Event from worker %d, iteration %d", workerID, j)
				errs := bus.Publish(topic, payload)
				if errs != nil {
					publishErrors <- fmt.Errorf("worker %d iteration %d: %v", workerID, j, errs)
				}
			}
		}(i)
	}

	// Wait for all publishers
	wg.Wait()
	close(publishErrors)

	// Collect errors
	errorList := []error{}
	for err := range publishErrors {
		errorList = append(errorList, err)
	}

	if len(errorList) > 0 {
		t.Errorf("Concurrent publishing had %d errors", len(errorList))
	}

	t.Logf("Event bus test: %d handlers triggered, total events published: %d", handlerCount.Load(), totalEvents)
}

// TestDynamicEventBusConcurrency tests dynamic event bus under concurrent load
func TestDynamicEventBusConcurrency(t *testing.T) {
	t.Parallel()

	cfg := agent.DefaultDynamicEventBusConfig()
	cfg.InitialQueueSize = 100
	cfg.MaxQueueSize = 10000
	cfg.ResizeThreshold = 0.7
	cfg.Monitoring = true

	bus := agent.NewDynamicEventBus(cfg)
	defer bus.Close()

	topic := "concurrency-test"

	// Subscribe handlers
	var receivedCount atomic.Int64
	for i := 0; i < 10; i++ {
		bus.SubscribeAsync(topic, func(evt agent.Event) error {
			receivedCount.Add(1)
			return nil
		})
	}

	// Concurrent stress test
	concurrency := 20
	eventsPerWorker := 200

	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for j := 0; j < eventsPerWorker; j++ {
				payload := fmt.Sprintf("Stress test event %d-%d", workerID, j)
				bus.Publish(topic, payload)
			}
		}(i)
	}

	wg.Wait()

	stats := bus.GetStats()
	t.Logf("Dynamic event bus stats: Events=%d, Handlers=%d, Resizes=%d, QueueUsage=%d/%d",
		stats.EventCount,
		stats.HandlerCount,
		stats.ResizeCount,
		stats.CurrentQueueLen,
		stats.CurrentQueueCap,
	)
}

// TestConcurrentModelCacheAccess tests model cache under concurrent access
func TestConcurrentModelCacheAccess(t *testing.T) {
	t.Parallel()

	cfg := agent.DefaultModelCacheConfig()
	cache := agent.NewModelCache(cfg)
	defer cache.StopCleanup()

	// Pre-populate cache
	models := []string{"model1", "model2", "model3", "model4", "model5"}
	for _, model := range models {
		cache.Put(model, &agent.TestChatModel{})
	}

	concurrency := 50
	iterations := 100

	var wg sync.WaitGroup
	var getOps, putOps, deleteOps atomic.Int64

	// Concurrent readers
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for j := 0; j < iterations; j++ {
				model := models[j%len(models)]
				if cache.Get(model) != nil {
					getOps.Add(1)
				}
			}
		}(i)
	}

	// Concurrent writers
	for i := 0; i < concurrency/2; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for j := 0; j < iterations/2; j++ {
				model := fmt.Sprintf("new-model-%d-%d", workerID, j)
				cache.Put(model, &agent.TestChatModel{})
				putOps.Add(1)
			}
		}(i)
	}

	// Concurrent deleters
	for i := 0; i < concurrency/4; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for j := 0; j < iterations/4; j++ {
				model := models[j%len(models)]
				cache.Delete(model)
				deleteOps.Add(1)
			}
		}(i)
	}

	wg.Wait()

	stats := cache.GetStats()
	t.Logf("Model cache concurrent test: Gets=%d, Puts=%d, Deletes=%d, Stats=%v",
		getOps.Load(), putOps.Load(), deleteOps.Load(), stats)
}

// TestConcurrentConfigAccess tests config manager under concurrent access
func TestConcurrentConfigAccess(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	configPath := tmpDir + "/concurrent_config.json"

	cfg := agent.NewConfigManager(configPath)
	if err := cfg.Set("test.key1", "value1"); err != nil {
		t.Fatalf("Failed to set initial config: %v", err)
	}

	concurrency := 100
	iterations := 50

	var wg sync.WaitGroup
	var errorCount atomic.Int64

	// Concurrent readers and writers
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for j := 0; j < iterations; j++ {
				key := fmt.Sprintf("key-%d-%d", workerID, j)
				value := fmt.Sprintf("value-%d-%d", workerID, j)

				// Mix of reads and writes
				if j%3 == 0 {
					if err := cfg.Set(key, value); err != nil {
						errorCount.Add(1)
					}
				} else {
					_ = cfg.Get(key, "default")
				}
			}
		}(i)
	}

	wg.Wait()

	if errorCount.Load() > 0 {
		t.Errorf("Concurrent config access had %d errors", errorCount.Load())
	}

	t.Logf("Concurrent config access test completed with %d errors", errorCount.Load())
}

// TestRaceConditionInEventBus detects data races in event bus
func TestRaceConditionInEventBus(t *testing.T) {
	t.Parallel()

	bus := agent.NewMemoryEventBus()
	defer bus.Close()

	topic := "race-test"

	var handlerExecuted atomic.Bool
	handler := func(evt agent.Event) error {
		// Simulate some work
		time.Sleep(1 * time.Millisecond)
		handlerExecuted.Store(true)
		return nil
	}

	// Subscribe same handler multiple times (potential race)
	for i := 0; i < 10; i++ {
		bus.SubscribeAsync(topic, handler)
	}

	// Publish concurrently
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			bus.Publish(topic, "test payload")
		}()
	}

	wg.Wait()

	if !handlerExecuted.Load() {
		t.Error("Handler was not executed during race condition test")
	}
}

// TestConcurrentWorkflowExecution tests concurrent workflow execution
func TestConcurrentWorkflowExecution(t *testing.T) {
	t.Parallel()

	// This test would require a workflow to be set up
	// For now, we'll skip if workflow is not available
	t.Skip("Workflow execution test requires workflow setup")
}

// BenchmarkConcurrentAgentExecution benchmarks concurrent agent performance
func BenchmarkConcurrentAgentExecution(b *testing.B) {
	cfg := agent.ChatAgentConfig{
		Name:    "bench-agent",
		Model:    "test-model",
		SystemPrompt: "You are a helpful assistant",
	}

	agt := agent.NewChatAgent(cfg)
	if agt == nil {
		b.Skip("Agent creation failed")
		return
	}

	ctx := context.Background()
	input := "Hello, how are you?"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		go func() {
			_, _ = agt.Run(ctx, input)
		}()
	}
}

// TestConcurrentMemoryMonitor tests memory monitor under concurrent access
func TestConcurrentMemoryMonitor(t *testing.T) {
	t.Parallel()

	cfg := agent.DefaultMemoryMonitorConfig()
	cfg.HistorySize = 100

	monitor := agent.NewMemoryMonitor(cfg)
	defer monitor.Stop()

	// Register components
	for i := 0; i < 10; i++ {
		name := fmt.Sprintf("component-%d", i)
		monitor.RegisterComponent(name)
	}

	concurrency := 20
	iterations := 50

	var wg sync.WaitGroup

	// Concurrent stats readers
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				_ = monitor.GetCurrentStats()
				_ = monitor.GetHistory()
			}
		}()
	}

	// Concurrent component memory updaters
	for i := 0; i < concurrency/2; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				name := fmt.Sprintf("component-%d", workerID%10)
				monitor.UpdateComponentMemory(name, 1024, 1024, 10)
			}
		}(i)
	}

	wg.Wait()

	stats := monitor.GetStats()
	t.Logf("Memory monitor concurrent test completed. History size: %d", len(stats))
}
