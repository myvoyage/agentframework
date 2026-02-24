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
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

type planStep struct {
	Route     string `json:"route"`
	Reasoning string `json:"reasoning,omitempty"`
	Input     string `json:"input,omitempty"`
}

// PlannerWorkflow represents a workflow that creates and executes a plan based on a model's decision.
type PlannerWorkflow struct {
	name       string
	model      ChatModel
	candidates map[string]WorkflowInterface
}

// NewPlannerWorkflow creates a new PlannerWorkflow instance.
func NewPlannerWorkflow(name string, m ChatModel, candidates map[string]WorkflowInterface) *PlannerWorkflow {
	return &PlannerWorkflow{
		name:       name,
		model:      m,
		candidates: candidates,
	}
}

func (w *PlannerWorkflow) Name() string {
	return w.name
}

func (w *PlannerWorkflow) Run(ctx context.Context, input string, opts ...model.Option) (*schema.Message, error) {
	if len(w.candidates) == 0 {
		return nil, fmt.Errorf("no candidates configured for planner workflow %s", w.name)
	}

	var keys []string
	for k := range w.candidates {
		keys = append(keys, k)
	}

	system := &schema.Message{
		Role: schema.System,
		Content: "You are a planner. Based on the user request and the available route keys: " +
			strings.Join(keys, ", ") +
			". Generate a JSON array plan.\n" +
			"Each element must be an object with fields:\n" +
			"- \"route\": one of the route keys (required)\n" +
			"- \"reasoning\": why you chose this step (optional)\n" +
			"- \"input\": specific input for this step (optional, if omitted, the previous step's output will be used)\n" +
			"Example: [{\"route\":\"" + keys[0] + "\", \"reasoning\":\"first step\", \"input\":\"custom input\"}].\n" +
			"Do not include any other text.",
	}

	user := &schema.Message{
		Role:    schema.User,
		Content: input,
	}

	resp, err := w.model.Generate(ctx, []*schema.Message{system, user}, opts...)
	if err != nil {
		return nil, err
	}

	content := strings.TrimSpace(resp.Content)
	// Remove markdown code blocks if present
	if strings.HasPrefix(content, "```json") {
		content = strings.TrimPrefix(content, "```json")
		content = strings.TrimSuffix(content, "```")
	} else if strings.HasPrefix(content, "```") {
		content = strings.TrimPrefix(content, "```")
		content = strings.TrimSuffix(content, "```")
	}
	content = strings.TrimSpace(content)

	var steps []planStep
	if err := json.Unmarshal([]byte(content), &steps); err != nil {
		return nil, fmt.Errorf("planner failed to produce valid JSON plan: %w", err)
	}
	if len(steps) == 0 {
		return nil, fmt.Errorf("planner returned empty plan")
	}

	var (
		currentInput = input
		lastResp     *schema.Message
	)

	for i, step := range steps {
		route := strings.TrimSpace(step.Route)
		if route == "" {
			return nil, fmt.Errorf("planner produced step with empty route")
		}
		target, ok := w.candidates[route]
		if !ok {
			return nil, fmt.Errorf("planner refers to unknown route %q", route)
		}

		// Log reasoning
		if step.Reasoning != "" {
			// Ideally use a logger, but for now we print to stdout or rely on middleware if we had one attached to planner
			// For this implementation, we'll format it into the "system" context if we could, but here we just execute.
			// Let's rely on the structured logger middleware to capture this if we add attributes later.
		}

		// Determine input for this step
		stepInput := currentInput
		if step.Input != "" {
			stepInput = step.Input
		}

		// If it's the first step and no input override, use original input.
		// If it's subsequent step and no input override, use previous output (already in currentInput).

		resp, err := target.Run(ctx, stepInput, opts...)
		if err != nil {
			return nil, fmt.Errorf("step %d (route %q) failed: %w", i, route, err)
		}
		lastResp = resp
		currentInput = resp.Content
	}

	if lastResp == nil {
		return &schema.Message{Role: schema.Assistant, Content: ""}, nil
	}

	return lastResp, nil
}

func (w *PlannerWorkflow) Resume(ctx context.Context, runID string, input string, opts ...model.Option) (*schema.Message, error) {
	return nil, fmt.Errorf("resume not supported for planner workflow")
}
