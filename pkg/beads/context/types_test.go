// Agent Framework - Context Types Tests
// Copyright (C) 2025  Agent Framework Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package context

import (
	"testing"
	"time"
)

// TestGenerateID 测试上下文 ID 生成
func TestGenerateID(t *testing.T) {
	ctxt := NewContext(ContextTypeFile, "Test File")
	ctxt.Workspace = "/test/workspace"

	id := ctxt.GenerateID()

	if id == "" {
		t.Error("expected non-empty ID")
	}

	if len(id) != 16 {
		t.Errorf("expected ID length 16, got %d", len(id))
	}

	// Same context should generate same ID
	id2 := ctxt.GenerateID()
	if id != id2 {
		t.Error("expected same ID for same context")
	}
}

// TestGenerateID_DifferentContexts 测试不同上下文生成不同 ID
func TestGenerateID_DifferentContexts(t *testing.T) {
	ctxt1 := NewContext(ContextTypeFile, "Test File")
	ctxt2 := NewContext(ContextTypeFile, "Different File")

	id1 := ctxt1.GenerateID()
	id2 := ctxt2.GenerateID()

	if id1 == id2 {
		t.Error("expected different IDs for different contexts")
	}
}

// TestGetContent_AllLayers 测试获取所有层级内容
func TestGetContent_AllLayers(t *testing.T) {
	ctxt := NewContext(ContextTypeFile, "Test")
	ctxt.Layers.L0 = &LayerSummary{Content: "summary"}
	ctxt.Layers.L1 = &LayerOverview{Content: "overview"}
	ctxt.Layers.L2 = &LayerDetails{Content: "details"}

	// Test L0
	content, err := ctxt.GetContent(LayerTypeL0)
	if err != nil {
		t.Errorf("expected no error for L0, got %v", err)
	}
	if content != "summary" {
		t.Errorf("expected L0 content 'summary', got '%s'", content)
	}

	// Test L1
	content, err = ctxt.GetContent(LayerTypeL1)
	if err != nil {
		t.Errorf("expected no error for L1, got %v", err)
	}
	if content != "overview" {
		t.Errorf("expected L1 content 'overview', got '%s'", content)
	}

	// Test L2
	content, err = ctxt.GetContent(LayerTypeL2)
	if err != nil {
		t.Errorf("expected no error for L2, got %v", err)
	}
	if content != "details" {
		t.Errorf("expected L2 content 'details', got '%s'", content)
	}
}

// TestGetContent_AutoSelectsL1 测试自动选择 L1
func TestGetContent_AutoSelectsL1(t *testing.T) {
	ctxt := NewContext(ContextTypeFile, "Test")
	ctxt.Layers.L1 = &LayerOverview{Content: "overview"}

	content, err := ctxt.GetContent(LayerAuto)
	if err != nil {
		t.Errorf("expected no error for Auto with L1, got %v", err)
	}
	if content != "overview" {
		t.Errorf("expected Auto to select L1 content 'overview', got '%s'", content)
	}
}

// TestGetContent_AutoSelectsL0 测试自动选择 L0
func TestGetContent_AutoSelectsL0(t *testing.T) {
	ctxt := NewContext(ContextTypeFile, "Test")
	ctxt.Layers.L0 = &LayerSummary{Content: "summary"}

	content, err := ctxt.GetContent(LayerAuto)
	if err != nil {
		t.Errorf("expected no error for Auto with L0, got %v", err)
	}
	if content != "summary" {
		t.Errorf("expected Auto to select L0 content 'summary', got '%s'", content)
	}
}

// TestGetContent_LayerNotFound 测试获取不存在的层级
func TestGetContent_LayerNotFound(t *testing.T) {
	ctxt := NewContext(ContextTypeFile, "Test")

	_, err := ctxt.GetContent(LayerTypeL0)
	if err == nil {
		t.Error("expected error when layer not found")
	}
}

// TestGetTotalTokens 测试获取总 Token 数
func TestGetTotalTokens(t *testing.T) {
	ctxt := NewContext(ContextTypeFile, "Test")
	ctxt.Layers.L0 = &LayerSummary{Tokens: 100}
	ctxt.Layers.L1 = &LayerOverview{Tokens: 500}
	ctxt.Layers.L2 = &LayerDetails{Tokens: 1000}

	total := ctxt.GetTotalTokens()
	if total != 1600 {
		t.Errorf("expected total 1600, got %d", total)
	}
}

// TestGetTotalTokens_NoLayers 测试无层级时返回 0
func TestGetTotalTokens_NoLayers(t *testing.T) {
	ctxt := NewContext(ContextTypeFile, "Test")

	total := ctxt.GetTotalTokens()
	if total != 0 {
		t.Errorf("expected 0 for no layers, got %d", total)
	}
}

// TestHasLayer 测试检查层级是否存在
func TestHasLayer(t *testing.T) {
	ctxt := NewContext(ContextTypeFile, "Test")
	ctxt.Layers.L0 = &LayerSummary{}
	ctxt.Layers.L2 = &LayerDetails{}

	if !ctxt.HasLayer(LayerTypeL0) {
		t.Error("expected L0 to exist")
	}
	if ctxt.HasLayer(LayerTypeL1) {
		t.Error("expected L1 to not exist")
	}
	if !ctxt.HasLayer(LayerTypeL2) {
		t.Error("expected L2 to exist")
	}
	if ctxt.HasLayer("invalid") {
		t.Error("expected invalid layer to not exist")
	}
}

// TestGetAvailableLayers 测试获取可用层级列表
func TestGetAvailableLayers(t *testing.T) {
	ctxt := NewContext(ContextTypeFile, "Test")
	ctxt.Layers.L0 = &LayerSummary{}
	ctxt.Layers.L2 = &LayerDetails{}

	layers := ctxt.GetAvailableLayers()

	if len(layers) != 2 {
		t.Errorf("expected 2 available layers, got %d", len(layers))
	}

	if layers[0] != LayerTypeL0 {
		t.Errorf("expected first layer L0, got %s", layers[0])
	}

	if layers[1] != LayerTypeL2 {
		t.Errorf("expected second layer L2, got %s", layers[1])
	}
}

// TestEntityRelation 测试实体关系结构
func TestEntityRelation(t *testing.T) {
	now := time.Now()
	relation := EntityRelation{
		Type:       "knows",
		EntityID:    "entity-1",
		Relation:    "knows entity-1",
		Confidence:  0.9,
		FirstSeen:   now,
		LastSeen:    now,
	}

	if relation.Type != "knows" {
		t.Errorf("expected type 'knows', got '%s'", relation.Type)
	}
	if relation.Confidence != 0.9 {
		t.Errorf("expected confidence 0.9, got %f", relation.Confidence)
	}
}

// TestProfileMemory 测试用户画像记忆结构
func TestProfileMemory(t *testing.T) {
	profile := &ProfileMemory{
		ID:     "profile-1",
		Name:   "Test User",
		Role:   "developer",
		Traits: map[string]string{
			"language": "Go",
			"level":    "senior",
		},
		Goals:       []string{"learn", "create"},
		Constraints: []string{"time"},
		UpdatedAt:    time.Now(),
	}

	if profile.GetID() != "profile-1" {
		t.Errorf("expected ID 'profile-1', got '%s'", profile.GetID())
	}

	if profile.Traits["language"] != "Go" {
		t.Errorf("expected trait language 'Go', got '%s'", profile.Traits["language"])
	}

	if len(profile.Goals) != 2 {
		t.Errorf("expected 2 goals, got %d", len(profile.Goals))
	}
}

// TestPreferenceMemory 测试用户偏好记忆结构
func TestPreferenceMemory(t *testing.T) {
	pref := &PreferenceMemory{
		ID:        "pref-1",
		Category:  "coding",
		Key:       "style",
		Value:     "functional",
		Confidence: 0.85,
		UpdatedAt: time.Now(),
	}

	if pref.GetID() != "pref-1" {
		t.Errorf("expected ID 'pref-1', got '%s'", pref.GetID())
	}

	if pref.Category != "coding" {
		t.Errorf("expected category 'coding', got '%s'", pref.Category)
	}

	if pref.Value != "functional" {
		t.Errorf("expected value 'functional', got '%s'", pref.Value)
	}
}

// TestEntityMemory 测试实体知识记忆结构
func TestEntityMemory(t *testing.T) {
	now := time.Now()
	entity := &EntityMemory{
		ID:   "entity-1",
		Type: "person",
		Name: "John Doe",
		Attributes: map[string]string{
			"age":  "30",
			"city": "NYC",
		},
		Relations: []EntityRelation{
			{
				Type:     "knows",
				EntityID: "entity-2",
			},
		},
		FirstSeen: now,
		LastSeen:  now,
	}

	if entity.GetID() != "entity-1" {
		t.Errorf("expected ID 'entity-1', got '%s'", entity.GetID())
	}

	if entity.Type != "person" {
		t.Errorf("expected type 'person', got '%s'", entity.Type)
	}

	if len(entity.Relations) != 1 {
		t.Errorf("expected 1 relation, got %d", len(entity.Relations))
	}
}

// TestEventMemory 测试事件记录记忆结构
func TestEventMemory(t *testing.T) {
	event := &EventMemory{
		ID:          "event-1",
		Type:        "decision",
		Title:       "Made a decision",
		Description: "Decided to use Go",
		OccurredAt:  time.Now(),
		Participants: []string{"user", "assistant"},
		Outcomes:    []string{"success"},
	}

	if event.GetID() != "event-1" {
		t.Errorf("expected ID 'event-1', got '%s'", event.GetID())
	}

	if event.Type != "decision" {
		t.Errorf("expected type 'decision', got '%s'", event.Type)
	}

	if len(event.Participants) != 2 {
		t.Errorf("expected 2 participants, got %d", len(event.Participants))
	}
}

// TestCaseMemory 测试案例库记忆结构
func TestCaseMemory(t *testing.T) {
	caseMem := &CaseMemory{
		ID:           "case-1",
		Domain:       "debugging",
		Problem:      "Memory leak",
		Solution:     "Fixed by closing connections",
		Lessons:      []string{"always close resources"},
		Tags:         []string{"memory", "leak"},
		CreatedAt:    time.Now(),
		AppliedCount: 5,
	}

	if caseMem.GetID() != "case-1" {
		t.Errorf("expected ID 'case-1', got '%s'", caseMem.GetID())
	}

	if caseMem.Domain != "debugging" {
		t.Errorf("expected domain 'debugging', got '%s'", caseMem.Domain)
	}

	if caseMem.AppliedCount != 5 {
		t.Errorf("expected applied count 5, got %d", caseMem.AppliedCount)
	}
}

// TestPatternMemory 测试模式识别记忆结构
func TestPatternMemory(t *testing.T) {
	pattern := &PatternMemory{
		ID:         "pattern-1",
		Category:   "coding",
		Pattern:    "Always handle errors",
		Frequency:  10,
		LastSeen:   time.Now(),
		Confidence: 0.95,
	}

	if pattern.GetID() != "pattern-1" {
		t.Errorf("expected ID 'pattern-1', got '%s'", pattern.GetID())
	}

	if pattern.Category != "coding" {
		t.Errorf("expected category 'coding', got '%s'", pattern.Category)
	}

	if pattern.Frequency != 10 {
		t.Errorf("expected frequency 10, got %d", pattern.Frequency)
	}
}

// TestTieredMemory 测试分层记忆结构
func TestTieredMemory(t *testing.T) {
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)

	tiered := &TieredMemory{
		Tier:            MemoryTierSession,
		CreatedAt:       now,
		ExpiresAt:       expiresAt,
		AccessCount:     5,
		LastAccessed:    now,
		ImportanceScore: 0.8,
	}

	if tiered.Tier != MemoryTierSession {
		t.Errorf("expected tier Session, got %s", tiered.Tier)
	}

	if tiered.AccessCount != 5 {
		t.Errorf("expected access count 5, got %d", tiered.AccessCount)
	}

	if tiered.ImportanceScore != 0.8 {
		t.Errorf("expected importance score 0.8, got %f", tiered.ImportanceScore)
	}
}

// TestMemoryWithMeta 测试带元数据的记忆包装
func TestMemoryWithMeta(t *testing.T) {
	meta := TieredMemory{
		Tier: MemoryTierLongTerm,
	}

	withMeta := &MemoryWithMeta{
		Meta: meta,
		Data: &ProfileMemory{ID: "profile-1"},
	}

	if withMeta.Meta.Tier != MemoryTierLongTerm {
		t.Errorf("expected tier LongTerm, got %s", withMeta.Meta.Tier)
	}
}

// TestNewContext_DefaultValues 测试新上下文默认值
func TestNewContext_DefaultValues(t *testing.T) {
	ctxt := NewContext(ContextTypeFile, "Test File")

	if ctxt.Type != ContextTypeFile {
		t.Errorf("expected type File, got %s", ctxt.Type)
	}

	if ctxt.Title != "Test File" {
		t.Errorf("expected title 'Test File', got '%s'", ctxt.Title)
	}

	if ctxt.Version != 1 {
		t.Errorf("expected version 1, got %d", ctxt.Version)
	}

	if ctxt.Metadata == nil {
		t.Error("expected metadata to be initialized")
	}

	if ctxt.TaskRefs == nil {
		t.Error("expected task refs to be initialized")
	}

	if len(ctxt.TaskRefs) != 0 {
		t.Errorf("expected empty task refs, got %d", len(ctxt.TaskRefs))
	}

	if ctxt.CreatedAt.IsZero() {
		t.Error("expected created at to be set")
	}

	if ctxt.UpdatedAt.IsZero() {
		t.Error("expected updated at to be set")
	}

	if ctxt.AccessedAt.IsZero() {
		t.Error("expected accessed at to be set")
	}
}

// TestContextLayers 测试上下文层级结构
func TestContextLayers(t *testing.T) {
	layers := ContextLayers{
		L0: &LayerSummary{
			Content:     "summary",
			Tokens:      100,
			GeneratedAt: time.Now(),
			Method:      "ai",
		},
		L1: &LayerOverview{
			Content:     "overview",
			Tokens:      500,
			Sections:    []string{"section1"},
			KeyPoints:   []string{"point1"},
			GeneratedAt: time.Now(),
			Method:      "ai",
		},
		L2: &LayerDetails{
			Content:     "details",
			Tokens:      1000,
			Format:      "markdown",
			Encoding:    "utf-8",
			Source:      "file",
			GeneratedAt: time.Now(),
			Metadata: map[string]string{
				"key": "value",
			},
		},
	}

	if layers.L0 == nil {
		t.Error("expected L0 to be set")
	}

	if layers.L1 == nil {
		t.Error("expected L1 to be set")
	}

	if layers.L2 == nil {
		t.Error("expected L2 to be set")
	}

	if layers.L2.Format != "markdown" {
		t.Errorf("expected format 'markdown', got '%s'", layers.L2.Format)
	}

	if layers.L2.Metadata["key"] != "value" {
		t.Errorf("expected metadata key 'value', got '%s'", layers.L2.Metadata["key"])
	}
}

// TestLayerDetails_DefaultEncoding 测试 L2 默认编码
func TestLayerDetails_DefaultEncoding(t *testing.T) {
	details := &LayerDetails{
		Content:     "test",
		Tokens:      10,
		Format:      "plain",
		GeneratedAt: time.Now(),
	}

	// Encoding should be empty string by default
	if details.Encoding != "" {
		t.Errorf("expected empty encoding, got '%s'", details.Encoding)
	}
}

// TestLayerDetails_EmptyMetadata 测试空元数据
func TestLayerDetails_EmptyMetadata(t *testing.T) {
	details := &LayerDetails{
		Content:     "test",
		Tokens:      10,
		GeneratedAt: time.Now(),
		Metadata:    map[string]string{},
	}

	if len(details.Metadata) != 0 {
		t.Errorf("expected empty metadata, got %d items", len(details.Metadata))
	}
}

// TestMemoryCollectionUpdateStruct 测试记忆集合更新结构
func TestMemoryCollectionUpdateStruct(t *testing.T) {
	profiles := []*ProfileMemory{{ID: "p1"}, {ID: "p2"}}
	update := &MemoryCollectionUpdate{
		Profiles: &profiles,
	}

	if update.Profiles == nil {
		t.Error("expected profiles to be set")
	}

	if len(*update.Profiles) != 2 {
		t.Errorf("expected 2 profiles, got %d", len(*update.Profiles))
	}
}

// TestContextUpdateStruct 测试上下文更新结构
func TestContextUpdateStruct(t *testing.T) {
	title := "Updated Title"
	metadata := map[string]string{"key": "value"}

	update := &ContextUpdate{
		Title:    &title,
		Metadata: &metadata,
	}

	if update.Title == nil {
		t.Error("expected title to be set")
	}

	if update.Metadata == nil {
		t.Error("expected metadata to be set")
	}
}

// TestGetContent_InvalidLayer 测试获取无效层级
func TestGetContent_InvalidLayer(t *testing.T) {
	ctxt := NewContext(ContextTypeFile, "Test")
	ctxt.Layers.L0 = &LayerSummary{Content: "summary"}

	_, err := ctxt.GetContent("invalid-layer")
	if err == nil {
		t.Error("expected error for invalid layer")
	}
}

// TestGetMemoryCount_Nil 测试 nil MemoryCollection
func TestGetMemoryCount_Nil(t *testing.T) {
	var mc *MemoryCollection
	count := mc.GetMemoryCount()
	if count != 0 {
		t.Errorf("expected 0 for nil collection, got %d", count)
	}
}

// TestMemoryCollection_IsEmpty_Nil 测试 nil MemoryCollection IsEmpty
func TestMemoryCollection_IsEmpty_Nil(t *testing.T) {
	var mc *MemoryCollection
	if !mc.IsEmpty() {
		t.Error("expected true for nil collection")
	}
}

// TestMemoryCollection_IsEmpty_NonEmpty 测试非空 MemoryCollection IsEmpty
func TestMemoryCollection_IsEmpty_NonEmpty(t *testing.T) {
	mc := &MemoryCollection{
		Profiles: []*ProfileMemory{{ID: "p1"}},
	}
	if mc.IsEmpty() {
		t.Error("expected false for non-empty collection")
	}
}

// TestGetTier_NilMetadata 测试没有元数据时的 GetTier
func TestGetTier_NilMetadata(t *testing.T) {
	emc := &EnhancedMemoryCollection{
		MemoryCollection: &MemoryCollection{},
	}

	tier := emc.GetTier("nonexistent-id")
	if tier != MemoryTierSession {
		t.Errorf("expected default Session tier, got %s", tier)
	}
}

// TestGetTier_UnknownID 测试未知 ID 的 GetTier
func TestGetTier_UnknownID(t *testing.T) {
	emc := &EnhancedMemoryCollection{
		MemoryCollection: &MemoryCollection{},
		MemoryMetadata: map[string]TieredMemory{
			"known-id": {Tier: MemoryTierLongTerm},
		},
	}

	tier := emc.GetTier("unknown-id")
	if tier != MemoryTierSession {
		t.Errorf("expected default Session tier for unknown ID, got %s", tier)
	}
}

// TestEnhancedMemoryCollection_IsEmpty_Nil 测试 nil EnhancedMemoryCollection
func TestEnhancedMemoryCollection_IsEmpty_Nil(t *testing.T) {
	var emc *EnhancedMemoryCollection
	if !emc.IsEmpty() {
		t.Error("expected true for nil enhanced collection")
	}
}

// TestEnhancedMemoryCollection_IsEmpty_NilMemoryCollection 测试 MemoryCollection 为 nil
func TestEnhancedMemoryCollection_IsEmpty_NilMemoryCollection(t *testing.T) {
	emc := &EnhancedMemoryCollection{
		MemoryCollection: nil,
	}
	if !emc.IsEmpty() {
		t.Error("expected true when MemoryCollection is nil")
	}
}

// TestEnhancedMemoryCollection_GetMemoryCount_Nil 测试 nil EnhancedMemoryCollection
func TestEnhancedMemoryCollection_GetMemoryCount_Nil(t *testing.T) {
	var emc *EnhancedMemoryCollection
	count := emc.GetMemoryCount()
	if count != 0 {
		t.Errorf("expected 0 for nil enhanced collection, got %d", count)
	}
}

// TestEnhancedMemoryCollection_GetMemoryCount_NilMemoryCollection 测试 MemoryCollection 为 nil
func TestEnhancedMemoryCollection_GetMemoryCount_NilMemoryCollection(t *testing.T) {
	emc := &EnhancedMemoryCollection{
		MemoryCollection: nil,
	}
	count := emc.GetMemoryCount()
	if count != 0 {
		t.Errorf("expected 0 when MemoryCollection is nil, got %d", count)
	}
}

// TestMemoryCollection_AddMemory_AllTypes 测试添加各种类型的记忆
func TestMemoryCollection_AddMemory_AllTypes(t *testing.T) {
	mc := &MemoryCollection{}

	// 添加 ProfileMemory
	profile := &ProfileMemory{ID: "p1", Name: "User1"}
	err := mc.AddMemory(MemoryTypeProfile, profile)
	if err != nil {
		t.Errorf("expected no error adding profile, got %v", err)
	}
	if len(mc.Profiles) != 1 {
		t.Errorf("expected 1 profile, got %d", len(mc.Profiles))
	}

	// 添加 PreferenceMemory
	pref := &PreferenceMemory{ID: "pref1", Category: "coding", Key: "style", Value: "functional"}
	err = mc.AddMemory(MemoryTypePreference, pref)
	if err != nil {
		t.Errorf("expected no error adding preference, got %v", err)
	}
	if len(mc.Preferences) != 1 {
		t.Errorf("expected 1 preference, got %d", len(mc.Preferences))
	}

	// 添加 EntityMemory
	entity := &EntityMemory{ID: "e1", Type: "person", Name: "John"}
	err = mc.AddMemory(MemoryTypeEntity, entity)
	if err != nil {
		t.Errorf("expected no error adding entity, got %v", err)
	}
	if len(mc.Entities) != 1 {
		t.Errorf("expected 1 entity, got %d", len(mc.Entities))
	}

	// 添加 EventMemory
	event := &EventMemory{ID: "ev1", Type: "decision", Title: "Decision"}
	err = mc.AddMemory(MemoryTypeEvent, event)
	if err != nil {
		t.Errorf("expected no error adding event, got %v", err)
	}
	if len(mc.Events) != 1 {
		t.Errorf("expected 1 event, got %d", len(mc.Events))
	}

	// 添加 CaseMemory
	caseMem := &CaseMemory{ID: "c1", Domain: "debugging", Problem: "Bug"}
	err = mc.AddMemory(MemoryTypeCase, caseMem)
	if err != nil {
		t.Errorf("expected no error adding case, got %v", err)
	}
	if len(mc.Cases) != 1 {
		t.Errorf("expected 1 case, got %d", len(mc.Cases))
	}

	// 添加 PatternMemory
	pattern := &PatternMemory{ID: "pat1", Category: "coding", Pattern: "Handle errors"}
	err = mc.AddMemory(MemoryTypePattern, pattern)
	if err != nil {
		t.Errorf("expected no error adding pattern, got %v", err)
	}
	if len(mc.Patterns) != 1 {
		t.Errorf("expected 1 pattern, got %d", len(mc.Patterns))
	}
}

// TestMemoryCollection_AddMemory_NilCollection 测试向 nil 集合添加记忆
func TestMemoryCollection_AddMemory_NilCollection(t *testing.T) {
	var mc *MemoryCollection
	profile := &ProfileMemory{ID: "p1", Name: "User1"}
	err := mc.AddMemory(MemoryTypeProfile, profile)
	if err == nil {
		t.Error("expected error for nil collection")
	}
	if err.Error() != "memory collection is nil" {
		t.Errorf("expected 'memory collection is nil' error, got %v", err)
	}
}

// TestMemoryCollection_AddMemory_UnknownType 测试添加未知类型的记忆
func TestMemoryCollection_AddMemory_UnknownType(t *testing.T) {
	mc := &MemoryCollection{}
	err := mc.AddMemory(MemoryType("unknown"), "invalid memory")
	if err == nil {
		t.Error("expected error for unknown memory type")
	}
}

// TestMemoryCollection_Clear_All 测试清空记忆集合
func TestMemoryCollection_Clear_All(t *testing.T) {
	mc := &MemoryCollection{
		Profiles:    []*ProfileMemory{{ID: "p1"}, {ID: "p2"}},
		Preferences: []*PreferenceMemory{{ID: "pref1"}},
		Entities:    []*EntityMemory{{ID: "e1"}},
		Events:      []*EventMemory{{ID: "ev1"}},
		Cases:       []*CaseMemory{{ID: "c1"}},
		Patterns:    []*PatternMemory{{ID: "pat1"}},
	}

	mc.Clear()

	if mc.Profiles != nil {
		t.Error("expected Profiles to be nil after Clear")
	}
	if mc.Preferences != nil {
		t.Error("expected Preferences to be nil after Clear")
	}
	if mc.Entities != nil {
		t.Error("expected Entities to be nil after Clear")
	}
	if mc.Events != nil {
		t.Error("expected Events to be nil after Clear")
	}
	if mc.Cases != nil {
		t.Error("expected Cases to be nil after Clear")
	}
	if mc.Patterns != nil {
		t.Error("expected Patterns to be nil after Clear")
	}

	if !mc.IsEmpty() {
		t.Error("expected collection to be empty after Clear")
	}
}

// TestMemoryCollection_Clear_Nil 测试清空 nil 集合
func TestMemoryCollection_Clear_Nil(t *testing.T) {
	var mc *MemoryCollection
	// 应该不会 panic
	mc.Clear()
}

// TestGetMemoriesByType_UnknownType 测试获取未知类型的记忆
func TestGetMemoriesByType_UnknownType(t *testing.T) {
	mc := &MemoryCollection{
		Profiles: []*ProfileMemory{{ID: "p1"}},
	}

	result := mc.GetMemoriesByType(MemoryType("unknown"))
	if result != nil {
		t.Error("expected nil for unknown memory type")
	}
}

// TestGetMemoriesByType_NilCollection 测试 nil 集合获取记忆
func TestGetMemoriesByType_NilCollection(t *testing.T) {
	var mc *MemoryCollection
	result := mc.GetMemoriesByType(MemoryTypeProfile)
	if result != nil {
		t.Error("expected nil for nil collection")
	}
}

// TestContextType_Values 测试所有上下文类型常量
func TestContextType_Values(t *testing.T) {
	types := []ContextType{
		ContextTypeProject,
		ContextTypeFile,
		ContextTypeCodebase,
		ContextTypeMemory,
		ContextTypeResource,
		ContextTypeSkill,
		ContextTypeConversation,
		ContextTypeSession,
		ContextTypeKnowledge,
		ContextTypeTask,
		ContextTypeCustom,
	}

	for _, ct := range types {
		if ct == "" {
			t.Errorf("context type should not be empty: %v", ct)
		}
	}
}

// TestLayerType_Values 测试所有层级类型常量
func TestLayerType_Values(t *testing.T) {
	types := []LayerType{
		LayerAuto,
		LayerTypeL0,
		LayerTypeL1,
		LayerTypeL2,
	}

	for _, lt := range types {
		if lt == "" {
			t.Errorf("layer type should not be empty: %v", lt)
		}
	}
}

// TestMemoryType_Values 测试所有记忆类型常量
func TestMemoryType_Values(t *testing.T) {
	types := []MemoryType{
		MemoryTypeProfile,
		MemoryTypePreference,
		MemoryTypeEntity,
		MemoryTypeEvent,
		MemoryTypeCase,
		MemoryTypePattern,
	}

	for _, mt := range types {
		if mt == "" {
			t.Errorf("memory type should not be empty: %v", mt)
		}
	}
}

// TestGetContent_EdgeCases 测试 GetContent 的边缘情况
func TestGetContent_EdgeCases(t *testing.T) {
	ctxt := NewContext(ContextTypeFile, "Test")

	// 测试 L0 为 nil 时获取 L0
	_, err := ctxt.GetContent(LayerTypeL0)
	if err == nil {
		t.Error("expected error when L0 is nil")
	}

	// 测试 L1 为 nil 时获取 L1
	_, err = ctxt.GetContent(LayerTypeL1)
	if err == nil {
		t.Error("expected error when L1 is nil")
	}

	// 测试 L2 为 nil 时获取 L2
	_, err = ctxt.GetContent(LayerTypeL2)
	if err == nil {
		t.Error("expected error when L2 is nil")
	}

	// 测试所有层级为空时使用 Auto
	_, err = ctxt.GetContent(LayerAuto)
	if err == nil {
		t.Error("expected error when all layers are nil")
	}
}

// TestGetTotalTokens_EmptyLayers 测试没有层级时的总 token 数
func TestGetTotalTokens_EmptyLayers(t *testing.T) {
	ctxt := NewContext(ContextTypeFile, "Test")
	// L0, L1, L2 都是 nil

	total := ctxt.GetTotalTokens()
	if total != 0 {
		t.Errorf("expected 0 for no layers, got %d", total)
	}
}

// TestGetAvailableLayers_Empty 测试没有层级时获取可用层级列表
func TestGetAvailableLayers_Empty(t *testing.T) {
	ctxt := NewContext(ContextTypeFile, "Test")
	// 所有层级都是 nil

	layers := ctxt.GetAvailableLayers()
	if len(layers) != 0 {
		t.Errorf("expected 0 available layers, got %d", len(layers))
	}
}

// TestHasLayer_InvalidLayer 测试检查无效的层级
func TestHasLayer_InvalidLayer(t *testing.T) {
	ctxt := NewContext(ContextTypeFile, "Test")

	// 测试无效的层级类型
	if ctxt.HasLayer("invalid") {
		t.Error("expected invalid layer to return false")
	}

	// 测试空字符串层级
	if ctxt.HasLayer("") {
		t.Error("expected empty string layer to return false")
	}
}

// TestMemoryTier_Values 测试所有记忆分层常量
func TestMemoryTier_Values(t *testing.T) {
	tiers := []MemoryTier{
		MemoryTierSession,
		MemoryTierDaily,
		MemoryTierLongTerm,
	}

	for _, tier := range tiers {
		if tier == "" {
			t.Errorf("memory tier should not be empty: %v", tier)
		}
	}
}

// TestSetTier_SetNewTier 测试设置新的记忆分层
func TestSetTier_SetNewTier(t *testing.T) {
	emc := &EnhancedMemoryCollection{
		MemoryCollection: &MemoryCollection{},
		MemoryMetadata:   make(map[string]TieredMemory),
	}

	expiresAt := time.Now().Add(24 * time.Hour)
	importance := 0.8

	emc.SetTier("mem-1", MemoryTierLongTerm, expiresAt, importance)

	meta, ok := emc.MemoryMetadata["mem-1"]
	if !ok {
		t.Error("expected metadata to be set")
	}

	if meta.Tier != MemoryTierLongTerm {
		t.Errorf("expected tier LongTerm, got %s", meta.Tier)
	}

	if meta.ImportanceScore != importance {
		t.Errorf("expected importance %f, got %f", importance, meta.ImportanceScore)
	}
}

// TestGetSessionMemories 测试获取会话记忆
func TestGetSessionMemories(t *testing.T) {
	emc := &EnhancedMemoryCollection{
		MemoryCollection: &MemoryCollection{
			Profiles: []*ProfileMemory{{ID: "p1"}},
		},
	}

	session := emc.GetSessionMemories()
	if session == nil {
		t.Error("expected session memories to be returned")
	}

	if len(session.Profiles) != 1 {
		t.Errorf("expected 1 profile, got %d", len(session.Profiles))
	}
}

// TestGetDailyMemories 测试获取每日记忆
func TestGetDailyMemories(t *testing.T) {
	emc := &EnhancedMemoryCollection{
		MemoryCollection: &MemoryCollection{
			Events: []*EventMemory{{ID: "e1"}},
		},
	}

	daily := emc.GetDailyMemories()
	if daily == nil {
		t.Error("expected daily memories to be returned")
	}

	if len(daily.Events) != 1 {
		t.Errorf("expected 1 event, got %d", len(daily.Events))
	}
}

// TestGetLongTermMemories 测试获取长期记忆
func TestGetLongTermMemories(t *testing.T) {
	emc := &EnhancedMemoryCollection{
		MemoryCollection: &MemoryCollection{
			Cases: []*CaseMemory{{ID: "c1"}},
		},
	}

	longTerm := emc.GetLongTermMemories()
	if longTerm == nil {
		t.Error("expected long term memories to be returned")
	}

	if len(longTerm.Cases) != 1 {
		t.Errorf("expected 1 case, got %d", len(longTerm.Cases))
	}
}

// TestSetTier_InitMetadata 测试设置分层时初始化元数据
func TestSetTier_InitMetadata(t *testing.T) {
	emc := &EnhancedMemoryCollection{
		MemoryCollection: &MemoryCollection{},
		// MemoryMetadata 是 nil
	}

	expiresAt := time.Now().Add(24 * time.Hour)
	emc.SetTier("mem-1", MemoryTierLongTerm, expiresAt, 0.8)

	if emc.MemoryMetadata == nil {
		t.Error("expected MemoryMetadata to be initialized")
	}

	meta, ok := emc.MemoryMetadata["mem-1"]
	if !ok {
		t.Error("expected metadata to be set for mem-1")
	}

	if meta.Tier != MemoryTierLongTerm {
		t.Errorf("expected tier LongTerm, got %s", meta.Tier)
	}
}

// TestGetContent_L0Content 测试获取 L0 内容
func TestGetContent_L0Content(t *testing.T) {
	ctxt := NewContext(ContextTypeFile, "test.txt")
	ctxt.Layers.L0 = &LayerSummary{Content: "summary content"}

	content, err := ctxt.GetContent(LayerTypeL0)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if content != "summary content" {
		t.Errorf("expected 'summary content', got '%s'", content)
	}
}

// TestGetContent_L1Content 测试获取 L1 内容
func TestGetContent_L1Content(t *testing.T) {
	ctxt := NewContext(ContextTypeFile, "test.txt")
	ctxt.Layers.L1 = &LayerOverview{Content: "overview content"}

	content, err := ctxt.GetContent(LayerTypeL1)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if content != "overview content" {
		t.Errorf("expected 'overview content', got '%s'", content)
	}
}

// TestGetContent_L2Content 测试获取 L2 内容
func TestGetContent_L2Content(t *testing.T) {
	ctxt := NewContext(ContextTypeFile, "test.txt")
	ctxt.Layers.L2 = &LayerDetails{Content: "details content"}

	content, err := ctxt.GetContent(LayerTypeL2)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if content != "details content" {
		t.Errorf("expected 'details content', got '%s'", content)
	}
}

// TestGetContent_AutoWithLayers 测试 Auto 模式获取内容
func TestGetContent_AutoWithLayers(t *testing.T) {
	ctxt := NewContext(ContextTypeFile, "test.txt")
	ctxt.Layers.L1 = &LayerOverview{Content: "overview"}
	ctxt.Layers.L2 = &LayerDetails{Content: "details"}

	content, err := ctxt.GetContent(LayerAuto)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Auto 应该优先选择 L1
	if content != "overview" {
		t.Errorf("expected L1 content 'overview', got '%s'", content)
	}
}

// TestGetTotalTokens_WithLayers 测试获取有层级的总 token 数
func TestGetTotalTokens_WithLayers(t *testing.T) {
	ctxt := NewContext(ContextTypeFile, "test.txt")
	ctxt.Layers.L0 = &LayerSummary{Tokens: 100}
	ctxt.Layers.L1 = &LayerOverview{Tokens: 500}

	total := ctxt.GetTotalTokens()
	if total != 600 {
		t.Errorf("expected 600 total tokens, got %d", total)
	}
}

// TestGetAvailableLayers_WithLayers 测试获取有层级时的可用层级列表
func TestGetAvailableLayers_WithLayers(t *testing.T) {
	ctxt := NewContext(ContextTypeFile, "test.txt")
	ctxt.Layers.L0 = &LayerSummary{Content: "summary"}
	ctxt.Layers.L1 = &LayerOverview{Content: "overview"}
	ctxt.Layers.L2 = &LayerDetails{Content: "details"}

	layers := ctxt.GetAvailableLayers()
	if len(layers) != 3 {
		t.Errorf("expected 3 available layers, got %d", len(layers))
	}
}

// TestProfileMemory_GetID 测试 ProfileMemory 的 GetID 方法
func TestProfileMemory_GetID(t *testing.T) {
	p := &ProfileMemory{ID: "profile-1"}
	if p.GetID() != "profile-1" {
		t.Errorf("expected ID 'profile-1', got '%s'", p.GetID())
	}
}

// TestPreferenceMemory_GetID 测试 PreferenceMemory 的 GetID 方法
func TestPreferenceMemory_GetID(t *testing.T) {
	p := &PreferenceMemory{ID: "pref-1"}
	if p.GetID() != "pref-1" {
		t.Errorf("expected ID 'pref-1', got '%s'", p.GetID())
	}
}

// TestEntityMemory_GetID 测试 EntityMemory 的 GetID 方法
func TestEntityMemory_GetID(t *testing.T) {
	e := &EntityMemory{ID: "entity-1"}
	if e.GetID() != "entity-1" {
		t.Errorf("expected ID 'entity-1', got '%s'", e.GetID())
	}
}

// TestEventMemory_GetID 测试 EventMemory 的 GetID 方法
func TestEventMemory_GetID(t *testing.T) {
	e := &EventMemory{ID: "event-1"}
	if e.GetID() != "event-1" {
		t.Errorf("expected ID 'event-1', got '%s'", e.GetID())
	}
}

// TestCaseMemory_GetID 测试 CaseMemory 的 GetID 方法
func TestCaseMemory_GetID(t *testing.T) {
	c := &CaseMemory{ID: "case-1"}
	if c.GetID() != "case-1" {
		t.Errorf("expected ID 'case-1', got '%s'", c.GetID())
	}
}

// TestPatternMemory_GetID 测试 PatternMemory 的 GetID 方法
func TestPatternMemory_GetID(t *testing.T) {
	p := &PatternMemory{ID: "pattern-1"}
	if p.GetID() != "pattern-1" {
		t.Errorf("expected ID 'pattern-1', got '%s'", p.GetID())
	}
}

// TestGetContent_AllLayersPresent 测试 Auto 模式所有层级都存在时
func TestGetContent_AllLayersPresent(t *testing.T) {
	ctxt := NewContext(ContextTypeFile, "test.txt")
	// 只有 L2 存在
	ctxt.Layers.L2 = &LayerDetails{Content: "details content"}

	content, err := ctxt.GetContent(LayerAuto)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if content != "details content" {
		t.Errorf("expected 'details content', got '%s'", content)
	}
}

// TestGetContent_AutoWithOnlyL0 测试 Auto 模式只有 L0
func TestGetContent_AutoWithOnlyL0(t *testing.T) {
	ctxt := NewContext(ContextTypeFile, "test.txt")
	ctxt.Layers.L0 = &LayerSummary{Content: "summary"}

	content, err := ctxt.GetContent(LayerAuto)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if content != "summary" {
		t.Errorf("expected 'summary', got '%s'", content)
	}
}

// TestGetContent_AutoWithOnlyL2 测试 Auto 模式只有 L2
func TestGetContent_AutoWithOnlyL2(t *testing.T) {
	ctxt := NewContext(ContextTypeFile, "test.txt")
	ctxt.Layers.L2 = &LayerDetails{Content: "details"}

	content, err := ctxt.GetContent(LayerAuto)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if content != "details" {
		t.Errorf("expected 'details', got '%s'", content)
	}
}

// TestMemoryCollection_GetMemoryCount_Empty 测试空记忆集合的计数
func TestMemoryCollection_GetMemoryCount_Empty(t *testing.T) {
	mc := &MemoryCollection{
		// 所有切片都是空的
	}

	count := mc.GetMemoryCount()
	if count != 0 {
		t.Errorf("expected 0, got %d", count)
	}
}

// TestMemoryCollection_GetMemoriesByType_AllTypes 测试获取所有类型的记忆
func TestMemoryCollection_GetMemoriesByType_AllTypes(t *testing.T) {
	profiles := []*ProfileMemory{{ID: "p1"}}
	prefs := []*PreferenceMemory{{ID: "pref1"}}
	entities := []*EntityMemory{{ID: "e1"}}
	events := []*EventMemory{{ID: "ev1"}}
	cases := []*CaseMemory{{ID: "c1"}}
	patterns := []*PatternMemory{{ID: "pat1"}}

	mc := &MemoryCollection{
		Profiles:    profiles,
		Preferences: prefs,
		Entities:    entities,
		Events:      events,
		Cases:       cases,
		Patterns:    patterns,
	}

	// 测试每种类型
	result := mc.GetMemoriesByType(MemoryTypeProfile)
	if result == nil {
		t.Error("expected profiles to be returned")
	}

	result = mc.GetMemoriesByType(MemoryTypePreference)
	if result == nil {
		t.Error("expected preferences to be returned")
	}

	result = mc.GetMemoriesByType(MemoryTypeEntity)
	if result == nil {
		t.Error("expected entities to be returned")
	}

	result = mc.GetMemoriesByType(MemoryTypeEvent)
	if result == nil {
		t.Error("expected events to be returned")
	}

	result = mc.GetMemoriesByType(MemoryTypeCase)
	if result == nil {
		t.Error("expected cases to be returned")
	}

	result = mc.GetMemoriesByType(MemoryTypePattern)
	if result == nil {
		t.Error("expected patterns to be returned")
	}
}

// TestContext_GetContentSize_Empty 测试空内容的大小
func TestContext_GetContentSize_Empty(t *testing.T) {
	ctxt := NewContext(ContextTypeFile, "test.txt")
	ctxt.Layers.L2 = &LayerDetails{Content: ""}

	size := ctxt.GetContentSize()
	if size != 0 {
		t.Errorf("expected 0, got %d", size)
	}
}

// TestContext_IsEmpty_WithContent 测试有内容时的 IsEmpty
func TestContext_IsEmpty_WithContent(t *testing.T) {
	ctxt := NewContext(ContextTypeFile, "test.txt")
	ctxt.Layers.L2 = &LayerDetails{Content: "some content"}

	if ctxt.IsEmpty() {
		t.Error("expected context to not be empty")
	}
}

// TestContext_String_WithDetails 测试带详情的 String 方法
func TestContext_String_WithDetails(t *testing.T) {
	ctxt := NewContext(ContextTypeProject, "Test Project")
	ctxt.Layers.L2 = &LayerDetails{Content: "details content"}

	str := ctxt.String()
	if str == "" {
		t.Error("expected String to return non-empty string")
	}
}

// TestNewContext_DifferentTypes 测试创建不同类型的上下文
func TestNewContext_DifferentTypes(t *testing.T) {
	types := []ContextType{
		ContextTypeProject,
		ContextTypeFile,
		ContextTypeCodebase,
		ContextTypeMemory,
		ContextTypeResource,
		ContextTypeSkill,
	}

	for _, ct := range types {
		ctxt := NewContext(ct, "test")
		if ctxt.Type != ct {
			t.Errorf("expected type %s, got %s", ct, ctxt.Type)
		}
		if ctxt.Title != "test" {
			t.Errorf("expected title 'test', got '%s'", ctxt.Title)
		}
	}
}

// TestContext_Layers_Operations 测试层级操作
func TestContext_Layers_Operations(t *testing.T) {
	ctxt := NewContext(ContextTypeFile, "test.txt")

	// 测试 HasLayer 对于不存在的层
	if ctxt.HasLayer(LayerType("invalid")) {
		t.Error("expected false for invalid layer type")
	}

	// 测试 HasLayer 当所有层都为 nil
	if ctxt.HasLayer(LayerTypeL0) {
		t.Error("expected L0 to not exist initially")
	}
	if ctxt.HasLayer(LayerTypeL1) {
		t.Error("expected L1 to not exist initially")
	}
	if ctxt.HasLayer(LayerTypeL2) {
		t.Error("expected L2 to not exist initially")
	}
}

// TestGetContent_AutoPriority 测试 Auto 模式的层级优先级
func TestGetContent_AutoPriority(t *testing.T) {
	ctxt := NewContext(ContextTypeFile, "test.txt")

	// 只有 L0 时应该返回 L0
	ctxt.Layers.L0 = &LayerSummary{Content: "L0 content"}
	content, err := ctxt.GetContent(LayerAuto)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if content != "L0 content" {
		t.Errorf("expected L0 content, got '%s'", content)
	}

	// 添加 L1 后应该返回 L1（优先级高于 L0）
	ctxt.Layers.L1 = &LayerOverview{Content: "L1 content"}
	content, err = ctxt.GetContent(LayerAuto)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if content != "L1 content" {
		t.Errorf("expected L1 content, got '%s'", content)
	}
}

// TestGetTotalTokens_PartialLayers 测试部分层级的 token 计数
func TestGetTotalTokens_PartialLayers(t *testing.T) {
	ctxt := NewContext(ContextTypeFile, "test.txt")

	// 只有 L0
	ctxt.Layers.L0 = &LayerSummary{Tokens: 100}
	if total := ctxt.GetTotalTokens(); total != 100 {
		t.Errorf("expected 100, got %d", total)
	}

	// 添加 L1
	ctxt.Layers.L1 = &LayerOverview{Tokens: 200}
	if total := ctxt.GetTotalTokens(); total != 300 {
		t.Errorf("expected 300, got %d", total)
	}

	// 添加 L2
	ctxt.Layers.L2 = &LayerDetails{Tokens: 300}
	if total := ctxt.GetTotalTokens(); total != 600 {
		t.Errorf("expected 600, got %d", total)
	}
}

// TestGetAvailableLayers_Order 测试可用层级列表的顺序
func TestGetAvailableLayers_Order(t *testing.T) {
	ctxt := NewContext(ContextTypeFile, "test.txt")

	// 按顺序添加层级
	ctxt.Layers.L2 = &LayerDetails{Content: "details"}
	ctxt.Layers.L0 = &LayerSummary{Content: "summary"}
	ctxt.Layers.L1 = &LayerOverview{Content: "overview"}

	layers := ctxt.GetAvailableLayers()

	// 应该有 3 个层级
	if len(layers) != 3 {
		t.Errorf("expected 3 layers, got %d", len(layers))
	}

	// 检查是否包含所有三个层级类型
	hasL0, hasL1, hasL2 := false, false, false
	for _, layer := range layers {
		if layer == LayerTypeL0 {
			hasL0 = true
		}
		if layer == LayerTypeL1 {
			hasL1 = true
		}
		if layer == LayerTypeL2 {
			hasL2 = true
		}
	}

	if !hasL0 || !hasL1 || !hasL2 {
		t.Error("expected all three layer types to be available")
	}
}

// TestMemoryCollection_GetMemoriesByType_EmptySlices 测试空切片的获取
func TestMemoryCollection_GetMemoriesByType_EmptySlices(t *testing.T) {
	mc := &MemoryCollection{
		Profiles:    []*ProfileMemory{},    // 空切片
		Preferences: []*PreferenceMemory{}, // 空切片
		Entities:    []*EntityMemory{},     // 空切片
		Events:      []*EventMemory{},      // 空切片
		Cases:       []*CaseMemory{},       // 空切片
		Patterns:    []*PatternMemory{},    // 空切片
	}

	// 所有切片都是空的，不应该返回 nil
	profiles := mc.GetMemoriesByType(MemoryTypeProfile)
	if profiles == nil {
		t.Error("expected non-nil result for empty profiles slice")
	}

	prefs := mc.GetMemoriesByType(MemoryTypePreference)
	if prefs == nil {
		t.Error("expected non-nil result for empty preferences slice")
	}

	// 检查长度是否为 0
	if p, ok := profiles.([]*ProfileMemory); ok {
		if len(p) != 0 {
			t.Errorf("expected empty slice, got %d items", len(p))
		}
	}
}
