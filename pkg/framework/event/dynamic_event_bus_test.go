// Agent Framework - Dynamic EventBus Tests
// Copyright (C) 2025  Agent Framework Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package event

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestDefaultDynamicEventBusConfig(t *testing.T) {
	config := DefaultDynamicEventBusConfig()

	if config.InitialQueueSize != 1000 {
		t.Errorf("expected InitialQueueSize 1000, got %d", config.InitialQueueSize)
	}
	if config.MaxQueueSize != 100000 {
		t.Errorf("expected MaxQueueSize 100000, got %d", config.MaxQueueSize)
	}
	if config.ResizeThreshold != 0.8 {
		t.Errorf("expected ResizeThreshold 0.8, got %f", config.ResizeThreshold)
	}
	if config.ResizeMultiplier != 2.0 {
		t.Errorf("expected ResizeMultiplier 2.0, got %f", config.ResizeMultiplier)
	}
	if !config.Monitoring {
		t.Error("expected Monitoring to be true")
	}
}

func TestNewDynamicEventBus(t *testing.T) {
	config := DynamicEventBusConfig{
		InitialQueueSize: 500,
		MaxQueueSize:     50000,
		ResizeThreshold:  0.7,
		ResizeMultiplier: 1.5,
		Monitoring:       false,
	}

	bus := NewDynamicEventBus(config)
	if bus == nil {
		t.Fatal("expected DynamicEventBus to be created")
	}

	if bus.cfg.InitialQueueSize != 500 {
		t.Errorf("expected InitialQueueSize 500, got %d", bus.cfg.InitialQueueSize)
	}
	if bus.cfg.MaxQueueSize != 50000 {
		t.Errorf("expected MaxQueueSize 50000, got %d", bus.cfg.MaxQueueSize)
	}
	if bus.cfg.ResizeThreshold != 0.7 {
		t.Errorf("expected ResizeThreshold 0.7, got %f", bus.cfg.ResizeThreshold)
	}
	if bus.cfg.ResizeMultiplier != 1.5 {
		t.Errorf("expected ResizeMultiplier 1.5, got %f", bus.cfg.ResizeMultiplier)
	}
	if bus.cfg.Monitoring {
		t.Error("expected Monitoring to be false")
	}
}

func TestNewDynamicEventBus_DefaultValues(t *testing.T) {
	config := DynamicEventBusConfig{}
	bus := NewDynamicEventBus(config)

	if bus.cfg.InitialQueueSize != 1000 {
		t.Errorf("expected default InitialQueueSize 1000, got %d", bus.cfg.InitialQueueSize)
	}
	if bus.cfg.MaxQueueSize != 0 {
		t.Errorf("expected default MaxQueueSize 0, got %d", bus.cfg.MaxQueueSize)
	}
	if bus.cfg.ResizeThreshold != 0.8 {
		t.Errorf("expected default ResizeThreshold 0.8, got %f", bus.cfg.ResizeThreshold)
	}
	if bus.cfg.ResizeMultiplier != 2.0 {
		t.Errorf("expected default ResizeMultiplier 2.0, got %f", bus.cfg.ResizeMultiplier)
	}
}

func TestDynamicEventBus_Subscribe(t *testing.T) {
	bus := NewDynamicEventBus(DefaultDynamicEventBusConfig())
	defer bus.Close()

	var received int32
	handler := func(event Event) error {
		atomic.AddInt32(&received, 1)
		return nil
	}

	_ = bus.Subscribe("test-topic", handler)

	// Publish an event
	bus.Publish("test-topic", "test-payload")

	// Wait for event to be processed
	time.Sleep(100 * time.Millisecond)

	if atomic.LoadInt32(&received) != 1 {
		t.Errorf("expected 1 event received, got %d", received)
	}
}

func TestDynamicEventBus_SubscribeAsync(t *testing.T) {
	bus := NewDynamicEventBus(DefaultDynamicEventBusConfig())
	defer bus.Close()

	var received int32
	handler := func(event Event) error {
		atomic.AddInt32(&received, 1)
		return nil
	}

	_ = bus.SubscribeAsync("test-topic", handler)

	// Publish an event
	bus.Publish("test-topic", "test-payload")

	// Wait for event to be processed
	time.Sleep(100 * time.Millisecond)

	if atomic.LoadInt32(&received) != 1 {
		t.Errorf("expected 1 event received, got %d", received)
	}
}

func TestDynamicEventBus_Publish(t *testing.T) {
	bus := NewDynamicEventBus(DefaultDynamicEventBusConfig())
	defer bus.Close()

	var received int32
	handler := func(event Event) error {
		atomic.AddInt32(&received, 1)
		return nil
	}

	bus.Subscribe("test-topic", handler)

	// Publish multiple events
	for i := 0; i < 5; i++ {
		bus.Publish("test-topic", i)
	}

	// Wait for events to be processed
	time.Sleep(200 * time.Millisecond)

	if atomic.LoadInt32(&received) != 5 {
		t.Errorf("expected 5 events received, got %d", received)
	}
}

func TestDynamicEventBus_Unsubscribe(t *testing.T) {
	bus := NewDynamicEventBus(DefaultDynamicEventBusConfig())
	defer bus.Close()

	var received int32
	handler := func(event Event) error {
		atomic.AddInt32(&received, 1)
		return nil
	}

	sub := bus.Subscribe("test-topic", handler)

	// Publish first event
	bus.Publish("test-topic", "test1")

	// Wait for event to be processed
	time.Sleep(50 * time.Millisecond)

	// Unsubscribe
	bus.Unsubscribe(sub)

	// Publish second event
	bus.Publish("test-topic", "test2")

	// Wait for event to be processed
	time.Sleep(50 * time.Millisecond)

	// Should only receive first event
	if atomic.LoadInt32(&received) != 1 {
		t.Errorf("expected 1 event received after unsubscribe, got %d", received)
	}
}

func TestDynamicEventBus_GetStats(t *testing.T) {
	bus := NewDynamicEventBus(DefaultDynamicEventBusConfig())
	defer bus.Close()

	// Publish some events
	handler := func(event Event) error {
		return nil
	}

	bus.Subscribe("test-topic", handler)

	for i := 0; i < 10; i++ {
		bus.Publish("test-topic", i)
	}

	// Wait for events to be processed
	time.Sleep(100 * time.Millisecond)

	stats := bus.GetStats()

	if stats.EventCount != 10 {
		t.Errorf("expected EventCount 10, got %d", stats.EventCount)
	}
}

func TestDynamicEventBus_Close(t *testing.T) {
	bus := NewDynamicEventBus(DefaultDynamicEventBusConfig())

	// Publish some events before closing
	handler := func(event Event) error {
		return nil
	}

	bus.Subscribe("test-topic", handler)
	bus.Publish("test-topic", "test1")

	// Wait for events to be processed
	time.Sleep(50 * time.Millisecond)

	// Close the bus
	err := bus.Close()
	if err != nil {
		t.Fatalf("failed to close bus: %v", err)
	}

	// After close, the bus should be shut down gracefully
	// We can't publish after close as it would panic
}

func TestDynamicEventBus_ConcurrentAccess(t *testing.T) {
	bus := NewDynamicEventBus(DefaultDynamicEventBusConfig())
	defer bus.Close()

	var received int32
	handler := func(event Event) error {
		atomic.AddInt32(&received, 1)
		return nil
	}

	var wg sync.WaitGroup
	const numOperations = 100

	// Subscribe concurrently
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			bus.Subscribe("test-topic", handler)
		}()
	}

	// Publish concurrently
	for i := 0; i < numOperations; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			bus.Publish("test-topic", i)
		}(i)
	}

	wg.Wait()

	// Wait for events to be processed
	time.Sleep(200 * time.Millisecond)

	if atomic.LoadInt32(&received) == 0 {
		t.Error("expected some events to be received")
	}
}

func TestNewDefaultDynamicEventBus(t *testing.T) {
	bus := NewDefaultDynamicEventBus()
	if bus == nil {
		t.Fatal("expected DynamicEventBus to be created")
	}
	defer bus.Close()

	if bus.cfg.InitialQueueSize != 1000 {
		t.Errorf("expected InitialQueueSize 1000, got %d", bus.cfg.InitialQueueSize)
	}
}
