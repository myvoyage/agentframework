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

// SPDX-License-Identifier: AGPL-3.0-or-later

package collaboration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

// MockAgent is a mock implementation of Agent for testing
type MockAgent struct {
	name         string
	response     string
	shouldFail   bool
	delay        time.Duration
	capabilities []string
}

func (m *MockAgent) Name() string {
	return m.name
}

func (m *MockAgent) Run(ctx context.Context, input string, opts ...model.Option) (*schema.Message, error) {
	if m.delay > 0 {
		time.Sleep(m.delay)
	}

	if m.shouldFail {
		return nil, fmt.Errorf("mock agent error")
	}

	return &schema.Message{
		Role:    schema.Assistant,
		Content: m.response,
	}, nil
}

func (m *MockAgent) Stream(ctx context.Context, input string, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	return nil, nil
}

func (m *MockAgent) GetCapabilities() []string {
	return m.capabilities
}

func (m *MockAgent) GetModel() string {
	return "mock-model"
}

// MockAgentWrapper wraps MockAgent to implement AgentWrapper
type MockAgentWrapper struct {
	*MockAgent
}

func NewMockAgentWrapper(name, response string, capabilities []string) *MockAgentWrapper {
	return &MockAgentWrapper{
		MockAgent: &MockAgent{
			name:         name,
			response:     response,
			capabilities: capabilities,
		},
	}
}

func (m *MockAgentWrapper) Type() string {
	return "mock"
}

func TestAgentTeam_AddMember(t *testing.T) {
	team := NewAgentTeam(TeamConfig{
		Name:        "test-team",
		Description: "Test team",
	})

	agent := NewMockAgentWrapper("agent-1", "response", []string{"coding"})
	member := team.AddMember(agent, "developer", []string{"coding"}, 5)

	if member == nil {
		t.Fatal("Expected member to be added")
	}

	if member.Agent.Name() != "agent-1" {
		t.Errorf("Expected agent name 'agent-1', got '%s'", member.Agent.Name())
	}

	if member.Role != "developer" {
		t.Errorf("Expected role 'developer', got '%s'", member.Role)
	}
}

func TestAgentTeam_RemoveMember(t *testing.T) {
	team := NewAgentTeam(TeamConfig{
		Name:        "test-team",
		Description: "Test team",
	})

	agent := NewMockAgentWrapper("agent-1", "response", []string{"coding"})
	team.AddMember(agent, "developer", []string{"coding"}, 5)

	err := team.RemoveMember("agent-1")
	if err != nil {
		t.Fatalf("Failed to remove member: %v", err)
	}

	_, err = team.GetMember("agent-1")
	if err == nil {
		t.Error("Expected error when getting removed member")
	}
}

func TestAgentTeam_ListMembers(t *testing.T) {
	team := NewAgentTeam(TeamConfig{
		Name:        "test-team",
		Description: "Test team",
	})

	team.AddMember(NewMockAgentWrapper("agent-1", "response", []string{"coding"}), "developer", []string{"coding"}, 5)
	team.AddMember(NewMockAgentWrapper("agent-2", "response", []string{"testing"}), "tester", []string{"testing"}, 3)

	members := team.ListMembers()

	if len(members) != 2 {
		t.Errorf("Expected 2 members, got %d", len(members))
	}
}

func TestMessageBus_PublishSubscribe(t *testing.T) {
	bus := NewMessageBus()
	ctx := context.Background()

	err := bus.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start message bus: %v", err)
	}
	defer bus.Stop()

	ch := make(chan *Message, 10)
	err = bus.Subscribe("agent-1", ch)
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	msg := &Message{
		ID:        "msg-1",
		Type:      MessageTypeTask,
		From:      "agent-2",
		To:        "agent-1",
		Content:   "test message",
		Timestamp: time.Now(),
	}

	err = bus.Publish(msg)
	if err != nil {
		t.Fatalf("Failed to publish message: %v", err)
	}

	select {
	case received := <-ch:
		if received.Content != "test message" {
			t.Errorf("Expected content 'test message', got '%s'", received.Content)
		}
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for message")
	}
}

func TestMessageBus_Broadcast(t *testing.T) {
	bus := NewMessageBus()
	ctx := context.Background()

	err := bus.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start message bus: %v", err)
	}
	defer bus.Stop()

	ch1 := make(chan *Message, 10)
	ch2 := make(chan *Message, 10)

	bus.Subscribe("agent-1", ch1)
	bus.Subscribe("agent-2", ch2)

	msg := &Message{
		ID:        "msg-1",
		Type:      MessageTypeBroadcast,
		From:      "sender",
		To:        "", // Empty = broadcast
		Content:   "broadcast message",
		Timestamp: time.Now(),
	}

	err = bus.Publish(msg)
	if err != nil {
		t.Fatalf("Failed to publish broadcast: %v", err)
	}

	// Both agents should receive the message
	receivedCount := 0
	timeout := time.After(1 * time.Second)

loop:
	for {
		select {
		case <-ch1:
			receivedCount++
		case <-ch2:
			receivedCount++
		case <-timeout:
			break loop
		}

		if receivedCount >= 2 {
			break
		}
	}

	if receivedCount != 2 {
		t.Errorf("Expected 2 receivers to get broadcast, got %d", receivedCount)
	}
}

func TestIntelligentRouter_RoundRobinStrategy(t *testing.T) {
	router := NewIntelligentRouter(RouterConfig{
		DefaultStrategy: StrategyRoundRobin,
	})

	ctx := context.Background()

	// Create mock members
	members := []*TeamMember{
		{
			Agent:        NewMockAgentWrapper("agent-1", "response", []string{"coding"}),
			Capabilities: []string{"coding"},
			State:        StateIdle,
		},
		{
			Agent:        NewMockAgentWrapper("agent-2", "response", []string{"coding"}),
			Capabilities: []string{"coding"},
			State:        StateIdle,
		},
	}

	task := &CollaborativeTask{
		ID:                   "task-1",
		Type:                 "test",
		Input:                "test input",
		RequiredCapabilities: []string{"coding"},
	}

	// First selection should be agent-1
	member, err := router.SelectMember(ctx, members, task)
	if err != nil {
		t.Fatalf("Failed to select member: %v", err)
	}

	if member.Agent.Name() != "agent-1" {
		t.Errorf("Expected first selection to be 'agent-1', got '%s'", member.Agent.Name())
	}

	// Second selection should be agent-2
	member, err = router.SelectMember(ctx, members, task)
	if err != nil {
		t.Fatalf("Failed to select member: %v", err)
	}

	if member.Agent.Name() != "agent-2" {
		t.Errorf("Expected second selection to be 'agent-2', got '%s'", member.Agent.Name())
	}
}

func TestIntelligentRouter_LeastLoadedStrategy(t *testing.T) {
	router := NewIntelligentRouter(RouterConfig{
		DefaultStrategy: StrategyLeastLoaded,
	})

	ctx := context.Background()

	// Create mock members with different loads
	members := []*TeamMember{
		{
			Agent:        NewMockAgentWrapper("agent-1", "response", []string{"coding"}),
			Capabilities: []string{"coding"},
			State:        StateBusy,
			ActiveTasks:  5,
		},
		{
			Agent:        NewMockAgentWrapper("agent-2", "response", []string{"coding"}),
			Capabilities: []string{"coding"},
			State:        StateIdle,
			ActiveTasks:  0,
		},
	}

	task := &CollaborativeTask{
		ID:                   "task-1",
		Type:                 "test",
		Input:                "test input",
		RequiredCapabilities: []string{"coding"},
	}

	// Should select agent-2 (least loaded)
	member, err := router.SelectMember(ctx, members, task)
	if err != nil {
		t.Fatalf("Failed to select member: %v", err)
	}

	if member.Agent.Name() != "agent-2" {
		t.Errorf("Expected least loaded selection to be 'agent-2', got '%s'", member.Agent.Name())
	}
}

func TestIntelligentRouter_CapabilityFiltering(t *testing.T) {
	router := NewIntelligentRouter(RouterConfig{
		DefaultStrategy: StrategyRoundRobin,
	})

	ctx := context.Background()

	// Create mock members with different capabilities
	members := []*TeamMember{
		{
			Agent:        NewMockAgentWrapper("agent-1", "response", []string{"coding"}),
			Capabilities: []string{"coding"},
			State:        StateIdle,
		},
		{
			Agent:        NewMockAgentWrapper("agent-2", "response", []string{"testing"}),
			Capabilities: []string{"testing"},
			State:        StateIdle,
		},
	}

	task := &CollaborativeTask{
		ID:                   "task-1",
		Type:                 "test",
		Input:                "test input",
		RequiredCapabilities: []string{"coding"}, // Only agent-1 has this
	}

	// Should only select agent-1 (has required capability)
	member, err := router.SelectMember(ctx, members, task)
	if err != nil {
		t.Fatalf("Failed to select member: %v", err)
	}

	if member.Agent.Name() != "agent-1" {
		t.Errorf("Expected selection to be 'agent-1' (has required capability), got '%s'", member.Agent.Name())
	}
}

func TestConsensusManager_MajorityConsensus(t *testing.T) {
	ctx := context.Background()

	team := NewAgentTeam(TeamConfig{
		Name:        "test-team",
		Description: "Test team",
	})

	// Add members with different responses
	team.AddMember(NewMockAgentWrapper("agent-1", "response-A", []string{"vote"}), "voter", []string{"vote"}, 1)
	team.AddMember(NewMockAgentWrapper("agent-2", "response-A", []string{"vote"}), "voter", []string{"vote"}, 1)
	team.AddMember(NewMockAgentWrapper("agent-3", "response-B", []string{"vote"}), "voter", []string{"vote"}, 1)

	consensusMgr := NewConsensusManager(ConsensusMajority, team, 10*time.Second)

	task := &CollaborativeTask{
		ID:                   "task-1",
		Type:                 "vote",
		Input:                "vote for A or B",
		RequiredCapabilities: []string{"vote"},
		Timeout:              5 * time.Second,
	}

	// Note: This test won't fully work without a running team and proper task execution
	// It's here to show the structure
	_ = consensusMgr
	_ = task
	_ = ctx
}

func TestOrchestrator_CreateWorkflow(t *testing.T) {
	team := NewAgentTeam(TeamConfig{
		Name:        "test-team",
		Description: "Test team",
	})

	orchestrator := NewOrchestrator(team)

	steps := []OrchestrationStep{
		{
			ID:   "step-1",
			Name: "First step",
			Type: OrchestrationSequential,
			Task: &CollaborativeTask{
				ID:    "task-1",
				Type:  "test",
				Input: "test input",
			},
		},
		{
			ID:   "step-2",
			Name: "Second step",
			Type: OrchestrationSequential,
			Task: &CollaborativeTask{
				ID:    "task-2",
				Type:  "test",
				Input: "test input",
			},
		},
	}

	workflow, err := orchestrator.CreateWorkflow("wf-1", "Test Workflow", OrchestrationSequential, steps)
	if err != nil {
		t.Fatalf("Failed to create workflow: %v", err)
	}

	if workflow.ID != "wf-1" {
		t.Errorf("Expected workflow ID 'wf-1', got '%s'", workflow.ID)
	}

	if workflow.Name != "Test Workflow" {
		t.Errorf("Expected workflow name 'Test Workflow', got '%s'", workflow.Name)
	}

	if len(workflow.Steps) != 2 {
		t.Errorf("Expected 2 steps, got %d", len(workflow.Steps))
	}
}

func TestOrchestrator_DeleteWorkflow(t *testing.T) {
	team := NewAgentTeam(TeamConfig{
		Name:        "test-team",
		Description: "Test team",
	})

	orchestrator := NewOrchestrator(team)

	steps := []OrchestrationStep{
		{
			ID:   "step-1",
			Name: "First step",
			Type: OrchestrationSequential,
			Task: &CollaborativeTask{
				ID:    "task-1",
				Type:  "test",
				Input: "test input",
			},
		},
	}

	_, err := orchestrator.CreateWorkflow("wf-1", "Test Workflow", OrchestrationSequential, steps)
	if err != nil {
		t.Fatalf("Failed to create workflow: %v", err)
	}

	err = orchestrator.DeleteWorkflow("wf-1")
	if err != nil {
		t.Fatalf("Failed to delete workflow: %v", err)
	}

	_, err = orchestrator.GetWorkflow("wf-1")
	if err == nil {
		t.Error("Expected error when getting deleted workflow")
	}
}

func TestAgentWrapper(t *testing.T) {
	mockAgent := &MockAgent{
		name:         "test-agent",
		response:     "test response",
		capabilities: []string{"coding", "testing"},
	}

	wrapper := NewDefaultAgentWrapper(mockAgent, []string{"coding", "testing"}, "gpt-4")

	if wrapper.Name() != "test-agent" {
		t.Errorf("Expected name 'test-agent', got '%s'", wrapper.Name())
	}

	if !wrapper.HasCapability("coding") {
		t.Error("Expected agent to have 'coding' capability")
	}

	if wrapper.HasCapability("writing") {
		t.Error("Expected agent to not have 'writing' capability")
	}

	wrapper.AddCapability("writing")
	if !wrapper.HasCapability("writing") {
		t.Error("Failed to add capability")
	}

	wrapper.RemoveCapability("writing")
	if wrapper.HasCapability("writing") {
		t.Error("Failed to remove capability")
	}
}

func TestRouterCache(t *testing.T) {
	cache := &RouterCache{
		entries: make(map[string]*CacheEntry),
		ttl:     5 * time.Minute,
	}

	member := &TeamMember{
		Agent:        NewMockAgentWrapper("agent-1", "response", []string{"coding"}),
		Capabilities: []string{"coding"},
	}

	// Test Set and Get
	cache.Set("key-1", member)
	retrieved := cache.Get("key-1")

	if retrieved == nil {
		t.Fatal("Failed to retrieve cached entry")
	}

	if retrieved.Agent.Name() != "agent-1" {
		t.Errorf("Expected cached agent name 'agent-1', got '%s'", retrieved.Agent.Name())
	}

	// Test non-existent key
	retrieved = cache.Get("key-2")
	if retrieved != nil {
		t.Error("Expected nil for non-existent key")
	}
}

// Benchmark tests

func BenchmarkIntelligentRouter_SelectMember(b *testing.B) {
	router := NewIntelligentRouter(RouterConfig{
		DefaultStrategy: StrategyIntelligent,
	})

	ctx := context.Background()

	members := make([]*TeamMember, 100)
	for i := 0; i < 100; i++ {
		members[i] = &TeamMember{
			Agent:        NewMockAgentWrapper("agent-"+string(rune(i)), "response", []string{"coding"}),
			Capabilities: []string{"coding"},
			State:        StateIdle,
		}
	}

	task := &CollaborativeTask{
		ID:                   "task-1",
		Type:                 "test",
		Input:                "test input",
		RequiredCapabilities: []string{"coding"},
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = router.SelectMember(ctx, members, task)
	}
}

func BenchmarkMessageBus_Publish(b *testing.B) {
	bus := NewMessageBus()
	ctx := context.Background()

	_ = bus.Start(ctx)
	defer bus.Stop()

	ch := make(chan *Message, 1000)
	_ = bus.Subscribe("agent-1", ch)

	msg := &Message{
		ID:        "msg-1",
		Type:      MessageTypeTask,
		From:      "sender",
		To:        "agent-1",
		Content:   "test message",
		Timestamp: time.Now(),
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = bus.Publish(msg)
	}
}
