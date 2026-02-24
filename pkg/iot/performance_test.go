// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later


package iot

import (
	"context"
	"sync"
	"testing"
	"time"

)

// BenchmarkDeviceDiscovery benchmarks device discovery across protocols.
func BenchmarkDeviceDiscovery(b *testing.B) {
	ctx := context.Background()

	adapters := []struct {
		name    string
		adapter ProtocolAdapter
	}{
		{"Zigbee", adapters.NewZigbeeAdapter()},
		{"Thread", adapters.NewThreadAdapter()},
		{"Z-Wave", adapters.NewZWaveAdapter()},
		{"NearLink", adapters.NewNearLinkAdapter()},
	}

	for _, a := range adapters {
		b.Run(a.name, func(b *testing.B) {
			err := a.adapter.Initialize(ctx, ProtocolConfig{
				Type: a.adapter.Type(),
				Hardware: HardwareConfig{
					Type:    "mock",
					Timeout: 5000,
				},
			})
			if err != nil {
				b.Skip("Skipping: initialization failed")
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = a.adapter.DiscoverDevices(ctx, 1*time.Second)
			}
		})
	}
}

// BenchmarkDeviceRead benchmarks reading device attributes.
func BenchmarkDeviceRead(b *testing.B) {
	ctx := context.Background()

	// Create mock device
	device := createMockDevice()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = device.Read(ctx, "state")
	}
}

// BenchmarkDeviceWrite benchmarks writing device attributes.
func BenchmarkDeviceWrite(b *testing.B) {
	ctx := context.Background()

	// Create mock device
	device := createMockDevice()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = device.Write(ctx, "state", "on")
	}
}

// BenchmarkBatchRead benchmarks batch reading attributes.
func BenchmarkBatchRead(b *testing.B) {
	ctx := context.Background()

	device := createBatchDevice()
	attributes := []string{"temp", "humidity", "pressure", "battery", "rssi"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = device.BatchRead(ctx, attributes)
	}
}

// BenchmarkBatchWrite benchmarks batch writing attributes.
func BenchmarkBatchWrite(b *testing.B) {
	ctx := context.Background()

	device := createBatchDevice()
	values := map[string]interface{}{
		"temp":     25.5,
		"humidity": 60,
		"pressure": 1013,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = device.BatchWrite(ctx, values)
	}
}

// BenchmarkEventBus benchmarks event bus performance.
func BenchmarkEventBus(b *testing.B) {
	ctx := context.Background()

	bus := NewEventBus()
	_ = bus.Start(ctx)
	defer func() {
		_ = bus.Stop(ctx)
	}()

	event := Event{
		Type:    "benchmark_event",
		Source:  "test",
		Payload: map[string]interface{}{"data": "test"},
	}

	b.Run("SingleSubscriber", func(b *testing.B) {
		_ = bus.Subscribe("benchmark_event", func(event Event) {})

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = bus.Publish(event)
		}
	})

	b.Run("MultipleSubscribers", func(b *testing.B) {
		for i := 0; i < 100; i++ {
			_ = bus.Subscribe("benchmark_event", func(event Event) {})
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = bus.Publish(event)
		}
	})
}

// BenchmarkScenarioExecution benchmarks scenario execution.
func BenchmarkScenarioExecution(b *testing.B) {
	ctx := context.Background()

	mgr := NewAdapterManager()
	engine := NewWorkflowEngine(mgr)
	_ = engine.Start(ctx)
	defer func() {
		_ = engine.Stop(ctx)
	}

	scenarios := []struct {
		name     string
		actions  int
		scenario *Scenario
	}{
		{
			name:    "Small",
			actions: 5,
			scenario: createScenario(5),
		},
		{
			name:    "Medium",
			actions: 20,
			scenario: createScenario(20),
		},
		{
			name:    "Large",
			actions: 100,
			scenario: createScenario(100),
		},
	}

	for _, s := range scenarios {
		b.Run(s.name, func(b *testing.B) {
			_ = engine.RegisterScenario(s.scenario)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = engine.ExecuteScenario(ctx, s.scenario.ID)
			}
		})
	}
}

// BenchmarkConcurrentDeviceAccess benchmarks concurrent device access.
func BenchmarkConcurrentDeviceAccess(b *testing.B) {
	ctx := context.Background()

	device := createMockDevice()

	b.Run("Read", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, _ = device.Read(ctx, "state")
			}
		})
	})

	b.Run("Write", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = device.Write(ctx, "state", "on")
			}
		})
	})

	b.Run("Mixed", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				if i%2 == 0 {
					_, _ = device.Read(ctx, "state")
				} else {
					_ = device.Write(ctx, "state", "on")
				}
				i++
			}
		})
	})
}

// BenchmarkWorkflowEngine benchmarks workflow engine performance.
func BenchmarkWorkflowEngine(b *testing.B) {
	ctx := context.Background()

	mgr := NewAdapterManager()
	engine := NewWorkflowEngine(mgr)
	_ = engine.Start(ctx)
	defer func() {
		_ = engine.Stop(ctx)
	}

	b.Run("RuleEvaluation", func(b *testing.B) {
		rule := &AutomationRule{
			ID:      "benchmark-rule",
			Name:    "Benchmark Rule",
			Enabled: true,
			Triggers: []Trigger{
				{Type: TriggerTypeEvent, Event: "test"},
			},
			Conditions: []Condition{
				{Type: ConditionTypeDeviceState, DeviceID: "test", Attribute: "state", Value: "on"},
			},
			Actions: []Action{
				{Type: ActionTypeNotification, Title: "Test"},
			},
		}
		_ = engine.RegisterRule(rule)

		event := Event{Type: "test", Source: "test"}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			engine.evaluateRule(ctx, rule, event)
		}
	})
}

// BenchmarkAdapterManager benchmarks adapter manager operations.
func BenchmarkAdapterManager(b *testing.B) {
	mgr := NewAdapterManager()

	// Register multiple adapters
	adapters := []ProtocolAdapter{
		adapters.NewZigbeeAdapter(),
		adapters.NewZWaveAdapter(),
		adapters.NewThreadAdapter(),
		adapters.NewNearLinkAdapter(),
	}

	for _, adapter := range adapters {
		_ = mgr.RegisterAdapter(adapter)
	}

	b.Run("GetAdapter", func(b *testing.B) {
		protocols := []ProtocolType{
			ProtocolZigbee,
			ProtocolZWave,
			ProtocolThread,
			ProtocolNearLink,
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			protocol := protocols[i%len(protocols)]
			_, _ = mgr.GetAdapter(protocol)
		}
	})
}

// BenchmarkMemoryUsage benchmarks memory usage for multiple devices.
func BenchmarkMemoryUsage(b *testing.B) {
	ctx := context.Background()

	deviceCounts := []int{10, 100, 1000, 10000}

	for _, count := range deviceCounts {
		b.Run(fmt.Sprintf("%dDevices", count), func(b *testing.B) {
			b.ReportAllocs()

			devices := make([]IoTDevice, count)
			for i := 0; i < count; i++ {
				devices[i] = createMockDevice()
			}

			// Measure memory for storing devices
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				for _, device := range devices {
					_, _ = device.Read(ctx, "state")
				}
			}
		})
	}
}

// Mock helper functions

type mockDevice struct {
	*BaseDevice
	state   interface{}
	stateMu sync.RWMutex
}

func createMockDevice() *mockDevice {
	info := &DeviceInfo{
		ID:       "mock-device-001",
		Name:     "Mock Device",
		Type:     DeviceTypeSensor,
		Protocol: ProtocolZigbee,
		Status:   DeviceStatusOnline,
		Properties: map[string]interface{}{
			"state": "off",
		},
		LastSeen:    time.Now(),
		Capabilities: []DeviceCapability{CapabilitySensor},
	}

	return &mockDevice{
		BaseDevice: NewBaseDevice(info),
		state:      "off",
	}
}

func (d *mockDevice) Read(ctx context.Context, attribute string) (interface{}, error) {
	d.stateMu.RLock()
	defer d.stateMu.RUnlock()
	return d.state, nil
}

func (d *mockDevice) Write(ctx context.Context, attribute string, value interface{}) error {
	d.stateMu.Lock()
	defer d.stateMu.Unlock()
	d.state = value
	return nil
}

type mockBatchDevice struct {
	*mockDevice
}

func createBatchDevice() *mockBatchDevice {
	base := createMockDevice()
	return &mockBatchDevice{mockDevice: base}
}

func (d *mockBatchDevice) BatchRead(ctx context.Context, attributes []string) (map[string]interface{}, error) {
	results := make(map[string]interface{})
	for _, attr := range attributes {
		results[attr] = "value"
	}
	return results, nil
}

func (d *mockBatchDevice) BatchWrite(ctx context.Context, values map[string]interface{}) error {
	d.stateMu.Lock()
	defer d.stateMu.Unlock()
	for k, v := range values {
		d.state = v
		_ = k
	}
	return nil
}

func createScenario(actionCount int) *Scenario {
	actions := make([]Action, actionCount)
	for i := 0; i < actionCount; i++ {
		actions[i] = Action{
			Type: ActionTypeNotification,
			Title: fmt.Sprintf("Action %d", i),
		}
	}

	return &Scenario{
		ID:   fmt.Sprintf("scenario-%d", actionCount),
		Name: fmt.Sprintf("Benchmark Scenario %d", actionCount),
		Actions: actions,
	}
}
