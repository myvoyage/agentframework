// Agent Framework - Enhanced Data Processing Skill
// SPDX-License-Identifier: AGPL-3.0-or-later

package skills

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/schema"
)

// DataProcessingSkill 数据处理技能
// 提供 JSON、XML、CSV 等数据格式的处理功能
type DataProcessingSkill struct {
	*AdvancedSkill
	config *DataProcessingConfig
}

// DataProcessingConfig 数据处理配置
type DataProcessingConfig struct {
	MaxInputSize  int64 // 最大输入大小
	MaxOutputSize int64 // 最大输出大小
	EnableQuery   bool  // 是否启用查询功能
	EnableConvert bool  // 是否启用转换功能
}

// NewDataProcessingSkill 创建新的数据处理技能
func NewDataProcessingSkill(config *DataProcessingConfig) (*DataProcessingSkill, error) {
	if config == nil {
		config = &DataProcessingConfig{
			MaxInputSize:  10 * 1024 * 1024, // 10MB
			MaxOutputSize: 10 * 1024 * 1024,
			EnableQuery:   true,
			EnableConvert: true,
		}
	}

	skill := &DataProcessingSkill{
		config: config,
	}

	skill.AdvancedSkill = NewAdvancedSkill(
		"data_processing",
		"Process and transform data in JSON, XML, CSV, and other formats",
		skill,
	)

	skill.BaseSkill.SetMetadata(SkillMetadata{
		Name:     "data_processing",
		Version:  "2.0.0",
		Category: "data",
		Tags:     []string{"data", "json", "xml", "csv", "transform"},
	})

	return skill, nil
}

// Info 返回技能信息
func (s *DataProcessingSkill) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: s.GetName(),
		Desc: s.GetDescription(),
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"operation": {
				Type:     "string",
				Desc:     "Operation: json_parse, json_stringify, xml_parse, xml_stringify, csv_parse, csv_stringify, transform, query, validate",
				Required: true,
			},
			"data": {
				Type:     "string",
				Desc:     "Input data to process",
				Required: false,
			},
			"query": {
				Type:     "string",
				Desc:     "Query expression (for query operation)",
				Required: false,
			},
			"transform": {
				Type:     "string",
				Desc:     "Transform expression (for transform operation)",
				Required: false,
			},
			"format": {
				Type:     "string",
				Desc:     "Output format: json, xml, csv",
				Required: false,
			},
		}),
	}, nil
}

// Validate 验证输入
func (s *DataProcessingSkill) Validate(ctx context.Context, input string) error {
	var params struct {
		Operation string `json:"operation"`
	}

	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return err
	}

	if params.Operation == "" {
		return fmt.Errorf("operation is required")
	}

	return nil
}

// Prepare 准备执行
func (s *DataProcessingSkill) Prepare(ctx context.Context, input string) (*ExecutionContext, error) {
	execCtx := NewExecutionContext()
	execCtx.SetMetadata("input", input)
	return execCtx, nil
}

// Execute 执行数据处理
func (s *DataProcessingSkill) Execute(ctx context.Context, execCtx *ExecutionContext) (interface{}, error) {
	input, _ := execCtx.GetMetadata("input")
	if input == nil {
		return nil, fmt.Errorf("input metadata not found")
	}

	inputStr, ok := input.(string)
	if !ok {
		return nil, fmt.Errorf("input metadata must be a string")
	}

	var params map[string]interface{}
	if err := json.Unmarshal([]byte(inputStr), &params); err != nil {
		return nil, err
	}

	operation, ok := params["operation"].(string)
	if !ok || operation == "" {
		return nil, fmt.Errorf("operation parameter is required")
	}

	switch operation {
	case "json_parse":
		return s.parseJSON(ctx, params)
	case "json_stringify":
		return s.stringifyJSON(ctx, params)
	case "xml_parse":
		return s.parseXML(ctx, params)
	case "xml_stringify":
		return s.stringifyXML(ctx, params)
	case "csv_parse":
		return s.parseCSV(ctx, params)
	case "csv_stringify":
		return s.stringifyCSV(ctx, params)
	case "transform":
		return s.transformData(ctx, params)
	case "query":
		return s.queryData(ctx, params)
	case "validate":
		return s.validateData(ctx, params)
	default:
		return nil, fmt.Errorf("unknown operation: %s", operation)
	}
}

// parseJSON 解析 JSON
func (s *DataProcessingSkill) parseJSON(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	data := params["data"].(string)

	var result interface{}
	if err := json.Unmarshal([]byte(data), &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return map[string]interface{}{
		"success":   true,
		"operation": "json_parse",
		"result":    result,
		"type":      detectType(result),
	}, nil
}

// stringifyJSON 序列化为 JSON
func (s *DataProcessingSkill) stringifyJSON(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	data := params["data"].(string)

	var input interface{}
	if err := json.Unmarshal([]byte(data), &input); err != nil {
		return nil, fmt.Errorf("failed to parse input: %w", err)
	}

	output, err := json.MarshalIndent(input, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to stringify: %w", err)
	}

	return map[string]interface{}{
		"success":   true,
		"operation": "json_stringify",
		"result":    string(output),
		"size":      len(output),
	}, nil
}

// parseXML 解析 XML
func (s *DataProcessingSkill) parseXML(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	data := params["data"].(string)

	var result interface{}
	if err := xml.Unmarshal([]byte(data), &result); err != nil {
		return nil, fmt.Errorf("failed to parse XML: %w", err)
	}

	return map[string]interface{}{
		"success":   true,
		"operation": "xml_parse",
		"result":    result,
	}, nil
}

// stringifyXML 序列化为 XML
func (s *DataProcessingSkill) stringifyXML(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	data := params["data"].(string)

	var input interface{}
	if err := json.Unmarshal([]byte(data), &input); err != nil {
		return nil, fmt.Errorf("failed to parse input: %w", err)
	}

	output, err := xml.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("failed to stringify XML: %w", err)
	}

	return map[string]interface{}{
		"success":   true,
		"operation": "xml_stringify",
		"result":    string(output),
		"size":      len(output),
	}, nil
}

// parseCSV 解析 CSV
func (s *DataProcessingSkill) parseCSV(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	data := params["data"].(string)

	reader := csv.NewReader(strings.NewReader(data))
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV: %w", err)
	}

	// 转换为 JSON 友好格式
	var headers []string
	if len(records) > 0 {
		headers = records[0]
	}

	var rows []map[string]string
	if len(records) > 1 {
		for _, record := range records[1:] {
			row := make(map[string]string)
			for i, value := range record {
				if i < len(headers) {
					row[headers[i]] = value
				}
			}
			rows = append(rows, row)
		}
	}

	return map[string]interface{}{
		"success":   true,
		"operation": "csv_parse",
		"headers":   headers,
		"rows":      rows,
		"count":     len(rows),
	}, nil
}

// stringifyCSV 序列化为 CSV
func (s *DataProcessingSkill) stringifyCSV(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	data := params["data"].(string)

	var input struct {
		Headers []string            `json:"headers"`
		Rows    []map[string]string `json:"rows"`
	}

	if err := json.Unmarshal([]byte(data), &input); err != nil {
		return nil, fmt.Errorf("failed to parse input: %w", err)
	}

	var builder strings.Builder
	writer := csv.NewWriter(&builder)

	// 写入表头
	if err := writer.Write(input.Headers); err != nil {
		return nil, fmt.Errorf("failed to write headers: %w", err)
	}

	// 写入数据行
	for _, row := range input.Rows {
		var record []string
		for _, header := range input.Headers {
			record = append(record, row[header])
		}
		if err := writer.Write(record); err != nil {
			return nil, fmt.Errorf("failed to write row: %w", err)
		}
	}

	writer.Flush()

	return map[string]interface{}{
		"success":   true,
		"operation": "csv_stringify",
		"result":    builder.String(),
		"rows":      len(input.Rows) + 1,
	}, nil
}

// transformData 转换数据
func (s *DataProcessingSkill) transformData(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	data := params["data"].(string)
	transformExpr := params["transform"].(string)

	// 简单实现：支持基本的数据转换
	// 格式：字段操作（如 uppercase(name), lowercase(name), concat(a,b)）

	var input interface{}
	if err := json.Unmarshal([]byte(data), &input); err != nil {
		return nil, fmt.Errorf("failed to parse input: %w", err)
	}

	result := input

	// 应用转换（这里提供简单的示例）
	switch strings.ToLower(transformExpr) {
	case "uppercase":
		if str, ok := input.(string); ok {
			result = strings.ToUpper(str)
		}
	case "lowercase":
		if str, ok := input.(string); ok {
			result = strings.ToLower(str)
		}
	case "reverse":
		if str, ok := input.(string); ok {
			result = reverseString(str)
		}
	default:
		return nil, fmt.Errorf("unsupported transform: %s", transformExpr)
	}

	return map[string]interface{}{
		"success":   true,
		"operation": "transform",
		"transform": transformExpr,
		"result":    result,
	}, nil
}

// queryData 查询数据
func (s *DataProcessingSkill) queryData(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	data := params["data"].(string)
	queryExpr := params["query"].(string)

	var input interface{}
	if err := json.Unmarshal([]byte(data), &input); err != nil {
		return nil, fmt.Errorf("failed to parse input: %w", err)
	}

	// 简单的 JSONPath 查询实现
	result, err := s.queryJSON(input, queryExpr)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}

	return map[string]interface{}{
		"success":   true,
		"operation": "query",
		"query":     queryExpr,
		"result":    result,
	}, nil
}

// validateData 验证数据
func (s *DataProcessingSkill) validateData(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	data := params["data"].(string)
	format := ""
	if f, ok := params["format"].(string); ok {
		format = f
	}

	var input interface{}
	var err error

	switch strings.ToLower(format) {
	case "json", "":
		err = json.Unmarshal([]byte(data), &input)
	case "xml":
		err = xml.Unmarshal([]byte(data), &input)
	default:
		// 尝试自动检测
		if strings.HasPrefix(strings.TrimSpace(data), "{") ||
			strings.HasPrefix(strings.TrimSpace(data), "[") {
			err = json.Unmarshal([]byte(data), &input)
		} else {
			err = xml.Unmarshal([]byte(data), &input)
		}
	}

	valid := err == nil

	return map[string]interface{}{
		"success":   valid,
		"operation": "validate",
		"format":    format,
		"valid":     valid,
		"error": func() string {
			if err != nil {
				return err.Error()
			}
			return ""
		}(),
	}, nil
}

// queryJSON 简单的 JSONPath 查询
func (s *DataProcessingSkill) queryJSON(data interface{}, query string) (interface{}, error) {
	// 简化实现：支持点号分隔的路径查询
	parts := strings.Split(query, ".")

	current := data
	for _, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			var ok bool
			current, ok = v[part]
			if !ok {
				return nil, fmt.Errorf("key not found: %s", part)
			}
		case []interface{}:
			// 支持数组索引
			idx := 0
			if _, err := fmt.Sscanf(part, "[%d]", &idx); err == nil {
				if idx >= 0 && idx < len(v) {
					current = v[idx]
				} else {
					return nil, fmt.Errorf("index out of range: %d", idx)
				}
			} else {
				return nil, fmt.Errorf("invalid array index: %s", part)
			}
		default:
			return nil, fmt.Errorf("cannot query into non-object type")
		}
	}

	return current, nil
}

// detectType 检测数据类型
func detectType(data interface{}) string {
	switch data.(type) {
	case map[string]interface{}:
		return "object"
	case []interface{}:
		return "array"
	case string:
		return "string"
	case float64:
		return "number"
	case bool:
		return "boolean"
	case nil:
		return "null"
	default:
		return "unknown"
	}
}

// reverseString 反转字符串
func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// Cleanup 清理资源
func (s *DataProcessingSkill) Cleanup(ctx context.Context, execCtx *ExecutionContext) error {
	return nil
}
