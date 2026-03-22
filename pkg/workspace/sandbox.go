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

// SPDX-License-Identifier: AGPL-3.0-or-later

package workspace

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// SandboxType defines the type of sandbox isolation
type SandboxType string

const (
	SandboxTypeNone    SandboxType = "none"      // No isolation
	SandboxTypeProcess  SandboxType = "process"   // Process-level isolation
	SandboxTypeDocker   SandboxType = "docker"   // Docker container isolation
	SandboxTypeVM       SandboxType = "vm"       // Virtual machine isolation (future)
)

// SandboxSession represents an isolated execution environment
type SandboxSession struct {
	ID       string       `json:"id"`
	Type     SandboxType  `json:"type"`
	Skills   []string     `json:"skills"`
	StartedAt time.Time   `json:"started_at"`
	Env      map[string]string `json:"env"`
}

// SandboxConfig contains sandbox configuration
type SandboxConfig struct {
	Type        SandboxType           `json:"type"`
	Timeout     time.Duration        `json:"timeout"`
	MemoryLimit string               `json:"memory_limit"`
	CPUQuota    float64              `json:"cpu_quota"`
	NetworkMode string              `json:"network_mode"`
	ReadOnly    bool                `json:"read_only"`
	AllowedDirs []string            `json:"allowed_dirs"`
}

// DefaultSandboxConfig returns the default sandbox configuration
func DefaultSandboxConfig() *SandboxConfig {
	return &SandboxConfig{
		Type:        SandboxTypeDocker,
		Timeout:     5 * time.Minute,
		MemoryLimit: "512m",
		CPUQuota:    1.0,
		NetworkMode: "none", // No network by default
		ReadOnly:    true,
		AllowedDirs: []string{"/workspace"},
	}
}

// SandboxManager manages sandbox execution environments
type SandboxManager struct {
	mu       sync.RWMutex
	config   *SandboxConfig
	sessions map[string]*SandboxSession
	workspace string
}

// NewSandboxManager creates a new sandbox manager
func NewSandboxManager(workspace string, config *SandboxConfig) (*SandboxManager, error) {
	if config == nil {
		config = DefaultSandboxConfig()
	}

	manager := &SandboxManager{
		config:    config,
		sessions: make(map[string]*SandboxSession),
		workspace: workspace,
	}

	// Ensure workspace directory exists
	if err := os.MkdirAll(workspace, 0755); err != nil {
		return nil, err
	}

	return manager, nil
}

// CreateSession creates a new sandbox session
func (m *SandboxManager) CreateSession(ctx context.Context, skills []string) (*SandboxSession, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	session := &SandboxSession{
		ID:        generateSessionID(),
		Type:      m.config.Type,
		Skills:    skills,
		StartedAt: time.Now(),
		Env:       make(map[string]string),
	}

	m.sessions[session.ID] = session

	// Setup based on sandbox type
	switch m.config.Type {
	case SandboxTypeDocker:
		if err := m.setupDockerSession(ctx, session); err != nil {
			delete(m.sessions, session.ID)
			return nil, fmt.Errorf("docker setup failed: %w", err)
		}
	case SandboxTypeProcess:
		if err := m.setupProcessSession(session); err != nil {
			delete(m.sessions, session.ID)
			return nil, fmt.Errorf("process setup failed: %w", err)
		}
	}

	return session, nil
}

// setupDockerSession sets up a Docker container session
func (m *SandboxManager) setupDockerSession(ctx context.Context, session *SandboxSession) error {
	// Check if Docker is available
	cmd := exec.CommandContext(ctx, "docker", "info")
	if err := cmd.Run(); err != nil {
		// Docker not available, fallback to process isolation
		session.Type = SandboxTypeProcess
		return nil
	}

	// Build container name
	containerName := fmt.Sprintf("openclaw-sandbox-%s", session.ID)

	// Prepare docker run command
	args := []string{
		"run",
		"--name", containerName,
		"--rm",
		"--network", m.config.NetworkMode,
		"--memory", m.config.MemoryLimit,
		"--cpus", fmt.Sprintf("%.2f", m.config.CPUQuota),
		"-v", fmt.Sprintf("%s:/workspace", m.workspace),
		"-w", "/workspace",
		"--security-opt", "no-new-privileges",
	}

	// Add read-only root filesystem if configured
	if m.config.ReadOnly {
		args = append(args, "--read-only")
	}

	// Add allowed directories as volumes
	for _, dir := range m.config.AllowedDirs {
		if dir != "/workspace" {
			absDir, err := filepath.Abs(dir)
			if err == nil {
				args = append(args, "-v", fmt.Sprintf("%s:%s:ro", absDir, dir))
			}
		}
	}

	// Use alpine image for lightweight execution
	args = append(args, "alpine:latest", "sleep", "3600")

	// Run container in background
	cmd = exec.CommandContext(ctx, "docker", args...)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	// Store container name in environment
	session.Env["CONTAINER_NAME"] = containerName
	session.Env["SANDBOX_TYPE"] = "docker"

	return nil
}

// setupProcessSession sets up process-level isolation
func (m *SandboxManager) setupProcessSession(session *SandboxSession) error {
	session.Env["SANDBOX_TYPE"] = "process"
	session.Env["SANDBOX_WORKSPACE"] = m.workspace
	return nil
}

// Execute runs a command in a sandbox session
func (m *SandboxManager) Execute(ctx context.Context, sessionID string, cmd string, args []string) (string, error) {
	m.mu.RLock()
	session, exists := m.sessions[sessionID]
	m.mu.RUnlock()

	if !exists {
		return "", fmt.Errorf("session not found: %s", sessionID)
	}

	// Create execution context with timeout
	execCtx, cancel := context.WithTimeout(ctx, m.config.Timeout)
	defer cancel()

	var result string
	var err error

	switch session.Type {
	case SandboxTypeDocker:
		result, err = m.executeInDocker(execCtx, session, cmd, args)
	case SandboxTypeProcess:
		result, err = m.executeInProcess(execCtx, session, cmd, args)
	default:
		result, err = m.executeDirect(execCtx, cmd, args)
	}

	return result, err
}

// executeInDocker executes command in Docker container
func (m *SandboxManager) executeInDocker(ctx context.Context, session *SandboxSession, cmd string, args []string) (string, error) {
	containerName := session.Env["CONTAINER_NAME"]

	// Check if container is still running
	checkCmd := exec.CommandContext(ctx, "docker", "inspect", "-f", "{{.State.Running}}", containerName)
	if err := checkCmd.Run(); err != nil {
		return "", fmt.Errorf("container not running: %s", containerName)
	}

	// Execute command in container
	execArgs := append([]string{"exec", containerName, cmd}, args...)
	execCmd := exec.CommandContext(ctx, "docker", execArgs...)
	execCmd.Dir = "/workspace"

	output, err := execCmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("execution failed: %w, output: %s", err, string(output))
	}

	return string(output), nil
}

// executeInProcess executes command with process isolation
func (m *SandboxManager) executeInProcess(ctx context.Context, session *SandboxSession, cmd string, args []string) (string, error) {
	execCmd := exec.CommandContext(ctx, cmd, args...)
	execCmd.Dir = m.workspace

	// Apply environment restrictions
	env := os.Environ()
	for k, v := range session.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	execCmd.Env = env

	output, err := execCmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("execution failed: %w, output: %s", err, string(output))
	}

	return string(output), nil
}

// executeDirect executes command without isolation
func (m *SandboxManager) executeDirect(ctx context.Context, cmd string, args []string) (string, error) {
	execCmd := exec.CommandContext(ctx, cmd, args...)
	execCmd.Dir = m.workspace

	output, err := execCmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("execution failed: %w, output: %s", err, string(output))
	}

	return string(output), nil
}

// DestroySession destroys a sandbox session
func (m *SandboxManager) DestroySession(ctx context.Context, sessionID string) error {
	m.mu.Lock()
	session, exists := m.sessions[sessionID]
	if exists {
		delete(m.sessions, sessionID)
	}
	m.mu.Unlock()

	if !exists {
		return nil
	}

	switch session.Type {
	case SandboxTypeDocker:
		containerName := session.Env["CONTAINER_NAME"]
		cmd := exec.CommandContext(ctx, "docker", "kill", containerName)
		return cmd.Run()
	}

	return nil
}

// GetSession returns a session by ID
func (m *SandboxManager) GetSession(sessionID string) *SandboxSession {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.sessions[sessionID]
}

// ListSessions returns all active sessions
func (m *SandboxManager) ListSessions() []*SandboxSession {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sessions := make([]*SandboxSession, 0, len(m.sessions))
	for _, s := range m.sessions {
		sessions = append(sessions, s)
	}
	return sessions
}

// Cleanup removes expired sessions
func (m *SandboxManager) Cleanup(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var expired []string
	for id, session := range m.sessions {
		if time.Since(session.StartedAt) > m.config.Timeout*2 {
			expired = append(expired, id)
		}
	}

	for _, id := range expired {
		session := m.sessions[id]
		delete(m.sessions, id)

		// Cleanup Docker container if needed
		if session.Type == SandboxTypeDocker {
			containerName := session.Env["CONTAINER_NAME"]
			exec.CommandContext(ctx, "docker", "kill", containerName).Run()
		}
	}

	return nil
}

// DockerInfo returns Docker system information
type DockerInfo struct {
	Available    bool   `json:"available"`
	Version      string `json:"version"`
	Containers   int    `json:"containers"`
	Images       int    `json:"images"`
}

// CheckDocker checks Docker availability
func CheckDocker(ctx context.Context) (*DockerInfo, error) {
	info := &DockerInfo{}

	// Check if docker is available
	cmd := exec.CommandContext(ctx, "docker", "version", "--format", "{{json .}}")
	output, err := cmd.Output()
	if err != nil {
		info.Available = false
		return info, nil
	}

	var version struct {
		Server struct {
			Version string `json:"Version"`
		} `json:"Server"`
	}
	if err := json.Unmarshal(output, &version); err == nil {
		info.Version = version.Server.Version
	}

	// Count containers
	cmd = exec.CommandContext(ctx, "docker", "ps", "-a", "-q")
	containers, _ := cmd.Output()
	info.Containers = len(strings.Split(strings.TrimSpace(string(containers)), "\n"))

	// Count images
	cmd = exec.CommandContext(ctx, "docker", "images", "-q")
	images, _ := cmd.Output()
	info.Images = len(strings.Split(strings.TrimSpace(string(images)), "\n"))

	info.Available = true
	return info, nil
}

func generateSessionID() string {
	return fmt.Sprintf("sess-%d", time.Now().UnixNano())
}
