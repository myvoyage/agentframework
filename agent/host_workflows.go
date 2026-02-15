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

	"github.com/cloudwego/eino/components/tool"
)

// buildWorkflows creates workflows from configuration
func (h *Host) buildWorkflows(ctx context.Context) error {
	var routingSpecs []WorkflowSpec
	var planningSpecs []WorkflowSpec

	for _, spec := range h.cfg.Workflows {
		if spec.Name == "" {
			return errors.New("workflow name is required")
		}

		var wf Workflow
		kind := strings.ToLower(spec.Kind)
		switch kind {
		case "routing":
			routingSpecs = append(routingSpecs, spec)
			continue
		case "planning":
			planningSpecs = append(planningSpecs, spec)
			continue
		case "sequential":
			// Implement basic sequential using DAG or specific struct
			if len(spec.Steps) > 0 {
				var agents []Agent
				for _, step := range spec.Steps {
					ag, ok := h.agents[step]
					if !ok {
						return fmt.Errorf("sequential workflow %q refers to unknown agent %q", spec.Name, step)
					}
					agents = append(agents, ag)
				}
				wf = NewSequentialWorkflow(spec.Name, agents...)
			}
		case "dag":
			dag := NewDAGWorkflow(spec.Name)
			// Add nodes
			for _, node := range spec.Nodes {
				// Determine node kind (default to agent if not specified)
				nodeKind := node.Kind
				if nodeKind == "" {
					nodeKind = "agent"
				}

				switch nodeKind {
				case "agent":
					// Existing logic: reference to defined agent
					ag, ok := h.agents[node.AgentName]
					if !ok {
						return fmt.Errorf("dag workflow %q node %q refers to unknown agent %q", spec.Name, node.ID, node.AgentName)
					}
					dag.AddNode(node.ID, ag)
				case "inline":
					// New logic: inline agent definition
					inlineName := node.InlineName
					if inlineName == "" {
						inlineName = node.ID // Use node ID as default inline agent name
					}

					inlineKind := node.InlineKind
					if inlineKind == "" {
						inlineKind = "chat" // Default to chat agent
					}

					inlineModel := node.InlineModel
					if inlineModel == "" {
						inlineModel = h.cfg.DefaultModel
					}

					// Create chat model for the inline agent
					var chatModel ChatModel
					var err error
					if strings.ToLower(inlineKind) != "human" {
						chatModel, err = h.modelFactory(ctx, inlineModel)
						if err != nil {
							return fmt.Errorf("create model %q for inline agent %q failed: %w", inlineModel, inlineName, err)
						}
					}

					// Create the inline agent based on kind
					var ag Agent
					switch strings.ToLower(inlineKind) {
					case "chat":
						// For now, inline agents don't support tools (this can be enhanced later)
						ca, err := NewChatAgent(ctx, ChatAgentConfig{
							Name:         inlineName,
							Instructions: node.InlineInstructions,
							Model:        chatModel,
							Tools:        []tool.BaseTool{}, // Empty tools for now
						})
						if err != nil {
							return fmt.Errorf("create inline ChatAgent %q failed: %w", inlineName, err)
						}
						ag = ca
					case "react":
						// For now, inline agents don't support tools (this can be enhanced later)
						rc, err := newReActFromModel(ctx, inlineName, chatModel, nil, []string{}) // Empty tools for now
						if err != nil {
							return fmt.Errorf("create inline ReActAgent %q failed: %w", inlineName, err)
						}
						ag = rc
					case "human":
						ag = NewHumanNode(inlineName, node.InlineInstructions)
					default:
						return fmt.Errorf("unknown inline agent kind %q for node %q", inlineKind, node.ID)
					}

					// Add the inline agent to the DAG
					dag.AddNode(node.ID, ag)
				default:
					return fmt.Errorf("dag workflow %q node %q has unknown kind %q", spec.Name, node.ID, nodeKind)
				}
			}
			// Add edges (simple map)
			for from, to := range spec.Edges {
				dag.AddEdge(from, to)
			}
			// Add edges (list map)
			for from, toList := range spec.EdgesList {
				for _, to := range toList {
					dag.AddEdge(from, to)
				}
			}
			wf = dag
		case "aggregating_parallel":
			if len(spec.Agents) == 0 {
				return fmt.Errorf("aggregating workflow %q has no agents", spec.Name)
			}
			if spec.Aggregator == "" {
				return fmt.Errorf("aggregating workflow %q has no aggregator", spec.Name)
			}
			agg, ok := h.agents[spec.Aggregator]
			if !ok {
				return fmt.Errorf("workflow %q refers to unknown aggregator %q", spec.Name, spec.Aggregator)
			}
			var agents []Agent
			for _, name := range spec.Agents {
				ag, ok := h.agents[name]
				if !ok {
					return fmt.Errorf("workflow %q refers to unknown agent %q", spec.Name, name)
				}
				agents = append(agents, ag)
			}
			wf = NewAggregatingParallelWorkflow(spec.Name, agg, agents...)
		default:
			return fmt.Errorf("unknown workflow kind %q for workflow %q", spec.Kind, spec.Name)
		}

		h.workflows[spec.Name] = wf
	}

	for _, spec := range routingSpecs {
		if len(spec.Routes) == 0 {
			return fmt.Errorf("routing workflow %q has no routes", spec.Name)
		}

		modelName := spec.Model
		if modelName == "" {
			modelName = h.cfg.DefaultModel
		}
		if modelName == "" {
			return fmt.Errorf("routing workflow %q has no model and defaultModel is empty", spec.Name)
		}

		m, err := h.modelFactory(ctx, modelName)
		if err != nil {
			return fmt.Errorf("create model %q for routing workflow %q failed: %w", modelName, spec.Name, err)
		}

		candidates := make(map[string]Workflow)
		for key, wfName := range spec.Routes {
			wf, ok := h.workflows[wfName]
			if !ok {
				return fmt.Errorf("routing workflow %q refers to unknown workflow %q for route %q", spec.Name, wfName, key)
			}
			candidates[key] = wf
		}

		rwf := NewRoutingWorkflow(spec.Name, m, candidates)
		h.workflows[spec.Name] = rwf
	}

	for _, spec := range planningSpecs {
		if len(spec.Routes) == 0 {
			return fmt.Errorf("planning workflow %q has no routes", spec.Name)
		}

		modelName := spec.Model
		if modelName == "" {
			modelName = h.cfg.DefaultModel
		}
		if modelName == "" {
			return fmt.Errorf("planning workflow %q has no model and defaultModel is empty", spec.Name)
		}

		m, err := h.modelFactory(ctx, modelName)
		if err != nil {
			return fmt.Errorf("create model %q for planning workflow %q failed: %w", modelName, spec.Name, err)
		}

		candidates := make(map[string]Workflow)
		for key, wfName := range spec.Routes {
			wf, ok := h.workflows[wfName]
			if !ok {
				return fmt.Errorf("planning workflow %q refers to unknown workflow %q for route %q", spec.Name, wfName, key)
			}
			candidates[key] = wf
		}

		pwf := NewPlannerWorkflow(spec.Name, m, candidates)
		h.workflows[spec.Name] = pwf
	}

	return nil
}
