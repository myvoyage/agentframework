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

package agent

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/cloudwego/eino-ext/components/model/ollama"
	"github.com/cloudwego/eino-ext/components/model/openai"
	// "github.com/cloudwego/eino-ext/components/model/vllm"
)

// ModelConfig contains the configuration for a specific model
type ModelConfig struct {
	Type    string         `json:"type"`     // e.g., "ollama", "vllm", "openai"
	Model   string         `json:"model"`    // e.g., "llama3", "gpt-4"
	APIKey  string         `json:"api_key"`  // for cloud models
	BaseURL string         `json:"base_url"` // for local models
	Options map[string]any `json:"options"`  // model-specific options

	// Enhanced configuration options
	Timeout       int               `json:"timeout"`        // Request timeout in seconds (default: 30)
	MaxRetries    int               `json:"max_retries"`    // Maximum number of retries (default: 0)
	RetryInterval int               `json:"retry_interval"` // Retry interval in seconds (default: 1)
	Temperature   float64           `json:"temperature"`    // Sampling temperature (default: 0.7)
	MaxTokens     int               `json:"max_tokens"`     // Maximum number of tokens to generate (default: 1024)
	TopP          float64           `json:"top_p"`          // Nucleus sampling parameter (default: 1.0)
	TopK          int               `json:"top_k"`          // Top-k sampling parameter (default: 0)
	LogLevel      string            `json:"log_level"`      // Log level for model operations (default: "info")
	Priority      int               `json:"priority"`       // Model priority for load balancing (default: 0)
	Enabled       bool              `json:"enabled"`        // Whether the model is enabled (default: true)
	Headers       map[string]string `json:"headers"`        // Additional HTTP headers
}

// DefaultModelFactoryConfig is the default configuration for the model factory
type DefaultModelFactoryConfig struct {
	Models map[string]ModelConfig `json:"models"` // model name to config mapping
}

// ModelConfigCache caches pre-validated and preprocessed model configurations for faster lookup
var modelConfigCache = struct {
	cache map[string]ModelConfig
	mutex sync.RWMutex
}{cache: make(map[string]ModelConfig)}

// supportedModelTypes contains all supported model types for quick lookup
var supportedModelTypes = map[string]bool{
	"ollama":   true,
	"openai":   true,
	"lmstudio": true,
	// "vllm":     true,
}

// validateModelConfig validates the model configuration with enhanced checks
func validateModelConfig(cfg ModelConfig, modelName string) error {
	// Normalize model type for consistent validation
	modelType := strings.ToLower(cfg.Type)

	// Validate model type
	if !supportedModelTypes[modelType] {
		supportedList := make([]string, 0, len(supportedModelTypes))
		for t := range supportedModelTypes {
			supportedList = append(supportedList, t)
		}
		return fmt.Errorf("model '%s': unsupported model type '%s', supported types are: %s",
			modelName, cfg.Type, strings.Join(supportedList, ", "))
	}

	// Check if model is enabled
	if !cfg.Enabled {
		return fmt.Errorf("model '%s': model is disabled", modelName)
	}

	// Validate required fields based on type with detailed error messages
	switch modelType {
	case "openai":
		if cfg.APIKey == "" {
			return fmt.Errorf("model '%s': APIKey is required for OpenAI models", modelName)
		}
		if cfg.Model == "" {
			return fmt.Errorf("model '%s': Model name is required for OpenAI models", modelName)
		}
		// Validate OpenAI-specific constraints
		if cfg.Temperature < 0 || cfg.Temperature > 2 {
			return fmt.Errorf("model '%s': temperature must be between 0 and 2, got %.2f", modelName, cfg.Temperature)
		}
	case "ollama":
		if cfg.Model == "" {
			return fmt.Errorf("model '%s': Model name is required for Ollama models", modelName)
		}
		// Validate Ollama-specific constraints
		if cfg.BaseURL != "" && !strings.HasPrefix(cfg.BaseURL, "http") {
			return fmt.Errorf("model '%s': Ollama BaseURL must be a valid URL starting with http/https, got '%s'", modelName, cfg.BaseURL)
		}
	case "lmstudio":
		if cfg.Model == "" {
			return fmt.Errorf("model '%s': Model name is required for LM Studio models", modelName)
		}
		if cfg.BaseURL == "" {
			return fmt.Errorf("model '%s': BaseURL is required for LM Studio models", modelName)
		}
	}

	// Validate common fields across all model types
	if cfg.Timeout < 0 {
		return fmt.Errorf("model '%s': timeout cannot be negative, got %d", modelName, cfg.Timeout)
	}
	if cfg.MaxRetries < 0 {
		return fmt.Errorf("model '%s': max_retries cannot be negative, got %d", modelName, cfg.MaxRetries)
	}
	if cfg.RetryInterval < 0 {
		return fmt.Errorf("model '%s': retry_interval cannot be negative, got %d", modelName, cfg.RetryInterval)
	}
	if cfg.Temperature < 0 {
		return fmt.Errorf("model '%s': temperature cannot be negative, got %.2f", modelName, cfg.Temperature)
	}
	if cfg.MaxTokens < 0 {
		return fmt.Errorf("model '%s': max_tokens cannot be negative, got %d", modelName, cfg.MaxTokens)
	}
	if cfg.TopP < 0 || cfg.TopP > 1 {
		return fmt.Errorf("model '%s': top_p must be between 0 and 1, got %.2f", modelName, cfg.TopP)
	}
	if cfg.TopK < 0 {
		return fmt.Errorf("model '%s': top_k cannot be negative, got %d", modelName, cfg.TopK)
	}

	return nil
}

// setDefaultModelConfig sets default values for model configuration with enhanced logic
func setDefaultModelConfig(cfg ModelConfig) ModelConfig {
	// Set default values for enhanced config options with more comprehensive defaults
	if cfg.Timeout <= 0 {
		cfg.Timeout = 30 // Default timeout: 30 seconds
	}
	if cfg.MaxRetries < 0 {
		cfg.MaxRetries = 0 // Default max retries: 0
	}
	if cfg.RetryInterval <= 0 {
		cfg.RetryInterval = 1 // Default retry interval: 1 second
	}
	if cfg.Temperature <= 0 {
		cfg.Temperature = 0.7 // Default temperature: 0.7
	}
	if cfg.MaxTokens <= 0 {
		cfg.MaxTokens = 1024 // Default max tokens: 1024
	}
	if cfg.TopP <= 0 {
		cfg.TopP = 1.0 // Default top_p: 1.0
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = "info" // Default log level: info
	}
	if cfg.Priority <= 0 {
		cfg.Priority = 0 // Default priority: 0
	}
	if cfg.Headers == nil {
		cfg.Headers = make(map[string]string) // Initialize headers map if nil
	}
	if cfg.Options == nil {
		cfg.Options = make(map[string]any) // Initialize options map if nil
	}
	if cfg.Enabled == false {
		cfg.Enabled = true // Default enabled: true
	}

	// Set type-specific defaults
	switch strings.ToLower(cfg.Type) {
	case "ollama":
		if cfg.BaseURL == "" {
			cfg.BaseURL = "http://localhost:11434" // Default Ollama endpoint
		}
	case "openai":
		if cfg.BaseURL == "" {
			cfg.BaseURL = "https://api.openai.com/v1" // Default OpenAI endpoint
		}
	case "lmstudio":
		if cfg.BaseURL == "" {
			cfg.BaseURL = "http://localhost:1234/v1" // Default LM Studio endpoint
		}
		if cfg.APIKey == "" {
			cfg.APIKey = "lm-studio" // Default LM Studio API key
		}
	}

	return cfg
}

// preprocessedModelConfig contains pre-validated and preprocessed model configurations
type preprocessedModelConfig struct {
	configs map[string]ModelConfig
	mutex   sync.RWMutex
}

// NewDefaultModelFactory creates a ModelFactory that supports multiple model types with optimized configuration lookup
func NewDefaultModelFactory(cfg DefaultModelFactoryConfig) ModelFactory {
	// Preprocess and validate all model configurations upfront
	preprocessed := &preprocessedModelConfig{
		configs: make(map[string]ModelConfig, len(cfg.Models)),
	}

	// Preprocess all model configurations at factory creation time
	for modelName, modelCfg := range cfg.Models {
		// Set defaults first
		processedCfg := setDefaultModelConfig(modelCfg)
		// Validate the processed configuration
		if err := validateModelConfig(processedCfg, modelName); err == nil {
			// Only add valid configurations to the preprocessed map
			preprocessed.configs[modelName] = processedCfg
		}
	}

	return func(ctx context.Context, modelName string) (ChatModel, error) {
		// Lookup preprocessed model config with read lock for concurrent safety
		preprocessed.mutex.RLock()
		modelCfg, ok := preprocessed.configs[modelName]
		preprocessed.mutex.RUnlock()

		if !ok {
			// If no preprocessed config found, create a default config with validation
			defaultCfg := setDefaultModelConfig(ModelConfig{
				Type:    "ollama",
				Model:   modelName,
				Enabled: true,
			})

			// Validate the default configuration
			if err := validateModelConfig(defaultCfg, modelName); err != nil {
				return nil, fmt.Errorf("failed to create default config for model '%s': %w", modelName, err)
			}

			// Cache the default configuration for future use
			preprocessed.mutex.Lock()
			preprocessed.configs[modelName] = defaultCfg
			preprocessed.mutex.Unlock()

			modelCfg = defaultCfg
		}

		// Create model based on type with detailed error handling
		switch strings.ToLower(modelCfg.Type) {
		case "ollama":
			ollamaCfg := &ollama.ChatModelConfig{
				Model:   modelCfg.Model,
				BaseURL: modelCfg.BaseURL,
			}
			model, err := ollama.NewChatModel(ctx, ollamaCfg)
			if err != nil {
				return nil, fmt.Errorf("model '%s': failed to create Ollama model '%s': %w",
					modelName, modelCfg.Model, err)
			}
			return model, nil

		// case "vllm":
		//  vllmCfg := &vllm.ChatModelConfig{
		//      Model:   modelCfg.Model,
		//      BaseURL: modelCfg.BaseURL,
		//  }
		//  model, err := vllm.NewChatModel(ctx, vllmCfg)
		//  if err != nil {
		//      return nil, fmt.Errorf("model '%s': failed to create vLLM model '%s': %w",
		//          modelName, modelCfg.Model, err)
		//  }
		//  return model, nil

		case "openai":
			openaiCfg := &openai.ChatModelConfig{
				Model:   modelCfg.Model,
				BaseURL: modelCfg.BaseURL,
				APIKey:  modelCfg.APIKey,
			}
			model, err := openai.NewChatModel(ctx, openaiCfg)
			if err != nil {
				return nil, fmt.Errorf("model '%s': failed to create OpenAI model '%s': %w",
					modelName, modelCfg.Model, err)
			}
			return model, nil

		case "lmstudio":
			// LM Studio uses OpenAI compatible API
			lmstudioCfg := &openai.ChatModelConfig{
				Model:   modelCfg.Model,
				BaseURL: modelCfg.BaseURL,
				APIKey:  modelCfg.APIKey,
			}
			model, err := openai.NewChatModel(ctx, lmstudioCfg)
			if err != nil {
				return nil, fmt.Errorf("model '%s': failed to create LM Studio model '%s': %w",
					modelName, modelCfg.Model, err)
			}
			return model, nil

		default:
			// This should not happen due to validation
			return nil, fmt.Errorf("model '%s': unsupported model type '%s' (this should not happen due to validation)",
				modelName, modelCfg.Type)
		}
	}
}

// NewDefaultModelFactoryWithCache creates a cached ModelFactory that supports multiple model types
func NewDefaultModelFactoryWithCache(cfg DefaultModelFactoryConfig, cacheConfig ModelCacheConfig) ModelFactory {
	factory := NewDefaultModelFactory(cfg)
	cache := NewModelCache(cacheConfig)
	return CachedModelFactory(factory, cache)
}

// SimpleModelFactory creates a simple ModelFactory that uses the same model type for all model names with enhanced validation
func SimpleModelFactory(modelType, defaultModel, baseURL string) ModelFactory {
	// Validate model type upfront to avoid repeated checks
	modelTypeLower := strings.ToLower(modelType)
	if !supportedModelTypes[modelTypeLower] {
		supportedList := make([]string, 0, len(supportedModelTypes))
		for t := range supportedModelTypes {
			supportedList = append(supportedList, t)
		}
		// Return a factory that always returns an error for invalid model types
		return func(ctx context.Context, modelName string) (ChatModel, error) {
			return nil, fmt.Errorf("simple model factory: unsupported model type '%s', supported types are: %s",
				modelType, strings.Join(supportedList, ", "))
		}
	}

	// Create a base configuration with defaults
	baseCfg := ModelConfig{
		Type:    modelType,
		Model:   defaultModel,
		BaseURL: baseURL,
		Enabled: true,
	}
	baseCfg = setDefaultModelConfig(baseCfg)

	return func(ctx context.Context, modelName string) (ChatModel, error) {
		if modelName == "" {
			modelName = defaultModel
		}

		if modelName == "" {
			return nil, fmt.Errorf("simple model factory: model name is required")
		}

		// Create model config from base config
		modelCfg := baseCfg
		modelCfg.Model = modelName

		// Validate the configuration
		if err := validateModelConfig(modelCfg, modelName); err != nil {
			return nil, fmt.Errorf("simple model factory: %w", err)
		}

		// Create model based on type with detailed error handling
		switch modelTypeLower {
		case "ollama":
			model, err := ollama.NewChatModel(ctx, &ollama.ChatModelConfig{
				Model:   modelCfg.Model,
				BaseURL: modelCfg.BaseURL,
			})
			if err != nil {
				return nil, fmt.Errorf("simple model factory: failed to create Ollama model '%s': %w",
					modelName, err)
			}
			return model, nil

		case "openai":
			model, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
				Model:   modelCfg.Model,
				BaseURL: modelCfg.BaseURL,
				APIKey:  modelCfg.APIKey,
			})
			if err != nil {
				return nil, fmt.Errorf("simple model factory: failed to create OpenAI model '%s': %w",
					modelName, err)
			}
			return model, nil

		case "lmstudio":
			model, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
				Model:   modelCfg.Model,
				BaseURL: modelCfg.BaseURL,
				APIKey:  modelCfg.APIKey,
			})
			if err != nil {
				return nil, fmt.Errorf("simple model factory: failed to create LM Studio model '%s': %w",
					modelName, err)
			}
			return model, nil

		default:
			// This should not happen due to upfront validation
			return nil, fmt.Errorf("simple model factory: unsupported model type '%s'", modelType)
		}
	}
}

// SimpleModelFactoryWithCache creates a cached simple ModelFactory that uses the same model type for all model names
func SimpleModelFactoryWithCache(modelType, defaultModel, baseURL string, cacheConfig ModelCacheConfig) ModelFactory {
	factory := SimpleModelFactory(modelType, defaultModel, baseURL)
	cache := NewModelCache(cacheConfig)
	return CachedModelFactory(factory, cache)
}

// NewModelFactoryWithConfig creates a ModelFactory with a specific ModelConfig with enhanced validation
func NewModelFactoryWithConfig(cfg ModelConfig) ModelFactory {
	// Pre-validate and preprocess the base configuration
	baseCfg := setDefaultModelConfig(cfg)

	return func(ctx context.Context, modelName string) (ChatModel, error) {
		// Create a copy of the base configuration to avoid modifying the original
		modelCfg := baseCfg
		if modelName != "" {
			modelCfg.Model = modelName
		}

		// Validate the configuration with model name context
		if err := validateModelConfig(modelCfg, modelName); err != nil {
			return nil, fmt.Errorf("model factory with config: %w", err)
		}

		// Create model based on type with detailed error handling
		switch strings.ToLower(modelCfg.Type) {
		case "ollama":
			model, err := ollama.NewChatModel(ctx, &ollama.ChatModelConfig{
				Model:   modelCfg.Model,
				BaseURL: modelCfg.BaseURL,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to create Ollama model '%s': %w",
					modelCfg.Model, err)
			}
			return model, nil

		case "openai":
			model, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
				Model:   modelCfg.Model,
				BaseURL: modelCfg.BaseURL,
				APIKey:  modelCfg.APIKey,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to create OpenAI model '%s': %w",
					modelCfg.Model, err)
			}
			return model, nil

		case "lmstudio":
			model, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
				Model:   modelCfg.Model,
				BaseURL: modelCfg.BaseURL,
				APIKey:  modelCfg.APIKey,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to create LM Studio model '%s': %w",
					modelCfg.Model, err)
			}
			return model, nil

		default:
			// This should not happen due to validation
			return nil, fmt.Errorf("unsupported model type '%s' (this should not happen due to validation)",
				modelCfg.Type)
		}
	}
}

// NewModelFactoryWithConfigAndCache creates a cached ModelFactory with a specific ModelConfig
func NewModelFactoryWithConfigAndCache(cfg ModelConfig, cacheConfig ModelCacheConfig) ModelFactory {
	factory := NewModelFactoryWithConfig(cfg)
	cache := NewModelCache(cacheConfig)
	return CachedModelFactory(factory, cache)
}
