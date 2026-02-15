// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
// Copyright (C) 2025  Agent Framework Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package code

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

// ExecutionMode 执行模式
type ExecutionMode string

const (
	// ModeYaegi 使用 yaegi 解释器执行
	ModeYaegi ExecutionMode = "yaegi"
	// ModeGoRun 使用 go run 执行
	ModeGoRun ExecutionMode = "go_run"
	// ModeAuto 自动选择执行模式
	ModeAuto ExecutionMode = "auto"
)

// YaegiInterpreter yaegi Go 解释器
type YaegiInterpreter struct {
	interp      *interp.Interpreter
	mu          sync.RWMutex
	initialized bool
	config      YaegiConfig
	cache       *LRUCache // 编译缓存
}

// YaegiConfig yaegi 配置
type YaegiConfig struct {
	// PreloadStdlib 是否预加载标准库
	PreloadStdlib bool `json:"preload_stdlib" yaml:"preload_stdlib"`
	// PreloadPackages 预加载的包列表
	PreloadPackages []string `json:"preload_packages" yaml:"preload_packages"`
	// EnableCache 是否启用编译缓存
	EnableCache bool `json:"enable_cache" yaml:"enable_cache"`
	// CacheCapacity 缓存容量（默认 100）
	CacheCapacity int `json:"cache_capacity" yaml:"cache_capacity"`
}

// NewYaegiInterpreter 创建 yaegi 解释器实例
func NewYaegiInterpreter(config YaegiConfig) (*YaegiInterpreter, error) {
	yi := &YaegiInterpreter{
		config: config,
	}

	// 初始化缓存
	if config.EnableCache {
		capacity := config.CacheCapacity
		if capacity <= 0 {
			capacity = 100 // 默认容量
		}
		yi.cache = NewLRUCache(capacity)
	}

	if err := yi.initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize yaegi interpreter: %w", err)
	}

	return yi, nil
}

// initialize 初始化解释器
func (yi *YaegiInterpreter) initialize() error {
	yi.mu.Lock()
	defer yi.mu.Unlock()

	// 创建解释器实例
	yi.interp = interp.New(interp.Options{})

	// 预加载标准库
	if yi.config.PreloadStdlib {
		if err := yi.interp.Use(stdlib.Symbols); err != nil {
			return fmt.Errorf("failed to load stdlib: %w", err)
		}
	}

	// 预加载常用包
	if len(yi.config.PreloadPackages) > 0 {
		for _, pkg := range yi.config.PreloadPackages {
			// 尝试导入包
			if _, err := yi.interp.Eval(fmt.Sprintf("import \"%s\"", pkg)); err != nil {
				// 如果导入失败，记录但不中断初始化
				continue
			}
		}
	}

	yi.initialized = true
	return nil
}

// Run 执行 Go 代码
func (yi *YaegiInterpreter) Run(ctx context.Context, code string, input string) (*ExecutionResult, error) {
	yi.mu.RLock()
	if !yi.initialized {
		yi.mu.RUnlock()
		return nil, fmt.Errorf("interpreter not initialized")
	}
	yi.mu.RUnlock()

	// 检查缓存
	if yi.config.EnableCache && yi.cache != nil {
		cacheKey := yi.cache.HashCode(code)
		if cached, ok := yi.cache.Get(cacheKey); ok {
			// 返回缓存结果的副本
			result := *cached
			// 缓存命中，执行时间接近 0
			result.Duration = time.Microsecond
			return &result, nil
		}
	}

	startTime := time.Now()

	// 捕获输出
	var stdout, stderr bytes.Buffer

	// 创建一个新的解释器实例用于执行（避免状态污染）
	execInterp := interp.New(interp.Options{
		Stdout: &stdout,
		Stderr: &stderr,
	})

	// 加载标准库
	if err := execInterp.Use(stdlib.Symbols); err != nil {
		return &ExecutionResult{
			Success:  false,
			Error:    fmt.Sprintf("failed to load stdlib: %v", err),
			ExitCode: -1,
			Duration: time.Since(startTime),
		}, nil
	}

	// 准备代码
	preparedCode := yi.prepareCode(code)

	// 执行代码（带超时控制）
	done := make(chan error, 1)
	go func() {
		_, err := execInterp.Eval(preparedCode)
		done <- err
	}()

	// 等待执行完成或超时
	var result *ExecutionResult
	select {
	case err := <-done:
		duration := time.Since(startTime)

		if err != nil {
			result = &ExecutionResult{
				Success:  false,
				Output:   stdout.String(),
				Error:    err.Error(),
				ExitCode: 1,
				Duration: duration,
			}
		} else {
			result = &ExecutionResult{
				Success:  true,
				Output:   stdout.String(),
				Error:    stderr.String(),
				ExitCode: 0,
				Duration: duration,
			}
		}

	case <-ctx.Done():
		result = &ExecutionResult{
			Success:  false,
			Error:    "execution timeout",
			ExitCode: -1,
			Duration: time.Since(startTime),
		}
	}

	// 缓存成功的结果
	if yi.config.EnableCache && yi.cache != nil && result.Success {
		cacheKey := yi.cache.HashCode(code)
		yi.cache.Put(cacheKey, result)
	}

	return result, nil
}

// prepareCode 准备代码（添加必要的包装）
func (yi *YaegiInterpreter) prepareCode(code string) string {
	// 如果代码已经是完整的程序（有 package 和 main），直接返回
	if strings.Contains(code, "package main") && strings.Contains(code, "func main()") {
		return code
	}

	var result strings.Builder

	// 添加 package 声明
	result.WriteString("package main\n\n")

	// 检查是否需要导入 fmt
	needsFmt := strings.Contains(code, "fmt.") || strings.Contains(code, "Println") || strings.Contains(code, "Printf")

	// 添加 import 语句
	if needsFmt {
		result.WriteString("import \"fmt\"\n\n")
	}

	// 检查代码中是否有函数定义
	hasFuncDef := strings.Contains(code, "func ")

	if hasFuncDef {
		// 如果有函数定义，添加代码然后添加 main 函数
		result.WriteString(code)
		result.WriteString("\n\nfunc main() {}\n")
	} else {
		// 如果没有函数定义，将代码包装在 main 函数中
		result.WriteString("func main() {\n")

		// 缩进代码
		lines := strings.Split(code, "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				result.WriteString("\t" + line + "\n")
			}
		}

		result.WriteString("}\n")
	}

	return result.String()
}

// IsAvailable 检查 yaegi 是否可用
func (yi *YaegiInterpreter) IsAvailable() bool {
	yi.mu.RLock()
	defer yi.mu.RUnlock()
	return yi.initialized
}

// Reset 重置解释器状态
func (yi *YaegiInterpreter) Reset() error {
	return yi.initialize()
}

// GetCacheStats 获取缓存统计信息
func (yi *YaegiInterpreter) GetCacheStats() CacheStats {
	if yi.cache == nil {
		return CacheStats{}
	}
	return yi.cache.GetStats()
}

// ClearCache 清空缓存
func (yi *YaegiInterpreter) ClearCache() {
	if yi.cache != nil {
		yi.cache.Clear()
	}
}
