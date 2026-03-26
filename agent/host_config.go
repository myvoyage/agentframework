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

// For network interaction purposes, when this Program is used over a network,
// the source code of the Program must be made available to users of the network.
// You can comply with this requirement by providing a link to the source code
// repository in your user interface or documentation.

// SPDX-License-Identifier: AGPL-3.0-or-later

package agent

import (
	"io"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type ThreadStoreSpec struct {
	Type        string `yaml:"type"`                  // memory|file|redis|sql
	Dir         string `yaml:"dir,omitempty"`         // for file
	RedisAddr   string `yaml:"redisAddr,omitempty"`   // for redis
	RedisPrefix string `yaml:"redisPrefix,omitempty"` // for redis
	DriverName  string `yaml:"driver,omitempty"`      // for sql
	DSN         string `yaml:"dsn,omitempty"`         // for sql
	Table       string `yaml:"table,omitempty"`       // for sql
	// Memory-specific configuration
	MaxMessages    int `yaml:"maxMessages,omitempty"`    // Maximum number of messages to keep per thread
	MaxMessageSize int `yaml:"maxMessageSize,omitempty"` // Maximum size of a single message in bytes
	TTL            int `yaml:"ttl,omitempty"`            // Time-to-live for threads in seconds
}

type AgentSpec struct {
	Name         string   `yaml:"name"`
	Kind         string   `yaml:"kind"`                   // chat|react
	Model        string   `yaml:"model"`                  // model key
	Instructions string   `yaml:"instructions,omitempty"` // system prompt
	Tools        []string `yaml:"tools,omitempty"`        // tool names
	Middlewares  []string `yaml:"middlewares,omitempty"`  // middleware keys
	// Memory management options
	MaxMessages    int     `yaml:"maxMessages,omitempty"`    // Maximum number of messages to keep in the thread history
	MaxMessageSize int     `yaml:"maxMessageSize,omitempty"` // Maximum size of a single message in bytes
	TrimRatio      float64 `yaml:"trimRatio,omitempty"`      // Ratio of messages to keep when trimming (0.0-1.0)
	EnableTrimming bool    `yaml:"enableTrimming,omitempty"` // Enable intelligent message trimming
}

// NodeSpec represents a node in a DAG workflow
type NodeSpec struct {
	ID        string `yaml:"id"`
	Kind      string `yaml:"kind,omitempty"`  // agent|inline
	AgentName string `yaml:"agent,omitempty"` // Reference to defined agent (for kind=agent)
	// Inline agent configuration (for kind=inline)
	InlineName         string   `yaml:"inlineName,omitempty"`         // Name of the inline agent
	InlineKind         string   `yaml:"inlineKind,omitempty"`         // chat|react
	InlineModel        string   `yaml:"inlineModel,omitempty"`        // model key
	InlineInstructions string   `yaml:"inlineInstructions,omitempty"` // system prompt
	InlineTools        []string `yaml:"inlineTools,omitempty"`        // tool names
	InlineMiddlewares  []string `yaml:"inlineMiddlewares,omitempty"`  // middleware keys
}

type WorkflowSpec struct {
	Name       string            `yaml:"name"`
	Kind       string            `yaml:"kind"`                 // sequential|aggregating_parallel|routing|dag
	Model      string            `yaml:"model,omitempty"`      // for routing
	Steps      []string          `yaml:"steps,omitempty"`      // for sequential
	Agents     []string          `yaml:"agents,omitempty"`     // for parallel / aggregating
	Aggregator string            `yaml:"aggregator,omitempty"` // aggregating workflow
	Routes     map[string]string `yaml:"routes,omitempty"`     // for routing: key -> workflow name
	Nodes      []NodeSpec        `yaml:"nodes,omitempty"`      // for DAG
	Edges      map[string]string `yaml:"edges,omitempty"`      // for DAG: from -> to (simple) or from -> [to1, to2]
	// Extended Edges support: from -> [to...]
	EdgesList map[string][]string `yaml:"edgesList,omitempty"` // explicit list support
}

// ModelCacheSpec represents model cache configuration
type ModelCacheSpec struct {
	Enabled         bool `yaml:"enabled,omitempty"`         // Whether model caching is enabled
	MaxSize         int  `yaml:"maxSize,omitempty"`         // Maximum number of models to cache
	TTL             int `yaml:"ttl,omitempty"`             // Time-to-live for cached models in seconds
	CleanupInterval int `yaml:"cleanupInterval,omitempty"` // Cleanup interval in seconds
}

// MemoryMonitorSpec represents memory monitor configuration
type MemoryMonitorSpec struct {
	Enabled        bool `yaml:"enabled,omitempty"`        // Whether memory monitoring is enabled
	Interval       int `yaml:"interval,omitempty"`       // Interval between measurements in seconds
	HistorySize    int `yaml:"historySize,omitempty"`    // Number of historical measurements to keep
	AlertThreshold int `yaml:"alertThreshold,omitempty"` // Memory usage threshold for alerts in MB
	AlertInterval  int `yaml:"alertInterval,omitempty"`  // Minimum interval between alerts in seconds
}

// MemoryManagementSpec represents memory management configuration
type MemoryManagementSpec struct {
	ModelCache    ModelCacheSpec    `yaml:"modelCache,omitempty"`    // Model cache configuration
	MemoryMonitor MemoryMonitorSpec `yaml:"memoryMonitor,omitempty"` // Memory monitor configuration
	Checkpoint    struct {
		MaxCheckpoints  int `yaml:"maxCheckpoints,omitempty"`  // Maximum number of checkpoints to keep
		TTL             int `yaml:"ttl,omitempty"`             // Time-to-live for checkpoints in seconds
		CleanupInterval int `yaml:"cleanupInterval,omitempty"` // Cleanup interval in seconds
	} `yaml:"checkpoint,omitempty"` // Checkpoint store configuration
}

// ContextStoreSpec represents context storage configuration
type ContextStoreSpec struct {
	Enabled   bool                `yaml:"enabled,omitempty"`   // Whether context storage is enabled
	Type      string              `yaml:"type,omitempty"`      // Context store type: openviking|memory|none
	OpenViking OpenVikingStoreSpec `yaml:"openviking,omitempty"` // OpenViking specific configuration
}

// OpenVikingStoreSpec represents OpenViking context store configuration
type OpenVikingStoreSpec struct {
	Endpoint   string `yaml:"endpoint,omitempty"`   // OpenViking server endpoint
	APIKey     string `yaml:"apiKey,omitempty"`     // API key for authentication
	Workspace  string `yaml:"workspace,omitempty"`  // Default workspace path
	Timeout    int `yaml:"timeout,omitempty"`    // Request timeout in seconds
	MaxRetries int `yaml:"maxRetries,omitempty"` // Maximum number of retries
	AutoSync   bool `yaml:"autoSync,omitempty"`   // Enable automatic context synchronization
}

// TokenCompressionSpec represents token compression configuration
type TokenCompressionSpec struct {
	Enabled             bool   `yaml:"enabled,omitempty"`             // Whether token compression is enabled
	Strategy            string `yaml:"strategy,omitempty"`            // Compression strategy: truncate|summarize|semantic|hybrid
	TargetTokens        int    `yaml:"targetTokens,omitempty"`        // Target token count after compression
	MinTokens           int    `yaml:"minTokens,omitempty"`           // Minimum token count to consider compression
	MaxTokens           int    `yaml:"maxTokens,omitempty"`           // Maximum token count to trigger compression
	PreserveSystemMessages bool `yaml:"preserveSystemMessages,omitempty"` // Whether to preserve system messages
	SummaryModelName    string `yaml:"summaryModelName,omitempty"`    // Model to use for summarization
	SummaryMaxTokens    int `yaml:"summaryMaxTokens,omitempty"`    // Maximum tokens for summaries
	Temperature         float64 `yaml:"temperature,omitempty"`         // Temperature for summary generation
	CheckInterval       int `yaml:"checkInterval,omitempty"`       // Check interval in seconds
	// Memory-based configuration
	SessionTokenLimit   int `yaml:"sessionTokenLimit,omitempty"`   // Token limit for session memory
	DailyTokenLimit     int `yaml:"dailyTokenLimit,omitempty"`     // Token limit for daily memory
	LongTermTokenLimit  int `yaml:"longTermTokenLimit,omitempty"`  // Token limit for long-term memory
}

// AsyncTaskSpec represents async task configuration
type AsyncTaskSpec struct {
	Enabled     bool `yaml:"enabled,omitempty"`     // Whether async task execution is enabled
	MaxTasks    int `yaml:"maxTasks,omitempty"`    // Maximum concurrent tasks
	TaskTimeout int `yaml:"taskTimeout,omitempty"` // Default task timeout in seconds
}

// SchedulerSpec represents scheduler configuration
type SchedulerSpec struct {
	Enabled   bool  `yaml:"enabled,omitempty"`   // Whether scheduler is enabled
	Timezone  string `yaml:"timezone,omitempty"`  // Timezone for cron jobs
	MaxJobs   int `yaml:"maxJobs,omitempty"`   // Maximum concurrent jobs
	JobTimeout int `yaml:"jobTimeout,omitempty"` // Job timeout in seconds
}

// HeartbeatSpec represents heartbeat service configuration
type HeartbeatSpec struct {
	Enabled  bool `yaml:"enabled,omitempty"`  // Whether heartbeat is enabled
	Interval int `yaml:"interval,omitempty"` // Heartbeat interval in seconds
	Timeout  int `yaml:"timeout,omitempty"`  // Heartbeat timeout in seconds
}

// =============================================================================
// Channel Configuration (微信/飞书/钉钉等渠道配置)
// =============================================================================

// ChannelsSpec represents all channel configurations
type ChannelsSpec struct {
	WeChat   *WeChatChannelSpec   `yaml:"wechat,omitempty"`   // 微信渠道配置
	Lark     *LarkChannelSpec     `yaml:"lark,omitempty"`     // 飞书渠道配置
	DingTalk *DingTalkChannelSpec `yaml:"dingtalk,omitempty"` // 钉钉渠道配置
	Telegram *TelegramChannelSpec `yaml:"telegram,omitempty"` // Telegram 渠道配置
	Slack    *SlackChannelSpec    `yaml:"slack,omitempty"`    // Slack 渠道配置
}

// WeChatChannelSpec represents WeChat channel configuration
type WeChatChannelSpec struct {
	Enabled bool `yaml:"enabled"` // Whether channel is enabled

	// Channel type: wecom/clawbot/mp/miniprogram
	Type string `yaml:"type"`

	// 企业微信配置
	CorpID     string `yaml:"corpId,omitempty"`
	AgentID    string `yaml:"agentId,omitempty"`
	CorpSecret string `yaml:"corpSecret,omitempty"`

	// 公众号/小程序配置
	AppID     string `yaml:"appId,omitempty"`
	AppSecret string `yaml:"appSecret,omitempty"`

	// 消息配置
	Token       string `yaml:"token,omitempty"`
	EncryptKey  string `yaml:"encryptKey,omitempty"`

	// ClawBot 配置
	ClawBotURL string `yaml:"clawBotUrl,omitempty"`

	// 服务配置
	Port       int  `yaml:"port,omitempty"`
	AutoReply  bool `yaml:"autoReply,omitempty"`  // 自动回复

	// 会话策略
	SessionPolicy string `yaml:"sessionPolicy,omitempty"` // open/restricted/private
}

// LarkChannelSpec represents Lark/Feishu channel configuration
type LarkChannelSpec struct {
	Enabled bool `yaml:"enabled"` // Whether channel is enabled

	// 域配置: feishu(国内版) / lark(国际版)
	Domain string `yaml:"domain"`

	// 应用配置
	AppID     string `yaml:"appId,omitempty"`
	AppSecret string `yaml:"appSecret,omitempty"`

	// 连接模式: websocket(推荐) / webhook
	ConnectionMode string `yaml:"connectionMode"`

	// Webhook 配置
	Port        int    `yaml:"port,omitempty"`
	EncryptKey  string `yaml:"encryptKey,omitempty"`
	VerifyToken string `yaml:"verifyToken,omitempty"`

	// 会话策略 (参考官方 OpenClaw 插件)
	DMPolicy      string   `yaml:"dmPolicy,omitempty"`      // pairing/allowlist/open/disabled
	DMAllowlist   []string `yaml:"dmAllowlist,omitempty"`   // 私信白名单
	GroupPolicy   string   `yaml:"groupPolicy,omitempty"`   // open/allowlist/disabled
	GroupAllowlist []string `yaml:"groupAllowlist,omitempty"` // 群聊白名单

	// 流式回复配置
	Streaming      bool `yaml:"streaming,omitempty"`      // 启用流式回复
	TextChunkLimit int  `yaml:"textChunkLimit,omitempty"` // 文本分块大小

	// 功能配置
	TypingIndicator   bool `yaml:"typingIndicator,omitempty"`   // 显示输入中状态
	ResolveSenderNames bool `yaml:"resolveSenderNames,omitempty"` // 解析发送者名称

	// 限流配置
	RateLimitEnabled bool `yaml:"rateLimitEnabled,omitempty"` // 启用限流
	RateLimitPerSec  int  `yaml:"rateLimitPerSec,omitempty"`  // 每秒请求数

	// Bot 信息
	BotName string `yaml:"botName,omitempty"`
}

// DingTalkChannelSpec represents DingTalk channel configuration
type DingTalkChannelSpec struct {
	Enabled bool `yaml:"enabled"` // Whether channel is enabled

	// 应用配置
	ClientID     string `yaml:"clientId,omitempty"`
	ClientSecret string `yaml:"clientSecret,omitempty"`
	AgentID      string `yaml:"agentId,omitempty"`

	// 连接配置
	Port       int    `yaml:"port,omitempty"`
	Token      string `yaml:"token,omitempty"`
	EncryptKey string `yaml:"encryptKey,omitempty"`

	// 会话策略
	SessionPolicy string `yaml:"sessionPolicy,omitempty"`
}

// TelegramChannelSpec represents Telegram channel configuration
type TelegramChannelSpec struct {
	Enabled bool `yaml:"enabled"` // Whether channel is enabled

	// Bot 配置
	BotToken string `yaml:"botToken,omitempty"`

	// Webhook 配置
	WebhookURL string `yaml:"webhookUrl,omitempty"`
	Port       int    `yaml:"port,omitempty"`

	// 会话策略
	AllowedChats []int64 `yaml:"allowedChats,omitempty"` // 允许的聊天 ID
}

// SlackChannelSpec represents Slack channel configuration
type SlackChannelSpec struct {
	Enabled bool `yaml:"enabled"` // Whether channel is enabled

	// App 配置
	BotToken    string `yaml:"botToken,omitempty"`
	AppToken    string `yaml:"appToken,omitempty"`
	SigningSecret string `yaml:"signingSecret,omitempty"`

	// Socket Mode
	SocketMode bool `yaml:"socketMode,omitempty"`

	// 会话策略
	AllowedChannels []string `yaml:"allowedChannels,omitempty"`
}

type HostConfig struct {
	Name           string                 `yaml:"name,omitempty"`           // application name
	Version        string                 `yaml:"version,omitempty"`        // application version
	DefaultModel   string                 `yaml:"defaultModel"`             // default model key
	Models         map[string]ModelConfig `yaml:"models,omitempty"`         // model configurations
	ThreadStore    ThreadStoreSpec        `yaml:"threadStore"`              // thread store config
	Memory         MemoryManagementSpec   `yaml:"memory,omitempty"`         // memory management config
	Agents         []AgentSpec            `yaml:"agents"`                   // agent declarations
	Workflows      []WorkflowSpec         `yaml:"workflows,omitempty"`      // workflow declarations
	SkillSystemDir string                 `yaml:"skillSystemDir,omitempty"` // skill system directory
	ContextStore   ContextStoreSpec       `yaml:"contextStore,omitempty"`  // context storage configuration
	Messaging      *MessagingConfig       `yaml:"messaging,omitempty"`      // messaging configuration
	Scheduler      *SchedulerSpec         `yaml:"scheduler,omitempty"`      // scheduler configuration
	Heartbeat      *HeartbeatSpec         `yaml:"heartbeat,omitempty"`      // heartbeat configuration
	AsyncTask      *AsyncTaskSpec         `yaml:"asyncTask,omitempty"`      // async task configuration
	TokenCompression *TokenCompressionSpec `yaml:"tokenCompression,omitempty"` // token compression configuration
	Channels       *ChannelsSpec          `yaml:"channels,omitempty"`       // channel configurations (微信/飞书/钉钉等)
	Extensions     map[string]interface{} `yaml:"-"`                        // extensions for plugins (not serialized)
}

func LoadHostConfig(r io.Reader) (*HostConfig, error) {
	// Read all content to apply env expansion
	content, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	// Expand environment variables: ${VAR} or $VAR
	expanded := os.ExpandEnv(string(content))

	// Normalize snake_case keys to camelCase for compatibility
	expanded = normalizeYAMLKeys(expanded)

	var cfg HostConfig
	if err := yaml.Unmarshal([]byte(expanded), &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// normalizeYAMLKeys converts snake_case keys to camelCase for compatibility
// This allows users to use either format in their config files
func normalizeYAMLKeys(content string) string {
	// Map of snake_case to camelCase field names
	replacements := map[string]string{
		"default_model:":      "defaultModel:",
		"skill_system_dir:":   "skillSystemDir:",
		"base_url:":           "baseurl:",
		"api_key:":            "apikey:",
		"max_tokens:":         "maxtokens:",
		"max_retries:":        "maxretries:",
		"retry_interval:":     "retryinterval:",
		"top_p:":              "topp:",
		"top_k:":              "topk:",
		"log_level:":          "loglevel:",
		"thread_store:":       "threadStore:",
		"context_store:":      "contextStore:",
		"async_task:":         "asyncTask:",
		"token_compression:":  "tokenCompression:",
		"max_messages:":       "maxMessages:",
		"max_message_size:":   "maxMessageSize:",
		"cleanup_interval:":   "cleanupInterval:",
		"job_timeout:":        "jobTimeout:",
		"task_timeout:":       "taskTimeout:",
		"target_tokens:":      "targetTokens:",
		"min_tokens:":         "minTokens:",
		"summary_model_name:": "summaryModelName:",
		"summary_max_tokens:": "summaryMaxTokens:",
		"check_interval:":     "checkInterval:",
		"session_token_limit:":  "sessionTokenLimit:",
		"daily_token_limit:":    "dailyTokenLimit:",
		"long_term_token_limit:": "longTermTokenLimit:",
		"preserve_system_messages:": "preserveSystemMessages:",
		"max_checkpoints:":      "maxCheckpoints:",
		"max_size:":             "maxSize:",
		"history_size:":         "historySize:",
		"alert_threshold:":     "alertThreshold:",
		"alert_interval:":      "alertInterval:",
		"auto_sync:":           "autoSync:",
		"corp_id:":             "corpId:",
		"agent_id:":            "agentId:",
		"corp_secret:":         "corpSecret:",
		"app_id:":              "appId:",
		"app_secret:":          "appSecret:",
		"encrypt_key:":         "encryptKey:",
		"claw_bot_url:":        "clawBotUrl:",
		"auto_reply:":          "autoReply:",
		"session_policy:":      "sessionPolicy:",
		"connection_mode:":     "connectionMode:",
		"verify_token:":        "verifyToken:",
		"dm_policy:":           "dmPolicy:",
		"dm_allowlist:":        "dmAllowlist:",
		"group_policy:":        "groupPolicy:",
		"group_allowlist:":     "groupAllowlist:",
		"text_chunk_limit:":    "textChunkLimit:",
		"typing_indicator:":    "typingIndicator:",
		"resolve_sender_names:": "resolveSenderNames:",
		"rate_limit_enabled:":  "rateLimitEnabled:",
		"rate_limit_per_sec:":  "rateLimitPerSec:",
		"bot_name:":            "botName:",
		"bot_token:":           "botToken:",
		"webhook_url:":         "webhookUrl:",
		"allowed_chats:":       "allowedChats:",
		"allowed_channels:":    "allowedChannels:",
		"socket_mode:":         "socketMode:",
		"signing_secret:":      "signingSecret:",
		"client_id:":           "clientId:",
		"client_secret:":       "clientSecret:",
	}

	result := content
	for snake, camel := range replacements {
		result = strings.ReplaceAll(result, snake, camel)
	}
	return result
}

func LoadHostConfigFile(path string) (*HostConfig, error) {
	// os.Open is replaced by os.ReadFile to support expansion easily
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return LoadHostConfig(f)
}

// SaveHostConfig saves the host configuration to a writer
func SaveHostConfig(w io.Writer, cfg *HostConfig) error {
	// Marshal the configuration to YAML
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	// Write the YAML data to the writer
	_, err = w.Write(data)
	return err
}

// SaveHostConfigFile saves the host configuration to a file
func SaveHostConfigFile(path string, cfg *HostConfig) error {
	// Create the file for writing
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	// Save the configuration to the file
	return SaveHostConfig(f, cfg)
}