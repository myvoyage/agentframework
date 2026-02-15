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
	"errors"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/flow/agent/react"
)

// buildAgents creates agents from configuration
func (h *Host) buildAgents(ctx context.Context, tools map[string]tool.BaseTool) error {
	for _, spec := range h.configMgr.GetHostConfig().Agents {
		if spec.Name == "" {
			return errors.New("agent name is required")
		}

		modelName := spec.Model
		if modelName == "" {
			modelName = h.configMgr.GetHostConfig().DefaultModel
		}

		// For HumanNode, we don't need a model.
		if strings.ToLower(spec.Kind) != "human" && modelName == "" {
			return fmt.Errorf("agent %q has no model and defaultModel is empty", spec.Name)
		}

		var chatModel ChatModel
		if strings.ToLower(spec.Kind) != "human" {
			var err error
			chatModel, err = h.modelFactory(ctx, modelName)
			if err != nil {
				return fmt.Errorf("create model %q for agent %q failed: %w", modelName, spec.Name, err)
			}
		}

		var ag Agent

		switch strings.ToLower(spec.Kind) {
		case "", "chat":
			var toolList []tool.BaseTool
			for _, name := range spec.Tools {
				t, ok := tools[name]
				if !ok {
					return fmt.Errorf("agent %q refers to unknown tool %q", spec.Name, name)
				}
				toolList = append(toolList, t)
			}
			ca, err := NewChatAgent(ctx, ChatAgentConfig{
				Name:         spec.Name,
				Instructions: spec.Instructions,
				Model:        chatModel,
				Tools:        toolList,
				MemoryOpts: MemoryOptions{
					MaxMessages:    spec.MaxMessages,
					MaxMessageSize: spec.MaxMessageSize,
					TrimRatio:      spec.TrimRatio,
					EnableTrimming: spec.EnableTrimming,
				},
			})
			if err != nil {
				return fmt.Errorf("create ChatAgent %q failed: %w", spec.Name, err)
			}
			ag = ca
		case "react":
			rc, err := newReActFromModel(ctx, spec.Name, chatModel, tools, spec.Tools)
			if err != nil {
				return fmt.Errorf("create ReActAgent %q failed: %w", spec.Name, err)
			}
			// Set memory management options
			rc.SetMemoryOptions(MemoryOptions{
				MaxMessages:    spec.MaxMessages,
				MaxMessageSize: spec.MaxMessageSize,
				TrimRatio:      spec.TrimRatio,
				EnableTrimming: spec.EnableTrimming,
			})
			ag = rc
		case "human":
			ag = NewHumanNode(spec.Name, "Waiting for human approval")
		default:
			return fmt.Errorf("unknown agent kind %q for agent %q", spec.Kind, spec.Name)
		}

		if len(spec.Middlewares) > 0 {
			var mws []AgentMiddleware
			for _, key := range spec.Middlewares {
				mw, ok := h.middlewares[key]
				if !ok {
					return fmt.Errorf("unknown middleware %q for agent %q", key, spec.Name)
				}
				mws = append(mws, mw)
			}
			ag = WrapAgent(ag, mws...)
		}

		h.agents[spec.Name] = ag
	}

	return nil
}

// newReActFromModel creates a ReActAgent from a ChatModel
func newReActFromModel(ctx context.Context, name string, m ChatModel, allTools map[string]tool.BaseTool, toolNames []string) (*ReActAgent, error) {
	tc, ok := m.(model.ToolCallingChatModel)
	if !ok {
		return nil, fmt.Errorf("underlying model for agent %q does not support tool calling", name)
	}

	inner, err := react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel: tc,
	})
	if err != nil {
		return nil, err
	}

	if len(toolNames) == 0 {
		return NewReActAgent(name, inner), nil
	}

	var selected []tool.BaseTool
	for _, tname := range toolNames {
		t, ok := allTools[tname]
		if !ok {
			return nil, fmt.Errorf("react agent %q refers to unknown tool %q", name, tname)
		}
		selected = append(selected, t)
	}

	opts, err := react.WithTools(ctx, selected...)
	if err != nil {
		return nil, err
	}

	return NewReActAgent(name, inner, opts...), nil
}
