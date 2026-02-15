// Agent Framework - Modular Skills Extension
// Copyright (C) 2025 Agent Framework Contributors
// SPDX-License-Identifier: Apache-2.0

package skills

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"plugin"
	"regexp"
	"runtime"
	"sync"
	"time"

	"github.com/cloudwego/eino/schema"
)

// ModularSkillRegistry 模块化技能注册表
type ModularSkillRegistry struct {
	registry    *SkillRegistry
	plugins     map[string]*plugin.Plugin
	skills      map[string]*ModularSkill
	config      ModularSkillConfig
	mu          sync.RWMutex
	loader      *SkillLoader
	monitor     *SkillMonitor
}

// ModularSkillConfig 模块化技能配置
type ModularSkillConfig struct {
	SkillDirs         []string `json:"skill_dirs"`           // 技能目录列表
	EnableHotReload   bool     `json:"enable_hot_reload"`     // 启用热重载
	ReloadInterval    int      `json:"reload_interval"`       // 重载检查间隔（秒）
	EnableVersioning  bool     `json:"enable_versioning"`     // 启用版本管理
	EnableDependency  bool     `json:"enable_dependency"`     // 启用依赖管理
	EnableMonitoring  bool     `json:"enable_monitoring"`     // 启用性能监控
	MaxConcurrentLoad int      `json:"max_concurrent_load"`   // 最大并发加载数
	CacheEnabled     bool     `json:"cache_enabled"`        // 启用技能缓存
}

// ModularSkill 模块化技能
type ModularSkill struct {
	*SkillEntry
	PluginPath   string            `json:"plugin_path"`    // 插件路径
	Dependencies []SkillDependency `json:"dependencies"`  // 依赖项
	Version      SkillVersion       `json:"version"`        // 版本信息
	LoadTime     time.Time         `json:"load_time"`      // 加载时间
	LastUsed     time.Time         `json:"last_used"`      // 最后使用时间
	Health       SkillHealth       `json:"health"`         // 健康状态
}

// SkillVersion 技能版本信息
type SkillVersion struct {
	Major      int    `json:"major"`       // 主版本号
	Minor      int    `json:"minor"`       // 次版本号
	Patch      int    `json:"patch"`       // 补丁版本号
	PreRelease string `json:"pre_release"` // 预发布版本
	BuildMeta  string `json:"build_meta"`  // 构建元数据
}

// SkillDependency 技能依赖项
type SkillDependency struct {
	ID          string `json:"id"`           // 依赖的技能ID
	MinVersion  string `json:"min_version"`  // 最低版本
	MaxVersion  string `json:"max_version"`  // 最高版本
	Optional    bool   `json:"optional"`     // 是否可选依赖
}

// SkillHealth 技能健康状态
type SkillHealth struct {
	Status      string    `json:"status"`       // healthy, degraded, unhealthy
	LastError   string    `json:"last_error"`   // 最后错误信息
	ErrorCount  int       `json:"error_count"`  // 错误计数
	LastCheck   time.Time `json:"last_check"`   // 最后检查时间
	Uptime      time.Duration `json:"uptime"`      // 运行时长
}

// SkillLoader 技能加载器
type SkillLoader struct {
	config      ModularSkillConfig
	loadSem     chan struct{}
	loading     map[string]bool
	mu          sync.RWMutex
}

// SkillMonitor 技能监控器
type SkillMonitor struct {
	registry    *ModularSkillRegistry
	metrics     map[string]*SkillMetrics
	mu          sync.RWMutex
}

// SkillMetrics 技能指标
type SkillMetrics struct {
	TotalInvocations   int64         `json:"total_invocations"`
	SuccessCount      int64         `json:"success_count"`
	FailureCount      int64         `json:"failure_count"`
	AvgLatency        time.Duration `json:"avg_latency"`
	MaxLatency        time.Duration `json:"max_latency"`
	MinLatency        time.Duration `json:"min_latency"`
	LastInvocation    time.Time      `json:"last_invocation"`
	ErrorRate         float64       `json:"error_rate"`
}

// NewModularSkillRegistry 创建模块化技能注册表
func NewModularSkillRegistry(config ModularSkillConfig) (*ModularSkillRegistry, error) {
	if len(config.SkillDirs) == 0 {
		config.SkillDirs = []string{"./skills"}
	}
	if config.ReloadInterval <= 0 {
		config.ReloadInterval = 60
	}
	if config.MaxConcurrentLoad <= 0 {
		config.MaxConcurrentLoad = 10
	}

	// 创建基础注册表
	baseRegistry := NewSkillRegistry()

	loader := &SkillLoader{
		config:  config,
		loadSem: make(chan struct{}, config.MaxConcurrentLoad),
		loading: make(map[string]bool),
	}

	monitor := &SkillMonitor{
		metrics: make(map[string]*SkillMetrics),
	}

	registry := &ModularSkillRegistry{
		registry: baseRegistry,
		plugins:  make(map[string]*plugin.Plugin),
		skills:  make(map[string]*ModularSkill),
		config:  config,
		loader:  loader,
		monitor: monitor,
	}
	monitor.registry = registry

	// 加载技能
	if err := registry.LoadAllSkills(); err != nil {
		return nil, fmt.Errorf("failed to load skills: %w", err)
	}

	// 启动热重载
	if config.EnableHotReload {
		go registry.startHotReload()
	}

	// 启动监控
	if config.EnableMonitoring {
		go registry.startMonitoring()
	}

	return registry, nil
}

// LoadAllSkills 加载所有技能
func (r *ModularSkillRegistry) LoadAllSkills() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var errors []error

	for _, skillDir := range r.config.SkillDirs {
		// 扫描技能目录
		skills, err := r.scanSkillsDirectory(skillDir)
		if err != nil {
			errors = append(errors, err)
			continue
		}

		// 加载每个技能
		for _, skillPath := range skills {
			if err := r.loadSkill(skillPath); err != nil {
				errors = append(errors, err)
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors loading skills: %v", errors)
	}

	return nil
}

// scanSkillsDirectory 扫描技能目录
func (r *ModularSkillRegistry) scanSkillsDirectory(dir string) ([]string, error) {
	var skills []string

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read skill directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			skillPath := filepath.Join(dir, entry.Name())
			// 检查是否为有效技能目录（包含 skill.json）
			if _, err := os.Stat(filepath.Join(skillPath, "skill.json")); err == nil {
				skills = append(skills, skillPath)
			}
		} else if filepath.Ext(entry.Name()) == ".so" {
			// 插件文件
			skills = append(skills, filepath.Join(dir, entry.Name()))
		}
	}

	return skills, nil
}

// loadSkill 加载技能
func (r *ModularSkillRegistry) loadSkill(path string) error {
	// 获取加载信号（控制并发）
	r.loader.mu.Lock()
	if r.loader.loading[path] {
		r.loader.mu.Unlock()
		return fmt.Errorf("skill %s is already loading", path)
	}
	r.loader.loading[path] = true
	r.loader.mu.Unlock()

	defer func() {
		r.loader.mu.Lock()
		delete(r.loader.loading, path)
		r.loader.mu.Unlock()
	}()

	// 读取技能元数据
	metadataPath := filepath.Join(path, "skill.json")
	metadata, err := r.readSkillMetadata(metadataPath)
	if err != nil {
		return fmt.Errorf("failed to read skill metadata: %w", err)
	}

	// 版本检查
	if r.config.EnableVersioning {
		if err := r.checkVersion(metadata); err != nil {
			return fmt.Errorf("version check failed: %w", err)
		}
	}

	// 依赖检查
	if r.config.EnableDependency {
		if err := r.checkDependencies(metadata); err != nil {
			return fmt.Errorf("dependency check failed: %w", err)
		}
	}

	// 加载插件（如果是 .so 文件）
	if filepath.Ext(path) == ".so" {
		plug, err := plugin.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open plugin: %w", err)
		}

		symbol, err := plug.Lookup("Skill")
		if err != nil {
			plug.Close()
			return fmt.Errorf("plugin does not export Skill symbol: %w", err)
		}

		skill, ok := symbol.(Skill)
		if !ok {
			plug.Close()
			return fmt.Errorf("plugin Skill symbol is not a Skill implementation")
		}

		r.plugins[metadata.ID] = plug

		// 创建技能条目
		entry := &SkillEntry{
			ID:          metadata.ID,
			Name:        metadata.Name,
			Description: metadata.Description,
			Category:    metadata.Category,
			Tags:        metadata.Tags,
			Version:     metadataVersionToString(metadata.Version),
			Enabled:     metadata.Enabled,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		skillEntry := &ModularSkill{
			SkillEntry: entry,
			PluginPath:  path,
			Dependencies: metadata.Dependencies,
			Version:     metadata.Version,
			LoadTime:    time.Now(),
			Health: SkillHealth{
				Status:    "healthy",
				LastCheck: time.Now(),
			},
		}

		r.skills[metadata.ID] = skillEntry
		r.registry.skills[metadata.ID] = entry

		return nil
	}

	// 动态加载 Go 代码
	// 这里需要实现 Go 代码动态编译和加载
	// 由于 Go 不支持原生动态加载，可以使用以下方案：
	// 1. 使用 gorgonia/eval 等解释器
	// 2. 使用 yaegi 等嵌入式脚本语言
	// 3. 使用 Go plugin 模式（需要预编译）

	return fmt.Errorf("dynamic Go skill loading not implemented")
}

// readSkillMetadata 读取技能元数据
func (r *ModularSkillRegistry) readSkillMetadata(path string) (*ModularSkillMetadata, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata file: %w", err)
	}

	var metadata ModularSkillMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse metadata: %w", err)
	}

	return &metadata, nil
}

// ModularSkillMetadata 模块化技能元数据
type ModularSkillMetadata struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Category    string            `json:"category"`
	Tags        []string          `json:"tags"`
	Version     SkillVersion      `json:"version"`
	Enabled     bool              `json:"enabled"`
	Dependencies []SkillDependency `json:"dependencies"`
	Config      map[string]string `json:"config"`
	Metadata    map[string]string `json:"metadata"`
}

// checkVersion 检查版本兼容性
func (r *ModularSkillRegistry) checkVersion(metadata *ModularSkillMetadata) error {
	// 检查框架版本兼容性
	// 这里可以实现更复杂的版本比较逻辑

	return nil
}

// checkDependencies 检查依赖项
func (r *ModularSkillRegistry) checkDependencies(metadata *ModularSkillMetadata) error {
	for _, dep := range metadata.Dependencies {
		// 检查依赖的技能是否存在
		r.mu.RLock()
		depSkill, exists := r.skills[dep.ID]
		r.mu.RUnlock()

		if !exists && !dep.Optional {
			return fmt.Errorf("required dependency %s not found", dep.ID)
		}

		if exists {
			// 检查版本兼容性
			if err := r.checkDependencyVersion(depSkill.Version, dep); err != nil {
				return fmt.Errorf("dependency %s version check failed: %w", dep.ID, err)
			}
		}
	}

	return nil
}

// checkDependencyVersion 检查依赖版本兼容性
func (r *ModularSkillRegistry) checkDependencyVersion(currentVersion string, dep SkillDependency) error {
	// 实现版本范围检查
	// 支持 semver 格式比较

	return nil
}

// InvokeSkill 调用技能
func (r *ModularSkillRegistry) InvokeSkill(ctx context.Context, skillID, input string) (string, error) {
	r.mu.RLock()
	skill, exists := r.skills[skillID]
	r.mu.RUnlock()

	if !exists {
		return "", fmt.Errorf("skill not found: %s", skillID)
	}

	if !skill.Enabled {
		return "", fmt.Errorf("skill is disabled: %s", skillID)
	}

	// 检查健康状态
	if skill.Health.Status == "unhealthy" {
		return "", fmt.Errorf("skill is unhealthy: %s", skillID)
	}

	// 更新监控指标
	if r.config.EnableMonitoring {
		r.updateMetrics(skillID, true, 0, nil)
	}

	// 执行技能
	startTime := time.Now()

	// 这里需要实际调用技能实现
	// 由于技能可能是插件或 Go 代码，需要统一接口

	duration := time.Since(startTime)

	// 更新监控指标
	if r.config.EnableMonitoring {
		r.updateMetrics(skillID, false, duration, nil)
	}

	// 更新使用统计
	skill.LastUsed = time.Now()
	skill.UsedCount++

	return "", nil
}

// updateMetrics 更新技能指标
func (r *ModularSkillRegistry) updateMetrics(skillID string, completed bool, duration time.Duration, err error) {
	r.monitor.mu.Lock()
	defer r.monitor.mu.Unlock()

	metrics, exists := r.monitor.metrics[skillID]
	if !exists {
		metrics = &SkillMetrics{
			MinLatency: duration,
			MaxLatency: duration,
		}
		r.monitor.metrics[skillID] = metrics
	}

	metrics.TotalInvocations++
	metrics.LastInvocation = time.Now()

	if completed {
		if err != nil {
			metrics.FailureCount++
		} else {
			metrics.SuccessCount++
		}

		// 更新延迟统计
		if duration > metrics.MaxLatency {
			metrics.MaxLatency = duration
		}
		if duration < metrics.MinLatency {
			metrics.MinLatency = duration
		}

		// 计算平均延迟
		total := metrics.SuccessCount + metrics.FailureCount
		if total > 0 {
			sum := time.Duration(0)
			// 这里应该维护一个累积和，简化处理
			metrics.AvgLatency = sum / time.Duration(total)
		}

		// 计算错误率
		metrics.ErrorRate = float64(metrics.FailureCount) / float64(total)
	}
}

// GetSkillMetrics 获取技能指标
func (r *ModularSkillRegistry) GetSkillMetrics(skillID string) (*SkillMetrics, error) {
	if !r.config.EnableMonitoring {
		return nil, fmt.Errorf("monitoring is disabled")
	}

	r.monitor.mu.RLock()
	defer r.monitor.mu.RUnlock()

	metrics, exists := r.monitor.metrics[skillID]
	if !exists {
		return nil, fmt.Errorf("skill not found: %s", skillID)
	}

	return metrics, nil
}

// GetAllMetrics 获取所有技能指标
func (r *ModularSkillRegistry) GetAllMetrics() map[string]*SkillMetrics {
	if !r.config.EnableMonitoring {
		return nil
	}

	r.monitor.mu.RLock()
	defer r.monitor.mu.RUnlock()

	// 返回副本
	result := make(map[string]*SkillMetrics)
	for id, metrics := range r.monitor.metrics {
		result[id] = metrics
	}

	return result
}

// startHotReload 启动热重载
func (r *ModularSkillRegistry) startHotReload() {
	ticker := time.NewTicker(time.Duration(r.config.ReloadInterval) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// 检查技能目录变化
		for _, skillDir := range r.config.SkillDirs {
			if err := r.checkAndReloadSkills(skillDir); err != nil {
				// 记录错误但不中断
				continue
			}
		}
	}
}

// checkAndReloadSkills 检查并重载技能
func (r *ModularSkillRegistry) checkAndReloadSkills(dir string) error {
	// 扫描变化
	// 这里需要实现文件变化检测和重载逻辑

	return nil
}

// startMonitoring 启动监控
func (r *ModularSkillRegistry) startMonitoring() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// 健康检查
		r.healthCheck()
	}
}

// healthCheck 健康检查
func (r *ModularSkillRegistry) healthCheck() {
	r.mu.Lock()
	defer r.mu.Unlock()

	for id, skill := range r.skills {
		// 检查错误率
		if metrics, exists := r.monitor.metrics[id]; exists {
			if metrics.ErrorRate > 0.5 { // 错误率超过50%
				skill.Health.Status = "unhealthy"
				skill.Health.ErrorCount++
			} else if metrics.ErrorRate > 0.2 { // 错误率超过20%
				skill.Health.Status = "degraded"
			} else {
				skill.Health.Status = "healthy"
			}
		}

		skill.Health.LastCheck = time.Now()
		skill.Health.Uptime = time.Since(skill.LoadTime)
	}
}

// UnloadSkill 卸载技能
func (r *ModularSkillRegistry) UnloadSkill(skillID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	skill, exists := r.skills[skillID]
	if !exists {
		return fmt.Errorf("skill not found: %s", skillID)
	}

	// 关闭插件
	if plug, ok := r.plugins[skillID]; ok {
		if err := plug.Close(); err != nil {
			return fmt.Errorf("failed to close plugin: %w", err)
		}
		delete(r.plugins, skillID)
	}

	// 从注册表中移除
	delete(r.skills, skillID)
	delete(r.registry.skills, skillID)

	return nil
}

// ReloadSkill 重载技能
func (r *ModularSkillRegistry) ReloadSkill(skillID string) error {
	r.mu.Lock()
	skill, exists := r.skills[skillID]
	r.mu.Unlock()

	if !exists {
		return fmt.Errorf("skill not found: %s", skillID)
	}

	// 卸载旧版本
	if err := r.UnloadSkill(skillID); err != nil {
		return fmt.Errorf("failed to unload skill: %w", err)
	}

	// 加载新版本
	if err := r.loadSkill(skill.PluginPath); err != nil {
		return fmt.Errorf("failed to reload skill: %w", err)
	}

	return nil
}

// ListSkills 列出所有技能
func (r *ModularSkillRegistry) ListSkills() []*ModularSkill {
	r.mu.RLock()
	defer r.mu.RUnlock()

	skills := make([]*ModularSkill, 0, len(r.skills))
	for _, skill := range r.skills {
		skills = append(skills, skill)
	}

	return skills
}

// GetSkill 获取技能
func (r *ModularSkillRegistry) GetSkill(skillID string) (*ModularSkill, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	skill, exists := r.skills[skillID]
	if !exists {
		return nil, fmt.Errorf("skill not found: %s", skillID)
	}

	return skill, nil
}

// Close 关闭注册表
func (r *ModularSkillRegistry) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 关闭所有插件
	var errors []error
	for id, plug := range r.plugins {
		if err := plug.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close plugin %s: %w", id, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors closing plugins: %v", errors)
	}

	return nil
}

// ==================== 辅助函数 ====================

// skillVersionToString 将版本转换为字符串
func skillVersionToString(v SkillVersion) string {
	result := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	if v.PreRelease != "" {
		result += "-" + v.PreRelease
	}
	if v.BuildMeta != "" {
		result += "+" + v.BuildMeta
	}
	return result
}

// parseSkillVersion 解析版本字符串
func parseSkillVersion(s string) (SkillVersion, error) {
	// 简化的版本解析
	var major, minor, patch int
	n, err := fmt.Sscanf(s, "%d.%d.%d", &major, &minor, &patch)
	if err != nil {
		return SkillVersion{}, fmt.Errorf("invalid version format: %s", s)
	}

	if n != 3 {
		return SkillVersion{}, fmt.Errorf("invalid version format: %s", s)
	}

	return SkillVersion{
		Major: major,
		Minor: minor,
		Patch: patch,
	}, nil
}

// compareVersions 比较两个版本
// 返回：-1 表示 v1 < v2，0 表示 v1 == v2，1 表示 v1 > v2
func compareVersions(v1, v2 string) int {
	ver1, err := parseSkillVersion(v1)
	if err != nil {
		return 0
	}

	ver2, err := parseSkillVersion(v2)
	if err != nil {
		return 0
	}

	if ver1.Major != ver2.Major {
		if ver1.Major < ver2.Major {
			return -1
		}
		return 1
	}

	if ver1.Minor != ver2.Minor {
		if ver1.Minor < ver2.Minor {
			return -1
		}
		return 1
	}

	if ver1.Patch != ver2.Patch {
		if ver1.Patch < ver2.Patch {
			return -1
		}
		return 1
	}

	return 0
}

// validateSkillID 验证技能 ID
func validateSkillID(id string) error {
	// 技能 ID 只能包含字母、数字、下划线和连字符
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, id)
	if !matched {
		return fmt.Errorf("invalid skill ID: %s", id)
	}
	return nil
}

// GetModuleInfo 获取模块信息
func GetModuleInfo() map[string]string {
	return map[string]string{
		"name":       "AgentFramework Skills",
		"version":    "1.0.0",
		"go_version": runtime.Version(),
		"os":         runtime.GOOS,
		"arch":       runtime.GOARCH,
	}
}
