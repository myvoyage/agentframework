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

	"github.com/cloudwego/eino/schema"

	"AgentFramework/agent/skills"
)

// ------------------------------
// 基础技能模板 - 提供常用技能的基础实现
// 开发者可以基于这些模板快速扩展和定制自己的技能
// ------------------------------

// skillsAdapter 适配器，将skills.Skill转换为agent.Skill
// 解决skills.SkillMetadata和agent.SkillMetadata类型不兼容的问题
// ------------------------------
type skillsAdapter struct {
	inner skills.Skill // 内部持有skills.Skill实例
}

// Info 返回技能的元信息
func (a *skillsAdapter) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return a.inner.Info(ctx)
}

// Invoke 执行技能，处理输入并返回输出
func (a *skillsAdapter) Invoke(ctx context.Context, input string) (string, error) {
	return a.inner.Invoke(ctx, input)
}

// IsEnabled 检查技能是否启用
func (a *skillsAdapter) IsEnabled(ctx context.Context) bool {
	return a.inner.IsEnabled(ctx)
}

// GetMetadata 获取技能的元数据，将skills.SkillMetadata转换为agent.SkillMetadata
func (a *skillsAdapter) GetMetadata(ctx context.Context) SkillMetadata {
	innerMeta := a.inner.GetMetadata(ctx)
	return SkillMetadata{
		Name:        innerMeta.Name,
		Version:     innerMeta.Version,
		Author:      innerMeta.Author,
		Description: innerMeta.Description,
		Category:    innerMeta.Category,
		Tags:        innerMeta.Tags,
	}
}

// NewHTTPRequestSkill 创建一个新的HTTP请求技能实例
func NewHTTPRequestSkill() Skill {
	return &skillsAdapter{inner: skills.NewHTTPRequestSkill()}
}

// NewFileOperationSkill 创建一个新的文件操作技能实例
func NewFileOperationSkill() Skill {
	skill, _ := skills.NewFileOperationSkill(nil)
	return &skillsAdapter{inner: skill}
}

// NewCodeExecutionSkill 创建一个新的代码执行技能实例
func NewCodeExecutionSkill() Skill {
	skill, _ := skills.NewCodeExecutionSkill(nil)
	return &skillsAdapter{inner: skill}
}

// NewDataProcessingSkill 创建一个新的数据处理技能实例
func NewDataProcessingSkill() Skill {
	skill, _ := skills.NewDataProcessingSkill(nil)
	return &skillsAdapter{inner: skill}
}
