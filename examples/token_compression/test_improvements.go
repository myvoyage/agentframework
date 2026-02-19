// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later


// +build ignore

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"AgentFramework/agent"
	"AgentFramework/agent/token"
)

func main() {
	ctx := context.Background()

	fmt.Println("=== LLM 摘要质量和效率测试 ===")
	fmt.Println("==============================")

	// 测试配置
	cfg := &agent.HostConfig{
		Name:         "TestHost",
		Version:      "1.0.0",
		DefaultModel: "gpt-4o-mini",
		TokenCompression: &agent.TokenCompressionSpec{
			Enabled:             true,
			Strategy:            "summarize",
			TargetTokens:        300,
			MinTokens:           100,
			MaxTokens:           4000,
			PreserveSystemMessages: true,
			SummaryModelName:    "gpt-4o-mini",
			SummaryMaxTokens:    300,
			Temperature:         0.3, // 使用较低的温度
			CheckInterval:       30,
		},
	}

	// 创建 Host
	host, err := agent.NewHost(ctx, cfg, mockModelFactory, nil)
	if err != nil {
		log.Fatalf("创建 Host 失败: %v", err)
	}

	// 测试消息
	testMessages := []interface{}{
		map[string]interface{}{
			"role":    "system",
			"content": "你是一个专业的软件开发工程师助手。回答要详细且专业。",
		},
		map[string]interface{}{
			"role":    "user",
			"content": "我需要实现一个并发安全的任务调度系统。请提供一些建议。",
		},
		map[string]interface{}{
			"role":    "assistant",
			"content": "实现并发安全的任务调度系统需要考虑多个关键因素。以下是一些建议：\n1. 使用 sync.Mutex 或 sync.RWMutex 保护共享数据\n2. 使用 channel 进行 goroutine 通信和同步\n3. 实现优雅的关闭机制\n4. 使用 context 进行取消操作\n5. 实现错误处理和重试机制",
		},
		map[string]interface{}{
			"role":    "user",
			"content": "如何正确地在 Go 中使用 context 进行取消？",
		},
		map[string]interface{}{
			"role":    "assistant",
			"content": "使用 context 进行取消操作的最佳实践：\n```go\nctx, cancel := context.WithCancel(context.Background())\n\ndefer cancel()\n\n// 启动 goroutine\nfor _, task := range tasks {\n    go func(t string) {\n        select {\n        case <-ctx.Done():\n            log.Println(\"任务取消\")\n            return\n        case result := <-process(task):\n            log.Println(\"结果:\", result)\n        }\n    }(task)\n}\n\n// 在需要时取消\nif err != nil {\n    cancel()\n}\n```",
		},
		map[string]interface{}{
			"role":    "user",
			"content": "如何处理大量任务的调度和错误重试？",
		},
		map[string]interface{}{
			"role":    "assistant",
			"content": "使用工作池模式和有限状态机：\n```go\ntype Task struct {\n    ID       string\n    State    string\n    Retries  int\n    MaxRetries int\n    CreatedAt time.Time\n}\n\nfunc NewWorkerPool(size int, tasks chan *Task) {\n    for i := 0; i < size; i++ {\n        go func() {\n            for task := range tasks {\n                processTask(task)\n            }\n        }()\n    }\n}\n```",
		},
	}

	// 测试摘要质量和效率
	testSummary(ctx, host, testMessages)

	fmt.Println("\n=== 测试完成 ===")
}

func testSummary(ctx context.Context, host *agent.Host, messages []interface{}) {
	counter := token.NewDefaultTokenCounter()

	// 计算原始消息的 Token 数
	originalTokens := counter.CountMessages(messages)
	fmt.Printf("\n原始消息 Token 数: %d\n", originalTokens)

	// 测试不同摘要长度
	for _, target := range []int{200, 300, 400} {
		fmt.Printf("\n=== 测试目标 Token 数: %d ===\n", target)

		// 执行压缩
		startTime := time.Now()
		compressed, err := host.CompressMessages(ctx, messages, target)
		duration := time.Since(time.Now()) - startTime

		if err != nil {
			fmt.Printf("压缩失败: %v\n", err)
			continue
		}

		compressedTokens := counter.CountMessages(compressed)
		fmt.Printf("实际输出 Token 数: %d\n", compressedTokens)

		// 打印摘要质量
		fmt.Println("\n压缩后的消息:")
		for i, msg := range compressed {
			if msgMap, ok := msg.(map[string]interface{}); ok {
				content := msgMap["content"].(string)
				fmt.Printf("%d. [%v] %s\n", i+1, msgMap["role"], content)
			}
		}

		// 打印性能数据
		fmt.Printf("\n性能数据:")
		fmt.Printf("\n耗时: %.2fms", duration.Milliseconds())
		fmt.Printf("\n压缩比: %.1f%%", (1.0-float64(compressedTokens)/float64(originalTokens))*100)
	}
}

func mockModelFactory(ctx context.Context, name string) (agent.ChatModel, error) {
	return &mockChatModel{}, nil
}

type mockChatModel struct{}

func (m *mockChatModel) Chat(ctx context.Context, messages []interface{}, maxTokens int, temperature float64) (string, error) {
	// 模拟 LLM 响应
	time.Sleep(200 * time.Millisecond)

	return `这是一个关于 Go 并发编程的对话。用户问了三个问题：

1. 如何实现并发安全的任务调度系统
2. 如何正确使用 context 进行取消操作
3. 如何处理大量任务的调度和错误重试

主要要点：
- 使用 sync.Mutex 或 sync.RWMutex 保护共享数据
- 使用 channel 进行 goroutine 通信和同步
- 实现优雅的关闭机制和 context 取消
- 使用工作池模式和有限状态机处理大量任务
- 实现错误处理和重试机制

代码示例包括了 context 取消和 worker pool 的实现。`, nil
}

func (m *mockChatModel) Embed(ctx context.Context, text string) ([]float32, error) {
	return []float32{0.1, 0.2, 0.3}, nil
}

func (m *mockChatModel) ToolCall(ctx context.Context, messages []interface{}, tools []map[string]interface{}, maxTokens int, temperature float64) (interface{}, error) {
	return nil, fmt.Errorf("tool call not implemented")
}
