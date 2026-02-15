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

package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	callbacksHelper "github.com/cloudwego/eino/utils/callbacks"
)

type AgentMiddleware func(Agent) Agent

func WrapAgent(ag Agent, mws ...AgentMiddleware) Agent {
	wrapped := ag
	for i := len(mws) - 1; i >= 0; i-- {
		wrapped = mws[i](wrapped)
	}
	return wrapped
}

type AgentFunc func(ctx context.Context, input string, opts ...model.Option) (*schema.Message, error)

type FuncAgent struct {
	name string
	fn   AgentFunc
}

func (f *FuncAgent) Name() string {
	return f.name
}

func (f *FuncAgent) Run(ctx context.Context, input string, opts ...model.Option) (*schema.Message, error) {
	return f.fn(ctx, input, opts...)
}

func NewLoggingMiddleware(logger func(name, input string, duration time.Duration, err error)) AgentMiddleware {
	return func(next Agent) Agent {
		return &FuncAgent{
			name: next.Name(),
			fn: func(ctx context.Context, input string, opts ...model.Option) (*schema.Message, error) {
				start := time.Now()
				resp, err := next.Run(ctx, input, opts...)
				logger(next.Name(), input, time.Since(start), err)
				return resp, err
			},
		}
	}
}

func NewStructuredLoggerMiddleware(w io.Writer) AgentMiddleware {
	return NewLoggingMiddleware(func(name, input string, duration time.Duration, err error) {
		logEntry := map[string]interface{}{
			"time":     time.Now().Format(time.RFC3339),
			"agent":    name,
			"input":    input,
			"duration": duration.String(),
		}
		if err != nil {
			logEntry["error"] = err.Error()
		}

		bytes, _ := json.Marshal(logEntry)
		fmt.Fprintln(w, string(bytes))
	})
}

type AgentRunMetrics struct {
	Name             string
	Input            string
	Duration         time.Duration
	Err              error
	PromptTokens     int64
	CompletionTokens int64
	TotalTokens      int64
	ToolCalls        int64
}

func NewTelemetryMiddleware(collector func(m AgentRunMetrics)) AgentMiddleware {
	return func(next Agent) Agent {
		return &FuncAgent{
			name: next.Name(),
			fn: func(ctx context.Context, input string, opts ...model.Option) (*schema.Message, error) {
				metrics := AgentRunMetrics{
					Name:  next.Name(),
					Input: input,
				}

				modelHandler := &callbacksHelper.ModelCallbackHandler{
					OnEnd: func(ctx context.Context, info *callbacks.RunInfo, output *model.CallbackOutput) context.Context {
						if output != nil && output.TokenUsage != nil {
							metrics.PromptTokens += int64(output.TokenUsage.PromptTokens)
							metrics.CompletionTokens += int64(output.TokenUsage.CompletionTokens)
							metrics.TotalTokens += int64(output.TokenUsage.TotalTokens)
						}
						return ctx
					},
				}

				toolHandler := &callbacksHelper.ToolCallbackHandler{
					OnStart: func(ctx context.Context, info *callbacks.RunInfo, input *tool.CallbackInput) context.Context {
						metrics.ToolCalls++
						return ctx
					},
				}

				handler := callbacksHelper.NewHandlerHelper().
					ChatModel(modelHandler).
					Tool(toolHandler).
					Handler()

				ctxWithCallbacks := callbacks.InitCallbacks(ctx, &callbacks.RunInfo{
					Name: next.Name(),
				}, handler)

				start := time.Now()
				resp, err := next.Run(ctxWithCallbacks, input, opts...)
				metrics.Duration = time.Since(start)
				metrics.Err = err

				if collector != nil {
					collector(metrics)
				}

				return resp, err
			},
		}
	}
}

func NewInputFilterMiddleware(filter func(input string) (string, error)) AgentMiddleware {
	return func(next Agent) Agent {
		return &FuncAgent{
			name: next.Name(),
			fn: func(ctx context.Context, input string, opts ...model.Option) (*schema.Message, error) {
				filtered, err := filter(input)
				if err != nil {
					return nil, err
				}
				return next.Run(ctx, filtered, opts...)
			},
		}
	}
}

func NewOutputFilterMiddleware(filter func(output string) (string, error)) AgentMiddleware {
	return func(next Agent) Agent {
		return &FuncAgent{
			name: next.Name(),
			fn: func(ctx context.Context, input string, opts ...model.Option) (*schema.Message, error) {
				resp, err := next.Run(ctx, input, opts...)
				if err != nil || resp == nil {
					return resp, err
				}
				checked, err := filter(resp.Content)
				if err != nil {
					return nil, err
				}
				resp.Content = checked
				return resp, nil
			},
		}
	}
}

type InputPolicy func(ctx context.Context, input string) error

func NewInputPolicyMiddleware(policy InputPolicy) AgentMiddleware {
	return func(next Agent) Agent {
		return &FuncAgent{
			name: next.Name(),
			fn: func(ctx context.Context, input string, opts ...model.Option) (*schema.Message, error) {
				if policy != nil {
					if err := policy(ctx, input); err != nil {
						return nil, err
					}
				}
				return next.Run(ctx, input, opts...)
			},
		}
	}
}
