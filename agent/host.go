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
	"time"

	"github.com/cloudwego/eino/components/tool"

	"AgentFramework/agent/messaging"
	"AgentFramework/agent/scheduler"
	"AgentFramework/agent/async"
	"AgentFramework/agent/heartbeat"
	"AgentFramework/agent/token"
)

type ModelFactory func(ctx context.Context, name string) (ChatModel, error)

type Host struct {
	cfg             *HostConfig
	configMgr       ConfigManager
	modelFactory    ModelFactory
	threadStore     ThreadStore
	toolRegistry    map[string]tool.BaseTool
	monitorMgr      *MonitorManager
	pluginMgr       PluginManager
	channelMgr      *messaging.ChannelManager
	scheduler      interface{} // *scheduler.Scheduler*/
	heartbeat      interface{} // *heartbeat.HeartbeatService*
	taskManager     interface{} // *async.TaskManager*/
	tokenCompressor interface{} // *token.MessageCompressor*/

	agents      map[string]Agent
	workflows   map[string]Workflow
	middlewares map[string]AgentMiddleware

	service *AgentService
}

// NewHost creates a new Host instance with the given configuration and options
func NewHost(ctx context.Context, cfg *HostConfig, mf ModelFactory, tr map[string]tool.BaseTool, opts ...HostOption) (*Host, error) {
	// Apply host options
	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, err
		}
	}

	// If no ModelFactory provided, create one from config
	if mf == nil {
		mf = NewDefaultModelFactory(DefaultModelFactoryConfig{
			Models: cfg.Models,
		})
	}

	// Create config manager
	configMgr := NewConfigManager(cfg)

	// Create monitor manager with memory monitor
	memoryMonitorConfig := MemoryMonitorConfig{
		Enabled:       cfg.Memory.MemoryMonitor.Enabled,
		Interval:      time.Duration(cfg.Memory.MemoryMonitor.Interval) * time.Second,
		HistorySize:   cfg.Memory.MemoryMonitor.HistorySize,
		AlertInterval: time.Duration(cfg.Memory.MemoryMonitor.AlertInterval) * time.Second,
	}

	memoryMonitor := NewMemoryMonitor(memoryMonitorConfig)

	monitorManager := NewMonitorManager(MonitorManagerConfig{
		Enabled:       cfg.Memory.MemoryMonitor.Enabled,
		Monitors:      []Monitor{memoryMonitor},
		Storage:       nil, // No storage by default
		AlertHandlers: []AlertHandler{},
	})

	// Create plugin manager
	pluginMgr := NewPluginManager()

	// Create async task manager (optional)
	var taskManagerInterface interface{}
	if cfg.AsyncTask != nil && cfg.AsyncTask.Enabled {
		taskManagerInterface = async.NewMemoryTaskManager()
		if sched, ok := taskManagerInterface.(interface{ Start(context.Context) error }); ok {
			if err := sched.Start(ctx); err != nil {
				return nil, fmt.Errorf("failed to start task manager: %w", err)
			}
		}
	}
	// Create scheduler (optional)
	var schedulerInterface interface{}
	if cfg.Scheduler != nil && cfg.Scheduler.Enabled {
		schedulerConfig := &scheduler.SchedulerConfig{
			Enabled:           true,
			Timezone:          cfg.Scheduler.Timezone,
			MaxConcurrentJobs: cfg.Scheduler.MaxJobs,
			JobTimeout:        time.Duration(cfg.Scheduler.JobTimeout) * time.Second,
			Logger:            &scheduler.DefaultLogger{},
		}
		schedulerInterface = scheduler.NewMemoryScheduler(schedulerConfig)
	}
	// Create heartbeat service (optional)
	var heartbeatInterface interface{}
	if cfg.Heartbeat != nil && cfg.Heartbeat.Enabled {
		heartbeatConfig := &heartbeat.HeartbeatConfig{
			Interval: time.Duration(cfg.Heartbeat.Interval) * time.Second,
			Timeout:  time.Duration(cfg.Heartbeat.Timeout) * time.Second,
			Logger:   &heartbeat.DefaultLogger{},
		}
		heartbeatInterface = heartbeat.NewMemoryHeartbeatService(heartbeatConfig)
	}
	// Create token compressor (optional)
	var tokenCompressorInterface interface{}
	if cfg.TokenCompression != nil && cfg.TokenCompression.Enabled {
		compressorConfig := &token.CompressConfig{
			Strategy:             token.CompressionStrategy(cfg.TokenCompression.Strategy),
			TargetTokens:         cfg.TokenCompression.TargetTokens,
			MinTokens:            cfg.TokenCompression.MinTokens,
			MaxTokens:            cfg.TokenCompression.MaxTokens,
			PreserveSystemMessages: cfg.TokenCompression.PreserveSystemMessages,
			SummaryModelName:    cfg.TokenCompression.SummaryModelName,
			SummaryMaxTokens:    cfg.TokenCompression.SummaryMaxTokens,
			Temperature:         cfg.TokenCompression.Temperature,
		}

		// 创建 LLM 压缩函数
		llmFunc := func(ctx context.Context, prompt string, maxTokens int) (string, error) {
			modelName := compressorConfig.SummaryModelName
			if modelName == "" {
				modelName = cfg.DefaultModel
			}

			model, err := mf(ctx, modelName)
			if err != nil {
				return "", fmt.Errorf("failed to get model for summarization: %w", err)
			}

			// 改进的系统提示，提供更详细的摘要指导
			systemPrompt := `你是一个专业的文本摘要专家。你的任务是将对话历史或长文本总结成简洁、准确、全面的摘要。

## 摘要要求：
1. **准确性**：保留所有关键信息，确保摘要与原文意思相符
2. **完整性**：包含所有重要的观点、决策、行动项目
3. **简洁性**：使用简洁的语言，避免冗余
4. **结构**：使用清晰的层次结构（如要点列表）
5. **客观性**：保持中立，避免个人观点

## 处理策略：
- 识别并突出关键信息
- 删除重复内容
- 简化复杂句子
- 保留所有专有名词和技术术语
- 对于对话历史，总结各参与者的主要观点

## 格式要求：
使用要点列表或段落形式，语言简洁明了。不要使用过于正式的语言，确保摘要易于理解。`

			// 构建优化的摘要请求
			messages := []interface{}{
				map[string]interface{}{
					"role":    "system",
					"content": systemPrompt,
				},
				map[string]interface{}{
					"role":    "user",
					"content": fmt.Sprintf("请总结以下内容，使用不超过 %d 个 token：\n\n%s", maxTokens, prompt),
				},
			}

			// 设置合理的参数，提高效率和质量
			result, err := model.Chat(ctx, messages, maxTokens, 0.3) // 使用较低的温度提高一致性
			if err != nil {
				return "", fmt.Errorf("failed to generate summary: %w", err)
			}

			// 清理和优化结果
			summary := strings.TrimSpace(result)
			if summary == "" {
				return "", fmt.Errorf("empty summary generated")
			}

			// 如果摘要仍然过长，进行截断
			if len(summary) > maxTokens*4 { // 估算字符数
				summary = summary[:maxTokens*4-3] + "..."
			}

			return summary, nil
		}

		tokenCompressorInterface = token.NewMessageCompressor(compressorConfig, llmFunc)
	}
	var channelMgr *messaging.ChannelManager
	if cfg.Messaging != nil && cfg.Messaging.Enabled {
		// 创建独立的 EventBus（避免循环依赖）
		eventBus := messaging.NewSimpleEventBus()

		channelConfig := messaging.ChannelManagerConfig{
			Logger:        &defaultLogger{},
			EventBus:      eventBus,
			EnableMetrics: cfg.Messaging.EnableMetrics,
		}
		var err error
		channelMgr, err = messaging.NewChannelManager(channelConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create channel manager: %w", err)
		}
	}

	h := &Host{
		cfg:          cfg,
		configMgr:    configMgr,
		modelFactory: mf,
		threadStore:  nil,
		toolRegistry: tr,
		monitorMgr:   monitorManager,
		pluginMgr:    pluginMgr,
		channelMgr:   channelMgr,
		scheduler:   schedulerInterface,
		taskManager: taskManagerInterface,
		heartbeat:   heartbeatInterface,
		tokenCompressor: tokenCompressorInterface,
		agents:       make(map[string]Agent),
		workflows:    make(map[string]Workflow),
		middlewares:  make(map[string]AgentMiddleware),
	}

	if err := h.initThreadStore(ctx); err != nil {
		return nil, err
	}

	h.service = NewAgentService(h.threadStore)

	h.registerDefaultMiddlewares()

	// Initialize skill system if configured
	if cfg.SkillSystemDir != "" {
		if err := h.InitializeSkillSystem(cfg.SkillSystemDir); err != nil {
			return nil, fmt.Errorf("failed to initialize skill system: %w", err)
		}
	}

	if err := h.buildAgents(ctx, tr); err != nil {
		return nil, err
	}

	if err := h.buildWorkflows(ctx); err != nil {
		return nil, err
	}

	// Start channel manager if configured
	if h.channelMgr != nil {
		if err := h.channelMgr.Start(ctx); err != nil {
			return nil, fmt.Errorf("failed to start channel manager: %w", err)
		}
	}

t// Start async task manager if configured	if h.taskManager != nil {		if tm, ok := h.taskManager.(interface{ Start(context.Context) error }); ok {			// TaskManager 在 Start 中已经启动，这里只需要检查			_ = tm		}	}
// Start scheduler if configured	if h.scheduler != nil {		if sched, ok := h.scheduler.(interface{ Start(context.Context) error }); ok {			if err := sched.Start(ctx); err != nil {				return nil, fmt.Errorf("failed to start scheduler: %w", err)			}		}	}	// Start heartbeat service if configured	if h.heartbeat != nil {		if hb, ok := h.heartbeat.(interface{ Start(context.Context) error }); ok {			if err := hb.Start(ctx); err != nil {				return nil, fmt.Errorf("failed to start heartbeat service: %w", err)			}		}	}
	return h, nil
}

func (h *Host) AddWorkflow(wf Workflow) {
	h.workflows[wf.Name()] = wf
}

func (h *Host) GetWorkflow(name string) (Workflow, error) {
	wf, ok := h.workflows[name]
	if !ok {
		return nil, fmt.Errorf("workflow %q not found", name)
	}
	return wf, nil
}

func (h *Host) GetAgent(name string) (Agent, error) {
	ag, ok := h.agents[name]
	if !ok {
		return nil, fmt.Errorf("agent %q not found", name)
	}
	return ag, nil
}

func (h *Host) Service() *AgentService {
	return h.service
}

func (h *Host) Config() *HostConfig {
	return h.cfg
}

func (h *Host) RegisterMiddleware(name string, middleware AgentMiddleware) {
	h.middlewares[name] = middleware
}

func (h *Host) ThreadStore() ThreadStore {
	return h.threadStore
}

// PluginManager returns the plugin manager
func (h *Host) PluginManager() PluginManager {
	return h.pluginMgr
}

// ChannelManager returns the channel manager (may be nil if not configured)
func (h *Host) ChannelManager() *messaging.ChannelManager {
	return h.channelMgr
}

// Agent returns an agent by name, alias for GetAgent
func (h *Host) Agent(name string) (Agent, bool) {
	ag, err := h.GetAgent(name)
	return ag, err == nil
}

// Workflow returns a workflow by name, alias for GetWorkflow
func (h *Host) Workflow(name string) (Workflow, bool) {
	wf, err := h.GetWorkflow(name)
	return wf, err == nil
}

// ListAgents returns a list of all agent names
func (h *Host) ListAgents() []string {
	agents := make([]string, 0, len(h.agents))
	for name := range h.agents {
		agents = append(agents, name)
	}
	return agents
}

// ListWorkflows returns a list of all workflow names
func (h *Host) ListWorkflows() []string {
	workflows := make([]string, 0, len(h.workflows))
	for name := range h.workflows {
		workflows = append(workflows, name)
	}
	return workflows
}

// WorkflowGraph represents a workflow graph structure for API response
// This is a simplified representation for the API
// It doesn't expose internal implementation details
// For complex workflows like DAG or Graph, it returns the actual graph structure
// For simple workflows, it returns a single-node graph
// For composite workflows, it returns the nested structure

// Node represents a node in the workflow graph
// ID, Name, Type fields are always present
// Other fields are optional depending on the workflow type
// Type can be: agent, workflow, human, inline
// AgentName is the name of the agent if the node is an agent
// WorkflowName is the name of the workflow if the node is a workflow
// Instructions are the instructions for the node if it's an inline agent
// Tools are the tools available to the node
// Children is used for nested workflows

// Edge represents an edge in the workflow graph
// From, To fields are always present
// Condition is optional for conditional edges

// WorkflowGraph represents a workflow graph structure
// Name is the workflow name
// Type is the workflow type (sequential, parallel, routing, planning, dag, graph)
// Nodes is the list of nodes in the graph
// Edges is the list of edges in the graph
// StartNode is the ID of the start node (for DAG and Graph workflows)

// GetWorkflowGraph returns the workflow graph structure for a given workflow name
// It returns a simplified graph representation based on the workflow type
func (h *Host) GetWorkflowGraph(name string) (interface{}, error) {
	wf, err := h.GetWorkflow(name)
	if err != nil {
		return nil, err
	}

	// For now, return a simple representation
	// We can enhance this later to return the actual graph structure
	// for different workflow types
	return map[string]interface{}{
		"name": wf.Name(),
		"type": "unknown", // We can determine the actual type by type assertion
		"nodes": []map[string]interface{}{
			{
				"id":   "root",
				"name": wf.Name(),
				"type": "workflow",
			},
		},
		"edges": []map[string]interface{}{},
	}, nil
}

// HostManager manages multiple Host instances
// It allows registering hosts by name and accessing their workflows
// This enables centralized management of multiple agent applications
// within a single process

type HostManager struct {
	hosts map[string]*Host
}

// NewHostManager creates a new HostManager instance
// This is used to manage multiple Host instances from different applications
func NewHostManager() *HostManager {
	return &HostManager{
		hosts: make(map[string]*Host),
	}
}

// Register adds a Host instance to the manager with the given app name
// If appName is empty, it uses the host's configuration name
func (hm *HostManager) Register(appName string, host *Host) {
	if appName == "" {
		appName = host.cfg.Name
	}
	hm.hosts[appName] = host
}

// Unregister removes a Host instance from the manager by app name
func (hm *HostManager) Unregister(appName string) {
	delete(hm.hosts, appName)
}

// Host returns a Host instance by app name
func (hm *HostManager) Host(appName string) (*Host, bool) {
	host, ok := hm.hosts[appName]
	return host, ok
}

// Workflow returns a workflow from a specific host by app name and workflow name
func (hm *HostManager) Workflow(appName, workflowName string) (Workflow, bool) {
	host, ok := hm.Host(appName)
	if !ok {
		return nil, false
	}
	return host.Workflow(workflowName)
}

// Agent returns an agent from a specific host by app name and agent name
func (hm *HostManager) Agent(appName, agentName string) (Agent, bool) {
	host, ok := hm.Host(appName)
	if !ok {
		return nil, false
	}
	return host.Agent(agentName)
}

// ListApps returns a list of all registered app names
func (hm *HostManager) ListApps() []string {
	apps := make([]string, 0, len(hm.hosts))
	for appName := range hm.hosts {
		apps = append(apps, appName)
	}
	return apps
}

// ListWorkflows returns a list of all workflow names for a given app
func (hm *HostManager) ListWorkflows(appName string) []string {
	host, ok := hm.Host(appName)
	if !ok {
		return nil
	}
	return host.ListWorkflows()
}

// ListAgents returns a list of all agent names for a given app
func (hm *HostManager) ListAgents(appName string) []string {
	host, ok := hm.Host(appName)
	if !ok {
		return nil
	}
	return host.ListAgents()
}
}

// ===== Scheduler 和 Heartbeat Getter 方法 =====

// Scheduler 返回调度器（可能为 nil）
func (h *Host) Scheduler() interface{} {
	return h.scheduler
}

// Heartbeat 返回心跳服务（可能为 nil）
func (h *Host) Heartbeat() interface{} {
	return h.heartbeat
}

// ScheduleJob 安排定时任务
func (h *Host) ScheduleJob(ctx context.Context, name string, handler interface{}) (string, error) {
	if h.scheduler == nil {
		return "", fmt.Errorf("scheduler not configured")
	}

	// 将通用函数转换为 JobHandler
	var jobHandler scheduler.JobHandler
	if hf, ok := handler.(func(context.Context) error); ok {
		jobHandler = hf
	} else {
		return "", fmt.Errorf("invalid handler type")
	}

	job := &scheduler.Job{
		Name:    name,
		Handler: jobHandler,
		Schedule: scheduler.JobSchedule{
			Type:     scheduler.ScheduleTypeInterval,
			Interval: 1 * time.Minute,
		},
	}

	if sched, ok := h.scheduler.(interface {
		ScheduleJob(context.Context, *scheduler.Job) (string, error)
	}); ok {
		return sched.ScheduleJob(ctx, job)
	}

	return "", fmt.Errorf("scheduler does not support ScheduleJob")
}

// SendHeartbeat 发送心跳
func (h *Host) SendHeartbeat(ctx context.Context) error {
	if h.heartbeat == nil {
		return fmt.Errorf("heartbeat service not configured")
	}

	if hb, ok := h.heartbeat.(interface {
		SendBeat(context.Context) error
	}); ok {
		return hb.SendBeat(ctx)
	}

	return fmt.Errorf("heartbeat service does not support SendBeat")
}
}

// ===== TokenCompressor Getter 方法 =====

// TokenCompressor 返回 Token 压缩器（可能为 nil）
func (h *Host) TokenCompressor() interface{} {
	return h.tokenCompressor
}

// CompressMessages 压缩消息列表到目标 Token 数量
func (h *Host) CompressMessages(ctx context.Context, messages []interface{}, targetTokens int) ([]interface{}, error) {
	if h.tokenCompressor == nil {
		return nil, fmt.Errorf("token compressor not configured")
	}

	if tc, ok := h.tokenCompressor.(interface {
		CompressMessages(context.Context, []interface{}, int) ([]interface{}, error)
	}); ok {
		return tc.CompressMessages(ctx, messages, targetTokens)
	}

	return nil, fmt.Errorf("token compressor does not support CompressMessages")
}

// CompressText 压缩文本到目标 Token 数量
func (h *Host) CompressText(ctx context.Context, text string, targetTokens int) (string, error) {
	if h.tokenCompressor == nil {
		return "", fmt.Errorf("token compressor not configured")
	}

	if tc, ok := h.tokenCompressor.(interface {
		CompressText(context.Context, string, int) (string, error)
	}); ok {
		return tc.CompressText(ctx, text, targetTokens)
	}

	return "", fmt.Errorf("token compressor does not support CompressText")
}

// SetTokenCompressionStrategy 设置压缩策略
func (h *Host) SetTokenCompressionStrategy(strategy string) error {
	if h.tokenCompressor == nil {
		return fmt.Errorf("token compressor not configured")
	}

	if tc, ok := h.tokenCompressor.(interface {
		SetStrategy(token.CompressionStrategy)
	}); ok {
		tc.SetStrategy(token.CompressionStrategy(strategy))
		return nil
	}

	return fmt.Errorf("token compressor does not support SetStrategy")
}

// GetTokenCompressionStats 获取压缩统计信息
func (h *Host) GetTokenCompressionStats() (*token.CompressionStats, error) {
	if h.tokenCompressor == nil {
		return nil, fmt.Errorf("token compressor not configured")
	}

	if tc, ok := h.tokenCompressor.(interface {
		GetStats() *token.CompressionStats
	}); ok {
		return tc.GetStats(), nil
	}

	return nil, fmt.Errorf("token compressor does not support GetStats")
}

// CountMessageTokens 计算消息列表的 Token 数量
func (h *Host) CountMessageTokens(messages []interface{}) int {
	counter := token.NewDefaultTokenCounter()
	return counter.CountMessages(messages)
}

// CountTextTokens 计算文本的 Token 数量
func (h *Host) CountTextTokens(text string) int {
	counter := token.NewDefaultTokenCounter()
	return counter.CountText(text)
}

// ===== TaskManager Getter 方法 =====

// TaskManager 返回任务管理器（可能为 nil）
func (h *Host) TaskManager() interface{} {
	return h.taskManager
}

// SubmitAsyncTask 提交异步任务
func (h *Host) SubmitAsyncTask(ctx context.Context, handler interface{}, opts ...interface{}) (interface{}, error) {
	if h.taskManager == nil {
		return nil, fmt.Errorf("task manager not configured")
	}

	// 将通用函数转换为 TaskFunc
	var taskFn async.TaskFunc
	if hf, ok := handler.(func(context.Context, async.ProgressCallback) (interface{}, error)); ok {
		taskFn = hf
	} else if hf, ok := handler.(func(context.Context) (interface{}, error)); ok {
		taskFn = func(ctx context.Context, progress async.ProgressCallback) (interface{}, error) {
			return hf(ctx)
		}
	} else {
		return nil, fmt.Errorf("invalid handler type")
	}

	// 转换选项
	var asyncOpts []async.TaskOption
	for _, opt := range opts {
		if o, ok := opt.(async.TaskOption); ok {
			asyncOpts = append(asyncOpts, o)
		}
	}

	if tm, ok := h.taskManager.(interface {
		Submit(context.Context, async.TaskFunc, ...async.TaskOption) (async.AsyncTask, error)
	}); ok {
		return tm.Submit(ctx, taskFn, asyncOpts...)
	}

	return nil, fmt.Errorf("task manager does not support Submit")
}

// GetAsyncTask 获取异步任务
func (h *Host) GetAsyncTask(taskID string) (async.AsyncTask, error) {
	if h.taskManager == nil {
		return nil, fmt.Errorf("task manager not configured")
	}

	if tm, ok := h.taskManager.(interface {
		Get(string) (async.AsyncTask, error)
	}); ok {
		return tm.Get(taskID)
	}

	return nil, fmt.Errorf("task manager does not support Get")
}

// ListAsyncTasks 列出异步任务
func (h *Host) ListAsyncTasks(opts ...interface{}) ([]async.AsyncTask, error) {
	if h.taskManager == nil {
		return nil, fmt.Errorf("task manager not configured")
	}

	// 转换选项
	var asyncOpts []async.TaskListOption
	for _, opt := range opts {
		if o, ok := opt.(async.TaskListOption); ok {
			asyncOpts = append(asyncOpts, o)
		}
	}

	if tm, ok := h.taskManager.(interface {
		List(...async.TaskListOption) []async.AsyncTask
	}); ok {
		return tm.List(asyncOpts...)
	}

	return nil, fmt.Errorf("task manager does not support List")
}

// CancelAsyncTask 取消异步任务
func (h *Host) CancelAsyncTask(taskID string) error {
	if h.taskManager == nil {
		return fmt.Errorf("task manager not configured")
	}

	if tm, ok := h.taskManager.(interface {
		Cancel(string) error
	}); ok {
		return tm.Cancel(taskID)
	}

	return fmt.Errorf("task manager does not support Cancel")
}

// WaitAsyncTask 等待异步任务完成
func (h *Host) WaitAsyncTask(ctx context.Context, taskID string) (interface{}, error) {
	if h.taskManager == nil {
		return nil, fmt.Errorf("task manager not configured")
	}

	if tm, ok := h.taskManager.(interface {
		WaitFor(context.Context, string) (interface{}, error)
	}); ok {
		return tm.WaitFor(ctx, taskID)
	}

	return nil, fmt.Errorf("task manager does not support WaitFor")
}

// GetTaskStats 获取任务统计信息
func (h *Host) GetTaskStats() (*async.TaskManagerStats, error) {
	if h.taskManager == nil {
		return nil, fmt.Errorf("task manager not configured")
	}

	if tm, ok := h.taskManager.(interface {
		GetStats() *async.TaskManagerStats
	}); ok {
		return tm.GetStats(), nil
	}

	return nil, fmt.Errorf("task manager does not support GetStats")
}
