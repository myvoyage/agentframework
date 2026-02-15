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

package vm

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// VMSandbox VM 级沙箱，提供更强的隔离
type VMSandbox struct {
	config       VMConfig
	vmType       VMType
	status       VMStatus
	process      *exec.Cmd
	mu           sync.RWMutex
	startTime    time.Time
	networkDisabled bool
}

// VMType 虚拟机类型
type VMType string

const (
	VMTypeWSL2  VMType = "wsl2"
	VMTypeLima  VMType = "lima"
	VMTypeQEMU  VMType = "qemu"
	VMTypeDocker VMType = "docker"
)

// VMStatus 虚拟机状态
type VMStatus string

const (
	VMStatusStopped  VMStatus = "stopped"
	VMStatusStarting VMStatus = "starting"
	VMStatusRunning  VMStatus = "running"
	VMStatusStopping VMStatus = "stopping"
	VMStatusError    VMStatus = "error"
)

// VMConfig 虚拟机配置
type VMConfig struct {
	// 通用配置
	Name           string            `json:"name"`
	CPUs           int               `json:"cpus"`
	Memory         string            `json:"memory"`       // e.g., "4GB", "4096MB"
	DiskSize       string            `json:"disk_size"`    // e.g., "50GB"
	OSImage        string            `json:"os_image"`     // e.g., "ubuntu:22.04"

	// 网络配置
	NetworkDisabled bool             `json:"network_disabled"`
	NetworkMode     string           `json:"network_mode"`  // "bridged", "nat", "host"

	// 安全配置
	ReadOnlyRoot    bool             `json:"read_only_root"`
	EnableSeccomp   bool             `json:"enable_seccomp"`
	AppArmorProfile string           `json:"apparmor_profile"`

	// 共享配置
	SharedDirs      map[string]string `json:"shared_dirs"`   // host_path -> vm_path

	// 超时配置
	StartupTimeout  time.Duration    `json:"startup_timeout"`
	ShutdownTimeout time.Duration    `json:"shutdown_timeout"`

	// 资源限制
	MaxProcesses    int              `json:"max_processes"`
	MaxFiles        int              `json:"max_files"`
	MaxMemory       string           `json:"max_memory"`
}

// VMResult 虚拟机操作结果
type VMResult struct {
	Success bool   `json:"success"`
	Output  string `json:"output"`
	Error   string `json:"error,omitempty"`
	Code    int    `json:"code,omitempty"`
}

// ExecResult 执行结果
type ExecResult struct {
	ExitCode int    `json:"exit_code"`
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
}

// NewVMSandbox 创建虚拟机沙箱
func NewVMSandbox(vmType VMType, config VMConfig) (*VMSandbox, error) {
	// 设置默认值
	if config.StartupTimeout == 0 {
		config.StartupTimeout = 5 * time.Minute
	}
	if config.ShutdownTimeout == 0 {
		config.ShutdownTimeout = 2 * time.Minute
	}
	if config.CPUs == 0 {
		config.CPUs = 2
	}
	if config.Memory == "" {
		config.Memory = "4GB"
	}

	sandbox := &VMSandbox{
		config:          config,
		vmType:          vmType,
		status:          VMStatusStopped,
		networkDisabled: config.NetworkDisabled,
	}

	// 检查虚拟机类型是否支持
	if err := sandbox.checkVMTypeSupport(); err != nil {
		return nil, err
	}

	return sandbox, nil
}

// Start 启动虚拟机
func (v *VMSandbox) Start(ctx context.Context) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	if v.status == VMStatusRunning {
		return fmt.Errorf("VM is already running")
	}

	v.status = VMStatusStarting
	v.startTime = time.Now()

	var err error
	switch v.vmType {
	case VMTypeWSL2:
		err = v.startWSL2(ctx)
	case VMTypeLima:
		err = v.startLima(ctx)
	case VMTypeQEMU:
		err = v.startQEMU(ctx)
	case VMTypeDocker:
		err = v.startDocker(ctx)
	default:
		err = fmt.Errorf("unsupported VM type: %s", v.vmType)
	}

	if err != nil {
		v.status = VMStatusError
		return fmt.Errorf("failed to start VM: %w", err)
	}

	v.status = VMStatusRunning
	return nil
}

// Stop 停止虚拟机
func (v *VMSandbox) Stop(ctx context.Context) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	if v.status != VMStatusRunning {
		return fmt.Errorf("VM is not running")
	}

	v.status = VMStatusStopping

	var err error
	switch v.vmType {
	case VMTypeWSL2:
		err = v.stopWSL2(ctx)
	case VMTypeLima:
		err = v.stopLima(ctx)
	case VMTypeQEMU:
		err = v.stopQEMU(ctx)
	case VMTypeDocker:
		err = v.stopDocker(ctx)
	default:
		err = fmt.Errorf("unsupported VM type: %s", v.vmType)
	}

	if err != nil {
		v.status = VMStatusError
		return fmt.Errorf("failed to stop VM: %w", err)
	}

	v.status = VMStatusStopped
	return nil
}

// Exec 在虚拟机中执行命令
func (v *VMSandbox) Exec(ctx context.Context, command string, args ...string) (*ExecResult, error) {
	v.mu.RLock()
	if v.status != VMStatusRunning {
		v.mu.RUnlock()
		return nil, fmt.Errorf("VM is not running")
	}
	v.mu.RUnlock()

	switch v.vmType {
	case VMTypeWSL2:
		return v.execWSL2(ctx, command, args...)
	case VMTypeLima:
		return v.execLima(ctx, command, args...)
	case VMTypeDocker:
		return v.execDocker(ctx, command, args...)
	default:
		return nil, fmt.Errorf("exec not supported for VM type: %s", v.vmType)
	}
}

// CopyToHost 从虚拟机复制文件到主机
func (v *VMSandbox) CopyToHost(ctx context.Context, vmPath, hostPath string) error {
	v.mu.RLock()
	defer v.mu.RUnlock()

	if v.status != VMStatusRunning {
		return fmt.Errorf("VM is not running")
	}

	switch v.vmType {
	case VMTypeWSL2:
		return v.copyToHostWSL2(ctx, vmPath, hostPath)
	case VMTypeLima:
		return v.copyToHostLima(ctx, vmPath, hostPath)
	case VMTypeDocker:
		return v.copyToHostDocker(ctx, vmPath, hostPath)
	default:
		return fmt.Errorf("copy not supported for VM type: %s", v.vmType)
	}
}

// CopyToVM 从主机复制文件到虚拟机
func (v *VMSandbox) CopyToVM(ctx context.Context, hostPath, vmPath string) error {
	v.mu.RLock()
	defer v.mu.RUnlock()

	if v.status != VMStatusRunning {
		return fmt.Errorf("VM is not running")
	}

	switch v.vmType {
	case VMTypeWSL2:
		return v.copyToVMWSL2(ctx, hostPath, vmPath)
	case VMTypeLima:
		return v.copyToVMLima(ctx, hostPath, vmPath)
	case VMTypeDocker:
		return v.copyToVMDocker(ctx, hostPath, vmPath)
	default:
		return fmt.Errorf("copy not supported for VM type: %s", v.vmType)
	}
}

// GetStatus 获取虚拟机状态
func (v *VMSandbox) GetStatus() VMStatus {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.status
}

// GetInfo 获取虚拟机信息
func (v *VMSandbox) GetInfo(ctx context.Context) (map[string]interface{}, error) {
	v.mu.RLock()
	defer v.mu.RUnlock()

	info := map[string]interface{}{
		"name":        v.config.Name,
		"type":        string(v.vmType),
		"status":      string(v.status),
		"cpus":        v.config.CPUs,
		"memory":      v.config.Memory,
		"disk_size":   v.config.DiskSize,
		"os_image":    v.config.OSImage,
		"network_disabled": v.networkDisabled,
	}

	if !v.startTime.IsZero() {
		info["uptime"] = time.Since(v.startTime).String()
	}

	// 获取额外信息
	var err error
	switch v.vmType {
	case VMTypeWSL2:
		extra, e := v.getInfoWSL2(ctx)
		if e == nil {
			for k, val := range extra {
				info[k] = val
			}
		}
	case VMTypeDocker:
		extra, e := v.getInfoDocker(ctx)
		if e == nil {
			for k, val := range extra {
				info[k] = val
			}
		}
	}

	_ = err // 避免未使用警告
	return info, nil
}

// ==================== WSL2 实现 ====================

func (v *VMSandbox) startWSL2(ctx context.Context) error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("WSL2 is only supported on Windows")
	}

	// 启动 WSL2 发行版
	distro := v.config.OSImage
	if distro == "" {
		distro = "Ubuntu-22.04"
	}

	cmd := exec.CommandContext(ctx, "wsl", "--distribution", distro, "--exec", "echo", "ready")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("WSL2 not available or distro not installed: %w", err)
	}

	return nil
}

func (v *VMSandbox) stopWSL2(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "wsl", "--shutdown")
	return cmd.Run()
}

func (v *VMSandbox) execWSL2(ctx context.Context, command string, args ...string) (*ExecResult, error) {
	distro := v.config.OSImage
	if distro == "" {
		distro = "Ubuntu-22.04"
	}

	wslArgs := []string{"--distribution", distro, "--exec", command}
	wslArgs = append(wslArgs, args...)

	cmd := exec.CommandContext(ctx, "wsl", wslArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return &ExecResult{
			ExitCode: cmd.ProcessState.ExitCode(),
			Stdout:   "",
			Stderr:   string(output),
		}, err
	}

	return &ExecResult{
		ExitCode: 0,
		Stdout:   string(output),
		Stderr:   "",
	}, nil
}

func (v *VMSandbox) copyToHostWSL2(ctx context.Context, vmPath, hostPath string) error {
	distro := v.config.OSImage
	if distro == "" {
		distro = "Ubuntu-22.04"
	}

	// WSL2 路径转换
	wslPath := fmt.Sprintf("\\\\wsl$\\%s\\%s", distro, strings.ReplaceAll(vmPath, "/", "\\"))

	// 复制文件
	cmd := exec.CommandContext(ctx, "cmd", "/c", "copy", wslPath, hostPath)
	return cmd.Run()
}

func (v *VMSandbox) copyToVMWSL2(ctx context.Context, hostPath, vmPath string) error {
	distro := v.config.OSImage
	if distro == "" {
		distro = "Ubuntu-22.04"
	}

	// 在 WSL2 中执行 cp 命令
	cmd := exec.CommandContext(ctx, "wsl", "--distribution", distro, "--exec", "cp", "/mnt"+strings.ReplaceAll(hostPath, ":", ""), vmPath)
	return cmd.Run()
}

func (v *VMSandbox) getInfoWSL2(ctx context.Context) (map[string]interface{}, error) {
	info := make(map[string]interface{})

	// 获取 WSL 版本
	cmd := exec.CommandContext(ctx, "wsl", "--version")
	if output, err := cmd.Output(); err == nil {
		info["wsl_version"] = strings.TrimSpace(string(output))
	}

	// 列出发行版
	cmd = exec.CommandContext(ctx, "wsl", "--list", "--verbose")
	if output, err := cmd.Output(); err == nil {
		info["distros"] = strings.TrimSpace(string(output))
	}

	return info, nil
}

// ==================== Lima 实现 ====================

func (v *VMSandbox) startLima(ctx context.Context) error {
	if runtime.GOOS != "darwin" {
		return fmt.Errorf("Lima is only supported on macOS")
	}

	// 检查 lima 是否安装
	if _, err := exec.LookPath("limactl"); err != nil {
		return fmt.Errorf("limactl not found: %w", err)
	}

	// 启动 lima 实例
	name := v.config.Name
	if name == "" {
		name = "default"
	}

	cmd := exec.CommandContext(ctx, "limactl", "start", name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to start lima: %s: %w", string(output), err)
	}

	return nil
}

func (v *VMSandbox) stopLima(ctx context.Context) error {
	name := v.config.Name
	if name == "" {
		name = "default"
	}

	cmd := exec.CommandContext(ctx, "limactl", "stop", name)
	return cmd.Run()
}

func (v *VMSandbox) execLima(ctx context.Context, command string, args ...string) (*ExecResult, error) {
	name := v.config.Name
	if name == "" {
		name = "default"
	}

	limaArgs := []string{"shell", name, command}
	limaArgs = append(limaArgs, args...)

	cmd := exec.CommandContext(ctx, "limactl", limaArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return &ExecResult{
			ExitCode: cmd.ProcessState.ExitCode(),
			Stdout:   "",
			Stderr:   string(output),
		}, err
	}

	return &ExecResult{
		ExitCode: 0,
		Stdout:   string(output),
		Stderr:   "",
	}, nil
}

func (v *VMSandbox) copyToHostLima(ctx context.Context, vmPath, hostPath string) error {
	name := v.config.Name
	if name == "" {
		name = "default"
	}

	cmd := exec.CommandContext(ctx, "limactl", "cp", name+":"+vmPath, hostPath)
	return cmd.Run()
}

func (v *VMSandbox) copyToVMLima(ctx context.Context, hostPath, vmPath string) error {
	name := v.config.Name
	if name == "" {
		name = "default"
	}

	cmd := exec.CommandContext(ctx, "limactl", "cp", hostPath, name+":"+vmPath)
	return cmd.Run()
}

// ==================== QEMU 实现 ====================

func (v *VMSandbox) startQEMU(ctx context.Context) error {
	// QEMU 实现比较复杂，需要更多配置
	// 这里提供基本框架
	return fmt.Errorf("QEMU support not fully implemented")
}

func (v *VMSandbox) stopQEMU(ctx context.Context) error {
	if v.process != nil {
		return v.process.Process.Kill()
	}
	return nil
}

// ==================== Docker 实现 ====================

func (v *VMSandbox) startDocker(ctx context.Context) error {
	// 检查 docker 是否可用
	if _, err := exec.LookPath("docker"); err != nil {
		return fmt.Errorf("docker not found: %w", err)
	}

	containerName := v.config.Name
	if containerName == "" {
		containerName = "agent-sandbox"
	}

	// 创建并启动容器
	args := []string{"run", "-d", "--name", containerName}

	// 添加资源限制
	args = append(args, "--cpus", fmt.Sprintf("%d", v.config.CPUs))
	args = append(args, "--memory", v.config.Memory)

	// 添加安全选项
	if v.config.EnableSeccomp {
		args = append(args, "--security-opt", "seccomp=default")
	}

	// 网络隔离
	if v.networkDisabled {
		args = append(args, "--network", "none")
	}

	// 挂载目录
	for hostPath, vmPath := range v.config.SharedDirs {
		args = append(args, "-v", fmt.Sprintf("%s:%s", hostPath, vmPath))
	}

	// 添加镜像
	image := v.config.OSImage
	if image == "" {
		image = "ubuntu:22.04"
	}
	args = append(args, image)

	// 保持容器运行
	args = append(args, "sleep", "infinity")

	cmd := exec.CommandContext(ctx, "docker", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to start docker container: %s: %w", string(output), err)
	}

	return nil
}

func (v *VMSandbox) stopDocker(ctx context.Context) error {
	containerName := v.config.Name
	if containerName == "" {
		containerName = "agent-sandbox"
	}

	// 停止并删除容器
	cmd := exec.CommandContext(ctx, "docker", "rm", "-f", containerName)
	return cmd.Run()
}

func (v *VMSandbox) execDocker(ctx context.Context, command string, args ...string) (*ExecResult, error) {
	containerName := v.config.Name
	if containerName == "" {
		containerName = "agent-sandbox"
	}

	dockerArgs := []string{"exec", containerName, command}
	dockerArgs = append(dockerArgs, args...)

	cmd := exec.CommandContext(ctx, "docker", dockerArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return &ExecResult{
			ExitCode: cmd.ProcessState.ExitCode(),
			Stdout:   "",
			Stderr:   string(output),
		}, err
	}

	return &ExecResult{
		ExitCode: 0,
		Stdout:   string(output),
		Stderr:   "",
	}, nil
}

func (v *VMSandbox) copyToHostDocker(ctx context.Context, vmPath, hostPath string) error {
	containerName := v.config.Name
	if containerName == "" {
		containerName = "agent-sandbox"
	}

	// 确保主机目录存在
	hostDir := filepath.Dir(hostPath)
	if err := os.MkdirAll(hostDir, 0755); err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, "docker", "cp", containerName+":"+vmPath, hostPath)
	return cmd.Run()
}

func (v *VMSandbox) copyToVMDocker(ctx context.Context, hostPath, vmPath string) error {
	containerName := v.config.Name
	if containerName == "" {
		containerName = "agent-sandbox"
	}

	cmd := exec.CommandContext(ctx, "docker", "cp", hostPath, containerName+":"+vmPath)
	return cmd.Run()
}

func (v *VMSandbox) getInfoDocker(ctx context.Context) (map[string]interface{}, error) {
	info := make(map[string]interface{})

	containerName := v.config.Name
	if containerName == "" {
		containerName = "agent-sandbox"
	}

	// 获取容器信息
	cmd := exec.CommandContext(ctx, "docker", "inspect", containerName)
	output, err := cmd.Output()
	if err != nil {
		return info, nil
	}

	var inspectData []map[string]interface{}
	if err := json.Unmarshal(output, &inspectData); err == nil && len(inspectData) > 0 {
		if config, ok := inspectData[0]["Config"].(map[string]interface{}); ok {
			info["image"] = config["Image"]
		}
		if state, ok := inspectData[0]["State"].(map[string]interface{}); ok {
			info["container_status"] = state["Status"]
		}
	}

	return info, nil
}

// ==================== 辅助函数 ====================

func (v *VMSandbox) checkVMTypeSupport() error {
	switch v.vmType {
	case VMTypeWSL2:
		if runtime.GOOS != "windows" {
			return fmt.Errorf("WSL2 requires Windows")
		}
		if _, err := exec.LookPath("wsl"); err != nil {
			return fmt.Errorf("WSL2 not found: %w", err)
		}

	case VMTypeLima:
		if runtime.GOOS != "darwin" {
			return fmt.Errorf("Lima requires macOS")
		}
		if _, err := exec.LookPath("limactl"); err != nil {
			return fmt.Errorf("limactl not found: %w", err)
		}

	case VMTypeDocker:
		if _, err := exec.LookPath("docker"); err != nil {
			return fmt.Errorf("docker not found: %w", err)
		}

	case VMTypeQEMU:
		if _, err := exec.LookPath("qemu-system-x86_64"); err != nil {
			return fmt.Errorf("qemu not found: %w", err)
		}
	}

	return nil
}

// CreateDefaultWSL2Config 创建默认 WSL2 配置
func CreateDefaultWSL2Config() VMConfig {
	return VMConfig{
		Name:      "agent-sandbox",
		CPUs:      2,
		Memory:    "4GB",
		DiskSize:  "50GB",
		OSImage:   "Ubuntu-22.04",
		StartupTimeout:  5 * time.Minute,
		ShutdownTimeout: 2 * time.Minute,
		NetworkDisabled: false,
		SharedDirs: make(map[string]string),
	}
}

// CreateDefaultDockerConfig 创建默认 Docker 配置
func CreateDefaultDockerConfig() VMConfig {
	return VMConfig{
		Name:      "agent-sandbox",
		CPUs:      2,
		Memory:    "4GB",
		OSImage:   "ubuntu:22.04",
		StartupTimeout:  2 * time.Minute,
		ShutdownTimeout: 1 * time.Minute,
		EnableSeccomp:   true,
		SharedDirs: make(map[string]string),
	}
}

// CreateDefaultLimaConfig 创建默认 Lima 配置
func CreateDefaultLimaConfig() VMConfig {
	return VMConfig{
		Name:      "default",
		CPUs:      4,
		Memory:    "8GB",
		DiskSize:  "100GB",
		OSImage:   "ubuntu:22.04",
		StartupTimeout:  5 * time.Minute,
		ShutdownTimeout: 2 * time.Minute,
		SharedDirs: make(map[string]string),
	}
}

// DetectBestVMType 检测最佳虚拟机类型
func DetectBestVMType() (VMType, error) {
	switch runtime.GOOS {
	case "windows":
		if _, err := exec.LookPath("wsl"); err == nil {
			return VMTypeWSL2, nil
		}
		if _, err := exec.LookPath("docker"); err == nil {
			return VMTypeDocker, nil
		}
		return "", fmt.Errorf("no suitable VM type found on Windows (WSL2 or Docker required)")

	case "darwin":
		if _, err := exec.LookPath("limactl"); err == nil {
			return VMTypeLima, nil
		}
		if _, err := exec.LookPath("docker"); err == nil {
			return VMTypeDocker, nil
		}
		return "", fmt.Errorf("no suitable VM type found on macOS (Lima or Docker required)")

	case "linux":
		if _, err := exec.LookPath("docker"); err == nil {
			return VMTypeDocker, nil
		}
		if _, err := exec.LookPath("qemu-system-x86_64"); err == nil {
			return VMTypeQEMU, nil
		}
		return "", fmt.Errorf("no suitable VM type found on Linux (Docker or QEMU required)")

	default:
		return "", fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}
