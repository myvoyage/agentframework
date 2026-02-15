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

// SPDX-License-Identifier: AGPL-3.0-or-later

package visual

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

// TestNewVisualModule tests the creation of a new visual module
func TestNewVisualModule(t *testing.T) {
	tests := []struct {
		name    string
		config  VisualConfig
		wantErr bool
	}{
		{
			name: "default config",
			config: VisualConfig{
				Enable:           true,
				AllowControl:     true,
				OCREnabled:       true,
				RecordingEnabled: true,
			},
			wantErr: false,
		},
		{
			name: "minimal config",
			config: VisualConfig{
				Enable: true,
			},
			wantErr: false,
		},
		{
			name: "disabled module",
			config: VisualConfig{
				Enable: false,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			module, err := NewVisualModule(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewVisualModule() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				if module == nil {
					t.Error("NewVisualModule() returned nil module")
				}
				if module.config.Port == 0 {
					t.Error("Port should have default value")
				}
				if module.config.Host == "" {
					t.Error("Host should have default value")
				}
			}
		})
	}
}

// TestGetTools tests the GetTools method
func TestGetTools(t *testing.T) {
	tests := []struct {
		name      string
		config    VisualConfig
		wantCount int
	}{
		{
			name: "enabled module",
			config: VisualConfig{
				Enable: true,
			},
			wantCount: 9, // 9 MCP tools
		},
		{
			name: "disabled module",
			config: VisualConfig{
				Enable: false,
			},
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			module, err := NewVisualModule(tt.config)
			if err != nil {
				t.Fatalf("NewVisualModule() error = %v", err)
			}

			ctx := context.Background()
			tools, err := module.GetTools(ctx)
			if err != nil {
				t.Errorf("GetTools() error = %v", err)
				return
			}

			if len(tools) != tt.wantCount {
				t.Errorf("GetTools() returned %d tools, want %d", len(tools), tt.wantCount)
			}
		})
	}
}

// TestCaptureScreen tests the screen capture functionality
func TestCaptureScreen(t *testing.T) {
	module, err := NewVisualModule(VisualConfig{
		Enable: true,
	})
	if err != nil {
		t.Fatalf("NewVisualModule() error = %v", err)
	}

	result, err := module.captureScreen(90)
	if err != nil {
		t.Errorf("captureScreen() error = %v", err)
		return
	}

	// Check result structure
	if success, ok := result["success"].(bool); !ok || !success {
		t.Error("captureScreen() should return success=true")
	}

	// Check statistics
	stats := module.GetStats()
	if captures, ok := stats["total_captures"].(int64); !ok || captures != 1 {
		t.Errorf("total_captures should be 1, got %v", captures)
	}
}

// TestCaptureRegion tests the region capture functionality
func TestCaptureRegion(t *testing.T) {
	module, err := NewVisualModule(VisualConfig{
		Enable: true,
	})
	if err != nil {
		t.Fatalf("NewVisualModule() error = %v", err)
	}

	result, err := module.captureRegion(0, 0, 100, 100)
	if err != nil {
		t.Errorf("captureRegion() error = %v", err)
		return
	}

	// Check result structure
	if success, ok := result["success"].(bool); !ok || !success {
		t.Error("captureRegion() should return success=true")
	}

	if x, ok := result["x"].(int); !ok || x != 0 {
		t.Errorf("x should be 0, got %v", x)
	}
}

// TestMoveMouse tests the mouse movement functionality
func TestMoveMouse(t *testing.T) {
	tests := []struct {
		name          string
		allowControl  bool
		x             int
		y             int
		expectSuccess bool
	}{
		{
			name:          "control enabled",
			allowControl:  true,
			x:             100,
			y:             100,
			expectSuccess: true,
		},
		{
			name:          "control disabled",
			allowControl:  false,
			x:             100,
			y:             100,
			expectSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			module, err := NewVisualModule(VisualConfig{
				Enable:       true,
				AllowControl: tt.allowControl,
			})
			if err != nil {
				t.Fatalf("NewVisualModule() error = %v", err)
			}

			result, err := module.moveMouse(tt.x, tt.y)
			if err != nil {
				t.Errorf("moveMouse() error = %v", err)
				return
			}

			success, ok := result["success"].(bool)
			if !ok {
				t.Error("result should have success field")
			}
			if success != tt.expectSuccess {
				t.Errorf("success = %v, want %v", success, tt.expectSuccess)
			}
		})
	}
}

// TestClick tests the mouse click functionality
func TestClick(t *testing.T) {
	tests := []struct {
		name          string
		allowControl  bool
		button        string
		expectSuccess bool
	}{
		{
			name:          "left click enabled",
			allowControl:  true,
			button:        "left",
			expectSuccess: true,
		},
		{
			name:          "right click enabled",
			allowControl:  true,
			button:        "right",
			expectSuccess: true,
		},
		{
			name:          "control disabled",
			allowControl:  false,
			button:        "left",
			expectSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			module, err := NewVisualModule(VisualConfig{
				Enable:       true,
				AllowControl: tt.allowControl,
			})
			if err != nil {
				t.Fatalf("NewVisualModule() error = %v", err)
			}

			result, err := module.click(tt.button, nil, nil)
			if err != nil {
				t.Errorf("click() error = %v", err)
				return
			}

			success, ok := result["success"].(bool)
			if !ok {
				t.Error("result should have success field")
			}
			if success != tt.expectSuccess {
				t.Errorf("success = %v, want %v", success, tt.expectSuccess)
			}
		})
	}
}

// TestTypeText tests the text input functionality
func TestTypeText(t *testing.T) {
	tests := []struct {
		name          string
		allowControl  bool
		text          string
		expectSuccess bool
	}{
		{
			name:          "control enabled",
			allowControl:  true,
			text:          "Hello World",
			expectSuccess: true,
		},
		{
			name:          "control disabled",
			allowControl:  false,
			text:          "Hello World",
			expectSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			module, err := NewVisualModule(VisualConfig{
				Enable:       true,
				AllowControl: tt.allowControl,
			})
			if err != nil {
				t.Fatalf("NewVisualModule() error = %v", err)
			}

			result, err := module.typeText(tt.text)
			if err != nil {
				t.Errorf("typeText() error = %v", err)
				return
			}

			success, ok := result["success"].(bool)
			if !ok {
				t.Error("result should have success field")
			}
			if success != tt.expectSuccess {
				t.Errorf("success = %v, want %v", success, tt.expectSuccess)
			}
		})
	}
}

// TestPerformOCR tests the OCR functionality
func TestPerformOCR(t *testing.T) {
	tests := []struct {
		name          string
		ocrEnabled    bool
		expectSuccess bool
	}{
		{
			name:          "OCR enabled",
			ocrEnabled:    true,
			expectSuccess: true,
		},
		{
			name:          "OCR disabled",
			ocrEnabled:    false,
			expectSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			module, err := NewVisualModule(VisualConfig{
				Enable:     true,
				OCREnabled: tt.ocrEnabled,
			})
			if err != nil {
				t.Fatalf("NewVisualModule() error = %v", err)
			}

			// Use a simple base64 encoded image (1x1 pixel)
			imageData := "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg=="

			result, err := module.performOCR(imageData, "eng")
			if err != nil {
				t.Errorf("performOCR() error = %v", err)
				return
			}

			success, ok := result["success"].(bool)
			if !ok {
				t.Error("result should have success field")
			}

			// If OCR is enabled but tesseract is not installed, the test may fail
			// This is expected behavior, so we just check the structure
			if tt.ocrEnabled && !success {
				// Check if error message indicates tesseract is not installed
				if errMsg, ok := result["error"].(string); ok {
					t.Logf("OCR failed (tesseract may not be installed): %s", errMsg)
				}
			} else if success != tt.expectSuccess {
				t.Errorf("success = %v, want %v", success, tt.expectSuccess)
			}
		})
	}
}

// TestRecording tests the video recording functionality
func TestRecording(t *testing.T) {
	module, err := NewVisualModule(VisualConfig{
		Enable:           true,
		RecordingEnabled: true,
	})
	if err != nil {
		t.Fatalf("NewVisualModule() error = %v", err)
	}

	// Start recording
	startResult, err := module.startRecording(10)
	if err != nil {
		t.Errorf("startRecording() error = %v", err)
		return
	}

	if success, ok := startResult["success"].(bool); !ok || !success {
		t.Error("startRecording() should return success=true")
	}

	// Wait a bit for some frames to be captured
	time.Sleep(200 * time.Millisecond)

	// Stop recording
	stopResult, err := module.stopRecording()
	if err != nil {
		t.Errorf("stopRecording() error = %v", err)
		return
	}

	if success, ok := stopResult["success"].(bool); !ok || !success {
		t.Error("stopRecording() should return success=true")
	}

	// Check frame count
	if frameCount, ok := stopResult["frame_count"].(int); !ok || frameCount < 0 {
		t.Errorf("frame_count should be >= 0, got %v", frameCount)
	}
}

// TestRecordingAlreadyInProgress tests starting recording when already recording
func TestRecordingAlreadyInProgress(t *testing.T) {
	module, err := NewVisualModule(VisualConfig{
		Enable:           true,
		RecordingEnabled: true,
	})
	if err != nil {
		t.Fatalf("NewVisualModule() error = %v", err)
	}

	// Start recording
	_, err = module.startRecording(10)
	if err != nil {
		t.Errorf("startRecording() error = %v", err)
		return
	}

	// Try to start again
	result, err := module.startRecording(10)
	if err != nil {
		t.Errorf("startRecording() error = %v", err)
		return
	}

	if success, ok := result["success"].(bool); !ok || success {
		t.Error("startRecording() should return success=false when already recording")
	}

	// Clean up
	module.stopRecording()
}

// TestStopRecordingNotStarted tests stopping recording when not started
func TestStopRecordingNotStarted(t *testing.T) {
	module, err := NewVisualModule(VisualConfig{
		Enable:           true,
		RecordingEnabled: true,
	})
	if err != nil {
		t.Fatalf("NewVisualModule() error = %v", err)
	}

	result, err := module.stopRecording()
	if err != nil {
		t.Errorf("stopRecording() error = %v", err)
		return
	}

	if success, ok := result["success"].(bool); !ok || success {
		t.Error("stopRecording() should return success=false when not recording")
	}
}

// TestGetStats tests the statistics functionality
func TestGetStats(t *testing.T) {
	module, err := NewVisualModule(VisualConfig{
		Enable:       true,
		AllowControl: true,
		OCREnabled:   true,
	})
	if err != nil {
		t.Fatalf("NewVisualModule() error = %v", err)
	}

	// Perform some operations
	module.captureScreen(90)
	module.moveMouse(100, 100)

	stats := module.GetStats()

	// Check all stat fields exist
	fields := []string{"total_captures", "total_controls", "total_ocr", "total_recordings"}
	for _, field := range fields {
		if _, ok := stats[field]; !ok {
			t.Errorf("stats should have %s field", field)
		}
	}

	// Check values
	if captures, ok := stats["total_captures"].(int64); !ok || captures != 1 {
		t.Errorf("total_captures should be 1, got %v", captures)
	}

	if controls, ok := stats["total_controls"].(int64); !ok || controls != 1 {
		t.Errorf("total_controls should be 1, got %v", controls)
	}
}

// TestClose tests the Close method
func TestClose(t *testing.T) {
	module, err := NewVisualModule(VisualConfig{
		Enable:           true,
		RecordingEnabled: true,
	})
	if err != nil {
		t.Fatalf("NewVisualModule() error = %v", err)
	}

	// Start recording
	module.startRecording(10)

	// Close should stop recording
	err = module.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}

	// Recording should be stopped
	if module.recorder.recording {
		t.Error("recording should be stopped after Close()")
	}
}

// TestMCPToolIntegration tests the MCP tool integration
func TestMCPToolIntegration(t *testing.T) {
	module, err := NewVisualModule(VisualConfig{
		Enable:       true,
		AllowControl: true,
		OCREnabled:   true,
	})
	if err != nil {
		t.Fatalf("NewVisualModule() error = %v", err)
	}

	ctx := context.Background()
	tools, err := module.GetTools(ctx)
	if err != nil {
		t.Fatalf("GetTools() error = %v", err)
	}

	// Test each tool has proper Info
	for i, tool := range tools {
		info, err := tool.Info(ctx)
		if err != nil {
			t.Errorf("Tool %d Info() error = %v", i, err)
			continue
		}

		if info.Name == "" {
			t.Errorf("Tool %d should have a name", i)
		}

		if info.Desc == "" {
			t.Errorf("Tool %d should have a description", i)
		}
	}
}

// TestVisualCaptureScreenTool tests the capture screen tool
func TestVisualCaptureScreenTool(t *testing.T) {
	module, err := NewVisualModule(VisualConfig{
		Enable: true,
	})
	if err != nil {
		t.Fatalf("NewVisualModule() error = %v", err)
	}

	tool := &visualCaptureScreenTool{module: module}
	ctx := context.Background()

	// Test Info
	info, err := tool.Info(ctx)
	if err != nil {
		t.Fatalf("Info() error = %v", err)
	}
	if info.Name != "visual_capture_screen" {
		t.Errorf("Name = %s, want visual_capture_screen", info.Name)
	}

	// Test InvokableRun
	args := map[string]any{"quality": 90}
	argsJSON, _ := json.Marshal(args)
	result, err := tool.InvokableRun(ctx, string(argsJSON))
	if err != nil {
		t.Errorf("InvokableRun() error = %v", err)
	}

	var resultMap map[string]any
	if err := json.Unmarshal([]byte(result), &resultMap); err != nil {
		t.Errorf("Failed to unmarshal result: %v", err)
	}
}

// TestVisualGetStatsTool tests the get stats tool
func TestVisualGetStatsTool(t *testing.T) {
	module, err := NewVisualModule(VisualConfig{
		Enable: true,
	})
	if err != nil {
		t.Fatalf("NewVisualModule() error = %v", err)
	}

	// Perform some operations
	module.captureScreen(90)

	tool := &visualGetStatsTool{module: module}
	ctx := context.Background()

	// Test InvokableRun
	result, err := tool.InvokableRun(ctx, "{}")
	if err != nil {
		t.Errorf("InvokableRun() error = %v", err)
	}

	var stats map[string]any
	if err := json.Unmarshal([]byte(result), &stats); err != nil {
		t.Errorf("Failed to unmarshal result: %v", err)
	}

	if captures, ok := stats["total_captures"].(float64); !ok || captures != 1 {
		t.Errorf("total_captures should be 1, got %v", captures)
	}
}

// TestConcurrentOperations tests concurrent access to the module
func TestConcurrentOperations(t *testing.T) {
	module, err := NewVisualModule(VisualConfig{
		Enable:       true,
		AllowControl: true,
	})
	if err != nil {
		t.Fatalf("NewVisualModule() error = %v", err)
	}

	// Run multiple operations concurrently
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			module.captureScreen(90)
			module.moveMouse(100, 100)
			module.GetStats()
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Check stats
	stats := module.GetStats()
	if captures, ok := stats["total_captures"].(int64); !ok || captures != 10 {
		t.Errorf("total_captures should be 10, got %v", captures)
	}
}
