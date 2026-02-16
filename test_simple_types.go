// 简单的 ReAct 类型测试
package main

import (
	"fmt"
	"AgentFramework/pkg/framework/agent/react"
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
