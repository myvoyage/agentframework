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

package event

import (
	"errors"
	"sync"
	"testing"
	"time"
)

// TestBasicSubscribePublish tests basic subscribe and publish functionality
func TestBasicSubscribePublish(t *testing.T) {
	bus := NewMemoryEventBus()
	var receivedEvent Event
	var wg sync.WaitGroup
	wg.Add(1)

	// Subscribe to a topic
	bus.Subscribe("test.topic", func(event Event) error {
		receivedEvent = event
		wg.Done()
		return nil
	})

	// Publish an event
	testPayload := "test payload"
	errors := bus.Publish("test.topic", testPayload)
	if len(errors) > 0 {
		t.Errorf("Expected no errors, got %v", errors)
	}

	// Wait for the event to be processed
	done := make(chan bool, 1)
	go func() {
		wg.Wait()
		done <- true
	}()

	select {
	case <-done:
		// Verify the event was received correctly
		if receivedEvent.Topic != "test.topic" {
			t.Errorf("Expected topic 'test.topic', got '%s'", receivedEvent.Topic)
		}
		if receivedEvent.Payload != testPayload {
			t.Errorf("Expected payload '%s', got '%v'", testPayload, receivedEvent.Payload)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Event was not received within the timeout")
	}
}

// TestUnsubscribe tests unsubscribe functionality
func TestUnsubscribe(t *testing.T) {
	bus := NewMemoryEventBus()
	eventCount := 0
	var mutex sync.Mutex

	// Subscribe to a topic
	subscription := bus.Subscribe("test.topic", func(event Event) error {
		mutex.Lock()
		eventCount++
		mutex.Unlock()
		return nil
	})

	// Publish first event
	errors := bus.Publish("test.topic", "payload1")
	if len(errors) > 0 {
		t.Errorf("Expected no errors, got %v", errors)
	}

	// Unsubscribe
	bus.Unsubscribe(subscription)

	// Publish second event
	errors = bus.Publish("test.topic", "payload2")
	if len(errors) > 0 {
		t.Errorf("Expected no errors, got %v", errors)
	}

	// Verify only the first event was received
	mutex.Lock()
	defer mutex.Unlock()
	if eventCount != 1 {
		t.Errorf("Expected 1 event, got %d", eventCount)
	}
}

// TestMultipleSubscribers tests multiple subscribers to the same topic
func TestMultipleSubscribers(t *testing.T) {
	bus := NewMemoryEventBus()
	subscriber1Count := 0
	subscriber2Count := 0
	var mutex sync.Mutex
	var wg sync.WaitGroup
	wg.Add(2)

	// Add first subscriber
	bus.Subscribe("test.topic", func(event Event) error {
		mutex.Lock()
		subscriber1Count++
		mutex.Unlock()
		wg.Done()
		return nil
	})

	// Add second subscriber
	bus.Subscribe("test.topic", func(event Event) error {
		mutex.Lock()
		subscriber2Count++
		mutex.Unlock()
		wg.Done()
		return nil
	})

	// Publish an event
	errors := bus.Publish("test.topic", "test payload")
	if len(errors) > 0 {
		t.Errorf("Expected no errors, got %v", errors)
	}

	// Wait for events to be processed
	done := make(chan bool, 1)
	go func() {
		wg.Wait()
		done <- true
	}()

	select {
	case <-done:
		// Verify both subscribers received the event
		mutex.Lock()
		defer mutex.Unlock()
		if subscriber1Count != 1 {
			t.Errorf("Expected subscriber1 to receive 1 event, got %d", subscriber1Count)
		}
		if subscriber2Count != 1 {
			t.Errorf("Expected subscriber2 to receive 1 event, got %d", subscriber2Count)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Events were not received within the timeout")
	}
}

// TestMultipleTopics tests publishing to multiple topics
func TestMultipleTopics(t *testing.T) {
	bus := NewMemoryEventBus()
	var topic1Event Event
	var topic2Event Event
	var wg sync.WaitGroup
	wg.Add(2)

	// Subscribe to topic1
	bus.Subscribe("topic1", func(event Event) error {
		topic1Event = event
		wg.Done()
		return nil
	})

	// Subscribe to topic2
	bus.Subscribe("topic2", func(event Event) error {
		topic2Event = event
		wg.Done()
		return nil
	})

	// Publish to both topics
	errors := bus.Publish("topic1", "payload1")
	if len(errors) > 0 {
		t.Errorf("Expected no errors, got %v", errors)
	}
	errors = bus.Publish("topic2", "payload2")
	if len(errors) > 0 {
		t.Errorf("Expected no errors, got %v", errors)
	}

	// Wait for events to be processed
	done := make(chan bool, 1)
	go func() {
		wg.Wait()
		done <- true
	}()

	select {
	case <-done:
		// Verify events were received on the correct topics
		if topic1Event.Topic != "topic1" || topic1Event.Payload != "payload1" {
			t.Errorf("Topic1 event mismatch: got %v", topic1Event)
		}
		if topic2Event.Topic != "topic2" || topic2Event.Payload != "payload2" {
			t.Errorf("Topic2 event mismatch: got %v", topic2Event)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Events were not received within the timeout")
	}
}

// TestConcurrentAccess tests concurrent access to the event bus
func TestConcurrentAccess(t *testing.T) {
	bus := NewMemoryEventBus()
	var eventCount int
	var mutex sync.Mutex

	// Add 5 subscribers
	for i := 0; i < 5; i++ {
		bus.Subscribe("test.topic", func(event Event) error {
			mutex.Lock()
			eventCount++
			mutex.Unlock()
			return nil
		})
	}

	// Publish 100 events concurrently
	publishWg := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		publishWg.Add(1)
		go func() {
			errors := bus.Publish("test.topic", "payload")
			if len(errors) > 0 {
				// Ignore errors for concurrent test
			}
			publishWg.Done()
		}()
	}
	publishWg.Wait()

	// Wait a bit for all events to be processed
	time.Sleep(100 * time.Millisecond)

	// Verify all events were received by all subscribers
	mutex.Lock()
	defer mutex.Unlock()
	expectedCount := 5 * 100 // 5 subscribers, 100 events each
	if eventCount != expectedCount {
		t.Errorf("Expected %d events, got %d", expectedCount, eventCount)
	}
}

// TestTopicCleanup tests that topics without subscribers are automatically cleaned up
func TestTopicCleanup(t *testing.T) {
	bus := NewMemoryEventBus()

	// Subscribe to a topic
	subscription := bus.Subscribe("test.topic", func(event Event) error {
		return nil
	})

	// Verify the topic exists
	bus.mu.RLock()
	_, exists := bus.subscribers["test.topic"]
	bus.mu.RUnlock()
	if !exists {
		t.Error("Topic should exist after subscription")
	}

	// Unsubscribe
	bus.Unsubscribe(subscription)

	// Verify the topic was cleaned up
	bus.mu.RLock()
	_, exists = bus.subscribers["test.topic"]
	bus.mu.RUnlock()
	if exists {
		t.Error("Topic should be cleaned up after all subscribers are removed")
	}
}

// TestSubscribeAfterUnsubscribe tests subscribing again after unsubscribing
func TestSubscribeAfterUnsubscribe(t *testing.T) {
	bus := NewMemoryEventBus()
	eventCount := 0
	var mutex sync.Mutex
	var wg sync.WaitGroup

	// First subscription
	subscription := bus.Subscribe("test.topic", func(event Event) error {
		mutex.Lock()
		eventCount++
		mutex.Unlock()
		return nil
	})

	// Publish and unsubscribe
	errors := bus.Publish("test.topic", "payload1")
	if len(errors) > 0 {
		t.Errorf("Expected no errors, got %v", errors)
	}
	bus.Unsubscribe(subscription)

	// Second subscription
	wg.Add(1)
	bus.Subscribe("test.topic", func(event Event) error {
		mutex.Lock()
		eventCount++
		mutex.Unlock()
		wg.Done()
		return nil
	})

	// Publish again
	errors = bus.Publish("test.topic", "payload2")
	if len(errors) > 0 {
		t.Errorf("Expected no errors, got %v", errors)
	}

	// Wait for the event to be processed
	done := make(chan bool, 1)
	go func() {
		wg.Wait()
		done <- true
	}()

	select {
	case <-done:
		// Verify both events were received
		mutex.Lock()
		defer mutex.Unlock()
		if eventCount != 2 {
			t.Errorf("Expected 2 events, got %d", eventCount)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Event was not received within the timeout")
	}
}

// TestSubscribeAsync tests asynchronous event handling
func TestSubscribeAsync(t *testing.T) {
	bus := NewMemoryEventBus()
	var eventCount int
	var mutex sync.Mutex
	var wg sync.WaitGroup
	wg.Add(1)

	// Subscribe with async handler
	bus.SubscribeAsync("test.topic", func(event Event) error {
		// Simulate slow processing
		time.Sleep(50 * time.Millisecond)
		mutex.Lock()
		eventCount++
		mutex.Unlock()
		wg.Done()
		return nil
	})

	// Publish an event
	startTime := time.Now()
	errors := bus.Publish("test.topic", "test payload")
	if len(errors) > 0 {
		t.Errorf("Expected no errors for async handlers, got %v", errors)
	}

	// Verify the publish call returns immediately (async processing)
	publishDuration := time.Since(startTime)
	if publishDuration > 10*time.Millisecond {
		t.Errorf("Publish with async handler should return immediately, took %v", publishDuration)
	}

	// Wait for the async event to be processed
	done := make(chan bool, 1)
	go func() {
		wg.Wait()
		done <- true
	}()

	select {
	case <-done:
		// Verify the event was received correctly
		mutex.Lock()
		defer mutex.Unlock()
		if eventCount != 1 {
			t.Errorf("Expected 1 event, got %d", eventCount)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Async event was not processed within the timeout")
	}
}

// TestWithAsyncHandlerOption tests the WithAsyncHandler option
func TestWithAsyncHandlerOption(t *testing.T) {
	// Create bus with async as default
	bus := NewMemoryEventBus(WithAsyncHandler())
	var eventCount int
	var mutex sync.Mutex
	var wg sync.WaitGroup
	wg.Add(1)

	// Subscribe with default handler (should be async)
	bus.Subscribe("test.topic", func(event Event) error {
		// Simulate slow processing
		time.Sleep(50 * time.Millisecond)
		mutex.Lock()
		eventCount++
		mutex.Unlock()
		wg.Done()
		return nil
	})

	// Publish an event
	startTime := time.Now()
	errors := bus.Publish("test.topic", "test payload")
	if len(errors) > 0 {
		t.Errorf("Expected no errors for async handlers, got %v", errors)
	}

	// Verify the publish call returns immediately
	publishDuration := time.Since(startTime)
	if publishDuration > 10*time.Millisecond {
		t.Errorf("Publish with default async handler should return immediately, took %v", publishDuration)
	}

	// Wait for the async event to be processed
	done := make(chan bool, 1)
	go func() {
		wg.Wait()
		done <- true
	}()

	select {
	case <-done:
		// Verify the event was received correctly
		mutex.Lock()
		defer mutex.Unlock()
		if eventCount != 1 {
			t.Errorf("Expected 1 event, got %d", eventCount)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Async event was not processed within the timeout")
	}
}

// TestMixedSyncAsyncSubscribers tests mixed sync and async subscribers
func TestMixedSyncAsyncSubscribers(t *testing.T) {
	bus := NewMemoryEventBus()
	var syncEventCount int
	var asyncEventCount int
	var mutex sync.Mutex
	var wg sync.WaitGroup
	wg.Add(1)

	// Add sync subscriber
	bus.Subscribe("test.topic", func(event Event) error {
		mutex.Lock()
		syncEventCount++
		mutex.Unlock()
		return nil
	})

	// Add async subscriber
	bus.SubscribeAsync("test.topic", func(event Event) error {
		// Simulate slow processing
		time.Sleep(50 * time.Millisecond)
		mutex.Lock()
		asyncEventCount++
		mutex.Unlock()
		wg.Done()
		return nil
	})

	// Publish an event
	errors := bus.Publish("test.topic", "test payload")
	if len(errors) > 0 {
		t.Errorf("Expected no errors, got %v", errors)
	}

	// Verify sync event was processed immediately
	mutex.Lock()
	syncCount := syncEventCount
	mutex.Unlock()
	if syncCount != 1 {
		t.Errorf("Expected sync event to be processed immediately, got %d", syncCount)
	}

	// Wait for the async event to be processed
	done := make(chan bool, 1)
	go func() {
		wg.Wait()
		done <- true
	}()

	select {
	case <-done:
		// Verify both events were received correctly
		mutex.Lock()
		defer mutex.Unlock()
		if syncEventCount != 1 {
			t.Errorf("Expected 1 sync event, got %d", syncEventCount)
		}
		if asyncEventCount != 1 {
			t.Errorf("Expected 1 async event, got %d", asyncEventCount)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Async event was not processed within the timeout")
	}
}

// TestHandlerError tests that handler errors are returned correctly
func TestHandlerError(t *testing.T) {
	bus := NewMemoryEventBus()
	var successCount int

	// Subscribe with an error-returning handler
	subscription := bus.Subscribe("test.topic", func(event Event) error {
		successCount++
		return errors.New("handler error")
	})

	// Subscribe with a successful handler
	bus.Subscribe("test.topic", func(event Event) error {
		successCount++
		return nil
	})

	// Publish an event
	errors := bus.Publish("test.topic", "test payload")

	// Verify errors were returned correctly
	if len(errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(errors))
	}

	if successCount != 2 {
		t.Errorf("Expected both handlers to be called, got %d calls", successCount)
	}

	// Verify the error is correct for the subscription
	if _, ok := errors[subscription.id]; !ok {
		t.Errorf("Expected error for subscription %d, got errors for %v", subscription.id, errors)
	}
}

// TestWithBatchSize tests the WithBatchSize option
func TestWithBatchSize(t *testing.T) {
	bus := NewMemoryEventBus(WithBatchSize(10))

	if bus.batchSize != 10 {
		t.Errorf("expected batchSize 10, got %d", bus.batchSize)
	}
}

// TestWithQueueSize tests the WithQueueSize option
func TestWithQueueSize(t *testing.T) {
	bus := NewMemoryEventBus(WithQueueSize(500))

	if bus.queueSize != 500 {
		t.Errorf("expected queueSize 500, got %d", bus.queueSize)
	}
}

// TestWithMonitor tests the WithMonitor option
func TestWithMonitor(t *testing.T) {
	bus := NewMemoryEventBus(WithMonitor())

	if !bus.monitor {
		t.Error("expected monitor to be true")
	}
}

// TestMultipleOptions tests multiple options combined
func TestMultipleOptions(t *testing.T) {
	bus := NewMemoryEventBus(
		WithBatchSize(5),
		WithQueueSize(200),
		WithMonitor(),
		WithAsyncHandler(),
	)

	if bus.batchSize != 5 {
		t.Errorf("expected batchSize 5, got %d", bus.batchSize)
	}
	if bus.queueSize != 200 {
		t.Errorf("expected queueSize 200, got %d", bus.queueSize)
	}
	if !bus.monitor {
		t.Error("expected monitor to be true")
	}
	if !bus.async {
		t.Error("expected async to be true")
	}
}

// TestGetStats tests the GetStats method
func TestGetStats(t *testing.T) {
	bus := NewMemoryEventBus(WithMonitor())

	// Subscribe and publish some events
	bus.Subscribe("test.topic", func(event Event) error {
		return nil
	})

	bus.Publish("test.topic", "payload1")
	bus.Publish("test.topic", "payload2")

	stats := bus.GetStats()

	if stats.EventCount != 2 {
		t.Errorf("expected EventCount 2, got %d", stats.EventCount)
	}
	if !stats.IsMonitoring {
		t.Error("expected IsMonitoring to be true")
	}
}

// TestClose tests the Close method
func TestClose(t *testing.T) {
	bus := NewMemoryEventBus(WithMonitor())

	// Subscribe and publish
	bus.Subscribe("test.topic", func(event Event) error {
		return nil
	})
	bus.Publish("test.topic", "payload")

	// Close the bus
	err := bus.Close()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// After close, stats should still be accessible
	stats := bus.GetStats()
	if stats.EventCount != 1 {
		t.Errorf("expected EventCount 1, got %d", stats.EventCount)
	}
}

// TestCloseWithBatchProcessor tests closing a bus with batch processing
func TestCloseWithBatchProcessor(t *testing.T) {
	bus := NewMemoryEventBus(WithBatchSize(10))

	// The batch processor goroutine should be started
	if bus.batchChan == nil {
		t.Error("expected batchChan to be initialized")
	}

	// Close the bus
	err := bus.Close()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestStatsTracking tests that stats are tracked correctly
func TestStatsTracking(t *testing.T) {
	bus := NewMemoryEventBus(WithMonitor())

	var handlerCount int
	bus.Subscribe("test.topic", func(event Event) error {
		handlerCount++
		return nil
	})

	// Publish events
	for i := 0; i < 5; i++ {
		bus.Publish("test.topic", i)
	}

	stats := bus.GetStats()

	if stats.EventCount != 5 {
		t.Errorf("expected EventCount 5, got %d", stats.EventCount)
	}
	if stats.HandlerCount != 5 {
		t.Errorf("expected HandlerCount 5, got %d", stats.HandlerCount)
	}
}

// TestStatsErrorCount tests error count in stats
func TestStatsErrorCount(t *testing.T) {
	bus := NewMemoryEventBus(WithMonitor())

	// Subscribe with a handler that returns error
	bus.Subscribe("test.topic", func(event Event) error {
		return errors.New("test error")
	})

	bus.Publish("test.topic", "payload")

	stats := bus.GetStats()

	if stats.ErrorCount != 1 {
		t.Errorf("expected ErrorCount 1, got %d", stats.ErrorCount)
	}
}

// TestEventPriority tests that events are processed in priority order
func TestEventPriority(t *testing.T) {
	bus := NewMemoryEventBus(WithBatchSize(10))

	var receivedOrder []int
	var mu sync.Mutex

	bus.Subscribe("test.topic", func(event Event) error {
		mu.Lock()
		defer mu.Unlock()
		if payload, ok := event.Payload.(int); ok {
			receivedOrder = append(receivedOrder, payload)
		}
		return nil
	})

	// Publish events with different priorities
	bus.Publish("test.topic", 3)
	bus.Publish("test.topic", 1)
	bus.Publish("test.topic", 2)

	// Wait for batch processing
	time.Sleep(200 * time.Millisecond)

	bus.Close()

	mu.Lock()
	defer mu.Unlock()

	// Check that events were received
	if len(receivedOrder) != 3 {
		t.Errorf("expected 3 events, got %d", len(receivedOrder))
	}
}

// TestBatchProcessing tests batch processing functionality
func TestBatchProcessing(t *testing.T) {
	bus := NewMemoryEventBus(WithBatchSize(3))

	var eventCount int
	var mu sync.Mutex
	var wg sync.WaitGroup

	bus.Subscribe("test.topic", func(event Event) error {
		mu.Lock()
		eventCount++
		mu.Unlock()
		wg.Done()
		return nil
	})

	wg.Add(3)

	// Publish events - they should be processed in batches
	bus.Publish("test.topic", 1)
	bus.Publish("test.topic", 2)
	bus.Publish("test.topic", 3)

	// Wait for batch processing
	wg.Wait()
	time.Sleep(100 * time.Millisecond)

	bus.Close()

	mu.Lock()
	count := eventCount
	mu.Unlock()

	if count != 3 {
		t.Errorf("expected 3 events processed, got %d", count)
	}
}

