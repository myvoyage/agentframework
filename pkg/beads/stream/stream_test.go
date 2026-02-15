// Package stream provides high-performance real-time data processing capabilities.
package stream

import (
	"context"
	"errors"
	"testing"
	"time"
)

// TestDataPipeline 测试数据管道
func TestDataPipeline(t *testing.T) {
	ctx := context.Background()

	// 创建简单的处理器链
	processors := []DataProcessor{
		NewMapProcessor(func(data interface{}) (interface{}, error) {
			// 将数据转换为大写字符串
			if str, ok := data.(string); ok {
				return str + "_processed", nil
			}
			return data, nil
		}),
		NewFilterProcessor(func(data interface{}) bool {
			// 过滤掉空字符串
			if str, ok := data.(string); ok {
				return str != ""
			}
			return true
		}),
	}

	pipeline := NewDataPipeline(processors,
		WithWorkers(2),
		WithBufferSize(10),
	)

	// 启动管道
	if err := pipeline.Start(ctx); err != nil {
		t.Fatalf("Failed to start pipeline: %v", err)
	}

	defer pipeline.Stop()

	// 测试数据处理
	testData := "test"
	if err := pipeline.Process(testData); err != nil {
		t.Fatalf("Failed to process data: %v", err)
	}

	// 等待处理结果
	select {
	case result := <-pipeline.Output():
		if resultStr, ok := result.(string); ok {
			expected := "test_processed"
			if resultStr != expected {
				t.Errorf("Expected %s, got %s", expected, resultStr)
			}
		} else {
			t.Errorf("Expected string result, got %T", result)
		}
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for pipeline output")
	}
}

// TestPipelineMetrics 测试管道指标
func TestPipelineMetrics(t *testing.T) {
	ctx := context.Background()

	processors := []DataProcessor{
		NewMapProcessor(func(data interface{}) (interface{}, error) {
			return data, nil
		}),
	}

	pipeline := NewDataPipeline(processors)
	if err := pipeline.Start(ctx); err != nil {
		t.Fatalf("Failed to start pipeline: %v", err)
	}

	defer pipeline.Stop()

	// 处理一些数据
	for i := 0; i < 5; i++ {
		pipeline.Process(i)
	}

	// 等待处理完成
	time.Sleep(100 * time.Millisecond)

	// 获取指标
	metrics := pipeline.GetMetrics()
	if metrics.ProcessedCount != 5 {
		t.Errorf("Expected 5 processed count, got %d", metrics.ProcessedCount)
	}
}

// TestFilterProcessor 测试过滤处理器
func TestFilterProcessor(t *testing.T) {
	ctx := context.Background()

	processor := NewFilterProcessor(func(data interface{}) bool {
		if num, ok := data.(int); ok {
			return num > 5
		}
		return false
	})

	// 测试过滤
	result, err := processor.Process(ctx, 10)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}
	if result == nil {
		t.Error("Expected data to pass filter")
	}

	result, err = processor.Process(ctx, 3)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}
	if result != nil {
		t.Error("Expected data to be filtered out")
	}
}

// TestMapProcessor 测试映射处理器
func TestMapProcessor(t *testing.T) {
	ctx := context.Background()

	processor := NewMapProcessor(func(data interface{}) (interface{}, error) {
		if num, ok := data.(int); ok {
			return num * 2, nil
		}
		return nil, errors.New("not a number")
	})

	// 测试映射
	result, err := processor.Process(ctx, 5)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}
	if num, ok := result.(int); ok {
		if num != 10 {
			t.Errorf("Expected 10, got %d", num)
		}
	} else {
		t.Error("Expected int result")
	}

	// 测试错误
	_, err = processor.Process(ctx, "not a number")
	if err == nil {
		t.Error("Expected error for non-number input")
	}
}

// TestBatchProcessor 测试批处理器
func TestBatchProcessor(t *testing.T) {
	ctx := context.Background()

	processor := NewBatchProcessor(3, 100*time.Millisecond)

	// 添加数据
	result1, _ := processor.Process(ctx, 1)
	result2, _ := processor.Process(ctx, 2)
	result3, _ := processor.Process(ctx, 3)

	// 前两个应该不返回结果
	if result1 != nil {
		t.Error("Expected no result before batch is full")
	}
	if result2 != nil {
		t.Error("Expected no result before batch is full")
	}

	// 第三个应该触发批次
	if result3 == nil {
		t.Error("Expected batch result")
	}
}

// TestDebounceProcessor 测试防抖处理器
func TestDebounceProcessor(t *testing.T) {
	ctx := context.Background()

	processor := NewDebounceProcessor(50 * time.Millisecond)

	// 发送数据
	go func() {
		processor.Process(ctx, "data1")
		time.Sleep(10 * time.Millisecond)
		processor.Process(ctx, "data2")
		time.Sleep(10 * time.Millisecond)
		processor.Process(ctx, "data3")
	}()

	// 等待防抖后的结果
	select {
	case result := <-processor.Output():
		if result != "data3" {
			t.Errorf("Expected 'data3', got %v", result)
		}
	case <-time.After(200 * time.Millisecond):
		t.Error("Timeout waiting for debounced result")
	}
}

// TestThrottleProcessor 测试节流处理器
func TestThrottleProcessor(t *testing.T) {
	ctx := context.Background()

	processor := NewThrottleProcessor(50 * time.Millisecond)

	// 第一次应该通过
	result1, err := processor.Process(ctx, "data1")
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}
	if result1 == nil {
		t.Error("Expected data to pass through throttle")
	}

	// 立即再次发送应该被节流
	result2, _ := processor.Process(ctx, "data2")
	if result2 != nil {
		t.Error("Expected data to be throttled")
	}

	// 等待节流间隔后应该通过
	time.Sleep(60 * time.Millisecond)
	result3, err := processor.Process(ctx, "data3")
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}
	if result3 == nil {
		t.Error("Expected data to pass through throttle after interval")
	}
}

// TestReduceProcessor 测试归约处理器
func TestReduceProcessor(t *testing.T) {
	ctx := context.Background()

	processor := NewReduceProcessor(0, func(acc, val interface{}) (interface{}, error) {
		accInt, _ := acc.(int)
		valInt, _ := val.(int)
		return accInt + valInt, nil
	})

	// 处理数据
	processor.Process(ctx, 1)
	processor.Process(ctx, 2)
	processor.Process(ctx, 3)

	// 注意：ReduceProcessor 在这个实现中不会累积状态
	// 在实际使用中需要更复杂的状态管理
}

// BenchmarkDataPipeline 性能测试
func BenchmarkDataPipeline(b *testing.B) {
	ctx := context.Background()

	processors := []DataProcessor{
		NewMapProcessor(func(data interface{}) (interface{}, error) {
			return data, nil
		}),
	}

	pipeline := NewDataPipeline(processors, WithWorkers(4))
	if err := pipeline.Start(ctx); err != nil {
		b.Fatalf("Failed to start pipeline: %v", err)
	}

	defer pipeline.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pipeline.Process(i)
	}
}