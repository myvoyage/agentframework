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
	"fmt"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

// GraphWorkforce is a workforce implementation that uses our own DAGWorkflow for orchestration
type GraphWorkforce struct {
	name      string
	workforce *SimpleWorkforce
	dag       *DAGWorkflow
}

// NewGraphWorkforce creates a new GraphWorkforce instance
func NewGraphWorkforce(name string) *GraphWorkforce {
	return &GraphWorkforce{
		name:      name,
		workforce: NewSimpleWorkforce(name),
		dag:       NewDAGWorkflow(name),
	}
}

// Name returns the name of the workforce
func (w *GraphWorkforce) Name() string {
	return w.name
}

// AddWorker adds a worker agent to the workforce and creates a corresponding graph node
func (w *GraphWorkforce) AddWorker(worker WorkerAgent) {
	w.workforce.AddWorker(worker)

	// Create a workflow adapter for the worker and add it to the DAG
	adapter := NewWorkforceWorkflowAdapter(worker.Name(), w.workforce, "", "")
	w.dag.AddNode(worker.Name(), adapter)
}

// RemoveWorker removes a worker agent from the workforce and its corresponding graph node
func (w *GraphWorkforce) RemoveWorker(name string) {
	w.workforce.RemoveWorker(name)
	// DAG nodes can't be removed, but they won't be used if no edges point to them
}

// GetWorker returns a worker agent by name
func (w *GraphWorkforce) GetWorker(name string) (WorkerAgent, bool) {
	return w.workforce.GetWorker(name)
}

// GetWorkersByRole returns all workers with the specified role
func (w *GraphWorkforce) GetWorkersByRole(role WorkerRole) []WorkerAgent {
	return w.workforce.GetWorkersByRole(role)
}

// GetWorkersByCapability returns all workers with the specified capability
func (w *GraphWorkforce) GetWorkersByCapability(capability string) []WorkerAgent {
	return w.workforce.GetWorkersByCapability(capability)
}

// ListWorkers returns all workers in the workforce
func (w *GraphWorkforce) ListWorkers() []WorkerAgent {
	return w.workforce.ListWorkers()
}

// AssignTask assigns a task to a specific worker using the graph
func (w *GraphWorkforce) AssignTask(ctx context.Context, workerName string, task string, opts ...model.Option) (*schema.Message, error) {
	// Directly assign the task to the worker
	return w.workforce.AssignTask(ctx, workerName, task, opts...)
}

// AssignTaskByRole assigns a task to the first available worker with the specified role
func (w *GraphWorkforce) AssignTaskByRole(ctx context.Context, role WorkerRole, task string, opts ...model.Option) (*schema.Message, error) {
	return w.workforce.AssignTaskByRole(ctx, role, task, opts...)
}

// AssignTaskByCapability assigns a task to the first available worker with the specified capability
func (w *GraphWorkforce) AssignTaskByCapability(ctx context.Context, capability string, task string, opts ...model.Option) (*schema.Message, error) {
	return w.workforce.AssignTaskByCapability(ctx, capability, task, opts...)
}

// AddEdge adds a directed edge between two workers in the graph
func (w *GraphWorkforce) AddEdge(fromWorker, toWorker string) error {
	// Check if both workers exist
	if _, ok := w.GetWorker(fromWorker); !ok {
		return fmt.Errorf("worker %s not found", fromWorker)
	}
	if _, ok := w.GetWorker(toWorker); !ok {
		return fmt.Errorf("worker %s not found", toWorker)
	}

	// Add edge to DAG
	w.dag.AddEdge(fromWorker, toWorker)
	return nil
}

// AddConditionalEdge adds a conditional edge between a worker and multiple possible next workers
// Note: This is a simplified implementation, as our DAGWorkflow doesn't support conditional edges directly
func (w *GraphWorkforce) AddConditionalEdge(fromWorker string, condition func(ctx context.Context, output *schema.Message) (string, error)) error {
	// Check if worker exists
	if _, ok := w.GetWorker(fromWorker); !ok {
		return fmt.Errorf("worker %s not found", fromWorker)
	}

	// For now, we'll just add the edge to all possible workers
	// In a real implementation, we would need to modify DAGWorkflow to support conditional edges
	return nil
}

// SetStartWorker sets the starting worker for the graph
func (w *GraphWorkforce) SetStartWorker(workerName string) error {
	// Check if worker exists
	if _, ok := w.GetWorker(workerName); !ok {
		return fmt.Errorf("worker %s not found", workerName)
	}

	// Our DAGWorkflow doesn't have a concept of start node, edges define the flow
	return nil
}

// Run executes the graph workflow with the given input
func (w *GraphWorkforce) Run(ctx context.Context, input string, opts ...model.Option) (*schema.Message, error) {
	// Execute the DAG
	return w.dag.Run(ctx, input, opts...)
}

// ParallelWorkforceTask represents a parallel task to be executed by multiple workers
type ParallelWorkforceTask struct {
	WorkerName string      `json:"worker_name"` // Specific worker name, or "" for role/capability
	Role       WorkerRole  `json:"role"`        // Worker role, if WorkerName is ""
	Capability string      `json:"capability"`  // Worker capability, if WorkerName and Role are ""
	Task       string      `json:"task"`        // Task to execute
	Input      interface{} `json:"input"`       // Additional input for the task
}

// ParallelWorkforceResult represents the result of a parallel task execution
type ParallelWorkforceResult struct {
	WorkerName string          `json:"worker_name"`
	Result     *schema.Message `json:"result"`
	Error      string          `json:"error,omitempty"`
}

// ParallelWorkforce is a workforce implementation that supports parallel task execution
type ParallelWorkforce struct {
	name      string
	workforce *SimpleWorkforce
}

// NewParallelWorkforce creates a new ParallelWorkforce instance
func NewParallelWorkforce(name string) *ParallelWorkforce {
	return &ParallelWorkforce{
		name:      name,
		workforce: NewSimpleWorkforce(name),
	}
}

// Name returns the name of the workforce
func (w *ParallelWorkforce) Name() string {
	return w.name
}

// AddWorker adds a worker agent to the workforce
func (w *ParallelWorkforce) AddWorker(worker WorkerAgent) {
	w.workforce.AddWorker(worker)
}

// RemoveWorker removes a worker agent from the workforce
func (w *ParallelWorkforce) RemoveWorker(name string) {
	w.workforce.RemoveWorker(name)
}

// GetWorker returns a worker agent by name
func (w *ParallelWorkforce) GetWorker(name string) (WorkerAgent, bool) {
	return w.workforce.GetWorker(name)
}

// GetWorkersByRole returns all workers with the specified role
func (w *ParallelWorkforce) GetWorkersByRole(role WorkerRole) []WorkerAgent {
	return w.workforce.GetWorkersByRole(role)
}

// GetWorkersByCapability returns all workers with the specified capability
func (w *ParallelWorkforce) GetWorkersByCapability(capability string) []WorkerAgent {
	return w.workforce.GetWorkersByCapability(capability)
}

// ListWorkers returns all workers in the workforce
func (w *ParallelWorkforce) ListWorkers() []WorkerAgent {
	return w.workforce.ListWorkers()
}

// AssignTask assigns a task to a specific worker
func (w *ParallelWorkforce) AssignTask(ctx context.Context, workerName string, task string, opts ...model.Option) (*schema.Message, error) {
	return w.workforce.AssignTask(ctx, workerName, task, opts...)
}

// AssignTaskByRole assigns a task to the first available worker with the specified role
func (w *ParallelWorkforce) AssignTaskByRole(ctx context.Context, role WorkerRole, task string, opts ...model.Option) (*schema.Message, error) {
	return w.workforce.AssignTaskByRole(ctx, role, task, opts...)
}

// AssignTaskByCapability assigns a task to the first available worker with the specified capability
func (w *ParallelWorkforce) AssignTaskByCapability(ctx context.Context, capability string, task string, opts ...model.Option) (*schema.Message, error) {
	return w.workforce.AssignTaskByCapability(ctx, capability, task, opts...)
}

// RunParallel executes multiple tasks in parallel
func (w *ParallelWorkforce) RunParallel(ctx context.Context, tasks []ParallelWorkforceTask, opts ...model.Option) ([]ParallelWorkforceResult, error) {
	results := make([]ParallelWorkforceResult, len(tasks))

	// Create a channel to receive results
	resultChan := make(chan ParallelWorkforceResult, len(tasks))

	// Execute tasks in parallel
	for _, task := range tasks {
		task := task
		go func() {
			var result ParallelWorkforceResult
			result.WorkerName = task.WorkerName

			var worker WorkerAgent
			var err error

			// Get the appropriate worker
			if task.WorkerName != "" {
				var ok bool
				worker, ok = w.GetWorker(task.WorkerName)
				if !ok {
					err = fmt.Errorf("worker %s not found", task.WorkerName)
				}
			} else if task.Role != "" {
				workers := w.GetWorkersByRole(task.Role)
				if len(workers) == 0 {
					err = fmt.Errorf("no workers found with role %s", task.Role)
				} else {
					worker = workers[0]
					result.WorkerName = worker.Name()
				}
			} else if task.Capability != "" {
				workers := w.GetWorkersByCapability(task.Capability)
				if len(workers) == 0 {
					err = fmt.Errorf("no workers found with capability %s", task.Capability)
				} else {
					worker = workers[0]
					result.WorkerName = worker.Name()
				}
			} else {
				err = fmt.Errorf("no worker selection criteria provided")
			}

			if err != nil {
				result.Error = err.Error()
				resultChan <- result
				return
			}

			// Execute the task
			msg, err := worker.Run(ctx, task.Task, opts...)
			if err != nil {
				result.Error = err.Error()
			} else {
				result.Result = msg
			}

			resultChan <- result
		}()
	}

	// Collect results
	for i := 0; i < len(tasks); i++ {
		results[i] = <-resultChan
	}

	return results, nil
}
