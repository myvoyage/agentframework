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
	"sync"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

// AgentTeam manages a group of agents working together
type AgentTeam struct {
	name        string
	description string
	members     []*TeamMember
	bus         *MessageBus
	scheduler   TaskScheduler
	router      *IntelligentRouter
	mu          sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
	running     bool
}

// TeamMember represents an agent in the team
type TeamMember struct {
	Agent          AgentWrapper
	Role           string
	Capabilities   []string
	State          MemberState
	Performance    MemberPerformance
	Priority       int // 0-9, higher is more important
	MaxConcurrency int
	ActiveTasks    int
	mu             sync.Mutex
}

// MemberState represents the current state of a team member
type MemberState int

const (
	StateIdle MemberState = iota
	StateBusy
	StateOffline
	StateOverloaded
)

func (s MemberState) String() string {
	switch s {
	case StateIdle:
		return "idle"
	case StateBusy:
		return "busy"
	case StateOffline:
		return "offline"
	case StateOverloaded:
		return "overloaded"
	default:
		return "unknown"
	}
}

// MemberPerformance tracks performance metrics for a team member
type MemberPerformance struct {
	TotalTasks   int64
	SuccessTasks int64
	FailedTasks  int64
	AvgDuration  time.Duration
	AvgCost      float64
	LastUsed     time.Time
	SuccessRate  float64
	ErrorRate    float64
	QualityScore float64 // 0.0-1.0
}

// AgentWrapper wraps the agent.Agent interface with additional metadata
type AgentWrapper interface {
	Name() string
	Type() string
	Run(ctx context.Context, input string, opts ...model.Option) (*schema.Message, error)
	Stream(ctx context.Context, input string, opts ...model.Option) (*schema.StreamReader[*schema.Message], error)
	GetCapabilities() []string
	GetModel() string
}

// TeamConfig configures an AgentTeam
type TeamConfig struct {
	Name          string
	Description   string
	MaxConcurrent int
	RouterConfig  RouterConfig
}

// NewAgentTeam creates a new agent team
func NewAgentTeam(cfg TeamConfig) *AgentTeam {
	ctx, cancel := context.WithCancel(context.Background())

	return &AgentTeam{
		name:        cfg.Name,
		description: cfg.Description,
		members:     make([]*TeamMember, 0),
		bus:         NewMessageBus(),
		scheduler:   NewDefaultTaskScheduler(cfg.MaxConcurrent),
		router:      NewIntelligentRouter(cfg.RouterConfig),
		ctx:         ctx,
		cancel:      cancel,
		running:     false,
	}
}

// Start starts the agent team
func (at *AgentTeam) Start(ctx context.Context) error {
	at.mu.Lock()
	defer at.mu.Unlock()

	if at.running {
		return fmt.Errorf("team already running")
	}

	at.ctx = ctx
	at.running = true

	// Start message bus
	if err := at.bus.Start(ctx); err != nil {
		return fmt.Errorf("failed to start message bus: %w", err)
	}

	// Start all team members
	for _, member := range at.members {
		go at.runMember(ctx, member)
	}

	// Start task scheduler
	go at.runTaskDispatcher(ctx)

	return nil
}

// Stop stops the agent team
func (at *AgentTeam) Stop() error {
	at.mu.Lock()
	defer at.mu.Unlock()

	if !at.running {
		return nil
	}

	at.cancel()
	at.running = false

	// Stop message bus
	at.bus.Stop()

	return nil
}

// AddMember adds an agent to the team
func (at *AgentTeam) AddMember(agent AgentWrapper, role string, capabilities []string, priority int) *TeamMember {
	at.mu.Lock()
	defer at.mu.Unlock()

	member := &TeamMember{
		Agent:          agent,
		Role:           role,
		Capabilities:   capabilities,
		State:          StateIdle,
		Performance:    MemberPerformance{},
		Priority:       priority,
		MaxConcurrency: 1,
		ActiveTasks:    0,
	}

	at.members = append(at.members, member)

	// Subscribe to message bus for this member
	if at.running {
		go at.runMember(at.ctx, member)
	}

	return member
}

// RemoveMember removes an agent from the team
func (at *AgentTeam) RemoveMember(agentName string) error {
	at.mu.Lock()
	defer at.mu.Unlock()

	for i, member := range at.members {
		if member.Agent.Name() == agentName {
			// Unsubscribe from message bus
			at.bus.Unsubscribe(agentName)

			// Remove from members list
			at.members = append(at.members[:i], at.members[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("member not found: %s", agentName)
}

// GetMember returns a team member by name
func (at *AgentTeam) GetMember(name string) (*TeamMember, error) {
	at.mu.RLock()
	defer at.mu.RUnlock()

	for _, member := range at.members {
		if member.Agent.Name() == name {
			return member, nil
		}
	}

	return nil, fmt.Errorf("member not found: %s", name)
}

// ListMembers returns all team members
func (at *AgentTeam) ListMembers() []*TeamMember {
	at.mu.RLock()
	defer at.mu.RUnlock()

	members := make([]*TeamMember, len(at.members))
	copy(members, at.members)
	return members
}

// AssignTask assigns a task to the best available agent
func (at *AgentTeam) AssignTask(ctx context.Context, task *CollaborativeTask) (*TaskResult, error) {
	// Select the best member for this task
	member, err := at.router.SelectMember(ctx, at.members, task)
	if err != nil {
		return nil, fmt.Errorf("failed to select member: %w", err)
	}

	// Submit task to scheduler
	result, err := at.scheduler.Submit(ctx, task, member)
	if err != nil {
		return nil, fmt.Errorf("failed to submit task: %w", err)
	}

	return result, nil
}

// AssignTaskWithStrategy assigns a task using a specific strategy
func (at *AgentTeam) AssignTaskWithStrategy(ctx context.Context, task *CollaborativeTask, strategy RoutingStrategy) (*TaskResult, error) {
	member, err := at.router.SelectMemberWithStrategy(ctx, at.members, task, strategy)
	if err != nil {
		return nil, fmt.Errorf("failed to select member: %w", err)
	}

	result, err := at.scheduler.Submit(ctx, task, member)
	if err != nil {
		return nil, fmt.Errorf("failed to submit task: %w", err)
	}

	return result, nil
}

// BroadcastTask sends a task to all matching agents
func (at *AgentTeam) BroadcastTask(ctx context.Context, task *CollaborativeTask) ([]*TaskResult, error) {
	at.mu.RLock()
	defer at.mu.RUnlock()

	var results []*TaskResult
	var wg sync.WaitGroup
	errChan := make(chan error, len(at.members))
	resultChan := make(chan *TaskResult, len(at.members))

	for _, member := range at.members {
		if !at.isMemberCapable(member, task) {
			continue
		}

		wg.Add(1)
		go func(m *TeamMember) {
			defer wg.Done()

			result, err := at.scheduler.Submit(ctx, task, m)
			if err != nil {
				errChan <- err
				return
			}

			resultChan <- result
		}(member)
	}

	go func() {
		wg.Wait()
		close(errChan)
		close(resultChan)
	}()

	// Collect results
	for err := range errChan {
		if err != nil {
			return nil, fmt.Errorf("task execution error: %w", err)
		}
	}

	for result := range resultChan {
		results = append(results, result)
	}

	return results, nil
}

// CollaborativeTask represents a task that can be executed by multiple agents
type CollaborativeTask struct {
	ID                   string
	Type                 string
	Input                string
	RequiredCapabilities []string
	Priority             int
	MaxRetries           int
	Timeout              time.Duration
	Context              map[string]interface{}
	CollaborationMode    CollaborationMode
	ExpectedAgents       []string // Specific agents that should work on this task
}

// CollaborationMode defines how agents collaborate on a task
type CollaborationMode int

const (
	// ModeSingle - single agent executes the task
	ModeSingle CollaborationMode = iota
	// ModeParallel - multiple agents work in parallel
	ModeParallel
	// ModeSequential - agents work in sequence
	ModeSequential
	// ModeConsensus - agents must reach consensus
	ModeConsensus
)

// TaskResult represents the result of a task execution
type TaskResult struct {
	TaskID    string
	AgentName string
	AgentRole string
	Output    string
	Success   bool
	Error     error
	Duration  time.Duration
	Cost      float64
	Timestamp time.Time
	Metadata  map[string]interface{}
}

// runMember runs the message handler for a team member
func (at *AgentTeam) runMember(ctx context.Context, member *TeamMember) {
	ch := make(chan *Message, 10)
	at.bus.Subscribe(member.Agent.Name(), ch)

	for {
		select {
		case msg := <-ch:
			at.handleMessage(ctx, member, msg)
		case <-ctx.Done():
			return
		}
	}
}

// runTaskDispatcher runs the task dispatcher
func (at *AgentTeam) runTaskDispatcher(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			at.updateMemberStates()
		case <-ctx.Done():
			return
		}
	}
}

// handleMessage handles an incoming message for a team member
func (at *AgentTeam) handleMessage(ctx context.Context, member *TeamMember, msg *Message) {
	switch msg.Type {
	case MessageTypeTask:
		at.executeTask(ctx, member, msg)
	case MessageTypeQuery:
		at.handleQuery(ctx, member, msg)
	case MessageTypeNotification:
		at.handleNotification(ctx, member, msg)
	}
}

// executeTask executes a task for a team member
func (at *AgentTeam) executeTask(ctx context.Context, member *TeamMember, msg *Message) {
	member.mu.Lock()
	if member.ActiveTasks >= member.MaxConcurrency {
		member.State = StateOverloaded
		member.mu.Unlock()
		msg.ReplyChan <- &Message{
			Type:      MessageTypeError,
			Content:   fmt.Sprintf("member %s is overloaded", member.Agent.Name()),
			Timestamp: time.Now(),
		}
		return
	}
	member.ActiveTasks++
	member.State = StateBusy
	member.mu.Unlock()

	defer func() {
		member.mu.Lock()
		member.ActiveTasks--
		if member.ActiveTasks == 0 {
			member.State = StateIdle
		}
		member.mu.Unlock()
	}()

	// Execute the task
	startTime := time.Now()
	result, err := member.Agent.Run(ctx, msg.Content)
	duration := time.Since(startTime)

	// Update performance metrics
	at.updateMemberPerformance(member, result, err, duration)

	// Send result back
	response := &Message{
		Type:      MessageTypeResult,
		Content:   result.Content,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"duration": duration,
			"success":  err == nil,
			"error":    err,
		},
	}

	if err != nil {
		response.Type = MessageTypeError
		response.Content = err.Error()
	}

	msg.ReplyChan <- response
}

// handleQuery handles a query message
func (at *AgentTeam) handleQuery(ctx context.Context, member *TeamMember, msg *Message) {
	// Query member capabilities
	if query, ok := msg.Metadata["query"].(string); ok {
		var response interface{}

		switch query {
		case "capabilities":
			response = member.Capabilities
		case "state":
			response = member.State.String()
		case "performance":
			response = member.Performance
		default:
			response = "unknown query"
		}

		msg.ReplyChan <- &Message{
			Type:      MessageTypeResult,
			Content:   fmt.Sprintf("%v", response),
			Timestamp: time.Now(),
		}
	}
}

// handleNotification handles a notification message
func (at *AgentTeam) handleNotification(ctx context.Context, member *TeamMember, msg *Message) {
	// Handle notifications (log, metrics, etc.)
	// For now, just log it
}

// updateMemberStates updates the state of all members
func (at *AgentTeam) updateMemberStates() {
	at.mu.RLock()
	defer at.mu.RUnlock()

	for _, member := range at.members {
		member.mu.Lock()

		// Update state based on active tasks
		if member.ActiveTasks == 0 && member.State == StateBusy {
			member.State = StateIdle
		} else if member.ActiveTasks >= member.MaxConcurrency {
			member.State = StateOverloaded
		}

		member.mu.Unlock()
	}
}

// updateMemberPerformance updates performance metrics for a member
func (at *AgentTeam) updateMemberPerformance(member *TeamMember, result *schema.Message, err error, duration time.Duration) {
	member.mu.Lock()
	defer member.mu.Unlock()

	perf := &member.Performance
	perf.TotalTasks++
	perf.LastUsed = time.Now()

	if err == nil {
		perf.SuccessTasks++
	} else {
		perf.FailedTasks++
	}

	// Update average duration
	if perf.TotalTasks == 1 {
		perf.AvgDuration = duration
	} else {
		perf.AvgDuration = time.Duration((int64(perf.AvgDuration)*int64(perf.TotalTasks-1) + int64(duration)) / int64(perf.TotalTasks))
	}

	// Calculate success rate
	perf.SuccessRate = float64(perf.SuccessTasks) / float64(perf.TotalTasks)
	perf.ErrorRate = float64(perf.FailedTasks) / float64(perf.TotalTasks)
}

// isMemberCapable checks if a member has the required capabilities
func (at *AgentTeam) isMemberCapable(member *TeamMember, task *CollaborativeTask) bool {
	if len(task.RequiredCapabilities) == 0 {
		return true
	}

	capMap := make(map[string]bool)
	for _, cap := range member.Capabilities {
		capMap[cap] = true
	}

	for _, reqCap := range task.RequiredCapabilities {
		if !capMap[reqCap] {
			return false
		}
	}

	return true
}

// GetPerformanceStats returns performance statistics for the team
func (at *AgentTeam) GetPerformanceStats() *TeamPerformanceStats {
	at.mu.RLock()
	defer at.mu.RUnlock()

	stats := &TeamPerformanceStats{
		TotalMembers:      len(at.members),
		IdleMembers:       0,
		BusyMembers:       0,
		OverloadedMembers: 0,
		OfflineMembers:    0,
		MemberStats:       make(map[string]*MemberPerformance),
	}

	for _, member := range at.members {
		switch member.State {
		case StateIdle:
			stats.IdleMembers++
		case StateBusy:
			stats.BusyMembers++
		case StateOverloaded:
			stats.OverloadedMembers++
		case StateOffline:
			stats.OfflineMembers++
		}

		stats.MemberStats[member.Agent.Name()] = &member.Performance
	}

	return stats
}

// TeamPerformanceStats represents performance statistics for the team
type TeamPerformanceStats struct {
	TotalMembers      int
	IdleMembers       int
	BusyMembers       int
	OverloadedMembers int
	OfflineMembers    int
	MemberStats       map[string]*MemberPerformance
}

// Name returns the team name
func (at *AgentTeam) Name() string {
	return at.name
}

// Description returns the team description
func (at *AgentTeam) Description() string {
	return at.description
}

// AssignBatchTasks assigns multiple tasks to the best available agents
func (at *AgentTeam) AssignBatchTasks(ctx context.Context, tasks []*CollaborativeTask) ([]*TaskResult, error) {
	// Submit all tasks to scheduler with member selection
	results, err := at.scheduler.SubmitBatch(ctx, tasks, at.members, at.router)
	if err != nil {
		return nil, fmt.Errorf("failed to submit batch tasks: %w", err)
	}

	return results, nil
}

