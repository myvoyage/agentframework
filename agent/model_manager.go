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
	"sync"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"

	"AgentFramework/agent/errors"
)

// ModelStatus represents the status of a model
type ModelStatus string

const (
	// ModelStatusIdle indicates the model is idle and ready to use
	ModelStatusIdle ModelStatus = "idle"
	// ModelStatusActive indicates the model is currently in use
	ModelStatusActive ModelStatus = "active"
	// ModelStatusHealthy indicates the model passed health check
	ModelStatusHealthy ModelStatus = "healthy"
	// ModelStatusUnhealthy indicates the model failed health check
	ModelStatusUnhealthy ModelStatus = "unhealthy"
	// ModelStatusLoading indicates the model is being loaded
	ModelStatusLoading ModelStatus = "loading"
	// ModelStatusUnloading indicates the model is being unloaded
	ModelStatusUnloading ModelStatus = "unloading"
)

// ModelInfo contains detailed information about a model
type ModelInfo struct {
	Status            ModelStatus `json:"status"`
	LastHealthCheck   time.Time   `json:"last_health_check"`
	HealthCheckResult string      `json:"health_check_result"`
	RequestCount      int64       `json:"request_count"`
	ErrorCount        int64       `json:"error_count"`
	LastRequest       time.Time   `json:"last_request"`
	CreatedAt         time.Time   `json:"created_at"`
}

// ModelManager manages multiple models and allows for dynamic selection and switching
type ModelManager struct {
	models              map[string]ChatModel
	modelInfo           map[string]*ModelInfo
	current             string
	mutex               sync.RWMutex
	factory             ModelFactory
	cache               ModelCacheInterface
	ctx                 context.Context
	cancel              context.CancelFunc
	healthCheckInterval time.Duration
}

// NewModelManager creates a new ModelManager instance
func NewModelManager(ctx context.Context, mf ModelFactory, initialModels ...string) (*ModelManager, error) {
	return NewModelManagerWithCache(ctx, mf, ModelCacheConfig{}, initialModels...)
}

// NewModelManagerWithCache creates a new ModelManager instance with model caching support
func NewModelManagerWithCache(ctx context.Context, mf ModelFactory, cacheConfig ModelCacheConfig, initialModels ...string) (*ModelManager, error) {
	mgrCtx, cancel := context.WithCancel(ctx)

	// Create model cache if config is provided
	var cache ModelCacheInterface
	if cacheConfig.MaxSize != 0 || cacheConfig.TTL != 0 || cacheConfig.CleanupInterval != 0 {
		cache = NewModelCache(cacheConfig)
	} else {
		// Create cache with default settings
		cache = NewModelCache(ModelCacheConfig{
			MaxSize:         100,
			TTL:             1 * time.Hour,
			CleanupInterval: 10 * time.Minute,
		})
	}

	mgr := &ModelManager{
		models:              make(map[string]ChatModel),
		modelInfo:           make(map[string]*ModelInfo),
		factory:             mf,
		cache:               cache,
		ctx:                 mgrCtx,
		cancel:              cancel,
		healthCheckInterval: 5 * time.Minute, // Default health check interval
	}

	// Initialize with initial models if provided
	for _, modelName := range initialModels {
		if err := mgr.LoadModel(modelName); err != nil {
			cancel()
			return nil, err
		}
		if mgr.current == "" {
			mgr.current = modelName
		}
	}

	// Start health check goroutine
	go mgr.runHealthChecks()

	return mgr, nil
}

// WithHealthCheckInterval sets the health check interval for the model manager
func (m *ModelManager) WithHealthCheckInterval(interval time.Duration) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.healthCheckInterval = interval
}

// runHealthChecks periodically checks the health of all loaded models
func (m *ModelManager) runHealthChecks() {
	ticker := time.NewTicker(m.healthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.checkAllModelsHealth()
		case <-m.ctx.Done():
			return
		}
	}
}

// checkAllModelsHealth checks the health of all loaded models
func (m *ModelManager) checkAllModelsHealth() {
	m.mutex.RLock()
	models := make([]string, 0, len(m.models))
	for name := range m.models {
		models = append(models, name)
	}
	m.mutex.RUnlock()

	for _, modelName := range models {
		m.checkModelHealth(modelName)
	}
}

// checkModelHealth checks the health of a specific model
func (m *ModelManager) checkModelHealth(modelName string) {
	m.mutex.RLock()
	model, ok := m.models[modelName]
	m.mutex.RUnlock()

	if !ok {
		return
	}

	// Simple health check: try to generate a small response
	ctx, cancel := context.WithTimeout(m.ctx, 5*time.Second)
	defer cancel()

	testMessage := &schema.Message{
		Role:    schema.User,
		Content: "ping",
	}

	_, err := model.Generate(ctx, []*schema.Message{testMessage})

	m.mutex.Lock()
	defer m.mutex.Unlock()

	info, ok := m.modelInfo[modelName]
	if !ok {
		return
	}

	info.LastHealthCheck = time.Now()
	if err != nil {
		info.Status = ModelStatusUnhealthy
		info.HealthCheckResult = fmt.Sprintf("Health check failed: %v", err)
	} else {
		info.Status = ModelStatusHealthy
		info.HealthCheckResult = "Health check passed"
	}
}

// LoadModel loads a model into the manager
func (m *ModelManager) LoadModel(name string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Check if model already exists
	if _, ok := m.models[name]; ok {
		return nil
	}

	// Try to get from cache first
	var model ChatModel
	var err error

	if m.cache != nil {
		if cachedModel := m.cache.Get(name); cachedModel != nil {
			model = cachedModel
		} else {
			// Create new model if not in cache
			model, err = m.factory(m.ctx, name)
			if err != nil {
				return err
			}
			// Cache the model
			m.cache.Put(name, model)
		}
	} else {
		// No cache, create new model
		model, err = m.factory(m.ctx, name)
		if err != nil {
			return err
		}
	}

	// Add to models map
	m.models[name] = model

	// Initialize model info
	m.modelInfo[name] = &ModelInfo{
		Status:            ModelStatusHealthy,
		LastHealthCheck:   time.Now(),
		HealthCheckResult: "Model loaded successfully",
		RequestCount:      0,
		ErrorCount:        0,
		LastRequest:       time.Time{},
		CreatedAt:         time.Now(),
	}

	return nil
}

// UnloadModel removes a model from the manager
func (m *ModelManager) UnloadModel(name string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Check if model exists
	if _, ok := m.models[name]; !ok {
		return nil // Model not found, nothing to do
	}

	// Update model status before unloading
	if info, ok := m.modelInfo[name]; ok {
		info.Status = ModelStatusUnloading
	}

	// Remove model
	delete(m.models, name)
	delete(m.modelInfo, name)

	// If current model was removed, switch to another model
	if m.current == name {
		for modelName := range m.models {
			m.current = modelName
			break
		}
		if m.current == name { // No models left
			m.current = ""
		}
	}

	return nil
}

// SwitchModel switches the current active model
func (m *ModelManager) SwitchModel(name string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Check if model exists
	if _, ok := m.models[name]; !ok {
		// Try to get from cache first
		var model ChatModel
		var err error

		if m.cache != nil {
			if cachedModel := m.cache.Get(name); cachedModel != nil {
				model = cachedModel
			} else {
				// Create new model if not in cache
				model, err = m.factory(m.ctx, name)
				if err != nil {
					return err
				}
				// Cache the model
				m.cache.Put(name, model)
			}
		} else {
			// No cache, create new model
			model, err = m.factory(m.ctx, name)
			if err != nil {
				return err
			}
		}

		// Add to models map
		m.models[name] = model

		// Initialize model info if not exists
		if _, ok := m.modelInfo[name]; !ok {
			m.modelInfo[name] = &ModelInfo{
				Status:            ModelStatusHealthy,
				LastHealthCheck:   time.Now(),
				HealthCheckResult: "Model loaded successfully",
				RequestCount:      0,
				ErrorCount:        0,
				LastRequest:       time.Time{},
				CreatedAt:         time.Now(),
			}
		}
	}

	// Update current model status
	if currentName := m.current; currentName != "" {
		if info, ok := m.modelInfo[currentName]; ok {
			info.Status = ModelStatusIdle
		}
	}

	// Update new current model status
	if info, ok := m.modelInfo[name]; ok {
		info.Status = ModelStatusActive
	}

	// Switch to the model
	m.current = name
	return nil
}

// CurrentModel returns the current active model
func (m *ModelManager) CurrentModel() (ChatModel, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if m.current == "" {
		return nil, ErrNoCurrentModel
	}

	model, ok := m.models[m.current]
	if !ok {
		return nil, ErrNoCurrentModel
	}

	return model, nil
}

// GetModel returns a specific model by name
func (m *ModelManager) GetModel(name string) (ChatModel, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	model, ok := m.models[name]
	if !ok {
		return nil, ErrModelNotFound
	}

	return model, nil
}

// ListModels returns a list of all loaded models
func (m *ModelManager) ListModels() []string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	models := make([]string, 0, len(m.models))
	for name := range m.models {
		models = append(models, name)
	}

	return models
}

// Generate uses the current model to generate a response
func (m *ModelManager) Generate(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	model, err := m.CurrentModel()
	if err != nil {
		return nil, err
	}

	// Get current model name for tracking
	m.mutex.RLock()
	modelName := m.current
	m.mutex.RUnlock()

	// Update model info before request
	m.mutex.Lock()
	if info, ok := m.modelInfo[modelName]; ok {
		info.RequestCount++
		info.LastRequest = time.Now()
	}
	m.mutex.Unlock()

	// Execute request
	resp, err := model.Generate(ctx, messages, opts...)

	// Update model info after request
	m.mutex.Lock()
	if info, ok := m.modelInfo[modelName]; ok {
		if err != nil {
			info.ErrorCount++
		}
	}
	m.mutex.Unlock()

	return resp, err
}

// Stream uses the current model to generate a streaming response
func (m *ModelManager) Stream(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	model, err := m.CurrentModel()
	if err != nil {
		return nil, err
	}

	// Get current model name for tracking
	m.mutex.RLock()
	modelName := m.current
	m.mutex.RUnlock()

	// Update model info before request
	m.mutex.Lock()
	if info, ok := m.modelInfo[modelName]; ok {
		info.RequestCount++
		info.LastRequest = time.Now()
	}
	m.mutex.Unlock()

	// Execute request
	resp, err := model.Stream(ctx, messages, opts...)

	// Update model info after request
	m.mutex.Lock()
	if info, ok := m.modelInfo[modelName]; ok {
		if err != nil {
			info.ErrorCount++
		}
	}
	m.mutex.Unlock()

	return resp, err
}

// ModelSelector is an interface for selecting models based on input or context
type ModelSelector interface {
	SelectModel(ctx context.Context, input string) (string, error)
}

// GetModelInfo returns detailed information about a specific model
func (m *ModelManager) GetModelInfo(name string) (*ModelInfo, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	info, ok := m.modelInfo[name]
	if !ok {
		return nil, ErrModelNotFound
	}

	// Return a copy to avoid external modification
	infoCopy := *info
	return &infoCopy, nil
}

// ListModelInfo returns detailed information about all loaded models
func (m *ModelManager) ListModelInfo() map[string]*ModelInfo {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Return copies to avoid external modification
	result := make(map[string]*ModelInfo, len(m.modelInfo))
	for name, info := range m.modelInfo {
		infoCopy := *info
		result[name] = &infoCopy
	}

	return result
}

// ManualCheckModelHealth manually checks the health of a specific model
func (m *ModelManager) ManualCheckModelHealth(name string) error {
	m.mutex.RLock()
	_, ok := m.models[name]
	m.mutex.RUnlock()

	if !ok {
		return ErrModelNotFound
	}

	m.checkModelHealth(name)

	// Return health status
	info, err := m.GetModelInfo(name)
	if err != nil {
		return err
	}

	if info.Status == ModelStatusUnhealthy {
		return errors.Newf(errors.ErrCodeModelExecution, "model %s is unhealthy: %s", name, info.HealthCheckResult)
	}

	return nil
}

// Close closes the ModelManager and releases all resources
func (m *ModelManager) Close() error {
	m.cancel()

	// Stop cache cleanup
	if m.cache != nil {
		m.cache.StopCleanup()
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Clear all models and model info
	m.models = make(map[string]ChatModel)
	m.modelInfo = make(map[string]*ModelInfo)
	m.current = ""

	return nil
}

// PreheatModels preloads models into the cache to improve initial hit rates and reduce cold start latency
// It accepts a list of model names and optional preheat options
func (m *ModelManager) PreheatModels(ctx context.Context, modelNames []string, opts ...PreheatCacheOption) error {
	// If no cache is available, we can't preheat
	if m.cache == nil {
		return nil
	}

	// Get the concrete ModelCache instance
	cache, ok := m.cache.(*ModelCache)
	if !ok {
		return fmt.Errorf("unsupported cache type for preheating")
	}

	// Use the existing PreheatCache function to preheat the models
	return PreheatCache(ctx, m.factory, cache, modelNames, opts...)
}

// SimpleModelSelector is a simple model selector that always returns the same model
type SimpleModelSelector struct {
	modelName string
}

// NewSimpleModelSelector creates a new SimpleModelSelector instance
func NewSimpleModelSelector(modelName string) *SimpleModelSelector {
	return &SimpleModelSelector{
		modelName: modelName,
	}
}

// SelectModel returns the configured model name
func (s *SimpleModelSelector) SelectModel(ctx context.Context, input string) (string, error) {
	return s.modelName, nil
}

// ContextualModelSelector selects models based on context or input content
type ContextualModelSelector struct {
	modelMap     map[string][]string // context keywords to model names
	defaultModel string
}

// NewContextualModelSelector creates a new ContextualModelSelector instance
func NewContextualModelSelector(modelMap map[string][]string, defaultModel string) *ContextualModelSelector {
	return &ContextualModelSelector{
		modelMap:     modelMap,
		defaultModel: defaultModel,
	}
}

// SelectModel selects a model based on input content
func (s *ContextualModelSelector) SelectModel(ctx context.Context, input string) (string, error) {
	// Simple keyword matching for demo purposes
	// In real implementation, this could use embeddings or more sophisticated matching
	for keyword, models := range s.modelMap {
		if containsKeyword(input, keyword) {
			return models[0], nil // Return first matching model
		}
	}

	return s.defaultModel, nil
}

// containsKeyword checks if input contains a keyword (case-insensitive)
func containsKeyword(input, keyword string) bool {
	// Simple implementation for demo
	// In real implementation, use proper string matching
	return len(input) >= len(keyword) &&
		containsSubstringIgnoreCase(input, keyword)
}

// containsSubstringIgnoreCase checks if a string contains a substring (case-insensitive)
func containsSubstringIgnoreCase(s, substr string) bool {
	// Simple implementation for demo
	// In real implementation, use strings.ToLower or similar
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if toLower(s[i+j]) != toLower(substr[j]) {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

// toLower converts a character to lowercase
func toLower(c byte) byte {
	if c >= 'A' && c <= 'Z' {
		return c + ('a' - 'A')
	}
	return c
}

// Common errors for ModelManager
var (
	ErrNoCurrentModel = errors.New(errors.ErrCodeNotFound, "no current model selected")
	ErrModelNotFound  = errors.New(errors.ErrCodeModelNotFound, "model not found")
)
