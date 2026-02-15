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
	"strings"
	"testing"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

// MockNode is a simple workflow node for testing
type MockNode struct {
	name   string
	suffix string
}

func (m *MockNode) Name() string {
	return m.name
}

func (m *MockNode) Run(ctx context.Context, input string, opts ...model.Option) (*schema.Message, error) {
	// Append suffix to input
	return &schema.Message{
		Role:    schema.Assistant,
		Content: input + m.suffix,
	}, nil
}

// Resume implements the Workflow interface for MockNode
func (m *MockNode) Resume(ctx context.Context, runID string, input string, opts ...model.Option) (*schema.Message, error) {
	// For testing purposes, we can just delegate to Run
	// In real implementations, this would handle resuming from a checkpoint
	return m.Run(ctx, input, opts...)
}

func TestGraphWorkflow_CritiqueRefine(t *testing.T) {
	// Scenario:
	// Start -> Writer (adds " wrote") -> Reviewer (adds " reviewed")
	// Reviewer -> Check: if length < 20, go back to Writer; else End.

	wf := NewGraphWorkflow("critique_refine_test")

	writer := &MockNode{name: "writer", suffix: " wrote"}
	reviewer := &MockNode{name: "reviewer", suffix: " reviewed"}

	wf.AddNode("writer", writer)
	wf.AddNode("reviewer", reviewer)
	wf.SetStartNode("writer")

	// Edge: Writer -> Reviewer
	wf.AddEdge("writer", "reviewer")

	// Conditional Edge: Reviewer -> Writer (loop) or End
	wf.AddConditionalEdge("reviewer", func(ctx context.Context, input string) (string, error) {
		// input here is the output of Reviewer
		fmt.Printf("Reviewer output: %s (len: %d)\n", input, len(input))
		if len(input) < 30 { // Loop until string is long enough
			return "writer", nil
		}
		return NodeEnd, nil
	})

	ctx := context.Background()
	initialInput := "Start"

	resp, err := wf.Run(ctx, initialInput)
	if err != nil {
		t.Fatalf("Workflow run failed: %v", err)
	}

	t.Logf("Final output: %s", resp.Content)

	// Verify logic
	// Iteration 1: "Start" -> "Start wrote" -> "Start wrote reviewed" (len 20) -> Loop to Writer
	// Iteration 2: "Start wrote reviewed" -> "Start wrote reviewed wrote" -> "Start wrote reviewed wrote reviewed" (len 39) -> End

	expectedSuffix := "wrote reviewed wrote reviewed"
	if !strings.HasSuffix(resp.Content, expectedSuffix) {
		t.Errorf("Expected suffix %q, got %q", expectedSuffix, resp.Content)
	}
}

func TestGraphWorkflow_MaxSteps(t *testing.T) {
	wf := NewGraphWorkflow("infinite_loop_test")

	nodeA := &MockNode{name: "A", suffix: "."}
	wf.AddNode("A", nodeA)
	wf.SetStartNode("A")

	// A -> A (Infinite loop)
	wf.AddEdge("A", "A")

	ctx := context.Background()
	_, err := wf.Run(ctx, "Start")

	if err == nil {
		t.Error("Expected error due to max steps exceeded, got nil")
	} else {
		if !strings.Contains(err.Error(), "exceeded max steps") {
			t.Errorf("Expected 'exceeded max steps' error, got: %v", err)
		}
	}
}
