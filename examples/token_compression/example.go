package main

import (
	"context"
	"fmt"

	"AgentFramework/agent"
	"AgentFramework/agent/token"
)

func main() {
	fmt.Println("AgentFramework Token 压缩功能示例")
	fmt.Println("===============================")

	ctx := context.Background()

	// 1. 创建 Token 压缩器
	config := token.DefaultCompressConfig()
	config.Strategy = token.StrategyHybrid
	config.TargetTokens = 500
	config.PreserveSystemMessages = true

	compressor := token.NewMessageCompressor(config, llmSummarizer)

	// 2. 模拟大量消息
	messages := generateConversation()

	counter := token.NewDefaultTokenCounter()
	originalTokens := counter.CountMessages(messages)
	fmt.Printf("\n原始消息 Token 数: %d\n", originalTokens)

	// 3. 执行压缩
	fmt.Println("\n执行 Token 压缩...")
	start := time.Now()
	compressed, err := compressor.CompressMessages(ctx, messages, config.TargetTokens)
	if err != nil {
		fmt.Printf("压缩失败: %v\n", err)
		return
	}
	compressedTokens := counter.CountMessages(compressed)

	fmt.Printf("\n压缩完成!")
	fmt.Printf("\n目标 Token 数: %d", config.TargetTokens)
	fmt.Printf("\n实际输出 Token: %d", compressedTokens)
	fmt.Printf("\n压缩比: %.1f%%", float64(compressedTokens)/float64(originalTokens)*100)
	fmt.Printf("\n耗时: %.2fms", time.Since(start).Milliseconds())

	// 4. 打印压缩结果
	fmt.Println("\n")
	printMessages(compressed)

	// 5. 显示统计信息
	stats := compressor.GetStats()
	fmt.Printf("\n统计信息:")
	fmt.Printf("\n总压缩次数: %d", stats.TotalCompressions)
	fmt.Printf("\n总输入 Token: %d", stats.TotalInputTokens)
	fmt.Printf("\n总输出 Token: %d", stats.TotalOutputTokens)
	fmt.Printf("\n平均压缩比: %.1f%%", stats.AverageRatio*100)

	if stats.LastCompression != nil {
		fmt.Printf("\n上次压缩:")
		fmt.Printf("\n策略: %s", stats.LastCompression.Strategy)
		fmt.Printf("\n输入: %d, 输出: %d", stats.LastCompression.InputTokens, stats.LastCompression.OutputTokens)
		fmt.Printf("\n压缩比: %.1f%%", stats.LastCompression.CompressionRatio*100)
		fmt.Printf("\n耗时: %dms", stats.LastCompression.DurationMs)
	}
}

func llmSummarizer(ctx context.Context, prompt string, maxTokens int) (string, error) {
	// 在实际应用中，这里会调用真实的 LLM API（如 OpenAI GPT）
	// 简单的模拟实现
	fmt.Println("\n  → 调用 LLM 进行摘要")
	return "这是一个模拟的 LLM 摘要。原始对话包含多个关于 Go 并发编程和错误处理的问题。", nil
}

func generateConversation() []interface{} {
	return []interface{}{
		map[string]interface{}{"role": "system", "content": "你是一个专业的 Go 语言开发工程师助手。回答要详细且专业。"},
		map[string]interface{}{"role": "user", "content": "我需要实现一个并发安全的任务调度系统。请提供一些建议。"},
		map[string]interface{}{"role": "assistant", "content": "实现并发安全的任务调度系统需要考虑多个关键因素。以下是一些建议：\n1. 使用 sync.Mutex 或 sync.RWMutex 保护共享数据\n2. 使用 channel 进行 goroutine 通信和同步\n3. 实现优雅的关闭机制\n4. 使用 context 进行取消操作\n5. 实现错误处理和重试机制"},
		map[string]interface{}{"role": "user", "content": "如何正确地在 Go 中使用 context 进行取消？"},
		map[string]interface{}{"role": "assistant", "content": "使用 context 进行取消操作的最佳实践：\n```go\nctx, cancel := context.WithCancel(context.Background())\n\ndefer cancel()\n\n// 启动 goroutine\nfor _, task := range tasks {\n    go func(t string) {\n        select {\n        case <-ctx.Done():\n            log.Println(\"任务取消\")\n            return\n        case result := <-process(task):\n            log.Println(\"结果:\", result)\n        }\n    }(task)\n}\n\n// 在需要时取消\nif err != nil {\n    cancel()\n}\n```"},
		map[string]interface{}{"role": "user", "content": "如何处理大量任务的调度和错误重试？"},
		map[string]interface{}{"role": "assistant", "content": "使用工作池模式和有限状态机：\n```go\ntype Task struct {\n    ID       string\n    State    string\n    Retries  int\n    MaxRetries int\n    CreatedAt time.Time\n}\n\nfunc NewWorkerPool(size int, tasks chan *Task) {\n    for i := 0; i < size; i++ {\n        go func() {\n            for task := range tasks {\n                processTask(task)\n            }\n        }()\n    }\n}\n```"},
	}
}

func printMessages(messages []interface{}) {
	fmt.Println("压缩后的消息:")
	for i, msg := range messages {
		if m, ok := msg.(map[string]interface{}); ok {
			fmt.Printf("\n[%d] %-8s %s", i+1, m["role"], truncate(m["content"], 60))
		}
	}
}

func truncate(s interface{}, n int) string {
	str := fmt.Sprintf("%v", s)
	if len(str) <= n {
		return str
	}
	return str[:n-3] + "..."
}
