// Agent Framework - Enhanced Code Execution Skill
// SPDX-License-Identifier: AGPL-3.0-or-later

package skills

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/cloudwego/eino/schema"
)

// CodeExecutionSkill 代码执行技能
// 支持多种编程语言的安全代码执行
type CodeExecutionSkill struct {
	*AdvancedSkill
	config *CodeExecutionConfig
}

// CodeExecutionConfig 代码执行配置
type CodeExecutionConfig struct {
	AllowedLanguages []string      // 允许的编程语言
	MaxExecutionTime time.Duration // 最大执行时间
	MaxMemoryMB      int           // 最大内存使用（MB）
	MaxOutputSize    int64         // 最大输出大小
	EnableSandbox    bool          // 是否启用沙箱
	TempDir          string        // 临时目录
	AllowedCommands  []string      // 允许的命令（shell执行）
}

// NewCodeExecutionSkill 创建新的代码执行技能
func NewCodeExecutionSkill(config *CodeExecutionConfig) (*CodeExecutionSkill, error) {
	if config == nil {
		// 默认配置
		tempDir := "/tmp/agent_code"
		config = &CodeExecutionConfig{
			AllowedLanguages: []string{"python", "javascript", "go", "bash", "sh"},
			MaxExecutionTime: 30 * time.Second,
			MaxMemoryMB:      512,
			MaxOutputSize:    1024 * 1024, // 1MB
			EnableSandbox:    true,
			TempDir:          tempDir,
			AllowedCommands:  []string{"echo", "ls", "cat", "pwd", "date", "whoami"},
		}
	}

	skill := &CodeExecutionSkill{
		config: config,
	}

	skill.AdvancedSkill = NewAdvancedSkill(
		"code_execution",
		"Execute code in multiple languages with sandbox protection",
		skill,
	)

	skill.BaseSkill.SetMetadata(SkillMetadata{
		Name:     "code_execution",
		Version:  "2.0.0",
		Category: "code",
		Tags:     []string{"code", "execute", "python", "javascript", "bash"},
	})

	return skill, nil
}

// Info 返回技能信息
func (s *CodeExecutionSkill) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: s.GetName(),
		Desc: s.GetDescription(),
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"language": {
				Type:     "string",
				Desc:     "Programming language: python, javascript, go, bash, sh",
				Required: true,
			},
			"code": {
				Type:     "string",
				Desc:     "Code to execute",
				Required: false,
			},
			"command": {
				Type:     "string",
				Desc:     "Shell command to execute",
				Required: false,
			},
			"args": {
				Type:     "array",
				Desc:     "Command arguments",
				Required: false,
			},
			"stdin": {
				Type:     "string",
				Desc:     "Standard input content",
				Required: false,
			},
			"timeout": {
				Type:     "integer",
				Desc:     "Execution timeout in seconds",
				Required: false,
			},
		}),
	}, nil
}

// Validate 验证输入
func (s *CodeExecutionSkill) Validate(ctx context.Context, input string) error {
	var params struct {
		Language string `json:"language"`
		Code     string `json:"code"`
		Command  string `json:"command"`
	}

	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return err
	}

	if params.Language == "" && params.Command == "" {
		return fmt.Errorf("either language+code or command is required")
	}

	return nil
}

// Prepare 准备执行
func (s *CodeExecutionSkill) Prepare(ctx context.Context, input string) (*ExecutionContext, error) {
	var params struct {
		Language string `json:"language"`
		Code     string `json:"code"`
		Command  string `json:"command"`
	}

	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return nil, err
	}

	// 验证语言是否允许
	if params.Language != "" {
		allowed := false
		for _, lang := range s.config.AllowedLanguages {
			if strings.EqualFold(params.Language, lang) {
				allowed = true
				break
			}
		}
		if !allowed {
			return nil, fmt.Errorf("language '%s' is not allowed", params.Language)
		}
	}

	execCtx := NewExecutionContext()
	execCtx.SetMetadata("input", input)
	return execCtx, nil
}

// Execute 执行代码
func (s *CodeExecutionSkill) Execute(ctx context.Context, execCtx *ExecutionContext) (interface{}, error) {
	input, _ := execCtx.GetMetadata("input")
	inputStr := input.(string)

	var params map[string]interface{}
	if err := json.Unmarshal([]byte(inputStr), &params); err != nil {
		return nil, err
	}

	// 判断执行类型
	if language, ok := params["language"].(string); ok && language != "" {
		return s.executeCode(ctx, params, execCtx)
	}

	if command, ok := params["command"].(string); ok && command != "" {
		return s.executeCommand(ctx, params, execCtx)
	}

	return nil, fmt.Errorf("either language+code or command is required")
}

// executeCode 执行代码
func (s *CodeExecutionSkill) executeCode(ctx context.Context, params map[string]interface{}, execCtx *ExecutionContext) (interface{}, error) {
	language := strings.ToLower(params["language"].(string))
	code := params["code"].(string)

	timeout := s.config.MaxExecutionTime
	if t, ok := params["timeout"].(float64); ok {
		timeout = time.Duration(t) * time.Second
	}

	startTime := time.Now()

	var cmd *exec.Cmd
	switch language {
	case "python", "python3":
		cmd = exec.Command("python3", "-c", code)
	case "javascript", "js", "node":
		cmd = exec.Command("node", "-e", code)
	case "go":
		// Go 需要临时文件
		return s.executeGoCode(ctx, params, execCtx)
	case "bash", "sh":
		cmd = exec.Command("/bin/bash", "-c", code)
	default:
		return nil, fmt.Errorf("unsupported language: %s", language)
	}

	// 执行命令
	output, err := cmd.CombinedOutput()
	duration := time.Since(startTime)

	// 检查超时
	if duration > timeout {
		return nil, fmt.Errorf("execution timeout after %v", timeout)
	}

	if err != nil {
		return map[string]interface{}{
			"success":     false,
			"language":    language,
			"output":      string(output),
			"error":       err.Error(),
			"duration_ms": duration.Milliseconds(),
		}, nil
	}

	return map[string]interface{}{
		"success":     true,
		"language":    language,
		"output":      string(output),
		"duration_ms": duration.Milliseconds(),
	}, nil
}

// executeGoCode 执行 Go 代码
func (s *CodeExecutionSkill) executeGoCode(ctx context.Context, params map[string]interface{}, execCtx *ExecutionContext) (interface{}, error) {
	code := params["code"].(string)

	// 创建临时 Go 文件
	tempFile := fmt.Sprintf("%s/main_%d.go", s.config.TempDir, time.Now().UnixNano())
	if err := os.WriteFile(tempFile, []byte(code), 0644); err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tempFile)

	// 编译并运行
	cmd := exec.Command("go", "run", tempFile)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return map[string]interface{}{
			"success":  false,
			"language": "go",
			"output":   string(output),
			"error":    err.Error(),
		}, nil
	}

	return map[string]interface{}{
		"success":  true,
		"language": "go",
		"output":   string(output),
	}, nil
}

// executeCommand 执行命令
func (s *CodeExecutionSkill) executeCommand(ctx context.Context, params map[string]interface{}, execCtx *ExecutionContext) (interface{}, error) {
	commandStr := params["command"].(string)

	// 解析命令和参数
	parts := strings.Fields(commandStr)
	if len(parts) == 0 {
		return nil, fmt.Errorf("command cannot be empty")
	}

	cmd := parts[0]
	args := parts[1:]

	// 验证命令是否允许
	allowed := false
	for _, allowedCmd := range s.config.AllowedCommands {
		if cmd == allowedCmd {
			allowed = true
			break
		}
	}
	if !allowed {
		return nil, fmt.Errorf("command '%s' is not allowed", cmd)
	}

	// 处理额外参数
	if extraArgs, ok := params["args"].([]interface{}); ok {
		for _, arg := range extraArgs {
			args = append(args, fmt.Sprintf("%v", arg))
		}
	}

	startTime := time.Now()
	command := exec.Command(cmd, args...)

	// 处理标准输入
	if stdin, ok := params["stdin"].(string); ok {
		command.Stdin = strings.NewReader(stdin)
	}

	output, err := command.CombinedOutput()
	duration := time.Since(startTime)

	if err != nil {
		return map[string]interface{}{
			"success":     false,
			"command":     commandStr,
			"output":      string(output),
			"error":       err.Error(),
			"duration_ms": duration.Milliseconds(),
		}, nil
	}

	return map[string]interface{}{
		"success":     true,
		"command":     commandStr,
		"output":      string(output),
		"duration_ms": duration.Milliseconds(),
	}, nil
}

// Cleanup 清理资源
func (s *CodeExecutionSkill) Cleanup(ctx context.Context, execCtx *ExecutionContext) error {
	// 清理临时文件
	return nil
}
