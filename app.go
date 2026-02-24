// Agent Framework - Main Application Entry Point
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"AgentFramework/agent"
	"AgentFramework/api"
	"AgentFramework/core"
)

// App struct wraps the core application for desktop usage
type App struct {
	core       *core.Application
	ctx        context.Context
	apiServer  *api.Server
	apiEnabled bool // 是否启用 API 服务器
}

// NewApp creates a new App application struct
// Refactored to use core.Application (DRY principle)
func NewApp() *App {
	return NewAppWithConfig(false, 0)
}

// NewAppWithConfig creates a new App with custom configuration
// enableApiServer: 是否启用内置 HTTP API 服务器
// apiPort: API 服务器端口（0 表示使用默认端口 8080）
func NewAppWithConfig(enableApiServer bool, apiPort int) *App {
	ctx := context.Background()

	// Initialize OpenTelemetry
	_, err := InitOpenTelemetry(ctx)
	if err != nil {
		fmt.Printf("Warning: Failed to initialize OpenTelemetry: %v\n", err)
	}

	// 创建默认的HostConfig
	defaultHostConfig := &agent.HostConfig{
		Models: map[string]agent.ModelConfig{
			"default": {
				Type:  "ollama",
				Model: "llama3",
			},
		},
		SkillSystemDir: ".skills", // 启用技能系统
	}

	// 创建模型工厂
	modelFactory := agent.NewModelFactoryWithConfig(agent.ModelConfig{
		Type:  "ollama",
		Model: "llama3",
	})

	// Create core application (shared between CLI and desktop)
	coreApp, err := core.NewApplication(ctx, defaultHostConfig, modelFactory, nil)
	if err != nil {
		panic(fmt.Errorf("failed to create core application: %w", err))
	}

	// Initialize core application
	if err := coreApp.Initialize(ctx); err != nil {
		panic(fmt.Errorf("failed to initialize core application: %w", err))
	}

	app := &App{
		core:       coreApp,
		ctx:        ctx,
		apiEnabled: enableApiServer,
	}

	// Create API server if enabled
	if enableApiServer {
		if apiPort == 0 {
			apiPort = 8080
		}
		app.apiServer = api.NewServer(ctx, coreApp, api.ServerConfig{
			Host: "localhost",
			Port: apiPort,
		})
	}

	return app
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	// 初始化工作流管理器
	a.core.GetWorkflowManager().Init(ctx)
	// 初始化文件浏览器
	a.core.GetFileExplorer().Init(ctx)

	// 启动 API 服务器（如果启用）
	if a.apiEnabled && a.apiServer != nil {
		go func() {
			if err := a.apiServer.Start(); err != nil {
				fmt.Printf("Error starting API server: %v\n", err)
			}
		}()
	}
}

// shutdown is called when the app is closing
func (a *App) shutdown(ctx context.Context) {
	// 停止 API 服务器
	if a.apiServer != nil {
		if err := a.apiServer.Shutdown(ctx); err != nil {
			fmt.Printf("Error shutting down API server: %v\n", err)
		}
	}
}

// getSkillSystemBaseDir returns the base directory for the skill system
func (a *App) getSkillSystemBaseDir() string {
	if a.core.GetHost() != nil {
		cfg := a.core.GetHost().Config()
		if cfg != nil && cfg.SkillSystemDir != "" {
			return cfg.SkillSystemDir
		}
	}
	return ".skills" // 默认值
}

// ========== Wails Bindings - 现有方法 ==========

// ClearCache 清空缓存
func (a *App) ClearCache() error {
	// TODO: 实现缓存清理逻辑
	return nil
}

// File operations
func (a *App) CopyFile(src, dst string) error {
	return a.core.GetFileExplorer().CopyFile(a.ctx, src, dst)
}

func (a *App) CreateDirectory(path string) error {
	return os.MkdirAll(path, 0755)
}

func (a *App) CreateFile(path, content string) error {
	return a.core.GetFileExplorer().WriteFile(a.ctx, path, content)
}

func (a *App) DeleteDirectory(path string) error {
	return os.RemoveAll(path)
}

func (a *App) DeleteFile(path string) error {
	return a.core.GetFileExplorer().DeleteFile(a.ctx, path)
}

func (a *App) DownloadFile(path string) error {
	// TODO: 实现文件下载逻辑
	return nil
}

// Workflow operations
func (a *App) CreateWorkflow(name, description, definition string) (string, error) {
	return a.core.GetWorkflowManager().CreateWorkflow(a.ctx, name, description, definition)
}

func (a *App) DeleteWorkflow(id string) error {
	return a.core.GetWorkflowManager().DeleteWorkflow(a.ctx, id)
}

func (a *App) ExecuteWorkflow(id, input string) (string, error) {
	return a.core.GetWorkflowManager().ExecuteWorkflow(a.ctx, id, input)
}

func (a *App) GetWorkflows() ([]*agent.WorkflowInfo, error) {
	return a.core.GetWorkflowManager().GetWorkflows(a.ctx)
}

func (a *App) GetWorkflow(id string) (*agent.WorkflowInfo, error) {
	return a.core.GetWorkflowManager().GetWorkflow(a.ctx, id)
}

func (a *App) UpdateWorkflow(id, name, description, definition string) error {
	return a.core.GetWorkflowManager().UpdateWorkflow(a.ctx, id, name, description, definition)
}

// Skill operations
func (a *App) DeleteSkill(id string) error {
	return a.core.GetSkillLibrary().UnregisterSkill(a.ctx, id)
}

func (a *App) DisableSkill(id string) error {
	return a.core.GetSkillLibrary().DisableSkill(a.ctx, id)
}

func (a *App) EnableSkill(id string) error {
	return a.core.GetSkillLibrary().EnableSkill(a.ctx, id)
}

// Config operations
func (a *App) GetConfig() (*agent.HostConfig, error) {
	return a.core.GetHost().Config(), nil
}

func (a *App) UpdateConfig(config map[string]interface{}) error {
	// TODO: 实现配置更新逻辑
	return nil
}

// ========== 新增的 Wails Bindings ==========

// ListFiles 列出目录文件
func (a *App) ListFiles(path string, depth int) ([]*agent.FileInfo, error) {
	return a.core.GetFileExplorer().ListFiles(a.ctx, path)
}

// ReadFile 读取文件内容
func (a *App) ReadFile(path string) (string, error) {
	return a.core.GetFileExplorer().ReadFile(a.ctx, path)
}

// ListAgents 列出所有可用的 agents
func (a *App) ListAgents() ([]string, error) {
	return a.core.GetHost().ListAgents(), nil
}

// GetAgentInfo 获取 agent 信息
func (a *App) GetAgentInfo(id string) (map[string]interface{}, error) {
	agent, err := a.core.GetHost().GetAgent(id)
	if err != nil {
		return nil, fmt.Errorf("agent %s not found: %w", id, err)
	}
	if agent == nil {
		return nil, fmt.Errorf("agent %s not found", id)
	}

	return map[string]interface{}{
		"id":   id,
		"name":  agent.Name(),
		"type":  fmt.Sprintf("%T", agent),
	}, nil
}

// ChatWithAgent 与指定 agent 进行对话
func (a *App) ChatWithAgent(agentID, message string) (string, error) {
	agent, err := a.core.GetHost().GetAgent(agentID)
	if err != nil {
		return "", fmt.Errorf("agent %s not found: %w", agentID, err)
	}
	if agent == nil {
		return "", fmt.Errorf("agent %s not found", agentID)
	}

	response, err := agent.Run(a.ctx, message)
	if err != nil {
		return "", err
	}

	// Convert response Message to string
	return response.Content, nil
}

// ListSkills 列出所有技能
func (a *App) ListSkills() ([]*agent.SkillMetadata, error) {
	skills := a.core.GetSkillLibrary().GetAllSkills(a.ctx)
	infos := make([]*agent.SkillMetadata, 0, len(skills))

	for _, skill := range skills {
		metadata := skill.GetMetadata(a.ctx)
		infos = append(infos, &metadata)
	}

	return infos, nil
}

// GetSkillInfo 获取技能信息
func (a *App) GetSkillInfo(id string) (*agent.SkillMetadata, error) {
	skill, found := a.core.GetSkillLibrary().GetSkill(a.ctx, id)
	if !found {
		return nil, fmt.Errorf("skill %s not found", id)
	}

	metadata := skill.GetMetadata(a.ctx)
	return &metadata, nil
}

// GetWorkflowVersions 获取工作流版本
func (a *App) GetWorkflowVersions(workflowID string) ([]*agent.WorkflowVersion, error) {
	return a.core.GetWorkflowManager().GetWorkflowVersions(a.ctx, workflowID)
}

// GetWorkflowExecutionResult 获取工作流执行结果
func (a *App) GetWorkflowExecutionResult(executionID string) (*agent.WorkflowExecutionResult, error) {
	return a.core.GetWorkflowManager().GetWorkflowExecutionResult(a.ctx, executionID)
}

// RegisterSkillFromMap 从 map 注册技能
func (a *App) RegisterSkillFromMap(skillData map[string]interface{}) (string, error) {
	id, _ := skillData["id"].(string)
	name, _ := skillData["name"].(string)
	if name == "" {
		name = id
	}
	_, _ = skillData["description"].(string)

	// TODO: 实现从 map 创建动态技能的逻辑
	// 目前返回一个占位符错误
	// 未来可以使用 agent.skills.NewAdvancedSkill 创建自定义技能
	return "", fmt.Errorf("dynamic skill registration from map is not yet implemented")
}

// ImportWorkflowFromJSON 从 JSON 导入工作流
func (a *App) ImportWorkflowFromJSON(workflowJSON string) (string, error) {
	var workflowDef struct {
		Name        string                 `json:"name"`
		Description string                 `json:"description"`
		Type        string                 `json:"type"`
		Definition  string                 `json:"definition"`
		Metadata    map[string]interface{} `json:"metadata"`
	}

	if err := json.Unmarshal([]byte(workflowJSON), &workflowDef); err != nil {
		return "", fmt.Errorf("failed to parse workflow JSON: %w", err)
	}

	return a.core.GetWorkflowManager().CreateWorkflow(
		a.ctx,
		workflowDef.Name,
		workflowDef.Description,
		workflowDef.Definition,
	)
}
