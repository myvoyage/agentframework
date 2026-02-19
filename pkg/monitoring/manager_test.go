// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
// Copyright (C) 2025  Agent Framework Contributors
//
// Monitoring Manager Tests
// SPDX-License-Identifier: AGPL-3.0-or-later

package monitoring

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestMonitoringManager_NewMonitoringManager(t *testing.T) {
	config := &MonitoringConfig{
		AgentID:     "test-agent",
		Version:      "1.0.0",
		Environment:  "test",
		Enabled:      true,
		Port:         9090,
	}

	manager, err := NewMonitoringManager(config)
	if err != nil {
		t.Fatalf("failed to create monitoring manager: %v", err)
	}

	if manager == nil {
		t.Fatal("expected manager to be created")
	}

	if manager.agentID != config.AgentID {
		t.Fatalf("expected agent ID %s, got %s", config.AgentID, manager.agentID)
	}

	if manager.version != config.Version {
		t.Fatalf("expected version %s, got %s", config.Version, manager.version)
	}

	if manager.environment != config.Environment {
		t.Fatalf("expected environment %s, got %s", config.Environment, manager.environment)
	}
}

func TestMonitoringManager_NewMonitoringManager_DefaultConfig(t *testing.T) {
	manager, err := NewMonitoringManager(nil)
	if err != nil {
		t.Fatalf("failed to create monitoring manager with default config: %v", err)
	}

	if manager == nil {
		t.Fatal("expected manager to be created with default config")
	}
}

func TestAgentMetrics_RecordAgentCreation(t *testing.T) {
	manager, _ := NewMonitoringManager(nil)
	metrics := manager.NewAgentMetrics("test-agent")

	initialAgents := manager.metrics.AgentTotalAgents
	if initialAgents == nil {
		t.Fatal("expected AgentTotalAgents to be initialized")
	}

	// Just verify the method runs without panic
	metrics.RecordAgentCreation(100 * time.Millisecond)
}

func TestAgentMetrics_RecordAgentStart(t *testing.T) {
	manager, _ := NewMonitoringManager(nil)
	metrics := manager.NewAgentMetrics("test-agent")

	metrics.RecordAgentStart()
	// Just verify the method runs without panic
}

func TestAgentMetrics_RecordAgentStop(t *testing.T) {
	manager, _ := NewMonitoringManager(nil)
	metrics := manager.NewAgentMetrics("test-agent")

	// Start then stop
	metrics.RecordAgentStart()
	metrics.RecordAgentStop()
	// Just verify the methods run without panic
}

func TestModelMetrics_RecordModelActivation(t *testing.T) {
	manager, _ := NewMonitoringManager(nil)
	metrics := manager.NewModelMetrics("gpt-4")

	metrics.RecordModelActivation()
	// Just verify the method runs without panic
}

func TestModelMetrics_RecordModelRequest(t *testing.T) {
	manager, _ := NewMonitoringManager(nil)
	metrics := manager.NewModelMetrics("gpt-4")

	// Record successful request
	metrics.RecordModelRequest(100*time.Millisecond, nil)

	// Record failed request
	metrics.RecordModelRequest(200*time.Millisecond, errors.New("test error"))
	// Just verify the methods run without panic
}

func TestModelMetrics_RecordCacheHit(t *testing.T) {
	manager, _ := NewMonitoringManager(nil)
	metrics := manager.NewModelMetrics("gpt-4")

	metrics.RecordCacheHit()
	// Just verify the method runs without panic
}

func TestSkillMetrics_RecordSkillCall(t *testing.T) {
	manager, _ := NewMonitoringManager(nil)
	metrics := manager.NewSkillMetrics("http-request", "network")

	metrics.RecordSkillCall()
	// Just verify the method runs without panic
}

func TestSkillMetrics_RecordSkillExecution(t *testing.T) {
	manager, _ := NewMonitoringManager(nil)
	metrics := manager.NewSkillMetrics("http-request", "network")

	// Record successful execution
	metrics.RecordSkillExecution(100*time.Millisecond, nil)

	// Record failed execution
	metrics.RecordSkillExecution(200*time.Millisecond, errors.New("test error"))
	// Just verify the methods run without panic
}

func TestWorkflowMetrics_RecordWorkflowStart(t *testing.T) {
	manager, _ := NewMonitoringManager(nil)
	metrics := manager.NewWorkflowMetrics("test-workflow")

	metrics.RecordWorkflowStart()
	// Just verify the method runs without panic
}

func TestWorkflowMetrics_RecordWorkflowComplete(t *testing.T) {
	manager, _ := NewMonitoringManager(nil)
	metrics := manager.NewWorkflowMetrics("test-workflow")

	// Record successful completion
	metrics.RecordWorkflowComplete(5*time.Second, nil)

	// Record failed completion
	metrics.RecordWorkflowComplete(10*time.Second, errors.New("test error"))
	// Just verify the methods run without panic
}

func TestSystemMetrics_NewSystemMetrics(t *testing.T) {
	manager, _ := NewMonitoringManager(nil)
	metrics := manager.NewSystemMetrics(15 * time.Second)

	if metrics == nil {
		t.Fatal("expected system metrics to be created")
	}

	if metrics.updateInterval != 15*time.Second {
		t.Fatalf("expected update interval 15s, got %v", metrics.updateInterval)
	}
}

func TestSystemMetrics_NewSystemMetrics_DefaultInterval(t *testing.T) {
	manager, _ := NewMonitoringManager(nil)
	metrics := manager.NewSystemMetrics(0)

	if metrics == nil {
		t.Fatal("expected system metrics to be created with default interval")
	}

	if metrics.updateInterval != 15*time.Second {
		t.Fatalf("expected default update interval 15s, got %v", metrics.updateInterval)
	}
}

func TestSystemMetrics_StartAndStop(t *testing.T) {
	manager, _ := NewMonitoringManager(nil)
	metrics := manager.NewSystemMetrics(100 * time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	metrics.Start(ctx)

	// Wait a bit for collection
	time.Sleep(150 * time.Millisecond)

	metrics.Stop()
}

func TestMonitoringManager_GetMetrics(t *testing.T) {
	manager, _ := NewMonitoringManager(nil)

	metrics := manager.GetMetrics()
	if metrics == nil {
		t.Fatal("expected metrics to be returned")
	}

	// Check some metrics are initialized
	if metrics.AgentTotalAgents == nil {
		t.Fatal("expected AgentTotalAgents to be initialized")
	}
	if metrics.ModelTotalModels == nil {
		t.Fatal("expected ModelTotalModels to be initialized")
	}
}

func TestMonitoringManager_GetRegistry(t *testing.T) {
	manager, _ := NewMonitoringManager(nil)

	registry := manager.GetRegistry()
	if registry == nil {
		t.Fatal("expected registry to be returned")
	}
}

func TestMonitoringManager_GetAgentMetrics(t *testing.T) {
	manager, _ := NewMonitoringManager(nil)

	metrics := manager.GetAgentMetrics("test-agent")
	if metrics == nil {
		t.Fatal("expected agent metrics to be returned")
	}

	if metrics.agentID != "test-agent" {
		t.Fatalf("expected agent ID test-agent, got %s", metrics.agentID)
	}
}

func TestMonitoringManager_GetModelMetrics(t *testing.T) {
	manager, _ := NewMonitoringManager(nil)

	metrics := manager.GetModelMetrics("gpt-4")
	if metrics == nil {
		t.Fatal("expected model metrics to be returned")
	}

	if metrics.modelID != "gpt-4" {
		t.Fatalf("expected model ID gpt-4, got %s", metrics.modelID)
	}
}

func TestMonitoringManager_GetSkillMetrics(t *testing.T) {
	manager, _ := NewMonitoringManager(nil)

	metrics := manager.GetSkillMetrics("http-request", "network")
	if metrics == nil {
		t.Fatal("expected skill metrics to be returned")
	}

	if metrics.skillID != "http-request" {
		t.Fatalf("expected skill ID http-request, got %s", metrics.skillID)
	}
	if metrics.category != "network" {
		t.Fatalf("expected category network, got %s", metrics.category)
	}
}

func TestMonitoringManager_GetWorkflowMetrics(t *testing.T) {
	manager, _ := NewMonitoringManager(nil)

	metrics := manager.GetWorkflowMetrics("test-workflow")
	if metrics == nil {
		t.Fatal("expected workflow metrics to be returned")
	}

	if metrics.workflowID != "test-workflow" {
		t.Fatalf("expected workflow ID test-workflow, got %s", metrics.workflowID)
	}
}

func TestMonitoringManager_GetSystemMetrics(t *testing.T) {
	manager, _ := NewMonitoringManager(nil)

	metrics := manager.GetSystemMetrics(15 * time.Second)
	if metrics == nil {
		t.Fatal("expected system metrics to be returned")
	}

	if metrics.updateInterval != 15*time.Second {
		t.Fatalf("expected update interval 15s, got %v", metrics.updateInterval)
	}
}

// Benchmark tests
func BenchmarkAgentMetrics_RecordAgentRun(b *testing.B) {
	manager, _ := NewMonitoringManager(nil)
	metrics := manager.NewAgentMetrics("test-agent")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		metrics.RecordAgentRun(100*time.Millisecond, nil)
	}
}

func BenchmarkModelMetrics_RecordModelRequest(b *testing.B) {
	manager, _ := NewMonitoringManager(nil)
	metrics := manager.NewModelMetrics("gpt-4")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		metrics.RecordModelRequest(100*time.Millisecond, nil)
	}
}

func BenchmarkSkillMetrics_RecordSkillCall(b *testing.B) {
	manager, _ := NewMonitoringManager(nil)
	metrics := manager.NewSkillMetrics("http-request", "network")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		metrics.RecordSkillCall()
	}
}

func BenchmarkWorkflowMetrics_RecordWorkflowComplete(b *testing.B) {
	manager, _ := NewMonitoringManager(nil)
	metrics := manager.NewWorkflowMetrics("test-workflow")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		metrics.RecordWorkflowComplete(5*time.Second, nil)
	}
}
