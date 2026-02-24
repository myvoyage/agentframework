// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later


// 简单的 ReAct 类型测试
package main

import (
	"fmt"
	"AgentFramework/pkg/framework/agent/react_experimental"
)

func main() {
	fmt.Println("Testing ReAct Agent Types")
	
	// 测试 1: ReActActionType
	t := react.ActionTypeTool
	fmt.Printf("✓ ReActActionType: %s (valid: %v)\n", t.String(), t.IsValid())
	
	// 测试 2: Capability
	c := react.CapabilityReasoning
	fmt.Printf("✓ Capability: %s (valid: %v)\n", c.String(), c.IsValid())
	
	// 测试 3: ReActStatus
	s := react.ReActStatusThinking
	fmt.Printf("✓ ReActStatus: %s\n", s.String())
	
	fmt.Println("\nAll basic type tests passed!")
}
