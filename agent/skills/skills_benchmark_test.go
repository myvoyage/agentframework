// Agent Framework - Skill System Benchmark Tests
// Copyright (C) 2025  Agent Framework Contributors

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// SPDX-License-Identifier: AGPL-3.0-or-later

package skills

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
)

// BenchmarkRegistryRegister benchmarks skill registration
func BenchmarkRegistryRegister(b *testing.B) {
	registry := NewSkillRegistry(&RegistryConfig{
		BaseDir:  ".bench_registry",
		AutoSave: false,
	})
	defer os.RemoveAll(".bench_registry")

	entry := &SkillEntry{
		ID:          "bench_skill",
		Name:        "Benchmark Skill",
		Description: "Used for benchmarking",
		Category:    "bench",
		Tags:        []string{"bench", "test"},
		InputSchema: &Schema{
			Type:       "object",
			Properties: map[string]*PropertyInfo{},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		entry.ID = fmt.Sprintf("skill_%d", i)
		registry.Register(entry)
	}
}

// BenchmarkRegistryQuery benchmarks skill queries
func BenchmarkRegistryQuery(b *testing.B) {
	registry := NewSkillRegistry(&RegistryConfig{
		BaseDir:  ".bench_registry",
		AutoSave: false,
	})
	defer os.RemoveAll(".bench_registry")

	// Register 1000 skills
	for i := 0; i < 1000; i++ {
		entry := &SkillEntry{
			ID:       fmt.Sprintf("skill_%d", i),
			Name:     fmt.Sprintf("Skill %d", i),
			Category: []string{"cat1", "cat2", "cat3"}[i%3],
			Tags:     []string{"tag1", "tag2", "tag3"},
		}
		registry.Register(entry)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		registry.GetByID(fmt.Sprintf("skill_%d", i%1000))
	}
}

// BenchmarkRegistryFind benchmarks skill find operations
func BenchmarkRegistryFind(b *testing.B) {
	registry := NewSkillRegistry(&RegistryConfig{
		BaseDir:  ".bench_registry",
		AutoSave: false,
	})
	defer os.RemoveAll(".bench_registry")

	// Register 1000 skills
	for i := 0; i < 1000; i++ {
		entry := &SkillEntry{
			ID:       fmt.Sprintf("skill_%d", i),
			Name:     fmt.Sprintf("Skill %d", i),
			Category: []string{"cat1", "cat2", "cat3"}[i%3],
			Tags:     []string{"tag1", "tag2", "tag3"},
		}
		registry.Register(entry)
	}

	query := &SkillQuery{
		Category: "cat1",
		Tags:     []string{"tag1"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		registry.Find(query)
	}
}

// BenchmarkCacheGet benchmarks cache get operations
func BenchmarkCacheGet(b *testing.B) {
	cache := NewMemoryCache(1024*1024, 1000, time.Minute)

	// Populate cache
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key_%d", i)
		data := fmt.Sprintf("data_%d", i)
		cache.Set(key, data, time.Minute)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get(fmt.Sprintf("key_%d", i%1000))
	}
}

// BenchmarkCacheSet benchmarks cache set operations
func BenchmarkCacheSet(b *testing.B) {
	cache := NewMemoryCache(1024*1024, 1000, time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key_%d", i%1000)
		data := fmt.Sprintf("data_%d", i)
		cache.Set(key, data, time.Minute)
	}
}

// BenchmarkMultiLevelCache benchmarks multi-level cache operations
func BenchmarkMultiLevelCache(b *testing.B) {
	cache := NewMultiLevelCache(DefaultCacheConfig())
	defer cache.Close()

	// Populate cache
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key_%d", i)
		data := fmt.Sprintf("data_%d", i)
		cache.Set(key, data, time.Minute)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get(fmt.Sprintf("key_%d", i%100))
	}
}

// BenchmarkExecutorPool benchmarks executor pool operations
func BenchmarkExecutorPool(b *testing.B) {
	config := &PoolConfig{
		MinConnections: 2,
		MaxConnections: 10,
		IdleTimeout:    5 * time.Minute,
		MaxLifetime:    30 * time.Minute,
		AcquireTimeout: 30 * time.Second,
	}

	pool, err := NewConnectionPool(config, func() (interface{}, error) {
		return &EnhancedSkillExecutor{}, nil
	}, nil)
	if err != nil {
		b.Fatal(err)
	}
	defer pool.Close()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		conn, _ := pool.Acquire(ctx)
		pool.Release(conn)
	}
}

// BenchmarkDefinitionLoad benchmarks definition loading
func BenchmarkDefinitionLoad(b *testing.B) {
	manager := NewDefinitionManager("agent/skills/definitions")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.Load("http_request")
	}
}

// BenchmarkLoaderMetadata benchmarks metadata loading
func BenchmarkLoaderMetadata(b *testing.B) {
	loader := NewProgressiveLoader("agent/skills/definitions")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		loader.ListSkills()
	}
}

// BenchmarkFullExecution benchmarks full skill execution
func BenchmarkFullExecution(b *testing.B) {
	ctx := context.Background()
	executor := NewEnhancedSkillExecutor(&DefaultExecutorConfig)

	// Setup test skill
	def := &SkillDefinition{
		ID:   "test_skill",
		Name: "Test Skill",
		Workflow: []WorkflowStep{
			{
				ID:     "step1",
				Name:   "Step 1",
				Action: "validate",
			},
			{
				ID:     "step2",
				Name:   "Step 2",
				Action: "execute",
			},
		},
	}

	executor.SetDefinition(def)

	input := `{"test": "data"}`
	execCtx := NewExecutionContext()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		executor.Execute(ctx, input, execCtx)
	}
}

// TestPerformanceReport generates a performance report
func TestPerformanceReport(t *testing.T) {
	fmt.Println("=== Skill System Performance Report ===")

	// Test Registry Performance
	fmt.Println("1. Registry Performance:")
	registry := NewSkillRegistry(&RegistryConfig{
		BaseDir:  ".bench_registry",
		AutoSave: false,
	})
	defer os.RemoveAll(".bench_registry")

	start := time.Now()
	for i := 0; i < 1000; i++ {
		entry := &SkillEntry{
			ID:       fmt.Sprintf("skill_%d", i),
			Name:     fmt.Sprintf("Skill %d", i),
			Category: "test",
		}
		registry.Register(entry)
	}
	fmt.Printf("   Register 1000 skills: %v\n", time.Since(start))

	start = time.Now()
	registry.Find(&SkillQuery{Category: "test"})
	fmt.Printf("   Query skills: %v\n", time.Since(start))

	// Test Cache Performance
	fmt.Println("\n2. Cache Performance:")
	cache := NewMemoryCache(1024*1024, 1000, time.Minute)

	start = time.Now()
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key_%d", i)
		data := fmt.Sprintf("data_%d", i)
		cache.Set(key, data, time.Minute)
	}
	fmt.Printf("   Cache 1000 items: %v\n", time.Since(start))

	start = time.Now()
	for i := 0; i < 1000; i++ {
		cache.Get(fmt.Sprintf("key_%d", i))
	}
	fmt.Printf("   Retrieve 1000 items: %v\n", time.Since(start))

	stats := cache.GetStats()
	fmt.Printf("   Cache hit rate: %.2f%%\n", stats["hit_rate"].(float64)*100)

	// Test Multi-level Cache
	fmt.Println("\n3. Multi-Level Cache Performance:")
	mlCache := NewMultiLevelCache(DefaultCacheConfig())
	defer mlCache.Close()

	start = time.Now()
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key_%d", i)
		data := fmt.Sprintf("data_%d", i)
		mlCache.Set(key, data, time.Minute)
	}
	fmt.Printf("   Cache 100 items: %v\n", time.Since(start))

	start = time.Now()
	for i := 0; i < 1000; i++ {
		mlCache.Get(fmt.Sprintf("key_%d", i%100))
	}
	fmt.Printf("   Retrieve 1000 times: %v\n", time.Since(start))

	stats = mlCache.GetStats()
	fmt.Printf("   Cache hit rate: %.2f%%\n", stats["hit_rate"].(float64)*100)

	// Test Loader Performance
	fmt.Println("\n4. Progressive Loader Performance:")
	loader := NewProgressiveLoader("agent/skills/definitions")

	start = time.Now()
	metas, _ := loader.ListSkills()
	fmt.Printf("   Load metadata for %d skills: %v\n", len(metas), time.Since(start))

	start = time.Now()
	for _, meta := range metas {
		loader.LoadSkill(meta.ID)
	}
	fmt.Printf("   Load full definitions: %v\n", time.Since(start))

	fmt.Println("\n=== Performance Report Complete ===")
}
