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
// with other code, such other code is not other code is not for that reason alone subject to any
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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"
)

// ChartGenerator 图表生成器
type ChartGenerator struct {
	runner PythonRunner
}

// NewChartGenerator 创建图表生成器
func NewChartGenerator(runner PythonRunner) *ChartGenerator {
	return &ChartGenerator{
		runner: runner,
	}
}

// Generate 生成图表
func (g *ChartGenerator) Generate(ctx context.Context, req ChartRequest) ([]byte, error) {
	// 设置默认值
	if req.Width == 0 {
		req.Width = 800
	}
	if req.Height == 0 {
		req.Height = 600
	}

	// 将数据转换为JSON
	dataJSON, err := json.Marshal(req.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	// 生成Python代码
	code := g.generateChartCode(req, string(dataJSON))

	// 执行Python代码
	output, err := g.runner.Execute(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to execute chart generation: %w", err)
	}

	// 解析输出（base64编码的图片）
	var result struct {
		Image string `json:"image"`
	}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		return nil, fmt.Errorf("failed to parse chart output: %w", err)
	}

	// 解码base64
	image, err := base64.StdEncoding.DecodeString(result.Image)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	return image, nil
}

// generateChartCode 生成图表代码
func (g *ChartGenerator) generateChartCode(req ChartRequest, dataJSON string) string {
	var code string

	switch req.Type {
	case "line":
		code = g.generateLineChart(req, dataJSON)
	case "bar":
		code = g.generateBarChart(req, dataJSON)
	case "scatter":
		code = g.generateScatterChart(req, dataJSON)
	case "pie":
		code = g.generatePieChart(req, dataJSON)
	case "histogram":
		code = g.generateHistogram(req, dataJSON)
	case "box":
		code = g.generateBoxPlot(req, dataJSON)
	default:
		code = g.generateLineChart(req, dataJSON)
	}

	return code
}

// generateLineChart 生成折线图代码
func (g *ChartGenerator) generateLineChart(req ChartRequest, dataJSON string) string {
	title := req.Title
	if title == "" {
		title = "Line Chart"
	}

	xLabel := req.XLabel
	if xLabel == "" && req.X != "" {
		xLabel = req.X
	}

	yLabel := req.YLabel
	if yLabel == "" && req.Y != "" {
		yLabel = req.Y
	}

	return fmt.Sprintf(`
import json
import matplotlib
matplotlib.use('Agg')
import matplotlib.pyplot as plt
import pandas as pd
import io
import base64

data = json.loads('''%s''')
df = pd.DataFrame(data)

plt.figure(figsize=(%d, %d))
plt.plot(df['%s'], df['%s'], marker='o', linewidth=2, markersize=6)
plt.title('%s', fontsize=14, fontweight='bold')
plt.xlabel('%s', fontsize=12)
plt.ylabel('%s', fontsize=12)
plt.grid(True, alpha=0.3)
plt.tight_layout()

buf = io.BytesIO()
plt.savefig(buf, format='png', dpi=100)
buf.seek(0)
img_base64 = base64.b64encode(buf.read()).decode('utf-8')
plt.close()

result = {"image": img_base64}
print(json.dumps(result))
`, dataJSON, req.Width/100, req.Height/100, req.X, req.Y, title, xLabel, yLabel)
}

// generateBarChart 生成柱状图代码
func (g *ChartGenerator) generateBarChart(req ChartRequest, dataJSON string) string {
	title := req.Title
	if title == "" {
		title = "Bar Chart"
	}

	xLabel := req.XLabel
	if xLabel == "" && req.X != "" {
		xLabel = req.X
	}

	yLabel := req.YLabel
	if yLabel == "" && req.Y != "" {
		yLabel = req.Y
	}

	return fmt.Sprintf(`
import json
import matplotlib
matplotlib.use('Agg')
import matplotlib.pyplot as plt
import pandas as pd
import io
import base64

data = json.loads('''%s''')
df = pd.DataFrame(data)

plt.figure(figsize=(%d, %d))
bars = plt.bar(df['%s'], df['%s'], color='steelblue', alpha=0.8)
plt.title('%s', fontsize=14, fontweight='bold')
plt.xlabel('%s', fontsize=12)
plt.ylabel('%s', fontsize=12)
plt.xticks(rotation=45, ha='right')
plt.grid(True, alpha=0.3, axis='y')
plt.tight_layout()

buf = io.BytesIO()
plt.savefig(buf, format='png', dpi=100)
buf.seek(0)
img_base64 = base64.b64encode(buf.read()).decode('utf-8')
plt.close()

result = {"image": img_base64}
print(json.dumps(result))
`, dataJSON, req.Width/100, req.Height/100, req.X, req.Y, title, xLabel, yLabel)
}

// generateScatterChart 生成散点图代码
func (g *ChartGenerator) generateScatterChart(req ChartRequest, dataJSON string) string {
	title := req.Title
	if title == "" {
		title = "Scatter Plot"
	}

	xLabel := req.XLabel
	if xLabel == "" && req.X != "" {
		xLabel = req.X
	}

	yLabel := req.YLabel
	if yLabel == "" && req.Y != "" {
		yLabel = req.Y
	}

	return fmt.Sprintf(`
import json
import matplotlib
matplotlib.use('Agg')
import matplotlib.pyplot as plt
import pandas as pd
import io
import base64

data = json.loads('''%s''')
df = pd.DataFrame(data)

plt.figure(figsize=(%d, %d))
plt.scatter(df['%s'], df['%s'], alpha=0.6, s=50, edgecolors='w', linewidth=0.5)
plt.title('%s', fontsize=14, fontweight='bold')
plt.xlabel('%s', fontsize=12)
plt.ylabel('%s', fontsize=12)
plt.grid(True, alpha=0.3)
plt.tight_layout()

buf = io.BytesIO()
plt.savefig(buf, format='png', dpi=100)
buf.seek(0)
img_base64 = base64.b64encode(buf.read()).decode('utf-8')
plt.close()

result = {"image": img_base64}
print(json.dumps(result))
`, dataJSON, req.Width/100, req.Height/100, req.X, req.Y, title, xLabel, yLabel)
}

// generatePieChart 生成饼图代码
func (g *ChartGenerator) generatePieChart(req ChartRequest, dataJSON string) string {
	title := req.Title
	if title == "" {
		title = "Pie Chart"
	}

	return fmt.Sprintf(`
import json
import matplotlib
matplotlib.use('Agg')
import matplotlib.pyplot as plt
import pandas as pd
import io
import base64

data = json.loads('''%s''')
df = pd.DataFrame(data)

plt.figure(figsize=(%d, %d))
plt.pie(df['%s'], labels=df['%s'], autopct='%%1.1f%%', startangle=90)
plt.title('%s', fontsize=14, fontweight='bold')
plt.axis('equal')
plt.tight_layout()

buf = io.BytesIO()
plt.savefig(buf, format='png', dpi=100)
buf.seek(0)
img_base64 = base64.b64encode(buf.read()).decode('utf-8')
plt.close()

result = {"image": img_base64}
print(json.dumps(result))
`, dataJSON, req.Width/100, req.Height/100, req.Y, req.X, title)
}

// generateHistogram 生成直方图代码
func (g *ChartGenerator) generateHistogram(req ChartRequest, dataJSON string) string {
	title := req.Title
	if title == "" {
		title = "Histogram"
	}

	xLabel := req.XLabel
	if xLabel == "" && req.X != "" {
		xLabel = req.X
	}

	return fmt.Sprintf(`
import json
import matplotlib
matplotlib.use('Agg')
import matplotlib.pyplot as plt
import pandas as pd
import numpy as np
import io
import base64

data = json.loads('''%s''')
df = pd.DataFrame(data)

plt.figure(figsize=(%d, %d))
plt.hist(df['%s'], bins=30, color='steelblue', alpha=0.7, edgecolor='black')
plt.title('%s', fontsize=14, fontweight='bold')
plt.xlabel('%s', fontsize=12)
plt.ylabel('Frequency', fontsize=12)
plt.grid(True, alpha=0.3, axis='y')
plt.tight_layout()

buf = io.BytesIO()
plt.savefig(buf, format='png', dpi=100)
buf.seek(0)
img_base64 = base64.b64encode(buf.read()).decode('utf-8')
plt.close()

result = {"image": img_base64}
print(json.dumps(result))
`, dataJSON, req.Width/100, req.Height/100, req.X, title, xLabel)
}

// generateBoxPlot 生成箱线图代码
func (g *ChartGenerator) generateBoxPlot(req ChartRequest, dataJSON string) string {
	title := req.Title
	if title == "" {
		title = "Box Plot"
	}

	return fmt.Sprintf(`
import json
import matplotlib
matplotlib.use('Agg')
import matplotlib.pyplot as plt
import pandas as pd
import io
import base64

data = json.loads('''%s''')
df = pd.DataFrame(data)

plt.figure(figsize=(%d, %d))
boxprops = dict(linestyle='-', linewidth=2, color='steelblue')
whiskerprops = dict(linestyle='-', linewidth=1.5, color='steelblue')
capprops = dict(linestyle='-', linewidth=1.5, color='steelblue')
medianprops = dict(linestyle='-', linewidth=2, color='red')

df[['%s']].plot(kind='box', boxprops=boxprops, whiskerprops=whiskerprops,
                   capprops=capprops, medianprops=medianprops)
plt.title('%s', fontsize=14, fontweight='bold')
plt.ylabel('Values', fontsize=12)
plt.grid(True, alpha=0.3, axis='y')
plt.tight_layout()

buf = io.BytesIO()
plt.savefig(buf, format='png', dpi=100)
buf.seek(0)
img_base64 = base64.b64encode(buf.read()).decode('utf-8')
plt.close()

result = {"image": img_base64}
print(json.dumps(result))
`, dataJSON, req.Width/100, req.Height/100, req.Y, title)
}

// GenerateMultiple 生成多个图表
func (g *ChartGenerator) GenerateMultiple(ctx context.Context, requests []ChartRequest) ([][]byte, error) {
	images := make([][]byte, len(requests))

	for i, req := range requests {
		img, err := g.Generate(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("failed to generate chart %d: %w", i, err)
		}
		images[i] = img
	}

	return images, nil
}

// GenerateWithTimestamp 生成带时间戳的图表
func (g *ChartGenerator) GenerateWithTimestamp(ctx context.Context, req ChartRequest) ([]byte, string, error) {
	img, err := g.Generate(ctx, req)
	if err != nil {
		return nil, "", err
	}

	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("chart_%s_%s.png", req.Type, timestamp)

	return img, filename, nil
}
