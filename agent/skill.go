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
// of the GNU Affero GPL version 3 as long as you maintain
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
	"sync"

	"github.com/cloudwego/eino/schema"

	"AgentFramework/agent/errors"
)

// Skill 定义了AI Agent可以使用的技能接口
// 技能是Agent执行特定任务的能力，如文件操作、网络请求、代码执行等
// 技能可以是内置的，也可以是第三方开发的
// 技能需要实现Info()和Invoke()方法
// Info()方法返回技能的元信息，包括名称、描述、参数等
// Invoke()方法执行技能，接收上下文和输入参数，返回执行结果
// 技能执行可能成功或失败，失败时返回错误信息
// 技能可以有多个参数，参数类型可以是基本类型、数组、对象等
// 技能可以有返回值，返回值类型可以是基本类型、数组、对象等
// 技能可以有副作用，如修改文件、发送网络请求等
// 技能应该是幂等的，即多次调用相同参数应该产生相同结果
// 技能应该是线程安全的，即可以同时被多个Agent调用
// 技能应该是可测试的，即可以独立测试，不需要依赖其他组件
// 技能应该是可扩展的，即可以通过配置调整行为
// 技能应该是可监控的，即可以记录执行时间、调用次数等
// 技能应该是可审计的，即可以记录调用者、参数、结果等
// 技能应该是安全的，即不允许越权操作，不允许执行恶意代码
// 技能应该是高效的，即执行时间短，资源消耗低
// 技能应该是可靠的，即不会崩溃，不会产生不可预期的结果
// 技能应该是文档化的，即提供清晰的使用说明

// Skill 技能接口，定义了技能的基本能力
type Skill interface {
	// Info 返回技能的元信息，包括名称、描述、参数等
	Info(ctx context.Context) (*schema.ToolInfo, error)

	// Invoke 执行技能，接收上下文和输入参数，返回执行结果
	Invoke(ctx context.Context, input string) (string, error)

	// IsEnabled 检查技能是否启用
	IsEnabled(ctx context.Context) bool

	// GetMetadata 获取技能的元数据，包括版本、作者、依赖等
	GetMetadata(ctx context.Context) SkillMetadata
}

// SkillMetadata 技能元数据，包含技能的版本、作者、依赖等信息
type SkillMetadata struct {
	Name         string         `json:"name"`         // 技能名称
	Version      string         `json:"version"`      // 技能版本
	Author       string         `json:"author"`       // 技能作者
	Description  string         `json:"description"`  // 技能描述
	Category     string         `json:"category"`     // 技能分类
	Tags         []string       `json:"tags"`         // 技能标签
	Dependencies []string       `json:"dependencies"` // 技能依赖
	License      string         `json:"license"`      // 技能许可证
	Homepage     string         `json:"homepage"`     // 技能主页
	Repository   string         `json:"repository"`   // 技能仓库
	Keywords     []string       `json:"keywords"`     // 技能关键词
	Config       map[string]any `json:"config"`       // 技能配置
}

// SkillLibrary 技能库接口，用于管理和检索技能
// 技能库负责技能的注册、检索、启用/禁用等操作
// 技能库可以从本地文件系统、远程服务器、数据库等加载技能
// 技能库可以缓存技能，提高检索效率
// 技能库可以支持技能的热加载，即无需重启系统即可更新技能
// 技能库可以支持技能的版本管理，即同一技能可以有多个版本
// 技能库可以支持技能的依赖管理，即自动加载依赖的技能
// 技能库可以支持技能的权限管理，即控制哪些Agent可以使用哪些技能
// 技能库可以支持技能的监控，即记录技能的调用情况
// 技能库可以支持技能的测试，即提供技能测试框架
// 技能库可以支持技能的文档生成，即自动生成技能文档

// SkillLibrary 技能库接口，用于管理和检索技能
type SkillLibrary interface {
	// RegisterSkill 注册技能
	RegisterSkill(ctx context.Context, skill Skill) error

	// UnregisterSkill 注销技能
	UnregisterSkill(ctx context.Context, skillName string) error

	// GetSkill 根据名称获取技能
	GetSkill(ctx context.Context, skillName string) (Skill, bool)

	// GetAllSkills 获取所有技能
	GetAllSkills(ctx context.Context) map[string]Skill

	// GetSkillsByCategory 根据分类获取技能
	GetSkillsByCategory(ctx context.Context, category string) map[string]Skill

	// GetSkillsByTag 根据标签获取技能
	GetSkillsByTag(ctx context.Context, tag string) map[string]Skill

	// EnableSkill 启用技能
	EnableSkill(ctx context.Context, skillName string) error

	// DisableSkill 禁用技能
	DisableSkill(ctx context.Context, skillName string) error

	// IsSkillEnabled 检查技能是否启用
	IsSkillEnabled(ctx context.Context, skillName string) bool

	// GetSkillMetadata 获取技能元数据
	GetSkillMetadata(ctx context.Context, skillName string) (SkillMetadata, bool)

	// LoadSkills 从外部加载技能
	LoadSkills(ctx context.Context, source string) error

	// ReloadSkills 重新加载技能
	ReloadSkills(ctx context.Context) error

	// LoadMCPSkills 从 MCP 服务器加载技能
	LoadMCPSkills(ctx context.Context, client *MCPClient) error

	// UnloadMCPSkills 卸载 MCP 技能
	UnloadMCPSkills(ctx context.Context) error

	// GetMCPSkills 获取所有 MCP 技能
	GetMCPSkills(ctx context.Context) map[string]Skill

	// RefreshMCPSkills 刷新 MCP 技能
	RefreshMCPSkills(ctx context.Context) error

	// Close 关闭技能库，释放资源
	Close(ctx context.Context) error
}

// SkillSpec 技能配置规范，用于定义技能的配置
type SkillSpec struct {
	Name         string         `yaml:"name"`         // 技能名称
	Type         string         `yaml:"type"`         // 技能类型
	Enabled      bool           `yaml:"enabled"`      // 是否启用
	Config       map[string]any `yaml:"config"`       // 技能配置
	Dependencies []string       `yaml:"dependencies"` // 技能依赖
	Tags         []string       `yaml:"tags"`         // 技能标签
	Category     string         `yaml:"category"`     // 技能分类
}

// DefaultSkillLibrary 技能库的默认实现
type DefaultSkillLibrary struct {
	skills        map[string]map[string]Skill // 技能映射，键为技能名称，值为版本到技能的映射
	latestSkills  map[string]string           // 最新版本的技能映射，键为技能名称，值为最新版本
	config        map[string]*SkillSpec       // 技能配置映射
	loadedSkills  map[string]bool             // 已加载的技能映射，用于依赖管理
	metadataCache map[string]*SkillMetadata   // 技能元数据缓存
	toolInfoCache map[string]*schema.ToolInfo // 技能工具信息缓存
	mcpSkills     map[string]Skill            // MCP 技能缓存
	mcpClient     *MCPClient                  // MCP 客户端引用
	mu            sync.RWMutex                // 读写锁，保证线程安全
	metadataMutex sync.RWMutex                // 元数据缓存读写锁
}

// NewSkillLibrary 创建一个新的技能库实例
func NewSkillLibrary() SkillLibrary {
	return &DefaultSkillLibrary{
		skills:        make(map[string]map[string]Skill),
		latestSkills:  make(map[string]string),
		config:        make(map[string]*SkillSpec),
		loadedSkills:  make(map[string]bool),
		metadataCache: make(map[string]*SkillMetadata),
		toolInfoCache: make(map[string]*schema.ToolInfo),
		mcpSkills:     make(map[string]Skill),
	}
}

// RegisterSkill 注册技能
func (sl *DefaultSkillLibrary) RegisterSkill(ctx context.Context, skill Skill) error {
	sl.mu.Lock()
	defer sl.mu.Unlock()

	info, err := skill.Info(ctx)
	if err != nil {
		return err
	}

	metadata := skill.GetMetadata(ctx)
	name := info.Name
	version := metadata.Version

	// 如果技能名称不存在，创建版本映射
	if _, exists := sl.skills[name]; !exists {
		sl.skills[name] = make(map[string]Skill)
	}

	// 注册技能版本
	sl.skills[name][version] = skill

	// 更新最新版本映射
	if latestVersion, exists := sl.latestSkills[name]; !exists || version > latestVersion {
		sl.latestSkills[name] = version
	}

	// 清空该技能的元数据缓存，确保下次获取时重新生成
	sl.metadataMutex.Lock()
	defer sl.metadataMutex.Unlock()
	delete(sl.metadataCache, name)
	delete(sl.toolInfoCache, name)

	return nil
}

// UnregisterSkill 注销技能
func (sl *DefaultSkillLibrary) UnregisterSkill(ctx context.Context, skillName string) error {
	sl.mu.Lock()
	defer sl.mu.Unlock()

	// 删除所有版本的技能
	delete(sl.skills, skillName)
	// 删除最新版本映射
	delete(sl.latestSkills, skillName)
	// 删除配置
	delete(sl.config, skillName)
	// 删除已加载状态
	delete(sl.loadedSkills, skillName)

	// 清空该技能的元数据缓存
	sl.metadataMutex.Lock()
	defer sl.metadataMutex.Unlock()
	delete(sl.metadataCache, skillName)
	delete(sl.toolInfoCache, skillName)

	return nil
}

// UnregisterSkillVersion 注销特定版本的技能
func (sl *DefaultSkillLibrary) UnregisterSkillVersion(ctx context.Context, skillName, version string) error {
	sl.mu.Lock()
	defer sl.mu.Unlock()

	if versionMap, exists := sl.skills[skillName]; exists {
		// 检查是否删除的是最新版本
		isLatestVersion := false
		if sl.latestSkills[skillName] == version {
			isLatestVersion = true
		}

		// 删除特定版本的技能
		delete(versionMap, version)

		// 如果该技能没有其他版本，删除整个技能映射
		if len(versionMap) == 0 {
			delete(sl.skills, skillName)
			delete(sl.latestSkills, skillName)
			delete(sl.config, skillName)
			delete(sl.loadedSkills, skillName)

			// 清空该技能的元数据缓存
			sl.metadataMutex.Lock()
			defer sl.metadataMutex.Unlock()
			delete(sl.metadataCache, skillName)
			delete(sl.toolInfoCache, skillName)
		} else {
			// 更新最新版本映射
			latestVersion := ""
			for v := range versionMap {
				if latestVersion == "" || v > latestVersion {
					latestVersion = v
				}
			}
			sl.latestSkills[skillName] = latestVersion

			// 如果删除的是最新版本，清空元数据缓存
			if isLatestVersion {
				sl.metadataMutex.Lock()
				defer sl.metadataMutex.Unlock()
				delete(sl.metadataCache, skillName)
				delete(sl.toolInfoCache, skillName)
			}
		}
	}

	return nil
}

// GetSkill 根据名称获取技能，返回最新版本
// GetSkillVersion 根据名称和版本获取技能
func (sl *DefaultSkillLibrary) GetSkillVersion(ctx context.Context, skillName, version string) (Skill, bool) {
	sl.mu.RLock()
	defer sl.mu.RUnlock()

	return sl.getSkillVersion(skillName, version)
}

// getSkillVersion 根据名称和版本获取技能（内部方法，不加锁）
func (sl *DefaultSkillLibrary) getSkillVersion(skillName, version string) (Skill, bool) {
	if versionMap, exists := sl.skills[skillName]; exists {
		skill, exists := versionMap[version]
		return skill, exists
	}
	return nil, false
}


// GetAllSkillVersions 获取所有技能的所有版本
func (sl *DefaultSkillLibrary) GetAllSkillVersions(ctx context.Context) map[string]map[string]Skill {
	sl.mu.RLock()
	defer sl.mu.RUnlock()

	// 返回副本，防止外部修改
	result := make(map[string]map[string]Skill)
	for skillName, versionMap := range sl.skills {
		result[skillName] = make(map[string]Skill)
		for version, skill := range versionMap {
			result[skillName][version] = skill
		}
	}
	return result
}

// GetSkillsByCategory 根据分类获取技能
func (sl *DefaultSkillLibrary) GetSkillsByCategory(ctx context.Context, category string) map[string]Skill {
	sl.mu.RLock()
	defer sl.mu.RUnlock()

	result := make(map[string]Skill)
	for skillName, latestVersion := range sl.latestSkills {
		if skill, exists := sl.getSkillVersion(skillName, latestVersion); exists {
			metadata := skill.GetMetadata(ctx)
			if metadata.Category == category {
				result[skillName] = skill
			}
		}
	}
	return result
}

// GetSkillsByTag 根据标签获取技能
func (sl *DefaultSkillLibrary) GetSkillsByTag(ctx context.Context, tag string) map[string]Skill {
	sl.mu.RLock()
	defer sl.mu.RUnlock()

	result := make(map[string]Skill)
	for skillName, latestVersion := range sl.latestSkills {
		if skill, exists := sl.getSkillVersion(skillName, latestVersion); exists {
			metadata := skill.GetMetadata(ctx)
			for _, t := range metadata.Tags {
				if t == tag {
					result[skillName] = skill
					break
				}
			}
		}
	}
	return result
}

// EnableSkill 启用技能，启用最新版本
func (sl *DefaultSkillLibrary) EnableSkill(ctx context.Context, skillName string) error {
	sl.mu.Lock()
	defer sl.mu.Unlock()

	// 获取最新版本
	latestVersion, exists := sl.latestSkills[skillName]
	if !exists {
		return fmt.Errorf("skill %s not found", skillName)
	}

	return sl.enableSkillVersion(skillName, latestVersion)
}

// EnableSkillVersion 启用特定版本的技能
func (sl *DefaultSkillLibrary) EnableSkillVersion(ctx context.Context, skillName, version string) error {
	sl.mu.Lock()
	defer sl.mu.Unlock()

	return sl.enableSkillVersion(skillName, version)
}

// enableSkillVersion 启用特定版本的技能（内部方法，不加锁）
func (sl *DefaultSkillLibrary) enableSkillVersion(skillName, version string) error {
	if skill, exists := sl.getSkillVersion(skillName, version); exists {
		// 检查是否是skillsAdapter类型，它支持SetEnabled方法
		if adapter, ok := skill.(*skillsAdapter); ok {
			adapter.inner.SetEnabled(true)
			return nil
		}
		return fmt.Errorf("skill %s version %s does not support enable/disable", skillName, version)
	}
	return fmt.Errorf("skill %s version %s not found", skillName, version)
}

// DisableSkill 禁用技能，禁用最新版本
func (sl *DefaultSkillLibrary) DisableSkill(ctx context.Context, skillName string) error {
	sl.mu.Lock()
	defer sl.mu.Unlock()

	// 获取最新版本
	latestVersion, exists := sl.latestSkills[skillName]
	if !exists {
		return fmt.Errorf("skill %s not found", skillName)
	}

	return sl.disableSkillVersion(skillName, latestVersion)
}

// DisableSkillVersion 禁用特定版本的技能
func (sl *DefaultSkillLibrary) DisableSkillVersion(ctx context.Context, skillName, version string) error {
	sl.mu.Lock()
	defer sl.mu.Unlock()

	return sl.disableSkillVersion(skillName, version)
}

// disableSkillVersion 禁用特定版本的技能（内部方法，不加锁）
func (sl *DefaultSkillLibrary) disableSkillVersion(skillName, version string) error {
	if skill, exists := sl.getSkillVersion(skillName, version); exists {
		// 检查是否是skillsAdapter类型，它支持SetEnabled方法
		if adapter, ok := skill.(*skillsAdapter); ok {
			adapter.inner.SetEnabled(false)
			return nil
		}
		return fmt.Errorf("skill %s version %s does not support enable/disable", skillName, version)
	}
	return fmt.Errorf("skill %s version %s not found", skillName, version)
}

// IsSkillEnabled 检查技能是否启用，检查最新版本
func (sl *DefaultSkillLibrary) IsSkillEnabled(ctx context.Context, skillName string) bool {
	sl.mu.RLock()
	defer sl.mu.RUnlock()

	// 获取最新版本
	latestVersion, exists := sl.latestSkills[skillName]
	if !exists {
		return false
	}

	return sl.isSkillVersionEnabled(ctx, skillName, latestVersion)
}

// IsSkillVersionEnabled 检查特定版本的技能是否启用
func (sl *DefaultSkillLibrary) IsSkillVersionEnabled(ctx context.Context, skillName, version string) bool {
	sl.mu.RLock()
	defer sl.mu.RUnlock()

	return sl.isSkillVersionEnabled(ctx, skillName, version)
}

// isSkillVersionEnabled 检查特定版本的技能是否启用（内部方法，不加锁）
func (sl *DefaultSkillLibrary) isSkillVersionEnabled(ctx context.Context, skillName, version string) bool {
	skill, found := sl.getSkillVersion(skillName, version)
	if !found {
		return false
	}
	return skill.IsEnabled(ctx)
}

// GetSkillMetadata 获取技能元数据，返回最新版本的元数据
func (sl *DefaultSkillLibrary) GetSkillMetadata(ctx context.Context, skillName string) (SkillMetadata, bool) {
	// 首先尝试从缓存获取
	sl.metadataMutex.RLock()
	if metadata, exists := sl.metadataCache[skillName]; exists {
		sl.metadataMutex.RUnlock()
		return *metadata, true
	}
	sl.metadataMutex.RUnlock()

	// 缓存未命中，获取最新版本并生成元数据
	sl.mu.RLock()
	latestVersion, exists := sl.latestSkills[skillName]
	if !exists {
		sl.mu.RUnlock()
		return SkillMetadata{}, false
	}

	// 获取技能实例
	skill, found := sl.getSkillVersion(skillName, latestVersion)
	sl.mu.RUnlock()

	if !found {
		return SkillMetadata{}, false
	}

	// 生成元数据
	metadata := skill.GetMetadata(ctx)

	// 存入缓存
	sl.metadataMutex.Lock()
	sl.metadataCache[skillName] = &metadata
	sl.metadataMutex.Unlock()

	return metadata, true
}

// GetSkillVersionMetadata 获取特定版本的技能元数据
func (sl *DefaultSkillLibrary) GetSkillVersionMetadata(ctx context.Context, skillName, version string) (SkillMetadata, bool) {
	sl.mu.RLock()
	defer sl.mu.RUnlock()

	return sl.getSkillVersionMetadata(ctx, skillName, version)
}

// getSkillVersionMetadata 获取特定版本的技能元数据（内部方法，不加锁）
func (sl *DefaultSkillLibrary) getSkillVersionMetadata(ctx context.Context, skillName, version string) (SkillMetadata, bool) {
	skill, found := sl.getSkillVersion(skillName, version)
	if !found {
		return SkillMetadata{}, false
	}
	return skill.GetMetadata(ctx), true
}

// LoadMCPSkills 从 MCP 服务器加载技能
func (sl *DefaultSkillLibrary) LoadMCPSkills(ctx context.Context, client *MCPClient) error {
	sl.mu.Lock()
	defer sl.mu.Unlock()

	// 保存 MCP 客户端引用
	sl.mcpClient = client

	// 发现并加载 MCP 技能
	if err := client.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize MCP client: %w", err)
	}

	tools := client.ListTools()
	for name, toolDef := range tools {
		// 转换为 schema.ToolInfo
		toolInfo := &schema.ToolInfo{
			Name:        name,
			Desc: toolDef.Description,
			ParamsOneOf: schema.NewParamsOneOfByParams(make(map[string]*schema.ParameterInfo)),
		}

		// 简单处理输入模式（在实际应用中可能需要更复杂的转换）
		if toolDef.InputSchema != nil {
			params := make(map[string]*schema.ParameterInfo)
			for paramName, paramSchema := range toolDef.InputSchema {
				params[paramName] = &schema.ParameterInfo{
					Type: "string",
					Desc: fmt.Sprintf("%v", paramSchema),
				}
			}
			toolInfo.ParamsOneOf = schema.NewParamsOneOfByParams(params)
		}

		skill := NewMCPSkillAdapter(client, toolInfo)
		sl.mcpSkills[name] = skill
	}

	return nil
}

// UnloadMCPSkills 卸载 MCP 技能
func (sl *DefaultSkillLibrary) UnloadMCPSkills(ctx context.Context) error {
	sl.mu.Lock()
	defer sl.mu.Unlock()

	sl.mcpSkills = make(map[string]Skill)
	sl.mcpClient = nil

	return nil
}

// GetMCPSkills 获取所有 MCP 技能
func (sl *DefaultSkillLibrary) GetMCPSkills(ctx context.Context) map[string]Skill {
	sl.mu.RLock()
	defer sl.mu.RUnlock()

	copySkills := make(map[string]Skill)
	for name, skill := range sl.mcpSkills {
		copySkills[name] = skill
	}

	return copySkills
}

// RefreshMCPSkills 刷新 MCP 技能
func (sl *DefaultSkillLibrary) RefreshMCPSkills(ctx context.Context) error {
	sl.mu.Lock()
	defer sl.mu.Unlock()

	if sl.mcpClient == nil {
		return fmt.Errorf("MCP client not connected")
	}

	// 清除现有 MCP 技能
	sl.mcpSkills = make(map[string]Skill)

	// 重新加载技能
	if err := sl.mcpClient.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize MCP client: %w", err)
	}

	tools := sl.mcpClient.ListTools()
	for name, toolDef := range tools {
		// 转换为 schema.ToolInfo
		toolInfo := &schema.ToolInfo{
			Name:        name,
			Desc: toolDef.Description,
			ParamsOneOf: schema.NewParamsOneOfByParams(make(map[string]*schema.ParameterInfo)),
		}

		// 简单处理输入模式（在实际应用中可能需要更复杂的转换）
		if toolDef.InputSchema != nil {
			params := make(map[string]*schema.ParameterInfo)
			for paramName, paramSchema := range toolDef.InputSchema {
				params[paramName] = &schema.ParameterInfo{
					Type: "string",
					Desc: fmt.Sprintf("%v", paramSchema),
				}
			}
			toolInfo.ParamsOneOf = schema.NewParamsOneOfByParams(params)
		}

		skill := NewMCPSkillAdapter(sl.mcpClient, toolInfo)
		sl.mcpSkills[name] = skill
	}

	return nil
}

// LoadSkills 从外部加载技能
func (sl *DefaultSkillLibrary) LoadSkills(ctx context.Context, source string) error {
	// 从外部加载技能的逻辑将在后续扩展
	return nil
}

// ReloadSkills 重新加载技能
func (sl *DefaultSkillLibrary) ReloadSkills(ctx context.Context) error {
	// 重新加载技能的逻辑将在后续扩展
	return nil
}

// Close 关闭技能库，释放资源
func (sl *DefaultSkillLibrary) Close(ctx context.Context) error {
	sl.mu.Lock()
	defer sl.mu.Unlock()

	if sl.mcpClient != nil {
		if err := sl.mcpClient.Disconnect(ctx); err != nil {
			return err
		}
	}

	return nil
}

// GetSkill 根据名称获取技能，优先返回内置技能，其次返回 MCP 技能
func (sl *DefaultSkillLibrary) GetSkill(ctx context.Context, skillName string) (Skill, bool) {
	sl.mu.RLock()
	defer sl.mu.RUnlock()

	// 首先检查内置技能
	latestVersion, exists := sl.latestSkills[skillName]
	if exists {
		if skill, found := sl.getSkillVersion(skillName, latestVersion); found {
			return skill, true
		}
	}

	// 检查 MCP 技能
	if skill, exists := sl.mcpSkills[skillName]; exists {
		return skill, true
	}

	return nil, false
}

// GetAllSkills 获取所有技能，包括内置技能和 MCP 技能
func (sl *DefaultSkillLibrary) GetAllSkills(ctx context.Context) map[string]Skill {
	sl.mu.RLock()
	defer sl.mu.RUnlock()

	result := make(map[string]Skill)

	// 添加内置技能
	for skillName, latestVersion := range sl.latestSkills {
		if skill, exists := sl.getSkillVersion(skillName, latestVersion); exists {
			result[skillName] = skill
		}
	}

	// 添加 MCP 技能
	for name, skill := range sl.mcpSkills {
		if _, exists := result[name]; !exists {
			result[name] = skill
		}
	}

	return result
}

// SkillExecutor 技能执行器，用于执行技能并处理结果
type SkillExecutor struct {
	skillLibrary SkillLibrary // 技能库
}

// NewSkillExecutor 创建一个新的技能执行器实例
func NewSkillExecutor(skillLibrary SkillLibrary) *SkillExecutor {
	return &SkillExecutor{
		skillLibrary: skillLibrary,
	}
}

// Execute 执行技能
func (se *SkillExecutor) Execute(ctx context.Context, skillName string, input string) (string, error) {
	// 从技能库获取技能
	skill, found := se.skillLibrary.GetSkill(ctx, skillName)
	if !found {
		return "", errors.New(errors.ErrCodeNotFound, "skill not found: "+skillName)
	}

	// 检查技能是否启用
	if !skill.IsEnabled(ctx) {
		return "", errors.New(errors.ErrCodeForbidden, "skill is disabled: "+skillName)
	}

	// 执行技能
	return skill.Invoke(ctx, input)
}
