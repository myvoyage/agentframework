// Agent Framework - Speech-to-Text Module
// Copyright (C) 2025 Agent Framework Contributors
// SPDX-License-Identifier: Apache-2.0

package voice

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// STTModule 语音转文字模块
type STTModule struct {
	config STTConfig
	mu     sync.RWMutex
	stats  *STTStats
	engine STTEngine
}

// STTConfig 语音转文字配置
type STTConfig struct {
	Engine           string   `json:"engine"`             // 引擎类型（local, whisper, api）
	Language         string   `json:"language"`           // 语言代码（zh-CN, en-US等）
	SampleRate       int      `json:"sample_rate"`         // 采样率（Hz）
	Channels         int      `json:"channels"`            // 声道数（1=单声道, 2=立体声）
	EnableAutoDetect  bool     `json:"enable_auto_detect"`   // 启用语言自动检测
	EnablePunctuation bool     `json:"enable_punctuation"`  // 启用标点符号
	EnableTimestamp  bool     `json:"enable_timestamp"`    // 启用时间戳
	MaxDuration     int      `json:"max_duration"`        // 最大录音时长（秒）
	TempDir         string   `json:"temp_dir"`            // 临时文件目录
}

// STTStats 语音转文字统计信息
type STTStats struct {
	TotalTranscriptions int64     `json:"total_transcriptions"`
	SuccessCount      int64     `json:"success_count"`
	FailureCount      int64     `json:"failure_count"`
	TotalAudioSeconds float64    `json:"total_audio_seconds"`
	TotalCharacters   int64     `json:"total_characters"`
	mu                  sync.RWMutex `json:"-"`
}

// STTEngine 语音转文字引擎接口
type STTEngine interface {
	Transcribe(ctx context.Context, audioPath string, options STTOptions) (*STTResult, error)
	TranscribeStream(ctx context.Context, audioStream io.Reader, options STTOptions) (*STTResult, error)
	GetCapabilities() STTCapabilities
	Close() error
}

// STTOptions 语音转文字选项
type STTOptions struct {
	Language         string        `json:"language"`
	EnableAutoDetect  bool          `json:"enable_auto_detect"`
	EnablePunctuation bool          `json:"enable_punctuation"`
	EnableTimestamp  bool          `json:"enable_timestamp"`
	Callback         func(string)   `json:"-"`
}

// STTCapabilities 语音转文字能力
type STTCapabilities struct {
	SampleRates     []int     `json:"sample_rates"`
	SupportedLanguages []string `json:"supported_languages"`
	Streaming       bool      `json:"streaming"`
	SpeakerDiarization bool `json:"speaker_diarization"` // 说话人分离
	Timestamps      bool      `json:"timestamps"`
}

// STTResult 语音转文字结果
type STTResult struct {
	Success    bool          `json:"success"`
	Text       string        `json:"text"`
	Language   string        `json:"language"`
	Duration   float64       `json:"duration"`
	Confidence float64       `json:"confidence,omitempty"`
	Segments  []TextSegment `json:"segments,omitempty"`
	Error      string        `json:"error,omitempty"`
}

// TextSegment 文本段落（带时间戳）
type TextSegment struct {
	Text      string  `json:"text"`
	StartTime int     `json:"start_time"`
	EndTime   int     `json:"end_time"`
	Speaker   int     `json:"speaker,omitempty"`
	Confidence float64 `json:"confidence,omitempty"`
}

// NewSTTModule 创建语音转文字模块实例
func NewSTTModule(config STTConfig) (*STTModule, error) {
	if config.SampleRate <= 0 {
		config.SampleRate = 16000 // 默认16kHz
	}
	if config.Channels <= 0 {
		config.Channels = 1 // 默认单声道
	}
	if config.Language == "" {
		config.Language = "zh-CN" // 默认中文
	}
	if config.MaxDuration <= 0 {
		config.MaxDuration = 300 // 默认5分钟
	}
	if config.TempDir == "" {
		config.TempDir = os.TempDir()
	}

	// 创建引擎
	var engine STTEngine
	var err error

	switch config.Engine {
	case "local":
		engine, err = NewLocalSTTEngine(config)
	case "whisper":
		engine, err = NewWhisperSTTEngine(config)
	case "api":
		engine, err = NewAPISTTEngine(config)
	default:
		engine, err = NewLocalSTTEngine(config) // 默认本地引擎
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create STT engine: %w", err)
	}

	stats := &STTStats{}

	return &STTModule{
		config: config,
		stats:  stats,
		engine: engine,
	}, nil
}

// GetTools 返回语音转文字模块的 MCP 工具列表
func (m *STTModule) GetTools(ctx context.Context) ([]tool.BaseTool, error) {
	tools := []tool.BaseTool{
		// 语音转文字工具
		&sttTranscribeTool{module: m},
		// 流式语音转文字工具
		&sttTranscribeStreamTool{module: m},
		// 获取能力工具
		&sttCapabilitiesTool{module: m},
		// 录音并转文字工具
		&sttRecordAndTranscribeTool{module: m},
	}

	return tools, nil
}

// 语音转文字工具
type sttTranscribeTool struct {
	module *STTModule
}

func (t *sttTranscribeTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "stt_transcribe",
		Desc: "Transcribe audio file to text using speech-to-text engine",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"audio_path": {
				Type:     "string",
				Desc:     "Path to audio file to transcribe",
				Required:  true,
			},
			"language": {
				Type: "string",
				Desc: "Language code (e.g., zh-CN, en-US, ja-JP)",
			},
			"enable_punctuation": {
				Type: "boolean",
				Desc: "Enable punctuation in output",
			},
			"enable_timestamp": {
				Type: "boolean",
				Desc: "Enable timestamp information",
			},
		}),
	}, nil
}

func (t *sttTranscribeTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var args struct {
		AudioPath         string `json:"audio_path"`
		Language          string `json:"language"`
		EnablePunctuation bool   `json:"enable_punctuation"`
		EnableTimestamp   bool   `json:"enable_timestamp"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	if args.Language == "" {
		args.Language = t.module.config.Language
	}

	result, err := t.module.transcribe(ctx, args.AudioPath, args.Language, args.EnablePunctuation, args.EnableTimestamp)
	if err != nil {
		return "", err
	}

	output, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// 流式语音转文字工具
type sttTranscribeStreamTool struct {
	module *STTModule
}

func (t *sttTranscribeStreamTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "stt_transcribe_stream",
		Desc: "Transcribe streaming audio to text in real-time",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"audio_data": {
				Type:     "string",
				Desc:     "Base64 encoded audio data",
				Required:  true,
			},
			"format": {
				Type: "string",
				Desc: "Audio format (wav, mp3, ogg, flac)",
			},
			"language": {
				Type: "string",
				Desc: "Language code for transcription",
			},
		}),
	}, nil
}

func (t *sttTranscribeStreamTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var args struct {
		AudioData string `json:"audio_data"`
		Format    string `json:"format"`
		Language  string `json:"language"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	result, err := t.module.transcribeStream(ctx, args.AudioData, args.Format, args.Language)
	if err != nil {
		return "", err
	}

	output, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// 获取能力工具
type sttCapabilitiesTool struct {
	module *STTModule
}

func (t *sttCapabilitiesTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name:        "stt_capabilities",
		Desc:        "Get speech-to-text engine capabilities",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{}),
	}, nil
}

func (t *sttCapabilitiesTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	caps := t.module.engine.GetCapabilities()
	output, err := json.Marshal(caps)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// 录音并转文字工具
type sttRecordAndTranscribeTool struct {
	module *STTModule
}

func (t *sttRecordAndTranscribeTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "stt_record_transcribe",
		Desc: "Record audio and transcribe to text in one operation",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"duration": {
				Type: "integer",
				Desc: "Recording duration in seconds",
			},
			"language": {
				Type: "string",
				Desc: "Language code for transcription",
			},
			"output_file": {
				Type: "string",
				Desc: "Optional output file path for audio recording",
			},
		}),
	}, nil
}

func (t *sttRecordAndTranscribeTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var args struct {
		Duration    int    `json:"duration"`
		Language    string `json:"language"`
		OutputFile  string `json:"output_file"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	result, err := t.module.recordAndTranscribe(ctx, args.Duration, args.Language, args.OutputFile)
	if err != nil {
		return "", err
	}

	output, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// Close 关闭语音转文字模块
func (m *STTModule) Close() error {
	return m.engine.Close()
}

// GetStats 获取统计信息
func (m *STTModule) GetStats() map[string]interface{} {
	m.stats.mu.RLock()
	defer m.stats.mu.RUnlock()

	return map[string]interface{}{
		"total_transcriptions":  m.stats.TotalTranscriptions,
		"success_count":        m.stats.SuccessCount,
		"failure_count":        m.stats.FailureCount,
		"total_audio_seconds":  m.stats.TotalAudioSeconds,
		"total_characters":      m.stats.TotalCharacters,
	}
}

// ==================== 核心功能实现 ====================

// transcribe 转录音频文件
func (m *STTModule) transcribe(ctx context.Context, audioPath, language string, enablePunctuation, enableTimestamp bool) (*STTResult, error) {
	m.stats.mu.Lock()
	m.stats.TotalTranscriptions++
	m.stats.mu.Unlock()

	// 验证文件存在
	if _, err := os.Stat(audioPath); os.IsNotExist(err) {
		return &STTResult{
			Success: false,
			Error:   fmt.Sprintf("Audio file not found: %s", audioPath),
		}, nil
	}

	// 获取音频时长
	duration, err := m.getAudioDuration(audioPath)
	if err != nil {
		return &STTResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to get audio duration: %v", err),
		}, nil
	}

	// 转录音频
	options := STTOptions{
		Language:         language,
		EnablePunctuation: enablePunctuation,
		EnableTimestamp:  enableTimestamp,
	}

	result, err := m.engine.Transcribe(ctx, audioPath, options)
	if err != nil {
		m.stats.mu.Lock()
		m.stats.FailureCount++
		m.stats.mu.Unlock()

		return &STTResult{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	// 更新统计
	m.stats.mu.Lock()
	m.stats.SuccessCount++
	m.stats.TotalAudioSeconds += duration
	m.stats.TotalCharacters += int64(len(result.Text))
	m.stats.mu.Unlock()

	return result, nil
}

// transcribeStream 流式转录音频
func (m *STTModule) transcribeStream(ctx context.Context, audioData, format, language string) (*STTResult, error) {
	m.stats.mu.Lock()
	m.stats.TotalTranscriptions++
	m.stats.mu.Unlock()

	// 解码音频数据
	// 这里简化处理，实际应该使用音频解码库
	audioReader := newMockAudioReader(audioData)

	options := STTOptions{
		Language: language,
	}

	result, err := m.engine.TranscribeStream(ctx, audioReader, options)
	if err != nil {
		m.stats.mu.Lock()
		m.stats.FailureCount++
		m.stats.mu.Unlock()

		return &STTResult{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	m.stats.mu.Lock()
	m.stats.SuccessCount++
	m.stats.TotalCharacters += int64(len(result.Text))
	m.stats.mu.Unlock()

	return result, nil
}

// recordAndTranscribe 录音并转录
func (m *STTModule) recordAndTranscribe(ctx context.Context, duration int, language, outputFile string) (*STTResult, error) {
	if duration <= 0 {
		duration = 10 // 默认10秒
	}
	if duration > m.config.MaxDuration {
		duration = m.config.MaxDuration
	}
	if language == "" {
		language = m.config.Language
	}

	// 创建临时音频文件
	tempFile := filepath.Join(m.config.TempDir, fmt.Sprintf("recording_%d.wav", time.Now().UnixNano()))
	if outputFile != "" {
		tempFile = outputFile
	}

	// 录音（平台特定实现）
	err := m.recordAudio(tempFile, duration)
	if err != nil {
		return &STTResult{
			Success: false,
			Error:   fmt.Sprintf("Recording failed: %v", err),
		}, nil
	}

	// 转录
	return m.transcribe(ctx, tempFile, language, m.config.EnablePunctuation, m.config.EnableTimestamp)
}

// getAudioDuration 获取音频时长
func (m *STTModule) getAudioDuration(audioPath string) (float64, error) {
	// 简化实现，实际应该使用音频库解析
	return 30.0, nil
}

// recordAudio 录音到文件
func (m *STTModule) recordAudio(outputPath string, duration int) error {
	// 平台特定录音实现
	// Windows: 使用 winmm 库
	// macOS: 使用 AVFoundation
	// Linux: 使用 ALSA 或 PulseAudio
	return fmt.Errorf("audio recording not implemented")
}

// ==================== 引擎实现 ====================

// LocalSTTEngine 本地语音转文字引擎
type LocalSTTEngine struct {
	config STTConfig
}

// NewLocalSTTEngine 创建本地引擎
func NewLocalSTTEngine(config STTConfig) (*LocalSTTEngine, error) {
	return &LocalSTTEngine{config: config}, nil
}

func (e *LocalSTTEngine) Transcribe(ctx context.Context, audioPath string, options STTOptions) (*STTResult, error) {
	// 使用平台特定 API
	// Windows: Speech Platform
	// macOS: Speech Framework
	// Linux: PocketSphinx 或 CMU Sphinx

	return &STTResult{
		Success:  false,
		Error:    "Local STT not implemented",
		Language: options.Language,
	}, nil
}

func (e *LocalSTTEngine) TranscribeStream(ctx context.Context, audioStream io.Reader, options STTOptions) (*STTResult, error) {
	return &STTResult{
		Success: false,
		Error:   "Stream transcription not implemented",
	}, nil
}

func (e *LocalSTTEngine) GetCapabilities() STTCapabilities {
	return STTCapabilities{
		SampleRates:       []int{8000, 16000, 44100, 48000},
		SupportedLanguages: []string{"zh-CN", "en-US", "ja-JP", "ko-KR"},
		Streaming:         false,
		SpeakerDiarization: false,
		Timestamps:        false,
	}
}

func (e *LocalSTTEngine) Close() error {
	return nil
}

// WhisperSTTEngine OpenAI Whisper 引擎
type WhisperSTTEngine struct {
	config STTConfig
	// whisper 模型加载
}

// NewWhisperSTTEngine 创建 Whisper 引擎
func NewWhisperSTTEngine(config STTConfig) (*WhisperSTTEngine, error) {
	return &WhisperSTTEngine{config: config}, nil
}

func (e *WhisperSTTEngine) Transcribe(ctx context.Context, audioPath string, options STTOptions) (*STTResult, error) {
	// 使用 ggergan/whisper 或 github.com/ggergan/go-whisper
	return &STTResult{
		Success: false,
		Error:   "Whisper STT not implemented",
	}, nil
}

func (e *WhisperSTTEngine) TranscribeStream(ctx context.Context, audioStream io.Reader, options STTOptions) (*STTResult, error) {
	return &STTResult{
		Success: false,
		Error:   "Whisper stream transcription not implemented",
	}, nil
}

func (e *WhisperSTTEngine) GetCapabilities() STTCapabilities {
	return STTCapabilities{
		SampleRates:       []int{16000},
		SupportedLanguages: []string{"zh", "en", "es", "fr", "de", "ja", "ko"},
		Streaming:         false,
		SpeakerDiarization: true,
		Timestamps:        true,
	}
}

func (e *WhisperSTTEngine) Close() error {
	return nil
}

// APISTTEngine API 语音转文字引擎（云服务）
type APISTTEngine struct {
	config STTConfig
	// API 客户端
}

// NewAPISTTEngine 创建 API 引擎
func NewAPISTTEngine(config STTConfig) (*APISTTEngine, error) {
	return &APISTTEngine{config: config}, nil
}

func (e *APISTTEngine) Transcribe(ctx context.Context, audioPath string, options STTOptions) (*STTResult, error) {
	// 调用云服务 API（如 Azure Speech, Google Speech-to-Text, AWS Transcribe）
	return &STTResult{
		Success: false,
		Error:   "API STT not implemented",
	}, nil
}

func (e *APISTTEngine) TranscribeStream(ctx context.Context, audioStream io.Reader, options STTOptions) (*STTResult, error) {
	return &STTResult{
		Success: false,
		Error:   "API stream transcription not implemented",
	}, nil
}

func (e *APISTTEngine) GetCapabilities() STTCapabilities {
	return STTCapabilities{
		SampleRates:       []int{8000, 16000, 44100, 48000},
		SupportedLanguages: []string{"zh-CN", "en-US", "ja-JP", "ko-KR", "es-ES", "fr-FR"},
		Streaming:         true,
		SpeakerDiarization: true,
		Timestamps:        true,
	}
}

func (e *APISTTEngine) Close() error {
	return nil
}

// ==================== 辅助函数 ====================

// mockAudioReader 模拟音频读取器
type mockAudioReader struct {
	data string
	pos  int
}

func newMockAudioReader(data string) *mockAudioReader {
	return &mockAudioReader{data: data, pos: 0}
}

func (r *mockAudioReader) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n = copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}
