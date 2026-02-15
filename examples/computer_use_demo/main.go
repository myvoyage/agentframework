// Agent Framework - Computer Use 使用示例
// 演示如何使用计算机控制智能体进行视觉 GUI 交互

package main

import (
	"context"
	"fmt"
	"time"

	"AgentFramework/agent/computer_use"
)

func main() {
	ctx := context.Background()

	// 示例1: 创建 Computer Use Agent
	fmt.Println("=== Computer Use Agent 基本使用 ===\n")
	basicUsageExample(ctx)

	// 示例2: 自然语言命令
	fmt.Println("\n=== 自然语言命令控制 ===\n")
	naturalLanguageExample(ctx)

	// 示例3: 操作历史
	fmt.Println("\n=== 操作历史记录 ===\n")
	historyExample(ctx)

	// 示例4: 跨平台支持
	fmt.Println("\n=== 跨平台功能演示 ===\n")
	crossPlatformExample(ctx)

	// 示例5: 高级用法
	fmt.Println("\n=== 高级用法 ===\n")
	advancedUsageExample(ctx)
}

// basicUsageExample 基本使用示例
func basicUsageExample(ctx context.Context) {
	// 创建 Computer Use Agent
	// 注意: 实际使用需要配置真实的 LLM 模型
	fmt.Println("创建 Computer Use Agent:")

	config := computer_use.ComputerUseConfig{
		// Model: model, // 需要实际的模型
		ScreenshotDir: "/tmp/computer_use_screenshots",
		MaxHistory:    100,
		DisplayWidth:  1920,
		DisplayHeight: 1080,
	}

	fmt.Printf("  配置: 截图目录=%s, 历史=%d, 分辨率=%dx%d\n",
		config.ScreenshotDir, config.MaxHistory, config.DisplayWidth, config.DisplayHeight)

	// agent, _ := computer_use.NewComputerUseAgent(config)
	fmt.Println("\n支持的操作类型:")
	actions := []struct {
		typ   computer_use.ActionType
		desc  string
		example string
	}{
		{computer_use.ActionTypeClick, "单击", "点击屏幕坐标 (100, 200)"},
		{computer_use.ActionTypeDoubleClick, "双击", "双击屏幕坐标 (100, 200)"},
		{computer_use.ActionTypeRightClick, "右键点击", "右键点击屏幕坐标 (100, 200)"},
		{computer_use.ActionTypeType, "输入文本", "输入文本 'Hello World'"},
		{computer_use.ActionTypeKeyPress, "按键", "按下 Enter 键"},
		{computer_use.ActionTypeScroll, "滚动", "向下滚动 300 像素"},
		{computer_use.ActionTypeWait, "等待", "等待 1000 毫秒"},
		{computer_use.ActionTypeScreenshot, "截图", "截取当前屏幕"},
	}

	for _, action := range actions {
		fmt.Printf("  %-15s - %s (例: %s)\n", action.typ, action.desc, action.example)
	}
}

// naturalLanguageExample 自然语言命令示例
func naturalLanguageExample(ctx context.Context) {
	commands := []string{
		"点击屏幕中央的登录按钮",
		"在搜索框输入 'AI Agent'",
		"按 Enter 键",
		"向下滚动页面",
		"截图保存",
	}

	fmt.Println("自然语言命令示例:")
	fmt.Println("Agent 会自动解析命令并执行相应操作:\n")

	for _, cmd := range commands {
		fmt.Printf("  命令: \"%s\"\n", cmd)

		// 模拟解析结果
		var actionType computer_use.ActionType
		switch {
		case cmd == "点击屏幕中央的登录按钮":
			actionType = computer_use.ActionTypeClick
		case cmd == "在搜索框输入 'AI Agent'":
			actionType = computer_use.ActionTypeType
		case cmd == "按 Enter 键":
			actionType = computer_use.ActionTypeKeyPress
		case cmd == "向下滚动页面":
			actionType = computer_use.ActionTypeScroll
		case cmd == "截图保存":
			actionType = computer_use.ActionTypeScreenshot
		}

		fmt.Printf("  → 解析为操作: %s\n\n", actionType)
	}

	fmt.Println("使用方法:")
	fmt.Println("  action, _ := agent.ExecuteCommand(ctx, \"点击登录按钮\")")
}

// historyExample 操作历史示例
func historyExample(ctx context.Context) {
	fmt.Println("操作历史功能:")
	fmt.Println()

	// 模拟一些操作
	actions := []computer_use.ComputerAction{
		{
			Type:      computer_use.ActionTypeClick,
			Timestamp: time.Now().Add(-10 * time.Second),
			Coordinates: &computer_use.Coordinates{X: 100, Y: 200},
		},
		{
			Type:      computer_use.ActionTypeType,
			Timestamp: time.Now().Add(-8 * time.Second),
			Text:      "Hello World",
		},
		{
			Type:      computer_use.ActionTypeKeyPress,
			Timestamp: time.Now().Add(-5 * time.Second),
			Key:       "Enter",
		},
	}

	fmt.Println("操作历史记录:")
	for i, action := range actions {
		fmt.Printf("  %d. [%s] %s", i+1, action.Timestamp.Format("15:04:05"), action.Type)
		if action.Coordinates != nil {
			fmt.Printf(" @ (%d, %d)", action.Coordinates.X, action.Coordinates.Y)
		}
		if action.Text != "" {
			fmt.Printf(" text=\"%s\"", action.Text)
		}
		if action.Key != "" {
			fmt.Printf(" key=%s", action.Key)
		}
		fmt.Println()
	}

	fmt.Println("\n使用方法:")
	fmt.Println("  history := agent.GetHistory()")
	fmt.Println("  for _, action := range history {")
	fmt.Println("    fmt.Printf(\"%v\\n\", action)")
	fmt.Println("  }")
}

// crossPlatformExample 跨平台支持示例
func crossPlatformExample(ctx context.Context) {
	fmt.Println("跨平台支持:")
	fmt.Println()

	platforms := []struct {
		os    string
		tools []string
	}{
		{
			os:    "macOS",
			tools: []string{"AppleScript", "osascript", "cliclick"},
		},
		{
			os:    "Linux",
			tools: []string{"xdotool", "ydotool"},
		},
		{
			os:    "Windows",
			tools: []string{"PowerShell", "user32.dll"},
		},
	}

	for _, platform := range platforms {
		fmt.Printf("  %s:\n", platform.os)
		for _, tool := range platform.tools {
			fmt.Printf("    - %s\n", tool)
		}
		fmt.Println()
	}

	fmt.Println("每个平台的功能支持:")
	features := []struct {
		feature string
		macos   bool
		linux   bool
		windows bool
	}{
		{"鼠标点击", true, true, true},
		{"文本输入", true, true, true},
		{"按键模拟", true, true, true},
		{"滚动", true, true, true},
		{"截图", true, true, true},
		{"光标移动", true, true, true},
	}

	fmt.Println("  功能              macOS  Linux  Windows")
	fmt.Println("  " + "------------------------------------------------")
	for _, f := range features {
		fmt.Printf("  %-18s  %-5s  %-5s  %-5s\n",
			f.feature,
			boolToIcon(f.macos),
			boolToIcon(f.linux),
			boolToIcon(f.windows))
	}
}

// advancedUsageExample 高级用法示例
func advancedUsageExample(ctx context.Context) {
	fmt.Println("高级用法:")
	fmt.Println()

	examples := []struct {
		title string
		code  string
		desc  string
	}{
		{
			title: "1. 自定义显示配置",
			code: `config := computer_use.ComputerUseConfig{
    Model:         model,
    ScreenshotDir: "/tmp/screenshots",
    DisplayWidth:  2560,  // 2K 显示器
    DisplayHeight: 1440,
    MaxHistory:    200,   // 保存更多历史
}`,
			desc: "根据实际显示配置自定义分辨率",
		},
		{
			title: "2. 获取显示信息",
			code: `display := agent.GetDisplayInfo()
fmt.Printf("分辨率: %dx%d\n", display.Width, display.Height)
fmt.Printf("缩放: %.2f\n", display.Scale)`,
			desc: "动态获取显示配置信息",
		},
		{
			title: "3. 获取光标位置",
			code: `cursor := agent.GetCursorInfo()
fmt.Printf("光标位置: (%d, %d)\n", cursor.X, cursor.Y)`,
			desc: "追踪当前光标坐标",
		},
		{
			title: "4. 组合操作序列",
			code: `// 点击搜索框
agent.ExecuteCommand(ctx, "点击搜索框")
// 输入搜索词
agent.ExecuteCommand(ctx, "输入 'AI Agent'")
// 按回车搜索
agent.ExecuteCommand(ctx, "按 Enter 键")`,
			desc: "执行复杂的操作序列",
		},
		{
			title: "5. 带超时的操作",
			code: `ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
action, err := agent.ExecuteCommand(ctx, "打开应用")`,
			desc: "使用 context 控制操作超时",
		},
	}

	for _, ex := range examples {
		fmt.Printf("%s\n", ex.title)
		fmt.Printf("描述: %s\n", ex.desc)
		fmt.Printf("代码:\n%s\n\n", ex.code)
	}

	fmt.Println("实际应用场景:")
	scenarios := []string{
		"自动化测试 - 模拟用户操作进行 UI 测试",
		"RPA 流程自动化 - 重复性任务的自动化",
		"辅助功能 - 帮助视障用户操作电脑",
		"演示录制 - 自动生成软件演示视频",
		"批量操作 - 自动执行批量 GUI 操作",
	}

	for _, scenario := range scenarios {
		fmt.Printf("  • %s\n", scenario)
	}
}

// boolToIcon 将布尔值转换为图标
func boolToIcon(b bool) string {
	if b {
		return "✅"
	}
	return "❌"
}

func init() {
	// 确保示例目录存在
	// os.MkdirAll("examples/computer_use_demo/screenshots", 0755)
}
