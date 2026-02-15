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
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
)

// SequentialWorkflow represents a sequential workflow that executes agents in order.
type SequentialWorkflow struct {
	name   string
	agents []Agent
	store  CheckpointStore
}

// NewSequentialWorkflow creates a new SequentialWorkflow instance.
func NewSequentialWorkflow(name string, agents ...Agent) *SequentialWorkflow {
	return &SequentialWorkflow{
		name:   name,
		agents: agents,
		store:  NewMemoryCheckpointStore(), // Default to memory store
	}
}

func (w *SequentialWorkflow) Name() string {
	return w.name
}

func (w *SequentialWorkflow) WithCheckpointStore(store CheckpointStore) Workflow {
	w.store = store
	return w
}

func (w *SequentialWorkflow) Run(ctx context.Context, input string, opts ...model.Option) (*schema.Message, error) {
	// Create a new run ID for this execution
	runID := uuid.NewString()

	// Create initial checkpoint
	cp := &Checkpoint{
		RunID:        runID,
		WorkflowName: w.name,
		Status:       StatusRunning,
		Input:        input,
		Progress:     0.0,
		State:        &WorkflowState{NodeStates: make(map[string]string)},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Save initial checkpoint
	if err := w.store.Save(ctx, cp); err != nil {
		return nil, err
	}

	var (
		text string = input
		resp *schema.Message
		err  error
	)

	for i, ag := range w.agents {
		// Update checkpoint with current progress
		cp.Progress = float64(i) / float64(len(w.agents))
		cp.UpdatedAt = time.Now()
		cp.State.NodeStates[fmt.Sprintf("step_%d", i)] = text
		if err := w.store.Save(ctx, cp); err != nil {
			return nil, err
		}

		resp, err = ag.Run(ctx, text, opts...)
		if err != nil {
			// Update checkpoint with error status
			cp.Status = StatusFailed
			cp.Error = err.Error()
			cp.UpdatedAt = time.Now()
			_ = w.store.Save(ctx, cp)
			return nil, err
		}
		text = resp.Content
	}

	// Update checkpoint with completion status
	cp.Status = StatusCompleted
	cp.Progress = 1.0
	cp.Output = resp.Content
	cp.UpdatedAt = time.Now()
	if err := w.store.Save(ctx, cp); err != nil {
		return nil, err
	}

	return resp, nil
}

func (w *SequentialWorkflow) Resume(ctx context.Context, runID string, input string, opts ...model.Option) (*schema.Message, error) {
	// Load checkpoint
	cp, err := w.store.Load(ctx, runID)
	if err != nil {
		return nil, err
	}

	// Check if workflow is already completed
	if cp.Status == StatusCompleted {
		return &schema.Message{Content: cp.Output}, nil
	}

	// Update checkpoint status to running
	cp.Status = StatusRunning
	cp.UpdatedAt = time.Now()
	if err := w.store.Save(ctx, cp); err != nil {
		return nil, err
	}

	// Determine which step to resume from
	step := 0
	if cp.State != nil && len(cp.State.NodeStates) > 0 {
		// Find the last completed step
		step = len(cp.State.NodeStates)
	}

	var (
		text string = input
		resp *schema.Message
	)

	// If we're resuming from the beginning, use the original input
	if step == 0 {
		text = cp.Input
	} else {
		// Otherwise, use the output from the last completed step
		text = cp.State.NodeStates[fmt.Sprintf("step_%d", step-1)]
	}

	// Continue execution from the current step
	for i := step; i < len(w.agents); i++ {
		// Update checkpoint with current progress
		cp.Progress = float64(i) / float64(len(w.agents))
		cp.UpdatedAt = time.Now()
		cp.State.NodeStates[fmt.Sprintf("step_%d", i)] = text
		if err := w.store.Save(ctx, cp); err != nil {
			return nil, err
		}

		resp, err = w.agents[i].Run(ctx, text, opts...)
		if err != nil {
			// Update checkpoint with error status
			cp.Status = StatusFailed
			cp.Error = err.Error()
			cp.UpdatedAt = time.Now()
			_ = w.store.Save(ctx, cp)
			return nil, err
		}
		text = resp.Content
	}

	// Update checkpoint with completion status
	cp.Status = StatusCompleted
	cp.Progress = 1.0
	cp.Output = resp.Content
	cp.UpdatedAt = time.Now()
	if err := w.store.Save(ctx, cp); err != nil {
		return nil, err
	}

	return resp, nil
}
