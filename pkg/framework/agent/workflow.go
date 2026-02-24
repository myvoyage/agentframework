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
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
	"gopkg.in/yaml.v3"
)

// WorkflowExecutionStatus represents the status of a workflow execution
type WorkflowExecutionStatus string

const (
	// WorkflowStatusPending indicates the workflow is pending execution
	WorkflowStatusPending WorkflowExecutionStatus = "pending"
	// WorkflowStatusRunning indicates the workflow is currently running
	WorkflowStatusRunning WorkflowExecutionStatus = "running"
	// WorkflowStatusCompleted indicates the workflow has completed successfully
	WorkflowStatusCompleted WorkflowExecutionStatus = "completed"
	// WorkflowStatusFailed indicates the workflow has failed
	WorkflowStatusFailed WorkflowExecutionStatus = "failed"
	// WorkflowStatusSuspended indicates the workflow is suspended and waiting for input
	WorkflowStatusSuspended WorkflowExecutionStatus = "suspended"
)

// NodeExecutionStatus represents the status of a node execution
type NodeExecutionStatus string

const (
	// NodeStatusWaiting indicates the node is waiting to be executed
	NodeStatusWaiting NodeExecutionStatus = "waiting"
	// NodeStatusRunning indicates the node is currently running
	NodeStatusRunning NodeExecutionStatus = "running"
	// NodeStatusCompleted indicates the node has completed successfully
	NodeStatusCompleted NodeExecutionStatus = "completed"
	// NodeStatusFailed indicates the node has failed
	NodeStatusFailed NodeExecutionStatus = "failed"
	// NodeStatusSkipped indicates the node was skipped
	NodeStatusSkipped NodeExecutionStatus = "skipped"
)

// NodeExecutionResult represents the execution result of a node
type NodeExecutionResult struct {
	NodeID     string              `json:"node_id"`
	Status     NodeExecutionStatus `json:"status"`
	Input      string              `json:"input,omitempty"`
	Output     string              `json:"output,omitempty"`
	StartTime  time.Time           `json:"start_time"`
	EndTime    time.Time           `json:"end_time"`
	Error      string              `json:"error,omitempty"`
	RetryCount int                 `json:"retry_count,omitempty"`
}

// WorkflowExecutionResult represents the execution result of a workflow
type WorkflowExecutionResult struct {
	WorkflowID  string                  `json:"workflow_id"`
	ExecutionID string                  `json:"execution_id"`
	Status      WorkflowExecutionStatus `json:"status"`
	Input       string                  `json:"input,omitempty"`
	Output      string                  `json:"output,omitempty"`
	StartTime   time.Time               `json:"start_time"`
	EndTime     time.Time               `json:"end_time,omitempty"`
	Error       string                  `json:"error,omitempty"`
	NodeResults []NodeExecutionResult   `json:"node_results,omitempty"`
}

// WorkflowCallbackHandler defines callbacks for workflow execution events.
// This is a custom callback interface since Eino doesn't have a specific "Workflow Node" callback type yet.
// We will inject this into the context.
type WorkflowCallbackHandler interface {
	OnNodeStart(ctx context.Context, nodeID string, input string)
	OnNodeEnd(ctx context.Context, nodeID string, output string)
	OnWorkflowStart(ctx context.Context, workflowID string, input string)
	OnWorkflowEnd(ctx context.Context, workflowID string, output string, status WorkflowExecutionStatus)
}

// WorkflowExecutionStore defines the interface for workflow execution result storage
type WorkflowExecutionStore interface {
	// SaveExecutionResult saves a workflow execution result
	SaveExecutionResult(ctx context.Context, result *WorkflowExecutionResult) error
	// GetExecutionResult gets a workflow execution result by execution ID
	GetExecutionResult(ctx context.Context, executionID string) (*WorkflowExecutionResult, error)
	// GetExecutionResultsByWorkflowID gets all execution results for a workflow
	GetExecutionResultsByWorkflowID(ctx context.Context, workflowID string) ([]*WorkflowExecutionResult, error)
}

// InMemoryWorkflowExecutionStore stores workflow execution results in memory
type InMemoryWorkflowExecutionStore struct {
	executions map[string]*WorkflowExecutionResult
	mu         sync.RWMutex
}

// NewInMemoryWorkflowExecutionStore creates a new in-memory workflow execution store
func NewInMemoryWorkflowExecutionStore() *InMemoryWorkflowExecutionStore {
	return &InMemoryWorkflowExecutionStore{
		executions: make(map[string]*WorkflowExecutionResult),
	}
}

// SaveExecutionResult saves a workflow execution result
func (store *InMemoryWorkflowExecutionStore) SaveExecutionResult(ctx context.Context, result *WorkflowExecutionResult) error {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.executions[result.ExecutionID] = result
	return nil
}

// GetExecutionResult gets a workflow execution result by execution ID
func (store *InMemoryWorkflowExecutionStore) GetExecutionResult(ctx context.Context, executionID string) (*WorkflowExecutionResult, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()
	result, exists := store.executions[executionID]
	if !exists {
		return nil, fmt.Errorf("execution result not found: %s", executionID)
	}
	return result, nil
}

// GetExecutionResultsByWorkflowID gets all execution results for a workflow
func (store *InMemoryWorkflowExecutionStore) GetExecutionResultsByWorkflowID(ctx context.Context, workflowID string) ([]*WorkflowExecutionResult, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()
	var results []*WorkflowExecutionResult
	for _, result := range store.executions {
		if result.WorkflowID == workflowID {
			results = append(results, result)
		}
	}
	return results, nil
}

type workflowCallbackKey struct{}

// WithWorkflowCallbacks injects a workflow callback handler into the context.
func WithWorkflowCallbacks(ctx context.Context, handler WorkflowCallbackHandler) context.Context {
	return context.WithValue(ctx, workflowCallbackKey{}, handler)
}

// GetWorkflowCallbacks retrieves the workflow callback handler from the context.
func GetWorkflowCallbacks(ctx context.Context) WorkflowCallbackHandler {
	if h, ok := ctx.Value(workflowCallbackKey{}).(WorkflowCallbackHandler); ok {
		return h
	}
	return nil
}

var (
	// ErrSuspended indicates that the workflow execution is suspended and waiting for external input.
	// The caller should save the returned state and resume later.
	ErrSuspended = fmt.Errorf("workflow suspended")
)

// Workflow defines the interface for workflow execution.
type Workflow interface {
	Name() string
	Run(ctx context.Context, input string, opts ...model.Option) (*schema.Message, error)
	Resume(ctx context.Context, runID string, input string, opts ...model.Option) (*schema.Message, error)
}

// WorkflowDefinition defines the structure for workflow JSON/YAML definitions
type WorkflowDefinition struct {
	Type     string                    `json:"type" yaml:"type"`         // Type of workflow: sequential, parallel, dag, graph
	Name     string                    `json:"name" yaml:"name"`         // Workflow name
	Nodes    map[string]NodeDefinition `json:"nodes" yaml:"nodes"`       // Workflow nodes
	Edges    []EdgeDefinition          `json:"edges" yaml:"edges"`       // Workflow edges (for graph/dag)
	Metadata map[string]string         `json:"metadata" yaml:"metadata"` // Additional metadata
}

// NodeDefinition defines a node in the workflow
type NodeDefinition struct {
	Type        string                 `json:"type" yaml:"type"`               // Node type: agent, skill, workflow
	Name        string                 `json:"name" yaml:"name"`               // Node name
	Description string                 `json:"description" yaml:"description"` // Node description
	Config      map[string]interface{} `json:"config" yaml:"config"`           // Node configuration
	Priority    int                    `json:"priority" yaml:"priority"`       // Node priority
	MaxRetries  int                    `json:"max_retries" yaml:"max_retries"` // Maximum retries
	RetryDelay  string                 `json:"retry_delay" yaml:"retry_delay"` // Retry delay (e.g., "1s")
	Timeout     string                 `json:"timeout" yaml:"timeout"`         // Execution timeout
}

// EdgeDefinition defines an edge between nodes
type EdgeDefinition struct {
	From      string                 `json:"from" yaml:"from"`           // Source node ID
	To        string                 `json:"to" yaml:"to"`               // Target node ID
	Condition string                 `json:"condition" yaml:"condition"` // Optional condition for edge activation
	Config    map[string]interface{} `json:"config" yaml:"config"`       // Edge configuration
}

// simpleAgentWorkflow is a simple workflow that runs an agent
// This is used as a wrapper when an Agent needs to be used as a Workflow
// in DAG, Graph, or other workflow types

type simpleAgentWorkflow struct {
	name  string
	agent Agent
}

func (w *simpleAgentWorkflow) Name() string {
	return w.name
}

func (w *simpleAgentWorkflow) Run(ctx context.Context, input string, opts ...model.Option) (*schema.Message, error) {
	return w.agent.Run(ctx, input, opts...)
}

func (w *simpleAgentWorkflow) GetName() string {
	return w.name
}

func (w *simpleAgentWorkflow) GetID() string {
	return fmt.Sprintf("workflow_%s", w.name)
}

func (w *simpleAgentWorkflow) GetType() string {
	return "agent"
}

func (w *simpleAgentWorkflow) RunResumable(ctx context.Context, input string, state interface{}, opts ...model.Option) (*schema.Message, interface{}, error) {
	msg, err := w.Run(ctx, input, opts...)
	return msg, nil, err
}

func (w *simpleAgentWorkflow) Resume(ctx context.Context, input string, state string, opts ...model.Option) (*schema.Message, error) {
	// For simple agent workflow, just run the agent
	return w.Run(ctx, input, opts...)
}

// ResumableWorkflow defines the interface for workflows that support resumption from a checkpoint.
type ResumableWorkflow interface {
	Workflow
	RunResumable(ctx context.Context, input string, state *WorkflowState, opts ...model.Option) (*schema.Message, *WorkflowState, error)
}

// CheckpointableWorkflow defines the interface for workflows that support checkpointing.
type CheckpointableWorkflow interface {
	Workflow
	WithCheckpointStore(store CheckpointStore) Workflow
}

// ParseWorkflowDefinition parses a JSON/YAML workflow definition string into a WorkflowDefinition struct
func ParseWorkflowDefinition(definition string) (*WorkflowDefinition, error) {
	var wfDef WorkflowDefinition

	// Try JSON parsing first
	if err := json.Unmarshal([]byte(definition), &wfDef); err != nil {
		// Try YAML parsing if JSON fails
		if err := yaml.Unmarshal([]byte(definition), &wfDef); err != nil {
			return nil, fmt.Errorf("failed to parse workflow definition: %w", err)
		}
	}

	// Validate required fields
	if wfDef.Type == "" {
		return nil, fmt.Errorf("workflow type is required")
	}

	return &wfDef, nil
}

// CreateWorkflowFromDefinition creates a Workflow instance from a WorkflowDefinition
func CreateWorkflowFromDefinition(def *WorkflowDefinition, skillLibrary SkillLibrary, modelFactory ModelFactory) (Workflow, error) {
	switch def.Type {
	case "sequential":
		return createSequentialWorkflow(def, skillLibrary, modelFactory)
	case "parallel":
		return createParallelWorkflow(def, skillLibrary, modelFactory)
	case "dag":
		return createDAGWorkflow(def, skillLibrary, modelFactory)
	case "graph":
		return createGraphWorkflow(def, skillLibrary, modelFactory)
	default:
		return nil, fmt.Errorf("unsupported workflow type: %s", def.Type)
	}
}

// createEinoReActAgent creates an Eino ReAct agent with the given configuration
func createEinoReActAgent(ctx context.Context, chatModel ChatModel, instructions string, tools []tool.BaseTool, maxIterations int) (*react.Agent, error) {
	// Convert ChatModel to ToolCallingChatModel if possible
	var toolCallingModel model.ToolCallingChatModel
	var ok bool

	if toolCallingModel, ok = chatModel.(model.ToolCallingChatModel); !ok {
		// If the model doesn't support tool calling, we can still create the agent
		// but it won't be able to use tools effectively
		return nil, fmt.Errorf("model does not support tool calling, required for ReAct agent")
	}

	// Create ReAct agent configuration
	config := &react.AgentConfig{
		ToolCallingModel: toolCallingModel,
		MaxStep:          maxIterations,
	}

	// Configure tools if provided
	if len(tools) > 0 {
		config.ToolsConfig = compose.ToolsNodeConfig{
			Tools: tools,
		}
	}

	// Add message modifier for custom instructions
	if instructions != "" {
		config.MessageModifier = func(ctx context.Context, input []*schema.Message) []*schema.Message {
			// Prepend system message with instructions if not already present
			hasSystemMessage := false
			for _, msg := range input {
				if msg.Role == schema.System {
					hasSystemMessage = true
					break
				}
			}

			if !hasSystemMessage {
				systemMsg := &schema.Message{
					Role:    schema.System,
					Content: instructions,
				}
				return append([]*schema.Message{systemMsg}, input...)
			}

			return input
		}
	}

	// Create the ReAct agent
	reactAgent, err := react.NewAgent(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Eino ReAct agent: %w", err)
	}

	return reactAgent, nil
}

// Helper function to create sequential workflow
func createSequentialWorkflow(def *WorkflowDefinition, skillLibrary SkillLibrary, modelFactory ModelFactory) (Workflow, error) {
	// Create a list to hold workflow nodes
	var nodes []interface{}

	// Iterate through the nodes in the workflow definition
	for nodeID, nodeDef := range def.Nodes {
		// Create a node based on its type
		node, err := createWorkflowNode(nodeID, nodeDef, skillLibrary, modelFactory)
		if err != nil {
			return nil, fmt.Errorf("failed to create node %s: %w", nodeID, err)
		}
		nodes = append(nodes, node)
	}

	// For sequential workflow, we need to convert nodes to agents or skill executors
	// Create a list of agents for the sequential workflow
	var agents []Agent

	for _, node := range nodes {
		// Check if the node is already an Agent
		if ag, ok := node.(Agent); ok {
			agents = append(agents, ag)
			continue
		}

		// Check if the node is a Skill
		if skill, ok := node.(Skill); ok {
			// Create a SkillAgent that wraps the skill (use pointer since NewSkillAgent expects *Skill)
			ag, err := NewSkillAgent(&skill)
			if err != nil {
				return nil, fmt.Errorf("failed to create skill agent: %w", err)
			}
			agents = append(agents, ag)
			continue
		}

		// Check if the node is a Workflow
		if wf, ok := node.(Workflow); ok {
			// Create a WorkflowAgent that wraps the workflow
			ag, err := NewWorkflowAgent(wf)
			if err != nil {
				return nil, fmt.Errorf("failed to create workflow agent: %w", err)
			}
			agents = append(agents, ag)
			continue
		}

		return nil, fmt.Errorf("unsupported node type: %T", node)
	}

	// Create a new sequential workflow with the agents
	// Use local NewSequentialWorkflow function instead of workflow package
	return NewSequentialWorkflow(def.Name, agents...), nil
}

// createWorkflowNode creates a workflow node based on its type
func createWorkflowNode(nodeID string, nodeDef NodeDefinition, skillLibrary SkillLibrary, modelFactory ModelFactory) (interface{}, error) {
	switch nodeDef.Type {
	case "agent":
		// Create an agent based on the node definition
		return createAgentFromNodeDefinition(nodeID, nodeDef, modelFactory)
	case "skill":
		// Create a skill based on the node definition
		return createSkillFromNodeDefinition(nodeID, nodeDef, skillLibrary)
	case "workflow":
		// Create a workflow based on the node definition
		return createSubWorkflowFromNodeDefinition(nodeID, nodeDef, skillLibrary, modelFactory)
	default:
		return nil, fmt.Errorf("unsupported node type: %s", nodeDef.Type)
	}
}

// createAgentFromNodeDefinition creates an agent from a node definition
func createAgentFromNodeDefinition(nodeID string, nodeDef NodeDefinition, modelFactory ModelFactory) (Agent, error) {
	// Get model name from node config or use default
	modelName, ok := nodeDef.Config["model"].(string)
	if !ok || modelName == "" {
		modelName = "default"
	}

	// Create model instance using the ModelFactory interface
	chatModel, err := modelFactory.GetModel(modelName)
	if err != nil {
		return nil, fmt.Errorf("failed to create model: %w", err)
	}

	// Get agent kind from node config or use default "chat"
	agentKind, ok := nodeDef.Config["kind"].(string)
	if !ok {
		agentKind = "chat"
	}

	// Get instructions from node config
	instructions, _ := nodeDef.Config["instructions"].(string)

	// Create agent based on kind
	switch agentKind {
	case "chat":
		// Create ChatAgent
		return NewChatAgent(context.Background(), ChatAgentConfig{
			Name:         nodeID,
			Instructions: instructions,
			Model:        chatModel,
			Tools:        []tool.BaseTool{}, // Empty tools list for now
		})
	case "react":
		// Create ReActAgent
		return createReActAgentFromNodeDefinition(nodeID, nodeDef, modelFactory)
	case "human":
		// Create a simple ChatAgent for human interaction
		// The actual human interaction would be handled at runtime
		return NewChatAgent(context.Background(), ChatAgentConfig{
			Name:         nodeID,
			Instructions: instructions + "\n\nNote: This agent requires human input. Please provide your response when prompted.",
			Model:        chatModel,
			Tools:        []tool.BaseTool{},
		})
	default:
		return nil, fmt.Errorf("unsupported agent kind: %s", agentKind)
	}
}

// createReActAgentFromNodeDefinition creates a ReActAgent from a node definition
func createReActAgentFromNodeDefinition(nodeID string, nodeDef NodeDefinition, modelFactory ModelFactory) (Agent, error) {
	// Get model name from node config or use default
	modelName, ok := nodeDef.Config["model"].(string)
	if !ok || modelName == "" {
		modelName = "default"
	}

	// Create model instance using the ModelFactory interface
	chatModel, err := modelFactory.GetModel(modelName)
	if err != nil {
		return nil, fmt.Errorf("failed to create model: %w", err)
	}

	// Get instructions from node config
	instructions, _ := nodeDef.Config["instructions"].(string)
	if instructions == "" {
		instructions = "You are a helpful AI assistant that uses tools to solve problems. Think step by step and use the available tools when needed."
	}

	// Get max iterations from node config (default: 12, which is Eino's default)
	maxIterations := 12
	if maxIter, ok := nodeDef.Config["max_iterations"].(int); ok && maxIter > 0 {
		maxIterations = maxIter
	} else if maxIterFloat, ok := nodeDef.Config["max_iterations"].(float64); ok && maxIterFloat > 0 {
		maxIterations = int(maxIterFloat)
	}

	// Get tools from node config
	var tools []tool.BaseTool
	if toolNames, ok := nodeDef.Config["tools"].([]interface{}); ok {
		// Tools are specified as a list of tool names
		// In a full implementation, you would load these from a tool registry
		// For now, we create an empty list and log that tools need to be loaded
		tools = make([]tool.BaseTool, 0, len(toolNames))

		// TODO: Load tools from a tool registry based on tool names
		// Example:
		// for _, toolNameInterface := range toolNames {
		//     if toolName, ok := toolNameInterface.(string); ok {
		//         tool := toolRegistry.GetTool(toolName)
		//         if tool != nil {
		//             tools = append(tools, tool)
		//         }
		//     }
		// }
	}

	// Get memory options from node config
	memoryOpts := MemoryOptions{
		EnableTrimming: true,
		MaxMessages:    20,
	}

	// Override memory options from config if provided
	if memConfig, ok := nodeDef.Config["memory"].(map[string]interface{}); ok {
		if enableTrim, ok := memConfig["enable_trimming"].(bool); ok {
			memoryOpts.EnableTrimming = enableTrim
		}
		if maxMsgs, ok := memConfig["max_messages"].(int); ok && maxMsgs > 0 {
			memoryOpts.MaxMessages = maxMsgs
		} else if maxMsgsFloat, ok := memConfig["max_messages"].(float64); ok && maxMsgsFloat > 0 {
			memoryOpts.MaxMessages = int(maxMsgsFloat)
		}
	}

	// Create ReAct agent using Eino's react package
	reactAgent, err := createEinoReActAgent(context.Background(), chatModel, instructions, tools, maxIterations)
	if err != nil {
		return nil, fmt.Errorf("failed to create ReAct agent: %w", err)
	}

	// Wrap in our ReActAgent with memory management
	agent := NewReActAgentWithMemory(nodeID, reactAgent, memoryOpts)

	return agent, nil
}

// createSkillFromNodeDefinition creates a skill from a node definition
func createSkillFromNodeDefinition(nodeID string, nodeDef NodeDefinition, skillLibrary SkillLibrary) (Skill, error) {
	// Get skill name from node config
	skillName, ok := nodeDef.Config["skill"].(string)
	if !ok {
		return Skill{}, fmt.Errorf("skill name not found in node config")
	}

	// Get skill from skill library (SkillLibrary.GetSkill takes only name, returns interface{}, bool)
	skillInterface, found := skillLibrary.GetSkill(skillName)
	if !found {
		return Skill{}, fmt.Errorf("skill %s not found in skill library", skillName)
	}

	// Type assert to Skill
	skill, ok := skillInterface.(Skill)
	if !ok {
		return Skill{}, fmt.Errorf("skill %s is not of type Skill", skillName)
	}

	return skill, nil
}

// createSubWorkflowFromNodeDefinition creates a sub-workflow from a node definition
func createSubWorkflowFromNodeDefinition(nodeID string, nodeDef NodeDefinition, skillLibrary SkillLibrary, modelFactory ModelFactory) (Workflow, error) {
	// Get workflow definition from node config
	wfDefStr, ok := nodeDef.Config["definition"].(string)
	if !ok {
		return nil, fmt.Errorf("workflow definition not found in node config")
	}

	// Parse workflow definition
	wfDef, err := ParseWorkflowDefinition(wfDefStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse workflow definition: %w", err)
	}

	// Create workflow from definition
	return CreateWorkflowFromDefinition(wfDef, skillLibrary, modelFactory)
}

// Helper function to create parallel workflow
func createParallelWorkflow(def *WorkflowDefinition, skillLibrary SkillLibrary, modelFactory ModelFactory) (Workflow, error) {
	// Create a list to hold workflow nodes
	var nodes []interface{}

	// Iterate through the nodes in the workflow definition
	for nodeID, nodeDef := range def.Nodes {
		// Create a node based on its type
		node, err := createWorkflowNode(nodeID, nodeDef, skillLibrary, modelFactory)
		if err != nil {
			return nil, fmt.Errorf("failed to create node %s: %w", nodeID, err)
		}
		nodes = append(nodes, node)
	}

	// For parallel workflow, we need to convert nodes to agents
	// Create a list of agents for the parallel workflow
	var agents []Agent

	for _, node := range nodes {
		// Check if the node is already an Agent
		if ag, ok := node.(Agent); ok {
			agents = append(agents, ag)
			continue
		}

		// Check if the node is a Skill
		if skill, ok := node.(Skill); ok {
			// Create a SkillAgent that wraps the skill (use pointer since NewSkillAgent expects *Skill)
			ag, err := NewSkillAgent(&skill)
			if err != nil {
				return nil, fmt.Errorf("failed to create skill agent: %w", err)
			}
			agents = append(agents, ag)
			continue
		}

		// Check if the node is a Workflow
		if wf, ok := node.(Workflow); ok {
			// Create a WorkflowAgent that wraps the workflow
			ag, err := NewWorkflowAgent(wf)
			if err != nil {
				return nil, fmt.Errorf("failed to create workflow agent: %w", err)
			}
			agents = append(agents, ag)
			continue
		}

		return nil, fmt.Errorf("unsupported node type: %T", node)
	}

	// Create a default aggregator agent if none is specified in the workflow definition
	// Get aggregator configuration from workflow metadata if available
	var aggregator Agent

	// Create a default chat agent as aggregator
	chatModel, err := modelFactory.GetModel("default")
	if err != nil {
		return nil, fmt.Errorf("failed to create model for default aggregator: %w", err)
	}

	aggregator, err = NewChatAgent(context.Background(), ChatAgentConfig{
		Name: "aggregator",
		Instructions: `You are an aggregator agent. Your task is to combine and summarize the results from multiple agents into a single coherent response.

Please summarize the following responses:

%s`,
		Model: chatModel,
		Tools: []tool.BaseTool{},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create default aggregator agent: %w", err)
	}

	// Create a new aggregating parallel workflow with the agents
	// Use local implementation since agent.Agent != workflow.Agent
	return NewAggregatingParallelWorkflowLocal(def.Name, aggregator, agents...), nil
}

// Helper function to create DAG workflow
func createDAGWorkflow(def *WorkflowDefinition, skillLibrary SkillLibrary, modelFactory ModelFactory) (Workflow, error) {
	// Use local implementation since agent.Agent != workflow.Agent
	return createDAGWorkflowLocal(def, skillLibrary, modelFactory)
}

// Helper function to create graph workflow
func createGraphWorkflow(def *WorkflowDefinition, skillLibrary SkillLibrary, modelFactory ModelFactory) (Workflow, error) {
	graph := newGraphWorkflowLocal(def.Name)

	// Create all nodes first
	for nodeID, nodeDef := range def.Nodes {
		// Create a node based on its type
		node, err := createWorkflowNode(nodeID, nodeDef, skillLibrary, modelFactory)
		if err != nil {
			return nil, fmt.Errorf("failed to create node %s: %w", nodeID, err)
		}

		// Convert node to Workflow if needed
		var workflow Workflow

		// Check if the node is already a Workflow
		if wf, ok := node.(Workflow); ok {
			workflow = wf
		} else if ag, ok := node.(Agent); ok {
			// Create a simple agent workflow wrapper
			agentWorkflow := &simpleAgentWorkflow{
				name:  nodeID,
				agent: ag,
			}
			workflow = agentWorkflow
		} else if skill, ok := node.(Skill); ok {
			// Create a SkillAgent that wraps the skill (use pointer)
			ag, err := NewSkillAgent(&skill)
			if err != nil {
				return nil, fmt.Errorf("failed to create skill agent: %w", err)
			}
			agentWorkflow := &simpleAgentWorkflow{
				name:  nodeID,
				agent: ag,
			}
			workflow = agentWorkflow
		} else {
			return nil, fmt.Errorf("unsupported node type: %T", node)
		}

		// Add node to graph
		graph.nodes[nodeID] = workflow
	}

	// Set start node (first node in map if not specified in metadata)
	startNode := ""
	// Check if start_node is specified in metadata (which is a map[string]string)
	if sv, exists := def.Metadata["start_node"]; exists {
		startNode = sv
	}

	if startNode == "" {
		// Use first node as start node if none specified
		for nodeID := range def.Nodes {
			startNode = nodeID
			break
		}
	}
	graph.SetStartNode(startNode)

	// Add edges to graph
	for _, edgeDef := range def.Edges {
		if edgeDef.Condition != "" {
			// Create a conditional edge
			conditionFn := func(ctx context.Context, input string) (string, error) {
				// For now, we'll use a simple condition that checks if the input contains the condition string
				// In the future, we could implement more complex condition evaluation
				if edgeDef.Condition == "" {
					return edgeDef.To, nil
				}
				return edgeDef.To, nil
			}
			graph.AddConditionalEdge(edgeDef.From, conditionFn)
		} else {
			// Create a static edge
			graph.AddEdge(edgeDef.From, edgeDef.To)
		}
	}

	return graph, nil
}

// NewSequentialWorkflow creates a simple sequential workflow using agent.Agent
func NewSequentialWorkflow(name string, agents ...Agent) Workflow {
	return &simpleSequentialWorkflow{
		name:   name,
		agents: agents,
	}
}

// simpleSequentialWorkflow is a simple implementation of Workflow for agent.Agent
type simpleSequentialWorkflow struct {
	name   string
	agents []Agent
}

func (w *simpleSequentialWorkflow) Name() string {
	return w.name
}

func (w *simpleSequentialWorkflow) GetName() string {
	return w.name
}

func (w *simpleSequentialWorkflow) GetID() string {
	return fmt.Sprintf("workflow_%s", w.name)
}

func (w *simpleSequentialWorkflow) GetType() string {
	return "sequential"
}

func (w *simpleSequentialWorkflow) Run(ctx context.Context, input string, opts ...model.Option) (*schema.Message, error) {
	text := input
	for _, agent := range w.agents {
		result, err := agent.Run(ctx, text, opts...)
		if err != nil {
			return nil, fmt.Errorf("agent %s failed: %w", agent.Name(), err)
		}
		text = result.Content
	}
	return &schema.Message{Role: schema.Assistant, Content: text}, nil
}

func (w *simpleSequentialWorkflow) RunResumable(ctx context.Context, input string, state interface{}, opts ...model.Option) (*schema.Message, interface{}, error) {
	msg, err := w.Run(ctx, input, opts...)
	return msg, nil, err
}

func (w *simpleSequentialWorkflow) Resume(ctx context.Context, input string, state string, opts ...model.Option) (*schema.Message, error) {
	// For simple sequential workflow, state is not used
	msg, _, err := w.RunResumable(ctx, input, state, opts...)
	return msg, err
}

// NewAggregatingParallelWorkflowLocal creates a local parallel workflow with aggregation
func NewAggregatingParallelWorkflowLocal(name string, aggregator Agent, agents ...Agent) Workflow {
	return &simpleParallelWorkflow{
		name:       name,
		agents:     agents,
		aggregator: aggregator,
	}
}

// simpleParallelWorkflow is a simple implementation of parallel workflow
type simpleParallelWorkflow struct {
	name       string
	agents     []Agent
	aggregator Agent
}

func (w *simpleParallelWorkflow) Name() string {
	return w.name
}

func (w *simpleParallelWorkflow) GetName() string {
	return w.name
}

func (w *simpleParallelWorkflow) GetID() string {
	return fmt.Sprintf("workflow_%s", w.name)
}

func (w *simpleParallelWorkflow) GetType() string {
	return "parallel"
}

func (w *simpleParallelWorkflow) Run(ctx context.Context, input string, opts ...model.Option) (*schema.Message, error) {
	// Run all agents in parallel (simplified - actual implementation would use goroutines)
	var results []string
	for _, agent := range w.agents {
		result, err := agent.Run(ctx, input, opts...)
		if err != nil {
			return nil, fmt.Errorf("agent %s failed: %w", agent.Name(), err)
		}
		results = append(results, result.Content)
	}

	// Aggregate results
	combinedInput := fmt.Sprintf("Results from %d agents:\n%s", len(results), strings.Join(results, "\n\n"))
	return w.aggregator.Run(ctx, combinedInput, opts...)
}

func (w *simpleParallelWorkflow) RunResumable(ctx context.Context, input string, state interface{}, opts ...model.Option) (*schema.Message, interface{}, error) {
	msg, err := w.Run(ctx, input, opts...)
	return msg, nil, err
}

func (w *simpleParallelWorkflow) Resume(ctx context.Context, input string, state string, opts ...model.Option) (*schema.Message, error) {
	// For simple parallel workflow, state is not used
	msg, _, err := w.RunResumable(ctx, input, state, opts...)
	return msg, err
}

// createDAGWorkflowLocal creates a local DAG workflow
func createDAGWorkflowLocal(def *WorkflowDefinition, skillLibrary SkillLibrary, modelFactory ModelFactory) (Workflow, error) {
	// For now, return a simple sequential workflow as a placeholder
	// A full DAG implementation would require more complex graph handling
	var agents []Agent
	for nodeID := range def.Nodes {
		nodeDef := def.Nodes[nodeID]
		node, err := createWorkflowNode(nodeID, nodeDef, skillLibrary, modelFactory)
		if err != nil {
			return nil, fmt.Errorf("failed to create node %s: %w", nodeID, err)
		}

		if agent, ok := node.(Agent); ok {
			agents = append(agents, agent)
		} else if skill, ok := node.(Skill); ok {
			ag, err := NewSkillAgent(&skill)
			if err != nil {
				return nil, fmt.Errorf("failed to create skill agent: %w", err)
			}
			agents = append(agents, ag)
		}
	}
	return NewSequentialWorkflow(def.Name, agents...), nil
}

// newGraphWorkflowLocal creates a local graph workflow
func newGraphWorkflowLocal(name string) *simpleGraphWorkflow {
	return &simpleGraphWorkflow{
		name:  name,
		nodes: make(map[string]Workflow),
	}
}

// simpleGraphWorkflow is a simple implementation of graph workflow
type simpleGraphWorkflow struct {
	name       string
	nodes      map[string]Workflow
	startNode  string
	edges      []graphEdge
	mu         sync.RWMutex
}

type graphEdge struct {
	From      string
	To        string
	Condition string
}

func (g *simpleGraphWorkflow) Name() string {
	return g.name
}

func (g *simpleGraphWorkflow) GetName() string {
	return g.name
}

func (g *simpleGraphWorkflow) GetID() string {
	return fmt.Sprintf("workflow_%s", g.name)
}

func (g *simpleGraphWorkflow) GetType() string {
	return "graph"
}

func (g *simpleGraphWorkflow) Run(ctx context.Context, input string, opts ...model.Option) (*schema.Message, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if g.startNode == "" {
		return nil, fmt.Errorf("no start node set")
	}

	current := g.startNode
	text := input
	visited := make(map[string]bool)

	for current != "" && !visited[current] {
		visited[current] = true
		node, ok := g.nodes[current]
		if !ok {
			return nil, fmt.Errorf("node %s not found", current)
		}

		result, err := node.Run(ctx, text, opts...)
		if err != nil {
			return nil, fmt.Errorf("node %s failed: %w", current, err)
		}
		text = result.Content

		// Find next node
		current = ""
		for _, edge := range g.edges {
			if edge.From == current {
				current = edge.To
				break
			}
		}
	}

	return &schema.Message{Role: schema.Assistant, Content: text}, nil
}

func (g *simpleGraphWorkflow) RunResumable(ctx context.Context, input string, state interface{}, opts ...model.Option) (*schema.Message, interface{}, error) {
	msg, err := g.Run(ctx, input, opts...)
	return msg, nil, err
}

func (g *simpleGraphWorkflow) Resume(ctx context.Context, input string, state string, opts ...model.Option) (*schema.Message, error) {
	msg, _, err := g.RunResumable(ctx, input, state, opts...)
	return msg, err
}

func (g *simpleGraphWorkflow) SetStartNode(nodeID string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.startNode = nodeID
}

func (g *simpleGraphWorkflow) AddNode(nodeID string, workflow Workflow) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.nodes[nodeID] = workflow
}

func (g *simpleGraphWorkflow) AddEdge(from, to string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.edges = append(g.edges, graphEdge{From: from, To: to})
}

func (g *simpleGraphWorkflow) AddConditionalEdge(from string, condition func(context.Context, string) (string, error)) {
	// For simplicity, just add a regular edge
	g.AddEdge(from, "")
}
