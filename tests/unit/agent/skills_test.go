// Agent Framework - Skills Test Suite
// SPDX-License-Identifier: AGPL-3.0-or-later

package skills

import (
	"context"
	"testing"
	"time"
)

// TestFileOperationSkill 测试文件操作技能
func TestFileOperationSkill(t *testing.T) {
	config := &FileOperationConfig{
		SandboxDir:        "./test_workspace",
		AllowedPaths:      []string{"./test_workspace"},
		MaxFileSize:       1024 * 1024,
		AllowedExts:       []string{".txt", ".json", ".md"},
		EnableSearch:      true,
		EnableCompression: false,
	}

	skill, err := NewFileOperationSkill(config)
	if err != nil {
		t.Fatalf("Failed to create skill: %v", err)
	}

	ctx := context.Background()

	// 测试 Info
	info, err := skill.Info(ctx)
	if err != nil {
		t.Fatalf("Failed to get info: %v", err)
	}
	if info.Name != "file_operation" {
		t.Errorf("Expected name 'file_operation', got '%s'", info.Name)
	}

	// 测试写文件
	input := `{
		"operation": "write",
		"path": "test.txt",
		"content": "Hello, World!"
	}`

	result, err := skill.Invoke(ctx, input)
	if err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	t.Logf("Write result: %s", result)

	// 测试读文件
	input = `{
		"operation": "read",
		"path": "test.txt"
	}`

	result, err = skill.Invoke(ctx, input)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	t.Logf("Read result: %s", result)

	// 测试文件存在
	input = `{
		"operation": "exists",
		"path": "test.txt"
	}`

	result, err = skill.Invoke(ctx, input)
	if err != nil {
		t.Fatalf("Failed to check exists: %v", err)
	}

	t.Logf("Exists result: %s", result)

	// 测试删除文件
	input = `{
		"operation": "delete",
		"path": "test.txt"
	}`

	result, err = skill.Invoke(ctx, input)
	if err != nil {
		t.Fatalf("Failed to delete file: %v", err)
	}

	t.Logf("Delete result: %s", result)
}

// TestCodeExecutionSkill 测试代码执行技能
func TestCodeExecutionSkill(t *testing.T) {
	config := &CodeExecutionConfig{
		AllowedLanguages: []string{"bash", "python"},
		MaxExecutionTime: 5 * time.Second,
		MaxMemoryMB:      256,
		MaxOutputSize:    1024 * 1024,
		EnableSandbox:    true,
		TempDir:          "/tmp/agent_test",
		AllowedCommands:  []string{"echo", "ls", "pwd", "date"},
	}

	skill, err := NewCodeExecutionSkill(config)
	if err != nil {
		t.Fatalf("Failed to create skill: %v", err)
	}

	ctx := context.Background()

	// 测试 echo 命令
	input := `{
		"command": "echo",
		"args": ["Hello", "World"]
	}`

	result, err := skill.Invoke(ctx, input)
	if err != nil {
		t.Fatalf("Failed to execute echo: %v", err)
	}

	t.Logf("Echo result: %s", result)

	// 测试 pwd 命令
	input = `{
		"command": "pwd"
	}`

	result, err = skill.Invoke(ctx, input)
	if err != nil {
		t.Fatalf("Failed to execute pwd: %v", err)
	}

	t.Logf("PWD result: %s", result)
}

// TestDataProcessingSkill 测试数据处理技能
func TestDataProcessingSkill(t *testing.T) {
	config := &DataProcessingConfig{
		MaxInputSize:  1024 * 1024,
		MaxOutputSize: 1024 * 1024,
		EnableQuery:   true,
		EnableConvert: true,
	}

	skill, err := NewDataProcessingSkill(config)
	if err != nil {
		t.Fatalf("Failed to create skill: %v", err)
	}

	ctx := context.Background()

	// 测试 JSON 解析
	input := `{
		"operation": "json_parse",
		"data": "{\"name\":\"John\",\"age\":30}"
	}`

	result, err := skill.Invoke(ctx, input)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	t.Logf("JSON parse result: %s", result)

	// 测试 JSON 序列化
	input = `{
		"operation": "json_stringify",
		"data": "{\"name\":\"John\",\"age\":30}"
	}`

	result, err = skill.Invoke(ctx, input)
	if err != nil {
		t.Fatalf("Failed to stringify JSON: %v", err)
	}

	t.Logf("JSON stringify result: %s", result)

	// 测试数据转换
	input = `{
		"operation": "transform",
		"data": "hello world",
		"transform": "uppercase"
	}`

	result, err = skill.Invoke(ctx, input)
	if err != nil {
		t.Fatalf("Failed to transform data: %v", err)
	}

	t.Logf("Transform result: %s", result)

	// 测试 CSV 解析
	input = `{
		"operation": "csv_parse",
		"data": "name,age\nJohn,30\nJane,25"
	}`

	result, err = skill.Invoke(ctx, input)
	if err != nil {
		t.Fatalf("Failed to parse CSV: %v", err)
	}

	t.Logf("CSV parse result: %s", result)
}

// TestSkillStats 测试技能统计
func TestSkillStats(t *testing.T) {
	config := &FileOperationConfig{
		SandboxDir:   "./test_workspace",
		AllowedPaths: []string{"./test_workspace"},
	}

	skill, _ := NewFileOperationSkill(config)

	stats := skill.GetStats()
	if stats == nil {
		t.Fatal("Stats is nil")
	}

	t.Logf("Initial stats - Total: %d, Success: %d, Failed: %d",
		stats.TotalCalls, stats.SuccessCalls, stats.FailedCalls)
}

// BenchmarkFileOperation 性能测试
func BenchmarkFileOperation(b *testing.B) {
	config := &FileOperationConfig{
		SandboxDir:   "./bench_workspace",
		AllowedPaths: []string{"./bench_workspace"},
	}

	skill, _ := NewFileOperationSkill(config)
	ctx := context.Background()

	input := `{
		"operation": "write",
		"path": "bench.txt",
		"content": "Benchmark test"
	}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		skill.Invoke(ctx, input)
	}
}

// BenchmarkDataProcessing 性能测试
func BenchmarkDataProcessing(b *testing.B) {
	skill, _ := NewDataProcessingSkill(nil)
	ctx := context.Background()

	input := `{
		"operation": "json_parse",
		"data": "{\"test\":\"data\"}"
	}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		skill.Invoke(ctx, input)
	}
}
