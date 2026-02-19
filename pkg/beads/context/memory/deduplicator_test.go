// Agent Framework - Context Memory Tests
// Copyright (C) 2025  Agent Framework Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package memory

import (
	"testing"
	"time"

	beadscontext "AgentFramework/pkg/beads/context"
)

// TestNewDeduplicator 测试创建去重器
func TestNewDeduplicator(t *testing.T) {
	d := NewDeduplicator()
	if d == nil {
		t.Fatal("expected deduplicator to be created")
	}
	if d.threshold != 0.85 {
		t.Errorf("expected threshold 0.85, got %f", d.threshold)
	}
}

// TestDeduplicate_NilMemories 测试去重空记忆
func TestDeduplicator_NilMemories(t *testing.T) {
	d := NewDeduplicator()
	ctx := beadscontext.Context{}

	result, err := d.Deduplicate(ctx, nil)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result == nil {
		t.Error("expected empty collection to be returned")
	}
	// Profiles is nil (zero value) which is fine for empty collection
	if len(result.Profiles) != 0 {
		t.Errorf("expected 0 profiles, got %d", len(result.Profiles))
	}
}

// TestDeduplicate_EmptyMemories 测试去重空记忆集合
func TestDeduplicator_EmptyMemories(t *testing.T) {
	d := NewDeduplicator()
	ctx := beadscontext.Context{}

	memories := &beadscontext.MemoryCollection{}
	result, err := d.Deduplicate(ctx, memories)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result == nil {
		t.Error("expected result to be returned")
	}
}

// TestDeduplicator_DeduplicateProfiles 测试去重 Profile
func TestDeduplicator_DeduplicateProfiles(t *testing.T) {
	d := NewDeduplicator()

	profiles := []*beadscontext.ProfileMemory{
		{ID: "1", Name: "User1"},
		{ID: "2", Name: "User2"},
		{ID: "1", Name: "User1 Updated"}, // Duplicate ID - should be merged
	}

	result := d.deduplicateProfiles(profiles)

	// Should have 2 profiles with different IDs: "1" and "2"
	if len(result) != 2 {
		t.Errorf("expected 2 profiles after deduplication, got %d", len(result))
	}
}

// TestDeduplicator_DeduplicateProfiles_Empty 测试去重空 Profile 列表
func TestDeduplicator_DeduplicateProfiles_Empty(t *testing.T) {
	d := NewDeduplicator()

	result := d.deduplicateProfiles([]*beadscontext.ProfileMemory{})

	if len(result) != 0 {
		t.Errorf("expected 0 profiles, got %d", len(result))
	}
}

// TestMergeProfiles_Deduplicator 测试合并 Profile（去重器）
func TestMergeProfiles_Deduplicator(t *testing.T) {
	d := NewDeduplicator()

	existing := &beadscontext.ProfileMemory{
		ID:     "1",
		Name:   "User",
		Role:   "Dev",
		Traits: map[string]string{"lang": "Go"},
		Goals:  []string{"goal1"},
	}

	new := &beadscontext.ProfileMemory{
		ID:     "1",
		Name:   "User",
		Role:   "Dev",
		Traits: map[string]string{"level": "senior"},
		Goals:  []string{"goal2"},
	}

	merged := d.mergeProfiles(existing, new)

	if merged.ID != "1" {
		t.Errorf("expected ID '1', got '%s'", merged.ID)
	}
	if len(merged.Traits) != 2 {
		t.Errorf("expected 2 traits, got %d", len(merged.Traits))
	}
}

// TestProfileSimilarity 测试 Profile 相似度计算
func TestProfileSimilarity_Deduplicator(t *testing.T) {
	d := NewDeduplicator()

	profile1 := &beadscontext.ProfileMemory{
		ID:     "1",
		Name:   "User",
		Role:   "Dev",
		Traits: map[string]string{"lang": "Go"},
	}
	profile2 := &beadscontext.ProfileMemory{
		ID:     "1",
		Name:   "User",
		Role:   "Dev",
		Traits: map[string]string{"lang": "Go"},
	}

	similarity := d.profileSimilarity(profile1, profile2)
	if similarity < 0.8 {
		t.Errorf("expected high similarity, got %f", similarity)
	}

	// Different ID
	profile3 := &beadscontext.ProfileMemory{ID: "2"}
	similarity = d.profileSimilarity(profile1, profile3)
	if similarity != 0.0 {
		t.Errorf("expected 0 similarity for different IDs, got %f", similarity)
	}
}

// TestDeduplicator_DeduplicatePreferences 测试去重 Preference
func TestDeduplicator_DeduplicatePreferences(t *testing.T) {
	d := NewDeduplicator()

	prefs := []*beadscontext.PreferenceMemory{
		{ID: "1", Category: "ui", Key: "theme", Value: "dark", Confidence: 0.5},
		{ID: "2", Category: "ui", Key: "theme", Value: "light", Confidence: 0.8},
		{ID: "3", Category: "ui", Key: "font", Value: "arial", Confidence: 0.6},
	}

	result := d.deduplicatePreferences(prefs)

	if len(result) != 2 {
		t.Errorf("expected 2 preferences after deduplication, got %d", len(result))
	}
}

// TestDeduplicator_DeduplicateEntities 测试去重 Entity
func TestDeduplicator_DeduplicateEntities(t *testing.T) {
	d := NewDeduplicator()
	now := time.Now()

	entities := []*beadscontext.EntityMemory{
		{ID: "1", Type: "person", Name: "John", Attributes: map[string]string{"age": "30"}, FirstSeen: now, LastSeen: now},
		{ID: "2", Type: "person", Name: "John", Attributes: map[string]string{"city": "NYC"}, FirstSeen: now, LastSeen: now.Add(time.Hour)},
	}

	result := d.deduplicateEntities(entities)

	if len(result) != 1 {
		t.Errorf("expected 1 entity after deduplication, got %d", len(result))
	}
}

// TestMergeEntities_Deduplicator 测试合并 Entity
func TestMergeEntities_Deduplicator(t *testing.T) {
	d := NewDeduplicator()
	now := time.Now()

	existing := &beadscontext.EntityMemory{
		ID:         "1",
		Type:       "person",
		Name:       "John",
		Attributes: map[string]string{"age": "30"},
		Relations:  []beadscontext.EntityRelation{{Type: "knows", EntityID: "2"}},
		FirstSeen:  now,
		LastSeen:   now,
	}

	new := &beadscontext.EntityMemory{
		ID:         "1",
		Type:       "person",
		Name:       "John",
		Attributes: map[string]string{"city": "NYC"},
		Relations:  []beadscontext.EntityRelation{{Type: "works_with", EntityID: "3"}},
		FirstSeen:  now,
		LastSeen:   now.Add(time.Hour),
	}

	merged := d.mergeEntities(existing, new)

	if len(merged.Attributes) != 2 {
		t.Errorf("expected 2 attributes, got %d", len(merged.Attributes))
	}
}

// TestDeduplicator_DeduplicateCases 测试去重 Case
func TestDeduplicator_DeduplicateCases(t *testing.T) {
	d := NewDeduplicator()

	cases := []*beadscontext.CaseMemory{
		{ID: "1", Domain: "debug", Problem: "Memory leak", Solution: "Fix connection"},
		{ID: "2", Domain: "debug", Problem: "Memory leak", Solution: "Another fix"},
		{ID: "3", Domain: "test", Problem: "Test failure", Solution: "Fix test"},
	}

	result := d.deduplicateCases(cases)

	if len(result) != 2 {
		t.Errorf("expected 2 cases after deduplication, got %d", len(result))
	}
}

// TestDeduplicator_DeduplicatePatterns 测试去重 Pattern
func TestDeduplicator_DeduplicatePatterns(t *testing.T) {
	d := NewDeduplicator()
	now := time.Now()

	patterns := []*beadscontext.PatternMemory{
		{ID: "1", Category: "coding", Pattern: "Use functions", Frequency: 5, Confidence: 0.7, LastSeen: now},
		{ID: "2", Category: "coding", Pattern: "Use functions", Frequency: 3, Confidence: 0.9, LastSeen: now.Add(time.Hour)},
	}

	result := d.deduplicatePatterns(patterns)

	if len(result) != 1 {
		t.Errorf("expected 1 pattern after deduplication, got %d", len(result))
	}
}

// TestStringSimilarity 测试字符串相似度计算
func TestStringSimilarity_Deduplicator(t *testing.T) {
	d := NewDeduplicator()

	similarity := d.stringSimilarity("hello world", "hello world")
	if similarity != 1.0 {
		t.Errorf("expected similarity 1.0, got %f", similarity)
	}

	similarity = d.stringSimilarity("", "")
	if similarity != 1.0 {
		t.Errorf("expected similarity 1.0 for empty strings, got %f", similarity)
	}

	similarity = d.stringSimilarity("hello", "")
	if similarity != 0.0 {
		t.Errorf("expected similarity 0.0, got %f", similarity)
	}
}

// TestMapSimilarity 测试 map 相似度计算
func TestMapSimilarity_Deduplicator(t *testing.T) {
	d := NewDeduplicator()

	map1 := map[string]string{"a": "1", "b": "2"}
	map2 := map[string]string{"a": "1", "b": "2"}
	similarity := d.mapSimilarity(map1, map2)
	if similarity != 1.0 {
		t.Errorf("expected similarity 1.0, got %f", similarity)
	}
}

// TestHashContent 测试内容哈希
func TestHashContent_Deduplicator(t *testing.T) {
	d := NewDeduplicator()

	content := "Hello World"
	hash1 := d.hashContent(content)
	hash2 := d.hashContent(content)

	if hash1 != hash2 {
		t.Error("expected same hash for same content")
	}

	emptyHash := d.hashContent("")
	if emptyHash != "" {
		t.Errorf("expected empty hash for empty content, got '%s'", emptyHash)
	}
}

// TestFindSimilar 测试查找相似记忆
func TestFindSimilar_Deduplicator(t *testing.T) {
	d := NewDeduplicator()
	ctx := beadscontext.Context{}

	memory := &beadscontext.ProfileMemory{ID: "1"}
	results, err := d.FindSimilar(ctx, memory)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if results == nil {
		t.Error("expected results slice to be returned")
	}
}

// TestShouldMerge_Deduplicator 测试判断是否应该合并
func TestShouldMerge_Deduplicator(t *testing.T) {
	d := NewDeduplicator()
	ctx := beadscontext.Context{}

	profile1 := &beadscontext.ProfileMemory{ID: "1", Name: "User", Role: "Dev"}
	profile2 := &beadscontext.ProfileMemory{ID: "1", Name: "User", Role: "Dev"}

	shouldMerge, err := d.ShouldMerge(ctx, profile1, profile2)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if !shouldMerge {
		t.Error("expected should merge to be true")
	}
}

// TestNewMemoryDeduplicator 测试创建记忆去重器
func TestNewMemoryDeduplicator(t *testing.T) {
	md := NewMemoryDeduplicator()
	if md == nil {
		t.Fatal("expected MemoryDeduplicator to be created")
	}
}

// TestDeduplicate_Events 测试事件不被去重
func TestDeduplicate_Events(t *testing.T) {
	d := NewDeduplicator()
	ctx := beadscontext.Context{}

	events := []*beadscontext.EventMemory{
		{ID: "1", Title: "Event 1", Description: "Same event"},
		{ID: "2", Title: "Event 1", Description: "Same event"},
		{ID: "3", Title: "Event 2"},
	}

	memories := &beadscontext.MemoryCollection{Events: events}
	result, err := d.Deduplicate(ctx, memories)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(result.Events) != 3 {
		t.Errorf("expected 3 events (no deduplication), got %d", len(result.Events))
	}
}

// TestSplitWords 测试分词
func TestSplitWords_Deduplicator(t *testing.T) {
	words := splitWords("hello world test")

	if len(words) != 3 {
		t.Errorf("expected 3 words, got %d", len(words))
	}
}

// TestJoinStrings 测试连接字符串
func TestJoinStrings_Deduplicator(t *testing.T) {
	strs := []string{"hello", "world", "test"}
	result := joinStrings(strs)

	if result != "hello world test" {
		t.Errorf("expected 'hello world test', got '%s'", result)
	}
}
