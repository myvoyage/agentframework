// Agent Framework - Voice Module Registry
// Copyright (C) 2025 Agent Framework Contributors
// SPDX-License-Identifier: Apache-2.0

package voice

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
)

// VoiceModule 语音模块集合
type VoiceModule struct {
	stt *STTModule
	tts *TTSModule
}

// VoiceConfig 语音模块配置
type VoiceConfig struct {
	STT STTConfig `json:"stt"`
	TTS TTSConfig `json:"tts"`
}

// DefaultVoiceConfig 返回默认配置
func DefaultVoiceConfig() *VoiceConfig {
	return &VoiceConfig{
		STT: STTConfig{
			Engine:           "local",
			Language:         "zh-CN",
			SampleRate:       16000,
			Channels:         1,
			EnableAutoDetect:  false,
			EnablePunctuation: true,
			EnableTimestamp:  false,
			MaxDuration:     300,
		},
		TTS: TTSConfig{
			Engine:         "local",
			Voice:          "default",
			Rate:           0,
			Pitch:          0,
			Volume:         1.0,
			OutputFormat:   "wav",
			OutputQuality:  "medium",
			EnableSSML:     false,
			EnableEmotion:  false,
			DefaultEmotion: "neutral",
		},
	}
}

// NewVoiceModule 创建语音模块
func NewVoiceModule(config *VoiceConfig) (*VoiceModule, error) {
	if config == nil {
		config = DefaultVoiceConfig()
	}

	// 创建 STT 模块
	stt, err := NewSTTModule(config.STT)
	if err != nil {
		return nil, fmt.Errorf("failed to create STT module: %w", err)
	}

	// 创建 TTS 模块
	tts, err := NewTTSModule(config.TTS)
	if err != nil {
		return nil, fmt.Errorf("failed to create TTS module: %w", err)
	}

	return &VoiceModule{
		stt: stt,
		tts: tts,
	}, nil
}

// GetAllTools 获取所有语音工具
func (m *VoiceModule) GetAllTools(ctx context.Context) ([]tool.BaseTool, error) {
	var allTools []tool.BaseTool

	// STT 工具
	sttTools, err := m.stt.GetTools(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get STT tools: %w", err)
	}
	allTools = append(allTools, sttTools...)

	// TTS 工具
	ttsTools, err := m.tts.GetTools(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get TTS tools: %w", err)
	}
	allTools = append(allTools, ttsTools...)

	return allTools, nil
}

// GetSTT 获取 STT 模块
func (m *VoiceModule) GetSTT() *STTModule {
	return m.stt
}

// GetTTS 获取 TTS 模块
func (m *VoiceModule) GetTTS() *TTSModule {
	return m.tts
}

// Close 关闭所有语音模块
func (m *VoiceModule) Close() error {
	if err := m.stt.Close(); err != nil {
		return fmt.Errorf("STT close error: %w", err)
	}

	if err := m.tts.Close(); err != nil {
		return fmt.Errorf("TTS close error: %w", err)
	}

	return nil
}

// GetAllStats 获取所有模块的统计信息
func (m *VoiceModule) GetAllStats() map[string]map[string]interface{} {
	stats := make(map[string]map[string]interface{})

	// Convert STT stats
	sttStats := m.stt.GetStats()
	stats["stt"] = make(map[string]interface{})
	for k, v := range sttStats {
		stats["stt"][k] = v
	}

	// Convert TTS stats
	ttsStats := m.tts.GetStats()
	stats["tts"] = make(map[string]interface{})
	for k, v := range ttsStats {
		stats["tts"][k] = v
	}

	return stats
}
