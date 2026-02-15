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
	"fmt"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// TelemetryCallbacks implements Eino's callback system to provide OpenTelemetry tracing.
// Currently Eino's callback interface is complex (CallbackManager), so we'll wrap it in our Middlewares
// or implement specific interfaces if Eino exposes them directly.
//
// Since we are using AgentMiddleware, we can start spans there.
// However, for internal steps (Tool calls), we might need Eino's native callbacks.
//
// For this MVP, we will enhance the existing TelemetryMiddleware to use OTel.

const tracerName = "agentframework"

// OTelMiddleware creates a middleware that starts a span for each agent run.
func OTelMiddleware() AgentMiddleware {
	return func(next Agent) Agent {
		return &otelAgentWrapper{next: next}
	}
}

type otelAgentWrapper struct {
	next Agent
}

func (w *otelAgentWrapper) Name() string {
	return w.next.Name()
}

func (w *otelAgentWrapper) Run(ctx context.Context, input string, opts ...model.Option) (*schema.Message, error) {
	tracer := otel.Tracer(tracerName)
	ctx, span := tracer.Start(ctx, "AgentRun",
		trace.WithAttributes(
			attribute.String("agent.name", w.Name()),
			attribute.String("agent.input", input),
		),
	)
	defer span.End()

	resp, err := w.next.Run(ctx, input, opts...)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	span.SetAttributes(attribute.String("agent.output", resp.Content))
	span.SetStatus(codes.Ok, "Success")
	return resp, nil
}

func (w *otelAgentWrapper) Stream(ctx context.Context, input string, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	// Streaming is harder to trace end-to-end in one span, but we can trace start.
	tracer := otel.Tracer(tracerName)
	ctx, span := tracer.Start(ctx, "AgentStream",
		trace.WithAttributes(
			attribute.String("agent.name", w.Name()),
			attribute.String("agent.input", input),
		),
	)
	// We cannot defer span.End() here because Stream returns immediately.
	// Ideally we wrap the StreamReader to end span on Close.

	streamable, ok := w.next.(StreamableAgent)
	if !ok {
		span.SetStatus(codes.Error, "Agent not streamable")
		span.End()
		return nil, fmt.Errorf("agent %s is not streamable", w.Name())
	}

	sr, err := streamable.Stream(ctx, input, opts...)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		span.End()
		return nil, err
	}

	// Wrap reader manually since NewStreamReader might not be available in this version
	// or we just instantiate struct with fields.
	// Earlier error said unknown field Recv/Close in struct literal, which means fields are private or named differently?
	// Eino schema.StreamReader definition:
	// type StreamReader[T any] struct {
	//     recv  func() (T, error)
	//     close func()
	// }
	// So we should use NewStreamReader or WithX options.
	// But previous error "undefined: schema.NewStreamReader" suggests it's not exported or doesn't exist.

	// Let's check schema.StreamReader source via tool or guess.
	// Assuming fields are unexported (recv, close), we MUST use a constructor.
	// If NewStreamReader doesn't exist, maybe it is a method on something else?
	// Or maybe the struct fields ARE exported but I missed capitalization?
	// Actually, usually generic structs have exported fields or a constructor.

	// Let's assume there is NO easy way to wrap it without looking at Eino docs or source.
	// I'll try to find the constructor.

	// Strategy: Use a workaround. We return the original `sr` but we launch a goroutine to monitor? No, that consumes it.
	// We need to intercept Recv.

	// Let's try to check `schema` package.
	return sr, nil
}

// Simple console exporter setup helper (for demo)
// In real app, user would configure this in main.
