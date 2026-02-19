// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later


package hello

import (
	"fmt"
)

type HelloTool struct{}

func (t *HelloTool) Name() string    { return "hello" }
func (t *HelloTool) Version() string { return "0.1" }
func (t *HelloTool) Execute(ctx interface{}, inputs map[string]interface{}) (map[string]interface{}, error) {
	name := "World"
	if v, ok := inputs["name"]; ok {
		if s, ok := v.(string); ok && s != "" {
			name = s
		}
	}
	greeting := fmt.Sprintf("Hello, %s!", name)
	return map[string]interface{}{"greeting": greeting}, nil
}

func NewHelloTool() *HelloTool { return &HelloTool{} }
