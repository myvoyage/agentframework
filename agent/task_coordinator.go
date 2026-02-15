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

// TaskSplitter is responsible for splitting a complex task into smaller subtasks
type TaskSplitter interface {
	// Split splits a complex task into smaller subtasks
	Split(ctx context.Context, task string) ([]ParallelWorkforceTask, error)
}

// ResultAggregator is responsible for aggregating results from multiple subtasks
type ResultAggregator interface {
	// Aggregate aggregates results from multiple subtasks into a single result
	Aggregate(ctx context.Context, results []ParallelWorkforceResult) (*schema.Message, error)
}

// TaskCoordinator coordinates the task splitting, execution, and result aggregation
type TaskCoordinator struct {
	workforce        Workforce
	taskSplitter     TaskSplitter
	resultAggregator ResultAggregator
	model            ChatModel
}

// NewTaskCoordinator creates a new TaskCoordinator instance
func NewTaskCoordinator(workforce Workforce, splitter TaskSplitter, aggregator ResultAggregator, model ChatModel) *TaskCoordinator {
	return &TaskCoordinator{
		workforce:        workforce,
		taskSplitter:     splitter,
		resultAggregator: aggregator,
		model:            model,
	}
}

// Coordinate executes a complex task by splitting it into subtasks, executing them in parallel, and aggregating the results
func (c *TaskCoordinator) Coordinate(ctx context.Context, task string, opts ...model.Option) (*schema.Message, error) {
	// 1. Split the task into subtasks
	subtasks, err := c.taskSplitter.Split(ctx, task)
	if err != nil {
		return nil, err
	}

	// 2. Execute subtasks in parallel
	var results []ParallelWorkforceResult
	if pw, ok := c.workforce.(*ParallelWorkforce); ok {
		results, err = pw.RunParallel(ctx, subtasks, opts...)
	} else {
		// Fallback to sequential execution if parallel execution is not supported
		results = make([]ParallelWorkforceResult, len(subtasks))
		for i, subtask := range subtasks {
			var result ParallelWorkforceResult
			var msg *schema.Message

			if subtask.WorkerName != "" {
				msg, err = c.workforce.AssignTask(ctx, subtask.WorkerName, subtask.Task, opts...)
			} else if subtask.Role != "" {
				msg, err = c.workforce.AssignTaskByRole(ctx, subtask.Role, subtask.Task, opts...)
			} else if subtask.Capability != "" {
				msg, err = c.workforce.AssignTaskByCapability(ctx, subtask.Capability, subtask.Task, opts...)
			} else {
				err = fmt.Errorf("no worker selection criteria provided")
			}

			if err != nil {
				result.Error = err.Error()
			} else {
				result.Result = msg
			}
			results[i] = result
		}
	}

	if err != nil {
		return nil, err
	}

	// 3. Aggregate the results
	return c.resultAggregator.Aggregate(ctx, results)
}

// SimpleTaskSplitter is a simple implementation of TaskSplitter that uses an LLM to split tasks
type SimpleTaskSplitter struct {
	model ChatModel
}

// NewSimpleTaskSplitter creates a new SimpleTaskSplitter instance
func NewSimpleTaskSplitter(model ChatModel) *SimpleTaskSplitter {
	return &SimpleTaskSplitter{
		model: model,
	}
}

// Split splits a complex task into smaller subtasks using an LLM
func (s *SimpleTaskSplitter) Split(ctx context.Context, task string) ([]ParallelWorkforceTask, error) {
	prompt := fmt.Sprintf(`
You are a task splitter. Split the following complex task into 2-5 smaller, independent subtasks that can be executed in parallel. Each subtask should be assigned to an appropriate worker role.

Available worker roles:
- developer: write and execute code
- browser: browse the web, search for information
- document: create, edit, and summarize documents
- multi_modal: process images, audio, and video
- researcher: research topics and synthesize information
- writer: write creative content
- reviewer: review and critique content
- analyst: analyze data and create charts

Task: %s

Return the subtasks in JSON format with no additional text. Each subtask should have:
- worker_name: (optional) specific worker name
- role: worker role to use if worker_name is not specified
- capability: (optional) worker capability to use if worker_name and role are not specified
- task: the subtask to execute

Example output:
[
  {"role": "researcher", "task": "Research the history of AI"},
  {"role": "writer", "task": "Write an introduction about AI"}
]
`, task)

	msg, err := s.model.Generate(ctx, []*schema.Message{
		{Role: schema.User, Content: prompt},
	})
	if err != nil {
		return nil, err
	}

	// Parse the JSON response
	// For simplicity, we'll implement a basic parser here
	// In a real implementation, you would use proper JSON parsing
	var subtasks []ParallelWorkforceTask

	// Simple parsing implementation (for demo purposes)
	content := msg.Content
	content = strings.TrimPrefix(content, "[")
	content = strings.TrimSuffix(content, "]")
	content = strings.TrimSpace(content)

	if content == "" {
		return []ParallelWorkforceTask{{Role: RoleResearcher, Task: task}}, nil
	}

	// Split into individual subtasks
	taskStrings := strings.Split(content, "}, {")
	for i, ts := range taskStrings {
		// Fix up the JSON formatting
		if i > 0 {
			ts = "{" + ts
		}
		if i < len(taskStrings)-1 {
			ts = ts + "}"
		}

		// Extract role and task from the string
		var role WorkerRole
		var taskStr string

		// Simple string parsing
		if strings.Contains(ts, "role") {
			rolePart := strings.Split(ts, "role")[1]
			rolePart = strings.Split(rolePart, ",")[0]
			rolePart = strings.TrimSpace(strings.Trim(rolePart, ":\"'"))
			role = WorkerRole(rolePart)
		}

		if strings.Contains(ts, "task") {
			taskPart := strings.Split(ts, "task")[1]
			taskPart = strings.Split(taskPart, ",")[0]
			taskPart = strings.TrimSpace(strings.Trim(taskPart, ":\"'"))
			taskStr = taskPart
		}

		if taskStr != "" {
			subtasks = append(subtasks, ParallelWorkforceTask{
				Role: role,
				Task: taskStr,
			})
		}
	}

	if len(subtasks) == 0 {
		// Fallback: return the original task as a single subtask
		subtasks = append(subtasks, ParallelWorkforceTask{
			Role: RoleResearcher,
			Task: task,
		})
	}

	return subtasks, nil
}

// SimpleResultAggregator is a simple implementation of ResultAggregator that uses an LLM to aggregate results
type SimpleResultAggregator struct {
	model ChatModel
}

// NewSimpleResultAggregator creates a new SimpleResultAggregator instance
func NewSimpleResultAggregator(model ChatModel) *SimpleResultAggregator {
	return &SimpleResultAggregator{
		model: model,
	}
}

// Aggregate aggregates results from multiple subtasks into a single result using an LLM
func (a *SimpleResultAggregator) Aggregate(ctx context.Context, results []ParallelWorkforceResult) (*schema.Message, error) {
	// Format the results for the LLM
	var resultStrings []string
	for i, result := range results {
		if result.Error != "" {
			resultStrings = append(resultStrings, fmt.Sprintf("Subtask %d (Error): %s", i+1, result.Error))
		} else if result.Result != nil {
			resultStrings = append(resultStrings, fmt.Sprintf("Subtask %d (%s): %s", i+1, result.WorkerName, result.Result.Content))
		} else {
			resultStrings = append(resultStrings, fmt.Sprintf("Subtask %d (%s): No result", i+1, result.WorkerName))
		}
	}

	resultsText := strings.Join(resultStrings, "\n")

	prompt := fmt.Sprintf(`
You are a result aggregator. Aggregate the following results from multiple subtasks into a single, coherent result. The aggregated result should be a comprehensive summary that combines all relevant information from the subtasks.

Subtask Results:
%s

Aggregated Result:`, resultsText)

	return a.model.Generate(ctx, []*schema.Message{
		{Role: schema.User, Content: prompt},
	})
}

// ComplexTaskAgent is an agent that can handle complex tasks by splitting them into subtasks
type ComplexTaskAgent struct {
	name         string
	coordinator  *TaskCoordinator
	instructions string
}

// NewComplexTaskAgent creates a new ComplexTaskAgent instance
func NewComplexTaskAgent(name string, coordinator *TaskCoordinator, instructions string) *ComplexTaskAgent {
	return &ComplexTaskAgent{
		name:         name,
		coordinator:  coordinator,
		instructions: instructions,
	}
}

// Name returns the name of the agent
func (a *ComplexTaskAgent) Name() string {
	return a.name
}

// Run executes the complex task agent
func (a *ComplexTaskAgent) Run(ctx context.Context, input string, opts ...model.Option) (*schema.Message, error) {
	// Add system instructions
	fullInput := fmt.Sprintf(`%s

Task: %s`, a.instructions, input)

	// Coordinate the task execution
	return a.coordinator.Coordinate(ctx, fullInput, opts...)
}

// Stream is not supported by ComplexTaskAgent
func (a *ComplexTaskAgent) Stream(ctx context.Context, input string, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	return nil, fmt.Errorf("streaming not supported by ComplexTaskAgent")
}
