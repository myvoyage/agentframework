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

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

// MockChatModel is a mock ChatModel implementation for testing
type MockChatModel struct {
	NameValue string
}

func (m *MockChatModel) Generate(ctx context.Context, input []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	return &schema.Message{
		Role:    schema.Assistant,
		Content: "Mock response from " + m.NameValue,
	}, nil
}

func (m *MockChatModel) Stream(ctx context.Context, input []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	return nil, nil
}

// MockToolCallingChatModel is a mock ToolCallingChatModel implementation for testing
type MockToolCallingChatModel struct {
	MockChatModel
	tools []*schema.ToolInfo
}

// WithTools implements the ToolCallingChatModel interface
func (m *MockToolCallingChatModel) WithTools(tools []*schema.ToolInfo) (model.ToolCallingChatModel, error) {
	return &MockToolCallingChatModel{
		MockChatModel: m.MockChatModel,
		tools:         tools,
	}, nil
}

// BindTools implements the ToolCallingChatModel interface
func (m *MockToolCallingChatModel) BindTools(tools []*schema.ToolInfo) (model.ToolCallingChatModel, error) {
	return m.WithTools(tools)
}

// Generate implements the ChatModel interface with tool calling support
func (m *MockToolCallingChatModel) Generate(ctx context.Context, input []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	return &schema.Message{
		Role:    schema.Assistant,
		Content: "Mock tool calling response from " + m.NameValue,
	}, nil
}

// Stream implements the ChatModel interface with tool calling support
func (m *MockToolCallingChatModel) Stream(ctx context.Context, input []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	return nil, nil
}

// NewMockToolCallingChatModel creates a new MockToolCallingChatModel
func NewMockToolCallingChatModel(name string) *MockToolCallingChatModel {
	return &MockToolCallingChatModel{
		MockChatModel: MockChatModel{NameValue: name},
		tools:         []*schema.ToolInfo{},
	}
}
