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

// SPDX-License-Identifier: AGPL-3.0-or-later

package skills

import (
	"sync"
	"time"
)

// SkillMetrics 技能执行指标收集器
type SkillMetrics struct {
	mu sync.RWMutex

	// 技能执行统计
	executionCount    int64           // 总执行次数
	successCount      int64           // 成功次数
	failureCount      int64           // 失败次数
	totalDuration     time.Duration   // 总执行时长
	maxDuration       time.Duration   // 最大执行时长
	minDuration       *time.Duration  // 最小执行时长

	// 按技能名称统计
	bySkill map[string]*SkillMetric

	// 时间窗口统计（最近 N 次执行）
	recentExecutions []ExecutionRecord
	maxRecentCount   int
}

// SkillMetric 单个技能的指标
type SkillMetric struct {
	SkillName    string
	ExecutionCount int64
	SuccessCount   int64
	FailureCount   int64
	TotalDuration  time.Duration
	AvgDuration    time.Duration
	MaxDuration    time.Duration
	MinDuration    *time.Duration
}

// ExecutionRecord 执行记录
type ExecutionRecord struct {
	SkillName   string
	StartTime   time.Time
	EndTime     time.Time
	Duration    time.Duration
	Success     bool
	Error       string
}

// NewSkillMetrics 创建新的指标收集器
func NewSkillMetrics(maxRecentCount int) *SkillMetrics {
	if maxRecentCount <= 0 {
		maxRecentCount = 100
	}
	return &SkillMetrics{
		bySkill:        make(map[string]*SkillMetric),
		recentExecutions: make([]ExecutionRecord, 0, maxRecentCount),
		maxRecentCount: maxRecentCount,
	}
}

// RecordExecution 记录一次技能执行
func (m *SkillMetrics) RecordExecution(skillName string, duration time.Duration, success bool, errMsg string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 更新总体统计
	m.executionCount++
	if success {
		m.successCount++
	} else {
		m.failureCount++
	}
	m.totalDuration += duration
	if duration > m.maxDuration {
		m.maxDuration = duration
	}
	if m.minDuration == nil || duration < *m.minDuration {
		m.minDuration = &duration
	}

	// 更新按技能统计
	if _, ok := m.bySkill[skillName]; !ok {
		m.bySkill[skillName] = &SkillMetric{
			SkillName: skillName,
		}
	}
	metric := m.bySkill[skillName]
	metric.ExecutionCount++
	if success {
		metric.SuccessCount++
	} else {
		metric.FailureCount++
	}
	metric.TotalDuration += duration
	metric.AvgDuration = metric.TotalDuration / time.Duration(metric.ExecutionCount)
	if duration > metric.MaxDuration {
		metric.MaxDuration = duration
	}
	if metric.MinDuration == nil || duration < *metric.MinDuration {
		metric.MinDuration = &duration
	}

	// 添加到最近执行记录
	record := ExecutionRecord{
		SkillName: skillName,
		StartTime: time.Now().Add(-duration),
		EndTime:   time.Now(),
		Duration:  duration,
		Success:   success,
		Error:     errMsg,
	}
	if len(m.recentExecutions) >= m.maxRecentCount {
		m.recentExecutions = m.recentExecutions[1:]
	}
	m.recentExecutions = append(m.recentExecutions, record)
}

// GetSummary 获取指标摘要
func (m *SkillMetrics) GetSummary() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	successRate := 0.0
	if m.executionCount > 0 {
		successRate = float64(m.successCount) / float64(m.executionCount) * 100
	}

	avgDuration := time.Duration(0)
	if m.executionCount > 0 {
		avgDuration = m.totalDuration / time.Duration(m.executionCount)
	}

	return map[string]interface{}{
		"execution_count": m.executionCount,
		"success_count":   m.successCount,
		"failure_count":   m.failureCount,
		"success_rate":    successRate,
		"total_duration":  m.totalDuration.String(),
		"avg_duration":    avgDuration.String(),
		"max_duration":   m.maxDuration.String(),
		"min_duration":   m.minDuration.String(),
		"skill_count":    len(m.bySkill),
	}
}

// GetSkillMetrics 获取指定技能的指标
func (m *SkillMetrics) GetSkillMetrics(skillName string) (map[string]interface{}, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	metric, ok := m.bySkill[skillName]
	if !ok {
		return nil, false
	}

	successRate := 0.0
	if metric.ExecutionCount > 0 {
		successRate = float64(metric.SuccessCount) / float64(metric.ExecutionCount) * 100
	}

	return map[string]interface{}{
		"skill_name":      metric.SkillName,
		"execution_count": metric.ExecutionCount,
		"success_count":   metric.SuccessCount,
		"failure_count":   metric.FailureCount,
		"success_rate":    successRate,
		"total_duration":  metric.TotalDuration.String(),
		"avg_duration":    metric.AvgDuration.String(),
		"max_duration":    metric.MaxDuration.String(),
		"min_duration":   metric.MinDuration.String(),
	}, true
}

// GetAllSkillMetrics 获取所有技能的指标
func (m *SkillMetrics) GetAllSkillMetrics() []map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]map[string]interface{}, 0, len(m.bySkill))
	for _, metric := range m.bySkill {
		successRate := 0.0
		if metric.ExecutionCount > 0 {
			successRate = float64(metric.SuccessCount) / float64(metric.ExecutionCount) * 100
		}
		result = append(result, map[string]interface{}{
			"skill_name":      metric.SkillName,
			"execution_count": metric.ExecutionCount,
			"success_count":   metric.SuccessCount,
			"failure_count":   metric.FailureCount,
			"success_rate":    successRate,
			"total_duration":  metric.TotalDuration.String(),
			"avg_duration":    metric.AvgDuration.String(),
			"max_duration":    metric.MaxDuration.String(),
			"min_duration":   metric.MinDuration.String(),
		})
	}
	return result
}

// GetRecentExecutions 获取最近执行记录
func (m *SkillMetrics) GetRecentExecutions(limit int) []ExecutionRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if limit <= 0 || limit > len(m.recentExecutions) {
		limit = len(m.recentExecutions)
	}

	// 返回最近的记录（从后往前）
	result := make([]ExecutionRecord, limit)
	copy(result, m.recentExecutions[len(m.recentExecutions)-limit:])
	return result
}

// Reset 重置所有指标
func (m *SkillMetrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.executionCount = 0
	m.successCount = 0
	m.failureCount = 0
	m.totalDuration = 0
	m.maxDuration = 0
	m.minDuration = nil
	m.bySkill = make(map[string]*SkillMetric)
	m.recentExecutions = make([]ExecutionRecord, 0, m.maxRecentCount)
}
