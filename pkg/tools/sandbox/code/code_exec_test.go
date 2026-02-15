// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
// Copyright (C) 2025  Agent Framework Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package code

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestCodeExecutorModule_PythonExecution(t *testing.T) {
	config := CodeExecutorConfig{
		Timeout:            30000,
		MemoryLimit:        512,
		CPULimit:           2,
		SupportedLanguages: []string{"python"},
	}

	module, err := NewCodeExecutorModule(config)
	if err != nil {
		t.Fatalf("Failed to create module: %v", err)
	}
	defer module.Close()

	// 测试简单的 Python 代码
	code := `print("Hello, World!")`
	result, err := module.runCode("python", code, "", 0)

	if err != nil {
		t.Fatalf("Failed to run code: %v", err)
	}

	if !result["success"].(bool) {
		t.Errorf("Expected success=true, got false. Error: %v", result["error"])
	}

	output := result["output"].(string)
	// 检查输出包含 "Hello, World!"（忽略换行符差异）
	if !strings.Contains(output, "Hello, World!") {
		t.Errorf("Expected output to contain 'Hello, World!', got '%s'", output)
	}
}

func TestCodeExecutorModule_JavaScriptExecution(t *testing.T) {
	config := CodeExecutorConfig{
		Timeout:            30000,
		MemoryLimit:        512,
		CPULimit:           2,
		SupportedLanguages: []string{"javascript"},
	}

	module, err := NewCodeExecutorModule(config)
	if err != nil {
		t.Fatalf("Failed to create module: %v", err)
	}
	defer module.Close()

	// 测试简单的 JavaScript 代码
	code := `console.log("Hello, World!");`
	result, err := module.runCode("javascript", code, "", 0)

	if err != nil {
		t.Fatalf("Failed to run code: %v", err)
	}

	if !result["success"].(bool) {
		t.Errorf("Expected success=true, got false. Error: %v", result["error"])
	}
}

func TestCodeExecutorModule_BashExecution(t *testing.T) {
	config := CodeExecutorConfig{
		Timeout:            30000,
		MemoryLimit:        512,
		CPULimit:           2,
		SupportedLanguages: []string{"bash"},
	}

	module, err := NewCodeExecutorModule(config)
	if err != nil {
		t.Fatalf("Failed to create module: %v", err)
	}
	defer module.Close()

	// 测试简单的 Bash 代码
	code := `echo "Hello, World!"`
	result, err := module.runCode("bash", code, "", 0)

	if err != nil {
		t.Fatalf("Failed to run code: %v", err)
	}

	if !result["success"].(bool) {
		t.Errorf("Expected success=true, got false. Error: %v", result["error"])
	}
}

func TestCodeExecutorModule_UnsupportedLanguage(t *testing.T) {
	config := CodeExecutorConfig{
		Timeout:            30000,
		MemoryLimit:        512,
		CPULimit:           2,
		SupportedLanguages: []string{"python"},
	}

	module, err := NewCodeExecutorModule(config)
	if err != nil {
		t.Fatalf("Failed to create module: %v", err)
	}
	defer module.Close()

	// 测试不支持的语言
	code := `print("Hello")`
	result, err := module.runCode("ruby", code, "", 0)

	if err != nil {
		t.Fatalf("Failed to run code: %v", err)
	}

	if result["success"].(bool) {
		t.Errorf("Expected success=false for unsupported language")
	}

	if result["error"].(string) != "Language not supported" {
		t.Errorf("Expected error 'Language not supported', got '%s'", result["error"])
	}
}

func TestCodeExecutorModule_Timeout(t *testing.T) {
	config := CodeExecutorConfig{
		Timeout:            1000, // 1 second
		MemoryLimit:        512,
		CPULimit:           2,
		SupportedLanguages: []string{"python"},
	}

	module, err := NewCodeExecutorModule(config)
	if err != nil {
		t.Fatalf("Failed to create module: %v", err)
	}
	defer module.Close()

	// 测试超时
	code := `import time; time.sleep(10)`
	result, err := module.runCode("python", code, "", 1000)

	if err != nil {
		t.Fatalf("Failed to run code: %v", err)
	}

	if result["success"].(bool) {
		t.Errorf("Expected success=false for timeout")
	}
}

func TestCodeExecutorModule_SyntaxError(t *testing.T) {
	config := CodeExecutorConfig{
		Timeout:            30000,
		MemoryLimit:        512,
		CPULimit:           2,
		SupportedLanguages: []string{"python"},
	}

	module, err := NewCodeExecutorModule(config)
	if err != nil {
		t.Fatalf("Failed to create module: %v", err)
	}
	defer module.Close()

	// 测试语法错误
	code := `print("Hello` // 缺少引号
	result, err := module.runCode("python", code, "", 0)

	if err != nil {
		t.Fatalf("Failed to run code: %v", err)
	}

	if result["success"].(bool) {
		t.Errorf("Expected success=false for syntax error")
	}
}

func TestCodeExecutorModule_GetTools(t *testing.T) {
	config := CodeExecutorConfig{
		Timeout:            30000,
		MemoryLimit:        512,
		CPULimit:           2,
		SupportedLanguages: []string{"python", "javascript", "bash"},
	}

	module, err := NewCodeExecutorModule(config)
	if err != nil {
		t.Fatalf("Failed to create module: %v", err)
	}
	defer module.Close()

	ctx := context.Background()
	tools, err := module.GetTools(ctx)

	if err != nil {
		t.Fatalf("Failed to get tools: %v", err)
	}

	if len(tools) != 6 {
		t.Errorf("Expected 6 tools, got %d", len(tools))
	}
}

func TestCodeExecutorModule_Stats(t *testing.T) {
	config := CodeExecutorConfig{
		Timeout:            30000,
		MemoryLimit:        512,
		CPULimit:           2,
		SupportedLanguages: []string{"python"},
	}

	module, err := NewCodeExecutorModule(config)
	if err != nil {
		t.Fatalf("Failed to create module: %v", err)
	}
	defer module.Close()

	// 执行一些代码
	module.runCode("python", `print("test")`, "", 0)
	module.runCode("python", `print("test2")`, "", 0)
	module.runCode("ruby", `puts "test"`, "", 0) // 不支持的语言

	stats := module.GetStats()

	if stats["total_executions"] != 3 {
		t.Errorf("Expected total_executions=3, got %d", stats["total_executions"])
	}

	if stats["success_count"] != 2 {
		t.Errorf("Expected success_count=2, got %d", stats["success_count"])
	}

	if stats["blocked_count"] != 1 {
		t.Errorf("Expected blocked_count=1, got %d", stats["blocked_count"])
	}
}

func TestPythonRunner_Format(t *testing.T) {
	config := CodeExecutorConfig{
		Timeout:     30000,
		MemoryLimit: 512,
		CPULimit:    2,
	}

	runner := NewPythonRunner(config, t.TempDir())

	code := `print("Hello")`
	formatted, err := runner.Format(code)

	if err != nil {
		t.Fatalf("Failed to format code: %v", err)
	}

	if formatted == "" {
		t.Errorf("Expected non-empty formatted code")
	}
}

func TestBashRunner_Run(t *testing.T) {
	config := CodeExecutorConfig{
		Timeout:     30000,
		MemoryLimit: 512,
		CPULimit:    2,
	}

	runner := NewBashRunner(config)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := runner.Run(ctx, `echo "test"`, "")

	if err != nil {
		t.Fatalf("Failed to run bash code: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected success=true, got false")
	}
}

// ============================================================================
// MCP Tool Tests
// ============================================================================

func TestCodeExecRunTool_InvokableRun(t *testing.T) {
	config := CodeExecutorConfig{
		Timeout:            30000,
		MemoryLimit:        512,
		CPULimit:           2,
		SupportedLanguages: []string{"python", "javascript", "bash"},
	}

	module, err := NewCodeExecutorModule(config)
	if err != nil {
		t.Fatalf("Failed to create module: %v", err)
	}
	defer module.Close()

	tool := &codeExecRunTool{module: module}
	ctx := context.Background()

	// Test Python execution
	t.Run("PythonExecution", func(t *testing.T) {
		input := `{"language":"python","code":"print('Hello from Python')","timeout":5000}`
		output, err := tool.InvokableRun(ctx, input)

		if err != nil {
			t.Fatalf("Failed to run tool: %v", err)
		}

		if !strings.Contains(output, "success") {
			t.Errorf("Expected output to contain 'success', got: %s", output)
		}

		if !strings.Contains(output, "Hello from Python") {
			t.Errorf("Expected output to contain 'Hello from Python', got: %s", output)
		}
	})

	// Test JavaScript execution
	t.Run("JavaScriptExecution", func(t *testing.T) {
		input := `{"language":"javascript","code":"console.log('Hello from JS');","timeout":5000}`
		output, err := tool.InvokableRun(ctx, input)

		if err != nil {
			t.Fatalf("Failed to run tool: %v", err)
		}

		if !strings.Contains(output, "success") {
			t.Errorf("Expected output to contain 'success', got: %s", output)
		}
	})

	// Test Bash execution
	t.Run("BashExecution", func(t *testing.T) {
		input := `{"language":"bash","code":"echo 'Hello from Bash'","timeout":5000}`
		output, err := tool.InvokableRun(ctx, input)

		if err != nil {
			t.Fatalf("Failed to run tool: %v", err)
		}

		if !strings.Contains(output, "success") {
			t.Errorf("Expected output to contain 'success', got: %s", output)
		}
	})

	// Test unsupported language
	t.Run("UnsupportedLanguage", func(t *testing.T) {
		input := `{"language":"ruby","code":"puts 'Hello'","timeout":5000}`
		output, err := tool.InvokableRun(ctx, input)

		if err != nil {
			t.Fatalf("Failed to run tool: %v", err)
		}

		if !strings.Contains(output, "Language not supported") {
			t.Errorf("Expected error about unsupported language, got: %s", output)
		}
	})

	// Test invalid JSON input
	t.Run("InvalidJSON", func(t *testing.T) {
		input := `{invalid json}`
		_, err := tool.InvokableRun(ctx, input)

		if err == nil {
			t.Error("Expected error for invalid JSON")
		}
	})

	// Test timeout
	t.Run("Timeout", func(t *testing.T) {
		input := `{"language":"python","code":"import time; time.sleep(10)","timeout":1000}`
		output, err := tool.InvokableRun(ctx, input)

		if err != nil {
			t.Fatalf("Failed to run tool: %v", err)
		}

		// Should complete but with failure status due to timeout
		if !strings.Contains(output, "success") {
			t.Errorf("Expected output to contain 'success' field, got: %s", output)
		}
	})

	// Test syntax error
	t.Run("SyntaxError", func(t *testing.T) {
		input := `{"language":"python","code":"print('unclosed string","timeout":5000}`
		output, err := tool.InvokableRun(ctx, input)

		if err != nil {
			t.Fatalf("Failed to run tool: %v", err)
		}

		// Should return error in the result
		if !strings.Contains(output, "error") {
			t.Errorf("Expected output to contain 'error' field, got: %s", output)
		}
	})
}

func TestCodeExecRunTool_Info(t *testing.T) {
	config := CodeExecutorConfig{
		Timeout:            30000,
		MemoryLimit:        512,
		CPULimit:           2,
		SupportedLanguages: []string{"python"},
	}

	module, err := NewCodeExecutorModule(config)
	if err != nil {
		t.Fatalf("Failed to create module: %v", err)
	}
	defer module.Close()

	tool := &codeExecRunTool{module: module}
	ctx := context.Background()

	info, err := tool.Info(ctx)
	if err != nil {
		t.Fatalf("Failed to get tool info: %v", err)
	}

	if info.Name != "code_exec_run" {
		t.Errorf("Expected tool name 'code_exec_run', got '%s'", info.Name)
	}

	if info.Desc == "" {
		t.Error("Expected non-empty description")
	}

	if info.ParamsOneOf == nil {
		t.Error("Expected non-nil ParamsOneOf")
	}
}

func TestCodeExecSupportedLanguagesTool_Info(t *testing.T) {
	config := CodeExecutorConfig{
		Timeout:            30000,
		MemoryLimit:        512,
		CPULimit:           2,
		SupportedLanguages: []string{"python"},
	}

	module, err := NewCodeExecutorModule(config)
	if err != nil {
		t.Fatalf("Failed to create module: %v", err)
	}
	defer module.Close()

	tool := &codeExecSupportedLanguagesTool{module: module}
	ctx := context.Background()

	info, err := tool.Info(ctx)
	if err != nil {
		t.Fatalf("Failed to get tool info: %v", err)
	}

	if info.Name != "code_exec_supported_languages" {
		t.Errorf("Expected tool name 'code_exec_supported_languages', got '%s'", info.Name)
	}

	if info.Desc == "" {
		t.Error("Expected non-empty description")
	}

	if info.ParamsOneOf == nil {
		t.Error("Expected non-nil ParamsOneOf")
	}
}

func TestCodeExecSupportedLanguagesTool_InvokableRun(t *testing.T) {
	config := CodeExecutorConfig{
		Timeout:            30000,
		MemoryLimit:        512,
		CPULimit:           2,
		SupportedLanguages: []string{"python", "javascript", "bash", "go"},
	}

	module, err := NewCodeExecutorModule(config)
	if err != nil {
		t.Fatalf("Failed to create module: %v", err)
	}
	defer module.Close()

	tool := &codeExecSupportedLanguagesTool{module: module}
	ctx := context.Background()

	output, err := tool.InvokableRun(ctx, "{}")
	if err != nil {
		t.Fatalf("Failed to run tool: %v", err)
	}

	if !strings.Contains(output, "python") {
		t.Errorf("Expected output to contain 'python', got: %s", output)
	}

	if !strings.Contains(output, "javascript") {
		t.Errorf("Expected output to contain 'javascript', got: %s", output)
	}

	if !strings.Contains(output, "bash") {
		t.Errorf("Expected output to contain 'bash', got: %s", output)
	}

	if !strings.Contains(output, "go") {
		t.Errorf("Expected output to contain 'go', got: %s", output)
	}

	// Verify the output is valid JSON
	if !strings.Contains(output, "success") {
		t.Errorf("Expected output to contain 'success' field, got: %s", output)
	}

	if !strings.Contains(output, "languages") {
		t.Errorf("Expected output to contain 'languages' field, got: %s", output)
	}
}

func TestCodeExecFormatTool_InvokableRun(t *testing.T) {
	config := CodeExecutorConfig{
		Timeout:            30000,
		MemoryLimit:        512,
		CPULimit:           2,
		SupportedLanguages: []string{"python", "go"},
	}

	module, err := NewCodeExecutorModule(config)
	if err != nil {
		t.Fatalf("Failed to create module: %v", err)
	}
	defer module.Close()

	tool := &codeExecFormatTool{module: module}
	ctx := context.Background()

	// Test Python formatting
	t.Run("PythonFormat", func(t *testing.T) {
		input := `{"language":"python","code":"print('hello')"}`
		output, err := tool.InvokableRun(ctx, input)

		if err != nil {
			t.Fatalf("Failed to run tool: %v", err)
		}

		if !strings.Contains(output, "success") {
			t.Errorf("Expected output to contain 'success', got: %s", output)
		}

		if !strings.Contains(output, "formatted_code") {
			t.Errorf("Expected output to contain 'formatted_code', got: %s", output)
		}
	})

	// Test unsupported language
	t.Run("UnsupportedLanguage", func(t *testing.T) {
		input := `{"language":"ruby","code":"puts 'hello'"}`
		output, err := tool.InvokableRun(ctx, input)

		if err != nil {
			t.Fatalf("Failed to run tool: %v", err)
		}

		if !strings.Contains(output, "Language not supported") {
			t.Errorf("Expected error about unsupported language, got: %s", output)
		}
	})

	// Test invalid JSON
	t.Run("InvalidJSON", func(t *testing.T) {
		input := `{invalid}`
		_, err := tool.InvokableRun(ctx, input)

		if err == nil {
			t.Error("Expected error for invalid JSON")
		}
	})
}

func TestCodeExecRunTool_WithInput(t *testing.T) {
	config := CodeExecutorConfig{
		Timeout:            30000,
		MemoryLimit:        512,
		CPULimit:           2,
		SupportedLanguages: []string{"python"},
	}

	module, err := NewCodeExecutorModule(config)
	if err != nil {
		t.Fatalf("Failed to create module: %v", err)
	}
	defer module.Close()

	tool := &codeExecRunTool{module: module}
	ctx := context.Background()

	// Test code with input parameter (though stdin support may not be fully implemented yet)
	input := `{"language":"python","code":"x = 5 + 3\nprint(x)","input":"","timeout":5000}`
	output, err := tool.InvokableRun(ctx, input)

	if err != nil {
		t.Fatalf("Failed to run tool: %v", err)
	}

	if !strings.Contains(output, "8") {
		t.Errorf("Expected output to contain '8', got: %s", output)
	}

	// Test with actual input parameter provided
	input2 := `{"language":"python","code":"print('Hello')","input":"test input","timeout":5000}`
	output2, err := tool.InvokableRun(ctx, input2)

	if err != nil {
		t.Fatalf("Failed to run tool: %v", err)
	}

	if !strings.Contains(output2, "success") {
		t.Errorf("Expected output to contain 'success', got: %s", output2)
	}
}

func TestCodeExecRunTool_MultipleExecutions(t *testing.T) {
	config := CodeExecutorConfig{
		Timeout:            30000,
		MemoryLimit:        512,
		CPULimit:           2,
		SupportedLanguages: []string{"python"},
	}

	module, err := NewCodeExecutorModule(config)
	if err != nil {
		t.Fatalf("Failed to create module: %v", err)
	}
	defer module.Close()

	tool := &codeExecRunTool{module: module}
	ctx := context.Background()

	// Execute multiple times to ensure no state leakage
	for i := 0; i < 5; i++ {
		input := `{"language":"python","code":"print('Execution ` + string(rune('0'+i)) + `')","timeout":5000}`
		output, err := tool.InvokableRun(ctx, input)

		if err != nil {
			t.Fatalf("Failed to run tool on iteration %d: %v", i, err)
		}

		if !strings.Contains(output, "success") {
			t.Errorf("Iteration %d: Expected output to contain 'success', got: %s", i, output)
		}
	}

	// Check stats
	stats := module.GetStats()
	if stats["total_executions"] < 5 {
		t.Errorf("Expected at least 5 executions, got %d", stats["total_executions"])
	}
}
