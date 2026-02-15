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
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// SWEAgent 软件工程智能体，专门用于代码审查、重构、调试和 Git 操作
type SWEAgent struct {
	name           string
	model          ChatModel
	tools          map[string]tool.InvokableTool
	gitClient      *GitClient
	codeAnalyzer   *CodeAnalyzer
	stateMachine   *StateMachine
	memoryManager  *MemoryManager
	thread         *Thread
	repoPath       string
	mu             sync.RWMutex
}

// GitClient Git 客户端
type GitClient struct {
	repoPath string
}

// CodeAnalyzer 代码分析器
type CodeAnalyzer struct {
	model ChatModel
}

// SWEAgentConfig SWEAgent 配置
type SWEAgentConfig struct {
	Name      string
	Model     ChatModel
	Tools     []tool.BaseTool
	RepoPath  string
	MemoryOpts MemoryOptions
}

// NewSWEAgent 创建软件工程智能体
func NewSWEAgent(ctx context.Context, cfg SWEAgentConfig) (*SWEAgent, error) {
	if cfg.Model == nil {
		return nil, fmt.Errorf("model is required")
	}

	toolMap := make(map[string]tool.InvokableTool)
	boundModel := cfg.Model

	if len(cfg.Tools) > 0 {
		var infos []*schema.ToolInfo
		for _, t := range cfg.Tools {
			info, err := t.Info(ctx)
			if err != nil {
				return nil, err
			}
			if info == nil {
				continue
			}
			infos = append(infos, info)

			if inv, ok := t.(tool.InvokableTool); ok {
				toolMap[info.Name] = inv
			}
		}

		if len(infos) > 0 {
			if m, ok := cfg.Model.(model.ToolCallingChatModel); ok {
				tm, err := m.WithTools(infos)
				if err != nil {
					return nil, err
				}
				boundModel = tm
			}
		}
	}

	// 创建 Git 客户端
	gitClient := &GitClient{
		repoPath: cfg.RepoPath,
	}

	// 创建代码分析器
	codeAnalyzer := &CodeAnalyzer{
		model: cfg.Model,
	}

	// 创建状态机
	stateMachine := NewStateMachineWithDefaults()

	// 创建内存管理器
	memoryManager := NewMemoryManager(cfg.MemoryOpts)

	return &SWEAgent{
		name:          cfg.Name,
		model:         boundModel,
		tools:         toolMap,
		gitClient:     gitClient,
		codeAnalyzer:  codeAnalyzer,
		stateMachine:  stateMachine,
		memoryManager: memoryManager,
		thread:        &Thread{ID: cfg.Name},
		repoPath:      cfg.RepoPath,
	}, nil
}

// Name 返回智能体名称
func (a *SWEAgent) Name() string {
	return a.name
}

// Run 执行软件工程任务
func (a *SWEAgent) Run(ctx context.Context, input string, opts ...model.Option) (*schema.Message, error) {
	// 转换到运行状态
	if err := a.stateMachine.Transition(ctx, StateRunning, "Starting SWE task", nil); err != nil {
		return nil, err
	}

	defer func() {
		if a.stateMachine.Current() == StateRunning {
			_ = a.stateMachine.Transition(context.Background(), StateFinished, "Task completed", nil)
		}
	}()

	userMsg := &schema.Message{
		Role:    schema.User,
		Content: input,
	}

	userMsg = a.memoryManager.ProcessMessage(userMsg)
	messages := a.buildMessages(userMsg)

	resp, err := a.model.Generate(ctx, messages, opts...)
	if err != nil {
		_ = a.stateMachine.Transition(ctx, StateError, "Model generation failed", map[string]any{
			"error": err.Error(),
		})
		return nil, err
	}

	// 处理工具调用
	if len(resp.ToolCalls) > 0 && len(a.tools) > 0 {
		toolMsgs, err := a.runTools(ctx, resp)
		if err != nil {
			_ = a.stateMachine.Transition(ctx, StateError, "Tool execution failed", map[string]any{
				"error": err.Error(),
			})
			return nil, err
		}
		messages = append(messages, resp)
		messages = append(messages, toolMsgs...)

		resp, err = a.model.Generate(ctx, messages, opts...)
		if err != nil {
			_ = a.stateMachine.Transition(ctx, StateError, "Model generation after tools failed", map[string]any{
				"error": err.Error(),
			})
			return nil, err
		}
	}

	resp = a.memoryManager.ProcessMessage(resp)
	a.thread.Messages = append(a.thread.Messages, userMsg, resp)
	a.thread.Messages = a.memoryManager.LimitHistory(a.thread.Messages)

	_ = a.stateMachine.Transition(ctx, StateFinished, "Task completed successfully", map[string]any{
		"response_length": len(resp.Content),
	})

	return resp, nil
}

// ReviewCode 审查代码
func (a *SWEAgent) ReviewCode(ctx context.Context, filePath string, content string) (*CodeReviewResult, error) {
	prompt := fmt.Sprintf(`请审查以下代码文件：%s

代码内容：
%s

请提供：
1. 代码质量评估（1-10分）
2. 潜在 Bug 和问题
3. 性能优化建议
4. 安全问题
5. 代码风格建议
6. 重构建议
7. 最佳实践建议

以 JSON 格式返回审查结果。`, filePath, content)

	systemMsg := &schema.Message{
		Role:    schema.System,
		Content: "你是一个专业的代码审查专家，精通多种编程语言、设计模式和最佳实践。你能够发现代码中的问题、性能瓶颈和安全漏洞。",
	}

	userMsg := &schema.Message{
		Role:    schema.User,
		Content: prompt,
	}

	resp, err := a.model.Generate(ctx, []*schema.Message{systemMsg, userMsg})
	if err != nil {
		return nil, err
	}

	var result CodeReviewResult
	if err := json.Unmarshal([]byte(resp.Content), &result); err != nil {
		// 如果解析失败，将文本内容放入摘要
		return &CodeReviewResult{
			Summary: resp.Content,
		}, nil
	}

	return &result, nil
}

// RefactorCode 重构代码
func (a *SWEAgent) RefactorCode(ctx context.Context, filePath string, content string, refactorType string) (*RefactorResult, error) {
	prompt := fmt.Sprintf(`请重构以下代码文件：%s

代码内容：
%s

重构类型：%s

请提供：
1. 重构后的代码
2. 重构说明
3. 改进点说明
4. 潜在影响分析
5. 测试建议

以 JSON 格式返回重构结果。`, filePath, content, refactorType)

	systemMsg := &schema.Message{
		Role:    schema.System,
		Content: "你是一个代码重构专家，精通设计模式、SOLID 原则和代码清洁之道。你能够将复杂的代码重构为简洁、可维护的代码。",
	}

	userMsg := &schema.Message{
		Role:    schema.User,
		Content: prompt,
	}

	resp, err := a.model.Generate(ctx, []*schema.Message{systemMsg, userMsg})
	if err != nil {
		return nil, err
	}

	var result RefactorResult
	if err := json.Unmarshal([]byte(resp.Content), &result); err != nil {
		// 如果解析失败，将文本内容放入说明
		return &RefactorResult{
			Description: resp.Content,
		}, nil
	}

	return &result, nil
}

// DebugCode 调试代码
func (a *SWEAgent) DebugCode(ctx context.Context, filePath string, content string, errorDescription string) (*DebugResult, error) {
	prompt := fmt.Sprintf(`请帮助调试以下代码：

文件路径：%s

代码内容：
%s

错误描述：
%s

请提供：
1. 问题根因分析
2. 修复建议
3. 修复后的代码
4. 预防类似问题的建议
5. 测试验证方法

以 JSON 格式返回调试结果。`, filePath, content, errorDescription)

	systemMsg := &schema.Message{
		Role:    schema.System,
		Content: "你是一个调试专家，能够快速定位代码问题、分析根因并提供有效的修复方案。",
	}

	userMsg := &schema.Message{
		Role:    schema.User,
		Content: prompt,
	}

	resp, err := a.model.Generate(ctx, []*schema.Message{systemMsg, userMsg})
	if err != nil {
		return nil, err
	}

	var result DebugResult
	if err := json.Unmarshal([]byte(resp.Content), &result); err != nil {
		// 如果解析失败，将文本内容放入分析
		return &DebugResult{
			Analysis: resp.Content,
		}, nil
	}

	return &result, nil
}

// AnalyzeCodebase 分析代码库
func (a *SWEAgent) AnalyzeCodebase(ctx context.Context, path string) (*CodebaseAnalysis, error) {
	files, err := a.findCodeFiles(path)
	if err != nil {
		return nil, err
	}

	analysis := &CodebaseAnalysis{
		Path:       path,
		TotalFiles: len(files),
		Languages:  make(map[string]int),
	}

	// 分析文件结构和语言分布
	for _, file := range files {
		ext := strings.TrimPrefix(filepath.Ext(file), ".")
		if ext != "" {
			analysis.Languages[ext]++
		}
	}

	// 生成分析报告
	prompt := fmt.Sprintf(`请分析以下代码库结构：

路径：%s
总文件数：%d
语言分布：%v

请提供：
1. 项目架构分析
2. 代码组织评估
3. 依赖关系分析
4. 技术栈识别
5. 潜在问题识别
6. 改进建议

以 JSON 格式返回分析结果。`, path, len(files), analysis.Languages)

	systemMsg := &schema.Message{
		Role:    schema.System,
		Content: "你是一个软件架构专家，能够分析代码库结构、识别设计模式、评估代码质量并提供架构改进建议。",
	}

	userMsg := &schema.Message{
		Role:    schema.User,
		Content: prompt,
	}

	resp, err := a.model.Generate(ctx, []*schema.Message{systemMsg, userMsg})
	if err != nil {
		return nil, err
	}

	var detail AnalysisDetail
	if err := json.Unmarshal([]byte(resp.Content), &detail); err != nil {
		analysis.Summary = resp.Content
	} else {
		analysis.Detail = &detail
	}

	return analysis, nil
}

// GitOperations Git 操作
func (a *SWEAgent) GitOperations(ctx context.Context, operation string, args map[string]string) (*GitResult, error) {
	result := &GitResult{
		Operation: operation,
	}

	switch operation {
	case "status":
		output, err := a.gitClient.Status(ctx)
		if err != nil {
			result.Error = err.Error()
			return result, err
		}
		result.Output = output

	case "diff":
		output, err := a.gitClient.Diff(ctx, args["file"])
		if err != nil {
			result.Error = err.Error()
			return result, err
		}
		result.Output = output

	case "log":
		output, err := a.gitClient.Log(ctx, args["limit"])
		if err != nil {
			result.Error = err.Error()
			return result, err
		}
		result.Output = output

	case "branch":
		output, err := a.gitClient.Branch(ctx)
		if err != nil {
			result.Error = err.Error()
			return result, err
		}
		result.Output = output

	case "commit":
		message := args["message"]
		if message == "" {
			return result, fmt.Errorf("commit message is required")
		}
		output, err := a.gitClient.Commit(ctx, message, args["files"])
		if err != nil {
			result.Error = err.Error()
			return result, err
		}
		result.Output = output

	case "create_branch":
		branchName := args["name"]
		if branchName == "" {
			return result, fmt.Errorf("branch name is required")
		}
		output, err := a.gitClient.CreateBranch(ctx, branchName)
		if err != nil {
			result.Error = err.Error()
			return result, err
		}
		result.Output = output

	case "checkout":
		branchName := args["branch"]
		if branchName == "" {
			return result, fmt.Errorf("branch name is required")
		}
		output, err := a.gitClient.Checkout(ctx, branchName)
		if err != nil {
			result.Error = err.Error()
			return result, err
		}
		result.Output = output

	default:
		return result, fmt.Errorf("unknown git operation: %s", operation)
	}

	return result, nil
}

// GenerateTests 生成测试
func (a *SWEAgent) GenerateTests(ctx context.Context, filePath string, content string, testFramework string) (*TestResult, error) {
	prompt := fmt.Sprintf(`请为以下代码生成测试：

文件路径：%s
代码内容：
%s
测试框架：%s

请提供：
1. 测试用例设计
2. 测试代码
3. Mock 数据（如需要）
4. 测试说明
5. 覆盖率预期

以 JSON 格式返回测试结果。`, filePath, content, testFramework)

	systemMsg := &schema.Message{
		Role:    schema.System,
		Content: "你是一个测试专家，能够设计全面的测试用例、编写可维护的测试代码并确保高覆盖率。",
	}

	userMsg := &schema.Message{
		Role:    schema.User,
		Content: prompt,
	}

	resp, err := a.model.Generate(ctx, []*schema.Message{systemMsg, userMsg})
	if err != nil {
		return nil, err
	}

	var result TestResult
	if err := json.Unmarshal([]byte(resp.Content), &result); err != nil {
		// 如果解析失败，将文本内容放入说明
		return &TestResult{
			Description: resp.Content,
		}, nil
	}

	return &result, nil
}

// buildMessages 构建消息列表
func (a *SWEAgent) buildMessages(latest *schema.Message) []*schema.Message {
	var msgs []*schema.Message

	// 添加系统提示
	msgs = append(msgs, &schema.Message{
		Role:    schema.System,
		Content: "你是一个专业的软件工程师，擅长代码审查、重构、调试、测试和 Git 操作。你遵循最佳实践、设计模式和 SOLID 原则。",
	})

	// 添加历史消息
	if a.thread != nil && len(a.thread.Messages) > 0 {
		msgs = append(msgs, a.thread.Messages...)
	}

	// 添加最新消息
	msgs = append(msgs, latest)

	return msgs
}

// runTools 执行工具调用
func (a *SWEAgent) runTools(ctx context.Context, toolCallMsg *schema.Message) ([]*schema.Message, error) {
	var toolMessages []*schema.Message

	for _, call := range toolCallMsg.ToolCalls {
		name := call.Function.Name
		t, ok := a.tools[name]
		if !ok {
			continue
		}

		output, err := t.InvokableRun(ctx, call.Function.Arguments)
		if err != nil {
			return nil, err
		}

		toolMsg := &schema.Message{
			Role:       schema.Tool,
			Content:    output,
			ToolCallID: call.ID,
		}

		toolMessages = append(toolMessages, toolMsg)
	}

	return toolMessages, nil
}

// findCodeFiles 查找代码文件
func (a *SWEAgent) findCodeFiles(path string) ([]string, error) {
	var files []string

	err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过隐藏目录和常见忽略目录
		if info.IsDir() {
			name := filepath.Base(filePath)
			if name == ".git" || name == "node_modules" || name == "vendor" || name == ".vscode" {
				return filepath.SkipDir
			}
			return nil
		}

		// 检查文件扩展名
		ext := strings.ToLower(filepath.Ext(filePath))
		codeExtensions := map[string]bool{
			".go":   true,
			".py":   true,
			".js":   true,
			".ts":   true,
			".java": true,
			".c":    true,
			".cpp":  true,
			".h":    true,
			".hpp":  true,
			".rs":   true,
			".rb":   true,
			".php":  true,
			".cs":   true,
			".kt":   true,
			".swift": true,
		}

		if codeExtensions[ext] {
			files = append(files, filePath)
		}

		return nil
	})

	return files, err
}

// ==================== Git Client Methods ====================

// Status 获取 Git 状态
func (g *GitClient) Status(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "status", "--porcelain")
	if g.repoPath != "" {
		cmd.Dir = g.repoPath
	}

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git status failed: %w", err)
	}

	return string(output), nil
}

// Diff 获取文件差异
func (g *GitClient) Diff(ctx context.Context, file string) (string, error) {
	args := []string{"diff"}
	if file != "" {
		args = append(args, file)
	}

	cmd := exec.CommandContext(ctx, "git", args...)
	if g.repoPath != "" {
		cmd.Dir = g.repoPath
	}

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git diff failed: %w", err)
	}

	return string(output), nil
}

// Log 获取提交历史
func (g *GitClient) Log(ctx context.Context, limit string) (string, error) {
	args := []string{"log", "--oneline", "-10"}
	if limit != "" {
		args = []string{"log", "--oneline", "-" + limit}
	}

	cmd := exec.CommandContext(ctx, "git", args...)
	if g.repoPath != "" {
		cmd.Dir = g.repoPath
	}

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git log failed: %w", err)
	}

	return string(output), nil
}

// Branch 获取分支列表
func (g *GitClient) Branch(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "branch", "-a")
	if g.repoPath != "" {
		cmd.Dir = g.repoPath
	}

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git branch failed: %w", err)
	}

	return string(output), nil
}

// Commit 提交更改
func (g *GitClient) Commit(ctx context.Context, message string, files string) (string, error) {
	// 先添加文件
	if files != "" {
		addCmd := exec.CommandContext(ctx, "git", "add", files)
		if g.repoPath != "" {
			addCmd.Dir = g.repoPath
		}
		if _, err := addCmd.Output(); err != nil {
			return "", fmt.Errorf("git add failed: %w", err)
		}
	}

	// 提交
	cmd := exec.CommandContext(ctx, "git", "commit", "-m", message)
	if g.repoPath != "" {
		cmd.Dir = g.repoPath
	}

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git commit failed: %w", err)
	}

	return string(output), nil
}

// CreateBranch 创建新分支
func (g *GitClient) CreateBranch(ctx context.Context, branchName string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "checkout", "-b", branchName)
	if g.repoPath != "" {
		cmd.Dir = g.repoPath
	}

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git create branch failed: %w", err)
	}

	return string(output), nil
}

// Checkout 切换分支
func (g *GitClient) Checkout(ctx context.Context, branchName string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "checkout", branchName)
	if g.repoPath != "" {
		cmd.Dir = g.repoPath
	}

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git checkout failed: %w", err)
	}

	return string(output), nil
}

// ==================== 数据结构定义 ====================

// CodeReviewResult 代码审查结果
type CodeReviewResult struct {
	Summary           string            `json:"summary"`
	QualityScore      int               `json:"quality_score,omitempty"`
	Issues            []CodeIssue       `json:"issues,omitempty"`
	Performance       []string          `json:"performance,omitempty"`
	Security          []string          `json:"security,omitempty"`
	Style             []string          `json:"style,omitempty"`
	Refactoring       []string          `json:"refactoring,omitempty"`
	BestPractices     []string          `json:"best_practices,omitempty"`
}

// CodeIssue 代码问题
type CodeIssue struct {
	Severity    string `json:"severity"` // critical, high, medium, low
	Line        int    `json:"line,omitempty"`
	Column      int    `json:"column,omitempty"`
	Description string `json:"description"`
	Suggestion  string `json:"suggestion,omitempty"`
}

// RefactorResult 重构结果
type RefactorResult struct {
	Description    string   `json:"description"`
	OriginalCode   string   `json:"original_code,omitempty"`
	RefactoredCode string   `json:"refactored_code,omitempty"`
	Improvements   []string `json:"improvements,omitempty"`
	ImpactAnalysis string   `json:"impact_analysis,omitempty"`
	Testing        string   `json:"testing,omitempty"`
}

// DebugResult 调试结果
type DebugResult struct {
	Analysis     string   `json:"analysis"`
	RootCause    string   `json:"root_cause,omitempty"`
	Fix          string   `json:"fix,omitempty"`
	FixedCode    string   `json:"fixed_code,omitempty"`
	Prevention   []string `json:"prevention,omitempty"`
	Verification string   `json:"verification,omitempty"`
}

// CodebaseAnalysis 代码库分析
type CodebaseAnalysis struct {
	Path       string                `json:"path"`
	TotalFiles int                   `json:"total_files"`
	Languages  map[string]int        `json:"languages"`
	Summary    string                `json:"summary,omitempty"`
	Detail     *AnalysisDetail       `json:"detail,omitempty"`
}

// AnalysisDetail 分析详情
type AnalysisDetail struct {
	Architecture    string   `json:"architecture"`
	Organization    string   `json:"organization"`
	Dependencies    string   `json:"dependencies"`
	TechStack       []string `json:"tech_stack"`
	Issues          []string `json:"issues"`
	Recommendations []string `json:"recommendations"`
}

// GitResult Git 操作结果
type GitResult struct {
	Operation string `json:"operation"`
	Output    string `json:"output,omitempty"`
	Error     string `json:"error,omitempty"`
}

// TestResult 测试结果
type TestResult struct {
	Description   string   `json:"description"`
	TestCases     []string `json:"test_cases,omitempty"`
	TestCode      string   `json:"test_code,omitempty"`
	MockData      string   `json:"mock_data,omitempty"`
	Notes         string   `json:"notes,omitempty"`
	Coverage      string   `json:"coverage,omitempty"`
}
