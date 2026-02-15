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

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// DataAnalysisAgent 数据分析智能体，专门用于数据统计和可视化
type DataAnalysisAgent struct {
	name              string
	model             ChatModel
	pythonRunner      PythonRunner
	chartGenerator    *ChartGenerator
	tools             map[string]tool.InvokableTool
	stateMachine      *StateMachine
	memoryManager     *MemoryManager
	thread            *Thread
}

// PythonRunner Python代码执行接口
type PythonRunner interface {
	Execute(ctx context.Context, code string) (string, error)
}

// DataAnalysisConfig 数据分析智能体配置
type DataAnalysisConfig struct {
	Name         string
	Model        ChatModel
	PythonRunner PythonRunner
	Tools        []tool.BaseTool
	MemoryOpts   MemoryOptions
}

// NewDataAnalysisAgent 创建数据分析智能体
func NewDataAnalysisAgent(ctx context.Context, cfg DataAnalysisConfig) (*DataAnalysisAgent, error) {
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

	// 创建图表生成器
	chartGen := &ChartGenerator{
		runner: cfg.PythonRunner,
	}

	// 创建状态机
	stateMachine := NewStateMachineWithDefaults()

	// 创建内存管理器
	memoryManager := NewMemoryManager(cfg.MemoryOpts)

	return &DataAnalysisAgent{
		name:           cfg.Name,
		model:          boundModel,
		pythonRunner:   cfg.PythonRunner,
		chartGenerator: chartGen,
		tools:          toolMap,
		stateMachine:   stateMachine,
		memoryManager:  memoryManager,
		thread:         &Thread{ID: cfg.Name},
	}, nil
}

// Name 返回智能体名称
func (a *DataAnalysisAgent) Name() string {
	return a.name
}

// Run 执行数据分析
func (a *DataAnalysisAgent) Run(ctx context.Context, input string, opts ...model.Option) (*schema.Message, error) {
	// 转换到运行状态
	if err := a.stateMachine.Transition(ctx, StateRunning, "Starting data analysis", nil); err != nil {
		return nil, err
	}

	defer func() {
		if a.stateMachine.Current() == StateRunning {
			_ = a.stateMachine.Transition(context.Background(), StateFinished, "Analysis completed", nil)
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

	_ = a.stateMachine.Transition(ctx, StateFinished, "Analysis completed successfully", map[string]any{
		"response_length": len(resp.Content),
	})

	return resp, nil
}

// Analyze 分析数据
func (a *DataAnalysisAgent) Analyze(ctx context.Context, data interface{}, analysisType string) (*AnalysisResult, error) {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	prompt := fmt.Sprintf(`请分析以下数据，类型：%s

数据：
%s

请提供：
1. 数据概览（行数、列数、数据类型）
2. 统计摘要（均值、中位数、标准差等）
3. 数据分布特征
4. 异常值检测
5. 相关性分析（如适用）
6. 主要发现和洞察

以JSON格式返回分析结果。`, analysisType, string(dataJSON))

	systemMsg := &schema.Message{
		Role:    schema.System,
		Content: "你是一个数据分析师，擅长统计分析、数据探索和洞察发现。",
	}

	userMsg := &schema.Message{
		Role:    schema.User,
		Content: prompt,
	}

	resp, err := a.model.Generate(ctx, []*schema.Message{systemMsg, userMsg})
	if err != nil {
		return nil, err
	}

	var result AnalysisResult
	if err := json.Unmarshal([]byte(resp.Content), &result); err != nil {
		// 如果解析失败，将文本内容放入概览
		result = AnalysisResult{
			Overview: resp.Content,
		}
	}

	return &result, nil
}

// GenerateChart 生成图表
func (a *DataAnalysisAgent) GenerateChart(ctx context.Context, req ChartRequest) ([]byte, error) {
	return a.chartGenerator.Generate(ctx, req)
}

// StatisticalSummary 统计摘要
func (a *DataAnalysisAgent) StatisticalSummary(ctx context.Context, data interface{}) (*StatisticalSummary, error) {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	code := fmt.Sprintf(`
import json
import pandas as pd
import numpy as np
from io import StringIO

data = json.loads('''%s''')
df = pd.DataFrame(data)

summary = {
    "rows": len(df),
    "columns": len(df.columns),
    "column_names": list(df.columns),
    "dtypes": {col: str(dtype) for col, dtype in df.dtypes.items()},
    "numeric_summary": {}
}

numeric_cols = df.select_dtypes(include=[np.number]).columns
for col in numeric_cols:
    summary["numeric_summary"][col] = {
        "count": int(df[col].count()),
        "mean": float(df[col].mean()),
        "median": float(df[col].median()),
        "std": float(df[col].std()),
        "min": float(df[col].min()),
        "max": float(df[col].max()),
        "q25": float(df[col].quantile(0.25)),
        "q75": float(df[col].quantile(0.75))
    }

print(json.dumps(summary, indent=2))
`, string(dataJSON))

	output, err := a.pythonRunner.Execute(ctx, code)
	if err != nil {
		return nil, err
	}

	var summary StatisticalSummary
	if err := json.Unmarshal([]byte(output), &summary); err != nil {
		return nil, fmt.Errorf("failed to parse summary: %w", err)
	}

	return &summary, nil
}

// GenerateReport 生成分析报告
func (a *DataAnalysisAgent) GenerateReport(ctx context.Context, data interface{}, template string) (string, error) {
	analysis, err := a.Analyze(ctx, data, "comprehensive")
	if err != nil {
		return "", err
	}

	prompt := fmt.Sprintf(`基于以下数据分析结果，生成一份分析报告：

模板：%s

分析结果：
%s

请生成一份专业、易读的数据分析报告。`, template, analysis.Overview)

	systemMsg := &schema.Message{
		Role:    schema.System,
		Content: "你是一个专业的数据报告撰写专家。",
	}

	userMsg := &schema.Message{
		Role:    schema.User,
		Content: prompt,
	}

	resp, err := a.model.Generate(ctx, []*schema.Message{systemMsg, userMsg})
	if err != nil {
		return "", err
	}

	return resp.Content, nil
}

// CorrelationAnalysis 相关性分析
func (a *DataAnalysisAgent) CorrelationAnalysis(ctx context.Context, data interface{}) (*CorrelationMatrix, error) {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	code := fmt.Sprintf(`
import json
import pandas as pd
import numpy as np

data = json.loads('''%s''')
df = pd.DataFrame(data)

# 只选择数值列
numeric_df = df.select_dtypes(include=[np.number])
corr_matrix = numeric_df.corr()

result = {
    "columns": list(corr_matrix.columns),
    "matrix": corr_matrix.to_dict()
}

print(json.dumps(result, indent=2))
`, string(dataJSON))

	output, err := a.pythonRunner.Execute(ctx, code)
	if err != nil {
		return nil, err
	}

	var result CorrelationMatrix
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		return nil, fmt.Errorf("failed to parse correlation matrix: %w", err)
	}

	return &result, nil
}

// buildMessages 构建消息列表
func (a *DataAnalysisAgent) buildMessages(latest *schema.Message) []*schema.Message {
	var msgs []*schema.Message

	// 添加系统提示
	msgs = append(msgs, &schema.Message{
		Role:    schema.System,
		Content: "你是一个专业的数据分析师，擅长使用Python进行数据分析、统计计算和可视化。你可以使用pandas、numpy、matplotlib等库来处理数据。",
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
func (a *DataAnalysisAgent) runTools(ctx context.Context, toolCallMsg *schema.Message) ([]*schema.Message, error) {
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

// ==================== 数据结构定义 ====================

// AnalysisResult 分析结果
type AnalysisResult struct {
	Overview           string                 `json:"overview"`
	Rows               int                    `json:"rows,omitempty"`
	Columns            int                    `json:"columns,omitempty"`
	ColumnTypes        map[string]string      `json:"column_types,omitempty"`
	Statistics         map[string]StatisticalMetric `json:"statistics,omitempty"`
	Distributions      map[string]DistributionInfo `json:"distributions,omitempty"`
	Outliers           []OutlierInfo          `json:"outliers,omitempty"`
	Correlations       map[string]float64     `json:"correlations,omitempty"`
	Insights           []string               `json:"insights,omitempty"`
}

// StatisticalSummary 统计摘要
type StatisticalSummary struct {
	Rows           int                    `json:"rows"`
	Columns        int                    `json:"columns"`
	ColumnNames    []string               `json:"column_names"`
	Dtypes         map[string]string      `json:"dtypes"`
	NumericSummary map[string]NumericStats `json:"numeric_summary"`
}

// NumericStats 数值统计
type NumericStats struct {
	Count   int     `json:"count"`
	Mean    float64 `json:"mean"`
	Median  float64 `json:"median"`
	Std     float64 `json:"std"`
	Min     float64 `json:"min"`
	Max     float64 `json:"max"`
	Q25     float64 `json:"q25"`
	Q75     float64 `json:"q75"`
}

// StatisticalMetric 统计指标
type StatisticalMetric struct {
	Mean   float64 `json:"mean"`
	Median float64 `json:"median"`
	Std    float64 `json:"std"`
	Min    float64 `json:"min"`
	Max    float64 `json:"max"`
}

// DistributionInfo 分布信息
type DistributionInfo struct {
	Type        string   `json:"type"` // normal, uniform, skewed, etc.
	Skewness    float64  `json:"skewness,omitempty"`
	Kurtosis    float64  `json:"kurtosis,omitempty"`
	Percentiles []float64 `json:"percentiles,omitempty"`
}

// OutlierInfo 异常值信息
type OutlierInfo struct {
	Column   string  `json:"column"`
	Index    int     `json:"index"`
	Value    float64 `json:"value"`
	Reason   string  `json:"reason"`
}

// CorrelationMatrix 相关性矩阵
type CorrelationMatrix struct {
	Columns []string            `json:"columns"`
	Matrix  map[string]map[string]float64 `json:"matrix"`
}

// ChartRequest 图表请求
type ChartRequest struct {
	Type     string                 `json:"type"`     // line, bar, scatter, pie, histogram, box
	Data     interface{}            `json:"data"`
	X        string                 `json:"x,omitempty"`        // X轴字段
	Y        string                 `json:"y,omitempty"`        // Y轴字段
	Title    string                 `json:"title,omitempty"`
	XLabel   string                 `json:"x_label,omitempty"`
	YLabel   string                 `json:"y_label,omitempty"`
	Width    int                    `json:"width,omitempty"`
	Height   int                    `json:"height,omitempty"`
	Options  map[string]interface{} `json:"options,omitempty"`
}

// ChartResponse 图表响应
type ChartResponse struct {
	Image    []byte `json:"image"`
	Format   string `json:"format"` // png, svg
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	Filename string `json:"filename,omitempty"`
}
