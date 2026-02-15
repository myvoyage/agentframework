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

package visual

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// VisualModule 可视化接管模块
type VisualModule struct {
	config   VisualConfig
	capture  *ScreenCapture
	control  *RemoteController
	ocr      *OCREngine
	recorder *VideoRecorder
	stats    *VisualStats
	mu       sync.RWMutex
}

// VisualConfig 可视化配置
type VisualConfig struct {
	Enable           bool   `json:"enable"`
	Port             int    `json:"port"`
	Host             string `json:"host"`
	AllowControl     bool   `json:"allow_control"`     // 是否允许远程控制
	OCREnabled       bool   `json:"ocr_enabled"`       // 是否启用 OCR
	RecordingEnabled bool   `json:"recording_enabled"` // 是否启用录制
}

// ScreenCapture 屏幕捕获组件
type ScreenCapture struct {
	quality int
	mu      sync.Mutex
}

// RemoteController 远程控制组件
type RemoteController struct {
	enabled bool
	mu      sync.Mutex
}

// OCREngine OCR 引擎组件
type OCREngine struct {
	enabled bool
	mu      sync.Mutex
}

// VideoRecorder 视频录制组件
type VideoRecorder struct {
	recording bool
	frames    [][]byte
	startTime time.Time
	mu        sync.Mutex
}

// VisualStats 可视化统计
type VisualStats struct {
	TotalCaptures   int64
	TotalControls   int64
	TotalOCR        int64
	TotalRecordings int64
	mu              sync.RWMutex
}

// NewVisualModule 创建可视化模块实例
func NewVisualModule(config VisualConfig) (*VisualModule, error) {
	// 设置默认值
	if config.Port == 0 {
		config.Port = 8080
	}
	if config.Host == "" {
		config.Host = "localhost"
	}

	// 创建屏幕捕获组件
	capture := &ScreenCapture{
		quality: 90, // 默认质量
	}

	// 创建远程控制组件
	control := &RemoteController{
		enabled: config.AllowControl,
	}

	// 创建 OCR 引擎
	ocr := &OCREngine{
		enabled: config.OCREnabled,
	}

	// 创建视频录制器
	recorder := &VideoRecorder{
		recording: false,
		frames:    make([][]byte, 0),
	}

	// 创建统计信息
	stats := &VisualStats{}

	return &VisualModule{
		config:   config,
		capture:  capture,
		control:  control,
		ocr:      ocr,
		recorder: recorder,
		stats:    stats,
	}, nil
}

// GetTools 返回可视化模块的 MCP 工具列表
func (m *VisualModule) GetTools(ctx context.Context) ([]tool.BaseTool, error) {
	// 如果可视化模块未启用，返回空列表
	if !m.config.Enable {
		return []tool.BaseTool{}, nil
	}

	tools := []tool.BaseTool{
		// 屏幕捕获工具
		&visualCaptureScreenTool{module: m},
		// 区域捕获工具
		&visualCaptureRegionTool{module: m},
		// 鼠标移动工具
		&visualMoveMouseTool{module: m},
		// 点击工具
		&visualClickTool{module: m},
		// 输入文本工具
		&visualTypeTool{module: m},
		// OCR 识别工具
		&visualOCRTool{module: m},
		// 开始录制工具
		&visualStartRecordingTool{module: m},
		// 停止录制工具
		&visualStopRecordingTool{module: m},
		// 获取统计信息工具
		&visualGetStatsTool{module: m},
	}

	return tools, nil
}

// ============================================================================
// MCP Tools Implementation
// ============================================================================

// 屏幕捕获工具
type visualCaptureScreenTool struct {
	module *VisualModule
}

func (t *visualCaptureScreenTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "visual_capture_screen",
		Desc: "Capture the entire screen and return as base64 encoded PNG",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"quality": {
				Type: "integer",
				Desc: "Image quality (1-100, default: 90)",
			},
		}),
	}, nil
}

func (t *visualCaptureScreenTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var args struct {
		Quality int `json:"quality"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	result, err := t.module.captureScreen(args.Quality)
	if err != nil {
		return "", err
	}

	output, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// 区域捕获工具
type visualCaptureRegionTool struct {
	module *VisualModule
}

func (t *visualCaptureRegionTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "visual_capture_region",
		Desc: "Capture a specific region of the screen",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"x": {
				Type:     "integer",
				Desc:     "X coordinate of top-left corner",
				Required: true,
			},
			"y": {
				Type:     "integer",
				Desc:     "Y coordinate of top-left corner",
				Required: true,
			},
			"width": {
				Type:     "integer",
				Desc:     "Width of the region",
				Required: true,
			},
			"height": {
				Type:     "integer",
				Desc:     "Height of the region",
				Required: true,
			},
		}),
	}, nil
}

func (t *visualCaptureRegionTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var args struct {
		X      int `json:"x"`
		Y      int `json:"y"`
		Width  int `json:"width"`
		Height int `json:"height"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	result, err := t.module.captureRegion(args.X, args.Y, args.Width, args.Height)
	if err != nil {
		return "", err
	}

	output, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// 鼠标移动工具
type visualMoveMouseTool struct {
	module *VisualModule
}

func (t *visualMoveMouseTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "visual_move_mouse",
		Desc: "Move the mouse cursor to specified coordinates",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"x": {
				Type:     "integer",
				Desc:     "X coordinate",
				Required: true,
			},
			"y": {
				Type:     "integer",
				Desc:     "Y coordinate",
				Required: true,
			},
		}),
	}, nil
}

func (t *visualMoveMouseTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var args struct {
		X int `json:"x"`
		Y int `json:"y"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	result, err := t.module.moveMouse(args.X, args.Y)
	if err != nil {
		return "", err
	}

	output, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// 点击工具
type visualClickTool struct {
	module *VisualModule
}

func (t *visualClickTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "visual_click",
		Desc: "Perform a mouse click at current or specified position",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"button": {
				Type: "string",
				Desc: "Mouse button: left, right, middle (default: left)",
			},
			"x": {
				Type: "integer",
				Desc: "X coordinate (optional, uses current position if not specified)",
			},
			"y": {
				Type: "integer",
				Desc: "Y coordinate (optional, uses current position if not specified)",
			},
		}),
	}, nil
}

func (t *visualClickTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var args struct {
		Button string `json:"button"`
		X      *int   `json:"x"`
		Y      *int   `json:"y"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	result, err := t.module.click(args.Button, args.X, args.Y)
	if err != nil {
		return "", err
	}

	output, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// 输入文本工具
type visualTypeTool struct {
	module *VisualModule
}

func (t *visualTypeTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "visual_type",
		Desc: "Type text using keyboard simulation",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"text": {
				Type:     "string",
				Desc:     "Text to type",
				Required: true,
			},
		}),
	}, nil
}

func (t *visualTypeTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var args struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	result, err := t.module.typeText(args.Text)
	if err != nil {
		return "", err
	}

	output, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// OCR 识别工具
type visualOCRTool struct {
	module *VisualModule
}

func (t *visualOCRTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "visual_ocr",
		Desc: "Perform OCR (Optical Character Recognition) on an image",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"image_data": {
				Type:     "string",
				Desc:     "Base64 encoded image data",
				Required: true,
			},
			"language": {
				Type: "string",
				Desc: "OCR language (default: eng)",
			},
		}),
	}, nil
}

func (t *visualOCRTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var args struct {
		ImageData string `json:"image_data"`
		Language  string `json:"language"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	result, err := t.module.performOCR(args.ImageData, args.Language)
	if err != nil {
		return "", err
	}

	output, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// 开始录制工具
type visualStartRecordingTool struct {
	module *VisualModule
}

func (t *visualStartRecordingTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "visual_start_recording",
		Desc: "Start recording screen activity",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"fps": {
				Type: "integer",
				Desc: "Frames per second (default: 10)",
			},
		}),
	}, nil
}

func (t *visualStartRecordingTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var args struct {
		FPS int `json:"fps"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	result, err := t.module.startRecording(args.FPS)
	if err != nil {
		return "", err
	}

	output, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// 停止录制工具
type visualStopRecordingTool struct {
	module *VisualModule
}

func (t *visualStopRecordingTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name:        "visual_stop_recording",
		Desc:        "Stop recording and return the recorded frames count",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{}),
	}, nil
}

func (t *visualStopRecordingTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	result, err := t.module.stopRecording()
	if err != nil {
		return "", err
	}

	output, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// 获取统计信息工具
type visualGetStatsTool struct {
	module *VisualModule
}

func (t *visualGetStatsTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name:        "visual_get_stats",
		Desc:        "Get visual module statistics",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{}),
	}, nil
}

func (t *visualGetStatsTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	result := t.module.GetStats()

	output, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// Close 关闭可视化模块，释放资源
func (m *VisualModule) Close() error {
	// 停止录制（如果正在录制）
	if m.recorder.recording {
		m.stopRecording()
	}
	return nil
}

// GetStats 获取统计信息
func (m *VisualModule) GetStats() map[string]any {
	m.stats.mu.RLock()
	defer m.stats.mu.RUnlock()

	return map[string]any{
		"total_captures":   m.stats.TotalCaptures,
		"total_controls":   m.stats.TotalControls,
		"total_ocr":        m.stats.TotalOCR,
		"total_recordings": m.stats.TotalRecordings,
	}
}

// ============================================================================
// Core Implementation Functions
// ============================================================================

// captureScreen 捕获整个屏幕
func (m *VisualModule) captureScreen(quality int) (map[string]any, error) {
	m.capture.mu.Lock()
	defer m.capture.mu.Unlock()

	// 更新统计
	m.stats.mu.Lock()
	m.stats.TotalCaptures++
	m.stats.mu.Unlock()

	// 设置默认质量
	if quality <= 0 || quality > 100 {
		quality = 90
	}

	// 使用平台特定的截图命令
	imageData, width, height, err := m.captureScreenNative()
	if err != nil {
		return map[string]any{
			"success": false,
			"error":   err.Error(),
		}, nil
	}

	return map[string]any{
		"success":    true,
		"image_data": imageData,
		"width":      width,
		"height":     height,
		"format":     "png",
		"message":    "Screen captured successfully",
	}, nil
}

// captureRegion 捕获屏幕区域
func (m *VisualModule) captureRegion(x, y, width, height int) (map[string]any, error) {
	m.capture.mu.Lock()
	defer m.capture.mu.Unlock()

	// 更新统计
	m.stats.mu.Lock()
	m.stats.TotalCaptures++
	m.stats.mu.Unlock()

	// 使用平台特定的区域截图命令
	imageData, err := m.captureRegionNative(x, y, width, height)
	if err != nil {
		return map[string]any{
			"success": false,
			"error":   err.Error(),
		}, nil
	}

	return map[string]any{
		"success":    true,
		"image_data": imageData,
		"x":          x,
		"y":          y,
		"width":      width,
		"height":     height,
		"format":     "png",
		"message":    "Region captured successfully",
	}, nil
}

// moveMouse 移动鼠标
func (m *VisualModule) moveMouse(x, y int) (map[string]any, error) {
	m.control.mu.Lock()
	defer m.control.mu.Unlock()

	if !m.control.enabled {
		return map[string]any{
			"success": false,
			"error":   "Remote control is disabled",
		}, nil
	}

	// 更新统计
	m.stats.mu.Lock()
	m.stats.TotalControls++
	m.stats.mu.Unlock()

	// 使用平台特定的鼠标移动命令
	err := m.moveMouseNative(x, y)
	if err != nil {
		return map[string]any{
			"success": false,
			"error":   err.Error(),
		}, nil
	}

	return map[string]any{
		"success": true,
		"x":       x,
		"y":       y,
		"message": "Mouse moved successfully",
	}, nil
}

// click 执行鼠标点击
func (m *VisualModule) click(button string, x, y *int) (map[string]any, error) {
	m.control.mu.Lock()
	defer m.control.mu.Unlock()

	if !m.control.enabled {
		return map[string]any{
			"success": false,
			"error":   "Remote control is disabled",
		}, nil
	}

	// 更新统计
	m.stats.mu.Lock()
	m.stats.TotalControls++
	m.stats.mu.Unlock()

	// 设置默认按钮
	if button == "" {
		button = "left"
	}

	// 如果指定了坐标，先移动鼠标
	if x != nil && y != nil {
		if err := m.moveMouseNative(*x, *y); err != nil {
			return map[string]any{
				"success": false,
				"error":   fmt.Sprintf("failed to move mouse: %v", err),
			}, nil
		}
	}

	// 使用平台特定的点击命令
	err := m.clickNative(button)
	if err != nil {
		return map[string]any{
			"success": false,
			"error":   err.Error(),
		}, nil
	}

	result := map[string]any{
		"success": true,
		"button":  button,
		"message": "Click performed successfully",
	}

	if x != nil && y != nil {
		result["x"] = *x
		result["y"] = *y
	}

	return result, nil
}

// typeText 输入文本
func (m *VisualModule) typeText(text string) (map[string]any, error) {
	m.control.mu.Lock()
	defer m.control.mu.Unlock()

	if !m.control.enabled {
		return map[string]any{
			"success": false,
			"error":   "Remote control is disabled",
		}, nil
	}

	// 更新统计
	m.stats.mu.Lock()
	m.stats.TotalControls++
	m.stats.mu.Unlock()

	// 使用平台特定的文本输入命令
	err := m.typeTextNative(text)
	if err != nil {
		return map[string]any{
			"success": false,
			"error":   err.Error(),
		}, nil
	}

	return map[string]any{
		"success": true,
		"text":    text,
		"length":  len(text),
		"message": "Text typed successfully",
	}, nil
}

// performOCR 执行 OCR 识别
func (m *VisualModule) performOCR(imageData, language string) (map[string]any, error) {
	m.ocr.mu.Lock()
	defer m.ocr.mu.Unlock()

	if !m.ocr.enabled {
		return map[string]any{
			"success": false,
			"error":   "OCR is disabled",
		}, nil
	}

	// 更新统计
	m.stats.mu.Lock()
	m.stats.TotalOCR++
	m.stats.mu.Unlock()

	// 设置默认语言
	if language == "" {
		language = "eng"
	}

	// 解码 base64 图像
	imgBytes, err := base64.StdEncoding.DecodeString(imageData)
	if err != nil {
		return map[string]any{
			"success": false,
			"error":   fmt.Sprintf("failed to decode image: %v", err),
		}, nil
	}

	// 使用 tesseract 进行 OCR
	text, err := m.performOCRNative(imgBytes, language)
	if err != nil {
		return map[string]any{
			"success": false,
			"error":   err.Error(),
		}, nil
	}

	return map[string]any{
		"success":  true,
		"text":     text,
		"language": language,
		"message":  "OCR completed successfully",
	}, nil
}

// startRecording 开始录制
func (m *VisualModule) startRecording(fps int) (map[string]any, error) {
	m.recorder.mu.Lock()
	defer m.recorder.mu.Unlock()

	if m.recorder.recording {
		return map[string]any{
			"success": false,
			"error":   "Recording already in progress",
		}, nil
	}

	// 设置默认 FPS
	if fps <= 0 {
		fps = 10
	}

	// 更新统计
	m.stats.mu.Lock()
	m.stats.TotalRecordings++
	m.stats.mu.Unlock()

	// 初始化录制
	m.recorder.recording = true
	m.recorder.frames = make([][]byte, 0)
	m.recorder.startTime = time.Now()

	// 启动录制协程
	go m.recordFrames(fps)

	return map[string]any{
		"success":    true,
		"fps":        fps,
		"start_time": m.recorder.startTime,
		"message":    "Recording started successfully",
	}, nil
}

// stopRecording 停止录制
func (m *VisualModule) stopRecording() (map[string]any, error) {
	m.recorder.mu.Lock()
	defer m.recorder.mu.Unlock()

	if !m.recorder.recording {
		return map[string]any{
			"success": false,
			"error":   "No recording in progress",
		}, nil
	}

	// 停止录制
	m.recorder.recording = false
	duration := time.Since(m.recorder.startTime)
	frameCount := len(m.recorder.frames)

	return map[string]any{
		"success":     true,
		"frame_count": frameCount,
		"duration_ms": duration.Milliseconds(),
		"message":     "Recording stopped successfully",
	}, nil
}

// recordFrames 录制帧（后台协程）
func (m *VisualModule) recordFrames(fps int) {
	interval := time.Second / time.Duration(fps)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		<-ticker.C

		m.recorder.mu.Lock()
		if !m.recorder.recording {
			m.recorder.mu.Unlock()
			return
		}
		m.recorder.mu.Unlock()

		// 捕获帧
		imageData, _, _, err := m.captureScreenNative()
		if err != nil {
			continue
		}

		// 解码 base64 并存储
		imgBytes, err := base64.StdEncoding.DecodeString(imageData)
		if err != nil {
			continue
		}

		m.recorder.mu.Lock()
		m.recorder.frames = append(m.recorder.frames, imgBytes)
		m.recorder.mu.Unlock()
	}
}

// ============================================================================
// Platform-Specific Native Implementations
// ============================================================================

// captureScreenNative 使用平台特定命令捕获屏幕
func (m *VisualModule) captureScreenNative() (string, int, int, error) {
	var cmd *exec.Cmd
	var imgData []byte
	var err error

	switch runtime.GOOS {
	case "windows":
		// Windows: 使用 PowerShell 截图
		cmd = exec.Command("powershell", "-Command",
			"Add-Type -AssemblyName System.Windows.Forms,System.Drawing;"+
				"$screens = [Windows.Forms.Screen]::AllScreens;"+
				"$top = ($screens.Bounds.Top | Measure-Object -Minimum).Minimum;"+
				"$left = ($screens.Bounds.Left | Measure-Object -Minimum).Minimum;"+
				"$width = ($screens.Bounds.Right | Measure-Object -Maximum).Maximum;"+
				"$height = ($screens.Bounds.Bottom | Measure-Object -Maximum).Maximum;"+
				"$bounds = [Drawing.Rectangle]::FromLTRB($left, $top, $width, $height);"+
				"$bmp = New-Object Drawing.Bitmap $bounds.width, $bounds.height;"+
				"$graphics = [Drawing.Graphics]::FromImage($bmp);"+
				"$graphics.CopyFromScreen($bounds.Location, [Drawing.Point]::Empty, $bounds.size);"+
				"$ms = New-Object IO.MemoryStream;"+
				"$bmp.Save($ms, [Drawing.Imaging.ImageFormat]::Png);"+
				"[Convert]::ToBase64String($ms.ToArray())")

		output, err := cmd.Output()
		if err != nil {
			return "", 0, 0, fmt.Errorf("failed to capture screen on Windows: %w", err)
		}
		return string(output), 1920, 1080, nil // 默认分辨率

	case "darwin":
		// macOS: 使用 screencapture
		tmpFile := fmt.Sprintf("/tmp/screenshot_%d.png", time.Now().UnixNano())
		cmd = exec.Command("screencapture", "-x", tmpFile)
		if err := cmd.Run(); err != nil {
			return "", 0, 0, fmt.Errorf("failed to capture screen on macOS: %w", err)
		}

		// 读取文件并转换为 base64
		imgData, err = exec.Command("base64", "-i", tmpFile).Output()
		if err != nil {
			return "", 0, 0, fmt.Errorf("failed to encode image: %w", err)
		}

		// 清理临时文件
		exec.Command("rm", tmpFile).Run()

		// 获取屏幕分辨率（默认值，实际应该解析 system_profiler 输出）
		width, height := 1920, 1080
		return string(imgData), width, height, nil

	case "linux":
		// Linux: 使用 scrot 或 import (ImageMagick)
		tmpFile := fmt.Sprintf("/tmp/screenshot_%d.png", time.Now().UnixNano())

		// 尝试使用 scrot
		cmd = exec.Command("scrot", tmpFile)
		if err := cmd.Run(); err != nil {
			// 如果 scrot 不可用，尝试 import
			cmd = exec.Command("import", "-window", "root", tmpFile)
			if err := cmd.Run(); err != nil {
				return "", 0, 0, fmt.Errorf("failed to capture screen on Linux (tried scrot and import): %w", err)
			}
		}

		// 读取文件并转换为 base64
		imgData, err = exec.Command("base64", "-w", "0", tmpFile).Output()
		if err != nil {
			return "", 0, 0, fmt.Errorf("failed to encode image: %w", err)
		}

		// 清理临时文件
		exec.Command("rm", tmpFile).Run()

		// 获取屏幕分辨率（默认值，实际应该解析 xdpyinfo 输出）
		width, height := 1920, 1080
		return string(imgData), width, height, nil

	default:
		return "", 0, 0, fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// captureRegionNative 使用平台特定命令捕获屏幕区域
func (m *VisualModule) captureRegionNative(x, y, width, height int) (string, error) {
	var cmd *exec.Cmd
	var imgData []byte
	var err error

	switch runtime.GOOS {
	case "windows":
		// Windows: 使用 PowerShell 截图指定区域
		psScript := fmt.Sprintf(
			"Add-Type -AssemblyName System.Windows.Forms,System.Drawing;"+
				"$bounds = New-Object Drawing.Rectangle %d,%d,%d,%d;"+
				"$bmp = New-Object Drawing.Bitmap $bounds.width, $bounds.height;"+
				"$graphics = [Drawing.Graphics]::FromImage($bmp);"+
				"$graphics.CopyFromScreen($bounds.Location, [Drawing.Point]::Empty, $bounds.size);"+
				"$ms = New-Object IO.MemoryStream;"+
				"$bmp.Save($ms, [Drawing.Imaging.ImageFormat]::Png);"+
				"[Convert]::ToBase64String($ms.ToArray())",
			x, y, width, height)

		cmd = exec.Command("powershell", "-Command", psScript)
		output, err := cmd.Output()
		if err != nil {
			return "", fmt.Errorf("failed to capture region on Windows: %w", err)
		}
		return string(output), nil

	case "darwin":
		// macOS: 使用 screencapture 捕获区域
		tmpFile := fmt.Sprintf("/tmp/screenshot_%d.png", time.Now().UnixNano())
		region := fmt.Sprintf("%d,%d,%d,%d", x, y, width, height)
		cmd = exec.Command("screencapture", "-x", "-R", region, tmpFile)
		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("failed to capture region on macOS: %w", err)
		}

		imgData, err = exec.Command("base64", "-i", tmpFile).Output()
		if err != nil {
			return "", fmt.Errorf("failed to encode image: %w", err)
		}

		exec.Command("rm", tmpFile).Run()
		return string(imgData), nil

	case "linux":
		// Linux: 使用 import 捕获区域
		tmpFile := fmt.Sprintf("/tmp/screenshot_%d.png", time.Now().UnixNano())
		geometry := fmt.Sprintf("%dx%d+%d+%d", width, height, x, y)
		cmd = exec.Command("import", "-window", "root", "-crop", geometry, tmpFile)
		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("failed to capture region on Linux: %w", err)
		}

		imgData, err = exec.Command("base64", "-w", "0", tmpFile).Output()
		if err != nil {
			return "", fmt.Errorf("failed to encode image: %w", err)
		}

		exec.Command("rm", tmpFile).Run()
		return string(imgData), nil

	default:
		return "", fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// moveMouseNative 使用平台特定命令移动鼠标
func (m *VisualModule) moveMouseNative(x, y int) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		// Windows: 使用 PowerShell
		psScript := fmt.Sprintf(
			"Add-Type -AssemblyName System.Windows.Forms;"+
				"[System.Windows.Forms.Cursor]::Position = New-Object System.Drawing.Point(%d,%d)",
			x, y)
		cmd = exec.Command("powershell", "-Command", psScript)

	case "darwin":
		// macOS: 使用 cliclick
		cmd = exec.Command("cliclick", fmt.Sprintf("m:%d,%d", x, y))

	case "linux":
		// Linux: 使用 xdotool
		cmd = exec.Command("xdotool", "mousemove", fmt.Sprintf("%d", x), fmt.Sprintf("%d", y))

	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to move mouse: %w", err)
	}

	return nil
}

// clickNative 使用平台特定命令执行鼠标点击
func (m *VisualModule) clickNative(button string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		// Windows: 使用 PowerShell
		mouseButton := "Left"
		if button == "right" {
			mouseButton = "Right"
		} else if button == "middle" {
			mouseButton = "Middle"
		}

		psScript := fmt.Sprintf(
			"Add-Type -AssemblyName System.Windows.Forms;"+
				"$pos = [System.Windows.Forms.Cursor]::Position;"+
				"[System.Windows.Forms.SendKeys]::SendWait('{%s}')",
			mouseButton)
		cmd = exec.Command("powershell", "-Command", psScript)

	case "darwin":
		// macOS: 使用 cliclick
		clickType := "c:."
		if button == "right" {
			clickType = "rc:."
		} else if button == "middle" {
			clickType = "mc:."
		}
		cmd = exec.Command("cliclick", clickType)

	case "linux":
		// Linux: 使用 xdotool
		buttonNum := "1"
		if button == "right" {
			buttonNum = "3"
		} else if button == "middle" {
			buttonNum = "2"
		}
		cmd = exec.Command("xdotool", "click", buttonNum)

	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to click: %w", err)
	}

	return nil
}

// typeTextNative 使用平台特定命令输入文本
func (m *VisualModule) typeTextNative(text string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		// Windows: 使用 PowerShell
		psScript := fmt.Sprintf(
			"Add-Type -AssemblyName System.Windows.Forms;"+
				"[System.Windows.Forms.SendKeys]::SendWait('%s')",
			text)
		cmd = exec.Command("powershell", "-Command", psScript)

	case "darwin":
		// macOS: 使用 cliclick
		cmd = exec.Command("cliclick", fmt.Sprintf("t:%s", text))

	case "linux":
		// Linux: 使用 xdotool
		cmd = exec.Command("xdotool", "type", text)

	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to type text: %w", err)
	}

	return nil
}

// performOCRNative 使用 tesseract 执行 OCR
func (m *VisualModule) performOCRNative(imgBytes []byte, language string) (string, error) {
	// 创建临时文件保存图像
	tmpFile := fmt.Sprintf("/tmp/ocr_input_%d.png", time.Now().UnixNano())
	if err := os.WriteFile(tmpFile, imgBytes, 0644); err != nil {
		return "", fmt.Errorf("failed to write temp file: %w", err)
	}
	defer os.Remove(tmpFile)

	// 使用 tesseract 进行 OCR
	outputBase := fmt.Sprintf("/tmp/ocr_output_%d", time.Now().UnixNano())
	cmd := exec.Command("tesseract", tmpFile, outputBase, "-l", language)

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("tesseract OCR failed: %w (make sure tesseract is installed)", err)
	}

	// 读取输出文件
	outputFile := outputBase + ".txt"
	defer os.Remove(outputFile)

	textBytes, err := os.ReadFile(outputFile)
	if err != nil {
		return "", fmt.Errorf("failed to read OCR output: %w", err)
	}

	return string(textBytes), nil
}

// encodeImageToPNG 将图像编码为 PNG 格式的 base64
func encodeImageToPNG(img image.Image) (string, error) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", fmt.Errorf("failed to encode PNG: %w", err)
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}
