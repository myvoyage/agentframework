// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later


// 测试 ReAct Agent 类型定义

package main

import (
	"fmt"
	"time"

	"AgentFramework/pkg/framework/agent/react"
)

func main() {
	fmt.Println("Testing ReAct Agent types...")
	
	// 测试 ReActActionType
	actionType := react.ActionTypeTool
	fmt.Printf("Action Type: %s\n", actionType.String())
	fmt.Printf("Is Valid: %v\n", actionType.IsValid())
	
	// 测试 Thought 创建
	thought := react.NewThought("Test content", "Test reasoning", 0.9)
	fmt.Printf("\nThought ID: %s\n", thought.ID)
	fmt.Printf("Thought Content: %s\n", thought.Content)
	
	// 测试 Action 创建
	action := react.NewAction(react.ActionTypeTool, "read_file", map[string]interface{}{
		"path": "/test/file.txt",
	})
	fmt.Printf("\nAction ID: %s\n", action.ID)
	fmt.Printf("Action Name: %s\n", action.Name)
	
	// 测试 Observation 创建
	observation := react.NewObservation(action.ID, true, "Success result", "")
	fmt.Printf("\nObservation ID: %s\n", observation.ID)
	fmt.Printf("Observation Summary: %s\n", observation.ResultSummary())
	
	// 测试 ReActState 创建
	state := &react.ReActState{
		SessionID: "test-session",
		AgentID:   "test-agent",
		Query:     "Test query",
		Status:    react.ReActStatusThinking,
		StartTime: time.Now(),
	}
	
	if err := state.Validate(); err != nil {
		fmt.Printf("\nState validation error: %v\n", err)
	} else {
		fmt.Printf("\nState is valid!\n")
	}
	
	fmt.Println("\nAll tests passed!")
}
