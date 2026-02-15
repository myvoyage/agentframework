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

// TestYaegiInterpreter_Basic tests basic yaegi functionality
func TestYaegiInterpreter_Basic(t *testing.T) {
	config := YaegiConfig{
		PreloadStdlib:   true,
		PreloadPackages: []string{"fmt"},
		EnableCache:     true,
	}

	yi, err := NewYaegiInterpreter(config)
	if err != nil {
		t.Fatalf("Failed to create yaegi interpreter: %v", err)
	}

	if !yi.IsAvailable() {
		t.Error("Yaegi interpreter should be available")
	}
}

// TestYaegiInterpreter_SimpleExecution tests simple code execution
func TestYaegiInterpreter_SimpleExecution(t *testing.T) {
	config := YaegiConfig{
		PreloadStdlib: true,
	}

	yi, err := NewYaegiInterpreter(config)
	if err != nil {
		t.Fatalf("Failed to create yaegi interpreter: %v", err)
	}

	tests := []struct {
		name    string
		code    string
		wantOut string
		wantErr bool
	}{
		{
			name:    "Simple print",
			code:    `fmt.Println("Hello, Yaegi!")`,
			wantOut: "Hello, Yaegi!",
			wantErr: false,
		},
		{
			name: "Variable and print",
			code: `
				x := 42
				fmt.Println(x)
			`,
			wantOut: "42",
			wantErr: false,
		},
		{
			name: "Function definition and call",
			code: `
				package main
				import "fmt"
				
				func add(a, b int) int {
					return a + b
				}
				
				func main() {
					result := add(3, 4)
					fmt.Println(result)
				}
			`,
			wantOut: "7",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			result, err := yi.Run(ctx, tt.code, "")

			if tt.wantErr {
				if err == nil && result.Success {
					t.Error("Expected error but got success")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if !result.Success {
					t.Errorf("Execution failed: %s", result.Error)
				}
				if !strings.Contains(result.Output, tt.wantOut) {
					t.Errorf("Expected output to contain %q, got %q", tt.wantOut, result.Output)
				}
			}
		})
	}
}

// TestYaegiInterpreter_Timeout tests timeout handling
func TestYaegiInterpreter_Timeout(t *testing.T) {
	config := YaegiConfig{
		PreloadStdlib: true,
	}

	yi, err := NewYaegiInterpreter(config)
	if err != nil {
		t.Fatalf("Failed to create yaegi interpreter: %v", err)
	}

	code := `
		package main
		import (
			"fmt"
			"time"
		)
		func main() {
			time.Sleep(10 * time.Second)
			fmt.Println("Done")
		}
	`

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	result, err := yi.Run(ctx, code, "")
	if err != nil {
		t.Logf("Got error (expected): %v", err)
	}

	if result.Success {
		t.Error("Expected timeout but execution succeeded")
	}

	if !strings.Contains(result.Error, "timeout") {
		t.Logf("Got error: %s (expected timeout)", result.Error)
	}
}

// TestGoRunner_ExecutionModes tests different execution modes
func TestGoRunner_ExecutionModes(t *testing.T) {
	config := CodeExecutorConfig{
		Timeout:     60000,
		MemoryLimit: 512,
		CPULimit:    2,
	}

	tempDir := t.TempDir()
	runner := NewGoRunner(config, tempDir)

	code := `
		package main
		import "fmt"
		func main() {
			fmt.Println("Hello from Go!")
		}
	`

	tests := []struct {
		name string
		mode ExecutionMode
	}{
		{"Yaegi mode", ModeYaegi},
		{"Go run mode", ModeGoRun},
		{"Auto mode", ModeAuto},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner.SetExecutionMode(tt.mode)

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			result, err := runner.Run(ctx, code, "")
			if err != nil {
				t.Fatalf("Execution failed: %v", err)
			}

			if !result.Success {
				t.Errorf("Expected success, got error: %s", result.Error)
			}

			if !strings.Contains(result.Output, "Hello from Go!") {
				t.Errorf("Expected output to contain 'Hello from Go!', got: %s", result.Output)
			}

			t.Logf("Mode: %s, Duration: %v", tt.mode, result.Duration)
		})
	}
}

// TestGoRunner_FallbackMechanism tests fallback from yaegi to go run
func TestGoRunner_FallbackMechanism(t *testing.T) {
	config := CodeExecutorConfig{
		Timeout:     60000,
		MemoryLimit: 512,
		CPULimit:    2,
	}

	tempDir := t.TempDir()
	runner := NewGoRunner(config, tempDir)
	runner.SetExecutionMode(ModeAuto)

	// Code that might fail in yaegi but work in go run
	code := `
		package main
		import "fmt"
		func main() {
			fmt.Println("Testing fallback")
		}
	`

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := runner.Run(ctx, code, "")
	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected success, got error: %s", result.Error)
	}

	// Check stats
	stats := runner.GetStats()
	t.Logf("Stats: %+v", stats)

	if stats["yaegi_executions"] == 0 && stats["go_run_executions"] == 0 {
		t.Error("Expected at least one execution")
	}
}

// TestGoRunner_Stats tests statistics tracking
func TestGoRunner_Stats(t *testing.T) {
	config := CodeExecutorConfig{
		Timeout:     60000,
		MemoryLimit: 512,
		CPULimit:    2,
	}

	tempDir := t.TempDir()
	runner := NewGoRunner(config, tempDir)

	code := `fmt.Println("Test")`

	// Execute with yaegi
	runner.SetExecutionMode(ModeYaegi)
	ctx1, cancel1 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel1()
	runner.Run(ctx1, code, "")

	// Execute with go run
	runner.SetExecutionMode(ModeGoRun)
	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()
	runner.Run(ctx2, code, "")

	stats := runner.GetStats()

	if stats["yaegi_executions"] == 0 {
		t.Error("Expected yaegi executions > 0")
	}

	if stats["go_run_executions"] == 0 {
		t.Error("Expected go_run executions > 0")
	}

	t.Logf("Final stats: %+v", stats)
}

// TestYaegiInterpreter_ErrorHandling tests error handling
func TestYaegiInterpreter_ErrorHandling(t *testing.T) {
	config := YaegiConfig{
		PreloadStdlib: true,
	}

	yi, err := NewYaegiInterpreter(config)
	if err != nil {
		t.Fatalf("Failed to create yaegi interpreter: %v", err)
	}

	tests := []struct {
		name    string
		code    string
		wantErr bool
	}{
		{
			name:    "Syntax error",
			code:    `fmt.Println("missing closing quote)`,
			wantErr: true,
		},
		{
			name:    "Undefined variable",
			code:    `fmt.Println(undefinedVar)`,
			wantErr: true,
		},
		{
			name:    "Type error",
			code:    `x := "string"; y := x + 5`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			result, err := yi.Run(ctx, tt.code, "")

			if tt.wantErr {
				if err == nil && result.Success {
					t.Error("Expected error but got success")
				}
			} else {
				if err != nil || !result.Success {
					t.Errorf("Expected success but got error: %v, result: %+v", err, result)
				}
			}
		})
	}
}

// BenchmarkYaegiVsGoRun benchmarks yaegi vs go run performance
func BenchmarkYaegiVsGoRun(b *testing.B) {
	config := CodeExecutorConfig{
		Timeout:     60000,
		MemoryLimit: 512,
		CPULimit:    2,
	}

	tempDir := b.TempDir()
	runner := NewGoRunner(config, tempDir)

	code := `
		package main
		import "fmt"
		func main() {
			sum := 0
			for i := 0; i < 100; i++ {
				sum += i
			}
			fmt.Println(sum)
		}
	`

	b.Run("Yaegi", func(b *testing.B) {
		runner.SetExecutionMode(ModeYaegi)
		ctx := context.Background()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			runner.Run(ctx, code, "")
		}
	})

	b.Run("GoRun", func(b *testing.B) {
		runner.SetExecutionMode(ModeGoRun)
		ctx := context.Background()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			runner.Run(ctx, code, "")
		}
	})
}

// TestYaegiInterpreter_ComplexCode tests complex code execution
func TestYaegiInterpreter_ComplexCode(t *testing.T) {
	config := YaegiConfig{
		PreloadStdlib:   true,
		PreloadPackages: []string{"fmt", "strings", "time"},
	}

	yi, err := NewYaegiInterpreter(config)
	if err != nil {
		t.Fatalf("Failed to create yaegi interpreter: %v", err)
	}

	code := `
		package main
		import (
			"fmt"
			"strings"
		)

		func processString(s string) string {
			return strings.ToUpper(s)
		}

		func main() {
			result := processString("hello world")
			fmt.Println(result)
		}
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := yi.Run(ctx, code, "")
	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected success, got error: %s", result.Error)
	}

	if !strings.Contains(result.Output, "HELLO WORLD") {
		t.Errorf("Expected output to contain 'HELLO WORLD', got: %s", result.Output)
	}
}
