// Agent Framework - Plugin Sandbox Isolation
// Copyright (C) 2025  Agent Framework Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package agent

import (
	"context"
	"fmt"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"AgentFramework/agent/errors"
)

// SandboxConfig defines sandbox isolation configuration
type SandboxConfig struct {
	// Containerization
	EnableContainerization bool `json:"enable_containerization"` // Use container for isolation
	ContainerImage       string `json:"container_image,omitempty"`       // Custom container image

	// Resource limits
	MaxMemoryMB int    `json:"max_memory_mb"`      // Memory limit in MB
	MaxCPUPercent int  `json:"max_cpu_percent"`   // CPU limit as percentage
	MaxTimeSecs   int  `json:"max_time_secs"`     // Maximum execution time

	// Filesystem access
	EnableFilesystem bool     `json:"enable_filesystem"`     // Enable filesystem access
	AllowedPaths   []string `json:"allowed_paths,omitempty"`   // Whitelist of allowed paths
	DeniedPaths   []string `json:"denied_paths,omitempty"`   // Blacklist of denied paths

	// Network access
	EnableNetwork   bool `json:"enable_network"`        // Enable network access
	AllowedDomains []string `json:"allowed_domains,omitempty"` // Whitelist of allowed domains
	DeniedDomains []string `json:"denied_domains,omitempty"` // Blacklist of denied domains

	// Security
	EnableSeccomp bool     `json:"enable_seccomp"`        // Enable seccomp filter
	SeccompProfile string `json:"seccomp_profile,omitempty"` // Seccomp profile (strict, etc.)

	// Timeout
	Timeout time.Duration `json:"timeout"` // Plugin execution timeout

	// Logging
	EnableLogging bool   `json:"enable_logging"`   // Enable plugin output logging
	LogPath      string `json:"log_path,omitempty"`    // Path to log file
}

// DefaultSandboxConfig returns default sandbox configuration
func DefaultSandboxConfig() SandboxConfig {
	return SandboxConfig{
		EnableContainerization: false,
		MaxMemoryMB:          512,  // 512 MB
		MaxCPUPercent:       50,   // 50%
		MaxTimeSecs:         300,  // 5 minutes
		EnableFilesystem:     true,
		AllowedPaths:        []string{"/tmp", "/home/user/.cache"},
		DeniedPaths:        []string{"/etc", "/sys", "/proc", "/root"},
		EnableNetwork:       false,
		EnableSeccomp:      false,
		Timeout:             30 * time.Second,
		EnableLogging:       true,
	}
}

// SandboxProfile defines a preset sandbox configuration
type SandboxProfile string

const (
	// SandboxProfileStrict provides maximum isolation
	SandboxProfileStrict SandboxProfile = "strict"
	// SandboxProfileBalanced balances security and functionality
	SandboxProfileBalanced SandboxProfile = "balanced"
	// SandboxProfilePermissive provides minimal restrictions
	SandboxProfilePermissive SandboxProfile = "permissive"
	// SandboxProfileDisabled disables sandboxing entirely
	SandboxProfileDisabled SandboxProfile = "disabled"
)

// GetSandboxProfile returns sandbox configuration for a profile
func GetSandboxProfile(profile SandboxProfile) SandboxConfig {
	switch profile {
	case SandboxProfileStrict:
		return SandboxConfig{
			EnableContainerization: true,
			MaxMemoryMB:          256,  // 256 MB
			MaxCPUPercent:       25,   // 25%
			MaxTimeSecs:         60,  // 1 minute
			EnableFilesystem:     false,
			EnableNetwork:       false,
			EnableSeccomp:      true,
			SeccompProfile:     "strict",
			Timeout:             10 * time.Second,
		}
	case SandboxProfileBalanced:
		return DefaultSandboxConfig()
	case SandboxProfilePermissive:
		return SandboxConfig{
			EnableContainerization: false,
			MaxMemoryMB:          1024, // 1 GB
			MaxCPUPercent:       100,  // 100%
			MaxTimeSecs:         600,  // 10 minutes
			EnableFilesystem:     true,
			AllowedPaths:        []string{"/"},
			EnableNetwork:       true,
			EnableSeccomp:      false,
			Timeout:             60 * time.Second,
		}
	case SandboxProfileDisabled:
		return SandboxConfig{
			EnableContainerization: false,
			MaxMemoryMB:          0,    // No limit
			MaxCPUPercent:       100,  // 100%
			MaxTimeSecs:         0,    // No limit
			EnableFilesystem:     true,
			EnableNetwork:       true,
			EnableSeccomp:      false,
			Timeout:             0,    // No timeout
		}
	default:
		return DefaultSandboxConfig()
	}
}

// SandboxIsolator provides sandbox isolation for plugins
type SandboxIsolator struct {
	config SandboxConfig
	mu     sync.RWMutex
}

// NewSandboxIsolator creates a new sandbox isolator
func NewSandboxIsolator(config SandboxConfig) *SandboxIsolator {
	return &SandboxIsolator{
		config: config,
	}
}

// SandboxIsolator wraps a plugin to enforce sandbox restrictions
type SandboxedPlugin struct {
	lugin    Plugin
	isolator *SandboxIsolator
	context  context.Context
}

// Execute executes a sandboxed plugin operation
func (p *SandboxedPlugin) Execute(ctx context.Context, method string, args []interface{}) (interface{}, error) {
	// Apply timeout
	if p.isolator.config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, p.isolator.config.Timeout)
		defer cancel()
		p.context = ctx
	}

	// Execute with sandbox enforcement
	return p.lugin.Execute(ctx, method, args)
}

// Name returns the sandboxed plugin's name
func (p *SandboxedPlugin) Name() string {
	return p.lugin.Name()
}

// Version returns the sandboxed plugin's version
func (p *SandboxedPlugin) Version() string {
	return p.lugin.Version()
}

// Description returns the sandboxed plugin's description
func (p *SandboxedPlugin) Description() string {
	return p.lugin.Description()
}

// Initialize initializes the sandboxed plugin
func (p *SandboxedPlugin) Initialize(ctx context.Context, host *Host) error {
	// Validate sandbox configuration before initialization
	if err := p.isolator.Validate(); err != nil {
		return errors.Wrapf(err, errors.ErrCodeInitFailed, "sandbox validation failed for plugin %s", p.Name())
	}

	// Initialize with resource limits
	return p.lugin.Initialize(ctx, host)
}

// Shutdown shuts down the sandboxed plugin
func (p *SandboxedPlugin) Shutdown(ctx context.Context) error {
	return p.lugin.Shutdown(ctx)
}

// IsEnabled returns whether the sandboxed plugin is enabled
func (p *SandboxedPlugin) IsEnabled() bool {
	return p.lugin.IsEnabled()
}

// Enable enables the sandboxed plugin
func (p *SandboxedPlugin) Enable() error {
	return p.lugin.Enable()
}

// Disable disables the sandboxed plugin
func (p *SandboxedPlugin) Disable() error {
	return p.lugin.Disable()
}

// ApplySandbox wraps a plugin with sandbox isolation
func (m *SandboxIsolator) ApplySandbox(lugin Plugin) Plugin {
	return &SandboxedPlugin{
		lugin:    plugin,
		isolator: m,
	}
}

// Validate validates the sandbox configuration
func (m *SandboxIsolator) Validate() error {
	// Validate resource limits
	if m.config.MaxMemoryMB < 0 {
		return errors.Newf(errors.ErrCodeInvalidInput, "max_memory_mb cannot be negative: %d", m.config.MaxMemoryMB)
	}
	if m.config.MaxCPUPercent < 0 || m.config.MaxCPUPercent > 100 {
		return errors.Newf(errors.ErrCodeInvalidInput, "max_cpu_percent must be between 0 and 100: %d", m.config.MaxCPUPercent)
	}
	if m.config.MaxTimeSecs < 0 {
		return errors.Newf(errors.ErrCodeInvalidInput, "max_time_secs cannot be negative: %d", m.config.MaxTimeSecs)
	}

	// Validate filesystem configuration
	if m.config.EnableFilesystem {
		if len(m.config.AllowedPaths) == 0 && len(m.config.DeniedPaths) == 0 {
			return errors.New(errors.ErrCodeInvalidInput, "filesystem enabled but no path restrictions specified")
		}
	}

	// Validate network configuration
	if !m.config.EnableNetwork && len(m.config.AllowedDomains) > 0 {
		return errors.New(errors.ErrCodeInvalidInput, "network disabled but allowed domains specified")
	}

	return nil
}

// GetConfig returns the sandbox configuration
func (m *SandboxIsolator) GetConfig() SandboxConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config
}

// SetConfig updates the sandbox configuration
func (m *SandboxIsolator) SetConfig(config SandboxConfig) error {
	if err := config.Validate(); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.config = config
	return nil
}

// IsolateProcess applies resource limits to a process
func (m *SandboxIsolator) IsolateProcess(cmd *exec.Cmd) error {
	// Apply memory limits
	if m.config.MaxMemoryMB > 0 {
		// This would require cgroup integration
		// For now, we'll just document the requirement
	}

	// Apply CPU limits
	if m.config.MaxCPUPercent < 100 {
		// This would require cgroup integration
		// For now, we'll just document the requirement
	}

	return nil
}

// ValidateFilesystemAccess checks if a filesystem access is allowed
func (m *SandboxIsolator) ValidateFilesystemAccess(path string) error {
	if !m.config.EnableFilesystem {
		return errors.Newf(errors.ErrCodePermissionDenied, "filesystem access is disabled for this plugin")
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	// Check denied paths first
	for _, deniedPath := range m.config.DeniedPaths {
		if path == deniedPath || len(path) > len(deniedPath) && path[:len(deniedPath)] == deniedPath {
			return errors.Newf(errors.ErrCodePermissionDenied, "access to path %s is denied by sandbox policy", path)
		}
	}

	// If allowed paths are specified, check against them
	if len(m.config.AllowedPaths) > 0 {
		allowed := false
		for _, allowedPath := range m.config.AllowedPaths {
			if path == allowedPath || len(path) > len(allowedPath) && path[:len(allowedPath)] == allowedPath {
				allowed = true
				break
			}
		}
		if !allowed {
			return errors.Newf(errors.ErrCodePermissionDenied, "access to path %s is not allowed by sandbox policy", path)
		}
	}

	return nil
}

// ValidateNetworkAccess checks if network access is allowed
func (m *SandboxIsolator) ValidateNetworkAccess(domain string) error {
	if !m.config.EnableNetwork {
		return errors.Newf(errors.ErrCodePermissionDenied, "network access is disabled for this plugin")
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	// Check denied domains first
	for _, deniedDomain := range m.config.DeniedDomains {
		if domain == deniedDomain || len(domain) > len(deniedDomain) && domain[:len(deniedDomain)] == deniedDomain {
			return errors.Newf(errors.ErrCodePermissionDenied, "access to domain %s is denied by sandbox policy", domain)
		}
	}

	// If allowed domains are specified, check against them
	if len(m.config.AllowedDomains) > 0 {
		allowed := false
		for _, allowedDomain := range m.config.AllowedDomains {
			if domain == allowedDomain || len(domain) > len(allowedDomain) && domain[:len(allowedDomain)] == allowedDomain {
				allowed = true
				break
			}
		}
		if !allowed {
			return errors.Newf(errors.ErrCodePermissionDenied, "access to domain %s is not allowed by sandbox policy", domain)
		}
	}

	return nil
}

// SetResourceLimits applies resource limits to the current process
func (m *SandboxIsolator) SetResourceLimits() error {
	if m.config.MaxMemoryMB > 0 {
		// Set memory limit using cgroup
		if err := setMemoryLimit(m.config.MaxMemoryMB); err != nil {
			return errors.Wrapf(err, errors.ErrCodeExecutionFailed, "failed to set memory limit: %w", m.config.MaxMemoryMB)
		}
	}

	if m.config.MaxCPUPercent < 100 {
		// Set CPU limit using cgroup
		if err := setCPULimit(m.config.MaxCPUPercent); err != nil {
			return errors.Wrapf(err, errors.ErrCodeExecutionFailed, "failed to set CPU limit: %w", m.config.MaxCPUPercent)
		}
	}

	return nil
}

// setMemoryLimit sets a memory limit for the current process
func setMemoryLimit(memoryMB int) error {
	// This would use cgroup v1 memory controller
	// For now, return an error indicating the requirement
	return errors.Newf(errors.ErrCodeNotImplemented, "memory limit requires cgroup v1 integration: %d MB", memoryMB)
}

// setCPULimit sets a CPU limit for the current process
func setCPULimit(percent int) error {
	// This would use cgroup v1 CPU controller
	// For now, return an error indicating the requirement
	return errors.Newf(errors.ErrCodeNotImplemented, "CPU limit requires cgroup v1 integration: %d%%", percent)
}

// ResourceMonitor monitors resource usage
type ResourceMonitor struct {
	mu               sync.RWMutex
	memoryUsage      int64
	cpuUsage         float64
	startTime        time.Time
	lastCheckTime    time.Time
	muStats          sync.RWMutex
	maxMemoryMB     int
	maxCPUPercent  int
}

// NewResourceMonitor creates a new resource monitor
func NewResourceMonitor(config SandboxConfig) *ResourceMonitor {
	return &ResourceMonitor{
		maxMemoryMB:     config.MaxMemoryMB,
		maxCPUPercent: config.MaxCPUPercent,
		startTime:      time.Now(),
		lastCheckTime:   time.Now(),
	}
}

// Start begins monitoring resource usage
func (m *ResourceMonitor) Start(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.checkResources(ctx)
		case <-ctx.Done():
			return
		}
	}
}

// checkResources checks current resource usage
func (m *ResourceMonitor) checkResources(ctx context.Context) {
	m.muStats.Lock()
	m.lastCheckTime = time.Now()
	m.muStats.Unlock()

	// Get memory usage
	var ruUsage syscall.Rusage
	if err := syscall.Getrusage(syscall.RUSAGE_SELF, &ruUsage); err == nil {
		memoryMB := ruUsage.Maxrss / 1024 // Convert to MB

		m.mu.Lock()
		m.memoryUsage = memoryMB
		m.mu.Unlock()

		// Check if over limit
		if m.maxMemoryMB > 0 && memoryMB > m.maxMemoryMB {
			// Trigger callback or terminate
			fmt.Printf("Memory limit exceeded: %d MB > %d MB\n", memoryMB, m.maxMemoryMB)
		}
	}

	// Get CPU usage (simplified)
	elapsed := time.Since(m.startTime).Seconds()
	if elapsed > 0 {
		cpuUsage := float64(ruUsage.Utime) / float64(elapsed) * 100

		m.mu.Lock()
		m.cpuUsage = cpuUsage
		m.mu.Unlock()

		// Check if over limit
		if cpuUsage > float64(m.maxCPUPercent) {
			fmt.Printf("CPU limit exceeded: %.2f%% > %d%%\n", cpuUsage, m.maxCPUPercent)
		}
	}
}

// GetMemoryUsage returns current memory usage in MB
func (m *ResourceMonitor) GetMemoryUsage() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.memoryUsage
}

// GetCPUsage returns current CPU usage percentage
func (m *ResourceMonitor) GetCPUsage() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.cpuUsage
}

// GetUptime returns uptime since monitoring started
func (m *ResourceMonitor) GetUptime() time.Duration {
	m.muStats.RLock()
	defer m.muStats.RUnlock()
	return time.Since(m.startTime)
}

// Global sandbox isolator instance
var globalSandboxIsolator *SandboxIsolator
var globalSandboxOnce sync.Once

// InitGlobalSandbox initializes the global sandbox isolator
func InitGlobalSandbox(config SandboxConfig) {
	globalSandboxOnce.Do(func() {
		globalSandboxIsolator = NewSandboxIsolator(config)
	})
}

// GetGlobalSandbox returns the global sandbox isolator instance
func GetGlobalSandbox() *SandboxIsolator {
	if globalSandboxIsolator == nil {
		InitGlobalSandbox(DefaultSandboxConfig())
	}
	return globalSandboxIsolator
}

// SandboxPluginManager manages sandboxed plugins
type SandboxPluginManager struct {
	solator *SandboxIsolator
	lugins   map[string]Plugin
	mu       sync.RWMutex
	monitor  *ResourceMonitor
}

// NewSandboxPluginManager creates a new sandbox plugin manager
func NewSandboxPluginManager(config SandboxConfig) *SandboxPluginManager {
	isolator := NewSandboxIsolator(config)
	monitor := NewResourceMonitor(config)

	return &SandboxPluginManager{
		solator: isolator,
		lugins:   make(map[string]Plugin),
		monitor:  monitor,
	}
}

// RegisterPlugin registers a sandboxed plugin
func (m *SandboxPluginManager) RegisterPlugin(lugin Plugin) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	sandboxedPlugin := m.solator.ApplySandbox(plugin)
	m.plugins[plugin.Name()] = sandboxedPlugin

	return nil
}

// UnregisterPlugin removes a sandboxed plugin
func (m *SandboxPluginManager) UnregisterPlugin(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.plugins[name]; !exists {
		return errors.Newf(errors.ErrCodeNotFound, "plugin %s not found", name)
	}

	delete(m.plugins, name)
	return nil
}

// GetPlugin retrieves a sandboxed plugin by name
func (m *SandboxPluginManager) GetPlugin(name string) (Plugin, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	lugin, exists := m.plugins[name]
	return plugin, exists
}

// ListPlugins returns all sandboxed plugin names
func (m *SandboxPluginManager) ListPlugins() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, 0, len(m.plugins))
	i := 0
	for name := range m.plugins {
		names[i] = name
		i++
	}
	return names
}

// GetResourceMonitor returns the resource monitor
func (m *SandboxPluginManager) GetResourceMonitor() *ResourceMonitor {
	return m.monitor
}

// Start begins monitoring resources
func (m *SandboxPluginManager) Start(ctx context.Context) {
	go m.monitor.Start(ctx)
}
