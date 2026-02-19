// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later


// Agent Framework - Skill System Validation
// 验证技能系统的核心功能

package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"AgentFramework/agent/skills"
)

func main() {
	fmt.Println("=== AgentFramework 技能系统验证 ===\n")

	ctx := context.Background()
	passed := 0
	failed := 0

	// 测试1: 注册表创建
	fmt.Println("测试1: 创建技能注册表")
	registry := skills.NewSkillRegistry(&skills.RegistryConfig{
		BaseDir:  ".test_registry",
		AutoSave: false,
	})
	if registry != nil {
		fmt.Println("✅ 通过")
		passed++
	} else {
		fmt.Println("❌ 失败")
		failed++
	}

	// 测试2: 技能注册
	fmt.Println("\n测试2: 注册技能")
	entry := &skills.SkillEntry{
		ID:          "test_skill",
		Name:        "测试技能",
		Description: "用于验证的测试技能",
		Category:    "test",
		Tags:        []string{"test"},
		Version:     "1.0.0",
		InputSchema: &skills.Schema{
			Type: "object",
		},
	}
	err := registry.Register(entry)
	if err == nil {
		fmt.Println("✅ 通过")
		passed++
	} else {
		fmt.Printf("❌ 失败: %v\n", err)
		failed++
	}

	// 测试3: 技能查询
	fmt.Println("\n测试3: 查询技能")
	if exists, ok := registry.GetByID("test_skill"); ok {
		fmt.Printf("✅ 通过 - 找到技能: %s\n", exists.Name)
		passed++
	} else {
		fmt.Println("❌ 失败 - 未找到技能")
		failed++
	}

	// 测试4: 定义管理器
	fmt.Println("\n测试4: 创建定义管理器")
	defManager := skills.NewDefinitionManager("agent/skills/definitions")
	if defManager != nil {
		fmt.Println("✅ 通过")
		passed++
	} else {
		fmt.Println("❌ 失败")
		failed++
	}

	// 测试5: 加载技能定义
	fmt.Println("\n测试5: 加载技能定义")
	definition, err := defManager.Load("http_request")
	if err == nil && definition != nil {
		fmt.Printf("✅ 通过 - 加载了定义: %s\n", definition.Name)
		passed++
	} else {
		fmt.Printf("❌ 失败: %v\n", err)
		failed++
	}

	// 测试6: 示例库
	fmt.Println("\n测试6: 创建示例库")
	library := skills.NewExampleLibrary(".skills/examples")
	err = library.CreateBuiltInTemplates()
	if err == nil && library.Count() > 0 {
		fmt.Printf("✅ 通过 - 创建了 %d 个模板\n", library.Count())
		passed++
	} else {
		fmt.Printf("❌ 失败: %v\n", err)
		failed++
	}

	// 测试7: 模板渲染
	fmt.Println("\n测试7: 渲染模板")
	data := map[string]interface{}{
		"FunctionName": "TestFunc",
		"Description":  "测试函数",
		"Params":       "ctx context.Context",
		"URLPattern":   `"https://example.com"`,
		"URLArgs":      "",
		"Headers":      map[string]string{},
	}
	code, err := library.Render(ctx, "http_get_request", data)
	if err == nil && len(code) > 0 {
		fmt.Printf("✅ 通过 - 生成了 %d 字节代码\n", len(code))
		passed++
	} else {
		fmt.Printf("❌ 失败: %v\n", err)
		failed++
	}

	// 测试8: 渐进式加载器
	fmt.Println("\n测试8: 创建渐进式加载器")
	loader := skills.NewProgressiveLoader("agent/skills/definitions")
	if loader != nil {
		fmt.Println("✅ 通过")
		passed++
	} else {
		fmt.Println("❌ 失败")
		failed++
	}

	// 测试9: 列出技能元数据
	fmt.Println("\n测试9: 列出技能元数据")
	metas, err := loader.ListSkills()
	if err == nil && len(metas) > 0 {
		fmt.Printf("✅ 通过 - 找到 %d 个技能\n", len(metas))
		passed++
	} else {
		fmt.Printf("❌ 失败: %v\n", err)
		failed++
	}

	// 测试10: 增强执行器
	fmt.Println("\n测试10: 创建增强执行器")
	config := &skills.ExecutorConfig{
		EnableRetry:    true,
		MaxRetries:     3,
		RetryDelay:     1 * time.Second,
		EnableTimeout:  true,
		DefaultTimeout: 30 * time.Second,
		EnableSkip:     true,
		EnableCache:    true,
		EnableLog:      true,
		LogLevel:       "info",
	}
	executor := skills.NewEnhancedSkillExecutor(config)
	if executor != nil {
		fmt.Println("✅ 通过")
		passed++
	} else {
		fmt.Println("❌ 失败")
		failed++
	}

	// 测试11: 多级缓存
	fmt.Println("\n测试11: 创建多级缓存")
	cache := skills.NewMultiLevelCache(skills.DefaultCacheConfig())
	defer cache.Close()
	if cache != nil {
		fmt.Println("✅ 通过")
		passed++
	} else {
		fmt.Println("❌ 失败")
		failed++
	}

	// 测试12: 缓存操作
	fmt.Println("\n测试12: 缓存读写操作")
	cache.Set("test_key", "test_value", time.Minute)
	if data, found := cache.Get("test_key"); found && data == "test_value" {
		fmt.Println("✅ 通过")
		passed++
	} else {
		fmt.Println("❌ 失败")
		failed++
	}

	// 测试13: 缓存统计
	fmt.Println("\n测试13: 获取缓存统计")
	stats := cache.GetStats()
	if stats != nil && len(stats) > 0 {
		fmt.Printf("✅ 通过 - 统计: %+v\n", stats)
		passed++
	} else {
		fmt.Println("❌ 失败")
		failed++
	}

	// 清理测试文件
	os.RemoveAll(".test_registry")

	// 输出总结
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Printf("验证结果: %d 通过, %d 失败\n", passed, failed)
	if failed == 0 {
		fmt.Println("🎉 所有测试通过！技能系统运行正常。")
	} else {
		fmt.Printf("⚠️  有 %d 个测试失败，请检查。\n", failed)
	}
	fmt.Println(strings.Repeat("=", 50))
}
