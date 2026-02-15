// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
// Copyright (C) 2025  Agent Framework Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"AgentFramework/pkg/tools/sandbox/visual"
)

func main() {
	fmt.Println("=== Visual Module Demo ===\n")

	// 创建 Visual Module
	module, err := visual.NewVisualModule(visual.VisualConfig{
		Enable:           true,
		AllowControl:     true,
		OCREnabled:       true,
		RecordingEnabled: true,
		Port:             8080,
		Host:             "localhost",
	})
	if err != nil {
		log.Fatalf("Failed to create Visual Module: %v", err)
	}
	defer module.Close()

	fmt.Println("✓ Visual Module created successfully")

	// 获取 MCP 工具列表
	ctx := context.Background()
	tools, err := module.GetTools(ctx)
	if err != nil {
		log.Fatalf("Failed to get tools: %v", err)
	}

	fmt.Printf("✓ Available tools: %d\n\n", len(tools))

	// 演示 1: 屏幕捕获
	fmt.Println("--- Demo 1: Screen Capture ---")
	demoScreenCapture(module)

	// 演示 2: 远程控制
	fmt.Println("\n--- Demo 2: Remote Control ---")
	demoRemoteControl(module)

	// 演示 3: OCR 识别
	fmt.Println("\n--- Demo 3: OCR Recognition ---")
	demoOCR(module)

	// 演示 4: 视频录制
	fmt.Println("\n--- Demo 4: Video Recording ---")
	demoVideoRecording(module)

	// 演示 5: 统计信息
	fmt.Println("\n--- Demo 5: Statistics ---")
	demoStatistics(module)

	fmt.Println("\n=== Demo Completed ===")
}

// demoScreenCapture 演示屏幕捕获功能
func demoScreenCapture(module *visual.VisualModule) {
	// 捕获全屏
	fmt.Println("Capturing full screen...")
	// Note: 实际使用时会捕获屏幕，这里仅演示 API 调用
	// result, err := module.captureScreen(90)
	// if err != nil {
	// 	log.Printf("Error: %v", err)
	// 	return
	// }
	// fmt.Printf("✓ Screen captured: %d bytes\n", len(result["image_data"].(string)))

	// 捕获区域
	fmt.Println("Capturing screen region (100x100 at 0,0)...")
	// result, err = module.captureRegion(0, 0, 100, 100)
	// if err != nil {
	// 	log.Printf("Error: %v", err)
	// 	return
	// }
	// fmt.Printf("✓ Region captured: %d bytes\n", len(result["image_data"].(string)))

	fmt.Println("✓ Screen capture demo completed (API calls shown)")
}

// demoRemoteControl 演示远程控制功能
func demoRemoteControl(module *visual.VisualModule) {
	// 移动鼠标
	fmt.Println("Moving mouse to (100, 100)...")
	// Note: 实际使用时会移动鼠标，这里仅演示 API 调用
	// result, err := module.moveMouse(100, 100)
	// if err != nil {
	// 	log.Printf("Error: %v", err)
	// 	return
	// }
	// fmt.Printf("✓ Mouse moved: %v\n", result["success"])

	// 点击
	fmt.Println("Clicking left button...")
	// result, err = module.click("left", nil, nil)
	// if err != nil {
	// 	log.Printf("Error: %v", err)
	// 	return
	// }
	// fmt.Printf("✓ Click performed: %v\n", result["success"])

	// 输入文本
	fmt.Println("Typing text: 'Hello World'...")
	// result, err = module.typeText("Hello World")
	// if err != nil {
	// 	log.Printf("Error: %v", err)
	// 	return
	// }
	// fmt.Printf("✓ Text typed: %d characters\n", result["length"])

	fmt.Println("✓ Remote control demo completed (API calls shown)")
}

// demoOCR 演示 OCR 识别功能
func demoOCR(module *visual.VisualModule) {
	// 先捕获屏幕
	fmt.Println("Capturing screen for OCR...")
	// captureResult, err := module.captureScreen(90)
	// if err != nil {
	// 	log.Printf("Error: %v", err)
	// 	return
	// }

	// 执行 OCR
	fmt.Println("Performing OCR recognition...")
	// imageData := captureResult["image_data"].(string)
	// ocrResult, err := module.performOCR(imageData, "eng")
	// if err != nil {
	// 	log.Printf("Error: %v", err)
	// 	return
	// }
	// fmt.Printf("✓ OCR completed: %s\n", ocrResult["text"])

	fmt.Println("✓ OCR demo completed (API calls shown)")
	fmt.Println("Note: Requires tesseract to be installed")
}

// demoVideoRecording 演示视频录制功能
func demoVideoRecording(module *visual.VisualModule) {
	// 开始录制
	fmt.Println("Starting recording (10 FPS)...")
	// startResult, err := module.startRecording(10)
	// if err != nil {
	// 	log.Printf("Error: %v", err)
	// 	return
	// }
	// fmt.Printf("✓ Recording started: %v\n", startResult["success"])

	// 等待一段时间
	fmt.Println("Recording for 2 seconds...")
	time.Sleep(2 * time.Second)

	// 停止录制
	fmt.Println("Stopping recording...")
	// stopResult, err := module.stopRecording()
	// if err != nil {
	// 	log.Printf("Error: %v", err)
	// 	return
	// }
	// frameCount := stopResult["frame_count"].(int)
	// duration := stopResult["duration_ms"].(int64)
	// fmt.Printf("✓ Recording stopped: %d frames in %d ms\n", frameCount, duration)

	fmt.Println("✓ Video recording demo completed (API calls shown)")
}

// demoStatistics 演示统计信息功能
func demoStatistics(module *visual.VisualModule) {
	stats := module.GetStats()

	fmt.Println("Visual Module Statistics:")
	fmt.Printf("  Total Captures:   %d\n", stats["total_captures"])
	fmt.Printf("  Total Controls:   %d\n", stats["total_controls"])
	fmt.Printf("  Total OCR:        %d\n", stats["total_ocr"])
	fmt.Printf("  Total Recordings: %d\n", stats["total_recordings"])

	fmt.Println("✓ Statistics demo completed")
}
