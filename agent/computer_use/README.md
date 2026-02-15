# Computer Use - 计算机控制代理

## 概述

`computer_use` 包提供了一个支持视觉 GUI 交互的计算机控制智能体。它使 AI 代理能够通过自然语言命令执行屏幕操作，包括点击、输入、截图等。

## 核心功能

### 支持的操作

| 操作类型 | 说明 | 参数 |
|---------|------|------|
| `click` | 单击 | `coordinates: {x, y}` |
| `double_click` | 双击 | `coordinates: {x, y}` |
| `right_click` | 右键点击 | `coordinates: {x, y}` |
| `type` | 输入文本 | `text: string` |
| `key_press` | 按键 | `key: string` |
| `scroll` | 滚动 | `coordinates: {x, y}` (x=水平, y=垂直) |
| `wait` | 等待 | `duration: int` (毫秒) |
| `screenshot` | 截图 | - |

### 跨平台支持

| 平台 | 截图 | 鼠标 | 键盘 | 滚动 |
|------|------|------|------|------|
| **macOS** | screencapture | cliclick | osascript | osascript |
| **Linux** | gnome-screenshot/scrot | xdotool/ydotool | xdotool/ydotool | xdotool |
| **Windows** | PowerShell | PowerShell (user32.dll) | PowerShell (SendKeys) | PowerShell (mouse_event) |

---

## 快速开始

### 创建代理

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/cloudwego/eino/components/model"
    "your-org/agentframework/agent/computer_use"
)

func main() {
    // 创建模型
    model, err := model.NewChatModel(ctx, &model.ChatModelConfig{
        Model: "gpt-4-vision-preview", // 需要支持视觉的模型
    })
    if err != nil {
        log.Fatal(err)
    }

    // 创建 Computer Use 代理
    agent, err := computer_use.NewComputerUseAgent(computer_use.ComputerUseConfig{
        Model:         model,
        ScreenshotDir: "/tmp/screenshots",
        MaxHistory:    100,
        DisplayWidth:  1920,
        DisplayHeight: 1080,
    })
    if err != nil {
        log.Fatal(err)
    }

    // 执行命令
    action, err := agent.ExecuteCommand(ctx, "点击屏幕中央")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("操作结果: %s\n", action.Result)
}
```

---

## API 文档

### ComputerUseAgent

#### 创建代理

```go
func NewComputerUseAgent(config ComputerUseConfig) (*ComputerUseAgent, error)
```

**配置参数**：
- `Model` - LLM 模型（需要支持视觉）
- `ScreenshotDir` - 截图保存目录（默认：系统临时目录）
- `MaxHistory` - 最大历史记录数（默认：100）
- `DisplayWidth` - 显示宽度（默认：1920）
- `DisplayHeight` - 显示高度（默认：1080）

#### 执行命令

```go
func (a *ComputerUseAgent) ExecuteCommand(ctx context.Context, command string) (*ComputerAction, error)
```

执行自然语言命令并返回操作结果。

**示例**：
```go
// 点击操作
action, _ := agent.ExecuteCommand(ctx, "点击坐标 (100, 200)")

// 输入文本
action, _ := agent.ExecuteCommand(ctx, "输入: Hello World")

// 按键
action, _ := agent.ExecuteCommand(ctx, "按下回车键")

// 等待
action, _ := agent.ExecuteCommand(ctx, "等待 2 秒")
```

#### 获取信息

```go
// 获取显示信息
display := agent.GetDisplayInfo()

// 获取光标位置
cursor := agent.GetCursorInfo()

// 获取操作历史
history := agent.GetHistory()
```

---

## ComputerAction 结构

```go
type ComputerAction struct {
    Type        ActionType             `json:"type"`        // 操作类型
    Timestamp   time.Time              `json:"timestamp"`   // 时间戳
    Coordinates *Coordinates           `json:"coordinates"` // 坐标（可选）
    Text        string                 `json:"text"`        // 文本内容（可选）
    Key         string                 `json:"key"`         // 按键（可选）
    Duration    int                    `json:"duration"`    // 持续时间（可选）
    Button      string                 `json:"button"`      // 鼠标按钮（可选）
    Metadata    map[string]interface{} `json:"metadata"`    // 元数据
    Result      string                 `json:"result"`      // 执行结果
    Error       string                 `json:"error"`       // 错误信息
}
```

---

## 坐标系统

- **原点 (0,0)**：屏幕左上角
- **X 轴**：向右增加
- **Y 轴**：向下增加

```
(0,0) ───────────────► X
  │
  │    ┌─────────┐
  │    │  Screen │
  │    │         │
  │    └─────────┘
  │
  ▼
  Y
```

---

## 工作流程

```
┌─────────────┐
│ 自然语言命令 │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│   截图获取   │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ LLM 解析命令 │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  执行操作    │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  记录历史    │
└─────────────┘
```

---

## 平台依赖

### macOS
```bash
# 安装 cliclick（鼠标/键盘控制）
brew install cliclick

# 截图自带（screencapture）
```

### Linux
```bash
# Ubuntu/Debian
sudo apt-get install xdotool gnome-screenshot

# 或使用 ydotool（Wayland 兼容）
sudo apt-get install ydotool
```

### Windows
- PowerShell 自带
- 部分功能需要管理员权限

---

## 安全注意事项

1. **权限控制**：代理拥有完整的系统控制权限
2. **操作验证**：建议实现操作确认机制
3. **屏幕隐私**：截图可能包含敏感信息
4. **资源限制**：设置合理的操作频率限制

---

## 高级用法

### 自定义命令解析

```go
// 使用简单解析器（不使用 LLM）
action := agent.simpleParseCommand("点击屏幕")

// 或实现自定义解析逻辑
```

### 操作历史

```go
history := agent.GetHistory()
for _, action := range history {
    fmt.Printf("%s: %s -> %s\n",
        action.Timestamp,
        action.Type,
        action.Result)
}
```

### 显示信息

```go
display := agent.GetDisplayInfo()
fmt.Printf("分辨率: %dx%d, 缩放: %.2f\n",
    display.Width,
    display.Height,
    display.Scale)
```

---

## 故障排除

### macOS 权限问题
```bash
# 授予终端辅助功能权限
系统偏好设置 > 安全性与隐私 > 辅助功能
```

### Linux 工具缺失
```bash
# 检查工具是否安装
which xdotool
which gnome-screenshot

# 安装缺失工具
sudo apt-get install xdotool gnome-screenshot
```

### Windows 权限问题
```powershell
# 以管理员身份运行 PowerShell
# Set-ExecutionPolicy RemoteSigned
```

---

## 示例项目

- 基础演示：`cmd/demo/main.go`
- 完整示例：查看 `examples/` 目录

---

## 相关文档

- [AIO 沙箱](../aiosandbox/) - 沙箱系统说明
- [Browser 工具](../aiosandbox/browser/) - 浏览器自动化
- [架构文档](../../doc/ARCHITECTURE_UNIFIED.md) - 系统架构

---

**最后更新**: 2025-02-03
