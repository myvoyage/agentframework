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

// Additional permission under GNU Affero General Public License version 3 section 7
// If you modify this Program, or any covered work, by linking or combining it
// with other code, such other code is not for that reason alone subject to any
// of the requirements of the GNU Affero GPL version 3 as long as you maintain
// the separation between the Program and the other code.

// For network interaction purposes, when this Program is used over a network,
// the source code of the Program must be made available to users of the network.
// You can comply with this requirement by providing a link to the source code
// repository in your user interface or documentation.

// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"AgentFramework/agent"

	"github.com/cloudwego/eino-ext/components/model/ollama"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/redis/go-redis/v9"
)

type SumParams struct {
	A float64 `json:"a" jsonschema_description:"first operand"`
	B float64 `json:"b" jsonschema_description:"second operand"`
}

func sumTool(_ context.Context, p *SumParams) (string, error) {
	return fmt.Sprintf("%f", p.A+p.B), nil
}

const demoHostConfigYAML = `
name: demo_app
version: 1.0.0
defaultModel: default
threadStore:
  type: memory
agents:
  - name: calc_chat
    kind: chat
    model: default
    instructions: "You are a calculator. Use tools to compute numeric results."
    tools: ["sum"]
    middlewares: ["logging", "telemetry", "safe_output"]
  - name: http_react
    kind: react
    model: default
    instructions: "You can call http_request to fetch data from HTTP APIs when necessary."
    tools: ["http_request"]
    middlewares: ["logging", "telemetry", "safe_output"]
  - name: aggregator
    kind: chat
    model: default
    instructions: "Summarize all agents' opinions into a concise Chinese conclusion."
    middlewares: ["logging", "telemetry", "safe_output"]
workflows:
  - name: host_sequential_calc
    kind: sequential
    steps: ["calc_chat", "aggregator"]
  - name: host_http_review
    kind: aggregating_parallel
    agents: ["calc_chat", "http_react"]
    aggregator: "aggregator"
  - name: host_router
    kind: routing
    model: default
    routes:
      calc: host_sequential_calc
      http_review: host_http_review
  - name: host_planner
    kind: planning
    model: default
    routes:
      calc: host_sequential_calc
      http_review: host_http_review
`

type HTTPRequestParams struct {
	URL    string `json:"url" jsonschema_description:"HTTP URL to request"`
	Method string `json:"method" jsonschema_description:"HTTP method, GET or POST"`
	Body   string `json:"body,omitempty" jsonschema_description:"Optional request body for POST"`
}

func httpRequestTool(ctx context.Context, p *HTTPRequestParams) (string, error) {
	if p.URL == "" {
		return "", fmt.Errorf("url is required")
	}

	method := strings.ToUpper(p.Method)
	if method == "" {
		method = http.MethodGet
	}

	if method != http.MethodGet && method != http.MethodPost {
		return "", fmt.Errorf("unsupported method %s", method)
	}

	if !strings.HasPrefix(p.URL, "http://") && !strings.HasPrefix(p.URL, "https://") {
		return "", fmt.Errorf("only http/https is allowed")
	}

	var body io.Reader
	if method == http.MethodPost && p.Body != "" {
		body = strings.NewReader(p.Body)
	}

	req, err := http.NewRequestWithContext(ctx, method, p.URL, body)
	if err != nil {
		return "", err
	}

	if method == http.MethodPost && p.Body != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if len(data) > 4096 {
		data = data[:4096]
	}

	return fmt.Sprintf("status: %d\nbody:\n%s", resp.StatusCode, string(data)), nil
}

func newOpenAIModel(ctx context.Context) (agent.ChatModel, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	modelName := os.Getenv("OPENAI_MODEL")
	baseURL := os.Getenv("OPENAI_BASE_URL")

	m, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		APIKey:  apiKey,
		Model:   modelName,
		BaseURL: baseURL,
	})
	if err != nil {
		return nil, err
	}
	return m, nil
}

func newOllamaModel(ctx context.Context) (agent.ChatModel, error) {
	baseURL := os.Getenv("OLLAMA_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	modelName := os.Getenv("OLLAMA_MODEL")
	if modelName == "" {
		modelName = "llama3.1"
	}

	m, err := ollama.NewChatModel(ctx, &ollama.ChatModelConfig{
		BaseURL: baseURL,
		Model:   modelName,
	})
	if err != nil {
		return nil, err
	}
	return m, nil
}

func newThreadStoreFromEnv() agent.ThreadStore {
	mode := os.Getenv("THREAD_STORE")
	switch mode {
	case "file":
		dir := os.Getenv("THREAD_STORE_DIR")
		if dir == "" {
			dir = "./threads"
		}
		fs, err := agent.NewFileThreadStore(dir)
		if err != nil {
			log.Printf("NewFileThreadStore failed, fallback to memory, err=%v", err)
			return agent.NewMemoryThreadStore()
		}
		return fs
	case "redis":
		addr := os.Getenv("REDIS_ADDR")
		if addr == "" {
			addr = "127.0.0.1:6379"
		}
		client := redis.NewClient(&redis.Options{
			Addr: addr,
		})
		return agent.NewRedisThreadStore(client, os.Getenv("REDIS_THREAD_PREFIX"))
	default:
		return agent.NewMemoryThreadStore()
	}
}

func main() {
	ctx := context.Background()

	var (
		chatModel agent.ChatModel
		err       error
	)

	if os.Getenv("OPENAI_API_KEY") != "" {
		chatModel, err = newOpenAIModel(ctx)
		if err != nil {
			log.Fatalf("newOpenAIModel failed, err=%v", err)
		}
	} else {
		chatModel, err = newOllamaModel(ctx)
		if err != nil {
			log.Fatalf("newOllamaModel failed, err=%v", err)
		}
	}

	sum, err := utils.InferTool("sum", "Calculate a + b", sumTool)
	if err != nil {
		log.Fatalf("InferTool failed, err=%v", err)
	}

	httpTool, err := utils.InferTool("http_request", "Perform an HTTP request to an external HTTP/HTTPS endpoint", httpRequestTool)
	if err != nil {
		log.Fatalf("InferTool http_request failed, err=%v", err)
	}

	calcAgent, err := agent.NewChatAgent(ctx, agent.ChatAgentConfig{
		Name:         "calculator",
		Instructions: "You are a precise calculator. Use tools to compute numeric results.",
		Model:        chatModel,
		Tools:        []tool.BaseTool{sum},
	})
	if err != nil {
		log.Fatalf("NewChatAgent failed, err=%v", err)
	}

	store := newThreadStoreFromEnv()
	svc := agent.NewAgentService(store)

	thread, resp, err := svc.Send(ctx, calcAgent, "", "帮我计算 1.5 加 2.5 等于多少？请用中文回答。", model.WithTemperature(0.2))
	if err != nil {
		log.Fatalf("first turn failed, err=%v", err)
	}

	fmt.Println("thread:", thread.ID)
	fmt.Println("assistant:", resp.Content)

	_, resp2, err := svc.Send(ctx, calcAgent, thread.ID, "基于刚才的结果，再计算乘以 3。", model.WithTemperature(0.2))
	if err != nil {
		log.Fatalf("second turn failed, err=%v", err)
	}

	fmt.Println("assistant:", resp2.Content)

	explainer, err := agent.NewChatAgent(ctx, agent.ChatAgentConfig{
		Name:         "explainer",
		Instructions: "You explain the previous assistant result in simple Chinese.",
		Model:        chatModel,
	})
	if err != nil {
		log.Fatalf("NewChatAgent explainer failed, err=%v", err)
	}

	logging := agent.NewLoggingMiddleware(func(name, input string, duration time.Duration, err error) {
		log.Printf("agent=%s input=%q duration=%s err=%v", name, input, duration, err)
	})

	loggedExplainer := agent.WrapAgent(explainer, logging)

	wf := agent.NewSequentialWorkflow("calc_then_explain", calcAgent, loggedExplainer)

	resp3, err := wf.Run(ctx, "请计算 10 + 20，并向非技术用户解释含义。", model.WithTemperature(0.3))
	if err != nil {
		log.Fatalf("workflow run failed, err=%v", err)
	}

	fmt.Println("workflow result:", resp3.Content)

	perfExpert, err := agent.NewChatAgent(ctx, agent.ChatAgentConfig{
		Name:         "perf_expert",
		Instructions: "你是性能专家，请从性能角度评价上一个回答。",
		Model:        chatModel,
	})
	if err != nil {
		log.Fatalf("NewChatAgent perf_expert failed, err=%v", err)
	}

	costExpert, err := agent.NewChatAgent(ctx, agent.ChatAgentConfig{
		Name:         "cost_expert",
		Instructions: "你是成本专家，请从成本角度评价上一个回答。",
		Model:        chatModel,
	})
	if err != nil {
		log.Fatalf("NewChatAgent cost_expert failed, err=%v", err)
	}

	maintainExpert, err := agent.NewChatAgent(ctx, agent.ChatAgentConfig{
		Name:         "maintain_expert",
		Instructions: "你是可维护性专家，请从可维护性角度评价上一个回答。",
		Model:        chatModel,
	})
	if err != nil {
		log.Fatalf("NewChatAgent maintain_expert failed, err=%v", err)
	}

	summaryAgent, err := agent.NewChatAgent(ctx, agent.ChatAgentConfig{
		Name:         "summary_agent",
		Instructions: "请综合多个专家的意见，用简明中文给出一个总结性结论。",
		Model:        chatModel,
	})
	if err != nil {
		log.Fatalf("NewChatAgent summary_agent failed, err=%v", err)
	}

	aggWf := agent.NewAggregatingParallelWorkflow("multi_expert_review", summaryAgent, perfExpert, costExpert, maintainExpert)

	aggInput := "请为一个新的在线支付系统设计总体方案，并分别从性能、成本和可维护性角度给出建议，最后给出综合结论。"

	aggResp, err := aggWf.Run(ctx, aggInput, model.WithTemperature(0.4))
	if err != nil {
		log.Fatalf("AggregatingParallelWorkflow run failed, err=%v", err)
	}

	fmt.Println("aggregating workflow result:", aggResp.Content)

	router := agent.NewRoutingWorkflow("demo_router", chatModel, map[string]agent.Workflow{
		"calc_then_explain":   wf,
		"multi_expert_review": aggWf,
	})

	routerResp, err := router.Run(ctx, "我更关心多专家对一个方案的并行评审，请选择合适的工作流。", model.WithTemperature(0.2))
	if err != nil {
		log.Printf("RoutingWorkflow run failed, err=%v", err)
	} else {
		fmt.Println("router workflow result:", routerResp.Content)
	}

	toolCallingModel, ok := chatModel.(model.ToolCallingChatModel)
	if !ok {
		log.Println("underlying model does not support tool call, skip ReActAgent demo")
		return
	}

	reactAgentImpl, err := react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel: toolCallingModel,
	})
	if err != nil {
		log.Fatalf("react.NewAgent failed, err=%v", err)
	}

	reactOpts, err := react.WithTools(ctx, sum, httpTool)
	if err != nil {
		log.Fatalf("react.WithTools failed, err=%v", err)
	}

	reactAgent := agent.NewReActAgent("react_calculator", reactAgentImpl, reactOpts...)

	telemetry := agent.NewTelemetryMiddleware(func(m agent.AgentRunMetrics) {
		log.Printf("telemetry agent=%s tokens_total=%d tools=%d duration=%s err=%v", m.Name, m.TotalTokens, m.ToolCalls, m.Duration, m.Err)
	})

	safeOutput := agent.NewOutputFilterMiddleware(func(output string) (string, error) {
		if output == "" {
			return output, nil
		}

		redacted := output

		if strings.Contains(redacted, "密码") {
			redacted = strings.ReplaceAll(redacted, "密码", "[敏感信息]")
		}

		lower := strings.ToLower(redacted)
		if strings.Contains(lower, "password") {
			redacted = strings.ReplaceAll(redacted, "password", "[redacted]")
			redacted = strings.ReplaceAll(redacted, "Password", "[redacted]")
		}

		return redacted, nil
	})

	// 演示结构化日志
	structuredLogger := agent.NewStructuredLoggerMiddleware(os.Stdout)
	wrappedReAct := agent.WrapAgent(reactAgent, telemetry, safeOutput, structuredLogger)

	thread2, resp4, err := svc.Send(ctx, wrappedReAct.(agent.ThreadAwareAgent), "", "请使用工具计算 3.5 加 4.5，然后给出一步步的解释。", model.WithTemperature(0.2))
	if err != nil {
		log.Fatalf("ReActAgent first turn failed, err=%v", err)
	}

	fmt.Println("react thread:", thread2.ID)
	fmt.Println("react assistant:", resp4.Content)

	_, resp5, err := svc.Send(ctx, wrappedReAct.(agent.ThreadAwareAgent), thread2.ID, "基于刚才的结果，再计算乘以 2。", model.WithTemperature(0.2))
	if err != nil {
		log.Fatalf("ReActAgent second turn failed, err=%v", err)
	}

	fmt.Println("react assistant:", resp5.Content)

	hostCfg, err := agent.LoadHostConfig(strings.NewReader(demoHostConfigYAML))
	if err != nil {
		log.Fatalf("LoadHostConfig failed, err=%v", err)
	}

	modelFactory := func(ctx context.Context, name string) (agent.ChatModel, error) {
		_ = ctx
		if name != "default" {
			log.Printf("modelFactory: unknown model %q, fallback to default", name)
		}
		return chatModel, nil
	}

	toolRegistry := map[string]tool.BaseTool{
		"sum":          sum,
		"http_request": httpTool,
	}

	host, err := agent.NewHost(ctx, hostCfg, modelFactory, toolRegistry)
	if err != nil {
		log.Fatalf("NewHost failed, err=%v", err)
	}

	wfFromHost, ok := host.Workflow("host_router")
	if !ok {
		log.Fatalf("workflow host_router not found in host")
	}

	hostResp, err := wfFromHost.Run(ctx, "请根据我的描述选择合适的工作流（calc 或 http_review），并给出最终中文总结。", model.WithTemperature(0.3))
	if err != nil {
		log.Fatalf("host router workflow run failed, err=%v", err)
	}

	fmt.Println("host router workflow result:", hostResp.Content)

	plannerFromHost, ok := host.Workflow("host_planner")
	if !ok {
		log.Fatalf("workflow host_planner not found in host")
	}

	hostPlannerResp, err := plannerFromHost.Run(
		ctx,
		"请先执行 calc（计算 10+5），然后把结果作为输入执行 http_review（假设需要评审该计算结果），最后给出综合结论。",
		model.WithTemperature(0.3),
	)
	if err != nil {
		log.Fatalf("host planner workflow run failed, err=%v", err)
	}

	fmt.Println("host planner workflow result:", hostPlannerResp.Content)

	hostManager := agent.NewHostManager()
	// Register without specifying app name, relying on config's name "demo_app"
	hostManager.Register("", host)

	mgrPlanner, ok := hostManager.Workflow("demo_app", "host_planner")
	if ok {
		mgrResp, err := mgrPlanner.Run(ctx, "从 HostManager 视角再跑一次 host_planner，测试参数传递。", model.WithTemperature(0.3))
		if err != nil {
			log.Printf("HostManager planner workflow run failed, err=%v", err)
		} else {
			fmt.Println("host manager planner workflow result:", mgrResp.Content)
		}
	}
}
