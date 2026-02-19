// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later


package pipelineengine

import (
	"context"
	"fmt"
	yaml "gopkg.in/yaml.v3"
	"time"

)

type PipelineEngine struct {
	registry ToolRegistry
}

func NewPipelineEngine(reg ToolRegistry) *PipelineEngine {
	return &PipelineEngine{registry: reg}
}

func (pe *PipelineEngine) LoadPipeline(data []byte) (*PipelineSpec, error) {
	var p PipelineSpec
	if err := yaml.Unmarshal(data, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

func (pe *PipelineEngine) RunPipeline(ctx context.Context, p *PipelineSpec) (*ExecutionContext, error) {
	if p == nil {
		return nil, fmt.Errorf("nil pipeline")
	}
	exec := &ExecutionContext{
		PipelineID: p.Id,
		StepIndex:  -1,
		Outputs:    map[string]interface{}{},
		Memory:     map[string]interface{}{},
		Logs:       []string{},
		CreatedAt:  time.Now(),
	}

	for i := 0; i < len(p.Steps); i++ {
		step := p.Steps[i]
		exec.StepIndex = i
		switch step.Type {
		case "task":
			// If there is a spec for the tool, validate inputs
			if spec, err := pe.registry.GetToolSpec(step.Tool); err == nil {
				if len(spec.InputsSchema) > 0 {
					if err := ValidateInputs(spec, step.Params); err != nil {
						return exec, err
					}
				}
			}
			tool, err := pe.registry.GetTool(step.Tool, "")
			if err != nil {
				return exec, err
			}
			out, err := tool.Execute(ctx, step.Params)
			if err != nil {
				return exec, err
			}
			// validate outputs if schema defined
			if spec, err := pe.registry.GetToolSpec(step.Tool); err == nil {
				if len(spec.OutputsSchema) > 0 {
					if err := ValidateOutputs(spec, out); err != nil {
						return exec, err
					}
				}
			}
			exec.Outputs[step.Id] = out
			exec.Logs = append(exec.Logs, fmt.Sprintf("step %s -> tool %s produced %v", step.Id, step.Tool, out))
		case "branch":
			if step.Condition == "true" {
				nextID := step.Next
				found := -1
				for idx, s := range p.Steps {
					if s.Id == nextID {
						found = idx
						break
					}
				}
				if found >= 0 {
					i = found - 1
				}
			} else {
				// terminate pipeline for MVP branch false or unspecified
				return exec, nil
			}
		case "loop":
			if step.Loop != nil {
				target := step.Loop.Do
				targetIdx := -1
				for idx, s := range p.Steps {
					if s.Id == target {
						targetIdx = idx
						break
					}
				}
				if targetIdx >= 0 {
					i = targetIdx - 1
				}
			}
		case "end":
			return exec, nil
		default:
			exec.Logs = append(exec.Logs, fmt.Sprintf("unsupported step type: %s", step.Type))
		}
	}
	return exec, nil
}

// InvokeTool exposes a bridge-friendly tool invocation
func (pe *PipelineEngine) InvokeTool(ctx context.Context, name string, inputs map[string]interface{}, contextData map[string]interface{}) (map[string]interface{}, error) {
	tool, err := pe.registry.GetTool(name, "")
	if err != nil {
		return nil, err
	}
	return tool.Execute(ctx, inputs)
}

// RunPipelineBridge is a bridge-friendly entry that currently calls RunPipeline directly
func (pe *PipelineEngine) RunPipelineBridge(ctx context.Context, p *PipelineSpec) (*ExecutionContext, error) {
	return pe.RunPipeline(ctx, p)
}

func yamlUnmarshal(data []byte, v interface{}) error {
	return yaml.Unmarshal(data, v)
}
