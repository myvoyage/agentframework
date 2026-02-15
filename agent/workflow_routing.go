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

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

// RoutingWorkflow represents a workflow that routes requests to different workflows based on a model's decision.
type RoutingWorkflow struct {
	name       string
	model      ChatModel
	candidates map[string]Workflow
}

// NewRoutingWorkflow creates a new RoutingWorkflow instance.
func NewRoutingWorkflow(name string, m ChatModel, candidates map[string]Workflow) *RoutingWorkflow {
	return &RoutingWorkflow{
		name:       name,
		model:      m,
		candidates: candidates,
	}
}

func (w *RoutingWorkflow) Name() string {
	return w.name
}

func (w *RoutingWorkflow) Run(ctx context.Context, input string, opts ...model.Option) (*schema.Message, error) {
	if len(w.candidates) == 0 {
		return nil, fmt.Errorf("no candidates configured for routing workflow %s", w.name)
	}

	var keys []string
	for k := range w.candidates {
		keys = append(keys, k)
	}

	system := &schema.Message{
		Role: schema.System,
		Content: "You are a router. Choose one route key from the list: " +
			strings.Join(keys, ", ") +
			". Only answer with the key itself.",
	}

	user := &schema.Message{
		Role:    schema.User,
		Content: input,
	}

	resp, err := w.model.Generate(ctx, []*schema.Message{system, user}, opts...)
	if err != nil {
		return nil, err
	}

	choice := strings.TrimSpace(resp.Content)
	if choice == "" {
		return nil, fmt.Errorf("router did not return a choice")
	}

	target, ok := w.candidates[choice]
	if !ok {
		return nil, fmt.Errorf("router returned unknown route %q", choice)
	}

	return target.Run(ctx, input, opts...)
}

func (w *RoutingWorkflow) Resume(ctx context.Context, runID string, input string, opts ...model.Option) (*schema.Message, error) {
	return nil, fmt.Errorf("resume not supported for routing workflow")
}
