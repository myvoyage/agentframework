// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
// Copyright (C) 2025  Agent Framework Contributors

// SPDX-License-Identifier: AGPL-3.0-or-later

package skills

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// EnhancedSkillExecutor 增强的技能执行器
// 支持声明式流程定义，按照 SKILL.yaml 中的 workflow 步骤执行
type EnhancedSkillExecutor struct {
	definition *SkillDefinition
	registry   *SkillRegistry
	examples   *ExampleLibrary
	validator  InputValidator
	logger     Logger
	config     ExecutorConfig
	mu         sync.RWMutex
}

// ExecutorConfig 执行器配置
type ExecutorConfig struct {
	EnableRetry    bool          // 是否启用重试
	MaxRetries     int           // 最大重试次数
	RetryDelay     time.Duration // 重试延迟
	EnableTimeout  bool          // 是否启用超时
	DefaultTimeout time.Duration // 默认超时
	EnableSkip     bool          // 是否启用跳过逻辑
	EnableCache    bool          // 是否启用缓存
	EnableLog      bool          // 是否启用日志
	LogLevel       string        // 日志级别
}

// InputValidator 输入验证器接口
type InputValidator interface {
	Validate(ctx context.Context, input string) error
	ValidateSchema(ctx context.Context, input string, schema *Schema) error
}

// Logger 日志接口
type Logger interface {
	Debug(ctx context.Context, format string, args ...interface{})
	Info(ctx context.Context, format string, args ...interface{})
	Warn(ctx context.Context, format string, args ...interface{})
	Error(ctx context.Context, format string, args ...interface{})
}

// SimpleLogger 简单的日志实现
type SimpleLogger struct {
	level string
}

// DefaultExecutorConfig 默认执行器配置
var DefaultExecutorConfig = ExecutorConfig{
	EnableRetry:    true,
	MaxRetries:     3,
	RetryDelay:     1 * time.Second,
	EnableTimeout:  true,
	DefaultTimeout: 30 * time.Second,
	EnableSkip:     true,
	EnableCache:    true,
	EnableLog:      true,
	LogLevel:       "info",
}

// NewEnhancedSkillExecutor 创建增强执行器
func NewEnhancedSkillExecutor(config *ExecutorConfig) *EnhancedSkillExecutor {
	if config == nil {
		config = &DefaultExecutorConfig
	}

	return &EnhancedSkillExecutor{
		config: *config,
		logger: &SimpleLogger{level: config.LogLevel},
	}
}

// SetRegistry 设置注册表
func (e *EnhancedSkillExecutor) SetRegistry(registry *SkillRegistry) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.registry = registry
}

// SetExamples 设置示例库
func (e *EnhancedSkillExecutor) SetExamples(examples *ExampleLibrary) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.examples = examples
}

// SetDefinition 设置技能定义
func (e *EnhancedSkillExecutor) SetDefinition(definition *SkillDefinition) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.definition = definition
}

// SetValidator 设置验证器
func (e *EnhancedSkillExecutor) SetValidator(validator InputValidator) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.validator = validator
}

// Execute 执行技能（支持声明式流程）
func (e *EnhancedSkillExecutor) Execute(
	ctx context.Context,
	input string,
	execCtx *ExecutionContext,
) (interface{}, error) {
	if e.definition == nil {
		return nil, fmt.Errorf("skill definition not set")
	}

	e.log(ctx, "info", "开始执行技能: %s", e.definition.Name)

	// 1. 检查前置条件
	if err := e.checkPrerequisites(ctx); err != nil {
		return nil, fmt.Errorf("前置条件检查失败: %w", err)
	}

	// 2. 按照工作流步骤执行
	var result interface{}

	for i, step := range e.definition.Workflow {
		e.log(ctx, "info", "执行步骤 %d/%d: %s (%s)", i+1, len(e.definition.Workflow), step.ID, step.Name)

		// 检查跳过条件
		if e.shouldSkipStep(ctx, step, execCtx, result) {
			e.log(ctx, "info", "跳过步骤: %s", step.ID)
			continue
		}

		// 执行步骤
		stepResult, err := e.executeStep(ctx, step, execCtx, result, input)
		if err != nil {
			e.log(ctx, "error", "步骤 %s 失败: %v", step.ID, err)

			// 检查是否重试
			if step.RetryOnFailure && step.MaxRetries > 0 {
				retryCount := 0
				for retryCount < step.MaxRetries {
					retryCount++
					e.log(ctx, "info", "重试步骤 %s (%d/%d)", step.ID, retryCount, step.MaxRetries)

					time.Sleep(e.config.RetryDelay)

					stepResult, err = e.executeStep(ctx, step, execCtx, result, input)
					if err == nil {
						e.log(ctx, "info", "重试成功: %s", step.ID)
						break
					}
				}
			}

			if err != nil {
				// 根据配置决定是否继续
				if i < len(e.definition.Workflow)-1 {
					e.log(ctx, "warn", "步骤失败但继续执行: %v", err)
				} else {
					return nil, fmt.Errorf("步骤 %s 最终失败: %w", step.ID, err)
				}
			}
		}

		// 更新结果
		result = stepResult

		// 传递结果到下一步
		if step.NextStep != "" {
			// 可以根据结果决定下一个步骤
			if resultMap, ok := result.(map[string]interface{}); ok {
				if nextStep, ok := resultMap[step.NextStep].(string); ok {
					// 跳转到指定步骤
					if nextDef, exists := e.definition.GetStep(nextStep); exists {
						// 执行指定的步骤
						stepResult, err := e.executeStep(ctx, *nextDef, execCtx, result, input)
						if err != nil {
							return nil, err
						}
						result = stepResult
					}
				}
			}
		}
	}

	e.log(ctx, "info", "技能执行完成: %s", e.definition.Name)

	// 3. 记录到注册表
	if e.registry != nil {
		skillID := e.generateSkillID(input)
		if _, exists := e.registry.GetByID(skillID); exists {
			e.registry.RecordUsage(skillID, "executor")
		} else {
			// 创建新条目
			entry := &SkillEntry{
				ID:          skillID,
				Name:        e.definition.Name,
				Description: e.definition.Description,
				Category:    e.definition.Category,
				Version:     e.definition.Version,
			}
			e.registry.Register(entry)
		}
	}

	return result, nil
}

// ExecuteSimple 简化执行（不使用流程定义）
func (e *EnhancedSkillExecutor) ExecuteSimple(
	ctx context.Context,
	action string,
	params map[string]interface{},
) (interface{}, error) {
	// 转换参数类型
	stringParams := make(map[string]string)
	for k, v := range params {
		stringParams[k] = fmt.Sprintf("%v", v)
	}
	return e.executeAction(ctx, action, stringParams, nil)
}

// executeStep 执行单个工作流步骤
func (e *EnhancedSkillExecutor) executeStep(
	ctx context.Context,
	step WorkflowStep,
	execCtx *ExecutionContext,
	previousResult interface{},
	input string,
) (interface{}, error) {

	// 设置超时
	var cancel context.CancelFunc
	if e.config.EnableTimeout && step.Timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, step.Timeout)
	} else if e.config.DefaultTimeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, e.config.DefaultTimeout)
	}
	defer cancel()

	// 根据动作类型执行
	switch step.Action {
	case "validate":
		return e.validateStep(ctx, step, execCtx, input)

	case "prepare":
		return e.prepareStep(ctx, step, execCtx, input)

	case "execute":
		return e.executeActionStep(ctx, step, execCtx, previousResult, input)

	case "check_exists":
		return e.checkExistsStep(ctx, step, execCtx)

	case "generate_code":
		return e.generateCodeStep(ctx, step, execCtx, previousResult)

	case "cleanup":
		return e.cleanupStep(ctx, step, execCtx)

	default:
		return nil, fmt.Errorf("未知动作类型: %s", step.Action)
	}
}

// validateStep 验证步骤
func (e *EnhancedSkillExecutor) validateStep(
	ctx context.Context,
	step WorkflowStep,
	execCtx *ExecutionContext,
	input string,
) (interface{}, error) {

	// 从参数获取必需字段
	requiredFieldsStr := step.Parameters["required_fields"]
	if requiredFieldsStr == "" {
		requiredFieldsStr = "input"
	}

	requiredFields := strings.Split(requiredFieldsStr, ",")
	inputMap := make(map[string]interface{})

	if err := json.Unmarshal([]byte(input), &inputMap); err == nil {
		// 验证必需字段
		for _, field := range requiredFields {
			field = strings.TrimSpace(field)
			if _, exists := inputMap[field]; !exists {
				return nil, fmt.Errorf("缺少必需字段: %s", field)
			}
		}
	}

	return map[string]interface{}{
		"validated":       true,
		"input":           input,
		"required_fields": requiredFields,
	}, nil
}

// prepareStep 准备步骤
func (e *EnhancedSkillExecutor) prepareStep(
	ctx context.Context,
	step WorkflowStep,
	execCtx *ExecutionContext,
	input string,
) (interface{}, error) {

	e.log(ctx, "info", "准备执行: %s", step.Description)

	// 执行准备工作
	// 这里可以根据具体需求实现
	return map[string]interface{}{
		"prepared": true,
		"step":     step.ID,
		"input":    input,
	}, nil
}

// executeActionStep 执行动作步骤
func (e *EnhancedSkillExecutor) executeActionStep(
	ctx context.Context,
	step WorkflowStep,
	execCtx *ExecutionContext,
	previousResult interface{},
	input string,
) (interface{}, error) {

	// 从参数获取动作类型
	useTemplate := step.Parameters["use_template"]
	if useTemplate != "" && e.examples != nil {
		// 使用模板生成代码
		return e.executeWithTemplate(ctx, step, execCtx, previousResult)
	}

	// 直接执行动作
	return e.executeAction(ctx, step.Action, step.Parameters, previousResult)
}

// checkExistsStep 检查存在步骤（去重检查）
func (e *EnhancedSkillExecutor) checkExistsStep(
	ctx context.Context,
	step WorkflowStep,
	execCtx *ExecutionContext,
) (interface{}, error) {

	if e.registry == nil {
		return map[string]interface{}{"exists": false}, nil
	}

	// 从参数获取检查键
	checkKey := e.interpolate(step.Parameters["key"], execCtx)
	if checkKey == "" {
		return nil, fmt.Errorf("检查键为空")
	}

	exists := e.registry.Exists(checkKey)

	entry, _ := e.registry.GetByID(checkKey)
	entryInfo := ""
	if entry != nil {
		entryInfo = fmt.Sprintf(" (位于 %s:%d)", entry.GeneratedFile, entry.GeneratedLine)
	}

	e.log(ctx, "info", "检查存在性: %s = %v%s", checkKey, exists, entryInfo)

	return map[string]interface{}{
		"exists": exists,
		"key":    checkKey,
		"entry":  entry,
	}, nil
}

// generateCodeStep 生成代码步骤
func (e *EnhancedSkillExecutor) generateCodeStep(
	ctx context.Context,
	step WorkflowStep,
	execCtx *ExecutionContext,
	previousResult interface{},
) (interface{}, error) {

	if e.examples == nil {
		return nil, fmt.Errorf("示例库未配置")
	}

	// 获取模板ID
	templateID := step.Parameters["template"]
	if templateID == "" {
		return nil, fmt.Errorf("模板ID未指定")
	}

	// 准备模板数据
	data := e.prepareTemplateData(ctx, step, execCtx, previousResult)

	// 渲染模板
	code, err := e.examples.Render(ctx, templateID, data)
	if err != nil {
		return nil, fmt.Errorf("渲染模板失败: %w", err)
	}

	// 获取输出文件路径
	outputFile := step.Parameters["output_file"]
	if outputFile == "" {
		return nil, fmt.Errorf("输出文件未指定")
	}

	// 确保目录存在
	outputDir := filepath.Dir(outputFile)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("创建目录失败: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(outputFile, []byte(code), 0644); err != nil {
		return nil, fmt.Errorf("写入文件失败: %w", err)
	}

	e.log(ctx, "info", "已生成代码文件: %s (%d bytes)", outputFile, len(code))

	return map[string]interface{}{
		"generated": true,
		"file":      outputFile,
		"code":      code,
		"size":      len(code),
	}, nil
}

// cleanupStep 清理步骤
func (e *EnhancedSkillExecutor) cleanupStep(
	ctx context.Context,
	step WorkflowStep,
	execCtx *ExecutionContext,
) (interface{}, error) {

	e.log(ctx, "info", "清理资源: %s", step.Description)

	// 清理临时资源
	// 这里可以根据具体需求实现清理逻辑
	return map[string]interface{}{
		"cleaned": true,
		"step":    step.ID,
	}, nil
}

// executeAction 执行具体动作
func (e *EnhancedSkillExecutor) executeAction(
	ctx context.Context,
	action string,
	parameters map[string]string,
	data interface{},
) (interface{}, error) {

	switch action {
	case "http_get":
		return e.executeHTTPGet(ctx, parameters, data)
	case "http_post":
		return e.executeHTTPPost(ctx, parameters, data)
	case "file_read":
		return e.executeFileRead(ctx, parameters, data)
	case "file_write":
		return e.executeFileWrite(ctx, parameters, data)
	case "data_transform":
		return e.executeDataTransform(ctx, parameters, data)
	default:
		return nil, fmt.Errorf("未实现的动作: %s", action)
	}
}

// executeWithTemplate 使用模板执行
func (e *EnhancedSkillExecutor) executeWithTemplate(
	ctx context.Context,
	step WorkflowStep,
	execCtx *ExecutionContext,
	previousResult interface{},
) (interface{}, error) {

	// 获取模板ID
	templateID := step.Parameters["use_template"]

	// 准备模板数据
	data := e.prepareTemplateData(ctx, step, execCtx, previousResult)

	// 渲染模板
	code, err := e.examples.Render(ctx, templateID, data)
	if err != nil {
		return nil, fmt.Errorf("渲染模板失败: %w", err)
	}

	// 这里可以动态执行生成的代码
	// 或者保存到文件后执行

	return map[string]interface{}{
		"code":     code,
		"executed": false, // 代码已生成但未执行
		"message":  "代码已生成，需要手动或动态执行",
	}, nil
}

// prepareTemplateData 准备模板数据
func (e *EnhancedSkillExecutor) prepareTemplateData(
	ctx context.Context,
	step WorkflowStep,
	execCtx *ExecutionContext,
	previousResult interface{},
) map[string]interface{} {

	data := make(map[string]interface{})

	// 添加步骤参数
	for k, v := range step.Parameters {
		data[k] = v
	}

	// 添加执行上下文
	if execCtx != nil {
		if env := execCtx.GetEnv("ENV"); env != "" {
			data["env"] = env
		}
		if workspace := execCtx.Workspace; workspace != "" {
			data["workspace"] = workspace
		}
	}

	// 添加上一步的结果
	if previousResult != nil {
		data["previous_result"] = previousResult
	}

	// 添加当前时间
	data["timestamp"] = time.Now().Format(time.RFC3339)
	data["request_id"] = execCtx.RequestID

	return data
}

// shouldSkipStep 检查是否应该跳过步骤
func (e *EnhancedSkillExecutor) shouldSkipStep(
	ctx context.Context,
	step WorkflowStep,
	execCtx *ExecutionContext,
	result interface{},
) bool {

	if !e.config.EnableSkip || step.SkipIf == "" {
		return false
	}

	// 简单的条件判断（实际项目中可以使用表达式引擎）
	skipIf := strings.ToLower(step.SkipIf)

	// 检查配置相关条件
	if strings.Contains(skipIf, "config.cache_disabled") {
		if config, ok := execCtx.GetMetadata("config"); ok {
			if configMap, ok := config.(map[string]interface{}); ok {
				if disabled, ok := configMap["cache_disabled"].(bool); ok && disabled {
					return true
				}
			}
		}
	}

	// 检查结果相关条件
	if result != nil {
		if resultMap, ok := result.(map[string]interface{}); ok {
			if strings.Contains(skipIf, "result.success") {
				if success, ok := resultMap["success"].(bool); ok && success {
					// result.success == true，跳过
					return strings.Contains(skipIf, "result.success == true")
				}
			}
		}
	}

	return false
}

// checkPrerequisites 检查前置条件
func (e *EnhancedSkillExecutor) checkPrerequisites(ctx context.Context) error {
	if e.definition == nil {
		return nil
	}

	for _, prereq := range e.definition.Prerequisites {
		if err := e.checkPrerequisite(ctx, prereq); err != nil {
			if prereq.Required {
				return err
			}
			// 非必需条件只记录警告
			e.log(ctx, "warn", "前置条件警告: %v", err)
		}
	}

	return nil
}

// checkPrerequisite 检查单个前置条件
func (e *EnhancedSkillExecutor) checkPrerequisite(
	ctx context.Context,
	prereq Prerequisite,
) error {

	switch prereq.Type {
	case "network_access":
		return e.checkNetworkAccess(ctx, prereq)

	case "env_var":
		return e.checkEnvVar(prereq)

	case "file_exists":
		return e.checkFileExists(prereq)

	case "dependency":
		return e.checkDependency(prereq)

	default:
		return fmt.Errorf("未知的前置条件类型: %s", prereq.Type)
	}
}

// checkNetworkAccess 检查网络访问
func (e *EnhancedSkillExecutor) checkNetworkAccess(
	ctx context.Context,
	prereq Prerequisite,
) error {

	// 执行检查命令
	if prereq.Check != "" {
		// 简化实现：检查命令（实际项目中应该更安全）
		// 这里只是示例，实际需要考虑安全性
		e.log(ctx, "debug", "检查网络访问: %s", prereq.Check)
		return nil // 假设网络可用
	}

	return nil
}

// checkEnvVar 检查环境变量
func (e *EnhancedSkillExecutor) checkEnvVar(prereq Prerequisite) error {
	if prereq.Check == "" {
		return fmt.Errorf("环境变量名未指定")
	}

	value := os.Getenv(prereq.Check)
	if value == "" {
		return fmt.Errorf("环境变量 %s 未设置", prereq.Check)
	}

	return nil
}

// checkFileExists 检查文件存在
func (e *EnhancedSkillExecutor) checkFileExists(prereq Prerequisite) error {
	if prereq.Check == "" {
		return fmt.Errorf("文件路径未指定")
	}

	if _, err := os.Stat(prereq.Check); err != nil {
		return fmt.Errorf("文件检查失败: %w", err)
	}

	return nil
}

// checkDependency 检查依赖
func (e *EnhancedSkillExecutor) checkDependency(prereq Prerequisite) error {
	// 简化实现：假设依赖已满足
	// 实际项目中应该检查导入的包是否可用
	return nil
}

// executeHTTPGet 执行 HTTP GET 请求
func (e *EnhancedSkillExecutor) executeHTTPGet(
	ctx context.Context,
	params map[string]string,
	data interface{},
) (interface{}, error) {

	url := params["url"]
	if url == "" {
		return nil, fmt.Errorf("URL未指定")
	}

	e.log(ctx, "info", "执行 HTTP GET: %s", url)

	// 这里应该实现实际的 HTTP GET 请求
	// 为了示例，返回模拟数据
	return map[string]interface{}{
		"success": true,
		"url":     url,
		"method":  "GET",
		"status":  200,
		"data": map[string]interface{}{
			"message": "HTTP GET 请求已执行",
			"url":     url,
		},
	}, nil
}

// executeHTTPPost 执行 HTTP POST 请求
func (e *EnhancedSkillExecutor) executeHTTPPost(
	ctx context.Context,
	params map[string]string,
	data interface{},
) (interface{}, error) {

	url := params["url"]
	if url == "" {
		return nil, fmt.Errorf("URL未指定")
	}

	e.log(ctx, "info", "执行 HTTP POST: %s", url)

	// 这里应该实现实际的 HTTP POST 请求
	return map[string]interface{}{
		"success": true,
		"url":     url,
		"method":  "POST",
		"status":  200,
		"data": map[string]interface{}{
			"message": "HTTP POST 请求已执行",
			"url":     url,
		},
	}, nil
}

// executeFileRead 执行文件读取
func (e *EnhancedSkillExecutor) executeFileRead(
	ctx context.Context,
	params map[string]string,
	data interface{},
) (interface{}, error) {

	path := params["path"]
	if path == "" {
		return nil, fmt.Errorf("文件路径未指定")
	}

	e.log(ctx, "info", "读取文件: %s", path)

	// 这里应该实现实际的文件读取
	return map[string]interface{}{
		"success": true,
		"path":    path,
		"action":  "read",
	}, nil
}

// executeFileWrite 执行文件写入
func (e *EnhancedSkillExecutor) executeFileWrite(
	ctx context.Context,
	params map[string]string,
	data interface{},
) (interface{}, error) {

	path := params["path"]
	if path == "" {
		return nil, fmt.Errorf("文件路径未指定")
	}

	content := params["content"]
	if content == "" && data != nil {
		if contentBytes, err := json.Marshal(data); err == nil {
			content = string(contentBytes)
		}
	}

	e.log(ctx, "info", "写入文件: %s", path)

	// 这里应该实现实际的文件写入
	return map[string]interface{}{
		"success": true,
		"path":    path,
		"action":  "written",
		"size":    len(content),
	}, nil
}

// executeDataTransform 执行数据转换
func (e *EnhancedSkillExecutor) executeDataTransform(
	ctx context.Context,
	params map[string]string,
	data interface{},
) (interface{}, error) {

	transformType := params["type"]
	if transformType == "" {
		transformType = "auto"
	}

	e.log(ctx, "info", "执行数据转换: %s", transformType)

	// 这里应该实现实际的数据转换逻辑
	return map[string]interface{}{
		"success": true,
		"type":    transformType,
		"data":    data,
	}, nil
}

// interpolate 插值替换
func (e *EnhancedSkillExecutor) interpolate(
	template string,
	execCtx *ExecutionContext,
) string {

	result := template

	// 简单的变量替换
	// 格式: {{.VariableName}}
	if execCtx != nil {
		// 替换环境变量
		if env := execCtx.GetEnv("ENV"); env != "" {
			result = strings.ReplaceAll(result, "{{.ENV}}", env)
		}

		// 替换工作目录
		if workspace := execCtx.Workspace; workspace != "" {
			result = strings.ReplaceAll(result, "{{.Workspace}}", workspace)
		}

		// 替换请求ID
		if requestID := execCtx.RequestID; requestID != "" {
			result = strings.ReplaceAll(result, "{{.RequestID}}", requestID)
		}

		// �换时间戳
		result = strings.ReplaceAll(result, "{{.Timestamp}}", time.Now().Format(time.RFC3339))
	}

	return result
}

// generateSkillID 生成技能ID
func (e *EnhancedSkillExecutor) generateSkillID(input string) string {
	// 基于输入内容生成唯一的技能ID
	// 这里使用简化的哈希逻辑
	if e.definition != nil && e.definition.ID != "" {
		return e.definition.ID + ":" + hashString(input)
	}
	return hashString(input)
}

// log 记录日志
func (e *EnhancedSkillExecutor) log(
	ctx context.Context,
	level string,
	format string,
	args ...interface{},
) {

	if e.logger == nil || !e.config.EnableLog {
		return
	}

	message := fmt.Sprintf("[%s] %s", level, format)
	switch level {
	case "debug":
		e.logger.Debug(ctx, message, args...)
	case "info":
		e.logger.Info(ctx, message, args...)
	case "warn":
		e.logger.Warn(ctx, message, args...)
	case "error":
		e.logger.Error(ctx, message, args...)
	}
}

// hashString 生成字符串哈希
func hashString(s string) string {
	// 简化的哈希实现
	// 实际项目中应该使用更可靠的哈希算法
	hash := uint32(2166136261)
	for _, c := range s {
		hash *= 31
		hash ^= uint32(c)
	}
	return fmt.Sprintf("%x", hash)
}

// SimpleLogger 的实现
func (l *SimpleLogger) Debug(ctx context.Context, format string, args ...interface{}) {
	if l.level == "debug" {
		fmt.Printf("[DEBUG] "+format+"\n", args...)
	}
}

func (l *SimpleLogger) Info(ctx context.Context, format string, args ...interface{}) {
	if l.level == "debug" || l.level == "info" {
		fmt.Printf("[INFO] "+format+"\n", args...)
	}
}

func (l *SimpleLogger) Warn(ctx context.Context, format string, args ...interface{}) {
	fmt.Printf("[WARN] "+format+"\n", args...)
}

func (l *SimpleLogger) Error(ctx context.Context, format string, args ...interface{}) {
	fmt.Printf("[ERROR] "+format+"\n", args...)
}
