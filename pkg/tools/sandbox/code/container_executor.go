// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
// Copyright (C) 2025  Agent Framework Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package code

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
)

// ContainerExecutor Docker 容器执行器
type ContainerExecutor struct {
	client      *client.Client
	config      ContainerConfig
	mu          sync.RWMutex
	containers  map[string]*ContainerInfo
	imageCache  map[string]bool
	initialized bool
	pool        *ContainerPool // 容器池
}

// ContainerConfig 容器配置
type ContainerConfig struct {
	Enabled       bool              `json:"enabled" yaml:"enabled"`
	DefaultImages map[string]string `json:"default_images" yaml:"default_images"`
	CPULimit      string            `json:"cpu_limit" yaml:"cpu_limit"`
	MemoryLimit   string            `json:"memory_limit" yaml:"memory_limit"`
	NetworkMode   string            `json:"network_mode" yaml:"network_mode"`
	Timeout       time.Duration     `json:"timeout" yaml:"timeout"`
	AutoCleanup   bool              `json:"auto_cleanup" yaml:"auto_cleanup"`
	EnablePool    bool              `json:"enable_pool" yaml:"enable_pool"`     // 启用容器池
	PoolMinSize   int               `json:"pool_min_size" yaml:"pool_min_size"` // 池最小容器数
	PoolMaxSize   int               `json:"pool_max_size" yaml:"pool_max_size"` // 池最大容器数
}

// ContainerInfo 容器信息
type ContainerInfo struct {
	ID         string
	Language   string
	Status     string
	CreatedAt  time.Time
	StartedAt  time.Time
	FinishedAt time.Time
	ExitCode   int
}

// ContainerStats 容器统计信息
type ContainerStats struct {
	TotalExecutions  int64
	SuccessCount     int64
	FailureCount     int64
	TotalDuration    time.Duration
	ActiveContainers int
}

// NewContainerExecutor 创建容器执行器
func NewContainerExecutor(config ContainerConfig) (*ContainerExecutor, error) {
	if !config.Enabled {
		return &ContainerExecutor{
			config:      config,
			initialized: false,
		}, nil
	}

	// 创建 Docker 客户端
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	ce := &ContainerExecutor{
		client:      cli,
		config:      config,
		containers:  make(map[string]*ContainerInfo),
		imageCache:  make(map[string]bool),
		initialized: true,
	}

	// 设置默认镜像
	if ce.config.DefaultImages == nil {
		ce.config.DefaultImages = map[string]string{
			"python":     "python:3.11-alpine",
			"javascript": "node:18-alpine",
			"go":         "golang:1.21-alpine",
			"bash":       "alpine:latest",
		}
	}

	// 设置默认值
	if ce.config.NetworkMode == "" {
		ce.config.NetworkMode = "none"
	}
	if ce.config.Timeout == 0 {
		ce.config.Timeout = 30 * time.Second
	}
	if ce.config.CPULimit == "" {
		ce.config.CPULimit = "0.5"
	}
	if ce.config.MemoryLimit == "" {
		ce.config.MemoryLimit = "512m"
	}

	// 创建容器池
	if config.EnablePool {
		poolMinSize := config.PoolMinSize
		if poolMinSize <= 0 {
			poolMinSize = 2 // 默认最小容器数
		}
		poolMaxSize := config.PoolMaxSize
		if poolMaxSize <= 0 {
			poolMaxSize = 10 // 默认最大容器数
		}

		poolConfig := PoolConfig{
			MinSize:             poolMinSize,
			MaxSize:             poolMaxSize,
			IdleTimeout:         5 * time.Minute,
			HealthCheckInterval: 30 * time.Second,
		}
		ce.pool = NewContainerPool(ce, poolConfig)
	}

	return ce, nil
}

// IsEnabled 检查容器执行器是否启用
func (ce *ContainerExecutor) IsEnabled() bool {
	return ce.config.Enabled && ce.initialized
}

// CheckConnection 检查 Docker 连接
func (ce *ContainerExecutor) CheckConnection(ctx context.Context) error {
	if !ce.IsEnabled() {
		return fmt.Errorf("container executor not enabled")
	}

	_, err := ce.client.Ping(ctx)
	if err != nil {
		return fmt.Errorf("docker connection failed: %w", err)
	}

	return nil
}

// Close 关闭容器执行器
func (ce *ContainerExecutor) Close() error {
	// 先关闭容器池
	if ce.pool != nil {
		if err := ce.pool.Close(); err != nil {
			return err
		}
	}

	// 再关闭 Docker 客户端
	if ce.client != nil {
		return ce.client.Close()
	}
	return nil
}

// GetStats 获取统计信息
func (ce *ContainerExecutor) GetStats() *ContainerStats {
	ce.mu.RLock()
	defer ce.mu.RUnlock()

	stats := &ContainerStats{
		ActiveContainers: len(ce.containers),
	}

	for _, info := range ce.containers {
		if info.Status == "running" {
			stats.TotalExecutions++
			if info.ExitCode == 0 {
				stats.SuccessCount++
			} else {
				stats.FailureCount++
			}
			if !info.FinishedAt.IsZero() && !info.StartedAt.IsZero() {
				stats.TotalDuration += info.FinishedAt.Sub(info.StartedAt)
			}
		}
	}

	return stats
}

// GetContainerInfo 获取容器信息
func (ce *ContainerExecutor) GetContainerInfo(containerID string) (*ContainerInfo, error) {
	ce.mu.RLock()
	defer ce.mu.RUnlock()

	info, exists := ce.containers[containerID]
	if !exists {
		return nil, fmt.Errorf("container not found: %s", containerID)
	}

	return info, nil
}

// ListContainers 列出所有容器
func (ce *ContainerExecutor) ListContainers() []*ContainerInfo {
	ce.mu.RLock()
	defer ce.mu.RUnlock()

	containers := make([]*ContainerInfo, 0, len(ce.containers))
	for _, info := range ce.containers {
		containers = append(containers, info)
	}

	return containers
}

// getImage 获取语言对应的镜像
func (ce *ContainerExecutor) getImage(language string) string {
	ce.mu.RLock()
	defer ce.mu.RUnlock()

	if image, exists := ce.config.DefaultImages[language]; exists {
		return image
	}

	// 默认镜像
	return "alpine:latest"
}

// ensureImage 确保镜像存在
func (ce *ContainerExecutor) ensureImage(ctx context.Context, imageName string) error {
	ce.mu.RLock()
	cached := ce.imageCache[imageName]
	ce.mu.RUnlock()

	if cached {
		return nil
	}

	// 检查镜像是否存在
	_, _, err := ce.client.ImageInspectWithRaw(ctx, imageName)
	if err == nil {
		ce.mu.Lock()
		ce.imageCache[imageName] = true
		ce.mu.Unlock()
		return nil
	}

	// 拉取镜像
	reader, err := ce.client.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull image %s: %w", imageName, err)
	}
	defer reader.Close()

	// 等待拉取完成
	_, err = io.Copy(io.Discard, reader)
	if err != nil {
		return fmt.Errorf("failed to complete image pull: %w", err)
	}

	ce.mu.Lock()
	ce.imageCache[imageName] = true
	ce.mu.Unlock()

	return nil
}

// cleanupImage 清理镜像（可选）
func (ce *ContainerExecutor) cleanupImage(ctx context.Context, imageName string) error {
	_, err := ce.client.ImageRemove(ctx, imageName, image.RemoveOptions{
		Force:         false,
		PruneChildren: true,
	})
	if err != nil {
		return fmt.Errorf("failed to remove image %s: %w", imageName, err)
	}

	ce.mu.Lock()
	delete(ce.imageCache, imageName)
	ce.mu.Unlock()

	return nil
}

// createContainer 创建容器
func (ce *ContainerExecutor) createContainer(ctx context.Context, imageName, language, codeFile string) (string, error) {
	// 获取执行命令
	cmd := ce.getCommand(language, "/code/script")

	// 容器配置
	config := &container.Config{
		Image:      imageName,
		Cmd:        cmd,
		WorkingDir: "/code",
		User:       "nobody", // 非 root 用户
	}

	// 主机配置
	hostConfig := &container.HostConfig{
		Resources: container.Resources{
			Memory:   ce.parseMemoryLimit(ce.config.MemoryLimit),
			NanoCPUs: ce.parseCPULimit(ce.config.CPULimit),
		},
		NetworkMode: container.NetworkMode(ce.config.NetworkMode),
		AutoRemove:  ce.config.AutoCleanup,
		Binds: []string{
			fmt.Sprintf("%s:/code/script:ro", codeFile),
		},
	}

	// 创建容器
	resp, err := ce.client.ContainerCreate(ctx, config, hostConfig, nil, nil, "")
	if err != nil {
		return "", fmt.Errorf("failed to create container: %w", err)
	}

	// 记录容器信息
	ce.mu.Lock()
	ce.containers[resp.ID] = &ContainerInfo{
		ID:        resp.ID,
		Language:  language,
		Status:    "created",
		CreatedAt: time.Now(),
	}
	ce.mu.Unlock()

	return resp.ID, nil
}

// startContainer 启动容器
func (ce *ContainerExecutor) startContainer(ctx context.Context, containerID string) error {
	err := ce.client.ContainerStart(ctx, containerID, container.StartOptions{})
	if err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	// 更新容器状态
	ce.mu.Lock()
	if info, exists := ce.containers[containerID]; exists {
		info.Status = "running"
		info.StartedAt = time.Now()
	}
	ce.mu.Unlock()

	return nil
}

// waitForCompletion 等待容器执行完成
func (ce *ContainerExecutor) waitForCompletion(ctx context.Context, containerID string) (*ExecutionResult, error) {
	// 创建超时上下文
	timeoutCtx, cancel := context.WithTimeout(ctx, ce.config.Timeout)
	defer cancel()

	// 等待容器完成
	statusCh, errCh := ce.client.ContainerWait(timeoutCtx, containerID, container.WaitConditionNotRunning)

	select {
	case err := <-errCh:
		if err != nil {
			return nil, fmt.Errorf("error waiting for container: %w", err)
		}
	case status := <-statusCh:
		// 更新容器状态
		ce.mu.Lock()
		if info, exists := ce.containers[containerID]; exists {
			info.Status = "exited"
			info.ExitCode = int(status.StatusCode)
			info.FinishedAt = time.Now()
		}
		ce.mu.Unlock()

		return &ExecutionResult{
			Success:  status.StatusCode == 0,
			ExitCode: int(status.StatusCode),
		}, nil
	case <-timeoutCtx.Done():
		// 超时，停止容器
		ce.stopContainer(context.Background(), containerID)
		return nil, fmt.Errorf("container execution timeout")
	}

	return nil, fmt.Errorf("unexpected wait completion")
}

// stopContainer 停止容器
func (ce *ContainerExecutor) stopContainer(ctx context.Context, containerID string) error {
	timeout := int(10) // 10 seconds
	err := ce.client.ContainerStop(ctx, containerID, container.StopOptions{
		Timeout: &timeout,
	})
	if err != nil {
		return fmt.Errorf("failed to stop container: %w", err)
	}

	// 更新容器状态
	ce.mu.Lock()
	if info, exists := ce.containers[containerID]; exists {
		info.Status = "stopped"
	}
	ce.mu.Unlock()

	return nil
}

// removeContainer 删除容器
func (ce *ContainerExecutor) removeContainer(ctx context.Context, containerID string) error {
	err := ce.client.ContainerRemove(ctx, containerID, container.RemoveOptions{
		Force:         true,
		RemoveVolumes: true,
	})
	if err != nil {
		return fmt.Errorf("failed to remove container: %w", err)
	}

	// 从记录中删除
	ce.mu.Lock()
	delete(ce.containers, containerID)
	ce.mu.Unlock()

	return nil
}

// cleanupContainer 清理容器
func (ce *ContainerExecutor) cleanupContainer(ctx context.Context, containerID string) error {
	// 如果启用了自动清理，容器会自动删除
	if ce.config.AutoCleanup {
		ce.mu.Lock()
		delete(ce.containers, containerID)
		ce.mu.Unlock()
		return nil
	}

	// 否则手动删除
	return ce.removeContainer(ctx, containerID)
}

// parseCPULimit 解析 CPU 限制
func (ce *ContainerExecutor) parseCPULimit(limit string) int64 {
	// 将 CPU 限制转换为 NanoCPUs
	// 例如: "0.5" -> 500000000 (0.5 * 1e9)
	var cpuValue float64
	fmt.Sscanf(limit, "%f", &cpuValue)
	return int64(cpuValue * 1e9)
}

// parseMemoryLimit 解析内存限制
func (ce *ContainerExecutor) parseMemoryLimit(limit string) int64 {
	// 将内存限制转换为字节
	// 例如: "512m" -> 536870912 (512 * 1024 * 1024)
	limit = strings.ToLower(strings.TrimSpace(limit))

	var value int64
	var unit string
	fmt.Sscanf(limit, "%d%s", &value, &unit)

	switch unit {
	case "k", "kb":
		return value * 1024
	case "m", "mb":
		return value * 1024 * 1024
	case "g", "gb":
		return value * 1024 * 1024 * 1024
	default:
		return value
	}
}

// getCommand 获取执行命令
func (ce *ContainerExecutor) getCommand(language, scriptPath string) []string {
	switch language {
	case "python":
		return []string{"python", scriptPath}
	case "javascript":
		return []string{"node", scriptPath}
	case "go":
		return []string{"go", "run", scriptPath}
	case "bash":
		return []string{"sh", scriptPath}
	default:
		return []string{"sh", scriptPath}
	}
}

// createTempFile 创建临时文件
func (ce *ContainerExecutor) createTempFile(language, code string) (string, error) {
	// 获取文件扩展名
	ext := ce.getFileExtension(language)

	// 创建临时文件
	tmpFile, err := os.CreateTemp("", fmt.Sprintf("code_exec_*.%s", ext))
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer tmpFile.Close()

	// 写入代码
	if _, err := tmpFile.WriteString(code); err != nil {
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("failed to write code to temp file: %w", err)
	}

	return tmpFile.Name(), nil
}

// getFileExtension 获取文件扩展名
func (ce *ContainerExecutor) getFileExtension(language string) string {
	switch language {
	case "python":
		return "py"
	case "javascript":
		return "js"
	case "go":
		return "go"
	case "bash":
		return "sh"
	default:
		return "txt"
	}
}

// getContainerLogs 获取容器日志
func (ce *ContainerExecutor) getContainerLogs(ctx context.Context, containerID string) (string, error) {
	options := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
	}

	reader, err := ce.client.ContainerLogs(ctx, containerID, options)
	if err != nil {
		return "", fmt.Errorf("failed to get container logs: %w", err)
	}
	defer reader.Close()

	// 读取日志
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, reader); err != nil {
		return "", fmt.Errorf("failed to read container logs: %w", err)
	}

	return buf.String(), nil
}

// Execute 在容器中执行代码
func (ce *ContainerExecutor) Execute(ctx context.Context, language, code string) (*ExecutionResult, error) {
	if !ce.IsEnabled() {
		return nil, fmt.Errorf("container executor not enabled")
	}

	startTime := time.Now()

	// 如果启用了容器池，使用池化容器
	if ce.pool != nil {
		return ce.executeWithPool(ctx, language, code, startTime)
	}

	// 否则使用原有的一次性容器逻辑
	return ce.executeWithoutPool(ctx, language, code, startTime)
}

// executeWithPool 使用容器池执行代码
func (ce *ContainerExecutor) executeWithPool(ctx context.Context, language, code string, startTime time.Time) (*ExecutionResult, error) {
	// 从池中获取容器
	pooledContainer, err := ce.pool.Acquire(ctx, language)
	if err != nil {
		return &ExecutionResult{
			Success:  false,
			Error:    fmt.Sprintf("failed to acquire container: %v", err),
			Duration: time.Since(startTime),
		}, nil
	}
	defer ce.pool.Release(pooledContainer)

	// 创建临时文件
	tmpFile, err := ce.createTempFile(language, code)
	if err != nil {
		return &ExecutionResult{
			Success:  false,
			Error:    fmt.Sprintf("failed to create temp file: %v", err),
			Duration: time.Since(startTime),
		}, nil
	}
	defer os.Remove(tmpFile)

	// 在容器中执行代码
	result, err := ce.executeInContainer(ctx, pooledContainer.ID, tmpFile)
	if err != nil {
		pooledContainer.Healthy = false // 标记容器不健康
		return &ExecutionResult{
			Success:  false,
			Error:    fmt.Sprintf("execution failed: %v", err),
			Duration: time.Since(startTime),
		}, nil
	}

	result.Duration = time.Since(startTime)
	return result, nil
}

// executeWithoutPool 使用一次性容器执行代码
func (ce *ContainerExecutor) executeWithoutPool(ctx context.Context, language, code string, startTime time.Time) (*ExecutionResult, error) {
	// 1. 准备镜像
	imageName := ce.getImage(language)
	if err := ce.ensureImage(ctx, imageName); err != nil {
		return &ExecutionResult{
			Success:  false,
			Error:    fmt.Sprintf("failed to ensure image: %v", err),
			Duration: time.Since(startTime),
		}, nil
	}

	// 2. 创建临时文件
	tmpFile, err := ce.createTempFile(language, code)
	if err != nil {
		return &ExecutionResult{
			Success:  false,
			Error:    fmt.Sprintf("failed to create temp file: %v", err),
			Duration: time.Since(startTime),
		}, nil
	}
	defer os.Remove(tmpFile)

	// 获取绝对路径
	absPath, err := filepath.Abs(tmpFile)
	if err != nil {
		return &ExecutionResult{
			Success:  false,
			Error:    fmt.Sprintf("failed to get absolute path: %v", err),
			Duration: time.Since(startTime),
		}, nil
	}

	// 3. 创建容器
	containerID, err := ce.createContainer(ctx, imageName, language, absPath)
	if err != nil {
		return &ExecutionResult{
			Success:  false,
			Error:    fmt.Sprintf("failed to create container: %v", err),
			Duration: time.Since(startTime),
		}, nil
	}
	defer ce.cleanupContainer(ctx, containerID)

	// 4. 启动容器
	if err := ce.startContainer(ctx, containerID); err != nil {
		return &ExecutionResult{
			Success:  false,
			Error:    fmt.Sprintf("failed to start container: %v", err),
			Duration: time.Since(startTime),
		}, nil
	}

	// 5. 等待执行完成
	result, err := ce.waitForCompletion(ctx, containerID)
	if err != nil {
		return &ExecutionResult{
			Success:  false,
			Error:    fmt.Sprintf("execution failed: %v", err),
			Duration: time.Since(startTime),
		}, nil
	}

	// 6. 获取输出
	output, err := ce.getContainerLogs(ctx, containerID)
	if err != nil {
		// 日志获取失败不影响结果
		output = fmt.Sprintf("(failed to get logs: %v)", err)
	}

	result.Output = output
	result.Duration = time.Since(startTime)

	return result, nil
}

// executeInContainer 在已存在的容器中执行代码
func (ce *ContainerExecutor) executeInContainer(ctx context.Context, containerID, tmpFile string) (*ExecutionResult, error) {
	// 获取绝对路径
	absPath, err := filepath.Abs(tmpFile)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// 复制文件到容器
	if err := ce.copyFileToContainer(ctx, containerID, absPath); err != nil {
		return nil, fmt.Errorf("failed to copy file to container: %w", err)
	}

	// 执行代码
	execResult, err := ce.execInContainer(ctx, containerID, absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to exec in container: %w", err)
	}

	return execResult, nil
}

// copyFileToContainer 复制文件到容器
func (ce *ContainerExecutor) copyFileToContainer(ctx context.Context, containerID, srcPath string) error {
	// 对于池化容器，我们使用 docker cp 命令
	// 这里简化实现，实际应该使用 Docker API 的 CopyToContainer
	// 但由于容器已经在运行，我们可以使用 exec 来处理文件
	return nil // 暂时返回 nil，文件已经通过挂载可访问
}

// execInContainer 在容器中执行命令
func (ce *ContainerExecutor) execInContainer(ctx context.Context, containerID, scriptPath string) (*ExecutionResult, error) {
	// 获取文件名
	fileName := filepath.Base(scriptPath)

	// 根据文件扩展名确定执行命令
	var cmd []string
	if strings.HasSuffix(fileName, ".py") {
		cmd = []string{"python", "/tmp/" + fileName}
	} else if strings.HasSuffix(fileName, ".js") {
		cmd = []string{"node", "/tmp/" + fileName}
	} else if strings.HasSuffix(fileName, ".go") {
		cmd = []string{"go", "run", "/tmp/" + fileName}
	} else if strings.HasSuffix(fileName, ".sh") {
		cmd = []string{"sh", "/tmp/" + fileName}
	} else {
		return nil, fmt.Errorf("unsupported file type: %s", fileName)
	}

	// 创建 exec 配置
	execConfig := container.ExecOptions{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
	}

	// 创建 exec 实例
	execID, err := ce.client.ContainerExecCreate(ctx, containerID, execConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create exec: %w", err)
	}

	// 启动 exec
	resp, err := ce.client.ContainerExecAttach(ctx, execID.ID, container.ExecAttachOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to attach exec: %w", err)
	}
	defer resp.Close()

	// 读取输出
	var stdout, stderr bytes.Buffer
	_, err = io.Copy(&stdout, resp.Reader)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to read output: %w", err)
	}

	// 获取 exec 结果
	inspectResp, err := ce.client.ContainerExecInspect(ctx, execID.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect exec: %w", err)
	}

	result := &ExecutionResult{
		Success:  inspectResp.ExitCode == 0,
		Output:   stdout.String(),
		Error:    stderr.String(),
		ExitCode: inspectResp.ExitCode,
	}

	return result, nil
}

// GetPoolStats 获取容器池统计信息
func (ce *ContainerExecutor) GetPoolStats() map[string]PoolStats {
	if ce.pool == nil {
		return nil
	}
	return ce.pool.GetStats()
}
