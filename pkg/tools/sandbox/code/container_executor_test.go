// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
// Copyright (C) 2025  Agent Framework Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package code

import (
	"context"
	"testing"
	"time"
)

// TestNewContainerExecutor 测试容器执行器创建
func TestNewContainerExecutor(t *testing.T) {
	tests := []struct {
		name    string
		config  ContainerConfig
		wantErr bool
	}{
		{
			name: "disabled container executor",
			config: ContainerConfig{
				Enabled: false,
			},
			wantErr: false,
		},
		{
			name: "enabled with default config",
			config: ContainerConfig{
				Enabled: true,
			},
			wantErr: false, // May fail if Docker not available
		},
		{
			name: "enabled with custom config",
			config: ContainerConfig{
				Enabled: true,
				DefaultImages: map[string]string{
					"python": "python:3.11-alpine",
				},
				CPULimit:    "1.0",
				MemoryLimit: "256m",
				NetworkMode: "none",
				Timeout:     10 * time.Second,
				AutoCleanup: true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ce, err := NewContainerExecutor(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewContainerExecutor() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if ce == nil {
				t.Error("NewContainerExecutor() returned nil")
				return
			}
			if ce.IsEnabled() != tt.config.Enabled {
				t.Errorf("IsEnabled() = %v, want %v", ce.IsEnabled(), tt.config.Enabled)
			}
		})
	}
}

// TestContainerExecutor_IsEnabled 测试 IsEnabled 方法
func TestContainerExecutor_IsEnabled(t *testing.T) {
	tests := []struct {
		name    string
		enabled bool
		want    bool
	}{
		{
			name:    "enabled",
			enabled: true,
			want:    true,
		},
		{
			name:    "disabled",
			enabled: false,
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ce, _ := NewContainerExecutor(ContainerConfig{
				Enabled: tt.enabled,
			})
			if got := ce.IsEnabled(); got != tt.want {
				t.Errorf("IsEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestContainerExecutor_CheckConnection 测试 Docker 连接检查
func TestContainerExecutor_CheckConnection(t *testing.T) {
	ce, err := NewContainerExecutor(ContainerConfig{
		Enabled: true,
	})
	if err != nil {
		t.Skip("Docker not available, skipping test")
		return
	}

	ctx := context.Background()
	err = ce.CheckConnection(ctx)
	if err != nil {
		t.Logf("Docker connection check failed (expected if Docker not running): %v", err)
	}
}

// TestContainerExecutor_GetImage 测试获取镜像
func TestContainerExecutor_GetImage(t *testing.T) {
	ce, _ := NewContainerExecutor(ContainerConfig{
		Enabled: true,
		DefaultImages: map[string]string{
			"python": "python:3.11-alpine",
			"go":     "golang:1.21-alpine",
		},
	})

	tests := []struct {
		name     string
		language string
		want     string
	}{
		{
			name:     "python",
			language: "python",
			want:     "python:3.11-alpine",
		},
		{
			name:     "go",
			language: "go",
			want:     "golang:1.21-alpine",
		},
		{
			name:     "unknown language",
			language: "unknown",
			want:     "alpine:latest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ce.getImage(tt.language)
			if got != tt.want {
				t.Errorf("getImage() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestContainerExecutor_ParseCPULimit 测试 CPU 限制解析
func TestContainerExecutor_ParseCPULimit(t *testing.T) {
	ce, _ := NewContainerExecutor(ContainerConfig{Enabled: false})

	tests := []struct {
		name  string
		limit string
		want  int64
	}{
		{
			name:  "0.5 CPU",
			limit: "0.5",
			want:  500000000,
		},
		{
			name:  "1.0 CPU",
			limit: "1.0",
			want:  1000000000,
		},
		{
			name:  "2.0 CPU",
			limit: "2.0",
			want:  2000000000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ce.parseCPULimit(tt.limit)
			if got != tt.want {
				t.Errorf("parseCPULimit() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestContainerExecutor_ParseMemoryLimit 测试内存限制解析
func TestContainerExecutor_ParseMemoryLimit(t *testing.T) {
	ce, _ := NewContainerExecutor(ContainerConfig{Enabled: false})

	tests := []struct {
		name  string
		limit string
		want  int64
	}{
		{
			name:  "512 MB",
			limit: "512m",
			want:  536870912,
		},
		{
			name:  "1 GB",
			limit: "1g",
			want:  1073741824,
		},
		{
			name:  "256 KB",
			limit: "256k",
			want:  262144,
		},
		{
			name:  "bytes",
			limit: "1024",
			want:  1024,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ce.parseMemoryLimit(tt.limit)
			if got != tt.want {
				t.Errorf("parseMemoryLimit() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestContainerExecutor_GetCommand 测试获取执行命令
func TestContainerExecutor_GetCommand(t *testing.T) {
	ce, _ := NewContainerExecutor(ContainerConfig{Enabled: false})

	tests := []struct {
		name       string
		language   string
		scriptPath string
		want       []string
	}{
		{
			name:       "python",
			language:   "python",
			scriptPath: "/code/script.py",
			want:       []string{"python", "/code/script.py"},
		},
		{
			name:       "javascript",
			language:   "javascript",
			scriptPath: "/code/script.js",
			want:       []string{"node", "/code/script.js"},
		},
		{
			name:       "go",
			language:   "go",
			scriptPath: "/code/script.go",
			want:       []string{"go", "run", "/code/script.go"},
		},
		{
			name:       "bash",
			language:   "bash",
			scriptPath: "/code/script.sh",
			want:       []string{"sh", "/code/script.sh"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ce.getCommand(tt.language, tt.scriptPath)
			if len(got) != len(tt.want) {
				t.Errorf("getCommand() length = %v, want %v", len(got), len(tt.want))
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("getCommand()[%d] = %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

// TestContainerExecutor_GetFileExtension 测试获取文件扩展名
func TestContainerExecutor_GetFileExtension(t *testing.T) {
	ce, _ := NewContainerExecutor(ContainerConfig{Enabled: false})

	tests := []struct {
		name     string
		language string
		want     string
	}{
		{
			name:     "python",
			language: "python",
			want:     "py",
		},
		{
			name:     "javascript",
			language: "javascript",
			want:     "js",
		},
		{
			name:     "go",
			language: "go",
			want:     "go",
		},
		{
			name:     "bash",
			language: "bash",
			want:     "sh",
		},
		{
			name:     "unknown",
			language: "unknown",
			want:     "txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ce.getFileExtension(tt.language)
			if got != tt.want {
				t.Errorf("getFileExtension() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestContainerExecutor_CreateTempFile 测试创建临时文件
func TestContainerExecutor_CreateTempFile(t *testing.T) {
	ce, _ := NewContainerExecutor(ContainerConfig{Enabled: false})

	tests := []struct {
		name     string
		language string
		code     string
		wantErr  bool
	}{
		{
			name:     "python code",
			language: "python",
			code:     "print('Hello, World!')",
			wantErr:  false,
		},
		{
			name:     "javascript code",
			language: "javascript",
			code:     "console.log('Hello, World!');",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile, err := ce.createTempFile(tt.language, tt.code)
			if (err != nil) != tt.wantErr {
				t.Errorf("createTempFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tmpFile == "" && !tt.wantErr {
				t.Error("createTempFile() returned empty path")
			}
		})
	}
}

// TestContainerExecutor_GetStats 测试获取统计信息
func TestContainerExecutor_GetStats(t *testing.T) {
	ce, _ := NewContainerExecutor(ContainerConfig{Enabled: false})

	stats := ce.GetStats()
	if stats == nil {
		t.Error("GetStats() returned nil")
		return
	}

	if stats.TotalExecutions < 0 {
		t.Errorf("TotalExecutions = %v, want >= 0", stats.TotalExecutions)
	}
	if stats.SuccessCount < 0 {
		t.Errorf("SuccessCount = %v, want >= 0", stats.SuccessCount)
	}
	if stats.FailureCount < 0 {
		t.Errorf("FailureCount = %v, want >= 0", stats.FailureCount)
	}
}

// TestContainerExecutor_ListContainers 测试列出容器
func TestContainerExecutor_ListContainers(t *testing.T) {
	ce, _ := NewContainerExecutor(ContainerConfig{Enabled: false})

	containers := ce.ListContainers()
	if containers == nil {
		t.Error("ListContainers() returned nil")
		return
	}

	if len(containers) != 0 {
		t.Errorf("ListContainers() length = %v, want 0", len(containers))
	}
}

// Benchmark tests
func BenchmarkContainerExecutor_ParseCPULimit(b *testing.B) {
	ce, _ := NewContainerExecutor(ContainerConfig{Enabled: false})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ce.parseCPULimit("0.5")
	}
}

func BenchmarkContainerExecutor_ParseMemoryLimit(b *testing.B) {
	ce, _ := NewContainerExecutor(ContainerConfig{Enabled: false})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ce.parseMemoryLimit("512m")
	}
}
