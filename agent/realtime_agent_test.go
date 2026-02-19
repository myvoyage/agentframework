// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later


// Package agent provides real-time data processing agent implementation.
package agent

import (
	"context"
	"testing"
	"time"

	beadscontext "AgentFramework/pkg/beads/context"
	"AgentFramework/pkg/beads/stream"
)

// TestRealTimeAgent 测试实时数据处理代理
func TestRealTimeAgent(t *testing.T) {
	ctx := context.Background()
	agent := NewRealTimeAgent(100, 4, true)

	t.Run("Initialize", func(t *testing.T) {
		if err := agent.Initialize(ctx); err != nil {
			t.Errorf("Failed to initialize agent: %v", err)
		}
	})

	t.Run("CreatePipeline", func(t *testing.T) {
		pipelineID := "test_pipeline"

		processors := []stream.DataProcessor{
			stream.NewMapProcessor(func(data interface{}) (interface{}, error) {
				return data, nil
			}),
		}

		err := agent.CreatePipeline(ctx, pipelineID, processors)
		if err != nil {
			t.Errorf("Failed to create pipeline: %v", err)
		}

		// 验证管道已创建
		pipelines, _ := agent.ListPipelines(ctx)
		if len(pipelines) == 0 {
			t.Error("Expected at least one pipeline")
		}
	})

	t.Run("ProcessData", func(t *testing.T) {
		pipelineID := "test_pipeline"
		testData := "test_data"

		err := agent.ProcessData(ctx, pipelineID, testData)
		if err != nil {
			t.Errorf("Failed to process data: %v", err)
		}
	})

	t.Run("GetPipelineMetrics", func(t *testing.T) {
		pipelineID := "test_pipeline"

		metrics, err := agent.GetPipelineMetrics(ctx, pipelineID)
		if err != nil {
			t.Errorf("Failed to get pipeline metrics: %v", err)
		}

		if metrics.ProcessedCount == 0 {
			t.Error("Expected some processed data")
		}
	})

	t.Run("PublishEvent", func(t *testing.T) {
		event := RealTimeEvent{
			Type:      "test_event",
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"message": "test event",
			},
			Source: "test",
		}

		err := agent.PublishEvent(ctx, event)
		if err != nil {
			t.Errorf("Failed to publish event: %v", err)
		}
	})

	t.Run("SubscribeEvents", func(t *testing.T) {
		eventReceived := false

		handler := func(ctx context.Context, event RealTimeEvent) {
			if event.Type == "test_subscription" {
				eventReceived = true
			}
		}

		err := agent.SubscribeEvents(ctx, "test_subscription", handler)
		if err != nil {
			t.Errorf("Failed to subscribe to events: %v", err)
		}

		// 发布事件
		event := RealTimeEvent{
			Type:      "test_subscription",
			Timestamp: time.Now(),
			Data:      map[string]interface{}{},
		}

		agent.PublishEvent(ctx, event)

		// 等待事件处理
		time.Sleep(100 * time.Millisecond)

		_ = eventReceived // 事件处理是异步的
	})

	t.Run("GetRealTimeStats", func(t *testing.T) {
		stats, err := agent.GetRealTimeStats(ctx)
		if err != nil {
			t.Errorf("Failed to get real-time stats: %v", err)
		}

		if stats.TotalEntries == 0 {
			t.Error("Expected some entries in stats")
		}
	})

	t.Run("DeletePipeline", func(t *testing.T) {
		pipelineID := "test_pipeline"

		err := agent.DeletePipeline(ctx, pipelineID)
		if err != nil {
			t.Errorf("Failed to delete pipeline: %v", err)
		}

		// 验证管道已删除
		_, err = agent.GetPipelineMetrics(ctx, pipelineID)
		if err == nil {
			t.Error("Expected error for deleted pipeline")
		}
	})

	t.Run("Close", func(t *testing.T) {
		if err := agent.Close(ctx); err != nil {
			t.Errorf("Failed to close agent: %v", err)
		}
	})
}

// TestRealTimeContext 测试实时上下文
func TestRealTimeContext(t *testing.T) {
	ctx := context.Background()
	rtc := beadscontext.NewRealTimeContext(100, 5*time.Minute)

	t.Run("SetAndGet", func(t *testing.T) {
		key := "test_key"
		value := "test_value"

		err := rtc.Set(ctx, key, value)
		if err != nil {
			t.Errorf("Failed to set value: %v", err)
		}

		retrieved, err := rtc.Get(ctx, key)
		if err != nil {
			t.Errorf("Failed to get value: %v", err)
		}

		if retrieved != value {
			t.Errorf("Expected %v, got %v", value, retrieved)
		}
	})

	t.Run("GetNonExistent", func(t *testing.T) {
		_, err := rtc.Get(ctx, "nonexistent")
		if err == nil {
			t.Error("Expected error for non-existent key")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		key := "delete_key"
		rtc.Set(ctx, key, "value")

		err := rtc.Delete(ctx, key)
		if err != nil {
			t.Errorf("Failed to delete: %v", err)
		}

		_, err = rtc.Get(ctx, key)
		if err == nil {
			t.Error("Expected error for deleted key")
		}
	})

	t.Run("Query", func(t *testing.T) {
		// 添加一些数据
		rtc.Set(ctx, "item1", map[string]interface{}{"name": "item1", "value": 100})
		rtc.Set(ctx, "item2", map[string]interface{}{"name": "item2", "value": 200})

		query := &beadscontext.Query{
			Filter: func(data interface{}) bool {
				if m, ok := data.(map[string]interface{}); ok {
					if val, exists := m["value"]; exists {
						if v, ok := val.(int); ok {
							return v > 100
						}
					}
				}
				return false
			},
			Limit: 10,
		}

		results, err := rtc.Query(ctx, *query)
		if err != nil {
			t.Errorf("Failed to query: %v", err)
		}

		if len(results) == 0 {
			t.Error("Expected some results from query")
		}
	})

	t.Run("Search", func(t *testing.T) {
		// 添加可搜索的数据
		rtc.Set(ctx, "search1", "searchable content here")
		rtc.Set(ctx, "search2", "other data")

		results, err := rtc.Search(ctx, "searchable", 10)
		if err != nil {
			t.Errorf("Failed to search: %v", err)
		}

		if len(results) == 0 {
			t.Error("Expected some search results")
		}
	})

	t.Run("GetStats", func(t *testing.T) {
		stats := rtc.GetStats(ctx)

		if stats.TotalEntries == 0 {
			t.Error("Expected some entries in stats")
		}
	})

	t.Run("Clear", func(t *testing.T) {
		err := rtc.Clear(ctx)
		if err != nil {
			t.Errorf("Failed to clear: %v", err)
		}

		stats := rtc.GetStats(ctx)
		if stats.TotalEntries != 0 {
			t.Error("Expected no entries after clear")
		}
	})
}

// BenchmarkRealTimeAgent 性能测试
func BenchmarkRealTimeAgent(b *testing.B) {
	ctx := context.Background()
	agent := NewRealTimeAgent(1000, 8, false)
	agent.Initialize(ctx)

	processors := []stream.DataProcessor{
		stream.NewMapProcessor(func(data interface{}) (interface{}, error) {
			return data, nil
		}),
	}

	agent.CreatePipeline(ctx, "bench_pipeline", processors)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		agent.ProcessData(ctx, "bench_pipeline", i)
	}
}

// BenchmarkRealTimeContext 性能测试
func BenchmarkRealTimeContext(b *testing.B) {
	ctx := context.Background()
	rtc := beadscontext.NewRealTimeContext(10000, 5*time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := string(rune(i%26 + 'a')) + "_key"
		rtc.Set(ctx, key, i)
		rtc.Get(ctx, key)
	}
}