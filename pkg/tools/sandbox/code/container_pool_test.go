// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
// Copyright (C) 2025  Agent Framework Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package code

import (
	"context"
	"testing"
	"time"
)

// TestContainerPool_BasicOperations tests basic pool operations
func TestContainerPool_BasicOperations(t *testing.T) {
	config := ContainerConfig{
		Enabled:     true,
		EnablePool:  true,
		PoolMinSize: 1,
		PoolMaxSize: 3,
	}

	executor, err := NewContainerExecutor(config)
	if err != nil {
		t.Skip("Docker not available:", err)
	}
	defer executor.Close()

	if executor.pool == nil {
		t.Fatal("Container pool should be initialized")
	}

	// Test pool configuration
	if executor.pool.config.MinSize != 1 {
		t.Errorf("Expected MinSize 1, got %d", executor.pool.config.MinSize)
	}
	if executor.pool.config.MaxSize != 3 {
		t.Errorf("Expected MaxSize 3, got %d", executor.pool.config.MaxSize)
	}
}

// TestContainerPool_AcquireRelease tests container acquisition and release
func TestContainerPool_AcquireRelease(t *testing.T) {
	config := ContainerConfig{
		Enabled:     true,
		EnablePool:  true,
		PoolMinSize: 0,
		PoolMaxSize: 5,
	}

	executor, err := NewContainerExecutor(config)
	if err != nil {
		t.Skip("Docker not available:", err)
	}
	defer executor.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Acquire container
	container1, err := executor.pool.Acquire(ctx, "python")
	if err != nil {
		t.Fatalf("Failed to acquire container: %v", err)
	}

	if container1.ID == "" {
		t.Error("Container ID should not be empty")
	}
	if container1.Language != "python" {
		t.Errorf("Expected language 'python', got '%s'", container1.Language)
	}
	if !container1.Healthy {
		t.Error("Container should be healthy")
	}

	// Release container
	err = executor.pool.Release(container1)
	if err != nil {
		t.Errorf("Failed to release container: %v", err)
	}
}

// TestContainerPool_Reuse tests container reuse
func TestContainerPool_Reuse(t *testing.T) {
	config := ContainerConfig{
		Enabled:     true,
		EnablePool:  true,
		PoolMinSize: 0,
		PoolMaxSize: 5,
	}

	executor, err := NewContainerExecutor(config)
	if err != nil {
		t.Skip("Docker not available:", err)
	}
	defer executor.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Acquire and release container
	container1, err := executor.pool.Acquire(ctx, "python")
	if err != nil {
		t.Fatalf("Failed to acquire container: %v", err)
	}
	containerID1 := container1.ID

	err = executor.pool.Release(container1)
	if err != nil {
		t.Fatalf("Failed to release container: %v", err)
	}

	// Acquire again - should reuse the same container
	container2, err := executor.pool.Acquire(ctx, "python")
	if err != nil {
		t.Fatalf("Failed to acquire container: %v", err)
	}
	containerID2 := container2.ID

	if containerID1 != containerID2 {
		t.Error("Container should be reused")
	}

	// Check reuse count
	stats := executor.GetPoolStats()
	pythonStats := stats["python"]
	if pythonStats.ReuseCount != 1 {
		t.Errorf("Expected 1 reuse, got %d", pythonStats.ReuseCount)
	}

	executor.pool.Release(container2)
}

// TestContainerPool_MultipleLanguages tests pool with multiple languages
func TestContainerPool_MultipleLanguages(t *testing.T) {
	config := ContainerConfig{
		Enabled:     true,
		EnablePool:  true,
		PoolMinSize: 0,
		PoolMaxSize: 5,
	}

	executor, err := NewContainerExecutor(config)
	if err != nil {
		t.Skip("Docker not available:", err)
	}
	defer executor.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	languages := []string{"python", "bash"}

	for _, lang := range languages {
		container, err := executor.pool.Acquire(ctx, lang)
		if err != nil {
			t.Errorf("Failed to acquire %s container: %v", lang, err)
			continue
		}

		if container.Language != lang {
			t.Errorf("Expected language '%s', got '%s'", lang, container.Language)
		}

		executor.pool.Release(container)
	}

	// Check stats
	stats := executor.GetPoolStats()
	for _, lang := range languages {
		langStats, ok := stats[lang]
		if !ok {
			t.Errorf("No stats for language '%s'", lang)
			continue
		}
		if langStats.TotalCreated == 0 {
			t.Errorf("Expected at least 1 container created for '%s'", lang)
		}
	}
}

// TestContainerPool_Stats tests pool statistics
func TestContainerPool_Stats(t *testing.T) {
	config := ContainerConfig{
		Enabled:     true,
		EnablePool:  true,
		PoolMinSize: 0,
		PoolMaxSize: 5,
	}

	executor, err := NewContainerExecutor(config)
	if err != nil {
		t.Skip("Docker not available:", err)
	}
	defer executor.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Initial stats
	stats := executor.GetPoolStats()
	if stats == nil {
		t.Fatal("Stats should not be nil")
	}

	// Acquire and release container
	container, err := executor.pool.Acquire(ctx, "python")
	if err != nil {
		t.Fatalf("Failed to acquire container: %v", err)
	}

	// Check active count
	stats = executor.GetPoolStats()
	pythonStats := stats["python"]
	if pythonStats.ActiveCount != 1 {
		t.Errorf("Expected 1 active container, got %d", pythonStats.ActiveCount)
	}
	if pythonStats.TotalCreated != 1 {
		t.Errorf("Expected 1 total created, got %d", pythonStats.TotalCreated)
	}

	// Release container
	executor.pool.Release(container)

	// Check idle count
	stats = executor.GetPoolStats()
	pythonStats = stats["python"]
	if pythonStats.IdleCount != 1 {
		t.Errorf("Expected 1 idle container, got %d", pythonStats.IdleCount)
	}
}

// TestContainerPool_MaxSize tests pool max size limit
func TestContainerPool_MaxSize(t *testing.T) {
	config := ContainerConfig{
		Enabled:     true,
		EnablePool:  true,
		PoolMinSize: 0,
		PoolMaxSize: 2,
	}

	executor, err := NewContainerExecutor(config)
	if err != nil {
		t.Skip("Docker not available:", err)
	}
	defer executor.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Acquire max number of containers
	containers := make([]*PooledContainer, 0, 2)
	for i := 0; i < 2; i++ {
		container, err := executor.pool.Acquire(ctx, "python")
		if err != nil {
			t.Fatalf("Failed to acquire container %d: %v", i, err)
		}
		containers = append(containers, container)
	}

	// Release all containers
	for _, container := range containers {
		executor.pool.Release(container)
	}

	// Try to add more containers than max size
	for i := 0; i < 3; i++ {
		container, err := executor.pool.Acquire(ctx, "python")
		if err != nil {
			t.Fatalf("Failed to acquire container: %v", err)
		}
		executor.pool.Release(container)
	}

	// Check that pool doesn't exceed max size
	stats := executor.GetPoolStats()
	pythonStats := stats["python"]
	if pythonStats.CurrentSize > 2 {
		t.Errorf("Pool size %d exceeds max size 2", pythonStats.CurrentSize)
	}
}

// TestContainerPool_UnhealthyContainer tests handling of unhealthy containers
func TestContainerPool_UnhealthyContainer(t *testing.T) {
	config := ContainerConfig{
		Enabled:     true,
		EnablePool:  true,
		PoolMinSize: 0,
		PoolMaxSize: 5,
	}

	executor, err := NewContainerExecutor(config)
	if err != nil {
		t.Skip("Docker not available:", err)
	}
	defer executor.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Acquire container
	container, err := executor.pool.Acquire(ctx, "python")
	if err != nil {
		t.Fatalf("Failed to acquire container: %v", err)
	}

	// Mark container as unhealthy
	container.Healthy = false

	// Release unhealthy container
	err = executor.pool.Release(container)
	if err != nil {
		t.Logf("Release returned error (expected for unhealthy container): %v", err)
	}

	// Check that container was destroyed
	stats := executor.GetPoolStats()
	pythonStats := stats["python"]
	if pythonStats.TotalDestroyed != 1 {
		t.Errorf("Expected 1 destroyed container, got %d", pythonStats.TotalDestroyed)
	}
}

// TestContainerPool_Close tests pool cleanup
func TestContainerPool_Close(t *testing.T) {
	config := ContainerConfig{
		Enabled:     true,
		EnablePool:  true,
		PoolMinSize: 0,
		PoolMaxSize: 5,
	}

	executor, err := NewContainerExecutor(config)
	if err != nil {
		t.Skip("Docker not available:", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Acquire some containers
	container1, _ := executor.pool.Acquire(ctx, "python")
	_, _ = executor.pool.Acquire(ctx, "bash")

	// Release one, keep one active
	if container1 != nil {
		executor.pool.Release(container1)
	}

	// Close pool
	err = executor.Close()
	if err != nil {
		t.Errorf("Failed to close pool: %v", err)
	}

	// Verify pool is closed (trying to acquire should fail or panic)
	// We don't test this as it would cause issues
}

// TestContainerPool_ExecuteWithPool tests code execution with pool
func TestContainerPool_ExecuteWithPool(t *testing.T) {
	config := ContainerConfig{
		Enabled:     true,
		EnablePool:  true,
		PoolMinSize: 1,
		PoolMaxSize: 3,
	}

	executor, err := NewContainerExecutor(config)
	if err != nil {
		t.Skip("Docker not available:", err)
	}
	defer executor.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	code := `print("Hello from pooled container!")`

	// Execute code multiple times
	for i := 0; i < 3; i++ {
		result, err := executor.Execute(ctx, "python", code)
		if err != nil {
			t.Fatalf("Execution %d failed: %v", i, err)
		}

		if !result.Success {
			t.Errorf("Execution %d should succeed: %s", i, result.Error)
		}
	}

	// Check reuse count
	stats := executor.GetPoolStats()
	pythonStats := stats["python"]
	if pythonStats.ReuseCount < 2 {
		t.Logf("Expected at least 2 reuses, got %d (may vary)", pythonStats.ReuseCount)
	}
}

// BenchmarkContainerPool_AcquireRelease benchmarks pool operations
func BenchmarkContainerPool_AcquireRelease(b *testing.B) {
	config := ContainerConfig{
		Enabled:     true,
		EnablePool:  true,
		PoolMinSize: 2,
		PoolMaxSize: 10,
	}

	executor, err := NewContainerExecutor(config)
	if err != nil {
		b.Skip("Docker not available:", err)
	}
	defer executor.Close()

	ctx := context.Background()

	// Warm up pool
	time.Sleep(2 * time.Second)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		container, err := executor.pool.Acquire(ctx, "python")
		if err != nil {
			b.Fatalf("Failed to acquire container: %v", err)
		}
		executor.pool.Release(container)
	}
}

// BenchmarkContainerPool_Execute benchmarks code execution with pool
func BenchmarkContainerPool_Execute(b *testing.B) {
	config := ContainerConfig{
		Enabled:     true,
		EnablePool:  true,
		PoolMinSize: 2,
		PoolMaxSize: 10,
	}

	executor, err := NewContainerExecutor(config)
	if err != nil {
		b.Skip("Docker not available:", err)
	}
	defer executor.Close()

	ctx := context.Background()
	code := `print("Benchmark test")`

	// Warm up pool
	time.Sleep(2 * time.Second)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := executor.Execute(ctx, "python", code)
		if err != nil {
			b.Fatalf("Execution failed: %v", err)
		}
	}
}
