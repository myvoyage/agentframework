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

package code

import (
	"testing"
	"time"
)

// TestCodeScanner_Scan tests code scanning for dangerous operations
func TestCodeScanner_Scan(t *testing.T) {
	scanner := NewCodeScanner()

	tests := []struct {
		name     string
		language string
		code     string
		wantWarn bool
	}{
		{
			name:     "Python safe code",
			language: "python",
			code:     "print('Hello, World!')",
			wantWarn: false,
		},
		{
			name:     "Python dangerous os.system",
			language: "python",
			code:     "import os\nos.system('rm -rf /')",
			wantWarn: true,
		},
		{
			name:     "Python dangerous eval",
			language: "python",
			code:     "eval('print(1+1)')",
			wantWarn: true,
		},
		{
			name:     "JavaScript safe code",
			language: "javascript",
			code:     "console.log('Hello, World!');",
			wantWarn: false,
		},
		{
			name:     "JavaScript dangerous eval",
			language: "javascript",
			code:     "eval('console.log(1+1)');",
			wantWarn: true,
		},
		{
			name:     "Bash safe code",
			language: "bash",
			code:     "echo 'Hello, World!'",
			wantWarn: false,
		},
		{
			name:     "Bash dangerous rm",
			language: "bash",
			code:     "rm -rf /",
			wantWarn: true,
		},
		{
			name:     "Go safe code",
			language: "go",
			code:     "fmt.Println(\"Hello, World!\")",
			wantWarn: false,
		},
		{
			name:     "Go dangerous syscall",
			language: "go",
			code:     "import \"syscall\"\nsyscall.Exit(1)",
			wantWarn: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			warnings, err := scanner.Scan(tt.language, tt.code)
			if err != nil {
				t.Fatalf("Scan() error = %v", err)
			}

			hasWarnings := len(warnings) > 0
			if hasWarnings != tt.wantWarn {
				t.Errorf("Scan() warnings = %v, wantWarn = %v", warnings, tt.wantWarn)
			}

			if hasWarnings {
				t.Logf("Detected warnings: %v", warnings)
			}
		})
	}
}

// TestCodeScanner_MultiplePatterns tests detection of multiple dangerous patterns
func TestCodeScanner_MultiplePatterns(t *testing.T) {
	scanner := NewCodeScanner()

	code := `
import os
import subprocess

os.system('ls')
subprocess.run(['echo', 'hello'])
eval('1+1')
`

	warnings, err := scanner.Scan("python", code)
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	if len(warnings) < 3 {
		t.Errorf("Expected at least 3 warnings, got %d: %v", len(warnings), warnings)
	}

	t.Logf("Detected %d warnings: %v", len(warnings), warnings)
}

// TestResourceLimiter_ValidateResourceLimits tests resource limit validation
func TestResourceLimiter_ValidateResourceLimits(t *testing.T) {
	tests := []struct {
		name          string
		memoryLimitMB int
		cpuLimit      int
		wantErr       bool
	}{
		{
			name:          "Valid limits",
			memoryLimitMB: 512,
			cpuLimit:      2,
			wantErr:       false,
		},
		{
			name:          "Zero limits (unlimited)",
			memoryLimitMB: 0,
			cpuLimit:      0,
			wantErr:       false,
		},
		{
			name:          "Negative memory limit",
			memoryLimitMB: -1,
			cpuLimit:      2,
			wantErr:       true,
		},
		{
			name:          "Negative CPU limit",
			memoryLimitMB: 512,
			cpuLimit:      -1,
			wantErr:       true,
		},
		{
			name:          "Memory limit too low",
			memoryLimitMB: 32,
			cpuLimit:      2,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateResourceLimits(tt.memoryLimitMB, tt.cpuLimit)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateResourceLimits() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestResourceLimiter_Creation tests resource limiter creation
func TestResourceLimiter_Creation(t *testing.T) {
	limiter := NewResourceLimiter(512, 2, 60*time.Second)
	if limiter == nil {
		t.Fatal("NewResourceLimiter() returned nil")
	}

	if limiter.GetMemoryLimit() != 512 {
		t.Errorf("Expected memory limit 512, got %d", limiter.GetMemoryLimit())
	}

	if limiter.GetCPULimit() != 2 {
		t.Errorf("Expected CPU limit 2, got %d", limiter.GetCPULimit())
	}

	if limiter.GetExecutionTimeLimit() != 60*time.Second {
		t.Errorf("Expected execution time limit 60s, got %v", limiter.GetExecutionTimeLimit())
	}
}

// TestCodeScanner_UnsupportedLanguage tests scanning unsupported language
func TestCodeScanner_UnsupportedLanguage(t *testing.T) {
	scanner := NewCodeScanner()

	warnings, err := scanner.Scan("ruby", "puts 'Hello'")
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	if len(warnings) != 0 {
		t.Errorf("Expected no warnings for unsupported language, got %v", warnings)
	}
}

// TestCodeScanner_EmptyCode tests scanning empty code
func TestCodeScanner_EmptyCode(t *testing.T) {
	scanner := NewCodeScanner()

	warnings, err := scanner.Scan("python", "")
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	if len(warnings) != 0 {
		t.Errorf("Expected no warnings for empty code, got %v", warnings)
	}
}

// TestCodeScanner_CaseSensitivity tests case sensitivity in pattern matching
func TestCodeScanner_CaseSensitivity(t *testing.T) {
	scanner := NewCodeScanner()

	// Test with uppercase (should not match)
	warnings1, _ := scanner.Scan("python", "OS.SYSTEM('ls')")
	if len(warnings1) > 0 {
		t.Logf("Note: Scanner detected uppercase pattern (case-insensitive): %v", warnings1)
	}

	// Test with lowercase (should match)
	warnings2, _ := scanner.Scan("python", "os.system('ls')")
	if len(warnings2) == 0 {
		t.Error("Expected warning for lowercase pattern")
	}
}
