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
	"strings"
	"testing"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

// MockAgent implements Agent interface for testing
type MockAgent struct {
	NameVal string
	RunFunc func(ctx context.Context, input string, opts ...model.Option) (*schema.Message, error)
}

func (m *MockAgent) Name() string {
	return m.NameVal
}

func (m *MockAgent) Run(ctx context.Context, input string, opts ...model.Option) (*schema.Message, error) {
	if m.RunFunc != nil {
		return m.RunFunc(ctx, input, opts...)
	}
	return &schema.Message{Role: schema.Assistant, Content: fmt.Sprintf("Processed by %s: %s", m.NameVal, input)}, nil
}

func TestDAGWorkflow(t *testing.T) {
	// Setup DAG:
	//       Start
	//      /     \
	//     A       B
	//      \     /
	//       End

	dag := NewDAGWorkflow("test_dag")

	// Node A: appends " -> A"
	nodeA := &MockAgent{
		NameVal: "NodeA",
		RunFunc: func(ctx context.Context, input string, opts ...model.Option) (*schema.Message, error) {
			time.Sleep(10 * time.Millisecond) // Simulate work
			return &schema.Message{Content: input + " -> A"}, nil
		},
	}

	// Node B: appends " -> B"
	nodeB := &MockAgent{
		NameVal: "NodeB",
		RunFunc: func(ctx context.Context, input string, opts ...model.Option) (*schema.Message, error) {
			time.Sleep(20 * time.Millisecond) // Simulate work
			return &schema.Message{Content: input + " -> B"}, nil
		},
	}

	// Node End: joins inputs
	nodeEnd := &MockAgent{
		NameVal: "NodeEnd",
		RunFunc: func(ctx context.Context, input string, opts ...model.Option) (*schema.Message, error) {
			// Expect input to be "Start -> A\n\nStart -> B" (order depends on map iteration, so checking contains)
			return &schema.Message{Content: "Final: " + input}, nil
		},
	}

	dag.AddNode("A", nodeA)
	dag.AddNode("B", nodeB)
	dag.AddNode("End", nodeEnd)

	// Edges: A->End, B->End
	// Note: DAGWorkflow implementation in previous turn didn't explicitly model a "Start" node that fans out.
	// It assumes inputs with 0 in-degree get the initial input.
	// So both A and B have 0 in-degree and will receive the initial workflow input.

	dag.AddEdge("A", "End")
	dag.AddEdge("B", "End")

	resp, err := dag.Run(context.Background(), "Start")
	if err != nil {
		t.Fatalf("DAG run failed: %v", err)
	}

	output := resp.Content
	// Validate structure
	if !strings.Contains(output, "Final:") {
		t.Errorf("Expected output to start with 'Final:', got %q", output)
	}
	if !strings.Contains(output, "Start -> A") {
		t.Errorf("Expected output to contain 'Start -> A', got %q", output)
	}
	if !strings.Contains(output, "Start -> B") {
		t.Errorf("Expected output to contain 'Start -> B', got %q", output)
	}

	t.Logf("DAG Output: %s", output)
}

func TestDAGWorkflow_Diamond(t *testing.T) {
	// Diamond shape:
	//      Start (virtual)
	//      /   \
	//     N1   N2
	//      \   /
	//       N3
	//       |
	//       N4

	dag := NewDAGWorkflow("diamond_dag")

	dag.AddNode("N1", &MockAgent{NameVal: "N1", RunFunc: func(ctx context.Context, i string, o ...model.Option) (*schema.Message, error) {
		return &schema.Message{Content: "1"}, nil
	}})
	dag.AddNode("N2", &MockAgent{NameVal: "N2", RunFunc: func(ctx context.Context, i string, o ...model.Option) (*schema.Message, error) {
		return &schema.Message{Content: "2"}, nil
	}})
	dag.AddNode("N3", &MockAgent{NameVal: "N3", RunFunc: func(ctx context.Context, i string, o ...model.Option) (*schema.Message, error) {
		// Expects "1" and "2" joined
		if !strings.Contains(i, "1") || !strings.Contains(i, "2") {
			return nil, fmt.Errorf("N3 received invalid input: %q", i)
		}
		return &schema.Message{Content: "3"}, nil
	}})
	dag.AddNode("N4", &MockAgent{NameVal: "N4", RunFunc: func(ctx context.Context, i string, o ...model.Option) (*schema.Message, error) {
		if i != "3" {
			return nil, fmt.Errorf("N4 expected '3', got %q", i)
		}
		return &schema.Message{Content: "4"}, nil
	}})

	dag.AddEdge("N1", "N3")
	dag.AddEdge("N2", "N3")
	dag.AddEdge("N3", "N4")

	resp, err := dag.Run(context.Background(), "init")
	if err != nil {
		t.Fatalf("Diamond DAG run failed: %v", err)
	}

	if resp.Content != "4" {
		t.Errorf("Expected final output '4', got %q", resp.Content)
	}
}
