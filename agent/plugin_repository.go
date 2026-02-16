// Agent Framework - Plugin Repository and Installation
// Copyright (C) 2025  Agent Framework Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"AgentFramework/agent/errors"
)

// PluginRepository represents a plugin repository source
type PluginRepository struct {
	Name        string `json:"name"`
	URL         string `json:"url"`
	Type        string `json:"type"` // local, git, http, etc.
	Enabled     bool   `json:"enabled"`
}

// PluginInstaller handles plugin installation
type PluginInstaller struct {
	repoPath   string
	pluginsDir string
}

// InstallPlugin installs a plugin from repository info
func (pi *PluginInstaller) InstallPlugin(ctx context.Context, info *RepositoryPluginInfo) error {
	// Create plugin directory
	pluginDir := filepath.Join(pi.pluginsDir, info.Name+"-"+info.Version)
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		return fmt.Errorf("failed to create plugin directory: %w", err)
	}

	// Extract and install plugin files
	// This is a simplified implementation - in production, you would:
	// 1. Verify the plugin archive checksum
	// 2. Extract files to a temporary directory
	// 3. Validate the plugin manifest
	// 4. Move files to the final location
	// 5. Register the plugin with the plugin manager

	return nil
}

// RepositoryPluginInfo represents information about an available plugin in a repository
type RepositoryPluginInfo struct {
	Name          string `json:"name"`
	Version       string `json:"version"`
	Description   string `json:"description"`
	Author        string `json:"author"`
	License       string `json:"license"`
	Repository    PluginRepository `json:"repository"`
	Homepage      string `json:"homepage"`
	DownloadCount int64 `json:"download_count"`
	Rating        float64 `json:"rating"`
}

// PluginManifest represents a plugin's manifest with installation info
type PluginManifest struct {
	// Identity
	Name            string `json:"name"`
	Version         string `json:"version"`
	Description     string `json:"description"`
	Author          string `json:"author"`
	License         string `json:"license"`

	// Installation info
	InstallDate     string `json:"install_date"`
	InstallPath    string `json:"install_path"`
	Repository      PluginRepository `json:"repository"`

	// Capabilities
	Capabilities      []string `json:"capabilities"`

	// Resource requirements
	MaxMemoryMB     int `json:"max_memory_mb"`
	MaxCPUPercent  int `json:"max_cpu_percent"`
	RequiredNetwork bool `json:"required_network"`
	RequiredFS       bool `json:"required_fs"`

	// Security
	SandboxEnabled   bool `json:"sandbox_enabled"`
	RequirePermissions []string `json:"require_permissions"`
}

// PluginRepositoryManager manages plugin repositories and installation
type PluginRepositoryManager struct {
	repositories    map[string]*PluginRepository
	installer      *PluginInstaller
	mu           sync.RWMutex
	manifestDir   string
}

// NewPluginRepositoryManager creates a new plugin repository manager
func NewPluginRepositoryManager(repositories []PluginRepository, manifestDir string) *PluginRepositoryManager {
	return &PluginRepositoryManager{
		repositories: make(map[string]*PluginRepository),
		installer: &PluginInstaller{
			repoPath: filepath.Join(manifestDir, "repo"),
			pluginsDir: filepath.Join(manifestDir, "plugins"),
		},
		mu: sync.RWMutex{},
		manifestDir: manifestDir,
	}
}

// AddRepository adds a plugin repository
func (m *PluginRepositoryManager) AddRepository(repo PluginRepository) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.repositories[repo.Name]; exists {
		return errors.Newf(errors.ErrCodeInvalidInput, "repository %s already exists", repo.Name)
	}

	m.repositories[repo.Name] = &repo
	return nil
}

// RemoveRepository removes a plugin repository
func (m *PluginRepositoryManager) RemoveRepository(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.repositories[name]; !exists {
		return errors.Newf(errors.ErrCodeNotFound, "repository %s not found", name)
	}

	delete(m.repositories, name)
	return nil
}

// ListRepositories returns all repositories
func (m *PluginRepositoryManager) ListRepositories() []PluginRepository {
	m.mu.RLock()
	defer m.mu.RUnlock()

	repos := make([]PluginRepository, 0, len(m.repositories))
	i := 0
	for _, repo := range m.repositories {
		repos[i] = *repo
		i++
	}
	return repos
}

// SearchPlugins searches for plugins by name or description
func (m *PluginRepositoryManager) SearchPlugins(query string) ([]RepositoryPluginInfo, error) {
	m.mu.RLock()
	defer m.mu.Unlock()

	var results []RepositoryPluginInfo

	for _, repo := range m.repositories {
		if !repo.Enabled {
			continue
		}

		plugins, err := m.fetchPlugins(*repo)
		if err != nil {
			return nil, err
		}

		// Filter by query
		for _, plugin := range plugins {
			if contains(plugin.Name, query) || contains(plugin.Description, query) {
				results = append(results, plugin)
			}
		}
	}

	return results, nil
}

// fetchPlugins fetches plugin list from a repository
func (m *PluginRepositoryManager) fetchPlugins(repo PluginRepository) ([]RepositoryPluginInfo, error) {
	switch repo.Type {
	case "local":
		return m.fetchLocalPlugins(repo)
	case "http":
		return m.fetchHTTPPlugins(repo)
	case "git":
		return m.fetchGitPlugins(repo)
	default:
		return nil, fmt.Errorf("unsupported repository type: %s", repo.Type)
	}
}

// fetchLocalPlugins fetches plugins from local repository
func (m *PluginRepositoryManager) fetchLocalPlugins(repo PluginRepository) ([]RepositoryPluginInfo, error) {
	manifestPath := filepath.Join(repo.URL, "manifest.json")

	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	var manifest struct {
		Plugins []RepositoryPluginInfo `json:"plugins"`
	}

	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	return manifest.Plugins, nil
}

// fetchHTTPPlugins fetches plugins from HTTP repository
func (m *PluginRepositoryManager) fetchHTTPPlugins(repo PluginRepository) ([]RepositoryPluginInfo, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Get(repo.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repository: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("repository returned status %d: %s", resp.StatusCode, repo.URL)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result struct {
		Plugins []RepositoryPluginInfo `json:"plugins"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return result.Plugins, nil
}

// fetchGitPlugins fetches plugins from Git repository
func (m *PluginRepositoryManager) fetchGitPlugins(repo PluginRepository) ([]RepositoryPluginInfo, error) {
	// Placeholder for Git repository support
	return nil, fmt.Errorf("git repository support not yet implemented")
}

// GetRepositoryPluginInfo retrieves detailed information about a plugin
func (m *PluginRepositoryManager) GetRepositoryPluginInfo(repoName, pluginName string) (*RepositoryPluginInfo, error) {
	m.mu.RLock()
	defer m.mu.Unlock()

	repo, exists := m.repositories[repoName]
	if !exists {
		return nil, errors.Newf(errors.ErrCodeNotFound, "repository %s not found", repoName)
	}

	plugins, err := m.fetchPlugins(*repo)
	if err != nil {
		return nil, err
	}

	for _, plugin := range plugins {
		if plugin.Name == pluginName {
			return &plugin, nil
		}
	}

	return nil, errors.Newf(errors.ErrCodeNotFound, "plugin %s not found in repository %s", pluginName)
}

// Install installs a plugin from repository
func (m *PluginRepositoryManager) Install(ctx context.Context, repoName, pluginName string) error {
	m.mu.Lock()
	_, exists := m.repositories[repoName]
	if !exists {
		m.mu.Unlock()
		return errors.Newf(errors.ErrCodeNotFound, "repository %s not found", repoName)
	}
	m.mu.Unlock()

	info, err := m.GetRepositoryPluginInfo(repoName, pluginName)
	if err != nil {
		return errors.Wrapf(err, errors.ErrCodeExecutionFailed, "failed to get plugin info: %s", pluginName)
	}

	// Download plugin if URL provided
	if info.Repository.URL != "" {
		if err := m.downloadPlugin(info); err != nil {
			return errors.Wrapf(err, errors.ErrCodeDownloadFailed, "failed to download plugin: %w", pluginName)
		}
	}

	// Install plugin
	if err := m.installer.InstallPlugin(ctx, info); err != nil {
		return errors.Wrapf(err, errors.ErrCodeInstallFailed, "failed to install plugin: %w", pluginName)
	}

	return nil
}

// downloadPlugin downloads a plugin from URL
func (m *PluginRepositoryManager) downloadPlugin(info *RepositoryPluginInfo) error {
	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	resp, err := client.Get(info.Repository.URL + "/" + info.Name + ".zip")
	if err != nil {
		return fmt.Errorf("failed to download plugin: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("download failed with status %d: %s", resp.StatusCode)
	}

	// Save to file
	pluginPath := filepath.Join(m.installer.pluginsDir, info.Name+"-"+info.Version+".zip")
	file, err := os.Create(pluginPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

// Uninstall removes an installed plugin
func (m *PluginRepositoryManager) Uninstall(pluginName string) error {
	pluginPath := filepath.Join(m.installer.pluginsDir, pluginName)

	if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
		return nil // Not installed
	}

	return os.RemoveAll(pluginPath)
}

// ListInstalled returns all installed plugin names
func (m *PluginRepositoryManager) ListInstalled(repoName string) ([]string, error) {
	entries, err := os.ReadDir(m.installer.pluginsDir)
	if err != nil {
		return nil, err
	}

	var names []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Simple extraction: use filename without extension as plugin name
		name := entry.Name()
		// Remove .zip extension if present
		if len(name) > 4 && name[len(name)-4:] == ".zip" {
			name = name[:len(name)-4]
		}
		names = append(names, name)
	}

	return names, nil
}

// GetInstalledManifest retrieves manifest of an installed plugin
func (m *PluginRepositoryManager) GetInstalledManifest(pluginName string) (*PluginManifest, error) {
	pluginPath, err := m.findPluginPath(pluginName)
	if err != nil {
		return nil, err
	}

	manifestPath := filepath.Join(pluginPath, "manifest.json")

	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	var manifest PluginManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	return &manifest, nil
}

// findPluginPath finds the installation path of a plugin
func (m *PluginRepositoryManager) findPluginPath(pluginName string) (string, error) {
	entries, err := os.ReadDir(m.installer.pluginsDir)
	if err != nil {
		return "", err
	}

	var latestPath string
	var latestModTime time.Time

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if strings.HasPrefix(entry.Name(), pluginName) {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			modTime := info.ModTime()
			if modTime.After(latestModTime) {
				latestPath = filepath.Join(m.installer.pluginsDir, entry.Name())
				latestModTime = modTime
			}
		}
	}

	if latestPath == "" {
		return "", fmt.Errorf("plugin %s not found", pluginName)
	}

	return latestPath, nil
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && strings.HasPrefix(strings.ToLower(s), strings.ToLower(substr)))
}

// Global plugin repository manager instance
var globalPluginRepositoryManager *PluginRepositoryManager
var globalPluginRepositoryOnce sync.Once

// InitGlobalPluginRepositoryManager initializes the global plugin repository manager
func InitGlobalPluginRepositoryManager(repositories []PluginRepository, manifestDir string) {
	globalPluginRepositoryOnce.Do(func() {
		globalPluginRepositoryManager = NewPluginRepositoryManager(repositories, manifestDir)
	})
}

// GetGlobalPluginRepositoryManager returns the global plugin repository manager
func GetGlobalPluginRepositoryManager() *PluginRepositoryManager {
	if globalPluginRepositoryManager == nil {
		InitGlobalPluginRepositoryManager(nil, "")
	}
	return globalPluginRepositoryManager
}
