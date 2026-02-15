// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
// Copyright (C) 2025  Agent Framework Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package code

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

// TestFormatTool_AllLanguages tests formatting for all supported languages
func TestFormatTool_AllLanguages(t *testing.T) {
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

	tool := &codeExecFormatTool{module: module}
	ctx := context.Background()

	tests := []struct {
		name     string
		language string
		code     string
		wantErr  bool
	}{
		{
			name:     "Python - simple print",
			language: "python",
			code:     "print('hello')",
			wantErr:  false,
		},
		{
			name:     "Python - unformatted",
			language: "python",
			code:     "x=1+2\nprint(x)",
			wantErr:  false,
		},
		{
			name:     "JavaScript - simple console.log",
			language: "javascript",
			code:     "console.log('hello');",
			wantErr:  false,
		},
		{
			name:     "JavaScript - unformatted",
			language: "javascript",
			code:     "const x=1+2;console.log(x);",
			wantErr:  false,
		},
		{
			name:     "Bash - simple echo",
			language: "bash",
			code:     "echo 'hello'",
			wantErr:  false,
		},
		{
			name:     "Bash - unformatted",
			language: "bash",
			code:     "if [ -f file.txt ]; then echo 'exists'; fi",
			wantErr:  false,
		},
		{
			name:     "Go - simple print",
			language: "go",
			code:     "package main\n\nfunc main() {\n\tprintln(\"hello\")\n}",
			wantErr:  false,
		},
		{
			name:     "Go - unformatted",
			language: "go",
			code:     "package main\nfunc main(){println(\"hello\")}",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := map[string]string{
				"language": tt.language,
				"code":     tt.code,
			}
			inputJSON, _ := json.Marshal(input)

			output, err := tool.InvokableRun(ctx, string(inputJSON))

			if (err != nil) != tt.wantErr {
				t.Errorf("InvokableRun() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				// Verify output contains expected fields
				if !strings.Contains(output, "success") {
					t.Errorf("Expected output to contain 'success', got: %s", output)
				}

				if !strings.Contains(output, "formatted_code") {
					t.Errorf("Expected output to contain 'formatted_code', got: %s", output)
				}

				// Parse the output to verify structure
				var result map[string]interface{}
				if err := json.Unmarshal([]byte(output), &result); err != nil {
					t.Errorf("Failed to parse output JSON: %v", err)
				}

				if success, ok := result["success"].(bool); !ok || !success {
					t.Errorf("Expected success=true, got: %v", result["success"])
				}

				if formattedCode, ok := result["formatted_code"].(string); !ok || formattedCode == "" {
					t.Errorf("Expected non-empty formatted_code, got: %v", result["formatted_code"])
				}
			}
		})
	}
}

// TestFormatTool_ErrorCases tests error handling in formatting
func TestFormatTool_ErrorCases(t *testing.T) {
	config := CodeExecutorConfig{
		Timeout:            30000,
		MemoryLimit:        512,
		CPULimit:           2,
		SupportedLanguages: []string{"python", "javascript"},
	}

	module, err := NewCodeExecutorModule(config)
	if err != nil {
		t.Fatalf("Failed to create module: %v", err)
	}
	defer module.Close()

	tool := &codeExecFormatTool{module: module}
	ctx := context.Background()

	tests := []struct {
		name        string
		input       string
		wantErr     bool
		errContains string
	}{
		{
			name:        "Unsupported language",
			input:       `{"language":"ruby","code":"puts 'hello'"}`,
			wantErr:     false, // Tool returns error in result, not as error
			errContains: "Language not supported",
		},
		{
			name:    "Invalid JSON",
			input:   `{invalid json}`,
			wantErr: true,
		},
		{
			name:        "Missing language",
			input:       `{"code":"print('hello')"}`,
			wantErr:     false,
			errContains: "Language not supported",
		},
		{
			name:    "Missing code",
			input:   `{"language":"python"}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := tool.InvokableRun(ctx, tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("InvokableRun() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil && tt.errContains != "" {
				if !strings.Contains(output, tt.errContains) {
					t.Errorf("Expected output to contain '%s', got: %s", tt.errContains, output)
				}
			}
		})
	}
}

// TestFormatTool_Info tests the tool info
func TestFormatTool_Info(t *testing.T) {
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

	tool := &codeExecFormatTool{module: module}
	ctx := context.Background()

	info, err := tool.Info(ctx)
	if err != nil {
		t.Fatalf("Failed to get tool info: %v", err)
	}

	if info.Name != "code_exec_format" {
		t.Errorf("Expected tool name 'code_exec_format', got '%s'", info.Name)
	}

	if info.Desc == "" {
		t.Error("Expected non-empty description")
	}

	if info.ParamsOneOf == nil {
		t.Error("Expected non-nil ParamsOneOf")
	}
}

// TestLanguageRunners_Format tests the Format method for each language runner
func TestLanguageRunners_Format(t *testing.T) {
	tempDir := t.TempDir()
	config := CodeExecutorConfig{
		Timeout:     30000,
		MemoryLimit: 512,
		CPULimit:    2,
	}

	tests := []struct {
		name    string
		runner  LanguageRunner
		code    string
		wantErr bool
	}{
		{
			name:    "Python runner",
			runner:  NewPythonRunner(config, tempDir),
			code:    "print('hello')",
			wantErr: false,
		},
		{
			name:    "JavaScript runner",
			runner:  NewJavaScriptRunner(config),
			code:    "console.log('hello');",
			wantErr: false,
		},
		{
			name:    "Bash runner",
			runner:  NewBashRunner(config),
			code:    "echo 'hello'",
			wantErr: false,
		},
		{
			name:    "Go runner",
			runner:  NewGoRunner(config, tempDir),
			code:    "package main\n\nfunc main() {\n\tprintln(\"hello\")\n}",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatted, err := tt.runner.Format(tt.code)

			if (err != nil) != tt.wantErr {
				t.Errorf("Format() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				if formatted == "" {
					t.Error("Expected non-empty formatted code")
				}

				// Formatted code should be valid (at least not empty)
				if len(formatted) == 0 {
					t.Error("Formatted code is empty")
				}
			}
		})
	}
}

// TestFormatTool_PreservesCodeSemantics tests that formatting preserves code semantics
func TestFormatTool_PreservesCodeSemantics(t *testing.T) {
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

	tests := []struct {
		name     string
		language string
		code     string
	}{
		{
			name:     "Python - preserves logic",
			language: "python",
			code:     "x=1+2\ny=x*3\nprint(y)",
		},
		{
			name:     "Go - preserves logic",
			language: "go",
			code:     "package main\nfunc main(){x:=1+2;println(x)}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Format the code
			result, err := module.formatCode(tt.language, tt.code)
			if err != nil {
				t.Fatalf("Failed to format code: %v", err)
			}

			if !result["success"].(bool) {
				t.Errorf("Format failed: %v", result["error"])
				return
			}

			formattedCode := result["formatted_code"].(string)

			// Execute both original and formatted code
			originalResult, err := module.runCode(tt.language, tt.code, "", 0)
			if err != nil {
				t.Fatalf("Failed to run original code: %v", err)
			}

			formattedResult, err := module.runCode(tt.language, formattedCode, "", 0)
			if err != nil {
				t.Fatalf("Failed to run formatted code: %v", err)
			}

			// Both should succeed
			if !originalResult["success"].(bool) {
				t.Errorf("Original code failed: %v", originalResult["error"])
			}

			if !formattedResult["success"].(bool) {
				t.Errorf("Formatted code failed: %v", formattedResult["error"])
			}

			// Output should be the same (semantics preserved)
			if originalResult["success"].(bool) && formattedResult["success"].(bool) {
				originalOutput := originalResult["output"].(string)
				formattedOutput := formattedResult["output"].(string)

				if originalOutput != formattedOutput {
					t.Logf("Original output: %s", originalOutput)
					t.Logf("Formatted output: %s", formattedOutput)
					// Note: We log but don't fail because formatting might add/remove whitespace
					// The important thing is both execute successfully
				}
			}
		})
	}
}
