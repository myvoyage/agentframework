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
	"testing"
	"time"


	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAdapterManager tests the adapter manager.
func TestAdapterManager(t *testing.T) {
	ctx := context.Background()

	// Create adapter manager
	mgr := NewAdapterManager()
	assert.NotNil(t, mgr)

	// Create and register Zigbee adapter
	zigbeeAdapter := adapters.NewZigbeeAdapter()
	err := zigbeeAdapter.Initialize(ctx, ProtocolConfig{
		Type: ProtocolZigbee,
		Hardware: HardwareConfig{
			Type:    "websocket",
			Timeout: 5000,
		},
		Metadata: map[string]string{
			"broker_url": "ws://localhost:8000/mqtt",
		},
	})
	require.NoError(t, err)

	err = mgr.RegisterAdapter(zigbeeAdapter)
	assert.NoError(t, err)

	// Get adapter
	adapter, err := mgr.GetAdapter(ProtocolZigbee)
	assert.NoError(t, err)
	assert.NotNil(t, adapter)
	assert.Equal(t, ProtocolZigbee, adapter.Type())

	// Get non-existent adapter
	_, err = mgr.GetAdapter(ProtocolZWave)
	assert.Error(t, err)
}

// TestWorkflowEngine tests the workflow engine.
func TestWorkflowEngine(t *testing.T) {
	ctx := context.Background()

	// Create workflow engine
	mgr := NewAdapterManager()
	engine := NewWorkflowEngine(mgr)
	assert.NotNil(t, engine)

	// Start engine
	err := engine.Start(ctx)
	assert.NoError(t, err)
	defer func() {
		_ = engine.Stop(ctx)
	}()

	// Register scenario
	scenario := &Scenario{
		ID:          "test-scenario",
		Name:        "Test Scenario",
		Description: "A test scenario",
		Actions: []Action{
			{
				Type:      ActionTypeNotification,
				Title:     "Test Notification",
				Message:   "This is a test",
			},
		},
	}

	err = engine.RegisterScenario(scenario)
	assert.NoError(t, err)

	// Get scenario
	retrieved, err := engine.GetScenario("test-scenario")
	assert.NoError(t, err)
	assert.Equal(t, "Test Scenario", retrieved.Name)

	// List scenarios
	scenarios := engine.ListScenarios()
	assert.GreaterOrEqual(t, len(scenarios), 1)

	// Execute scenario
	err = engine.ExecuteScenario(ctx, "test-scenario")
	assert.NoError(t, err)

	// Unregister scenario
	err = engine.UnregisterScenario("test-scenario")
	assert.NoError(t, err)

	// Verify scenario is removed
	_, err = engine.GetScenario("test-scenario")
	assert.Error(t, err)
}

// TestAutomationRule tests automation rules.
func TestAutomationRule(t *testing.T) {
	ctx := context.Background()

	// Create workflow engine
	mgr := NewAdapterManager()
	engine := NewWorkflowEngine(mgr)

	err := engine.Start(ctx)
	assert.NoError(t, err)
	defer func() {
		_ = engine.Stop(ctx)
	}()

	// Register rule
	rule := &AutomationRule{
		ID:      "test-rule",
		Name:    "Test Rule",
		Enabled: true,
		Triggers: []Trigger{
			{
				Type:  TriggerTypeEvent,
				Event: "test_event",
			},
		},
		Actions: []Action{
			{
				Type:      ActionTypeNotification,
				Title:     "Test Action",
				Message:   "Rule triggered",
			},
		},
	}

	err = engine.RegisterRule(rule)
	assert.NoError(t, err)

	// List rules
	rules := engine.ListRules()
	assert.GreaterOrEqual(t, len(rules), 1)

	// Unregister rule
	err = engine.UnregisterRule("test-rule")
	assert.NoError(t, err)
}

// TestWorkflowRegistration tests workflow registration.
func TestWorkflowRegistration(t *testing.T) {
	ctx := context.Background()

	mgr := NewAdapterManager()
	engine := NewWorkflowEngine(mgr)

	err := engine.Start(ctx)
	assert.NoError(t, err)
	defer func() {
		_ = engine.Stop(ctx)
	}()

	// Register workflow
	workflow := &Workflow{
		ID:          "test-workflow",
		Name:        "Test Workflow",
		Description: "A test workflow",
		Enabled:     false,
		Triggers: []Trigger{
			{
				Type:     TriggerTypeSchedule,
				Interval: 60,
			},
		},
		Actions: []Action{
			{
				Type:      ActionTypeNotification,
				Title:     "Workflow Action",
				Message:   "Workflow executed",
			},
		},
	}

	err = engine.RegisterWorkflow(workflow)
	assert.NoError(t, err)

	// Get workflow
	retrieved, err := engine.GetWorkflow("test-workflow")
	assert.NoError(t, err)
	assert.Equal(t, "Test Workflow", retrieved.Name)

	// Enable workflow
	err = engine.EnableWorkflow(ctx, "test-workflow")
	assert.NoError(t, err)

	// Check workflow is enabled
	enabledWorkflow, _ := engine.GetWorkflow("test-workflow")
	assert.True(t, enabledWorkflow.Enabled)

	// Disable workflow
	err = engine.DisableWorkflow("test-workflow")
	assert.NoError(t, err)

	// Check workflow is disabled
	disabledWorkflow, _ := engine.GetWorkflow("test-workflow")
	assert.False(t, disabledWorkflow.Enabled)
}

// TestEventBus tests the event bus.
func TestEventBus(t *testing.T) {
	ctx := context.Background()

	// Create event bus
	bus := NewEventBus()
	assert.NotNil(t, bus)

	// Start event bus
	err := bus.Start(ctx)
	assert.NoError(t, err)
	defer func() {
		_ = bus.Stop(ctx)
	}()

	// Subscribe to event
	received := false
	err = bus.Subscribe("test_event", func(event Event) {
		received = true
	})
	assert.NoError(t, err)

	// Publish event
	err = bus.Publish(Event{
		Type:    "test_event",
		Source:  "test",
		Payload: map[string]interface{}{"test": "data"},
	})
	assert.NoError(t, err)

	// Wait for event to be processed
	time.Sleep(100 * time.Millisecond)
	assert.True(t, received)
}

// TestTaskScheduler tests the task scheduler.
func TestTaskScheduler(t *testing.T) {
	ctx := context.Background()

	// Create scheduler
	scheduler := NewTaskScheduler()
	assert.NotNil(t, scheduler)

	// Start scheduler
	err := scheduler.Start(ctx)
	assert.NoError(t, err)
	defer func() {
		_ = scheduler.Stop(ctx)
	}()

	// Schedule task
	executed := false
	err = scheduler.Schedule("test-task", "0 * * * *", func() {
		executed = true
	})
	assert.NoError(t, err)
}

// TestGetProtocolFromDeviceID tests protocol extraction.
func TestGetProtocolFromDeviceID(t *testing.T) {
	tests := []struct {
		deviceID   string
		expected   ProtocolType
	}{
		{"zigbee-node-001", ProtocolZigbee},
		{"zwave-node-002", ProtocolZWave},
		{"thread-00:11:22:33:44:55", ProtocolThread},
		{"nearlink-001", ProtocolNearLink},
		{"unknown-device", ProtocolUnknown},
		{"short", ProtocolUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.deviceID, func(t *testing.T) {
			result := GetProtocolFromDeviceID(tt.deviceID)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestMultiProtocolAdapterManager tests managing multiple protocols.
func TestMultiProtocolAdapterManager(t *testing.T) {
	ctx := context.Background()

	// Create adapter manager
	mgr := NewAdapterManager()

	// Register multiple adapters
	adapters := []struct {
		protocol ProtocolType
		adapter  ProtocolAdapter
	}{
		{ProtocolZigbee, adapters.NewZigbeeAdapter()},
		{ProtocolZWave, adapters.NewZWaveAdapter()},
		{ProtocolThread, adapters.NewThreadAdapter()},
		{ProtocolNearLink, adapters.NewNearLinkAdapter()},
	}

	for _, a := range adapters {
		err := a.adapter.Initialize(ctx, ProtocolConfig{
			Type: a.protocol,
			Hardware: HardwareConfig{
				Type:    "mock",
				Timeout: 5000,
			},
		})
		require.NoError(t, err)

		err = mgr.RegisterAdapter(a.adapter)
		assert.NoError(t, err)
	}

	// Verify all adapters are registered
	for _, a := range adapters {
		adapter, err := mgr.GetAdapter(a.protocol)
		assert.NoError(t, err)
		assert.NotNil(t, adapter)
		assert.Equal(t, a.protocol, adapter.Type())
	}
}

// TestScenarioExecutionOrder tests scenario execution order.
func TestScenarioExecutionOrder(t *testing.T) {
	ctx := context.Background()

	mgr := NewAdapterManager()
	engine := NewWorkflowEngine(mgr)

	err := engine.Start(ctx)
	assert.NoError(t, err)
	defer func() {
		_ = engine.Stop(ctx)
	}()

	// Create scenario with ordered actions
	executionOrder := make([]string, 0)
	scenario := &Scenario{
		ID:   "ordered-scenario",
		Name: "Ordered Scenario",
		Actions: []Action{
			{
				Type: ActionTypeDelay,
				Value: 10 * time.Millisecond,
			},
			{
				Type: ActionTypeSetVariable,
				Metadata: map[string]interface{}{
					"step": "1",
				},
			},
			{
				Type: ActionTypeDelay,
				Value: 10 * time.Millisecond,
			},
			{
				Type: ActionTypeSetVariable,
				Metadata: map[string]interface{}{
					"step": "2",
				},
			},
		},
	}

	// Mock execute action to record order
	originalExecuteAction := engine.executeAction
	engine.executeAction = func(ctx context.Context, action Action) error {
		if action.Type == ActionTypeSetVariable {
			executionOrder = append(executionOrder, action.Metadata["step"].(string))
		}
		return originalExecuteAction(ctx, action)
	}

	err = engine.RegisterScenario(scenario)
	assert.NoError(t, err)

	err = engine.ExecuteScenario(ctx, "ordered-scenario")
	assert.NoError(t, err)

	// Verify execution order
	assert.Equal(t, []string{"1", "2"}, executionOrder)
}

// BenchmarkScenarioExecution benchmarks scenario execution.
func BenchmarkScenarioExecution(b *testing.B) {
	ctx := context.Background()

	mgr := NewAdapterManager()
	engine := NewWorkflowEngine(mgr)

	_ = engine.Start(ctx)
	defer func() {
		_ = engine.Stop(ctx)
	}()

	// Create scenario with multiple actions
	scenario := &Scenario{
		ID:   "benchmark-scenario",
		Name: "Benchmark Scenario",
		Actions: make([]Action, 10),
	}

	for i := 0; i < 10; i++ {
		scenario.Actions[i] = Action{
			Type: ActionTypeNotification,
			Title: "Benchmark Action",
		}
	}

	_ = engine.RegisterScenario(scenario)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = engine.ExecuteScenario(ctx, "benchmark-scenario")
	}
}

// BenchmarkEventPublish benchmarks event publishing.
func BenchmarkEventPublish(b *testing.B) {
	ctx := context.Background()

	bus := NewEventBus()
	_ = bus.Start(ctx)
	defer func() {
		_ = bus.Stop(ctx)
	}()

	// Subscribe multiple handlers
	for i := 0; i < 10; i++ {
		_ = bus.Subscribe("test_event", func(event Event) {
			// Handle event
		})
	}

	event := Event{
		Type:    "test_event",
		Source:  "benchmark",
		Payload: map[string]interface{}{"test": "data"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = bus.Publish(event)
	}
}
