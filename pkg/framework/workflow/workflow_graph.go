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
)

// GraphWorkflow implements a cyclic graph workflow.
// It uses a state machine approach where each node returns the next node to execute.
type GraphWorkflow struct {
	name      string
	nodes     map[string]WorkflowInterface
	startNode string
	// edges maps nodeID -> condition function that returns next nodeID
	// If condition is nil, it's a terminal node (or use END)
	edges map[string]EdgeCondition
}

type EdgeCondition func(ctx context.Context, input string) (string, error)

const (
	NodeEnd = "__END__"
)

func NewGraphWorkflow(name string) *GraphWorkflow {
	return &GraphWorkflow{
		name:  name,
		nodes: make(map[string]WorkflowInterface),
		edges: make(map[string]EdgeCondition),
	}
}

func (w *GraphWorkflow) Name() string {
	return w.name
}

func (w *GraphWorkflow) AddNode(id string, wf WorkflowInterface) {
	w.nodes[id] = wf
}

func (w *GraphWorkflow) SetStartNode(id string) {
	w.startNode = id
}

// AddEdge adds a static edge (unconditional transition)
func (w *GraphWorkflow) AddEdge(from, to string) {
	w.edges[from] = func(ctx context.Context, input string) (string, error) {
		return to, nil
	}
}

// AddConditionalEdge adds a dynamic edge based on output
func (w *GraphWorkflow) AddConditionalEdge(from string, condition EdgeCondition) {
	w.edges[from] = condition
}

func (w *GraphWorkflow) Run(ctx context.Context, input string, opts ...model.Option) (*schema.Message, error) {
	// For simplicity, GraphWorkflow is not fully Resumable in this first implementation
	// unless we track the current node in the state.
	// But let's implement the basic run loop.

	if w.startNode == "" {
		return nil, fmt.Errorf("graph workflow %s has no start node", w.name)
	}

	currentNodeID := w.startNode
	currentInput := input
	var lastResp *schema.Message

	// Max steps to prevent infinite loops (default 100)
	maxSteps := 100
	step := 0

	for {
		if currentNodeID == NodeEnd {
			break
		}
		if step >= maxSteps {
			return nil, fmt.Errorf("graph workflow exceeded max steps (%d)", maxSteps)
		}

		node, ok := w.nodes[currentNodeID]
		if !ok {
			return nil, fmt.Errorf("unknown node: %s", currentNodeID)
		}

		// Emit Node Start
		if cb := GetWorkflowCallbacks(ctx); cb != nil {
			cb.OnNodeStart(ctx, currentNodeID, currentInput)
		}

		resp, err := node.Run(ctx, currentInput, opts...)
		if err != nil {
			return nil, err
		}
		lastResp = resp
		currentInput = resp.Content

		// Emit Node End
		if cb := GetWorkflowCallbacks(ctx); cb != nil {
			cb.OnNodeEnd(ctx, currentNodeID, resp.Content)
		}

		// Determine next node
		condition, ok := w.edges[currentNodeID]
		if !ok {
			// No outgoing edge -> End
			currentNodeID = NodeEnd
		} else {
			nextNode, err := condition(ctx, currentInput)
			if err != nil {
				return nil, err
			}
			currentNodeID = nextNode
		}
		step++
	}

	if lastResp == nil {
		return &schema.Message{Role: schema.Assistant, Content: ""}, nil
	}
	return lastResp, nil
}

func (w *GraphWorkflow) Resume(ctx context.Context, runID string, input string, opts ...model.Option) (*schema.Message, error) {
	return nil, fmt.Errorf("resume not supported for graph workflow")
}
