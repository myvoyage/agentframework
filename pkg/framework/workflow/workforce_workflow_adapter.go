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

package workflow

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"

	"AgentFramework/agent/errors"
)

// WorkforceWorkflowAdapter adapts a Workforce to the Workflow interface
// This allows Workforce instances to be used as nodes in a DAGWorkflow
type WorkforceWorkflowAdapter struct {
	name       string
	workforce  Workforce
	role       WorkerRole
	capability string
}

// NewWorkforceWorkflowAdapter creates a new WorkforceWorkflowAdapter instance
func NewWorkforceWorkflowAdapter(name string, workforce Workforce, role WorkerRole, capability string) *WorkforceWorkflowAdapter {
	return &WorkforceWorkflowAdapter{
		name:       name,
		workforce:  workforce,
		role:       role,
		capability: capability,
	}
}

// Name returns the name of the workflow adapter
func (a *WorkforceWorkflowAdapter) Name() string {
	return a.name
}

// Run executes the workforce as a workflow
func (a *WorkforceWorkflowAdapter) Run(ctx context.Context, input string, opts ...model.Option) (*schema.Message, error) {
	// Execute the task using the appropriate workforce method
	if a.role != "" {
		return a.workforce.AssignTaskByRole(ctx, a.role, input, opts...)
	} else if a.capability != "" {
		return a.workforce.AssignTaskByCapability(ctx, a.capability, input, opts...)
	} else {
		// If no role or capability is specified, use the first available worker
		workers := a.workforce.ListWorkers()
		if len(workers) == 0 {
			return nil, errors.New(errors.ErrCodeNotFound, "no workers available in workforce")
		}
		return a.workforce.AssignTask(ctx, workers[0].Name(), input, opts...)
	}
}

// Resume implements the Workflow interface but is not supported for WorkforceWorkflowAdapter
func (a *WorkforceWorkflowAdapter) Resume(ctx context.Context, runID string, input string, opts ...model.Option) (*schema.Message, error) {
	return nil, fmt.Errorf("resume not supported for workforce workflow adapter")
}

// GraphWorkforceWorkflowAdapter adapts a GraphWorkforce to the Workflow interface
type GraphWorkforceWorkflowAdapter struct {
	name           string
	graphWorkforce *GraphWorkforce
}

// NewGraphWorkforceWorkflowAdapter creates a new GraphWorkforceWorkflowAdapter instance
func NewGraphWorkforceWorkflowAdapter(name string, graphWorkforce *GraphWorkforce) *GraphWorkforceWorkflowAdapter {
	return &GraphWorkforceWorkflowAdapter{
		name:           name,
		graphWorkforce: graphWorkforce,
	}
}

// Name returns the name of the workflow adapter
func (a *GraphWorkforceWorkflowAdapter) Name() string {
	return a.name
}

// Run executes the graph workforce as a workflow
func (a *GraphWorkforceWorkflowAdapter) Run(ctx context.Context, input string, opts ...model.Option) (*schema.Message, error) {
	return a.graphWorkforce.Run(ctx, input, opts...)
}

// Resume implements the Workflow interface but is not supported for GraphWorkforceWorkflowAdapter
func (a *GraphWorkforceWorkflowAdapter) Resume(ctx context.Context, runID string, input string, opts ...model.Option) (*schema.Message, error) {
	return nil, fmt.Errorf("resume not supported for graph workforce workflow adapter")
}

// TaskCoordinatorWorkflowAdapter adapts a TaskCoordinator to the Workflow interface
type TaskCoordinatorWorkflowAdapter struct {
	name        string
	coordinator *TaskCoordinator
}

// NewTaskCoordinatorWorkflowAdapter creates a new TaskCoordinatorWorkflowAdapter instance
func NewTaskCoordinatorWorkflowAdapter(name string, coordinator *TaskCoordinator) *TaskCoordinatorWorkflowAdapter {
	return &TaskCoordinatorWorkflowAdapter{
		name:        name,
		coordinator: coordinator,
	}
}

// Name returns the name of the workflow adapter
func (a *TaskCoordinatorWorkflowAdapter) Name() string {
	return a.name
}

// Run executes the task coordinator as a workflow
func (a *TaskCoordinatorWorkflowAdapter) Run(ctx context.Context, input string, opts ...model.Option) (*schema.Message, error) {
	return a.coordinator.Coordinate(ctx, input, opts...)
}

// Resume implements the Workflow interface but is not supported for TaskCoordinatorWorkflowAdapter
func (a *TaskCoordinatorWorkflowAdapter) Resume(ctx context.Context, runID string, input string, opts ...model.Option) (*schema.Message, error) {
	return nil, fmt.Errorf("resume not supported for task coordinator workflow adapter")
}

// WorkforceDAGWorkflow is a DAGWorkflow that supports Workforce nodes
type WorkforceDAGWorkflow struct {
	*DAGWorkflow
}

// NewWorkforceDAGWorkflow creates a new WorkforceDAGWorkflow instance
func NewWorkforceDAGWorkflow(name string) *WorkforceDAGWorkflow {
	return &WorkforceDAGWorkflow{
		DAGWorkflow: NewDAGWorkflow(name),
	}
}

// AddWorkforceNode adds a workforce node to the DAG
func (w *WorkforceDAGWorkflow) AddWorkforceNode(nodeID string, workforce Workforce, role WorkerRole, capability string) {
	adapter := NewWorkforceWorkflowAdapter(nodeID, workforce, role, capability)
	w.AddNode(nodeID, adapter)
}

// AddGraphWorkforceNode adds a graph workforce node to the DAG
func (w *WorkforceDAGWorkflow) AddGraphWorkforceNode(nodeID string, graphWorkforce *GraphWorkforce) {
	adapter := NewGraphWorkforceWorkflowAdapter(nodeID, graphWorkforce)
	w.AddNode(nodeID, adapter)
}

// AddTaskCoordinatorNode adds a task coordinator node to the DAG
func (w *WorkforceDAGWorkflow) AddTaskCoordinatorNode(nodeID string, coordinator *TaskCoordinator) {
	adapter := NewTaskCoordinatorWorkflowAdapter(nodeID, coordinator)
	w.AddNode(nodeID, adapter)
}
