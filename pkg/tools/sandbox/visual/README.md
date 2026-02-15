# Visual Module

Visual Module 是 AIO Sandbox 的可视化控制模块，提供屏幕捕获、远程控制、OCR 识别和视频录制功能。

## 功能特性

- 🖥️ **屏幕捕获**: 全屏或区域截图，支持 PNG 格式
- 🖱️ **远程控制**: 鼠标移动、点击和键盘输入
- 📝 **OCR 识别**: 基于 Tesseract 的文字识别
- 🎥 **视频录制**: 屏幕录制和帧捕获
- 🔒 **权限控制**: 可配置的远程控制权限
- 📊 **统计信息**: 实时操作统计
- 🌐 **跨平台**: 支持 Windows、macOS 和 Linux

## 快速开始

### 安装依赖

#### Windows
- PowerShell（系统自带）
- .NET Framework（系统自带）

#### macOS
```bash
# 安装 cliclick（用于远程控制）
brew install cliclick

# 安装 tesseract（可选，用于 OCR）
brew install tesseract
```

#### Linux
```bash
# 安装屏幕捕获工具
sudo apt install scrot
# 或
sudo apt install imagemagick

# 安装远程控制工具
sudo apt install xdotool

# 安装 tesseract（可选，用于 OCR）
sudo apt install tesseract-ocr
```

### 基本使用

```go
package main

import (
    "context"
    "fmt"
    "github.com/your-org/agent-framework/agent/aiosandbox/visual"
)

func main() {
    // 创建 Visual Module
    module, err := visual.NewVisualModule(visual.VisualConfig{
        Enable:           true,
        AllowControl:     true,
        OCREnabled:       true,
        RecordingEnabled: true,
    })
    if err != nil {
        panic(err)
    }
    defer module.Close()

    // 获取 MCP 工具
    ctx := context.Background()
    tools, err := module.GetTools(ctx)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Available tools: %d\n", len(tools))
}
```

## 配置选项

```go
type VisualConfig struct {
    Enable          bool   // 是否启用模块（默认: false）
    Port            int    // 服务端口（默认: 8080）
    Host            string // 服务主机（默认: localhost）
    AllowControl    bool   // 是否允许远程控制（默认: false）
    OCREnabled      bool   // 是否启用 OCR（默认: false）
    RecordingEnabled bool  // 是否启用录制（默认: false）
}
```

## MCP 工具

Visual Module 提供 9 个 MCP 工具：

### 1. visual_capture_screen

捕获整个屏幕。

**参数**:
- `quality` (integer, 可选): 图像质量 (1-100, 默认: 90)

**返回**:
```json
{
  "success": true,
  "image_data": "base64_encoded_png",
  "width": 1920,
  "height": 1080,
  "format": "png",
  "message": "Screen captured successfully"
}
```

### 2. visual_capture_region

捕获屏幕的指定区域。

**参数**:
- `x` (integer, 必需): 左上角 X 坐标
- `y` (integer, 必需): 左上角 Y 坐标
- `width` (integer, 必需): 区域宽度
- `height` (integer, 必需): 区域高度

**返回**:
```json
{
  "success": true,
  "image_data": "base64_encoded_png",
  "x": 0,
  "y": 0,
  "width": 100,
  "height": 100,
  "format": "png",
  "message": "Region captured successfully"
}
```

### 3. visual_move_mouse

移动鼠标到指定坐标。

**参数**:
- `x` (integer, 必需): X 坐标
- `y` (integer, 必需): Y 坐标

**返回**:
```json
{
  "success": true,
  "x": 100,
  "y": 100,
  "message": "Mouse moved successfully"
}
```

### 4. visual_click

执行鼠标点击。

**参数**:
- `button` (string, 可选): 鼠标按钮 (left/right/middle, 默认: left)
- `x` (integer, 可选): X 坐标（如果指定，先移动鼠标）
- `y` (integer, 可选): Y 坐标（如果指定，先移动鼠标）

**返回**:
```json
{
  "success": true,
  "button": "left",
  "message": "Click performed successfully"
}
```

### 5. visual_type

输入文本。

**参数**:
- `text` (string, 必需): 要输入的文本

**返回**:
```json
{
  "success": true,
  "text": "Hello World",
  "length": 11,
  "message": "Text typed successfully"
}
```

### 6. visual_ocr

对图像执行 OCR 识别。

**参数**:
- `image_data` (string, 必需): Base64 编码的图像数据
- `language` (string, 可选): OCR 语言 (默认: eng)

**返回**:
```json
{
  "success": true,
  "text": "Recognized text",
  "language": "eng",
  "message": "OCR completed successfully"
}
```

### 7. visual_start_recording

开始录制屏幕。

**参数**:
- `fps` (integer, 可选): 帧率 (默认: 10)

**返回**:
```json
{
  "success": true,
  "fps": 10,
  "start_time": "2026-01-30T10:00:00Z",
  "message": "Recording started successfully"
}
```

### 8. visual_stop_recording

停止录制屏幕。

**参数**: 无

**返回**:
```json
{
  "success": true,
  "frame_count": 100,
  "duration_ms": 10000,
  "message": "Recording stopped successfully"
}
```

### 9. visual_get_stats

获取统计信息。

**参数**: 无

**返回**:
```json
{
  "total_captures": 10,
  "total_controls": 5,
  "total_ocr": 2,
  "total_recordings": 1
}
```

## 使用示例

### 屏幕捕获

```go
// 捕获全屏
result, err := module.captureScreen(90)
if err != nil {
    log.Fatal(err)
}

imageData := result["image_data"].(string)
fmt.Printf("Captured image: %d bytes\n", len(imageData))
```

### 远程控制

```go
// 移动鼠标
result, err := module.moveMouse(100, 100)
if err != nil {
    log.Fatal(err)
}

// 点击
result, err = module.click("left", nil, nil)
if err != nil {
    log.Fatal(err)
}

// 输入文本
result, err = module.typeText("Hello World")
if err != nil {
    log.Fatal(err)
}
```

### OCR 识别

```go
// 先捕获屏幕
captureResult, _ := module.captureScreen(90)
imageData := captureResult["image_data"].(string)

// 执行 OCR
ocrResult, err := module.performOCR(imageData, "eng")
if err != nil {
    log.Fatal(err)
}

text := ocrResult["text"].(string)
fmt.Printf("Recognized text: %s\n", text)
```

### 视频录制

```go
// 开始录制（10 FPS）
startResult, err := module.startRecording(10)
if err != nil {
    log.Fatal(err)
}

// 等待一段时间
time.Sleep(5 * time.Second)

// 停止录制
stopResult, err := module.stopRecording()
if err != nil {
    log.Fatal(err)
}

frameCount := stopResult["frame_count"].(int)
fmt.Printf("Recorded %d frames\n", frameCount)
```

### 获取统计信息

```go
stats := module.GetStats()
fmt.Printf("Total captures: %d\n", stats["total_captures"])
fmt.Printf("Total controls: %d\n", stats["total_controls"])
fmt.Printf("Total OCR: %d\n", stats["total_ocr"])
fmt.Printf("Total recordings: %d\n", stats["total_recordings"])
```

## 权限控制

Visual Module 支持权限控制，可以通过配置禁用远程控制功能：

```go
module, err := visual.NewVisualModule(visual.VisualConfig{
    Enable:       true,
    AllowControl: false, // 禁用远程控制
    OCREnabled:   true,
})

// 尝试移动鼠标会失败
result, _ := module.moveMouse(100, 100)
// result["success"] == false
// result["error"] == "Remote control is disabled"
```

## 并发安全

Visual Module 的所有操作都是并发安全的，可以在多个 goroutine 中同时调用：

```go
// 并发捕获屏幕
for i := 0; i < 10; i++ {
    go func() {
        module.captureScreen(90)
    }()
}
```

## 错误处理

所有操作都返回详细的错误信息：

```go
result, err := module.moveMouse(100, 100)
if err != nil {
    log.Printf("Error: %v\n", err)
    return
}

if !result["success"].(bool) {
    errorMsg := result["error"].(string)
    log.Printf("Operation failed: %s\n", errorMsg)
}
```

## 平台差异

### Windows
- 使用 PowerShell 和 .NET Framework
- 屏幕捕获速度较快
- 远程控制功能完整

### macOS
- 需要安装 cliclick
- 屏幕捕获使用系统命令
- 可能需要授予辅助功能权限

### Linux
- 需要安装 scrot/ImageMagick 和 xdotool
- 需要 X11 环境
- 可能需要配置 DISPLAY 环境变量

## 性能考虑

- 屏幕捕获: ~0.5-1 秒/次（取决于分辨率）
- 远程控制: ~0.5 秒/操作
- OCR 识别: ~1-3 秒/图像（取决于图像大小）
- 录制: 内存占用随帧数增加

## 故障排查

### 屏幕捕获失败

**Windows**:
- 检查 PowerShell 是否可用
- 检查 .NET Framework 是否安装

**macOS**:
- 检查是否授予了屏幕录制权限
- 系统偏好设置 > 安全性与隐私 > 屏幕录制

**Linux**:
- 检查 scrot 或 ImageMagick 是否安装
- 检查 DISPLAY 环境变量

### 远程控制失败

**Windows**:
- 检查 PowerShell 执行策略

**macOS**:
- 安装 cliclick: `brew install cliclick`
- 授予辅助功能权限

**Linux**:
- 安装 xdotool: `sudo apt install xdotool`
- 检查 X11 是否运行

### OCR 失败

所有平台:
- 安装 tesseract
- 检查语言包是否安装
- 验证图像格式正确

## 许可证

AGPL-3.0-or-later

## 贡献

欢迎提交 Issue 和 Pull Request！

## 相关文档

- [IMPLEMENTATION_SUMMARY.md](./IMPLEMENTATION_SUMMARY.md) - 实现总结
- [visual_test.go](./visual_test.go) - 测试用例
- [AIO Sandbox 文档](../README.md) - 主文档
