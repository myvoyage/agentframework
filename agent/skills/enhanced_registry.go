// Agent Framework - Enhanced Skill Registry
// Copyright (C) 2025 Agent Framework Contributors
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package skills

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v3"
)

// EnhancedSkillRegistry 增强的技能注册表
// 整合 PicoClaw 的依赖检查、Lingti-Bot 的执行器模式
type EnhancedSkillRegistry struct {
	skills            map[string]*EnhancedSkillDefinition
	executors         map[string]ActionExecutor
	dependencyChecker *DependencyChecker
	installer         *SkillInstaller
	watcher           *fsnotify.Watcher
	mu                sync.RWMutex
	baseDir           string
}

// NewEnhancedSkillRegistry 创建增强的技能注册表
func NewEnhancedSkillRegistry(baseDir string) (*EnhancedSkillRegistry, error) {
	if baseDir == "" {
		baseDir = ".skills"
	}

	// 确保目录存在
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("create skills directory failed: %w", err)
	}

	// 创建文件监视器
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("create watcher failed: %w", err)
	}

	registry := &EnhancedSkillRegistry{
		skills:            make(map[string]*EnhancedSkillDefinition),
		executors:         make(map[string]ActionExecutor),
		dependencyChecker: NewDependencyChecker(),
		installer:         NewSkillInstaller(baseDir),
		watcher:           watcher,
		baseDir:           baseDir,
	}

	// 注册默认执行器
	shellExec := NewShellExecutor(30 * time.Second)
	registry.executors["shell"] = shellExec
	registry.executors["sh"] = shellExec // 别名

	httpExec := NewHTTPExecutor()
	registry.executors["http"] = httpExec
	registry.executors["https"] = httpExec // 别名

	// 注册新执行器
	templateExec := NewTemplateExecutor()
	registry.executors["template"] = templateExec

	fileExec := NewFileExecutor(".")
	registry.executors["file"] = fileExec

	jsonExec := NewJSONExecutor()
	registry.executors["json"] = jsonExec

	// EmailExecutor - 需要配置，这里创建默认实例
	// 实际使用时应该通过配置设置 SMTP 服务器信息
	emailExec := NewEmailExecutor("smtp.example.com", 587, "", "", "noreply@example.com")
	registry.executors["email"] = emailExec

	// WorkflowExecutor 需要其他执行器，在最后创建
	executorsMap := map[string]ActionExecutor{
		"shell":    shellExec,
		"sh":       shellExec,
		"http":     httpExec,
		"https":    httpExec,
		"template": templateExec,
		"file":     fileExec,
		"json":     jsonExec,
		"email":    emailExec,
	}
	workflowExec := NewWorkflowExecutor(executorsMap)
	registry.executors["workflow"] = workflowExec

	// 启动文件监视
	go registry.watchFiles()

	return registry, nil
}

// RegisterExecutor 注册执行器
func (r *EnhancedSkillRegistry) RegisterExecutor(execType string, executor ActionExecutor) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.executors[execType] = executor
}

// GetExecutor 获取执行器
func (r *EnhancedSkillRegistry) GetExecutor(execType string) (ActionExecutor, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	exec, ok := r.executors[execType]
	return exec, ok
}

// LoadSkillFromDirectory 从目录加载技能
func (r *EnhancedSkillRegistry) LoadSkillFromDirectory(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read directory failed: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		skillDir := filepath.Join(dir, entry.Name())
		skillFile := filepath.Join(skillDir, "SKILL.md")

		if _, err := os.Stat(skillFile); err != nil {
			continue // 跳过没有 SKILL.md 的目录
		}

		// 加载技能定义
		def, err := r.loadSkillFromFile(skillFile)
		if err != nil {
			fmt.Printf("Warning: failed to load skill %s: %v\n", entry.Name(), err)
			continue
		}

		// 验证技能定义
		if err := def.Validate(); err != nil {
			fmt.Printf("Warning: invalid skill definition %s: %v\n", entry.Name(), err)
			continue
		}

		// 检查依赖
		if def.Prerequisites != nil {
			result, err := r.dependencyChecker.Check(context.Background(), def.Prerequisites)
			if err != nil {
				fmt.Printf("Warning: dependency check failed for %s: %v\n", entry.Name(), err)
				continue
			}

			if !result.Satisfied {
				fmt.Printf("Warning: skill %s prerequisites not satisfied:\n", entry.Name())
				for _, missing := range result.Missing {
					fmt.Printf("  - Missing: %s\n", missing)
				}
				for _, hint := range result.InstallHints {
					fmt.Printf("  - Install: %s\n", hint.Label)
					fmt.Printf("    Command: %s\n", hint.Command)
				}
				continue
			}
		}

		// 注册技能
		r.mu.Lock()
		r.skills[def.ID] = def
		r.mu.Unlock()

		fmt.Printf("Loaded skill: %s (%s)\n", def.Name, def.ID)
	}

	return nil
}

// loadSkillFromFile 从文件加载技能定义
func (r *EnhancedSkillRegistry) loadSkillFromFile(filePath string) (*EnhancedSkillDefinition, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read file failed: %w", err)
	}

	// 分离 frontmatter 和内容
	parts := strings.SplitN(string(data), "---", 3)
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid SKILL.md format")
	}

	// 解析 YAML frontmatter
	var def EnhancedSkillDefinition
	if err := yaml.Unmarshal([]byte(parts[1]), &def); err != nil {
		return nil, fmt.Errorf("parse YAML failed: %w", err)
	}

	// 转换 action 中的 timeout 类型
	for i := range def.Actions {
		// YAML 中 timeout 是 int64（纳秒），需要转换为 time.Duration
		if def.Actions[i].Timeout == 0 {
			// 如果没有设置超时，使用默认值 30 秒
			def.Actions[i].Timeout = 30 * time.Second
		}
		// 注意：如果 YAML 中是整数，YAML 解析器会尝试将其作为 int64 解析
		// 但由于 Timeout 字段类型是 time.Duration，解析会失败
		// 因此需要在 SKILL.md 中使用字符串格式，如 "30s"
	}

	def.SourceFile = filePath
	def.LoadedAt = time.Now()

	return &def, nil
}

// FindByTrigger 根据触发器查找技能
func (r *EnhancedSkillRegistry) FindByTrigger(triggerType, pattern string) ([]*EnhancedSkillDefinition, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var results []*EnhancedSkillDefinition

	for _, skill := range r.skills {
		for _, trigger := range skill.Triggers {
			if trigger.Type == triggerType {
				if pattern == "" || strings.Contains(pattern, trigger.Pattern) {
					results = append(results, skill)
					break // 同一个技能只添加一次
				}
			}
		}
	}

	return results, nil
}

// GetSkill 获取技能
func (r *EnhancedSkillRegistry) GetSkill(skillID string) (*EnhancedSkillDefinition, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	skill, ok := r.skills[skillID]
	return skill, ok
}

// ListSkills 列出所有技能
func (r *EnhancedSkillRegistry) ListSkills() []*EnhancedSkillDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	results := make([]*EnhancedSkillDefinition, 0, len(r.skills))
	for _, skill := range r.skills {
		results = append(results, skill)
	}

	return results
}

// ExecuteSkill 执行技能
func (r *EnhancedSkillRegistry) ExecuteSkill(ctx context.Context, skillID string, actionID string, vars map[string]string) (string, error) {
	r.mu.RLock()
	skill, ok := r.skills[skillID]
	r.mu.RUnlock()

	if !ok {
		return "", fmt.Errorf("skill not found: %s", skillID)
	}

	// 查找动作
	var action *Action
	for _, a := range skill.Actions {
		if a.ID == actionID {
			action = &a
			break
		}
	}

	if action == nil {
		return "", fmt.Errorf("action not found: %s", actionID)
	}

	// 获取执行器
	executor, ok := r.GetExecutor(action.Type)
	if !ok {
		return "", fmt.Errorf("executor not found for type: %s", action.Type)
	}

	// 执行动作
	return executor.Execute(ctx, action, vars)
}

// InstallFromGitHub 从 GitHub 安装技能
func (r *EnhancedSkillRegistry) InstallFromGitHub(ctx context.Context, repo string) error {
	return r.installer.InstallFromGitHub(ctx, repo)
}

// watchFiles 监视文件变化
func (r *EnhancedSkillRegistry) watchFiles() {
	// 监视技能目录
	if err := r.watcher.Add(r.baseDir); err != nil {
		fmt.Printf("Warning: failed to watch skills directory: %v\n", err)
		return
	}

	for {
		select {
		case event, ok := <-r.watcher.Events:
			if !ok {
				return
			}

			// 只处理创建和写入事件
			if event.Op&fsnotify.Create == fsnotify.Create || event.Op&fsnotify.Write == fsnotify.Write {
				// 检查是否是 SKILL.md 文件
				if filepath.Base(event.Name) == "SKILL.md" {
					fmt.Printf("Skill file changed: %s\n", event.Name)
					// 重新加载技能
					skillDir := filepath.Dir(event.Name)
					if err := r.LoadSkillFromDirectory(skillDir); err != nil {
						fmt.Printf("Warning: failed to reload skill: %v\n", err)
					}
				}
			}
		case err, ok := <-r.watcher.Errors:
			if !ok {
				return
			}
			fmt.Printf("Watcher error: %v\n", err)
		}
	}
}

// Close 关闭注册表
func (r *EnhancedSkillRegistry) Close() error {
	return r.watcher.Close()
}
