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

package computer_use

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

// ComputerUseAgent 计算机使用智能体，支持视觉 GUI 交互
type ComputerUseAgent struct {
	model          ChatModel
	screenshotPath string
	display        DisplayInfo
	cursor         CursorInfo
	mu             sync.RWMutex
	lastAction     *ComputerAction
	actionHistory  []ComputerAction
	maxHistory     int
}

// ChatModel LLM模型接口
type ChatModel interface {
	Generate(ctx context.Context, input []*schema.Message, opts ...model.Option) (*schema.Message, error)
}

// DisplayInfo 显示信息
type DisplayInfo struct {
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	Scale       float64 `json:"scale"`
	Primary     bool   `json:"primary"`
	DisplayName string  `json:"display_name"`
}

// CursorInfo 光标信息
type CursorInfo struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// ComputerAction 计算机操作
type ComputerAction struct {
	Type        ActionType             `json:"type"`
	Timestamp   time.Time              `json:"timestamp"`
	Coordinates *Coordinates           `json:"coordinates,omitempty"`
	Text        string                 `json:"text,omitempty"`
	Key         string                 `json:"key,omitempty"`
	Duration    int                    `json:"duration,omitempty"`
	Button      string                 `json:"button,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Result      string                 `json:"result,omitempty"`
	Error       string                 `json:"error,omitempty"`
}

// ActionType 操作类型
type ActionType string

const (
	ActionTypeClick      ActionType = "click"
	ActionTypeDoubleClick ActionType = "double_click"
	ActionTypeRightClick ActionType = "right_click"
	ActionTypeDrag       ActionType = "drag"
	ActionTypeType       ActionType = "type"
	ActionTypeKeyPress   ActionType = "key_press"
	ActionTypeScroll     ActionType = "scroll"
	ActionTypeWait       ActionType = "wait"
	ActionTypeScreenshot ActionType = "screenshot"
)

// Coordinates 坐标
type Coordinates struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// ComputerUseConfig Computer Use 配置
type ComputerUseConfig struct {
	Model         ChatModel
	ScreenshotDir string
	MaxHistory    int
	DisplayWidth  int
	DisplayHeight int
}

// NewComputerUseAgent 创建计算机使用智能体
func NewComputerUseAgent(config ComputerUseConfig) (*ComputerUseAgent, error) {
	if config.Model == nil {
		return nil, fmt.Errorf("model is required")
	}

	// 创建截图目录
	if config.ScreenshotDir == "" {
		config.ScreenshotDir = os.TempDir() + "/computer_use_screenshots"
	}
	if err := os.MkdirAll(config.ScreenshotDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create screenshot directory: %w", err)
	}

	// 获取显示信息
	display := DisplayInfo{
		Width:   config.DisplayWidth,
		Height:  config.DisplayHeight,
		Scale:   1.0,
		Primary: true,
	}

	if display.Width == 0 || display.Height == 0 {
		display = getDefaultDisplayInfo()
	}

	maxHistory := config.MaxHistory
	if maxHistory == 0 {
		maxHistory = 100
	}

	return &ComputerUseAgent{
		model:         config.Model,
		screenshotPath: config.ScreenshotDir,
		display:       display,
		cursor:        CursorInfo{X: 0, Y: 0},
		actionHistory: make([]ComputerAction, 0),
		maxHistory:    maxHistory,
	}, nil
}

// ExecuteCommand 执行自然语言命令
func (a *ComputerUseAgent) ExecuteCommand(ctx context.Context, command string) (*ComputerAction, error) {
	// 分析命令
	action, err := a.parseCommand(ctx, command)
	if err != nil {
		return nil, fmt.Errorf("failed to parse command: %w", err)
	}

	// 执行操作
	result, err := a.executeAction(ctx, action)
	if err != nil {
		action.Error = err.Error()
	} else {
		action.Result = result
	}

	// 记录历史
	a.mu.Lock()
	a.actionHistory = append(a.actionHistory, *action)
	if len(a.actionHistory) > a.maxHistory {
		a.actionHistory = a.actionHistory[1:]
	}
	a.lastAction = action
	a.mu.Unlock()

	return action, err
}

// parseCommand 解析命令为操作
func (a *ComputerUseAgent) parseCommand(ctx context.Context, command string) (*ComputerAction, error) {
	// 获取当前截图
	screenshot, err := a.TakeScreenshot(ctx)
	if err != nil {
		return nil, err
	}

	// 构建提示词
	prompt := a.buildParsePrompt(command, screenshot)

	systemMsg := &schema.Message{
		Role:    schema.System,
		Content: a.getSystemPrompt(),
	}

	userMsg := &schema.Message{
		Role:    schema.User,
		Content: prompt,
	}

	// 调用模型
	resp, err := a.model.Generate(ctx, []*schema.Message{systemMsg, userMsg})
	if err != nil {
		return nil, err
	}

	// 解析响应
	var action ComputerAction
	if err := json.Unmarshal([]byte(resp.Content), &action); err != nil {
		// 如果解析失败，尝试简单解析
		return a.simpleParseCommand(command), nil
	}

	action.Timestamp = time.Now()
	return &action, nil
}

// executeAction 执行操作
func (a *ComputerUseAgent) executeAction(ctx context.Context, action *ComputerAction) (string, error) {
	switch action.Type {
	case ActionTypeClick:
		return a.click(action.Coordinates.X, action.Coordinates.Y)

	case ActionTypeDoubleClick:
		return a.doubleClick(action.Coordinates.X, action.Coordinates.Y)

	case ActionTypeRightClick:
		return a.rightClick(action.Coordinates.X, action.Coordinates.Y)

	case ActionTypeType:
		return a.typing(action.Text)

	case ActionTypeKeyPress:
		return a.keyPress(action.Key)

	case ActionTypeScroll:
		return a.scroll(action.Coordinates.X, action.Coordinates.Y)

	case ActionTypeWait:
		return a.wait(time.Duration(action.Duration) * time.Millisecond)

	case ActionTypeScreenshot:
		return a.TakeScreenshot(ctx)

	default:
		return "", fmt.Errorf("unknown action type: %s", action.Type)
	}
}

// TakeScreenshot 截图
func (a *ComputerUseAgent) TakeScreenshot(ctx context.Context) (string, error) {
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s/screenshot_%s.png", a.screenshotPath, timestamp)

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		// macOS
		cmd = exec.Command("screencapture", "-x", filename)
	case "linux":
		// Linux (尝试多个工具)
		if _, err := exec.LookPath("gnome-screenshot"); err == nil {
			cmd = exec.Command("gnome-screenshot", "-f", filename)
		} else if _, err := exec.LookPath("scrot"); err == nil {
			cmd = exec.Command("scrot", filename)
		} else {
			return "", fmt.Errorf("no screenshot tool found (gnome-screenshot or scrot required)")
		}
	case "windows":
		// Windows (需要第三方工具或 PowerShell)
		psScript := fmt.Sprintf(`
Add-Type -AssemblyName System.Windows.Forms
$bounds = [System.Windows.Forms.Screen]::PrimaryScreen.Bounds
$bmp = New-Object System.Drawing.Bitmap $bounds.width, $bounds.height
$graphics = [System.Drawing.Graphics]::FromImage($bmp)
$graphics.CopyFromScreen($bounds.Location, [System.Drawing.Point]::Empty, $bounds.size)
$bmp.Save("%s")
$graphics.Dispose()
$bmp.Dispose()
`, filename)
		cmd = exec.Command("powershell", "-Command", psScript)
	default:
		return "", fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("screenshot failed: %w", err)
	}

	// 编码为 base64
	imgData, err := os.ReadFile(filename)
	if err != nil {
		return filename, nil // 返回文件名，即使编码失败
	}

	base64Str := base64.StdEncoding.EncodeToString(imgData)
	return fmt.Sprintf("data:image/png;base64,%s", base64Str), nil
}

// click 点击
func (a *ComputerUseAgent) click(x, y int) (string, error) {
	if err := moveCursor(x, y); err != nil {
		return "", err
	}
	time.Sleep(50 * time.Millisecond)
	return clickMouse("left")
}

// doubleClick 双击
func (a *ComputerUseAgent) doubleClick(x, y int) (string, error) {
	if err := moveCursor(x, y); err != nil {
		return "", err
	}

	// 快速双击
	for i := 0; i < 2; i++ {
		if _, err := clickMouse("left"); err != nil {
			return "", err
		}
		time.Sleep(50 * time.Millisecond)
	}

	return "double clicked", nil
}

// rightClick 右键点击
func (a *ComputerUseAgent) rightClick(x, y int) (string, error) {
	if err := moveCursor(x, y); err != nil {
		return "", err
	}
	time.Sleep(50 * time.Millisecond)
	return clickMouse("right")
}

// typing 输入文本
func (a *ComputerUseAgent) typing(text string) (string, error) {
	switch runtime.GOOS {
	case "darwin":
		// macOS 使用 osascript
		script := fmt.Sprintf(`tell application "System Events" to keystroke "%s"`, escapeAppleScript(text))
		cmd := exec.Command("osascript", "-e", script)
		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("typing failed: %w", err)
		}
	case "linux":
		// Linux 使用 xdotool 或 ydotool
		if _, err := exec.LookPath("xdotool"); err == nil {
			cmd := exec.Command("xdotool", "type", "--", text)
			if err := cmd.Run(); err != nil {
				return "", fmt.Errorf("typing failed: %w", err)
			}
		} else if _, err := exec.LookPath("ydotool"); err == nil {
			cmd := exec.Command("ydotool", "type", "--key-delay", "5", text)
			if err := cmd.Run(); err != nil {
				return "", fmt.Errorf("typing failed: %w", err)
			}
		} else {
			return "", fmt.Errorf("no typing tool found (xdotool or ydotool required)")
		}
	case "windows":
		// Windows 使用 PowerShell
		psScript := fmt.Sprintf(`
Add-Type -AssemblyName System.Windows.Forms
$text = '%s'
[System.Windows.Forms.SendKeys]::SendWait($text)
`, escapePowerShell(text))
		cmd := exec.Command("powershell", "-Command", psScript)
		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("typing failed: %w", err)
		}
	default:
		return "", fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return fmt.Sprintf("typed: %s", text), nil
}

// keyPress 按键
func (a *ComputerUseAgent) keyPress(key string) (string, error) {
	switch runtime.GOOS {
	case "darwin":
		script := fmt.Sprintf(`tell application "System Events" to key code %s`, keyCodeForMac(key))
		cmd := exec.Command("osascript", "-e", script)
		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("key press failed: %w", err)
		}
	case "linux":
		if _, err := exec.LookPath("xdotool"); err == nil {
			cmd := exec.Command("xdotool", "key", key)
			if err := cmd.Run(); err != nil {
				return "", fmt.Errorf("key press failed: %w", err)
			}
		} else if _, err := exec.LookPath("ydotool"); err == nil {
			cmd := exec.Command("ydotool", "key", key)
			if err := cmd.Run(); err != nil {
				return "", fmt.Errorf("key press failed: %w", err)
			}
		} else {
			return "", fmt.Errorf("no key tool found (xdotool or ydotool required)")
		}
	case "windows":
		psScript := fmt.Sprintf(`
Add-Type -AssemblyName System.Windows.Forms
[System.Windows.Forms.SendKeys]::SendWait("%s")
`, mapKeyToSendKeys(key))
		cmd := exec.Command("powershell", "-Command", psScript)
		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("key press failed: %w", err)
		}
	default:
		return "", fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return fmt.Sprintf("key pressed: %s", key), nil
}

// scroll 滚动
func (a *ComputerUseAgent) scroll(x, y int) (string, error) {
	switch runtime.GOOS {
	case "darwin":
		// 模拟滚轮滚动
		script := fmt.Sprintf(`tell application "System Events" to scroll wheel %d by ¬
clicking button %d`, int(math.Abs(float64(y))), mapScrollDirection(y))
		cmd := exec.Command("osascript", "-e", script)
		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("scroll failed: %w", err)
		}
	case "linux":
		if _, err := exec.LookPath("xdotool"); err == nil {
			clicks := int(math.Abs(float64(y)))
			button := mapScrollDirection(y)
			for i := 0; i < clicks; i++ {
				cmd := exec.Command("xdotool", "click", "--repeat", "1", "--delay", "50", fmt.Sprintf("%d", button))
				if err := cmd.Run(); err != nil {
					return "", fmt.Errorf("scroll failed: %w", err)
				}
			}
		} else {
			return "", fmt.Errorf("no scroll tool found (xdotool required)")
		}
	case "windows":
		// Windows 使用 mouse_event 滚轮滚动
		// MOUSEEVENTF_WHEEL = 0x0800
		scrollAmount := int(math.Abs(float64(y))) * 120 // WHEEL_DELTA = 120
		if y < 0 {
			scrollAmount = -scrollAmount
		}

		psScript := fmt.Sprintf(`
Add-Type -TypeDefinition '
using System;
using System.Runtime.InteropServices;
public class Win32 {
    [DllImport("user32.dll")]
    public static extern void mouse_event(uint dwFlags, uint dx, uint dy, uint cButtons, uint dwExtraInfo);
}
'
[Win32]::mouse_event(0x0800, 0, 0, 0, %d)
`, uint32(scrollAmount&0xFFFFFFFF))
		cmd := exec.Command("powershell", "-Command", psScript)
		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("Windows scroll failed: %w", err)
		}
	default:
		return "", fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return fmt.Sprintf("scrolled: x=%d, y=%d", x, y), nil
}

// wait 等待
func (a *ComputerUseAgent) wait(duration time.Duration) (string, error) {
	time.Sleep(duration)
	return fmt.Sprintf("waited: %v", duration), nil
}

// GetHistory 获取操作历史
func (a *ComputerUseAgent) GetHistory() []ComputerAction {
	a.mu.RLock()
	defer a.mu.RUnlock()

	history := make([]ComputerAction, len(a.actionHistory))
	copy(history, a.actionHistory)
	return history
}

// GetDisplayInfo 获取显示信息
func (a *ComputerUseAgent) GetDisplayInfo() DisplayInfo {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.display
}

// GetCursorInfo 获取光标信息
func (a *ComputerUseAgent) GetCursorInfo() CursorInfo {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.cursor
}

// buildParsePrompt 构建解析提示词
func (a *ComputerUseAgent) buildParsePrompt(command string, screenshot string) string {
	return fmt.Sprintf(`# 计算机控制命令解析

用户命令: %s

显示信息:
- 分辨率: %dx%d
- 缩放: %.2f

当前截图: %s

请分析命令并返回计算机操作。坐标应该在显示范围内。
返回JSON格式，例如：
{
  "type": "click",
  "coordinates": {"x": 100, "y": 200}
}

支持的操作类型:
- click: 单击
- double_click: 双击
- right_click: 右键点击
- type: 输入文本 (需要 text 字段)
- key_press: 按键 (需要 key 字段)
- scroll: 滚动 (coordinates.x 为水平滚动, coordinates.y 为垂直滚动)
- wait: 等待 (需要 duration 字段，单位毫秒)
- screenshot: 截图
`, command, a.display.Width, a.display.Height, a.display.Scale, screenshot)
}

// getSystemPrompt 获取系统提示词
func (a *ComputerUseAgent) getSystemPrompt() string {
	return `你是一个计算机控制专家。将自然语言命令解析为结构化的计算机操作。

坐标系统:
- 原点 (0,0) 在屏幕左上角
- X 轴向右增加
- Y 轴向下增加

注意事项:
1. 确保坐标在显示范围内
2. 对于模糊的命令，做出合理的假设
3. 优先使用精确坐标而非描述
4. 考虑用户意图，提供最佳的操作序列`
}

// simpleParseCommand 简单命令解析（备用方案）
func (a *ComputerUseAgent) simpleParseCommand(command string) *ComputerAction {
	action := &ComputerAction{
		Timestamp: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	lowerCmd := strings.ToLower(command)

	switch {
	case strings.Contains(lowerCmd, "点击") || strings.Contains(lowerCmd, "click"):
		action.Type = ActionTypeClick
		// 默认点击中心
		action.Coordinates = &Coordinates{
			X: a.display.Width / 2,
			Y: a.display.Height / 2,
		}

	case strings.Contains(lowerCmd, "双击") || strings.Contains(lowerCmd, "double click"):
		action.Type = ActionTypeDoubleClick
		action.Coordinates = &Coordinates{
			X: a.display.Width / 2,
			Y: a.display.Height / 2,
		}

	case strings.Contains(lowerCmd, "右键") || strings.Contains(lowerCmd, "right click"):
		action.Type = ActionTypeRightClick
		action.Coordinates = &Coordinates{
			X: a.display.Width / 2,
			Y: a.display.Height / 2,
		}

	case strings.Contains(lowerCmd, "输入") || strings.Contains(lowerCmd, "type"):
		action.Type = ActionTypeType
		// 提取输入内容
		if idx := strings.Index(command, ":"); idx != -1 && idx < len(command)-1 {
			action.Text = strings.TrimSpace(command[idx+1:])
		}

	case strings.Contains(lowerCmd, "截图") || strings.Contains(lowerCmd, "screenshot"):
		action.Type = ActionTypeScreenshot

	case strings.Contains(lowerCmd, "等待") || strings.Contains(lowerCmd, "wait"):
		action.Type = ActionTypeWait
		action.Duration = 1000 // 默认1秒

	default:
		action.Type = ActionTypeWait
		action.Duration = 500
	}

	return action
}

// ==================== 辅助函数 ====================

// getDefaultDisplayInfo 获取默认显示信息
func getDefaultDisplayInfo() DisplayInfo {
	// 尝试从环境变量或系统获取
	// 这里返回一个合理的默认值
	return DisplayInfo{
		Width:   1920,
		Height:  1080,
		Scale:   1.0,
		Primary: true,
	}
}

// moveCursor 移动光标
func moveCursor(x, y int) error {
	switch runtime.GOOS {
	case "darwin":
		cmd := exec.Command("cliclick", "c:"+fmt.Sprintf("%d,%d", x, y))
		return cmd.Run()
	case "linux":
		if _, err := exec.LookPath("xdotool"); err == nil {
			cmd := exec.Command("xdotool", "mousemove", fmt.Sprintf("%d", x), fmt.Sprintf("%d", y))
			return cmd.Run()
		} else if _, err := exec.LookPath("ydotool"); err == nil {
			cmd := exec.Command("ydotool", "mousemove", fmt.Sprintf("%d:%d", x, y))
			return cmd.Run()
		} else {
			return fmt.Errorf("no mouse tool found (xdotool or ydotool required)")
		}
	case "windows":
		psScript := fmt.Sprintf(`
Add-Type -TypeDefinition '
using System;
using System.Runtime.InteropServices;
public class Mouse {
  [DllImport("user32.dll")]
  public static extern bool SetCursorPos(int x, int y);
}
'
[Mouse]::SetCursorPos(%d, %d)
`, x, y)
		cmd := exec.Command("powershell", "-Command", psScript)
		return cmd.Run()
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// clickMouse 点击鼠标
func clickMouse(button string) (string, error) {
	var buttonCode int
	switch strings.ToLower(button) {
	case "left":
		buttonCode = 1
	case "right":
		buttonCode = 3
	case "middle":
		buttonCode = 2
	default:
		return "", fmt.Errorf("unknown button: %s", button)
	}

	switch runtime.GOOS {
	case "darwin":
		cmd := exec.Command("cliclick", "c:"+fmt.Sprintf("%d", buttonCode))
		if err := cmd.Run(); err != nil {
			return "", err
		}
		return fmt.Sprintf("clicked: %s", button), nil
	case "linux":
		if _, err := exec.LookPath("xdotool"); err == nil {
			cmd := exec.Command("xdotool", "click", fmt.Sprintf("%d", buttonCode))
			if err := cmd.Run(); err != nil {
				return "", err
			}
		} else if _, err := exec.LookPath("ydotool"); err == nil {
			cmd := exec.Command("ydotool", "click", fmt.Sprintf("0x%d", buttonCode))
			if err := cmd.Run(); err != nil {
				return "", err
			}
		} else {
			return "", fmt.Errorf("no mouse tool found")
		}
		return fmt.Sprintf("clicked: %s", button), nil
	case "windows":
		// Windows 使用 PowerShell 调用 user32.dll 的 mouse_event 函数
		var mouseEventFlags string
		switch strings.ToLower(button) {
		case "left":
			mouseEventFlags = "0x02, 0, 0, 0, 0; 0x04, 0, 0, 0, 0" // MOUSEEVENTF_LEFTDOWN, MOUSEEVENTF_LEFTUP
		case "right":
			mouseEventFlags = "0x08, 0, 0, 0, 0; 0x10, 0, 0, 0, 0" // MOUSEEVENTF_RIGHTDOWN, MOUSEEVENTF_RIGHTUP
		case "middle":
			mouseEventFlags = "0x20, 0, 0, 0, 0; 0x40, 0, 0, 0, 0" // MOUSEEVENTF_MIDDLEDOWN, MOUSEEVENTF_MIDDLEUP
		default:
			return "", fmt.Errorf("unknown button: %s", button)
		}

		psScript := fmt.Sprintf(`
Add-Type -TypeDefinition '
using System;
using System.Runtime.InteropServices;
public class Win32 {
    [DllImport("user32.dll")]
    public static extern void mouse_event(uint dwFlags, uint dx, uint dy, uint cButtons, uint dwExtraInfo);
}
'
$events = @(%s)
foreach ($event in $events) {
    [Win32]::mouse_event($event[0], $event[1], $event[2], $event[3], $event[4])
    Start-Sleep -Milliseconds 10
}
`, mouseEventFlags)
		cmd := exec.Command("powershell", "-Command", psScript)
		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("Windows mouse click failed: %w", err)
		}
		return fmt.Sprintf("clicked: %s", button), nil
	default:
		return "", fmt.Errorf("unsupported platform")
	}
}

// escapeAppleScript 转义 AppleScript 字符串
func escapeAppleScript(s string) string {
	// 简单转义
	return strings.ReplaceAll(strings.ReplaceAll(s, "\\", "\\\n"), "\"", "\\\"")
}

// escapePowerShell 转义 PowerShell 字符串
func escapePowerShell(s string) string {
	// 简单转义
	return strings.ReplaceAll(strings.ReplaceAll(s, "`", "``"), "\"", "`\"")
}

// keyCodeForMac 获取 Mac 按键代码
func keyCodeForMac(key string) string {
	keyCodes := map[string]string{
		"return": "36",
		"tab":    "48",
		"space":  "49",
		"escape": "53",
		"enter":  "36",
		"cmd":    "55",
		"ctrl":   "59",
		"shift":  "56",
		"option": "58",
	}

	if code, ok := keyCodes[strings.ToLower(key)]; ok {
		return code
	}
	return "49" // 默认空格
}

// mapKeyToSendKeys 映射按键到 SendKeys 格式
func mapKeyToSendKeys(key string) string {
	sendKeysMap := map[string]string{
		"enter":   "~",
		"return":  "~",
		"tab":     "{TAB}",
		"escape":  "{ESC}",
		"ctrl":    "^",
		"shift":   "+",
		"alt":     "%",
		"delete":  "{DEL}",
		"backspace": "{BS}",
	}

	if mapped, ok := sendKeysMap[strings.ToLower(key)]; ok {
		return mapped
	}
	return key
}

// mapScrollDirection 映射滚动方向
func mapScrollDirection(y int) int {
	if y > 0 {
		return 5 // 向下滚动
	}
	return 4 // 向上滚动
}
