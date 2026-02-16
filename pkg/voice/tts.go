// Agent Framework - Text-to-Speech Module
// Copyright (C) 2025 Agent Framework Contributors
// SPDX-License-Identifier: Apache-2.0

package voice

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// TTSModule 文字转语音模块
type TTSModule struct {
	config TTSConfig
	mu     sync.RWMutex
	stats   *TTSStats
	engine  TTSEngine
}

// TTSConfig 文字转语音配置
type TTSConfig struct {
	Engine             string   `json:"engine"`               // 引擎类型（local, api, edge）
	Voice              string   `json:"voice"`                // 说话人ID
	Rate               int      `json:"rate"`                 // 语速（-10到10，0为正常）
	Pitch              int      `json:"pitch"`                // 音调（-10到10，0为正常）
	Volume             float64  `json:"volume"`               // 音量（0.0到1.0）
	OutputFormat       string   `json:"output_format"`        // 输出格式（wav, mp3, ogg）
	OutputQuality      string   `json:"output_quality"`       // 输出质量（low, medium, high）
	EnableSSML        bool     `json:"enable_ssml"`          // 启用SSML标签支持
	EnableEmotion      bool     `json:"enable_emotion"`        // 启用情感表达
	DefaultEmotion     string   `json:"default_emotion"`       // 默认情感（neutral, happy, sad, angry）
	TempDir            string   `json:"temp_dir"`             // 临时文件目录
}

// TTSStats 文字转语音统计信息
type TTSStats struct {
	TotalSyntheses int64     `json:"total_syntheses"`
	SuccessCount   int64     `json:"success_count"`
	FailureCount   int64     `json:"failure_count"`
	TotalCharacters int64     `json:"total_characters"`
	TotalDuration  float64   `json:"total_duration_seconds"` // 合成音频总时长
	mu                  sync.RWMutex `json:"-"`
}

// TTSEngine 文字转语音引擎接口
type TTSEngine interface {
	Synthesize(ctx context.Context, text string, options TTSOptions) (*TTSResult, error)
	SynthesizeToFile(ctx context.Context, text string, outputPath string, options TTSOptions) error
	GetVoices() []TTSVoice
	GetCapabilities() TTSCapabilities
	Close() error
}

// TTSOptions 文字转语音选项
type TTSOptions struct {
	Rate           int     `json:"rate"`            // 语速调整
	Pitch          int     `json:"pitch"`           // 音调调整
	Volume         float64 `json:"volume"`          // 音量调整
	Voice          string  `json:"voice"`           // 说话人
	Emotion        string  `json:"emotion"`         // 情感
	OutputFormat   string  `json:"output_format"`   // 输出格式
	OutputQuality  string  `json:"output_quality"`  // 输出质量
	EnableSSML     bool    `json:"enable_ssml"`    // 启用SSML
}

// TTSVoice 说话人信息
type TTSVoice struct {
	ID          string   `json:"id"`            // 说话人ID
	Name        string   `json:"name"`          // 显示名称
	Gender      string   `json:"gender"`        // 性别（male, female, neutral）
	Language    string   `json:"language"`      // 语言代码
	Age         string   `json:"age"`           // 年龄段（young, adult, elderly）
	Style       string   `json:"style"`         // 风格（formal, casual）
	Description string   `json:"description"`   // 描述
}

// TTSCapabilities 文字转语音能力
type TTSCapabilities struct {
	SampleRates       []int     `json:"sample_rates"`        // 支持采样率
	SupportedFormats  []string  `json:"supported_formats"`   // 支持输出格式
	SupportedVoices   []string  `json:"supported_voices"`    // 支持的说话人ID
	Streaming        bool      `json:"streaming"`           // 支持流式输出
	SSMLSupport      bool      `json:"ssml_support"`       // SSML标签支持
	EmotionSupport   bool      `json:"emotion_support"`    // 情感表达支持
	VoiceCloning    bool      `json:"voice_cloning"`      // 声音克隆支持
	MaxCharacters   int       `json:"max_characters"`      // 单次最大字符数
}

// TTSResult 文字转语音结果
type TTSResult struct {
	Success        bool        `json:"success"`
	AudioData     []byte      `json:"audio_data,omitempty"`      // 音频数据
	AudioPath     string      `json:"audio_path,omitempty"`      // 音频文件路径
	Duration      float64     `json:"duration"`                  // 音频时长（秒）
	Format        string      `json:"format"`                    // 音频格式
	SampleRate    int         `json:"sample_rate"`               // 采样率
	Voice         string      `json:"voice"`                    // 使用的说话人
	Characters    int         `json:"characters"`                // 字符数
	Error         string      `json:"error,omitempty"`          // 错误信息
}

// NewTTSModule 创建文字转语音模块实例
func NewTTSModule(config TTSConfig) (*TTSModule, error) {
	if config.Rate < -10 {
		config.Rate = -10
	}
	if config.Rate > 10 {
		config.Rate = 10
	}
	if config.Pitch < -10 {
		config.Pitch = -10
	}
	if config.Pitch > 10 {
		config.Pitch = 10
	}
	if config.Volume <= 0 {
		config.Volume = 1.0
	}
	if config.Volume > 1.0 {
		config.Volume = 1.0
	}
	if config.OutputFormat == "" {
		config.OutputFormat = "wav"
	}
	if config.OutputQuality == "" {
		config.OutputQuality = "medium"
	}
	if config.TempDir == "" {
		config.TempDir = os.TempDir()
	}

	// 创建引擎
	var engine TTSEngine
	var err error

	switch config.Engine {
	case "local":
		engine, err = NewLocalTTSEngine(config)
	case "api":
		engine, err = NewAPITTSEngine(config)
	case "edge":
		engine, err = NewEdgeTTSEngine(config)
	default:
		engine, err = NewLocalTTSEngine(config) // 默认本地引擎
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create TTS engine: %w", err)
	}

	stats := &TTSStats{}

	return &TTSModule{
		config: config,
		stats:  stats,
		engine: engine,
	}, nil
}

// GetTools 返回文字转语音模块的 MCP 工具列表
func (m *TTSModule) GetTools(ctx context.Context) ([]tool.BaseTool, error) {
	tools := []tool.BaseTool{
		// 文字转语音工具
		&ttsSynthesizeTool{module: m},
		// 文字转语音到文件工具
		&ttsSynthesizeToFileTool{module: m},
		// 获取说话人列表工具
		&ttsVoicesTool{module: m},
		// 获取能力工具
		&ttsCapabilitiesTool{module: m},
		// 批量合成工具
		&ttsBatchTool{module: m},
		// SSML 合成工具
		&ttsSSMLTool{module: m},
	}

	return tools, nil
}

// 文字转语音工具
type ttsSynthesizeTool struct {
	module *TTSModule
}

func (t *ttsSynthesizeTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "tts_synthesize",
		Desc: "Convert text to speech using text-to-speech engine",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"text": {
				Type:     "string",
				Desc:     "Text to convert to speech",
				Required:  true,
			},
			"voice": {
				Type: "string",
				Desc: "Voice ID to use for synthesis",
			},
			"rate": {
				Type: "integer",
				Desc: "Speaking rate adjustment (-10 to 10, 0 is normal)",
			},
			"pitch": {
				Type: "integer",
				Desc: "Pitch adjustment (-10 to 10, 0 is normal)",
			},
			"volume": {
				Type: "number",
				Desc: "Volume level (0.0 to 1.0)",
			},
			"emotion": {
				Type: "string",
				Desc: "Emotion for speech (neutral, happy, sad, angry)",
			},
			"output_format": {
				Type: "string",
				Desc: "Output audio format (wav, mp3, ogg)",
			},
		}),
	}, nil
}

func (t *ttsSynthesizeTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var args struct {
		Text         string  `json:"text"`
		Voice        string  `json:"voice"`
		Rate         int     `json:"rate"`
		Pitch        int     `json:"pitch"`
		Volume       float64 `json:"volume"`
		Emotion      string  `json:"emotion"`
		OutputFormat string  `json:"output_format"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	result, err := t.module.synthesize(ctx, args.Text, args.Voice, args.Rate, args.Pitch, args.Volume, args.Emotion, args.OutputFormat)
	if err != nil {
		return "", err
	}

	output, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// 文字转语音到文件工具
type ttsSynthesizeToFileTool struct {
	module *TTSModule
}

func (t *ttsSynthesizeToFileTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "tts_synthesize_to_file",
		Desc: "Convert text to speech and save to file",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"text": {
				Type:     "string",
				Desc:     "Text to convert to speech",
				Required:  true,
			},
			"output_path": {
				Type:     "string",
				Desc:     "Output file path for audio",
				Required:  true,
			},
			"voice": {
				Type: "string",
				Desc: "Voice ID to use for synthesis",
			},
			"rate": {
				Type: "integer",
				Desc: "Speaking rate adjustment",
			},
			"output_format": {
				Type: "string",
				Desc: "Output audio format",
			},
		}),
	}, nil
}

func (t *ttsSynthesizeToFileTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var args struct {
		Text       string `json:"text"`
		OutputPath string `json:"output_path"`
		Voice      string `json:"voice"`
		Rate       int    `json:"rate"`
		OutputFormat string `json:"output_format"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	err := t.module.synthesizeToFile(ctx, args.Text, args.OutputPath, args.Voice, args.Rate, args.OutputFormat)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(`{"success":true,"output_path":"%s"}`, args.OutputPath), nil
}

// 获取说话人列表工具
type ttsVoicesTool struct {
	module *TTSModule
}

func (t *ttsVoicesTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name:        "tts_voices",
		Desc:        "Get list of available voices for synthesis",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"language": {
				Type: "string",
				Desc: "Filter voices by language code",
			},
			"gender": {
				Type: "string",
				Desc: "Filter voices by gender (male, female, neutral)",
			},
		}),
	}, nil
}

func (t *ttsVoicesTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var args struct {
		Language string `json:"language"`
		Gender   string `json:"gender"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	result := t.module.getVoices(args.Language, args.Gender)
	output, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// 获取能力工具
type ttsCapabilitiesTool struct {
	module *TTSModule
}

func (t *ttsCapabilitiesTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name:        "tts_capabilities",
		Desc:        "Get text-to-speech engine capabilities",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{}),
	}, nil
}

func (t *ttsCapabilitiesTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	caps := t.module.getCapabilities()
	output, err := json.Marshal(caps)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// 批量合成工具
type ttsBatchTool struct {
	module *TTSModule
}

func (t *ttsBatchTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "tts_batch",
		Desc: "Synthesize multiple texts in batch",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"texts": {
				Type:     "array",
				Desc:     "Array of texts to synthesize",
				Required:  true,
			},
			"output_dir": {
				Type:     "string",
				Desc:     "Output directory for audio files",
				Required:  true,
			},
			"voice": {
				Type: "string",
				Desc: "Voice ID to use for all syntheses",
			},
		}),
	}, nil
}

func (t *ttsBatchTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var args struct {
		Texts     []string `json:"texts"`
		OutputDir string   `json:"output_dir"`
		Voice     string   `json:"voice"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	result, err := t.module.synthesizeBatch(ctx, args.Texts, args.OutputDir, args.Voice)
	if err != nil {
		return "", err
	}

	output, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// SSML 合成工具
type ttsSSMLTool struct {
	module *TTSModule
}

func (t *ttsSSMLTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "tts_ssml",
		Desc: "Synthesize speech from SSML (Speech Synthesis Markup Language)",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"ssml": {
				Type:     "string",
				Desc:     "SSML content to synthesize",
				Required:  true,
			},
			"output_path": {
				Type: "string",
				Desc: "Output file path for audio (optional, returns audio data if not specified)",
			},
		}),
	}, nil
}

func (t *ttsSSMLTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var args struct {
		SSML       string `json:"ssml"`
		OutputPath string `json:"output_path"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	result, err := t.module.synthesizeSSML(ctx, args.SSML, args.OutputPath)
	if err != nil {
		return "", err
	}

	output, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// Close 关闭文字转语音模块
func (m *TTSModule) Close() error {
	return m.engine.Close()
}

// GetStats 获取统计信息
func (m *TTSModule) GetStats() map[string]int64 {
	m.stats.mu.RLock()
	defer m.stats.mu.RUnlock()

	return map[string]int64{
		"total_syntheses":  m.stats.TotalSyntheses,
		"success_count":    m.stats.SuccessCount,
		"failure_count":    m.stats.FailureCount,
		"total_characters": m.stats.TotalCharacters,
	}
}

// ==================== 核心功能实现 ====================

// synthesize 合成语音
func (m *TTSModule) synthesize(ctx context.Context, text, voice string, rate, pitch int, volume float64, emotion, outputFormat string) (*TTSResult, error) {
	m.stats.mu.Lock()
	m.stats.TotalSyntheses++
	m.stats.mu.Unlock()

	if voice == "" {
		voice = m.config.Voice
	}
	if rate == 0 {
		rate = m.config.Rate
	}
	if pitch == 0 {
		pitch = m.config.Pitch
	}
	if volume == 0 {
		volume = m.config.Volume
	}
	if outputFormat == "" {
		outputFormat = m.config.OutputFormat
	}
	if emotion == "" {
		emotion = m.config.DefaultEmotion
	}

	options := TTSOptions{
		Rate:          rate,
		Pitch:         pitch,
		Volume:        volume,
		Voice:         voice,
		Emotion:       emotion,
		OutputFormat:  outputFormat,
		OutputQuality: m.config.OutputQuality,
		EnableSSML:    m.config.EnableSSML,
	}

	result, err := m.engine.Synthesize(ctx, text, options)
	if err != nil {
		m.stats.mu.Lock()
		m.stats.FailureCount++
		m.stats.mu.Unlock()

		return &TTSResult{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	m.stats.mu.Lock()
	m.stats.SuccessCount++
	m.stats.TotalCharacters += int64(result.Characters)
	m.stats.TotalDuration += result.Duration
	m.stats.mu.Unlock()

	return result, nil
}

// synthesizeToFile 合成语音到文件
func (m *TTSModule) synthesizeToFile(ctx context.Context, text, outputPath, voice string, rate int, outputFormat string) error {
	if voice == "" {
		voice = m.config.Voice
	}
	if rate == 0 {
		rate = m.config.Rate
	}
	if outputFormat == "" {
		outputFormat = m.config.OutputFormat
	}

	options := TTSOptions{
		Rate:          rate,
		Pitch:         m.config.Pitch,
		Volume:        m.config.Volume,
		Voice:         voice,
		OutputFormat:  outputFormat,
		OutputQuality: m.config.OutputQuality,
	}

	return m.engine.SynthesizeToFile(ctx, text, outputPath, options)
}

// getVoices 获取说话人列表
func (m *TTSModule) getVoices(language, gender string) map[string]any {
	voices := m.engine.GetVoices()

	var filtered []TTSVoice
	for _, voice := range voices {
		if language != "" && voice.Language != language {
			continue
		}
		if gender != "" && voice.Gender != gender {
			continue
		}
		filtered = append(filtered, voice)
	}

	return map[string]any{
		"success": true,
		"voices":  filtered,
		"count":   len(filtered),
	}
}

// getCapabilities 获取能力信息
func (m *TTSModule) getCapabilities() map[string]any {
	caps := m.engine.GetCapabilities()

	return map[string]any{
		"success":        true,
		"capabilities":  caps,
	}
}

// synthesizeBatch 批量合成
func (m *TTSModule) synthesizeBatch(ctx context.Context, texts []string, outputDir, voice string) (map[string]any, error) {
	if voice == "" {
		voice = m.config.Voice
	}

	results := make([]TTSResult, 0, len(texts))

	for i, text := range texts {
		filename := fmt.Sprintf("output_%03d.%s", i+1, m.config.OutputFormat)
		outputPath := filepath.Join(outputDir, filename)

		err := m.synthesizeToFile(ctx, text, outputPath, voice, m.config.Rate, m.config.OutputFormat)
		if err != nil {
			results = append(results, TTSResult{
				Success: false,
				Error:   err.Error(),
			})
			continue
		}

		results = append(results, TTSResult{
			Success:    true,
			AudioPath:  outputPath,
			Characters: len(text),
		})
	}

	successCount := 0
	for _, r := range results {
		if r.Success {
			successCount++
		}
	}

	return map[string]any{
		"success":   true,
		"total":     len(texts),
		"succeeded": successCount,
		"failed":    len(texts) - successCount,
		"results":   results,
	}, nil
}

// synthesizeSSML 合成 SSML
func (m *TTSModule) synthesizeSSML(ctx context.Context, ssml, outputPath string) (*TTSResult, error) {
	m.stats.mu.Lock()
	m.stats.TotalSyntheses++
	m.stats.mu.Unlock()

	// SSML 合成需要引擎支持
	options := TTSOptions{
		Rate:          m.config.Rate,
		Pitch:         m.config.Pitch,
		Volume:        m.config.Volume,
		Voice:         m.config.Voice,
		OutputFormat:  m.config.OutputFormat,
		OutputQuality: m.config.OutputQuality,
		EnableSSML:    true,
	}

	if outputPath != "" {
		err := m.engine.SynthesizeToFile(ctx, ssml, outputPath, options)
		if err != nil {
			m.stats.mu.Lock()
			m.stats.FailureCount++
			m.stats.mu.Unlock()

			return &TTSResult{
				Success: false,
				Error:   err.Error(),
			}, nil
		}

		m.stats.mu.Lock()
		m.stats.SuccessCount++
		m.stats.mu.Unlock()

		return &TTSResult{
			Success:   true,
			AudioPath: outputPath,
		}, nil
	}

	result, err := m.engine.Synthesize(ctx, ssml, options)
	if err != nil {
		m.stats.mu.Lock()
		m.stats.FailureCount++
		m.stats.mu.Unlock()

		return &TTSResult{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	m.stats.mu.Lock()
	m.stats.SuccessCount++
	m.stats.TotalCharacters += int64(result.Characters)
	m.stats.TotalDuration += result.Duration
	m.stats.mu.Unlock()

	return result, nil
}

// ==================== 引擎实现 ====================

// LocalTTSEngine 本地文字转语音引擎
type LocalTTSEngine struct {
	config TTSConfig
}

// NewLocalTTSEngine 创建本地引擎
func NewLocalTTSEngine(config TTSConfig) (*LocalTTSEngine, error) {
	return &LocalTTSEngine{config: config}, nil
}

func (e *LocalTTSEngine) Synthesize(ctx context.Context, text string, options TTSOptions) (*TTSResult, error) {
	// 使用平台特定 API
	// Windows: SAPI (Speech API)
	// macOS: say 命令
	// Linux: espeak-ng / festival
	
	switch runtime.GOOS {
	case "windows":
		return e.synthesizeWindows(ctx, text, options)
	case "darwin":
		return e.synthesizeMacOS(ctx, text, options)
	case "linux":
		return e.synthesizeLinux(ctx, text, options)
	default:
		return &TTSResult{
			Success: false,
			Error:   fmt.Sprintf("unsupported platform: %s", runtime.GOOS),
		}, nil
	}
}

func (e *LocalTTSEngine) synthesizeWindows(ctx context.Context, text string, options TTSOptions) (*TTSResult, error) {
	// 使用 PowerShell 调用 SAPI
	psScript := fmt.Sprintf(`
Add-Type -AssemblyName System.Speech
$speak = New-Object System.Speech.Synthesis.SpeechSynthesizer
$speak.Speak("%s")
`, escapePowerShell(text))
	
	// 创建临时音频文件
	tmpFile := filepath.Join(e.config.TempDir, fmt.Sprintf("tts_%d.wav", time.Now().UnixNano()))
	psScriptToFile := fmt.Sprintf(`
Add-Type -AssemblyName System.Speech
$speak = New-Object System.Speech.Synthesis.SpeechSynthesizer
$speak.SetOutputToWaveFile("%s")
$speak.Speak("%s")
$speak.SetOutputToNull()
`, tmpFile, escapePowerShell(text))
	
	cmd := exec.CommandContext(ctx, "powershell", "-command", psScriptToFile)
	if err := cmd.Run(); err != nil {
		// 回退到仅播放（不保存文件）
		cmd = exec.CommandContext(ctx, "powershell", "-command", psScript)
		if err := cmd.Run(); err != nil {
			return &TTSResult{
				Success: false,
				Error:   fmt.Sprintf("Windows TTS failed: %v", err),
			}, nil
		}
		return &TTSResult{
			Success:   true,
			Format:    "wav",
			Voice:     options.Voice,
			Characters: len(text),
		}, nil
	}
	
	// 读取音频文件
	audioData, err := os.ReadFile(tmpFile)
	if err != nil {
		return &TTSResult{
			Success: false,
			Error:   fmt.Sprintf("failed to read audio file: %v", err),
		}, nil
	}
	
	return &TTSResult{
		Success:    true,
		AudioData:  audioData,
		AudioPath:  tmpFile,
		Format:     "wav",
		Voice:      options.Voice,
		Characters: len(text),
	}, nil
}

func (e *LocalTTSEngine) synthesizeMacOS(ctx context.Context, text string, options TTSOptions) (*TTSResult, error) {
	// 使用 macOS say 命令
	tmpFile := filepath.Join(e.config.TempDir, fmt.Sprintf("tts_%d.aiff", time.Now().UnixNano()))
	
	args := []string{}
	if options.Voice != "" {
		args = append(args, "-v", options.Voice)
	}
	if options.Rate != 0 {
		// say 命令的语速范围是不同的
		rate := 175 + (options.Rate * 20) // 默认 175
		args = append(args, "-r", fmt.Sprintf("%d", rate))
	}
	args = append(args, "-o", tmpFile, text)
	
	cmd := exec.CommandContext(ctx, "say", args...)
	if err := cmd.Run(); err != nil {
		return &TTSResult{
			Success: false,
			Error:   fmt.Sprintf("macOS TTS failed: %v", err),
		}, nil
	}
	
	// 读取音频文件
	audioData, err := os.ReadFile(tmpFile)
	if err != nil {
		return &TTSResult{
			Success: false,
			Error:   fmt.Sprintf("failed to read audio file: %v", err),
		}, nil
	}
	
	return &TTSResult{
		Success:    true,
		AudioData:  audioData,
		AudioPath:  tmpFile,
		Format:     "aiff",
		Voice:      options.Voice,
		Characters: len(text),
	}, nil
}

func (e *LocalTTSEngine) synthesizeLinux(ctx context.Context, text string, options TTSOptions) (*TTSResult, error) {
	// 使用 espeak-ng 或 espeak
	var cmd *exec.Cmd
	tmpFile := filepath.Join(e.config.TempDir, fmt.Sprintf("tts_%d.wav", time.Now().UnixNano()))
	
	// 检查 espeak-ng 是否可用
	if _, err := exec.LookPath("espeak-ng"); err == nil {
		args := []string{"-w", tmpFile}
		if options.Voice != "" {
			args = append(args, "-v", options.Voice)
		}
		if options.Rate != 0 {
			rate := 175 + (options.Rate * 20)
			args = append(args, "-s", fmt.Sprintf("%d", rate))
		}
		args = append(args, text)
		cmd = exec.CommandContext(ctx, "espeak-ng", args...)
	} else if _, err := exec.LookPath("espeak"); err == nil {
		args := []string{"-w", tmpFile}
		if options.Voice != "" {
			args = append(args, "-v", options.Voice)
		}
		args = append(args, text)
		cmd = exec.CommandContext(ctx, "espeak", args...)
	} else {
		return &TTSResult{
			Success: false,
			Error:   "neither espeak-ng nor espeak is available",
		}, nil
	}
	
	if err := cmd.Run(); err != nil {
		return &TTSResult{
			Success: false,
			Error:   fmt.Sprintf("Linux TTS failed: %v", err),
		}, nil
	}
	
	// 读取音频文件
	audioData, err := os.ReadFile(tmpFile)
	if err != nil {
		return &TTSResult{
			Success: false,
			Error:   fmt.Sprintf("failed to read audio file: %v", err),
		}, nil
	}
	
	return &TTSResult{
		Success:    true,
		AudioData:  audioData,
		AudioPath:  tmpFile,
		Format:     "wav",
		Voice:      options.Voice,
		Characters: len(text),
	}, nil
}

func (e *LocalTTSEngine) SynthesizeToFile(ctx context.Context, text string, outputPath string, options TTSOptions) error {
	result, err := e.Synthesize(ctx, text, options)
	if err != nil {
		return err
	}
	
	if !result.Success {
		return fmt.Errorf(result.Error)
	}
	
	if result.AudioPath != "" && result.AudioPath != outputPath {
		// 复制文件到目标路径
		data, err := os.ReadFile(result.AudioPath)
		if err != nil {
			return err
		}
		return os.WriteFile(outputPath, data, 0644)
	}
	
	return nil
}

func (e *LocalTTSEngine) GetVoices() []TTSVoice {
	// 返回平台可用的说话人列表
	voices := []TTSVoice{}
	
	switch runtime.GOOS {
	case "darwin":
		// 获取 macOS 可用声音
		cmd := exec.Command("say", "-v", "?")
		output, err := cmd.Output()
		if err == nil {
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				if line == "" {
					continue
				}
				parts := strings.SplitN(line, " ", 2)
				if len(parts) > 0 {
					voices = append(voices, TTSVoice{
						ID:       strings.TrimSpace(parts[0]),
						Name:     strings.TrimSpace(parts[0]),
						Language: "en-US", // 默认
					})
				}
			}
		}
	case "linux":
		// 获取 espeak-ng 可用声音
		cmd := exec.Command("espeak-ng", "--voices")
		output, err := cmd.Output()
		if err == nil {
			lines := strings.Split(string(output), "\n")
			for i, line := range lines {
				if i == 0 || line == "" {
					continue
				}
				fields := strings.Fields(line)
				if len(fields) >= 2 {
					voices = append(voices, TTSVoice{
						ID:       fields[1],
						Name:     fields[1],
						Language: fields[0],
					})
				}
			}
		}
	case "windows":
		// Windows 默认声音
		voices = []TTSVoice{
			{ID: "default", Name: "Microsoft David", Gender: "male", Language: "en-US"},
			{ID: "default-female", Name: "Microsoft Zira", Gender: "female", Language: "en-US"},
		}
	}
	
	if len(voices) == 0 {
		voices = []TTSVoice{
			{ID: "default", Name: "Default Voice", Gender: "neutral", Language: "zh-CN"},
		}
	}
	
	return voices
}

func (e *LocalTTSEngine) GetCapabilities() TTSCapabilities {
	return TTSCapabilities{
		SampleRates:      []int{8000, 16000, 22050, 44100, 48000},
		SupportedFormats: []string{"wav", "aiff"},
		SupportedVoices:  []string{"default"},
		Streaming:        false,
		SSMLSupport:      false,
		EmotionSupport:   false,
		VoiceCloning:     false,
		MaxCharacters:    5000,
	}
}

func (e *LocalTTSEngine) Close() error {
	return nil
}

// APITTSEngine API 文字转语音引擎（云服务）
type APITTSEngine struct {
	config    TTSConfig
	apiKey    string
	endpoint  string
	provider  string // "azure", "google", "aws"
}

// NewAPITTSEngine 创建 API 引擎
func NewAPITTSEngine(config TTSConfig) (*APITTSEngine, error) {
	return &APITTSEngine{
		config:   config,
		provider: "default", // 可以从配置读取
	}, nil
}

func (e *APITTSEngine) Synthesize(ctx context.Context, text string, options TTSOptions) (*TTSResult, error) {
	// 调用云服务 API（如 Azure Speech, Google TTS, AWS Polly）
	// 这里提供一个通用的框架实现
	
	// 实际实现需要根据 provider 选择不同的 API
	switch e.provider {
	case "azure":
		return e.synthesizeAzure(ctx, text, options)
	case "google":
		return e.synthesizeGoogle(ctx, text, options)
	case "aws":
		return e.synthesizeAWS(ctx, text, options)
	default:
		return &TTSResult{
			Success: false,
			Error:   "API TTS provider not configured",
		}, nil
	}
}

func (e *APITTSEngine) synthesizeAzure(ctx context.Context, text string, options TTSOptions) (*TTSResult, error) {
	// Azure Speech Service TTS 实现
	// 需要: github.com/Microsoft/cognitive-services-speech-sdk-go
	return &TTSResult{
		Success: false,
		Error:   "Azure TTS requires cognitive-services-speech-sdk-go",
	}, nil
}

func (e *APITTSEngine) synthesizeGoogle(ctx context.Context, text string, options TTSOptions) (*TTSResult, error) {
	// Google Cloud Text-to-Speech 实现
	// 需要: cloud.google.com/go/texttospeech
	return &TTSResult{
		Success: false,
		Error:   "Google TTS requires cloud.google.com/go/texttospeech",
	}, nil
}

func (e *APITTSEngine) synthesizeAWS(ctx context.Context, text string, options TTSOptions) (*TTSResult, error) {
	// AWS Polly 实现
	// 需要: github.com/aws/aws-sdk-go-v2/service/polly
	return &TTSResult{
		Success: false,
		Error:   "AWS Polly requires aws-sdk-go-v2",
	}, nil
}

func (e *APITTSEngine) SynthesizeToFile(ctx context.Context, text string, outputPath string, options TTSOptions) error {
	result, err := e.Synthesize(ctx, text, options)
	if err != nil {
		return err
	}
	
	if !result.Success {
		return fmt.Errorf(result.Error)
	}
	
	if len(result.AudioData) > 0 {
		return os.WriteFile(outputPath, result.AudioData, 0644)
	}
	
	return fmt.Errorf("no audio data returned")
}

func (e *APITTSEngine) GetVoices() []TTSVoice {
	// 从 API 获取可用说话人列表
	return []TTSVoice{
		{ID: "api-default", Name: "API Default Voice", Gender: "neutral", Language: "zh-CN"},
	}
}

func (e *APITTSEngine) GetCapabilities() TTSCapabilities {
	return TTSCapabilities{
		SampleRates:      []int{8000, 16000, 22050, 44100, 48000},
		SupportedFormats: []string{"wav", "mp3", "ogg"},
		SupportedVoices:  []string{"api-default"},
		Streaming:        true,
		SSMLSupport:      true,
		EmotionSupport:   true,
		VoiceCloning:     false,
		MaxCharacters:    5000,
	}
}

func (e *APITTSEngine) Close() error {
	return nil
}

// EdgeTTSEngine 边缘计算文字转语音引擎
type EdgeTTSEngine struct {
	config TTSConfig
}

// NewEdgeTTSEngine 创建边缘引擎
func NewEdgeTTSEngine(config TTSConfig) (*EdgeTTSEngine, error) {
	return &EdgeTTSEngine{config: config}, nil
}

func (e *EdgeTTSEngine) Synthesize(ctx context.Context, text string, options TTSOptions) (*TTSResult, error) {
	// 使用本地边缘模型（如 Sherpa-onnx, Coqui TTS）
	// 这里提供框架实现，实际需要加载本地模型
	
	// 检查是否有可用的边缘 TTS 工具
	if _, err := exec.LookPath("sherpa-onnx-offline-tts"); err == nil {
		return e.synthesizeSherpaOnnx(ctx, text, options)
	}
	
	return &TTSResult{
		Success: false,
		Error:   "Edge TTS requires Sherpa-onnx or similar local TTS engine",
	}, nil
}

func (e *EdgeTTSEngine) synthesizeSherpaOnnx(ctx context.Context, text string, options TTSOptions) (*TTSResult, error) {
	// Sherpa-onnx 离线 TTS 实现
	tmpFile := filepath.Join(e.config.TempDir, fmt.Sprintf("tts_%d.wav", time.Now().UnixNano()))
	
	cmd := exec.CommandContext(ctx, "sherpa-onnx-offline-tts",
		"--output", tmpFile,
		"--text", text,
	)
	
	if err := cmd.Run(); err != nil {
		return &TTSResult{
			Success: false,
			Error:   fmt.Sprintf("Sherpa-onnx TTS failed: %v", err),
		}, nil
	}
	
	// 读取音频文件
	audioData, err := os.ReadFile(tmpFile)
	if err != nil {
		return &TTSResult{
			Success: false,
			Error:   fmt.Sprintf("failed to read audio file: %v", err),
		}, nil
	}
	
	return &TTSResult{
		Success:    true,
		AudioData:  audioData,
		AudioPath:  tmpFile,
		Format:     "wav",
		Characters: len(text),
	}, nil
}

func (e *EdgeTTSEngine) SynthesizeToFile(ctx context.Context, text string, outputPath string, options TTSOptions) error {
	result, err := e.Synthesize(ctx, text, options)
	if err != nil {
		return err
	}
	
	if !result.Success {
		return fmt.Errorf(result.Error)
	}
	
	if len(result.AudioData) > 0 {
		return os.WriteFile(outputPath, result.AudioData, 0644)
	}
	
	return fmt.Errorf("no audio data returned")
}

func (e *EdgeTTSEngine) GetVoices() []TTSVoice {
	return []TTSVoice{
		{ID: "edge-default", Name: "Edge Model Voice", Gender: "neutral", Language: "zh-CN"},
	}
}

func (e *EdgeTTSEngine) GetCapabilities() TTSCapabilities {
	return TTSCapabilities{
		SampleRates:      []int{16000, 22050, 24000},
		SupportedFormats: []string{"wav"},
		SupportedVoices:  []string{"edge-default"},
		Streaming:        false,
		SSMLSupport:      false,
		EmotionSupport:   true,
		VoiceCloning:     false,
		MaxCharacters:    1000,
	}
}

func (e *EdgeTTSEngine) Close() error {
	return nil
}

// ==================== 辅助函数 ====================

// escapePowerShell 转义 PowerShell 特殊字符
func escapePowerShell(s string) string {
	s = strings.ReplaceAll(s, "'", "''")
	s = strings.ReplaceAll(s, "\"", "`\"")
	s = strings.ReplaceAll(s, "\n", "`n")
	s = strings.ReplaceAll(s, "\r", "`r")
	return s
}

// encodeAudioBase64 将音频数据编码为 Base64
func encodeAudioBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}
