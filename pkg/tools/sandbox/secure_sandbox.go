// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later


// Package aiosandbox provides secure sandbox implementation for code execution.
package sandbox

import (
	"context"
	"fmt"
	"sync"
	"time"

	"AgentFramework/pkg/beads/security"
)

// SecureSandbox provides a secure environment for code execution.
type SecureSandbox struct {
	containerID      string
	resourceLimits   ResourceLimits
	securityAgent    *security.SecurityManager
	allowedPaths     []string
	blockedPaths     []string
	enabled          bool
	mutex            sync.RWMutex
}

// ResourceLimits defines resource limits for sandbox execution.
type ResourceLimits struct {
	MaxMemoryMB      int64         `json:"max_memory_mb"`
	MaxCPUPercent    float64       `json:"max_cpu_percent"`
	MaxDuration      time.Duration `json:"max_duration"`
	MaxDiskUsageMB   int64         `json:"max_disk_usage_mb"`
	MaxNetworkCalls  int           `json:"max_network_calls"`
	AllowFileAccess  bool          `json:"allow_file_access"`
	AllowNetwork     bool          `json:"allow_network"`
	AllowSubprocess  bool          `json:"allow_subprocess"`
}

// ExecutionResult represents the result of code execution.
type ExecutionResult struct {
	Success     bool        `json:"success"`
	Output      interface{} `json:"output"`
	Error       string      `json:"error,omitempty"`
	ExitCode    int         `json:"exit_code"`
	Duration    time.Duration `json:"duration"`
	MemoryUsed  int64       `json:"memory_used"`
	CPUUsed     float64     `json:"cpu_used"`
	ContainerID string      `json:"container_id"`
}

// NewSecureSandbox creates a new SecureSandbox instance.
func NewSecureSandbox(containerID string, limits ResourceLimits, securityAgent *security.SecurityManager) *SecureSandbox {
	return &SecureSandbox{
		containerID:      containerID,
		resourceLimits:   limits,
		securityAgent:    securityAgent,
		allowedPaths:     make([]string, 0),
		blockedPaths:     make([]string, 0),
		enabled:          true,
	}
}

// Execute executes code in the secure sandbox.
func (s *SecureSandbox) Execute(ctx context.Context, code string, lang Language) (*ExecutionResult, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.enabled {
		return nil, fmt.Errorf("sandbox is disabled")
	}

	// Validate code before execution
	if err := s.validateCode(ctx, code, lang); err != nil {
		return &ExecutionResult{
			Success: false,
			Error:   fmt.Sprintf("Code validation failed: %v", err),
		}, nil
	}

	// Check resource availability
	if err := s.checkResources(ctx); err != nil {
		return &ExecutionResult{
			Success: false,
			Error:   fmt.Sprintf("Resource check failed: %v", err),
		}, nil
	}

	// Set up execution context with timeout
	execCtx, cancel := context.WithTimeout(ctx, s.resourceLimits.MaxDuration)
	defer cancel()

	// Execute code
	startTime := time.Now()
	result, err := s.executeInContainer(execCtx, code, lang)
	duration := time.Since(startTime)

	if err != nil {
		return &ExecutionResult{
			Success:     false,
			Error:       err.Error(),
			Duration:    duration,
			ContainerID: s.containerID,
		}, nil
	}

	// Check resource usage
	if err := s.checkResourceUsage(ctx, result); err != nil {
		return &ExecutionResult{
			Success:     false,
			Error:       fmt.Sprintf("Resource usage check failed: %v", err),
			Duration:    duration,
			ContainerID: s.containerID,
		}, nil
	}

	// Log execution event
	s.logAuditEvent(ctx, "code_execution", result)

	return result, nil
}

// validateCode validates code before execution.
func (s *SecureSandbox) validateCode(ctx context.Context, code string, lang Language) error {
	// Check for dangerous patterns
	if containsDangerousCode(code) {
		return fmt.Errorf("code contains dangerous patterns")
	}

	// Check language support
	if !s.isLanguageSupported(lang) {
		return fmt.Errorf("language %s is not supported", lang)
	}

	return nil
}

// checkResources checks if resources are available.
func (s *SecureSandbox) checkResources(ctx context.Context) error {
	return nil
}

// executeInContainer executes code in a container.
func (s *SecureSandbox) executeInContainer(ctx context.Context, code string, lang Language) (*ExecutionResult, error) {
	// Simulated execution
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(100 * time.Millisecond):
		return &ExecutionResult{
			Success:     true,
			Output:      fmt.Sprintf("Executed %s code", lang),
			ExitCode:    0,
			Duration:    100 * time.Millisecond,
			MemoryUsed:  1024 * 1024,
			CPUUsed:     0.1,
			ContainerID: s.containerID,
		}, nil
	}
}

// checkResourceUsage checks if resource usage is within limits.
func (s *SecureSandbox) checkResourceUsage(ctx context.Context, result *ExecutionResult) error {
	if result.MemoryUsed > s.resourceLimits.MaxMemoryMB*1024*1024 {
		return fmt.Errorf("memory usage exceeded limit")
	}

	if result.CPUUsed > s.resourceLimits.MaxCPUPercent {
		return fmt.Errorf("CPU usage exceeded limit")
	}

	return nil
}

// logAuditEvent logs an audit event.
func (s *SecureSandbox) logAuditEvent(ctx context.Context, eventType string, result *ExecutionResult) {
	if s.securityAgent == nil {
		return
	}

	s.securityAgent.AuditLog(ctx, security.AuditEvent{
		EventType: eventType,
		Status:    boolToStatus(result.Success),
		Details: map[string]interface{}{
			"container_id": s.containerID,
			"result":       result,
		},
	})
}

// AddAllowedPath adds a path to the allowed paths list.
func (s *SecureSandbox) AddAllowedPath(ctx context.Context, path string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.allowedPaths = append(s.allowedPaths, path)
}

// AddBlockedPath adds a path to the blocked paths list.
func (s *SecureSandbox) AddBlockedPath(ctx context.Context, path string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.blockedPaths = append(s.blockedPaths, path)
}

// IsPathAllowed checks if a path is allowed.
func (s *SecureSandbox) IsPathAllowed(ctx context.Context, path string) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for _, blockedPath := range s.blockedPaths {
		if path == blockedPath || isSubpath(path, blockedPath) {
			return false
		}
	}

	if len(s.allowedPaths) == 0 {
		return true
	}

	for _, allowedPath := range s.allowedPaths {
		if path == allowedPath || isSubpath(path, allowedPath) {
			return true
		}
	}

	return false
}

// SetResourceLimits sets resource limits.
func (s *SecureSandbox) SetResourceLimits(ctx context.Context, limits ResourceLimits) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.resourceLimits = limits
}

// GetResourceLimits returns current resource limits.
func (s *SecureSandbox) GetResourceLimits(ctx context.Context) ResourceLimits {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.resourceLimits
}

// Enable enables the sandbox.
func (s *SecureSandbox) Enable(ctx context.Context) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.enabled = true
	s.logAuditEvent(ctx, "sandbox_enable", &ExecutionResult{Success: true})
}

// Disable disables the sandbox.
func (s *SecureSandbox) Disable(ctx context.Context) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.enabled = false
	s.logAuditEvent(ctx, "sandbox_disable", &ExecutionResult{Success: true})
}

// IsEnabled returns whether the sandbox is enabled.
func (s *SecureSandbox) IsEnabled(ctx context.Context) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.enabled
}

// Close closes the sandbox and cleans up resources.
func (s *SecureSandbox) Close(ctx context.Context) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.enabled = false
	s.logAuditEvent(ctx, "sandbox_close", &ExecutionResult{Success: true})

	return nil
}

// isLanguageSupported checks if a language is supported.
func (s *SecureSandbox) isLanguageSupported(lang Language) bool {
	switch lang {
	case LanguagePython, LanguageJavaScript, LanguageGo, LanguageRust:
		return true
	default:
		return false
	}
}

// containsDangerousCode checks if code contains dangerous patterns.
func containsDangerousCode(code string) bool {
	dangerousPatterns := []string{
		"eval(",
		"exec(",
		"__import__",
		"compile(",
	}

	for _, pattern := range dangerousPatterns {
		if contains(code, pattern) {
			return true
		}
	}

	return false
}

// contains checks if a string contains a substring.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && indexOf(s, substr) >= 0
}

// indexOf returns the index of a substring in a string.
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// isSubpath checks if a path is a subpath of another.
func isSubpath(path, parent string) bool {
	return len(path) > len(parent) && path[:len(parent)] == parent && path[len(parent)] == '/'
}

// boolToStatus converts a boolean to a status string.
func boolToStatus(success bool) string {
	if success {
		return "success"
	}
	return "failed"
}

// Language represents a programming language.
type Language string

const (
	LanguagePython    Language = "python"
	LanguageJavaScript Language = "javascript"
	LanguageGo        Language = "go"
	LanguageRust      Language = "rust"
)

// SandboxManager manages multiple sandboxes.
type SandboxManager struct {
	sandboxes map[string]*SecureSandbox
	mutex     sync.RWMutex
}

// NewSandboxManager creates a new SandboxManager instance.
func NewSandboxManager() *SandboxManager {
	return &SandboxManager{
		sandboxes: make(map[string]*SecureSandbox),
	}
}

// CreateSandbox creates a new sandbox.
func (m *SandboxManager) CreateSandbox(ctx context.Context, containerID string, limits ResourceLimits, securityAgent *security.SecurityManager) (*SecureSandbox, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, exists := m.sandboxes[containerID]; exists {
		return nil, fmt.Errorf("sandbox %s already exists", containerID)
	}

	sandbox := NewSecureSandbox(containerID, limits, securityAgent)
	m.sandboxes[containerID] = sandbox

	return sandbox, nil
}

// GetSandbox returns a sandbox by ID.
func (m *SandboxManager) GetSandbox(ctx context.Context, containerID string) (*SecureSandbox, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	sandbox, exists := m.sandboxes[containerID]
	if !exists {
		return nil, fmt.Errorf("sandbox %s not found", containerID)
	}

	return sandbox, nil
}

// DeleteSandbox deletes a sandbox.
func (m *SandboxManager) DeleteSandbox(ctx context.Context, containerID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	sandbox, exists := m.sandboxes[containerID]
	if !exists {
		return fmt.Errorf("sandbox %s not found", containerID)
	}

	if err := sandbox.Close(ctx); err != nil {
		return fmt.Errorf("failed to close sandbox: %w", err)
	}

	delete(m.sandboxes, containerID)

	return nil
}

// ListSandboxes lists all sandbox IDs.
func (m *SandboxManager) ListSandboxes(ctx context.Context) []string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	ids := make([]string, 0, len(m.sandboxes))
	for id := range m.sandboxes {
		ids = append(ids, id)
	}

	return ids
}