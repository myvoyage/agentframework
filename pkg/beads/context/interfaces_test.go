// Agent Framework - Context Interfaces Tests
// Copyright (C) 2025  Agent Framework Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package context

import (
	"testing"
	"time"
)

func TestWithSearchLayer(t *testing.T) {
	opts := &SearchOptions{}

	WithSearchLayer(LayerTypeL0)(opts)

	if opts.Layer != LayerTypeL0 {
		t.Errorf("expected LayerTypeL0, got %s", opts.Layer)
	}
}

func TestWithMaxResults(t *testing.T) {
	opts := &SearchOptions{}

	WithMaxResults(10)(opts)

	if opts.MaxResults != 10 {
		t.Errorf("expected MaxResults 10, got %d", opts.MaxResults)
	}
}

func TestWithMinScore(t *testing.T) {
	opts := &SearchOptions{}

	WithMinScore(0.8)(opts)

	if opts.MinScore != 0.8 {
		t.Errorf("expected MinScore 0.8, got %f", opts.MinScore)
	}
}

func TestWithContextType(t *testing.T) {
	opts := &SearchOptions{}

	WithContextType(ContextTypeFile)(opts)

	if opts.ContextType != ContextTypeFile {
		t.Errorf("expected ContextTypeFile, got %s", opts.ContextType)
	}
}

func TestWithWorkspace(t *testing.T) {
	opts := &SearchOptions{}

	WithWorkspace("/test/workspace")(opts)

	if opts.Workspace != "/test/workspace" {
		t.Errorf("expected '/test/workspace', got %s", opts.Workspace)
	}
}

func TestDefaultMemoryCompressionConfig(t *testing.T) {
	config := DefaultMemoryCompressionConfig()

	if config == nil {
		t.Fatal("expected config to be returned")
	}

	// Check retention durations
	if config.SessionRetentionDuration != 24*time.Hour {
		t.Errorf("expected SessionRetentionDuration 24h, got %v", config.SessionRetentionDuration)
	}
	if config.DailyRetentionDuration != 7*24*time.Hour {
		t.Errorf("expected DailyRetentionDuration 7*24h, got %v", config.DailyRetentionDuration)
	}
	if config.LongTermRetentionDuration != 365*24*time.Hour {
		t.Errorf("expected LongTermRetentionDuration 365*24h, got %v", config.LongTermRetentionDuration)
	}

	// Check max memories
	if config.MaxSessionMemories != 100 {
		t.Errorf("expected MaxSessionMemories 100, got %d", config.MaxSessionMemories)
	}
	if config.MaxDailyMemories != 500 {
		t.Errorf("expected MaxDailyMemories 500, got %d", config.MaxDailyMemories)
	}
	if config.MaxLongTermMemories != 1000 {
		t.Errorf("expected MaxLongTermMemories 1000, got %d", config.MaxLongTermMemories)
	}

	// Check compression threshold
	if config.CompressionThreshold != 0.3 {
		t.Errorf("expected CompressionThreshold 0.3, got %f", config.CompressionThreshold)
	}

	// Check compression interval
	if config.CompressionInterval != 1*time.Hour {
		t.Errorf("expected CompressionInterval 1h, got %v", config.CompressionInterval)
	}

	// Check LLM config
	if config.ModelName != "gpt-4" {
		t.Errorf("expected ModelName 'gpt-4', got %s", config.ModelName)
	}
	if config.MaxTokens != 2000 {
		t.Errorf("expected MaxTokens 2000, got %d", config.MaxTokens)
	}
	if config.Temperature != 0.3 {
		t.Errorf("expected Temperature 0.3, got %f", config.Temperature)
	}

	// Check preserve flags
	if !config.PreserveSystemMemories {
		t.Error("expected PreserveSystemMemories to be true")
	}
	if !config.EnableAsyncCompression {
		t.Error("expected EnableAsyncCompression to be true")
	}
}

func TestSearchOptions(t *testing.T) {
	opts := &SearchOptions{}

	// Apply multiple options
	WithSearchLayer(LayerTypeL1)(opts)
	WithMaxResults(20)(opts)
	WithMinScore(0.7)(opts)
	WithContextType(ContextTypeCodebase)(opts)
	WithWorkspace("/my/workspace")(opts)

	if opts.Layer != LayerTypeL1 {
		t.Errorf("expected LayerTypeL1, got %s", opts.Layer)
	}
	if opts.MaxResults != 20 {
		t.Errorf("expected MaxResults 20, got %d", opts.MaxResults)
	}
	if opts.MinScore != 0.7 {
		t.Errorf("expected MinScore 0.7, got %f", opts.MinScore)
	}
	if opts.ContextType != ContextTypeCodebase {
		t.Errorf("expected ContextTypeCodebase, got %s", opts.ContextType)
	}
	if opts.Workspace != "/my/workspace" {
		t.Errorf("expected '/my/workspace', got %s", opts.Workspace)
	}
}

func TestLayerTypeValues(t *testing.T) {
	// Test that layer type constants are properly defined
	layerTypes := []LayerType{
		LayerTypeL0,
		LayerTypeL1,
		LayerTypeL2,
		LayerAuto,
	}

	expectedValues := []string{"l0", "l1", "l2", "auto"}

	for i, lt := range layerTypes {
		if string(lt) != expectedValues[i] {
			t.Errorf("expected %s, got %s", expectedValues[i], lt)
		}
	}
}

func TestContextTypeValues(t *testing.T) {
	// Test that context type constants are properly defined
	contextTypes := []ContextType{
		ContextTypeProject,
		ContextTypeFile,
		ContextTypeCodebase,
		ContextTypeConversation,
		ContextTypeSession,
	}

	expectedValues := []string{"project", "file", "codebase", "conversation", "session"}

	for i, ct := range contextTypes {
		if string(ct) != expectedValues[i] {
			t.Errorf("expected %s, got %s", expectedValues[i], ct)
		}
	}
}

func TestMemoryTierValues(t *testing.T) {
	// Test that memory tier constants are properly defined
	tiers := []MemoryTier{
		MemoryTierSession,
		MemoryTierDaily,
		MemoryTierLongTerm,
	}

	expectedValues := []string{"session", "daily", "longterm"}

	for i, mt := range tiers {
		if string(mt) != expectedValues[i] {
			t.Errorf("expected %s, got %s", expectedValues[i], mt)
		}
	}
}

func TestMemoryTypeValues(t *testing.T) {
	// Test that memory type constants are properly defined
	memoryTypes := []MemoryType{
		MemoryTypeProfile,
		MemoryTypePreference,
		MemoryTypeEntity,
		MemoryTypeEvent,
		MemoryTypeCase,
		MemoryTypePattern,
	}

	expectedValues := []string{"profile", "preference", "entity", "event", "case", "pattern"}

	for i, mt := range memoryTypes {
		if string(mt) != expectedValues[i] {
			t.Errorf("expected %s, got %s", expectedValues[i], mt)
		}
	}
}

func TestContextEventTypeValues(t *testing.T) {
	// Test that context event type constants are properly defined
	eventTypes := []ContextEventType{
		ContextEventCreated,
		ContextEventUpdated,
		ContextEventDeleted,
		ContextEventAssociated,
		ContextEventDissociated,
	}

	expectedValues := []string{"context.created", "context.updated", "context.deleted", "context.associated", "context.dissociated"}

	for i, et := range eventTypes {
		if string(et) != expectedValues[i] {
			t.Errorf("expected %s, got %s", expectedValues[i], et)
		}
	}
}

func TestMemoryDeduplicationValues(t *testing.T) {
	// Test that memory deduplication constants are properly defined
	dedupTypes := []MemoryDeduplication{
		DedupDecisionCreate,
		DedupDecisionUpdate,
		DedupDecisionMerge,
		DedupDecisionSkip,
	}

	expectedValues := []string{"create", "update", "merge", "skip"}

	for i, dt := range dedupTypes {
		if string(dt) != expectedValues[i] {
			t.Errorf("expected %s, got %s", expectedValues[i], dt)
		}
	}
}

func TestContextFilter(t *testing.T) {
	// Test ContextFilter struct
	filter := &ContextFilter{
		ContextID: stringPtr("test-id"),
		Type:      contextTypePtr(ContextTypeFile),
		Workspace: stringPtr("/test/workspace"),
		Metadata: map[string]string{
			"key1": "value1",
		},
	}

	if *filter.ContextID != "test-id" {
		t.Errorf("expected ContextID 'test-id', got '%s'", *filter.ContextID)
	}
	if *filter.Type != ContextTypeFile {
		t.Errorf("expected Type ContextTypeFile, got '%s'", *filter.Type)
	}
	if *filter.Workspace != "/test/workspace" {
		t.Errorf("expected Workspace '/test/workspace', got '%s'", *filter.Workspace)
	}
	if filter.Metadata["key1"] != "value1" {
		t.Errorf("expected metadata key1 to be 'value1', got '%s'", filter.Metadata["key1"])
	}
}

func TestTaskWithContext(t *testing.T) {
	// Test TaskWithContext struct
	twc := &TaskWithContext{
		Task: nil, // Would be beads.Task in real usage
		Contexts: []*Context{
			{ID: "context-1", Title: "Context 1"},
			{ID: "context-2", Title: "Context 2"},
		},
	}

	if len(twc.Contexts) != 2 {
		t.Errorf("expected 2 contexts, got %d", len(twc.Contexts))
	}
	if twc.Contexts[0].ID != "context-1" {
		t.Errorf("expected context ID 'context-1', got '%s'", twc.Contexts[0].ID)
	}
}

func TestContextStoreConfig(t *testing.T) {
	// Test ContextStoreConfig struct
	config := &ContextStoreConfig{
		Type:    "memory",
		Enabled: true,
		Config: map[string]interface{}{
			"max_size": 1000,
		},
	}

	if config.Type != "memory" {
		t.Errorf("expected Type 'memory', got '%s'", config.Type)
	}
	if !config.Enabled {
		t.Error("expected Enabled to be true")
	}
	if config.Config["max_size"] != 1000 {
		t.Errorf("expected max_size 1000, got %v", config.Config["max_size"])
	}
}

func TestContextEvent(t *testing.T) {
	// Test ContextEvent struct
	event := &ContextEvent{
		Type:      ContextEventCreated,
		ContextID: "test-context-id",
		TaskID:    "test-task-id",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"key1": "value1",
			"key2": 123,
		},
	}

	if event.Type != ContextEventCreated {
		t.Errorf("expected Type ContextEventCreated, got '%s'", event.Type)
	}
	if event.ContextID != "test-context-id" {
		t.Errorf("expected ContextID 'test-context-id', got '%s'", event.ContextID)
	}
	if event.TaskID != "test-task-id" {
		t.Errorf("expected TaskID 'test-task-id', got '%s'", event.TaskID)
	}
	if event.Data["key1"] != "value1" {
		t.Errorf("expected Data key1 to be 'value1', got '%v'", event.Data["key1"])
	}
}

func TestContextStoreStats(t *testing.T) {
	// Test ContextStoreStats struct
	stats := &ContextStoreStats{
		TotalContexts: 100,
		TotalTasks:     50,
		ByType: map[ContextType]int64{
			ContextTypeFile:     30,
			ContextTypeProject:  20,
		},
		StorageSize:   1000000,
		CacheHitRate:  0.85,
		AvgAccessTime: 100 * time.Millisecond,
	}

	if stats.TotalContexts != 100 {
		t.Errorf("expected TotalContexts 100, got %d", stats.TotalContexts)
	}
	if stats.TotalTasks != 50 {
		t.Errorf("expected TotalTasks 50, got %d", stats.TotalTasks)
	}
	if stats.ByType[ContextTypeFile] != 30 {
		t.Errorf("expected 30 file contexts, got %d", stats.ByType[ContextTypeFile])
	}
	if stats.StorageSize != 1000000 {
		t.Errorf("expected StorageSize 1000000, got %d", stats.StorageSize)
	}
	if stats.CacheHitRate != 0.85 {
		t.Errorf("expected CacheHitRate 0.85, got %f", stats.CacheHitRate)
	}
}

func TestVFSPath(t *testing.T) {
	// Test VFSPath struct
	path := &VFSPath{
		Scheme:    "viking",
		Workspace: "test-workspace",
		Path:      "/test/path",
		Layer:     LayerTypeL0,
		Query: map[string]string{
			"param1": "value1",
		},
	}

	if path.Scheme != "viking" {
		t.Errorf("expected Scheme 'viking', got '%s'", path.Scheme)
	}
	if path.Workspace != "test-workspace" {
		t.Errorf("expected Workspace 'test-workspace', got '%s'", path.Workspace)
	}
	if path.Path != "/test/path" {
		t.Errorf("expected Path '/test/path', got '%s'", path.Path)
	}
	if path.Layer != LayerTypeL0 {
		t.Errorf("expected Layer LayerTypeL0, got '%s'", path.Layer)
	}
	if path.Query["param1"] != "value1" {
		t.Errorf("expected Query param1 to be 'value1', got '%s'", path.Query["param1"])
	}
}

func TestVFSFileInfo(t *testing.T) {
	// Test VFSFileInfo struct
	info := &VFSFileInfo{
		URI:     "viking://workspace/path/file.txt",
		Name:    "file.txt",
		Size:    1024,
		Type:    "file",
		ModTime: time.Now(),
		Layers: LayerAvailability{
			L0: true,
			L1: true,
			L2: false,
		},
	}

	if info.URI != "viking://workspace/path/file.txt" {
		t.Errorf("expected URI 'viking://workspace/path/file.txt', got '%s'", info.URI)
	}
	if info.Name != "file.txt" {
		t.Errorf("expected Name 'file.txt', got '%s'", info.Name)
	}
	if info.Size != 1024 {
		t.Errorf("expected Size 1024, got %d", info.Size)
	}
	if info.Type != "file" {
		t.Errorf("expected Type 'file', got '%s'", info.Type)
	}
	if !info.Layers.L0 || !info.Layers.L1 || info.Layers.L2 {
		t.Error("expected Layers.L0 and Layers.L1 to be true, L2 to be false")
	}
}

func TestVFSSearchResult(t *testing.T) {
	// Test VFSSearchResult struct
	result := &VFSSearchResult{
		URI:     "viking://workspace/path/file.txt",
		Score:   0.95,
		Snippet: "matching content snippet",
		Layer:   LayerTypeL1,
	}

	if result.URI != "viking://workspace/path/file.txt" {
		t.Errorf("expected URI 'viking://workspace/path/file.txt', got '%s'", result.URI)
	}
	if result.Score != 0.95 {
		t.Errorf("expected Score 0.95, got %f", result.Score)
	}
	if result.Snippet != "matching content snippet" {
		t.Errorf("expected Snippet 'matching content snippet', got '%s'", result.Snippet)
	}
	if result.Layer != LayerTypeL1 {
		t.Errorf("expected Layer LayerTypeL1, got '%s'", result.Layer)
	}
}

func TestLayerAvailability(t *testing.T) {
	// Test LayerAvailability struct
	la := &LayerAvailability{
		L0: true,
		L1: true,
		L2: false,
	}

	if !la.L0 {
		t.Error("expected L0 to be true")
	}
	if !la.L1 {
		t.Error("expected L1 to be true")
	}
	if la.L2 {
		t.Error("expected L2 to be false")
	}
}

func TestMessage_Struct(t *testing.T) {
	// Test Message struct
	msg := &Message{
		Role:      "user",
		Content:   "test message",
		Timestamp: time.Now().Unix(),
		Metadata: map[string]string{
			"source": "chat",
		},
	}

	if msg.Role != "user" {
		t.Errorf("expected Role 'user', got '%s'", msg.Role)
	}
	if msg.Content != "test message" {
		t.Errorf("expected Content 'test message', got '%s'", msg.Content)
	}
	if msg.Metadata["source"] != "chat" {
		t.Errorf("expected Metadata source to be 'chat', got '%s'", msg.Metadata["source"])
	}
}

// Helper functions for tests
func stringPtr(s string) *string {
	return &s
}

func contextTypePtr(ct ContextType) *ContextType {
	return &ct
}
