// ReAct Agent 简单示例
// 演示如何使用 ReAct Agent 框架

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.uber.org/zap"

	"AgentFramework/pkg/framework/agent/react"
	"AgentFramework/pkg/errors"
)

func main() {
	// 初始化 logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	ctx := context.Background()

	// 创建 ReAct Agent 配置
	config := &react.ReActConfig{
		ModelName:            "gpt-4",
		MaxIterations:       10,
		MaxTokens:           2000,
		Temperature:         0.7,
		EnableReflection:    true,
		EnableToolExecution: true,
		ToolTimeout:         time.Second * 30,
		Logger:              logger,
	}

	// 使用全局工厂创建 ReAct Agent
	agent, err := react.GetGlobalReActAgentFactory().CreateAgent(ctx, config)
	if err != nil {
		logger.Error("Failed to create ReAct Agent",
			zap.Error(err),
		)
		os.Exit(1)
	}

	// 启动 ReAct 循环
	state, err := agent.Start(ctx, "如何使用 Python 读取一个 CSV 文件？")
	if err != nil {
		logger.Error("Failed to start ReAct loop",
			zap.Error(err),
		)
		os.Exit(1)
	}

	// 执行多个步骤
	for i := 0; i < 5; i++ {
		state, err = agent.Step(ctx, state)
		if err != nil {
			logger.Error("Failed to execute step",
				zap.Error(err),
			)
			break
		}

		if state.Status == react.ReActStatusCompleted || state.Status == react.ReActStatusFailed {
			break
		}

		// 显示当前思考
		if len(state.Thoughts) > 0 {
			lastThought := state.Thoughts[len(state.Thoughts)-1]
			fmt.Printf("Thought %d: %s\n", i+1, lastThought.Reasoning)
		}
	}

	// 获取最终结果
	if len(state.Observations) > 0 {
		lastObs := state.Observations[len(state.Observations)-1]
		fmt.Printf("\nFinal Result: %s\n", lastObs.ResultSummary())
	}

	fmt.Println("\nReAct Agent execution completed!")
}
