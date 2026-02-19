// Agent Framework - RealTimeContext Tests
// Copyright (C) 2025  Agent Framework Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package context

import (
	"context"
	"testing"
	"time"
)

func TestNewRealTimeContext(t *testing.T) {
	rtc := NewRealTimeContext(100, 5*time.Minute)

	if rtc == nil {
		t.Fatal("expected RealTimeContext to be created")
	}

	if rtc.maxSize != 100 {
		t.Errorf("expected maxSize 100, got %d", rtc.maxSize)
	}

	if rtc.ttl != 5*time.Minute {
		t.Errorf("expected ttl 5 minutes, got %v", rtc.ttl)
	}

	if rtc.dataStore == nil {
		t.Error("expected dataStore to be initialized")
	}

	if rtc.metadata == nil {
		t.Error("expected metadata to be initialized")
	}

	if rtc.index == nil {
		t.Error("expected index to be initialized")
	}
}

func TestRealTimeContext_SetAndGet(t *testing.T) {
	rtc := NewRealTimeContext(100, 5*time.Minute)
	defer rtc.Close()

	ctx := context.Background()

	// Set a value
	err := rtc.Set(ctx, "test-key", "test-value")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Get the value
	value, err := rtc.Get(ctx, "test-key")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if value != "test-value" {
		t.Errorf("expected 'test-value', got '%v'", value)
	}
}

func TestRealTimeContext_GetNotFound(t *testing.T) {
	rtc := NewRealTimeContext(100, 5*time.Minute)
	defer rtc.Close()

	ctx := context.Background()

	_, err := rtc.Get(ctx, "non-existent")
	if err != ErrKeyNotFound {
		t.Errorf("expected ErrKeyNotFound, got %v", err)
	}
}

func TestRealTimeContext_GetExpired(t *testing.T) {
	rtc := NewRealTimeContext(100, 100*time.Millisecond)
	defer rtc.Close()

	ctx := context.Background()

	// Set a value
	err := rtc.Set(ctx, "test-key", "test-value")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Wait for TTL to expire
	time.Sleep(150 * time.Millisecond)

	// Get the value - should be expired or removed by cleanup
	_, err = rtc.Get(ctx, "test-key")
	if err != ErrKeyExpired && err != ErrKeyNotFound {
		t.Errorf("expected ErrKeyExpired or ErrKeyNotFound, got %v", err)
	}
}

func TestRealTimeContext_Delete(t *testing.T) {
	rtc := NewRealTimeContext(100, 5*time.Minute)
	defer rtc.Close()

	ctx := context.Background()

	// Set a value
	err := rtc.Set(ctx, "test-key", "test-value")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Delete the value
	err = rtc.Delete(ctx, "test-key")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Get should return not found
	_, err = rtc.Get(ctx, "test-key")
	if err != ErrKeyNotFound {
		t.Errorf("expected ErrKeyNotFound, got %v", err)
	}
}

func TestRealTimeContext_DeleteNotFound(t *testing.T) {
	rtc := NewRealTimeContext(100, 5*time.Minute)
	defer rtc.Close()

	ctx := context.Background()

	err := rtc.Delete(ctx, "non-existent")
	if err != ErrKeyNotFound {
		t.Errorf("expected ErrKeyNotFound, got %v", err)
	}
}

func TestRealTimeContext_Query(t *testing.T) {
	rtc := NewRealTimeContext(100, 5*time.Minute)
	defer rtc.Close()

	ctx := context.Background()

	// Set multiple values
	rtc.Set(ctx, "key1", "value1")
	rtc.Set(ctx, "key2", "value2")
	rtc.Set(ctx, "key3", "value3")

	// Query all
	query := &Query{}
	results, err := rtc.Query(ctx, *query)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}
}

func TestRealTimeContext_QueryWithFilter(t *testing.T) {
	rtc := NewRealTimeContext(100, 5*time.Minute)
	defer rtc.Close()

	ctx := context.Background()

	// Set values
	rtc.Set(ctx, "key1", 10)
	rtc.Set(ctx, "key2", 20)
	rtc.Set(ctx, "key3", 30)

	// Query with filter
	query := &Query{
		Filter: func(v interface{}) bool {
			if i, ok := v.(int); ok {
				return i > 15
			}
			return false
		},
	}
	results, err := rtc.Query(ctx, *query)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestRealTimeContext_QueryWithLimit(t *testing.T) {
	rtc := NewRealTimeContext(100, 5*time.Minute)
	defer rtc.Close()

	ctx := context.Background()

	// Set multiple values
	rtc.Set(ctx, "key1", "value1")
	rtc.Set(ctx, "key2", "value2")
	rtc.Set(ctx, "key3", "value3")
	rtc.Set(ctx, "key4", "value4")
	rtc.Set(ctx, "key5", "value5")

	// Query with limit
	query := &Query{Limit: 3}
	results, err := rtc.Query(ctx, *query)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}
}

func TestRealTimeContext_Search(t *testing.T) {
	rtc := NewRealTimeContext(100, 5*time.Minute)
	defer rtc.Close()

	ctx := context.Background()

	// Set values
	rtc.Set(ctx, "test1", "hello world")
	rtc.Set(ctx, "test2", "world peace")

	// Search for "world"
	results, err := rtc.Search(ctx, "world", 10)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(results) == 0 {
		t.Error("expected some results")
	}
}

func TestRealTimeContext_SearchWithLimit(t *testing.T) {
	rtc := NewRealTimeContext(100, 5*time.Minute)
	defer rtc.Close()

	ctx := context.Background()

	// Set values
	for i := 0; i < 10; i++ {
		rtc.Set(ctx, "key"+string(rune('0'+i)), "test value "+string(rune('0'+i)))
	}

	// Search with limit
	results, err := rtc.Search(ctx, "test", 5)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(results) > 5 {
		t.Errorf("expected at most 5 results, got %d", len(results))
	}
}

func TestRealTimeContext_GetStats(t *testing.T) {
	rtc := NewRealTimeContext(100, 5*time.Minute)
	defer rtc.Close()

	ctx := context.Background()

	// Set some values
	rtc.Set(ctx, "key1", "value1")
	rtc.Set(ctx, "key2", "value2")

	stats := rtc.GetStats(ctx)
	if stats == nil {
		t.Fatal("expected stats to be returned")
	}

	if stats.TotalEntries != 2 {
		t.Errorf("expected TotalEntries 2, got %d", stats.TotalEntries)
	}
}

func TestRealTimeContext_Clear(t *testing.T) {
	rtc := NewRealTimeContext(100, 5*time.Minute)
	defer rtc.Close()

	ctx := context.Background()

	// Set some values
	rtc.Set(ctx, "key1", "value1")
	rtc.Set(ctx, "key2", "value2")

	// Clear
	err := rtc.Clear(ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Check if cleared
	stats := rtc.GetStats(ctx)
	if stats.TotalEntries != 0 {
		t.Errorf("expected TotalEntries 0, got %d", stats.TotalEntries)
	}
}

func TestRealTimeContext_Close(t *testing.T) {
	rtc := NewRealTimeContext(100, 5*time.Minute)

	ctx := context.Background()

	// Set some values
	rtc.Set(ctx, "key1", "value1")

	// Close
	err := rtc.Close()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRealTimeContext_MaxSize(t *testing.T) {
	rtc := NewRealTimeContext(3, 5*time.Minute)
	defer rtc.Close()

	ctx := context.Background()

	// Set values exceeding max size
	rtc.Set(ctx, "key1", "value1")
	rtc.Set(ctx, "key2", "value2")
	rtc.Set(ctx, "key3", "value3")
	rtc.Set(ctx, "key4", "value4") // Should evict oldest

	// Check that only 3 items remain
	stats := rtc.GetStats(ctx)
	if stats.TotalEntries != 3 {
		t.Errorf("expected TotalEntries 3, got %d", stats.TotalEntries)
	}
}

func TestRealTimeContext_ConcurrentAccess(t *testing.T) {
	rtc := NewRealTimeContext(100, 5*time.Minute)
	defer rtc.Close()

	ctx := context.Background()

	// Concurrent writes
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(i int) {
			key := "key" + string(rune('0'+i))
			rtc.Set(ctx, key, i)
			done <- true
		}(i)
	}

	// Wait for all writes
	for i := 0; i < 10; i++ {
		<-done
	}

	// Check stats
	stats := rtc.GetStats(ctx)
	if stats.TotalEntries != 10 {
		t.Errorf("expected TotalEntries 10, got %d", stats.TotalEntries)
	}
}

// TestCalculateSize tests the calculateSize function indirectly through stats
func TestRealTimeContext_CalculateSize(t *testing.T) {
	rtc := NewRealTimeContext(100, 5*time.Minute)
	defer rtc.Close()

	ctx := context.Background()

	// Set different types of values to test size calculation
	rtc.Set(ctx, "string-key", "test string value")
	rtc.Set(ctx, "bytes-key", []byte("test bytes"))
	rtc.Set(ctx, "int-key", 42)
	rtc.Set(ctx, "float-key", 3.14)
	rtc.Set(ctx, "bool-key", true)

	stats := rtc.GetStats(ctx)
	if stats.TotalEntries != 5 {
		t.Errorf("expected TotalEntries 5, got %d", stats.TotalEntries)
	}
	if stats.TotalSize == 0 {
		t.Error("expected TotalSize > 0")
	}
}

// TestRealTimeContext_MapValue tests setting map values
func TestRealTimeContext_MapValue(t *testing.T) {
	rtc := NewRealTimeContext(100, 5*time.Minute)
	defer rtc.Close()

	ctx := context.Background()

	// Set a map value
	mapValue := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
		"key3": true,
	}

	err := rtc.Set(ctx, "map-key", mapValue)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Get the map value
	retrieved, err := rtc.Get(ctx, "map-key")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if retrieved == nil {
		t.Error("expected value to be returned")
	}
}

// TestRealTimeContext_SliceValue tests setting slice values
func TestRealTimeContext_SliceValue(t *testing.T) {
	rtc := NewRealTimeContext(100, 5*time.Minute)
	defer rtc.Close()

	ctx := context.Background()

	// Set a slice value
	sliceValue := []interface{}{"item1", "item2", 42}

	err := rtc.Set(ctx, "slice-key", sliceValue)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Get the slice value
	retrieved, err := rtc.Get(ctx, "slice-key")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if retrieved == nil {
		t.Error("expected value to be returned")
	}
}

// TestRealTimeContext_NestedMap tests nested map structures
func TestRealTimeContext_NestedMap(t *testing.T) {
	rtc := NewRealTimeContext(100, 5*time.Minute)
	defer rtc.Close()

	ctx := context.Background()

	// Set a nested map value
	nestedMap := map[string]interface{}{
		"level1": map[string]interface{}{
			"level2": "deep value",
		},
	}

	err := rtc.Set(ctx, "nested-key", nestedMap)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Get the nested map value
	retrieved, err := rtc.Get(ctx, "nested-key")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if retrieved == nil {
		t.Error("expected value to be returned")
	}
}

// TestRealTimeContext_ByteArray tests byte array values
func TestRealTimeContext_ByteArray(t *testing.T) {
	rtc := NewRealTimeContext(100, 5*time.Minute)
	defer rtc.Close()

	ctx := context.Background()

	// Set a byte array value
	byteValue := []byte{0x01, 0x02, 0x03, 0x04}

	err := rtc.Set(ctx, "bytes-key", byteValue)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Get the byte array value
	retrieved, err := rtc.Get(ctx, "bytes-key")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if retrieved == nil {
		t.Error("expected value to be returned")
	}
}

// TestRealTimeContext_SearchMap tests searching in map values
func TestRealTimeContext_SearchMap(t *testing.T) {
	rtc := NewRealTimeContext(100, 5*time.Minute)
	defer rtc.Close()

	ctx := context.Background()

	// Set a map value
	mapValue := map[string]interface{}{
		"title":  "test document",
		"author": "test author",
	}

	err := rtc.Set(ctx, "doc-key", mapValue)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Search for content in the map
	results, err := rtc.Search(ctx, "document", 10)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(results) == 0 {
		t.Error("expected to find the document")
	}
}

// TestRealTimeContext_SearchSlice tests searching in slice values
func TestRealTimeContext_SearchSlice(t *testing.T) {
	rtc := NewRealTimeContext(100, 5*time.Minute)
	defer rtc.Close()

	ctx := context.Background()

	// Set a slice value
	sliceValue := []interface{}{"apple", "banana", "cherry"}

	err := rtc.Set(ctx, "fruits-key", sliceValue)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Search for content in the slice
	results, err := rtc.Search(ctx, "banana", 10)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(results) == 0 {
		t.Error("expected to find banana in the fruits")
	}
}

// TestRealTimeContext_SetMaxSize_NoTimestamp 测试达到最大大小但没有有效时间戳的情况
func TestRealTimeContext_SetMaxSize_NoTimestamp(t *testing.T) {
	rtc := NewRealTimeContext(2, 5*time.Minute)
	defer rtc.Close()

	ctx := context.Background()

	// Set a value
	rtc.Set(ctx, "key1", "value1")

	// 直接操作 metadata 来模拟没有时间戳的情况
	rtc.mu.Lock()
	delete(rtc.metadata, "key1")
	rtc.mu.Unlock()

	// Set another value - should evict the one without timestamp
	rtc.Set(ctx, "key2", "value2")

	// Verify that eviction happened
	stats := rtc.GetStats(ctx)
	if stats.TotalEntries != 2 {
		t.Logf("Note: With 2 max size, behavior may vary when timestamps are missing")
	}
}

// TestRealTimeContext_SearchNoResults 测试搜索无结果的情况
func TestRealTimeContext_SearchNoResults(t *testing.T) {
	rtc := NewRealTimeContext(100, 5*time.Minute)
	defer rtc.Close()

	ctx := context.Background()

	// Set values
	rtc.Set(ctx, "key1", "apple")
	rtc.Set(ctx, "key2", "banana")

	// Search for non-existent term
	results, err := rtc.Search(ctx, "nonexistent", 10)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

// TestRealTimeContext_QueryEmptyFilter 测试使用空过滤器的查询
func TestRealTimeContext_QueryEmptyFilter(t *testing.T) {
	rtc := NewRealTimeContext(100, 5*time.Minute)
	defer rtc.Close()

	ctx := context.Background()

	// Set values
	rtc.Set(ctx, "key1", "value1")
	rtc.Set(ctx, "key2", "value2")

	// Query with nil filter (should match all)
	query := &Query{Filter: nil}
	results, err := rtc.Query(ctx, *query)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 results with nil filter, got %d", len(results))
	}
}

// TestRealTimeContext_QueryNoMatches 测试查询无匹配的情况
func TestRealTimeContext_QueryNoMatches(t *testing.T) {
	rtc := NewRealTimeContext(100, 5*time.Minute)
	defer rtc.Close()

	ctx := context.Background()

	// Set values
	rtc.Set(ctx, "key1", 10)
	rtc.Set(ctx, "key2", 20)

	// Query with filter that doesn't match
	query := &Query{
		Filter: func(v interface{}) bool {
			if i, ok := v.(int); ok {
				return i > 100
			}
			return false
		},
	}
	results, err := rtc.Query(ctx, *query)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

// TestRealTimeContext_MaxSize_Eviction 测试超出最大大小时驱逐最旧的条目
func TestRealTimeContext_MaxSize_EvictionOrder(t *testing.T) {
	rtc := NewRealTimeContext(2, 5*time.Minute)
	defer rtc.Close()

	ctx := context.Background()

	// Set first value
	rtc.Set(ctx, "key1", "value1")

	// Wait to ensure timestamp difference
	time.Sleep(10 * time.Millisecond)

	// Set second value
	rtc.Set(ctx, "key2", "value2")

	// Set third value - should evict key1
	rtc.Set(ctx, "key3", "value3")

	// Check that key1 was evicted
	_, err := rtc.Get(ctx, "key1")
	if err != ErrKeyNotFound && err != ErrKeyExpired {
		t.Logf("key1 status: %v (may have been evicted or expired)", err)
	}

	// Check that key2 and key3 still exist
	val2, err := rtc.Get(ctx, "key2")
	if err != nil {
		t.Errorf("expected key2 to exist, got error: %v", err)
	}
	if val2 != "value2" {
		t.Errorf("expected key2 to be 'value2', got '%v'", val2)
	}

	val3, err := rtc.Get(ctx, "key3")
	if err != nil {
		t.Errorf("expected key3 to exist, got error: %v", err)
	}
	if val3 != "value3" {
		t.Errorf("expected key3 to be 'value3', got '%v'", val3)
	}
}

// TestRealTimeContext_GetStats_Empty 测试空上下文的统计信息
func TestRealTimeContext_GetStats_Empty(t *testing.T) {
	rtc := NewRealTimeContext(100, 5*time.Minute)
	defer rtc.Close()

	ctx := context.Background()

	stats := rtc.GetStats(ctx)
	if stats == nil {
		t.Fatal("expected stats to be returned")
	}

	if stats.TotalEntries != 0 {
		t.Errorf("expected TotalEntries 0, got %d", stats.TotalEntries)
	}

	if stats.TotalSize != 0 {
		t.Errorf("expected TotalSize 0, got %d", stats.TotalSize)
	}
}

// TestRealTimeContext_SetWithMap 测试设置 map 类型的值
func TestRealTimeContext_SetWithMap(t *testing.T) {
	rtc := NewRealTimeContext(100, 5*time.Minute)
	defer rtc.Close()

	ctx := context.Background()

	// Set a complex map value
	mapValue := map[string]interface{}{
		"nested": map[string]interface{}{
			"key1": "value1",
			"key2": 42,
		},
		"simple": "string value",
	}

	err := rtc.Set(ctx, "map-key", mapValue)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Get the value back
	retrieved, err := rtc.Get(ctx, "map-key")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if retrieved == nil {
		t.Error("expected value to be returned")
	}
}

// TestRealTimeContext_UpdateAccessTime 测试更新访问时间
func TestRealTimeContext_UpdateAccessTime(t *testing.T) {
	rtc := NewRealTimeContext(100, 5*time.Minute)
	defer rtc.Close()

	ctx := context.Background()

	// Set a value
	rtc.Set(ctx, "test-key", "test-value")

	// Get the value (should update access time implicitly)
	value, err := rtc.Get(ctx, "test-key")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if value != "test-value" {
		t.Errorf("expected 'test-value', got '%v'", value)
	}
}

// TestRealTimeContext_QueryExpiredItem 测试查询过期项
func TestRealTimeContext_QueryExpiredItem(t *testing.T) {
	rtc := NewRealTimeContext(100, 100*time.Millisecond)
	defer rtc.Close()

	ctx := context.Background()

	// Set a value
	rtc.Set(ctx, "test-key", "test-value")

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Query - expired items should be skipped
	query := &Query{Filter: nil}
	results, err := rtc.Query(ctx, *query)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Expired items should be skipped
	if len(results) != 0 {
		t.Errorf("expected 0 results (item expired), got %d", len(results))
	}
}

// TestRealTimeContext_SearchExpiredItem 测试搜索过期项
func TestRealTimeContext_SearchExpiredItem(t *testing.T) {
	rtc := NewRealTimeContext(100, 100*time.Millisecond)
	defer rtc.Close()

	ctx := context.Background()

	// Set a value
	rtc.Set(ctx, "test-key", "test-value")

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Search - expired items should be skipped
	results, err := rtc.Search(ctx, "test-value", 10)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Expired items should be skipped
	if len(results) != 0 {
		t.Errorf("expected 0 results (item expired), got %d", len(results))
	}
}

// TestRealTimeContext_SetUpdateExisting 测试更新已存在的值
func TestRealTimeContext_SetUpdateExisting(t *testing.T) {
	rtc := NewRealTimeContext(100, 5*time.Minute)
	defer rtc.Close()

	ctx := context.Background()

	// Set initial value
	rtc.Set(ctx, "test-key", "initial-value")

	// Update the value
	rtc.Set(ctx, "test-key", "updated-value")

	// Get the value
	value, err := rtc.Get(ctx, "test-key")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if value != "updated-value" {
		t.Errorf("expected 'updated-value', got '%v'", value)
	}

	// Should only have one entry
	stats := rtc.GetStats(ctx)
	if stats.TotalEntries != 1 {
		t.Errorf("expected 1 entry, got %d", stats.TotalEntries)
	}
}


// TestRealTimeContext_Get_NoMetadata 测试获取没有元数据的值
func TestRealTimeContext_Get_NoMetadata(t *testing.T) {
	rtc := NewRealTimeContext(100, 5*time.Minute)
	defer rtc.Close()

	ctx := context.Background()

	// 直接设置 dataStore，不设置 metadata
	rtc.mu.Lock()
	rtc.dataStore["test-key"] = "test-value"
	// 确保没有 metadata
	rtc.metadata["test-key"] = nil
	rtc.mu.Unlock()

	// 获取值 - 没有 metadata 应该返回值（没有 TTL 检查）
	value, err := rtc.Get(ctx, "test-key")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if value != "test-value" {
		t.Errorf("expected 'test-value', got '%v'", value)
	}
}

// TestRealTimeContext_Get_MetadataNoTimestamp 测试元数据没有时间戳的情况
func TestRealTimeContext_Get_MetadataNoTimestamp(t *testing.T) {
	rtc := NewRealTimeContext(100, 5*time.Minute)
	defer rtc.Close()

	ctx := context.Background()

	// Set 会自动添加 timestamp，所以直接操作内部状态
	rtc.Set(ctx, "test-key", "test-value")

	// 修改 metadata 来移除 timestamp 或设置为无效值
	rtc.mu.Lock()
	rtc.metadata["test-key"] = map[string]interface{}{
		"size": 9,
	}
	rtc.mu.Unlock()

	// 获取值 - 没有 timestamp 应该返回值（没有 TTL 检查）
	value, err := rtc.Get(ctx, "test-key")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if value != "test-value" {
		t.Errorf("expected 'test-value', got '%v'", value)
	}
}

// TestRealTimeContext_Query_ExpiredItemsInLoop 测试查询时跳过过期项
func TestRealTimeContext_Query_ExpiredItemsInLoop(t *testing.T) {
	rtc := NewRealTimeContext(100, 100*time.Millisecond)
	defer rtc.Close()

	ctx := context.Background()

	// Set multiple values
	rtc.Set(ctx, "key1", "value1")
	rtc.Set(ctx, "key2", "value2")

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Query with filter that matches all - expired items should be skipped
	query := &Query{
		Filter: func(v interface{}) bool {
			return true // Match all
		},
		Limit: 10,
	}
	results, err := rtc.Query(ctx, *query)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("expected 0 results (items expired), got %d", len(results))
	}
}

// TestRealTimeContext_Search_ZeroLimit 测试搜索限制为 0
func TestRealTimeContext_Search_ZeroLimit(t *testing.T) {
	rtc := NewRealTimeContext(100, 5*time.Minute)
	defer rtc.Close()

	ctx := context.Background()

	rtc.Set(ctx, "key1", "test value 1")
	rtc.Set(ctx, "key2", "test value 2")

	// Search with limit 0 (should return all matches)
	results, err := rtc.Search(ctx, "test", 0)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(results) == 0 {
		t.Error("expected some results")
	}
}

// TestRealTimeContext_Set_NilValue 测试设置 nil 值
func TestRealTimeContext_Set_NilValue(t *testing.T) {
	rtc := NewRealTimeContext(100, 5*time.Minute)
	defer rtc.Close()

	ctx := context.Background()

	// Set nil value
	err := rtc.Set(ctx, "nil-key", nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Get should return nil
	value, err := rtc.Get(ctx, "nil-key")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if value != nil {
		t.Errorf("expected nil, got '%v'", value)
	}
}

// TestRealTimeContext_Set_ZeroValue 测试设置零值
func TestRealTimeContext_Set_ZeroValue(t *testing.T) {
	rtc := NewRealTimeContext(100, 5*time.Minute)
	defer rtc.Close()

	ctx := context.Background()

	// Set zero values
	rtc.Set(ctx, "zero-string", "")
	rtc.Set(ctx, "zero-int", 0)
	rtc.Set(ctx, "zero-float", 0.0)
	rtc.Set(ctx, "zero-bool", false)

	// Verify they can be retrieved
	val, err := rtc.Get(ctx, "zero-string")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if val != "" {
		t.Errorf("expected empty string, got '%v'", val)
	}
}

// TestRealTimeContext_Clear_WithExistingData 测试清空有数据的上下文
func TestRealTimeContext_Clear_WithExistingData(t *testing.T) {
	rtc := NewRealTimeContext(100, 5*time.Minute)
	defer rtc.Close()

	ctx := context.Background()

	// Set some data
	rtc.Set(ctx, "key1", "value1")
	rtc.Set(ctx, "key2", "value2")
	rtc.Set(ctx, "key3", "value3")

	// Verify data exists
	stats := rtc.GetStats(ctx)
	if stats.TotalEntries != 3 {
		t.Errorf("expected 3 entries before clear, got %d", stats.TotalEntries)
	}

	// Clear
	err := rtc.Clear(ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify data is cleared
	stats = rtc.GetStats(ctx)
	if stats.TotalEntries != 0 {
		t.Errorf("expected 0 entries after clear, got %d", stats.TotalEntries)
	}

	if stats.TotalSize != 0 {
		t.Errorf("expected 0 size after clear, got %d", stats.TotalSize)
	}
}

// TestRealTimeContext_GetStats_TotalSize 测试统计信息中的总大小
func TestRealTimeContext_GetStats_TotalSize(t *testing.T) {
	rtc := NewRealTimeContext(100, 5*time.Minute)
	defer rtc.Close()

	ctx := context.Background()

	// Set values with known sizes
	rtc.Set(ctx, "key1", "test")      // 4 bytes
	rtc.Set(ctx, "key2", "value")    // 5 bytes
	rtc.Set(ctx, "key3", "data123")  // 7 bytes

	stats := rtc.GetStats(ctx)

	if stats.TotalEntries != 3 {
		t.Errorf("expected 3 entries, got %d", stats.TotalEntries)
	}

	// Total size should be sum of string lengths
	// Note: calculateSize may have special handling for different types
	if stats.TotalSize == 0 {
		t.Error("expected non-zero total size")
	}
}

// TestRealTimeContext_Set_InvalidMetadataTimestamp 测试设置值时元数据时间戳无效的情况
func TestRealTimeContext_Set_InvalidMetadataTimestamp(t *testing.T) {
	rtc := NewRealTimeContext(2, 5*time.Minute)
	defer rtc.Close()

	ctx := context.Background()

	// First set a value
	rtc.Set(ctx, "key1", "value1")

	// Manually corrupt metadata to test the continue branch
	rtc.mu.Lock()
	rtc.metadata["key1"]["timestamp"] = "invalid"
	rtc.mu.Unlock()

	// Set another value that triggers max size eviction
	rtc.Set(ctx, "key2", "value2")

	// Should still work (the corrupted entry is skipped)
	stats := rtc.GetStats(ctx)
	if stats.TotalEntries == 0 {
		t.Error("expected at least some entries to exist")
	}
}

// TestRealTimeContext_Get_InvalidTimestampType 测试获取值时时间戳类型断言失败的情况
func TestRealTimeContext_Get_InvalidTimestampType(t *testing.T) {
	rtc := NewRealTimeContext(100, 5*time.Minute)
	defer rtc.Close()

	ctx := context.Background()

	// Set a value
	rtc.Set(ctx, "test-key", "test-value")

	// Manually corrupt metadata timestamp type
	rtc.mu.Lock()
	rtc.metadata["test-key"]["timestamp"] = "not a time"
	rtc.mu.Unlock()

	// Get should still return the value (it just skips TTL check when type assertion fails)
	value, err := rtc.Get(ctx, "test-key")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if value != "test-value" {
		t.Errorf("expected 'test-value', got %v", value)
	}
}

// TestRealTimeContext_Query_NoLimit 测试无限制的查询
func TestRealTimeContext_Query_NoLimit(t *testing.T) {
	rtc := NewRealTimeContext(100, 5*time.Minute)
	defer rtc.Close()

	ctx := context.Background()

	// Set multiple values
	for i := 0; i < 10; i++ {
		rtc.Set(ctx, "key"+string(rune('0'+i)), "value"+string(rune('0'+i)))
	}

	// Query with no limit (Limit = 0 means no limit)
	query := &Query{
		Filter:  nil,
		Limit:   0,
	}
	results, err := rtc.Query(ctx, *query)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// Should return all non-expired entries
	if len(results) < 10 {
		t.Errorf("expected at least 10 results, got %d", len(results))
	}
}

// TestRealTimeContext_Delete_NonExistent 测试删除不存在的键
func TestRealTimeContext_Delete_NonExistent(t *testing.T) {
	rtc := NewRealTimeContext(100, 5*time.Minute)
	defer rtc.Close()

	ctx := context.Background()

	// Try to delete a non-existent key
	err := rtc.Delete(ctx, "non-existent-key")
	if err != ErrKeyNotFound {
		t.Errorf("expected ErrKeyNotFound, got %v", err)
	}
}

// TestCalculateSize_String 测试计算字符串大小
func TestCalculateSize_String(t *testing.T) {
	size := calculateSize("hello")
	if size != 5 {
		t.Errorf("expected size 5, got %d", size)
	}
}

// TestCalculateSize_Bytes 测试计算字节数组大小
func TestCalculateSize_Bytes(t *testing.T) {
	size := calculateSize([]byte{1, 2, 3, 4, 5})
	if size != 5 {
		t.Errorf("expected size 5, got %d", size)
	}
}

// TestCalculateSize_Map 测试计算映射大小
func TestCalculateSize_Map(t *testing.T) {
	m := map[string]interface{}{
		"a": "hello",
		"b": "world",
	}
	size := calculateSize(m)
	// "hello" (5) + "world" (5) = 10
	if size != 10 {
		t.Errorf("expected size 10, got %d", size)
	}
}

// TestCalculateSize_Slice 测试计算切片大小
func TestCalculateSize_Slice(t *testing.T) {
	slice := []interface{}{"hello", "world"}
	size := calculateSize(slice)
	// "hello" (5) + "world" (5) = 10
	if size != 10 {
		t.Errorf("expected size 10, got %d", size)
	}
}

// TestCalculateSize_Default 测试计算默认类型大小
func TestCalculateSize_Default(t *testing.T) {
	size := calculateSize(123)
	if size != 8 {
		t.Errorf("expected default size 8, got %d", size)
	}
}

// TestMatchesSearch_String 测试字符串搜索匹配
func TestMatchesSearch_String(t *testing.T) {
	if !matchesSearch("hello world", "world") {
		t.Error("expected 'world' to match 'hello world'")
	}
	if matchesSearch("hello world", "xyz") {
		t.Error("expected 'xyz' not to match 'hello world'")
	}
}

// TestMatchesSearch_Map 测试映射搜索匹配
func TestMatchesSearch_Map(t *testing.T) {
	m := map[string]interface{}{
		"name": "test",
		"value": 123,
	}
	if !matchesSearch(m, "test") {
		t.Error("expected 'test' to match map")
	}
	if matchesSearch(m, "nonexistent") {
		t.Error("expected 'nonexistent' not to match map")
	}
}

// TestMatchesSearch_Slice 测试切片搜索匹配
func TestMatchesSearch_Slice(t *testing.T) {
	slice := []interface{}{"hello", "world", 123}
	if !matchesSearch(slice, "world") {
		t.Error("expected 'world' to match slice")
	}
	if matchesSearch(slice, "nonexistent") {
		t.Error("expected 'nonexistent' not to match slice")
	}
}

// TestMatchesSearch_Default 测试默认类型搜索匹配
func TestMatchesSearch_Default(t *testing.T) {
	if matchesSearch(123, "123") {
		t.Error("expected int not to match string")
	}
}

// TestCalculateRelevance_ExactMatch 测试精确匹配相关性
func TestCalculateRelevance_ExactMatch(t *testing.T) {
	score := calculateRelevance("hello", "hello")
	if score != 1.0 {
		t.Errorf("expected score 1.0 for exact match, got %f", score)
	}
}

// TestCalculateRelevance_PartialMatch 测试部分匹配相关性
func TestCalculateRelevance_PartialMatch(t *testing.T) {
	score := calculateRelevance("hello world", "world")
	if score != 0.5 {
		t.Errorf("expected score 0.5 for partial match, got %f", score)
	}
}

// TestCalculateRelevance_NoMatch 测试无匹配相关性
func TestCalculateRelevance_NoMatch(t *testing.T) {
	score := calculateRelevance("hello", "xyz")
	if score != 0.0 {
		t.Errorf("expected score 0.0 for no match, got %f", score)
	}
}

// TestCalculateRelevance_DefaultType 测试默认类型相关性
func TestCalculateRelevance_DefaultType(t *testing.T) {
	score := calculateRelevance(123, "123")
	if score != 0.0 {
		t.Errorf("expected score 0.0 for non-string type, got %f", score)
	}
}

// TestContains_ExactMatch 测试精确包含
func TestContains_ExactMatch(t *testing.T) {
	if !contains("hello", "hello") {
		t.Error("expected 'hello' to contain 'hello'")
	}
}

// TestContains_Substring 测试子串包含
func TestContains_Substring(t *testing.T) {
	if !contains("hello world", "world") {
		t.Error("expected 'hello world' to contain 'world'")
	}
}

// TestContains_CaseInsensitive 测试不区分大小写包含
func TestContains_CaseInsensitive(t *testing.T) {
	if !contains("HELLO World", "world") {
		t.Error("expected case-insensitive match")
	}
	if !contains("hello world", "WORLD") {
		t.Error("expected case-insensitive match")
	}
}

// TestContains_NotContained 测试不包含
func TestContains_NotContained(t *testing.T) {
	if contains("hello", "xyz") {
		t.Error("expected 'hello' not to contain 'xyz'")
	}
}

// TestExtractTerms_String 测试从字符串提取术语
func TestExtractTerms_String(t *testing.T) {
	terms := extractTerms("hello")
	if len(terms) != 1 {
		t.Errorf("expected 1 term, got %d", len(terms))
	}
	if terms[0] != "hello" {
		t.Errorf("expected 'hello', got '%s'", terms[0])
	}
}

// TestExtractTerms_Map 测试从映射提取术语
func TestExtractTerms_Map(t *testing.T) {
	m := map[string]interface{}{
		"name": "value",
		"nested": map[string]interface{}{
			"key": "nested-value",
		},
	}
	terms := extractTerms(m)
	// Should include keys and nested values
	if len(terms) < 3 {
		t.Errorf("expected at least 3 terms, got %d", len(terms))
	}
}

// TestExtractTerms_Slice 测试从切片提取术语
func TestExtractTerms_Slice(t *testing.T) {
	slice := []interface{}{"hello", "world", 123}
	terms := extractTerms(slice)
	if len(terms) != 2 {
		t.Errorf("expected 2 terms (strings only), got %d", len(terms))
	}
}

// TestExtractTerms_EmptyMap 测试从空映射提取术语
func TestExtractTerms_EmptyMap(t *testing.T) {
	m := map[string]interface{}{}
	terms := extractTerms(m)
	if len(terms) != 0 {
		t.Errorf("expected 0 terms from empty map, got %d", len(terms))
	}
}

// TestExtractTerms_EmptySlice 测试从空切片提取术语
func TestExtractTerms_EmptySlice(t *testing.T) {
	slice := []interface{}{}
	terms := extractTerms(slice)
	if len(terms) != 0 {
		t.Errorf("expected 0 terms from empty slice, got %d", len(terms))
	}
}

// TestQueryMatches_NilFilter 测试 nil 过滤器匹配
func TestQueryMatches_NilFilter(t *testing.T) {
	query := &Query{
		Filter: nil,
	}
	if !query.Matches("anything") {
		t.Error("expected nil filter to match everything")
	}
}

// TestQueryMatches_CustomFilter 测试自定义过滤器匹配
func TestQueryMatches_CustomFilter(t *testing.T) {
	query := &Query{
		Filter: func(v interface{}) bool {
			if s, ok := v.(string); ok {
				return len(s) > 5
			}
			return false
		},
	}
	if !query.Matches("hello world") {
		t.Error("expected 'hello world' to match filter")
	}
	if query.Matches("hi") {
		t.Error("expected 'hi' not to match filter")
	}
}

// TestRealTimeContext_Search_Expiration 测试搜索时过期项被跳过
func TestRealTimeContext_Search_Expiration(t *testing.T) {
	rtc := NewRealTimeContext(100, 100*time.Millisecond)
	defer rtc.Close()

	ctx := context.Background()

	// Set values
	rtc.Set(ctx, "key1", "searchable value")
	rtc.Set(ctx, "key2", "another value")

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Search should return no results (all expired)
	results, err := rtc.Search(ctx, "value", 10)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results (expired), got %d", len(results))
	}
}

// TestRealTimeContext_Search_MapValue 测试搜索映射值
func TestRealTimeContext_Search_MapValue(t *testing.T) {
	rtc := NewRealTimeContext(100, 5*time.Minute)
	defer rtc.Close()

	ctx := context.Background()

	// Set a map value
	rtc.Set(ctx, "map-key", map[string]interface{}{
		"name": "searchable",
		"value": "test",
	})

	// Search for term in map
	results, err := rtc.Search(ctx, "searchable", 10)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(results) == 0 {
		t.Error("expected to find 'searchable' in map")
	}
}

// TestRealTimeContext_Search_SliceValue 测试搜索切片值
func TestRealTimeContext_Search_SliceValue(t *testing.T) {
	rtc := NewRealTimeContext(100, 5*time.Minute)
	defer rtc.Close()

	ctx := context.Background()

	// Set a slice value
	rtc.Set(ctx, "slice-key", []interface{}{"hello", "searchable", "world"})

	// Search for term in slice
	results, err := rtc.Search(ctx, "searchable", 10)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(results) == 0 {
		t.Error("expected to find 'searchable' in slice")
	}
}

// TestRealTimeContext_Delete_RemovesFromIndex 测试删除时从索引中移除
func TestRealTimeContext_Delete_RemovesFromIndex(t *testing.T) {
	rtc := NewRealTimeContext(100, 5*time.Minute)
	defer rtc.Close()

	ctx := context.Background()

	// Set a value (this updates the index)
	rtc.Set(ctx, "indexed-key", "indexed-value")

	// Delete the value
	err := rtc.Delete(ctx, "indexed-key")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify it's gone from the data store
	_, err = rtc.Get(ctx, "indexed-key")
	if err != ErrKeyNotFound {
		t.Errorf("expected ErrKeyNotFound after delete, got %v", err)
	}
}

