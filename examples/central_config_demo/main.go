// Agent Framework - Central Configuration Demo
// Copyright (C) 2025  Agent Framework Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"AgentFramework/agent"
	"AgentFramework/agent/config"
)

func main() {
	ctx := context.Background()

	// 使用新的配置系统
	// 初始化配置
	config.InitCentralConfig("config/default_config.yaml")
	if err != nil {
		log.Fatalf("Failed to initialize config: %v", err)
	}

	fmt.Println("=== 中央配置管理演示 ===")
	fmt.Println("\n1. 从配置读取内存限制：")
	memLimit := config.GetInt("memory.historySize", 1000)
	fmt.Printf("  - 历史大小: %d 条目\n", memLimit)

	fmt.Println("\n2. 从配置读取工作线程数：")
	workerCount := config.GetInt("worker.count", 5)
	fmt.Printf("  - 工作线程数: %d\n", workerCount)

	fmt.Println("\n3. 从配置读取缓存配置：")
	cacheMax := config.GetInt("cache.maxSize", 200)
	cacheTTL := config.GetDuration("cache.ttl", 1*time.Hour)
	fmt.Printf("  - 缓存最大值: %d\n", cacheMax)
	fmt.Printf("  - TTL: %v\n", cacheTTL)

	fmt.Println("\n4. 动态加载模型：")
	modelName := config.GetString("model.defaultModel", "llama3")
	fmt.Printf("  - 默认模型: %s\n", modelName)

	fmt.Println("\n5. 创建 Agent 时使用配置化参数")
	fmt.Println("\n--- 创建 Chat Agent ---")
	chatModel, err := model.NewChatModel(ctx, &model.ChatModelConfig{
		Model: modelName,
	})
	if err != nil {
		log.Fatalf("Failed to create model: %v", err)
	}

	chatAgent, err := agent.NewChatAgent(
		agent.WithName("配置演示助手"),
		agent.WithInstructions("你是一个友好的AI助手，使用中央配置系统管理参数"),
		agent.WithModel(chatModel),
	)

	if err != nil {
		log.Fatalf("Failed to create chat agent: %v", err)
	}

	// 运行一次对话
	response, err := chatAgent.Run(ctx, "请告诉我当前的内存限制是多少")
	if err != nil {
		log.Fatalf("Failed to run chat agent: %v", err)
	}

	fmt.Printf("回答: %s\n", response.Content)
}
