// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
// Copyright (C) 2025  Agent Framework Contributors

// SPDX-License-Identifier: AGPL-3.0-or-later

package skills

import (
	"context"

	"github.com/cloudwego/eino/schema"
)

// SkillMetadata 技能元数据
// 包含技能的基本信息，如名称、版本、作者等
// ------------------------------

// Skill 技能接口定义
// 所有技能都必须实现这个接口
// ------------------------------
type Skill interface {
	// Info 返回技能的元信息
	Info(ctx context.Context) (*schema.ToolInfo, error)
	// Invoke 执行技能，处理输入并返回输出
	Invoke(ctx context.Context, input string) (string, error)
	// IsEnabled 检查技能是否启用
	IsEnabled(ctx context.Context) bool
	// SetEnabled 设置技能是否启用
	SetEnabled(enabled bool)
	// GetMetadata 获取技能的元数据
	GetMetadata(ctx context.Context) SkillMetadata
	// SetMetadata 设置技能的元数据
	SetMetadata(metadata SkillMetadata)
}
