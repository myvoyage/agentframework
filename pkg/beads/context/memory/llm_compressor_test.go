// Agent Framework - Context Memory Tests
// Copyright (C) 2025  Agent Framework Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package memory

import (
	stdctx "context"
	"testing"
	"time"

	"AgentFramework/agent"
	beadscontext "AgentFramework/pkg/beads/context"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

// MockChatModel is a mock implementation of agent.ChatModel for testing
type MockChatModel struct {
	GenerateFunc func(ctx stdctx.Context, msgs []*schema.Message, opts ...model.Option) (*schema.Message, error)
}

func (m *MockChatModel) Generate(ctx stdctx.Context, msgs []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	if m.GenerateFunc != nil {
		return m.GenerateFunc(ctx, msgs, opts...)
	}
	return &schema.Message{
		Content: `{"score": 0.8}`,
	}, nil
}

func (m *MockChatModel) Stream(ctx stdctx.Context, msgs []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	return nil, nil
}

var _ agent.ChatModel = (*MockChatModel)(nil)

// TestNewLLMCompressor 测试创建新的 LLM 压缩器
func TestNewLLMCompressor(t *testing.T) {
	mockModel := &MockChatModel{}
	config := &beadscontext.MemoryCompressionConfig{
		MaxSessionMemories:     10,
		MaxDailyMemories:       20,
		MaxLongTermMemories:    50,
		EnableAsyncCompression: false,
		CompressionThreshold:   0.7,
	}

	compressor := NewLLMCompressor(mockModel, config)

	if compressor == nil {
		t.Fatal("expected compressor to be created")
	}
	if compressor.config != config {
		t.Error("expected config to be set")
	}
	if compressor.model == nil {
		t.Error("expected model to be set")
	}
}

// TestNewLLMCompressor_NilConfig 测试使用默认配置创建压缩器
func TestNewLLMCompressor_NilConfig(t *testing.T) {
	mockModel := &MockChatModel{}
	compressor := NewLLMCompressor(mockModel, nil)

	if compressor == nil {
		t.Fatal("expected compressor to be created")
	}
	if compressor.config == nil {
		t.Error("expected default config to be set")
	}
}

// TestCompressMemories_NilMemories 测试压缩空记忆
func TestCompressMemories_NilMemories(t *testing.T) {
	mockModel := &MockChatModel{}
	compressor := NewLLMCompressor(mockModel, nil)

	ctx := stdctx.Background()
	result, err := compressor.CompressMemories(ctx, nil, beadscontext.MemoryTierSession)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result == nil {
		t.Error("expected empty memory collection to be returned")
	}
}

// TestCompressMemories_EmptyMemories 测试压缩空记忆集合
func TestCompressMemories_EmptyMemories(t *testing.T) {
	mockModel := &MockChatModel{}
	compressor := NewLLMCompressor(mockModel, nil)

	ctx := stdctx.Background()
	memories := NewMemoryCollection()
	result, err := compressor.CompressMemories(ctx, memories, beadscontext.MemoryTierSession)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result == nil {
		t.Error("expected result to be returned")
	}
}

// TestCompressMemories_BelowTarget 测试记忆数量低于目标
func TestCompressMemories_BelowTarget(t *testing.T) {
	mockModel := &MockChatModel{}
	config := &beadscontext.MemoryCompressionConfig{
		MaxSessionMemories: 10,
	}
	compressor := NewLLMCompressor(mockModel, config)

	ctx := stdctx.Background()
	memories := NewMemoryCollection()
	memories.Profiles = append(memories.Profiles, &beadscontext.ProfileMemory{ID: "1"})

	result, err := compressor.CompressMemories(ctx, memories, beadscontext.MemoryTierSession)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(result.Profiles) != 1 {
		t.Errorf("expected 1 profile, got %d", len(result.Profiles))
	}
}

// TestCompressMemories_AsyncMode 测试异步压缩模式
func TestCompressMemories_AsyncMode(t *testing.T) {
	mockModel := &MockChatModel{}
	config := &beadscontext.MemoryCompressionConfig{
		MaxSessionMemories:     1,
		EnableAsyncCompression: true,
	}
	compressor := NewLLMCompressor(mockModel, config)

	ctx := stdctx.Background()
	memories := NewMemoryCollection()
	for i := 0; i < 10; i++ {
		memories.Profiles = append(memories.Profiles, &beadscontext.ProfileMemory{
			ID:   string(rune('a' + i)),
			Name: "Test",
		})
	}

	result, err := compressor.CompressMemories(ctx, memories, beadscontext.MemoryTierSession)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	// Async mode returns immediately with original memories
	if result == nil {
		t.Error("expected result to be returned")
	}
}

// TestSummarizeByType_UnknownType 测试未知类型
func TestSummarizeByType_UnknownType(t *testing.T) {
	mockModel := &MockChatModel{}
	compressor := NewLLMCompressor(mockModel, nil)

	ctx := stdctx.Background()
	memories := NewMemoryCollection()

	_, err := compressor.SummarizeByType(ctx, memories, beadscontext.MemoryType("unknown"))

	if err == nil {
		t.Error("expected error for unknown memory type")
	}
}

// TestExtractEssentials_NilMemories 测试从空集合提取
func TestExtractEssentials_NilMemories(t *testing.T) {
	mockModel := &MockChatModel{}
	compressor := NewLLMCompressor(mockModel, nil)

	ctx := stdctx.Background()
	result, err := compressor.ExtractEssentials(ctx, nil, 10)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result == nil {
		t.Error("expected empty collection to be returned")
	}
}

// TestMergeMemories_BothNil 测试合并两个空集合
func TestMergeMemories_BothNil(t *testing.T) {
	mockModel := &MockChatModel{}
	compressor := NewLLMCompressor(mockModel, nil)

	ctx := stdctx.Background()
	result, err := compressor.MergeMemories(ctx, nil, nil)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result != nil {
		t.Error("expected nil result when both inputs are nil")
	}
}

// TestMergeMemories_BaseNil 测试合并空基础集合
func TestMergeMemories_BaseNil(t *testing.T) {
	mockModel := &MockChatModel{}
	compressor := NewLLMCompressor(mockModel, nil)

	ctx := stdctx.Background()
	delta := NewMemoryCollection()
	delta.Profiles = append(delta.Profiles, &beadscontext.ProfileMemory{ID: "1"})

	result, err := compressor.MergeMemories(ctx, nil, delta)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result == nil {
		t.Error("expected result to be returned")
	}
}

// TestMergeMemories_DeltaNil 测试合并空增量集合
func TestMergeMemories_DeltaNil(t *testing.T) {
	mockModel := &MockChatModel{}
	compressor := NewLLMCompressor(mockModel, nil)

	ctx := stdctx.Background()
	base := NewMemoryCollection()
	base.Profiles = append(base.Profiles, &beadscontext.ProfileMemory{ID: "1"})

	result, err := compressor.MergeMemories(ctx, base, nil)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result == nil {
		t.Error("expected result to be returned")
	}
	if len(result.Profiles) != 1 {
		t.Errorf("expected 1 profile, got %d", len(result.Profiles))
	}
}

// TestMergeProfiles 测试合并 Profile
func TestMergeProfiles(t *testing.T) {
	mockModel := &MockChatModel{}
	compressor := NewLLMCompressor(mockModel, nil)

	base := []*beadscontext.ProfileMemory{
		{ID: "1", Name: "Base1"},
		{ID: "2", Name: "Base2"},
	}
	delta := []*beadscontext.ProfileMemory{
		{ID: "2", Name: "Delta2"},
		{ID: "3", Name: "Delta3"},
	}

	result := compressor.mergeProfiles(base, delta)

	// Should have 3 profiles (base 1, base 2, delta 3)
	if len(result) != 3 {
		t.Errorf("expected 3 profiles, got %d", len(result))
	}
}

// TestMergePreferences 测试合并 Preference
func TestMergePreferences(t *testing.T) {
	mockModel := &MockChatModel{}
	compressor := NewLLMCompressor(mockModel, nil)

	base := []*beadscontext.PreferenceMemory{
		{ID: "1", Category: "cat", Key: "key1", Confidence: 0.5},
	}
	delta := []*beadscontext.PreferenceMemory{
		{ID: "2", Category: "cat", Key: "key1", Confidence: 0.8},
		{ID: "3", Category: "cat", Key: "key2", Confidence: 0.6},
	}

	result := compressor.mergePreferences(base, delta)

	// Should have 2 preferences (key1 with higher confidence, key2)
	if len(result) != 2 {
		t.Errorf("expected 2 preferences, got %d", len(result))
	}
}

// TestMergeEntities 测试合并 Entity
func TestMergeEntities(t *testing.T) {
	mockModel := &MockChatModel{}
	compressor := NewLLMCompressor(mockModel, nil)

	now := time.Now()
	base := []*beadscontext.EntityMemory{
		{
			ID:         "1",
			Type:       "person",
			Name:       "John",
			Attributes: map[string]string{"age": "30"},
			Relations:  []beadscontext.EntityRelation{{Type: "knows", EntityID: "2"}},
			LastSeen:   now,
		},
	}
	delta := []*beadscontext.EntityMemory{
		{
			ID:         "2",
			Type:       "person",
			Name:       "John",
			Attributes: map[string]string{"city": "NYC"},
			Relations:  []beadscontext.EntityRelation{{Type: "works_with", EntityID: "3"}},
			LastSeen:   now.Add(time.Hour),
		},
	}

	result := compressor.mergeEntities(base, delta)

	// Should merge entities with same type:name
	if len(result) != 1 {
		t.Errorf("expected 1 merged entity, got %d", len(result))
	}
	if result[0].Attributes["city"] != "NYC" {
		t.Error("expected city attribute to be merged")
	}
}

// TestMergeEvents 测试合并 Event
func TestMergeEvents(t *testing.T) {
	mockModel := &MockChatModel{}
	compressor := NewLLMCompressor(mockModel, nil)

	base := []*beadscontext.EventMemory{
		{ID: "1", Title: "Event1"},
	}
	delta := []*beadscontext.EventMemory{
		{ID: "2", Title: "Event2"},
	}

	result := compressor.mergeEvents(base, delta)

	// Should concatenate events
	if len(result) != 2 {
		t.Errorf("expected 2 events, got %d", len(result))
	}
}

// TestMergeCases 测试合并 Case
func TestMergeCases(t *testing.T) {
	mockModel := &MockChatModel{}
	compressor := NewLLMCompressor(mockModel, nil)

	base := []*beadscontext.CaseMemory{
		{ID: "1", Domain: "debug", Lessons: []string{"lesson1"}},
	}
	delta := []*beadscontext.CaseMemory{
		{ID: "2", Domain: "debug", Lessons: []string{"lesson2"}},
		{ID: "3", Domain: "testing", Lessons: []string{"lesson3"}},
	}

	result := compressor.mergeCases(base, delta)

	// Should merge cases with same domain
	if len(result) != 2 {
		t.Errorf("expected 2 cases, got %d", len(result))
	}
}

// TestMergePatterns 测试合并 Pattern
func TestMergePatterns(t *testing.T) {
	mockModel := &MockChatModel{}
	compressor := NewLLMCompressor(mockModel, nil)

	base := []*beadscontext.PatternMemory{
		{ID: "1", Pattern: "pattern1"},
	}
	delta := []*beadscontext.PatternMemory{
		{ID: "2", Pattern: "pattern2"},
	}

	result := compressor.mergePatterns(base, delta)

	// Should concatenate patterns
	if len(result) != 2 {
		t.Errorf("expected 2 patterns, got %d", len(result))
	}
}

// TestTopProfilesByRecency 测试按时间排序 Profile
func TestTopProfilesByRecency(t *testing.T) {
	mockModel := &MockChatModel{}
	compressor := NewLLMCompressor(mockModel, nil)

	now := time.Now()
	profiles := []*beadscontext.ProfileMemory{
		{ID: "1", UpdatedAt: now.Add(-2 * time.Hour)},
		{ID: "2", UpdatedAt: now},
		{ID: "3", UpdatedAt: now.Add(-1 * time.Hour)},
	}

	result := compressor.topProfilesByRecency(profiles, 2)

	if len(result) != 2 {
		t.Errorf("expected 2 profiles, got %d", len(result))
	}
	if result[0].ID != "2" {
		t.Errorf("expected first profile to be ID 2, got %s", result[0].ID)
	}
	if result[1].ID != "3" {
		t.Errorf("expected second profile to be ID 3, got %s", result[1].ID)
	}
}

// TestTopPreferencesByConfidence 测试按置信度排序 Preference
func TestTopPreferencesByConfidence(t *testing.T) {
	mockModel := &MockChatModel{}
	compressor := NewLLMCompressor(mockModel, nil)

	prefs := []*beadscontext.PreferenceMemory{
		{ID: "1", Confidence: 0.5},
		{ID: "2", Confidence: 0.9},
		{ID: "3", Confidence: 0.7},
	}

	result := compressor.topPreferencesByConfidence(prefs, 2)

	if len(result) != 2 {
		t.Errorf("expected 2 preferences, got %d", len(result))
	}
	if result[0].ID != "2" {
		t.Errorf("expected first pref to be ID 2, got %s", result[0].ID)
	}
}

// TestTopEntitiesByRecency 测试按时间排序 Entity
func TestTopEntitiesByRecency(t *testing.T) {
	mockModel := &MockChatModel{}
	compressor := NewLLMCompressor(mockModel, nil)

	now := time.Now()
	entities := []*beadscontext.EntityMemory{
		{ID: "1", LastSeen: now.Add(-2 * time.Hour)},
		{ID: "2", LastSeen: now},
		{ID: "3", LastSeen: now.Add(-1 * time.Hour)},
	}

	result := compressor.topEntitiesByRecency(entities, 2)

	if len(result) != 2 {
		t.Errorf("expected 2 entities, got %d", len(result))
	}
	if result[0].ID != "2" {
		t.Errorf("expected first entity to be ID 2, got %s", result[0].ID)
	}
}

// TestTopEventsByRecency 测试按时间排序 Event
func TestTopEventsByRecency(t *testing.T) {
	mockModel := &MockChatModel{}
	compressor := NewLLMCompressor(mockModel, nil)

	now := time.Now()
	events := []*beadscontext.EventMemory{
		{ID: "1", OccurredAt: now.Add(-2 * time.Hour)},
		{ID: "2", OccurredAt: now},
		{ID: "3", OccurredAt: now.Add(-1 * time.Hour)},
	}

	result := compressor.topEventsByRecency(events, 2)

	if len(result) != 2 {
		t.Errorf("expected 2 events, got %d", len(result))
	}
	if result[0].ID != "2" {
		t.Errorf("expected first event to be ID 2, got %s", result[0].ID)
	}
}

// TestTopCasesByUsage 测试按使用次数排序 Case
func TestTopCasesByUsage(t *testing.T) {
	mockModel := &MockChatModel{}
	compressor := NewLLMCompressor(mockModel, nil)

	cases := []*beadscontext.CaseMemory{
		{ID: "1", AppliedCount: 5},
		{ID: "2", AppliedCount: 10},
		{ID: "3", AppliedCount: 7},
	}

	result := compressor.topCasesByUsage(cases, 2)

	if len(result) != 2 {
		t.Errorf("expected 2 cases, got %d", len(result))
	}
	if result[0].ID != "2" {
		t.Errorf("expected first case to be ID 2, got %s", result[0].ID)
	}
}

// TestTopPatternsByFrequency 测试按频率排序 Pattern
func TestTopPatternsByFrequency(t *testing.T) {
	mockModel := &MockChatModel{}
	compressor := NewLLMCompressor(mockModel, nil)

	patterns := []*beadscontext.PatternMemory{
		{ID: "1", Frequency: 5, Confidence: 0.5},
		{ID: "2", Frequency: 10, Confidence: 0.8},
		{ID: "3", Frequency: 7, Confidence: 0.9},
	}

	result := compressor.topPatternsByFrequency(patterns, 2)

	if len(result) != 2 {
		t.Errorf("expected 2 patterns, got %d", len(result))
	}
}

// TestSummarizeProfiles 测试摘要 Profiles
func TestSummarizeProfiles(t *testing.T) {
	mockModel := &MockChatModel{}
	compressor := NewLLMCompressor(mockModel, nil)

	ctx := stdctx.Background()
	profiles := []*beadscontext.ProfileMemory{
		{
			ID:     "1",
			Name:   "User1",
			Role:   "Dev",
			Traits: map[string]string{"lang": "Go"},
			Goals:  []string{"goal1"},
		},
		{
			ID:     "2",
			Name:   "User2",
			Role:   "QA",
			Traits: map[string]string{"lang": "Python"},
			Goals:  []string{"goal2"},
		},
	}

	result, err := compressor.summarizeProfiles(ctx, profiles)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected result to be returned")
	}

	merged, ok := result.(*beadscontext.ProfileMemory)
	if !ok {
		t.Fatal("expected ProfileMemory result")
	}
	if len(merged.Goals) != 2 {
		t.Errorf("expected 2 goals, got %d", len(merged.Goals))
	}
}

// TestSummarizeProfiles_Empty 测试摘要空 Profiles
func TestSummarizeProfiles_Empty(t *testing.T) {
	mockModel := &MockChatModel{}
	compressor := NewLLMCompressor(mockModel, nil)

	ctx := stdctx.Background()
	result, err := compressor.summarizeProfiles(ctx, []*beadscontext.ProfileMemory{})

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result != nil {
		t.Error("expected nil result for empty profiles")
	}
}

// TestSummarizePreferences 测试摘要 Preferences
func TestSummarizePreferences(t *testing.T) {
	mockModel := &MockChatModel{}
	compressor := NewLLMCompressor(mockModel, nil)

	ctx := stdctx.Background()
	prefs := []*beadscontext.PreferenceMemory{
		{ID: "1", Category: "ui", Key: "theme", Confidence: 0.5},
		{ID: "2", Category: "ui", Key: "theme", Confidence: 0.8},
		{ID: "3", Category: "coding", Key: "style", Confidence: 0.6},
	}

	result, err := compressor.summarizePreferences(ctx, prefs)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected result to be returned")
	}

	summarized, ok := result.([]*beadscontext.PreferenceMemory)
	if !ok {
		t.Fatal("expected preference slice result")
	}
	// Should have 2 preferences (one per category: ui with highest confidence, coding)
	if len(summarized) != 2 {
		t.Errorf("expected 2 preferences, got %d", len(summarized))
	}
}

// TestSummarizeEntities 测试摘要 Entities
func TestSummarizeEntities(t *testing.T) {
	mockModel := &MockChatModel{}
	compressor := NewLLMCompressor(mockModel, nil)

	ctx := stdctx.Background()
	now := time.Now()
	entities := []*beadscontext.EntityMemory{
		{
			ID:         "1",
			Type:       "person",
			Name:       "John",
			Attributes: map[string]string{"age": "30"},
			LastSeen:   now,
		},
		{
			ID:         "2",
			Type:       "person",
			Name:       "John",
			Attributes: map[string]string{"city": "NYC"},
			LastSeen:   now,
		},
	}

	result, err := compressor.summarizeEntities(ctx, entities)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected result to be returned")
	}

	summarized, ok := result.([]*beadscontext.EntityMemory)
	if !ok {
		t.Fatal("expected entity slice result")
	}
	// Should have 1 entity (merged)
	if len(summarized) != 1 {
		t.Errorf("expected 1 merged entity, got %d", len(summarized))
	}
}

// TestSummarizeCases 测试摘要 Cases
func TestSummarizeCases(t *testing.T) {
	mockModel := &MockChatModel{}
	compressor := NewLLMCompressor(mockModel, nil)

	ctx := stdctx.Background()
	cases := []*beadscontext.CaseMemory{
		{ID: "1", AppliedCount: 5},
		{ID: "2", AppliedCount: 10},
	}

	result, err := compressor.summarizeCases(ctx, cases)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected result to be returned")
	}
}

// TestSummarizePatterns 测试摘要 Patterns
func TestSummarizePatterns(t *testing.T) {
	mockModel := &MockChatModel{}
	compressor := NewLLMCompressor(mockModel, nil)

	ctx := stdctx.Background()
	patterns := []*beadscontext.PatternMemory{
		{ID: "1", Frequency: 5},
		{ID: "2", Frequency: 10},
	}

	result, err := compressor.summarizePatterns(ctx, patterns)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected result to be returned")
	}
}

// TestBuildImportancePrompt 测试构建重要性评分 Prompt
func TestBuildImportancePrompt(t *testing.T) {
	mockModel := &MockChatModel{}
	compressor := NewLLMCompressor(mockModel, nil)

	memory := &beadscontext.ProfileMemory{
		ID:   "1",
		Name: "Test",
	}

	prompt := compressor.buildImportancePrompt(memory)

	if prompt == "" {
		t.Error("expected prompt to be generated")
	}
}

// TestBuildEventSummaryPrompt 测试构建事件摘要 Prompt
func TestBuildEventSummaryPrompt(t *testing.T) {
	mockModel := &MockChatModel{}
	compressor := NewLLMCompressor(mockModel, nil)

	events := []*beadscontext.EventMemory{
		{ID: "1", Title: "Event1", Description: "Desc1"},
		{ID: "2", Title: "Event2", Description: "Desc2"},
	}

	prompt := compressor.buildEventSummaryPrompt(events)

	if prompt == "" {
		t.Error("expected prompt to be generated")
	}
	if !containsString(prompt, "Event1") {
		t.Error("expected prompt to contain Event1")
	}
}

// TestExtractTopProfiles 测试提取顶级 Profiles
func TestExtractTopProfiles(t *testing.T) {
	mockModel := &MockChatModel{}
	compressor := NewLLMCompressor(mockModel, nil)

	ctx := stdctx.Background()
	now := time.Now()
	profiles := []*beadscontext.ProfileMemory{
		{ID: "1", UpdatedAt: now.Add(-2 * time.Hour)},
		{ID: "2", UpdatedAt: now},
		{ID: "3", UpdatedAt: now.Add(-1 * time.Hour)},
	}

	result, err := compressor.extractTopProfiles(ctx, profiles, 2)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 profiles, got %d", len(result))
	}
}

// TestExtractTopPreferences 测试提取顶级 Preferences
func TestExtractTopPreferences(t *testing.T) {
	mockModel := &MockChatModel{}
	compressor := NewLLMCompressor(mockModel, nil)

	ctx := stdctx.Background()
	prefs := []*beadscontext.PreferenceMemory{
		{ID: "1", Confidence: 0.5},
		{ID: "2", Confidence: 0.9},
		{ID: "3", Confidence: 0.7},
	}

	result, err := compressor.extractTopPreferences(ctx, prefs, 2)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 preferences, got %d", len(result))
	}
}

// TestExtractTopEntities 测试提取顶级 Entities
func TestExtractTopEntities(t *testing.T) {
	mockModel := &MockChatModel{}
	compressor := NewLLMCompressor(mockModel, nil)

	ctx := stdctx.Background()
	now := time.Now()
	entities := []*beadscontext.EntityMemory{
		{ID: "1", LastSeen: now.Add(-2 * time.Hour)},
		{ID: "2", LastSeen: now},
		{ID: "3", LastSeen: now.Add(-1 * time.Hour)},
	}

	result, err := compressor.extractTopEntities(ctx, entities, 2)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 entities, got %d", len(result))
	}
}

// TestExtractTopEvents 测试提取顶级 Events
func TestExtractTopEvents(t *testing.T) {
	mockModel := &MockChatModel{}
	compressor := NewLLMCompressor(mockModel, nil)

	ctx := stdctx.Background()
	now := time.Now()
	events := []*beadscontext.EventMemory{
		{ID: "1", OccurredAt: now.Add(-2 * time.Hour)},
		{ID: "2", OccurredAt: now},
		{ID: "3", OccurredAt: now.Add(-1 * time.Hour)},
	}

	result, err := compressor.extractTopEvents(ctx, events, 2)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 events, got %d", len(result))
	}
}

// TestExtractTopCases 测试提取顶级 Cases
func TestExtractTopCases(t *testing.T) {
	mockModel := &MockChatModel{}
	compressor := NewLLMCompressor(mockModel, nil)

	ctx := stdctx.Background()
	cases := []*beadscontext.CaseMemory{
		{ID: "1", AppliedCount: 5},
		{ID: "2", AppliedCount: 10},
		{ID: "3", AppliedCount: 7},
	}

	result, err := compressor.extractTopCases(ctx, cases, 2)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 cases, got %d", len(result))
	}
}

// TestExtractTopPatterns 测试提取顶级 Patterns
func TestExtractTopPatterns(t *testing.T) {
	mockModel := &MockChatModel{}
	compressor := NewLLMCompressor(mockModel, nil)

	ctx := stdctx.Background()
	patterns := []*beadscontext.PatternMemory{
		{ID: "1", Frequency: 5, Confidence: 0.5},
		{ID: "2", Frequency: 10, Confidence: 0.8},
		{ID: "3", Frequency: 7, Confidence: 0.9},
	}

	result, err := compressor.extractTopPatterns(ctx, patterns, 2)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 patterns, got %d", len(result))
	}
}

// Helper function
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestMin 辅助函数测试
func TestMin(t *testing.T) {
	tests := []struct {
		name     string
		a, b     int
		expected int
	}{
		{"a less than b", 5, 10, 5},
		{"a greater than b", 10, 5, 5},
		{"equal values", 5, 5, 5},
		{"negative values", -5, -10, -10},
		{"zero and positive", 0, 5, 0},
		{"negative and zero", -5, 0, -5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := min(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

// TestNewMemoryCollection 测试创建新的记忆集合
func TestNewMemoryCollection(t *testing.T) {
	collection := NewMemoryCollection()

	if collection == nil {
		t.Fatal("expected collection to be created")
	}
	if collection.Profiles == nil {
		t.Error("expected profiles to be initialized")
	}
	if collection.Preferences == nil {
		t.Error("expected preferences to be initialized")
	}
	if collection.Entities == nil {
		t.Error("expected entities to be initialized")
	}
	if collection.Events == nil {
		t.Error("expected events to be initialized")
	}
	if collection.Cases == nil {
		t.Error("expected cases to be initialized")
	}
	if collection.Patterns == nil {
		t.Error("expected patterns to be initialized")
	}
	if len(collection.Profiles) != 0 {
		t.Error("expected empty profiles slice")
	}
}
