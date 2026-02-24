// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later


// Package iot provides IoT workflow automation and scenario management.
package iot

import (
	"context"
	"fmt"
	"sync"
	"time"

)

// WorkflowEngine manages IoT automation workflows.
type WorkflowEngine struct {
	workflows    map[string]*Workflow
	scenarios    map[string]*Scenario
	rules        map[string]*AutomationRule
	adapterMgr   *AdapterManager
	eventBus     *EventBus
	scheduler    *TaskScheduler
	mutex        sync.RWMutex
	isRunning    bool
}

// NewWorkflowEngine creates a new workflow engine.
func NewWorkflowEngine(adapterMgr *AdapterManager) *WorkflowEngine {
	return &WorkflowEngine{
		workflows:  make(map[string]*Workflow),
		scenarios:  make(map[string]*Scenario),
		rules:      make(map[string]*AutomationRule),
		adapterMgr: adapterMgr,
		eventBus:   NewEventBus(),
		scheduler:  NewTaskScheduler(),
	}
}

// Start starts the workflow engine.
func (e *WorkflowEngine) Start(ctx context.Context) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if e.isRunning {
		return fmt.Errorf("workflow engine already running")
	}

	// Note: EventBus auto-starts in NewEventBus, no need to call Start()

	// Start scheduler
	if err := e.scheduler.Start(ctx); err != nil {
		return fmt.Errorf("failed to start scheduler: %w", err)
	}

	// Start all active workflows
	for _, workflow := range e.workflows {
		if workflow.Enabled {
			go e.runWorkflow(ctx, workflow)
		}
	}

	e.isRunning = true
	return nil
}

// Stop stops the workflow engine.
func (e *WorkflowEngine) Stop(ctx context.Context) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if !e.isRunning {
		return nil
	}

	// Stop scheduler
	if err := e.scheduler.Stop(ctx); err != nil {
		return fmt.Errorf("failed to stop scheduler: %w", err)
	}

	// Note: EventBus will stop when its context is cancelled

	e.isRunning = false
	return nil
}

// RegisterWorkflow registers a workflow.
func (e *WorkflowEngine) RegisterWorkflow(workflow *Workflow) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if _, exists := e.workflows[workflow.ID]; exists {
		return fmt.Errorf("workflow already exists: %s", workflow.ID)
	}

	e.workflows[workflow.ID] = workflow

	// Subscribe to workflow trigger events
	for _, trigger := range workflow.Triggers {
		if trigger.Type == TriggerTypeEvent {
			_ = e.eventBus.Subscribe(trigger.Event, func(ctx context.Context, event Event) {
				if workflow.Enabled {
					go e.executeWorkflow(context.Background(), workflow, event)
				}
			})
		}
	}

	return nil
}

// UnregisterWorkflow unregisters a workflow.
func (e *WorkflowEngine) UnregisterWorkflow(workflowID string) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	workflow, exists := e.workflows[workflowID]
	if !exists {
		return fmt.Errorf("workflow not found: %s", workflowID)
	}

	// Disable workflow
	workflow.Enabled = false

	delete(e.workflows, workflowID)
	return nil
}

// GetWorkflow retrieves a workflow.
func (e *WorkflowEngine) GetWorkflow(workflowID string) (*Workflow, error) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	workflow, exists := e.workflows[workflowID]
	if !exists {
		return nil, fmt.Errorf("workflow not found: %s", workflowID)
	}

	return workflow, nil
}

// ListWorkflows lists all workflows.
func (e *WorkflowEngine) ListWorkflows() []*Workflow {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	workflows := make([]*Workflow, 0, len(e.workflows))
	for _, workflow := range e.workflows {
		workflows = append(workflows, workflow)
	}
	return workflows
}

// EnableWorkflow enables a workflow.
func (e *WorkflowEngine) EnableWorkflow(ctx context.Context, workflowID string) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	workflow, exists := e.workflows[workflowID]
	if !exists {
		return fmt.Errorf("workflow not found: %s", workflowID)
	}

	workflow.Enabled = true
	go e.runWorkflow(ctx, workflow)

	return nil
}

// DisableWorkflow disables a workflow.
func (e *WorkflowEngine) DisableWorkflow(workflowID string) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	workflow, exists := e.workflows[workflowID]
	if !exists {
		return fmt.Errorf("workflow not found: %s", workflowID)
	}

	workflow.Enabled = false
	return nil
}

// RegisterScenario registers a scenario.
func (e *WorkflowEngine) RegisterScenario(scenario *Scenario) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if _, exists := e.scenarios[scenario.ID]; exists {
		return fmt.Errorf("scenario already exists: %s", scenario.ID)
	}

	e.scenarios[scenario.ID] = scenario
	return nil
}

// UnregisterScenario unregisters a scenario.
func (e *WorkflowEngine) UnregisterScenario(scenarioID string) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if _, exists := e.scenarios[scenarioID]; !exists {
		return fmt.Errorf("scenario not found: %s", scenarioID)
	}

	delete(e.scenarios, scenarioID)
	return nil
}

// GetScenario retrieves a scenario.
func (e *WorkflowEngine) GetScenario(scenarioID string) (*Scenario, error) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	scenario, exists := e.scenarios[scenarioID]
	if !exists {
		return nil, fmt.Errorf("scenario not found: %s", scenarioID)
	}

	return scenario, nil
}

// ListScenarios lists all scenarios.
func (e *WorkflowEngine) ListScenarios() []*Scenario {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	scenarios := make([]*Scenario, 0, len(e.scenarios))
	for _, scenario := range e.scenarios {
		scenarios = append(scenarios, scenario)
	}
	return scenarios
}

// ExecuteScenario executes a scenario.
func (e *WorkflowEngine) ExecuteScenario(ctx context.Context, scenarioID string) error {
	scenario, err := e.GetScenario(scenarioID)
	if err != nil {
		return err
	}

	for _, action := range scenario.Actions {
		if err := e.executeAction(ctx, action); err != nil {
			return fmt.Errorf("failed to execute action %s: %w", action.ID, err)
		}
	}

	return nil
}

// RegisterRule registers an automation rule.
func (e *WorkflowEngine) RegisterRule(rule *AutomationRule) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if _, exists := e.rules[rule.ID]; exists {
		return fmt.Errorf("rule already exists: %s", rule.ID)
	}

	e.rules[rule.ID] = rule

	// Subscribe to rule trigger events
	for _, trigger := range rule.Triggers {
		if trigger.Type == TriggerTypeEvent {
			_ = e.eventBus.Subscribe(trigger.Event, func(ctx context.Context, event Event) {
				if rule.Enabled {
					go e.evaluateRule(context.Background(), rule, event)
				}
			})
		}
	}

	return nil
}

// UnregisterRule unregisters a rule.
func (e *WorkflowEngine) UnregisterRule(ruleID string) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if _, exists := e.rules[ruleID]; !exists {
		return fmt.Errorf("rule not found: %s", ruleID)
	}

	delete(e.rules, ruleID)
	return nil
}

// ListRules lists all rules.
func (e *WorkflowEngine) ListRules() []*AutomationRule {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	rules := make([]*AutomationRule, 0, len(e.rules))
	for _, rule := range e.rules {
		rules = append(rules, rule)
	}
	return rules
}

// runWorkflow runs a workflow with its triggers.
func (e *WorkflowEngine) runWorkflow(ctx context.Context, workflow *Workflow) {
	for {
		if !workflow.Enabled {
			return
		}

		for _, trigger := range workflow.Triggers {
			if trigger.Type == TriggerTypeSchedule {
				// Scheduled trigger
				interval := time.Duration(trigger.Interval) * time.Second
				ticker := time.NewTicker(interval)
				defer ticker.Stop()

				select {
				case <-ticker.C:
					_ = e.executeWorkflow(ctx, workflow, Event{})
				case <-ctx.Done():
					return
				}
			}
		}

		time.Sleep(1 * time.Second)
	}
}

// executeWorkflow executes a workflow.
func (e *WorkflowEngine) executeWorkflow(ctx context.Context, workflow *Workflow, event Event) error {
	// Check conditions
	for _, condition := range workflow.Conditions {
		if !e.evaluateCondition(condition, event) {
			return nil // Condition not met
		}
	}

	// Execute actions
	for _, action := range workflow.Actions {
		if err := e.executeAction(ctx, action); err != nil {
			return fmt.Errorf("failed to execute action: %w", err)
		}
	}

	return nil
}

// evaluateRule evaluates and executes a rule.
func (e *WorkflowEngine) evaluateRule(ctx context.Context, rule *AutomationRule, event Event) {
	// Check all conditions
	allConditionsMet := true
	for _, condition := range rule.Conditions {
		if !e.evaluateCondition(condition, event) {
			allConditionsMet = false
			break
		}
	}

	if !allConditionsMet {
		return
	}

	// Execute actions
	for _, action := range rule.Actions {
		_ = e.executeAction(ctx, action)
	}
}

// evaluateCondition evaluates a condition.
func (e *WorkflowEngine) evaluateCondition(condition Condition, event Event) bool {
	switch condition.Type {
	case ConditionTypeDeviceState:
		// Check device state
		device, err := e.adapterMgr.GetDevice(context.Background(), condition.DeviceID)
		if err != nil {
			return false
		}

		state, err := device.Read(context.Background(), condition.Attribute)
		if err != nil {
			return false
		}

		return fmt.Sprintf("%v", state) == condition.Value

	case ConditionTypeTime:
		// Check time condition
		now := time.Now()
		currentTime := now.Format("15:04")
		return currentTime == condition.Value

	case ConditionTypeEvent:
		// Check event condition
		return string(event.Type) == condition.Event

	case ConditionTypeExpression:
		// Evaluate expression (simplified)
		// In a real implementation, this would use an expression evaluator
		return true
	}

	return false
}

// executeAction executes an action.
func (e *WorkflowEngine) executeAction(ctx context.Context, action Action) error {
	switch action.Type {
	case ActionTypeDeviceControl:
		// Control device
		device, err := e.adapterMgr.GetDevice(ctx, action.DeviceID)
		if err != nil {
			return err
		}

		if action.Attribute != "" && action.Value != nil {
			return device.Write(ctx, action.Attribute, action.Value)
		}

	case ActionTypeDelay:
		// Delay
		if duration, ok := action.Value.(time.Duration); ok {
			time.Sleep(duration)
		}

	case ActionTypeNotification:
		// Send notification
		e.eventBus.Publish(Event{
			ID:        generateEventID(),
			Type:      EventType("notification"),
			Source:    "workflow_engine",
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"title":   action.Title,
				"message": action.Message,
			},
		})

	case ActionTypeScenario:
		// Execute scenario
		if scenarioID, ok := action.Value.(string); ok {
			return e.ExecuteScenario(ctx, scenarioID)
		}

	case ActionTypeHTTPRequest:
		// Send HTTP request (placeholder)
		_ = action.Value

	case ActionTypeSetVariable:
		// Set variable (placeholder)
		_ = action.Value
	}

	return nil
}

// Workflow represents an automation workflow.
type Workflow struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Enabled     bool            `json:"enabled"`
	Triggers    []Trigger       `json:"triggers"`
	Conditions  []Condition     `json:"conditions"`
	Actions     []Action        `json:"actions"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// Trigger represents a workflow trigger.
type Trigger struct {
	Type     TriggerType `json:"type"`
	Event    string      `json:"event,omitempty"`
	Interval int         `json:"interval,omitempty"`
	Cron     string      `json:"cron,omitempty"`
}

// TriggerType defines trigger types.
type TriggerType string

const (
	TriggerTypeEvent    TriggerType = "event"
	TriggerTypeSchedule TriggerType = "schedule"
	TriggerTypeManual   TriggerType = "manual"
)

// Condition represents a condition.
type Condition struct {
	Type      ConditionType `json:"type"`
	DeviceID  string        `json:"device_id,omitempty"`
	Attribute string        `json:"attribute,omitempty"`
	Value     string        `json:"value,omitempty"`
	Event     string        `json:"event,omitempty"`
	Expression string       `json:"expression,omitempty"`
}

// ConditionType defines condition types.
type ConditionType string

const (
	ConditionTypeDeviceState  ConditionType = "device_state"
	ConditionTypeTime         ConditionType = "time"
	ConditionTypeEvent        ConditionType = "event"
	ConditionTypeExpression   ConditionType = "expression"
)

// Action represents an action.
type Action struct {
	ID        string                 `json:"id"`
	Type      ActionType             `json:"type"`
	DeviceID  string                 `json:"device_id,omitempty"`
	Attribute string                 `json:"attribute,omitempty"`
	Value     interface{}            `json:"value,omitempty"`
	Title     string                 `json:"title,omitempty"`
	Message   string                 `json:"message,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// ActionType defines action types.
type ActionType string

const (
	ActionTypeDeviceControl ActionType = "device_control"
	ActionTypeDelay         ActionType = "delay"
	ActionTypeNotification  ActionType = "notification"
	ActionTypeScenario      ActionType = "scenario"
	ActionTypeHTTPRequest   ActionType = "http_request"
	ActionTypeSetVariable   ActionType = "set_variable"
)

// Scenario represents a collection of actions.
type Scenario struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Icon        string          `json:"icon,omitempty"`
	Actions     []Action        `json:"actions"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// AutomationRule represents a simple automation rule.
type AutomationRule struct {
	ID         string      `json:"id"`
	Name       string      `json:"name"`
	Enabled    bool        `json:"enabled"`
	Triggers   []Trigger   `json:"triggers"`
	Conditions []Condition `json:"conditions"`
	Actions    []Action    `json:"actions"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// TaskScheduler manages scheduled tasks.
type TaskScheduler struct {
	tasks   map[string]*ScheduledTask
	mutex   sync.RWMutex
	running bool
}

// ScheduledTask represents a scheduled task.
type ScheduledTask struct {
	ID       string
	Schedule string
	Task     func()
	NextRun  time.Time
}

// NewTaskScheduler creates a new task scheduler.
func NewTaskScheduler() *TaskScheduler {
	return &TaskScheduler{
		tasks: make(map[string]*ScheduledTask),
	}
}

// Start starts the scheduler.
func (s *TaskScheduler) Start(ctx context.Context) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.running = true
	return nil
}

// Stop stops the scheduler.
func (s *TaskScheduler) Stop(ctx context.Context) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.running = false
	return nil
}

// Schedule schedules a task.
func (s *TaskScheduler) Schedule(id string, schedule string, task func()) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.tasks[id] = &ScheduledTask{
		ID:       id,
		Schedule: schedule,
		Task:     task,
	}

	return nil
}

// AdapterManager manages IoT adapters.
type AdapterManager struct {
	adapters map[ProtocolType]ProtocolAdapter
	mutex    sync.RWMutex
}

// NewAdapterManager creates a new adapter manager.
func NewAdapterManager() *AdapterManager {
	return &AdapterManager{
		adapters: make(map[ProtocolType]ProtocolAdapter),
	}
}

// RegisterAdapter registers an adapter.
func (m *AdapterManager) RegisterAdapter(adapter ProtocolAdapter) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.adapters[adapter.Type()] = adapter
	return nil
}

// GetAdapter retrieves an adapter.
func (m *AdapterManager) GetAdapter(protocol ProtocolType) (ProtocolAdapter, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	adapter, exists := m.adapters[protocol]
	if !exists {
		return nil, fmt.Errorf("adapter not found: %s", protocol)
	}

	return adapter, nil
}

// GetDevice retrieves a device from any adapter.
func (m *AdapterManager) GetDevice(ctx context.Context, deviceID string) (IoTDevice, error) {
	protocol := GetProtocolFromDeviceID(deviceID)

	adapter, err := m.GetAdapter(protocol)
	if err != nil {
		return nil, err
	}

	return adapter.GetDevice(ctx, deviceID)
}

// ListDevices lists all devices from all adapters.
func (m *AdapterManager) ListDevices(ctx context.Context) ([]IoTDevice, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var allDevices []IoTDevice
	for _, adapter := range m.adapters {
		devices, err := adapter.ListDevices(ctx)
		if err != nil {
			continue
		}
		allDevices = append(allDevices, devices...)
	}

	return allDevices, nil
}
