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

package agent

import (
	"context"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"

	"AgentFramework/agent/errors"
)

// WorkerRole defines the role of a WorkerAgent
type WorkerRole string

// Common worker roles
const (
	RoleDeveloper  WorkerRole = "developer"
	RoleBrowser    WorkerRole = "browser"
	RoleDocument   WorkerRole = "document"
	RoleMultiModal WorkerRole = "multi_modal"
	RoleResearcher WorkerRole = "researcher"
	RoleWriter     WorkerRole = "writer"
	RoleReviewer   WorkerRole = "reviewer"
	RoleAnalyst    WorkerRole = "analyst"
)

// WorkerAgent is an agent that specializes in a specific role
type WorkerAgent interface {
	Agent

	// Role returns the specialized role of the worker
	Role() WorkerRole

	// Capabilities returns the list of capabilities this worker has
	Capabilities() []string
}

// BaseWorkerAgent provides common functionality for all worker agents
type BaseWorkerAgent struct {
	name          string
	role          WorkerRole
	capabilities  []string
	model         ChatModel
	tools         map[string]tool.InvokableTool
	instructions  string
	thread        *Thread
	memoryManager *MemoryManager
}

// NewBaseWorkerAgent creates a new BaseWorkerAgent instance
func NewBaseWorkerAgent(name string, role WorkerRole, capabilities []string, model ChatModel, instructions string, tools []tool.BaseTool, opts ...MemoryOptions) (*BaseWorkerAgent, error) {
	toolMap := make(map[string]tool.InvokableTool)
	for _, t := range tools {
		if inv, ok := t.(tool.InvokableTool); ok {
			info, err := t.Info(context.Background())
			if err != nil {
				continue
			}
			toolMap[info.Name] = inv
		}
	}

	// Create memory manager with provided or default options
	var memoryManager *MemoryManager
	if len(opts) > 0 {
		memoryManager = NewMemoryManager(opts[0])
	} else {
		memoryManager = DefaultMemoryManager()
	}

	return &BaseWorkerAgent{
		name:          name,
		role:          role,
		capabilities:  capabilities,
		model:         model,
		tools:         toolMap,
		instructions:  instructions,
		thread:        &Thread{ID: name},
		memoryManager: memoryManager,
	}, nil
}

// Name returns the name of the agent
func (a *BaseWorkerAgent) Name() string {
	return a.name
}

// Role returns the role of the worker
func (a *BaseWorkerAgent) Role() WorkerRole {
	return a.role
}

// Capabilities returns the capabilities of the worker
func (a *BaseWorkerAgent) Capabilities() []string {
	return a.capabilities
}

// SetMemoryOptions updates the memory management options
func (a *BaseWorkerAgent) SetMemoryOptions(opts MemoryOptions) {
	a.memoryManager.SetOptions(opts)
}

// GetMemoryOptions returns the current memory management options
func (a *BaseWorkerAgent) GetMemoryOptions() MemoryOptions {
	return a.memoryManager.GetOptions()
}

// ClearHistory clears the message history
func (a *BaseWorkerAgent) ClearHistory() {
	a.thread.Messages = a.memoryManager.ClearHistory()
}

// Run executes the worker agent with memory management
func (a *BaseWorkerAgent) Run(ctx context.Context, input string, opts ...model.Option) (*schema.Message, error) {
	userMsg := &schema.Message{
		Role:    schema.User,
		Content: input,
	}

	// Build messages including instructions and history
	messages := []*schema.Message{
		{Role: schema.System, Content: a.instructions},
	}

	// Add message history if enabled
	if a.memoryManager.GetOptions().EnableTrimming && len(a.thread.Messages) > 0 {
		messages = append(messages, a.thread.Messages...)
	}

	messages = append(messages, userMsg)

	resp, err := a.model.Generate(ctx, messages, opts...)
	if err != nil {
		return nil, err
	}

	// Add messages to history for future context
	if a.memoryManager.GetOptions().EnableTrimming {
		a.thread.Messages = append(a.thread.Messages, userMsg, resp)
		// Apply memory management
		a.thread.Messages = a.memoryManager.LimitHistory(a.thread.Messages)
	}

	return resp, nil
}

// DeveloperAgent is a worker agent that specializes in writing and executing code
type DeveloperAgent struct {
	*BaseWorkerAgent
}

// NewDeveloperAgent creates a new DeveloperAgent instance
func NewDeveloperAgent(name string, model ChatModel, instructions string, tools []tool.BaseTool, opts ...MemoryOptions) (*DeveloperAgent, error) {
	base, err := NewBaseWorkerAgent(
		name,
		RoleDeveloper,
		[]string{"write_code", "execute_code", "debug_code", "analyze_code"},
		model,
		instructions,
		tools,
		opts...,
	)
	if err != nil {
		return nil, err
	}

	return &DeveloperAgent{BaseWorkerAgent: base}, nil
}

// BrowserAgent is a worker agent that specializes in browsing the web
type BrowserAgent struct {
	*BaseWorkerAgent
}

// NewBrowserAgent creates a new BrowserAgent instance
func NewBrowserAgent(name string, model ChatModel, instructions string, tools []tool.BaseTool, opts ...MemoryOptions) (*BrowserAgent, error) {
	base, err := NewBaseWorkerAgent(
		name,
		RoleBrowser,
		[]string{"search_web", "browse_url", "extract_content", "click_element"},
		model,
		instructions,
		tools,
		opts...,
	)
	if err != nil {
		return nil, err
	}

	return &BrowserAgent{BaseWorkerAgent: base}, nil
}

// DocumentAgent is a worker agent that specializes in handling documents
type DocumentAgent struct {
	*BaseWorkerAgent
}

// NewDocumentAgent creates a new DocumentAgent instance
func NewDocumentAgent(name string, model ChatModel, instructions string, tools []tool.BaseTool, opts ...MemoryOptions) (*DocumentAgent, error) {
	base, err := NewBaseWorkerAgent(
		name,
		RoleDocument,
		[]string{"create_document", "edit_document", "read_document", "convert_document", "summarize_document"},
		model,
		instructions,
		tools,
		opts...,
	)
	if err != nil {
		return nil, err
	}

	return &DocumentAgent{BaseWorkerAgent: base}, nil
}

// MultiModalAgent is a worker agent that specializes in handling multi-modal content
type MultiModalAgent struct {
	*BaseWorkerAgent
}

// NewMultiModalAgent creates a new MultiModalAgent instance
func NewMultiModalAgent(name string, model ChatModel, instructions string, tools []tool.BaseTool, opts ...MemoryOptions) (*MultiModalAgent, error) {
	base, err := NewBaseWorkerAgent(
		name,
		RoleMultiModal,
		[]string{"process_image", "generate_image", "process_audio", "generate_audio", "process_video", "generate_video"},
		model,
		instructions,
		tools,
		opts...,
	)
	if err != nil {
		return nil, err
	}

	return &MultiModalAgent{BaseWorkerAgent: base}, nil
}

// ResearcherAgent is a worker agent that specializes in research
type ResearcherAgent struct {
	*BaseWorkerAgent
}

// NewResearcherAgent creates a new ResearcherAgent instance
func NewResearcherAgent(name string, model ChatModel, instructions string, tools []tool.BaseTool, opts ...MemoryOptions) (*ResearcherAgent, error) {
	base, err := NewBaseWorkerAgent(
		name,
		RoleResearcher,
		[]string{"search_information", "analyze_data", "synthesize_results", "cite_sources"},
		model,
		instructions,
		tools,
		opts...,
	)
	if err != nil {
		return nil, err
	}

	return &ResearcherAgent{BaseWorkerAgent: base}, nil
}

// WriterAgent is a worker agent that specializes in writing content
type WriterAgent struct {
	*BaseWorkerAgent
}

// NewWriterAgent creates a new WriterAgent instance
func NewWriterAgent(name string, model ChatModel, instructions string, tools []tool.BaseTool, opts ...MemoryOptions) (*WriterAgent, error) {
	base, err := NewBaseWorkerAgent(
		name,
		RoleWriter,
		[]string{"write_content", "edit_content", "summarize_content", "format_content"},
		model,
		instructions,
		tools,
		opts...,
	)
	if err != nil {
		return nil, err
	}

	return &WriterAgent{BaseWorkerAgent: base}, nil
}

// ReviewerAgent is a worker agent that specializes in reviewing content
type ReviewerAgent struct {
	*BaseWorkerAgent
}

// NewReviewerAgent creates a new ReviewerAgent instance
func NewReviewerAgent(name string, model ChatModel, instructions string, tools []tool.BaseTool, opts ...MemoryOptions) (*ReviewerAgent, error) {
	base, err := NewBaseWorkerAgent(
		name,
		RoleReviewer,
		[]string{"review_content", "provide_feedback", "check_quality", "verify_facts"},
		model,
		instructions,
		tools,
		opts...,
	)
	if err != nil {
		return nil, err
	}

	return &ReviewerAgent{BaseWorkerAgent: base}, nil
}

// Workforce is a collection of worker agents that can collaborate on tasks
type Workforce interface {
	// Name returns the name of the workforce
	Name() string

	// AddWorker adds a worker agent to the workforce
	AddWorker(worker WorkerAgent)

	// RemoveWorker removes a worker agent from the workforce
	RemoveWorker(name string)

	// GetWorker returns a worker agent by name
	GetWorker(name string) (WorkerAgent, bool)

	// GetWorkersByRole returns all workers with the specified role
	GetWorkersByRole(role WorkerRole) []WorkerAgent

	// GetWorkersByCapability returns all workers with the specified capability
	GetWorkersByCapability(capability string) []WorkerAgent

	// ListWorkers returns all workers in the workforce
	ListWorkers() []WorkerAgent

	// AssignTask assigns a task to a specific worker
	AssignTask(ctx context.Context, workerName string, task string, opts ...model.Option) (*schema.Message, error)

	// AssignTaskByRole assigns a task to the first available worker with the specified role
	AssignTaskByRole(ctx context.Context, role WorkerRole, task string, opts ...model.Option) (*schema.Message, error)

	// AssignTaskByCapability assigns a task to the first available worker with the specified capability
	AssignTaskByCapability(ctx context.Context, capability string, task string, opts ...model.Option) (*schema.Message, error)
}

// SimpleWorkforce is a simple implementation of the Workforce interface
type SimpleWorkforce struct {
	name    string
	workers map[string]WorkerAgent
	roleMap map[WorkerRole][]WorkerAgent
	capMap  map[string][]WorkerAgent
}

// NewSimpleWorkforce creates a new SimpleWorkforce instance
func NewSimpleWorkforce(name string) *SimpleWorkforce {
	return &SimpleWorkforce{
		name:    name,
		workers: make(map[string]WorkerAgent),
		roleMap: make(map[WorkerRole][]WorkerAgent),
		capMap:  make(map[string][]WorkerAgent),
	}
}

// Name returns the name of the workforce
func (w *SimpleWorkforce) Name() string {
	return w.name
}

// AddWorker adds a worker agent to the workforce
func (w *SimpleWorkforce) AddWorker(worker WorkerAgent) {
	// Add to workers map
	w.workers[worker.Name()] = worker

	// Add to role map
	w.roleMap[worker.Role()] = append(w.roleMap[worker.Role()], worker)

	// Add to capability map
	for _, cap := range worker.Capabilities() {
		w.capMap[cap] = append(w.capMap[cap], worker)
	}
}

// RemoveWorker removes a worker agent from the workforce
func (w *SimpleWorkforce) RemoveWorker(name string) {
	// Get worker
	worker, ok := w.workers[name]
	if !ok {
		return
	}

	// Remove from workers map
	delete(w.workers, name)

	// Remove from role map
	role := worker.Role()
	for i, workerInMap := range w.roleMap[role] {
		if workerInMap.Name() == name {
			w.roleMap[role] = append(w.roleMap[role][:i], w.roleMap[role][i+1:]...)
			break
		}
	}

	// Remove from capability map
	for _, cap := range worker.Capabilities() {
		for i, workerInMap := range w.capMap[cap] {
			if workerInMap.Name() == name {
				w.capMap[cap] = append(w.capMap[cap][:i], w.capMap[cap][i+1:]...)
				break
			}
		}
	}
}

// GetWorker returns a worker agent by name
func (w *SimpleWorkforce) GetWorker(name string) (WorkerAgent, bool) {
	worker, ok := w.workers[name]
	return worker, ok
}

// GetWorkersByRole returns all workers with the specified role
func (w *SimpleWorkforce) GetWorkersByRole(role WorkerRole) []WorkerAgent {
	return w.roleMap[role]
}

// GetWorkersByCapability returns all workers with the specified capability
func (w *SimpleWorkforce) GetWorkersByCapability(capability string) []WorkerAgent {
	return w.capMap[capability]
}

// ListWorkers returns all workers in the workforce
func (w *SimpleWorkforce) ListWorkers() []WorkerAgent {
	workers := make([]WorkerAgent, 0, len(w.workers))
	for _, worker := range w.workers {
		workers = append(workers, worker)
	}
	return workers
}

// AssignTask assigns a task to a specific worker
func (w *SimpleWorkforce) AssignTask(ctx context.Context, workerName string, task string, opts ...model.Option) (*schema.Message, error) {
	worker, ok := w.workers[workerName]
	if !ok {
		return nil, errors.New(errors.ErrCodeNotFound, "worker not found")
	}

	return worker.Run(ctx, task, opts...)
}

// AssignTaskByRole assigns a task to the first available worker with the specified role
func (w *SimpleWorkforce) AssignTaskByRole(ctx context.Context, role WorkerRole, task string, opts ...model.Option) (*schema.Message, error) {
	workers := w.roleMap[role]
	if len(workers) == 0 {
		return nil, errors.New(errors.ErrCodeNotFound, "no workers found with specified role")
	}

	// Use the first worker for now
	return workers[0].Run(ctx, task, opts...)
}

// AssignTaskByCapability assigns a task to the first available worker with the specified capability
func (w *SimpleWorkforce) AssignTaskByCapability(ctx context.Context, capability string, task string, opts ...model.Option) (*schema.Message, error) {
	workers := w.capMap[capability]
	if len(workers) == 0 {
		return nil, errors.New(errors.ErrCodeNotFound, "no workers found with specified capability")
	}

	// Use the first worker for now
	return workers[0].Run(ctx, task, opts...)
}
