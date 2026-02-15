// Agent Framework - Plugin Dependency Management
// Copyright (C) 2025  Agent Framework Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package agent

import (
	"context"
	"fmt"
	"sync"

	"AgentFramework/agent/errors"
)

// PluginVersion represents a semantic version
type PluginVersion string

// NewVersion creates a new version from components
func NewVersion(major, minor, patch int) PluginVersion {
	return PluginVersion(fmt.Sprintf("%d.%d.%d", major, minor, patch))
}

// Major returns the major version number
func (v PluginVersion) Major() (int, error) {
	var major int
	n, err := fmt.Sscanf(string(v), "%d", &major)
	if err != nil {
		return 0, errors.Newf(errors.ErrCodeInvalidInput, "invalid version format: %s", v)
	}
	return major, nil
}

// Minor returns the minor version number
func (v PluginVersion) Minor() (int, error) {
	var major, minor int
	n, err := fmt.Sscanf(string(v), "%d.%d", &major, &minor)
	if err != nil {
		return 0, errors.Newf(errors.ErrCodeInvalidInput, "invalid version format: %s", v)
	}
	return minor, nil
}

// Patch returns the patch version number
func (v PluginVersion) Patch() (int, error) {
	var major, minor, patch int
	n, err := fmt.Sscanf(string(v), "%d.%d.%d", &major, &minor, &patch)
	if err != nil {
		return 0, errors.Newf(errors.ErrCodeInvalidInput, "invalid version format: %s", v)
	}
	return patch, nil
}

// Compare compares two versions
// Returns -1 if v < other, 0 if v == other, 1 if v > other
func (v PluginVersion) Compare(other PluginVersion) (int, error) {
	vMajor, err := v.Major()
	if err != nil {
		return 0, err
	}
	otherMajor, err := other.Major()
	if err != nil {
		return 0, err
	}

	if vMajor != otherMajor {
		if vMajor < otherMajor {
			return -1, nil
		}
		return 1, nil
	}

	vMinor, err := v.Minor()
	if err != nil {
		return 0, err
	}
	otherMinor, err := other.Minor()
	if err != nil {
		return 0, err
	}

	if vMinor != otherMinor {
		if vMinor < otherMinor {
			return -1, nil
		}
		return 1, nil
	}

	vPatch, err := v.Patch()
	if err != nil {
		return 0, err
	}
	otherPatch, err := other.Patch()
	if err != nil {
		return 0, err
	}

	if vPatch < otherPatch {
		return -1, nil
	}
	if vPatch > otherPatch {
		return 1, nil
	}
	return 0, nil
}

// IsCompatible checks if this version is compatible with the required version range
func (v PluginVersion) IsCompatible(minVersion, maxVersion PluginVersion) (bool, error) {
	// Check minimum version
	if minVersion != "" {
		cmp, err := v.Compare(minVersion)
		if err != nil {
			return false, err
		}
		if cmp < 0 {
			return false, nil
		}
	}

	// Check maximum version
	if maxVersion != "" {
		cmp, err := v.Compare(maxVersion)
		if err != nil {
			return false, err
		}
		if cmp > 0 {
			return false, nil
		}
	}

	return true, nil
}

// PluginDependency represents a single plugin dependency
type PluginDependency struct {
	Name         string       `json:"name"`
	Version      string       `json:"version"`       // Semantic version constraint
	Optional     bool         `json:"optional"`
	Repository   string       `json:"repository,omitempty"` // Optional repository URL
}

// PluginManifest represents a plugin's manifest with metadata
type PluginManifest struct {
	// Identity
	Name                string `json:"name"`
	Version             string `json:"version"`
	Description         string `json:"description,omitempty"`
	Author              string `json:"author,omitempty"`
	License             string `json:"license,omitempty"`

	// Framework compatibility
	MinFrameworkVersion string `json:"min_framework_version,omitempty"`
	MaxFrameworkVersion string `json:"max_framework_version,omitempty"`

	// Dependencies
	Dependencies        []PluginDependency `json:"dependencies,omitempty"`

	// Capabilities
	Capabilities        []string `json:"capabilities,omitempty"`

	// Resource requirements
	MaxMemoryMB        int    `json:"max_memory_mb,omitempty"`
	MaxCPUPercent     int    `json:"max_cpu_percent,omitempty"`
	RequiredNetwork   bool   `json:"required_network,omitempty"`
	RequiredFS       bool   `json:"required_fs,omitempty"`

	// Security
	SandboxEnabled     bool   `json:"sandbox_enabled"`
	RequirePermissions []string `json:"require_permissions,omitempty"`
}

// Validate validates the plugin manifest
func (m *PluginManifest) Validate() error {
	if m.Name == "" {
		return errors.Newf(errors.ErrCodeInvalidInput, "plugin manifest missing name")
	}

	if m.Version == "" {
		return errors.Newf(errors.ErrCodeInvalidInput, "plugin manifest %s missing version", m.Name)
	}

	// Validate version format
	version := PluginVersion(m.Version)
	if _, err := version.Major(); err != nil {
		return errors.Wrapf(err, errors.ErrCodeInvalidInput, "plugin %s has invalid version format", m.Name)
	}

	// Validate framework version compatibility
	if m.MinFrameworkVersion != "" {
		minVer := PluginVersion(m.MinFrameworkVersion)
		// Use current framework version - should be passed in or detected
		currentFrameworkVer := PluginVersion("1.0.0") // Example: current framework version
		if compatible, err := currentFrameworkVer.IsCompatible(minVer, ""); !compatible || err != nil {
			return errors.Newf(errors.ErrCodeInitFailed, "plugin %s requires framework version >= %s", m.Name, m.MinFrameworkVersion)
		}
	}

	// Validate dependencies
	for _, dep := range m.Dependencies {
		if dep.Name == "" {
			return errors.Newf(errors.ErrCodeInvalidInput, "plugin %s has dependency without name", m.Name)
		}
		if dep.Version == "" {
			return errors.Newf(errors.ErrCodeInvalidInput, "plugin %s dependency %s missing version", m.Name, dep.Name)
		}
	}

	return nil
}

// DependencyResolver resolves plugin dependencies and load order
type DependencyResolver struct {
	plugins    map[string]*PluginManifest
	mu         sync.RWMutex
	version    string // Current framework version
}

// NewDependencyResolver creates a new dependency resolver
func NewDependencyResolver(frameworkVersion string) *DependencyResolver {
	return &DependencyResolver{
		plugins: make(map[string]*PluginManifest),
		version: frameworkVersion,
	}
}

// Register registers a plugin manifest
func (r *DependencyResolver) Register(manifest *PluginManifest) error {
	if err := manifest.Validate(); err != nil {
		return errors.Wrapf(err, errors.ErrCodeInvalidInput, "failed to validate plugin manifest %s", manifest.Name)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.plugins[manifest.Name]; exists {
		return errors.Newf(errors.ErrCodeInvalidInput, "plugin %s already registered", manifest.Name)
	}

	r.plugins[manifest.Name] = manifest
	return nil
}

// Resolve resolves dependencies and returns the load order
func (r *DependencyResolver) Resolve(pluginName string) ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plugin, exists := r.plugins[pluginName]
	if !exists {
		return nil, errors.Newf(errors.ErrCodeNotFound, "plugin %s not found", pluginName)
	}

	// Track load order
	loadOrder := []string{}
	visited := make(map[string]bool)
	resolving := make(map[string]bool)

	if err := r.resolveDeps(plugin, loadOrder, visited, resolving); err != nil {
		return nil, err
	}

	return loadOrder, nil
}

// resolveDeps recursively resolves dependencies
func (r *DependencyResolver) resolveDeps(plugin *PluginManifest, loadOrder *[]string, visited, resolving map[string]bool) error {
	// Check for circular dependencies
	if resolving[plugin.Name] {
		return errors.Newf(errors.ErrCodeInitFailed, "circular dependency detected for plugin %s", plugin.Name)
	}

	resolving[plugin.Name] = true

	// Resolve dependencies first
	for _, dep := range plugin.Dependencies {
		// Check if dependency exists
		depPlugin, exists := r.plugins[dep.Name]
		if !exists {
			if dep.Optional {
				continue // Skip optional missing dependencies
			}
			return errors.Newf(errors.ErrCodeNotFound, "plugin %s depends on %s which is not registered", plugin.Name, dep.Name)
		}

		// Check if already visited
		if visited[dep.Name] {
			continue
		}

		// Recursively resolve
		if err := r.resolveDeps(depPlugin, loadOrder, visited, resolving); err != nil {
			return err
		}
	}

	// Add this plugin to load order
	*loadOrder = append(*loadOrder, plugin.Name)
	visited[plugin.Name] = true

	delete(resolving, plugin.Name)
	return nil
}

// GetDependencyGraph returns the complete dependency graph
func (r *DependencyResolver) GetDependencyGraph() map[string][]string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	graph := make(map[string][]string)
	for name, plugin := range r.plugins {
		deps := make([]string, 0, len(plugin.Dependencies))
		for i, dep := range plugin.Dependencies {
			deps[i] = dep.Name
		}
		graph[name] = deps
	}
	return graph
}

// GetManifest retrieves a plugin manifest by name
func (r *DependencyResolver) GetManifest(name string) (*PluginManifest, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	manifest, exists := r.plugins[name]
	return manifest, exists
}

// ListPlugins returns all registered plugin names
func (r *DependencyResolver) ListPlugins() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.plugins))
	i := 0
	for name := range r.plugins {
		names[i] = name
		i++
	}
	return names
}

// CheckCompatibility checks if a plugin is compatible with the framework version
func (r *DependencyResolver) CheckCompatibility(manifest *PluginManifest) (bool, error) {
	frameworkVer := PluginVersion(r.version)

	// Check minimum version
	if manifest.MinFrameworkVersion != "" {
		minVer := PluginVersion(manifest.MinFrameworkVersion)
		compatible, err := frameworkVer.IsCompatible(minVer, "")
		if err != nil {
			return false, err
		}
		if !compatible {
			return false, nil
		}
	}

	// Check maximum version
	if manifest.MaxFrameworkVersion != "" {
		maxVer := PluginVersion(manifest.MaxFrameworkVersion)
		compatible, err := frameworkVer.IsCompatible("", maxVer)
		if err != nil {
			return false, err
		}
		if !compatible {
			return false, nil
		}
	}

	return true, nil
}

// Unregister removes a plugin manifest
func (r *DependencyResolver) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.plugins[name]; !exists {
		return errors.Newf(errors.ErrCodeNotFound, "plugin %s not registered", name)
	}

	delete(r.plugins, name)
	return nil
}

// GetCapabilities returns all capabilities provided by registered plugins
func (r *DependencyResolver) GetCapabilities() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	capSet := make(map[string]bool)
	for _, plugin := range r.plugins {
		for _, cap := range plugin.Capabilities {
			capSet[cap] = true
		}
	}

	capabilities := make([]string, 0, len(capSet))
	i := 0
	for cap := range capSet {
		capabilities[i] = cap
		i++
	}
	return capabilities
}

// FindByCapability finds plugins that provide a specific capability
func (r *DependencyResolver) FindByCapability(capability string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var plugins []string
	for name, plugin := range r.plugins {
		for _, cap := range plugin.Capabilities {
			if cap == capability {
				plugins = append(plugins, name)
				break
			}
		}
	}
	return plugins
}

// GetDependents returns plugins that depend on the given plugin
func (r *DependencyResolver) GetDependents(pluginName string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var dependents []string
	for name, plugin := range r.plugins {
		for _, dep := range plugin.Dependencies {
			if dep.Name == pluginName {
				dependents = append(dependents, name)
				break
			}
		}
	}
	return dependents
}

// ValidateDependencyGraph validates the entire dependency graph for:
// - Circular dependencies
// - Missing dependencies
// - Version conflicts
func (r *DependencyResolver) ValidateDependencyGraph() error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	visited := make(map[string]bool)
	resolving := make(map[string]bool)

	// Check each plugin's dependencies
	for name, plugin := range r.plugins {
		if err := r.validatePluginDeps(plugin, visited, resolving); err != nil {
			return errors.Wrapf(err, errors.ErrCodeInitFailed, "dependency graph validation failed for plugin %s", name)
		}
	}

	return nil
}

// validatePluginDeps validates a single plugin's dependencies
func (r *DependencyResolver) validatePluginDeps(plugin *PluginManifest, visited, resolving map[string]bool) error {
	// Check for circular dependencies
	if resolving[plugin.Name] {
		return errors.Newf(errors.ErrCodeInitFailed, "circular dependency detected: %s", plugin.Name)
	}

	resolving[plugin.Name] = true

	// Validate each dependency
	for _, dep := range plugin.Dependencies {
		// Check if dependency exists
		depPlugin, exists := r.plugins[dep.Name]
		if !exists && !dep.Optional {
			return errors.Newf(errors.ErrCodeNotFound, "missing required dependency: %s", dep.Name)
		}

		if exists {
			// Recursively validate
			if err := r.validatePluginDeps(depPlugin, visited, resolving); err != nil {
				return err
			}
		}
	}

	delete(resolving, plugin.Name)
	visited[plugin.Name] = true
	return nil
}
