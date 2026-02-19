// Agent Framework - Context Types Methods Tests
// Copyright (C) 2025  Agent Framework Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package context

import (
	"testing"
	"time"
)

func TestNewContext(t *testing.T) {
	ctx := NewContext(ContextTypeFile, "test-context")

	if ctx == nil {
		t.Fatal("expected context to be created")
	}

	if ctx.Type != ContextTypeFile {
		t.Errorf("expected type %s, got %s", ContextTypeFile, ctx.Type)
	}

	if ctx.Title != "test-context" {
		t.Errorf("expected title 'test-context', got '%s'", ctx.Title)
	}

	if ctx.Version != 1 {
		t.Errorf("expected version 1, got %d", ctx.Version)
	}

	if ctx.Metadata == nil {
		t.Error("expected metadata to be initialized")
	}

	if ctx.TaskRefs == nil {
		t.Error("expected task refs to be initialized")
	}

	if ctx.CreatedAt.IsZero() {
		t.Error("expected created_at to be set")
	}

	if ctx.UpdatedAt.IsZero() {
		t.Error("expected updated_at to be set")
	}

	if ctx.AccessedAt.IsZero() {
		t.Error("expected accessed_at to be set")
	}
}

func TestContext_GenerateID(t *testing.T) {
	ctx := NewContext(ContextTypeFile, "test-context")
	ctx.Workspace = "/test/workspace"

	id := ctx.GenerateID()
	if id == "" {
		t.Error("expected ID to be generated")
	}

	if len(id) != 16 {
		t.Errorf("expected ID length 16, got %d", len(id))
	}

	// Same context should generate same ID
	ctx2 := NewContext(ContextTypeFile, "test-context")
	ctx2.Workspace = "/test/workspace"

	// Different created_at time will result in different ID
	// So we need to set the same created_at time
	ctx2.CreatedAt = ctx.CreatedAt

	id2 := ctx2.GenerateID()
	if id != id2 {
		t.Errorf("expected same ID for same context, got %s and %s", id, id2)
	}

	// Different context should generate different ID
	ctx3 := NewContext(ContextTypeFile, "different-context")
	id3 := ctx3.GenerateID()
	if id == id3 {
		t.Error("expected different ID for different context")
	}
}

func TestContext_GetContent(t *testing.T) {
	ctx := NewContext(ContextTypeFile, "test-context")

	// Test with no layers
	content, err := ctx.GetContent(LayerTypeL0)
	if err == nil {
		t.Error("expected error when getting content from non-existent layer")
	}
	if content != "" {
		t.Errorf("expected empty content, got '%s'", content)
	}

	// Test with L0 layer
	ctx.Layers.L0 = &LayerSummary{Content: "L0 content"}
	content, err = ctx.GetContent(LayerTypeL0)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if content != "L0 content" {
		t.Errorf("expected 'L0 content', got '%s'", content)
	}

	// Test with L1 layer
	ctx.Layers.L1 = &LayerOverview{Content: "L1 content"}
	content, err = ctx.GetContent(LayerTypeL1)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if content != "L1 content" {
		t.Errorf("expected 'L1 content', got '%s'", content)
	}

	// Test with L2 layer
	ctx.Layers.L2 = &LayerDetails{Content: "L2 content"}
	content, err = ctx.GetContent(LayerTypeL2)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if content != "L2 content" {
		t.Errorf("expected 'L2 content', got '%s'", content)
	}

	// Test with LayerAuto
	ctx2 := NewContext(ContextTypeFile, "test-context")
	ctx2.Layers.L1 = &LayerOverview{Content: "L1 content"}
	content, err = ctx2.GetContent(LayerAuto)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if content != "L1 content" {
		t.Errorf("expected 'L1 content', got '%s'", content)
	}
}

func TestContext_GetTotalTokens(t *testing.T) {
	ctx := NewContext(ContextTypeFile, "test-context")

	// Test with no layers
	total := ctx.GetTotalTokens()
	if total != 0 {
		t.Errorf("expected 0 tokens, got %d", total)
	}

	// Test with layers
	ctx.Layers.L0 = &LayerSummary{Tokens: 100}
	ctx.Layers.L1 = &LayerOverview{Tokens: 2000}
	ctx.Layers.L2 = &LayerDetails{Tokens: 5000}

	total = ctx.GetTotalTokens()
	if total != 7100 {
		t.Errorf("expected 7100 tokens, got %d", total)
	}
}

func TestContext_HasLayer(t *testing.T) {
	ctx := NewContext(ContextTypeFile, "test-context")

	// Test with no layers
	if ctx.HasLayer(LayerTypeL0) {
		t.Error("expected L0 layer to not exist")
	}
	if ctx.HasLayer(LayerTypeL1) {
		t.Error("expected L1 layer to not exist")
	}
	if ctx.HasLayer(LayerTypeL2) {
		t.Error("expected L2 layer to not exist")
	}

	// Test with layers
	ctx.Layers.L0 = &LayerSummary{}
	if !ctx.HasLayer(LayerTypeL0) {
		t.Error("expected L0 layer to exist")
	}

	ctx.Layers.L1 = &LayerOverview{}
	if !ctx.HasLayer(LayerTypeL1) {
		t.Error("expected L1 layer to exist")
	}

	ctx.Layers.L2 = &LayerDetails{}
	if !ctx.HasLayer(LayerTypeL2) {
		t.Error("expected L2 layer to exist")
	}
}

func TestContext_GetAvailableLayers(t *testing.T) {
	ctx := NewContext(ContextTypeFile, "test-context")

	// Test with no layers
	layers := ctx.GetAvailableLayers()
	if len(layers) != 0 {
		t.Errorf("expected 0 layers, got %d", len(layers))
	}

	// Test with L0 layer
	ctx.Layers.L0 = &LayerSummary{}
	layers = ctx.GetAvailableLayers()
	if len(layers) != 1 {
		t.Errorf("expected 1 layer, got %d", len(layers))
	}
	if layers[0] != LayerTypeL0 {
		t.Errorf("expected L0 layer, got %s", layers[0])
	}

	// Test with L0 and L1 layers
	ctx.Layers.L1 = &LayerOverview{}
	layers = ctx.GetAvailableLayers()
	if len(layers) != 2 {
		t.Errorf("expected 2 layers, got %d", len(layers))
	}

	// Test with all layers
	ctx.Layers.L2 = &LayerDetails{}
	layers = ctx.GetAvailableLayers()
	if len(layers) != 3 {
		t.Errorf("expected 3 layers, got %d", len(layers))
	}
}

func TestContext_UpdateAccessTime(t *testing.T) {
	ctx := NewContext(ContextTypeFile, "test-context")
	originalTime := ctx.AccessedAt

	// Wait a bit to ensure time difference
	time.Sleep(10 * time.Millisecond)

	ctx.UpdateAccessTime()

	if !ctx.AccessedAt.After(originalTime) {
		t.Error("expected accessed_at to be updated")
	}
}

func TestContext_IncrementVersion(t *testing.T) {
	ctx := NewContext(ContextTypeFile, "test-context")
	originalVersion := ctx.Version
	originalTime := ctx.UpdatedAt

	// Wait a bit to ensure time difference
	time.Sleep(10 * time.Millisecond)

	ctx.IncrementVersion()

	if ctx.Version != originalVersion+1 {
		t.Errorf("expected version %d, got %d", originalVersion+1, ctx.Version)
	}

	if !ctx.UpdatedAt.After(originalTime) {
		t.Error("expected updated_at to be updated")
	}
}

func TestMemoryCollection_GetMemoryCount(t *testing.T) {
	mc := &MemoryCollection{}

	// Test with empty collection
	count := mc.GetMemoryCount()
	if count != 0 {
		t.Errorf("expected 0 memories, got %d", count)
	}

	// Test with memories
	mc.Profiles = []*ProfileMemory{{ID: "1"}}
	mc.Preferences = []*PreferenceMemory{{ID: "1"}, {ID: "2"}}
	mc.Entities = []*EntityMemory{{ID: "1"}, {ID: "2"}, {ID: "3"}}
	mc.Events = []*EventMemory{{ID: "1"}}
	mc.Cases = []*CaseMemory{{ID: "1"}, {ID: "2"}}
	mc.Patterns = []*PatternMemory{{ID: "1"}}

	count = mc.GetMemoryCount()
	if count != 10 {
		t.Errorf("expected 10 memories, got %d", count)
	}

	// Test with nil collection
	var nilMC *MemoryCollection
	count = nilMC.GetMemoryCount()
	if count != 0 {
		t.Errorf("expected 0 memories for nil collection, got %d", count)
	}
}

func TestMemoryCollection_GetMemoriesByType(t *testing.T) {
	mc := &MemoryCollection{
		Profiles:    []*ProfileMemory{{ID: "1"}},
		Preferences: []*PreferenceMemory{{ID: "2"}},
		Entities:    []*EntityMemory{{ID: "3"}},
		Events:      []*EventMemory{{ID: "4"}},
		Cases:       []*CaseMemory{{ID: "5"}},
		Patterns:    []*PatternMemory{{ID: "6"}},
	}

	// Test getting profiles
	profiles := mc.GetMemoriesByType(MemoryTypeProfile)
	if profiles == nil {
		t.Error("expected profiles to be returned")
	}
	if len(profiles.([]*ProfileMemory)) != 1 {
		t.Errorf("expected 1 profile, got %d", len(profiles.([]*ProfileMemory)))
	}

	// Test getting preferences
	prefs := mc.GetMemoriesByType(MemoryTypePreference)
	if prefs == nil {
		t.Error("expected preferences to be returned")
	}

	// Test getting entities
	entities := mc.GetMemoriesByType(MemoryTypeEntity)
	if entities == nil {
		t.Error("expected entities to be returned")
	}

	// Test getting events
	events := mc.GetMemoriesByType(MemoryTypeEvent)
	if events == nil {
		t.Error("expected events to be returned")
	}

	// Test getting cases
	cases := mc.GetMemoriesByType(MemoryTypeCase)
	if cases == nil {
		t.Error("expected cases to be returned")
	}

	// Test getting patterns
	patterns := mc.GetMemoriesByType(MemoryTypePattern)
	if patterns == nil {
		t.Error("expected patterns to be returned")
	}

	// Test with nil collection
	var nilMC *MemoryCollection
	result := nilMC.GetMemoriesByType(MemoryTypeProfile)
	if result != nil {
		t.Error("expected nil for nil collection")
	}
}

func TestMemoryCollection_AddMemory(t *testing.T) {
	mc := &MemoryCollection{}

	// Test adding profile
	profile := &ProfileMemory{ID: "1", Name: "Test Profile"}
	err := mc.AddMemory(MemoryTypeProfile, profile)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(mc.Profiles) != 1 {
		t.Errorf("expected 1 profile, got %d", len(mc.Profiles))
	}

	// Test adding preference
	pref := &PreferenceMemory{ID: "2", Key: "test", Value: "value"}
	err = mc.AddMemory(MemoryTypePreference, pref)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(mc.Preferences) != 1 {
		t.Errorf("expected 1 preference, got %d", len(mc.Preferences))
	}

	// Test adding entity
	entity := &EntityMemory{ID: "3", Name: "Test Entity", Type: "person"}
	err = mc.AddMemory(MemoryTypeEntity, entity)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(mc.Entities) != 1 {
		t.Errorf("expected 1 entity, got %d", len(mc.Entities))
	}

	// Test adding event
	event := &EventMemory{ID: "4", Title: "Test Event"}
	err = mc.AddMemory(MemoryTypeEvent, event)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(mc.Events) != 1 {
		t.Errorf("expected 1 event, got %d", len(mc.Events))
	}

	// Test adding case
	caseMem := &CaseMemory{ID: "5", Domain: "test"}
	err = mc.AddMemory(MemoryTypeCase, caseMem)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(mc.Cases) != 1 {
		t.Errorf("expected 1 case, got %d", len(mc.Cases))
	}

	// Test adding pattern
	pattern := &PatternMemory{ID: "6", Pattern: "test pattern"}
	err = mc.AddMemory(MemoryTypePattern, pattern)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(mc.Patterns) != 1 {
		t.Errorf("expected 1 pattern, got %d", len(mc.Patterns))
	}

	// Test adding invalid type
	err = mc.AddMemory(MemoryType("invalid"), "test")
	if err == nil {
		t.Error("expected error for invalid memory type")
	}

	// Test with nil collection
	var nilMC *MemoryCollection
	err = nilMC.AddMemory(MemoryTypeProfile, profile)
	if err == nil {
		t.Error("expected error for nil collection")
	}
}

func TestMemoryCollection_Clear(t *testing.T) {
	mc := &MemoryCollection{
		Profiles:    []*ProfileMemory{{ID: "1"}},
		Preferences: []*PreferenceMemory{{ID: "2"}},
		Entities:    []*EntityMemory{{ID: "3"}},
		Events:      []*EventMemory{{ID: "4"}},
		Cases:       []*CaseMemory{{ID: "5"}},
		Patterns:    []*PatternMemory{{ID: "6"}},
	}

	mc.Clear()

	if mc.Profiles != nil {
		t.Error("expected profiles to be cleared")
	}
	if mc.Preferences != nil {
		t.Error("expected preferences to be cleared")
	}
	if mc.Entities != nil {
		t.Error("expected entities to be cleared")
	}
	if mc.Events != nil {
		t.Error("expected events to be cleared")
	}
	if mc.Cases != nil {
		t.Error("expected cases to be cleared")
	}
	if mc.Patterns != nil {
		t.Error("expected patterns to be cleared")
	}

	// Test with nil collection (should not panic)
	var nilMC *MemoryCollection
	nilMC.Clear()
}

func TestMemoryCollection_IsEmpty(t *testing.T) {
	mc := &MemoryCollection{}

	// Test with empty collection
	if !mc.IsEmpty() {
		t.Error("expected collection to be empty")
	}

	// Test with memories
	mc.Profiles = []*ProfileMemory{{ID: "1"}}
	if mc.IsEmpty() {
		t.Error("expected collection to not be empty")
	}

	// Test with nil collection
	var nilMC *MemoryCollection
	if !nilMC.IsEmpty() {
		t.Error("expected nil collection to be empty")
	}
}

func TestMemory_GetID(t *testing.T) {
	// Test ProfileMemory
	profile := &ProfileMemory{ID: "test-id"}
	if profile.GetID() != "test-id" {
		t.Errorf("expected 'test-id', got '%s'", profile.GetID())
	}

	// Test PreferenceMemory
	pref := &PreferenceMemory{ID: "pref-id"}
	if pref.GetID() != "pref-id" {
		t.Errorf("expected 'pref-id', got '%s'", pref.GetID())
	}

	// Test EntityMemory
	entity := &EntityMemory{ID: "entity-id"}
	if entity.GetID() != "entity-id" {
		t.Errorf("expected 'entity-id', got '%s'", entity.GetID())
	}

	// Test EventMemory
	event := &EventMemory{ID: "event-id"}
	if event.GetID() != "event-id" {
		t.Errorf("expected 'event-id', got '%s'", event.GetID())
	}

	// Test CaseMemory
	caseMem := &CaseMemory{ID: "case-id"}
	if caseMem.GetID() != "case-id" {
		t.Errorf("expected 'case-id', got '%s'", caseMem.GetID())
	}

	// Test PatternMemory
	pattern := &PatternMemory{ID: "pattern-id"}
	if pattern.GetID() != "pattern-id" {
		t.Errorf("expected 'pattern-id', got '%s'", pattern.GetID())
	}
}

func TestEnhancedMemoryCollection_GetTier(t *testing.T) {
	emc := &EnhancedMemoryCollection{
		MemoryCollection: &MemoryCollection{},
		MemoryMetadata: make(map[string]TieredMemory),
	}

	// Test with no metadata
	tier := emc.GetTier("non-existent")
	if tier != MemoryTierSession {
		t.Errorf("expected default tier %s, got %s", MemoryTierSession, tier)
	}

	// Test with metadata
	emc.MemoryMetadata["test-id"] = TieredMemory{
		Tier: MemoryTierLongTerm,
	}

	tier = emc.GetTier("test-id")
	if tier != MemoryTierLongTerm {
		t.Errorf("expected tier %s, got %s", MemoryTierLongTerm, tier)
	}
}

func TestEnhancedMemoryCollection_SetTier(t *testing.T) {
	emc := &EnhancedMemoryCollection{
		MemoryCollection: &MemoryCollection{},
	}

	// Set tier for a memory
	emc.SetTier("test-id", MemoryTierDaily, time.Now().Add(24*time.Hour), 0.8)

	if emc.MemoryMetadata == nil {
		t.Fatal("expected memory metadata to be initialized")
	}

	meta, ok := emc.MemoryMetadata["test-id"]
	if !ok {
		t.Error("expected metadata to be set")
	}

	if meta.Tier != MemoryTierDaily {
		t.Errorf("expected tier %s, got %s", MemoryTierDaily, meta.Tier)
	}
}

func TestEnhancedMemoryCollection_IsEmpty(t *testing.T) {
	emc := &EnhancedMemoryCollection{
		MemoryCollection: &MemoryCollection{},
	}

	// Test with empty collection
	if !emc.IsEmpty() {
		t.Error("expected enhanced memory collection to be empty")
	}

	// Test with memories
	emc.Profiles = []*ProfileMemory{{ID: "1"}}
	if emc.IsEmpty() {
		t.Error("expected enhanced memory collection to not be empty")
	}
}

func TestEnhancedMemoryCollection_GetMemoryCount(t *testing.T) {
	emc := &EnhancedMemoryCollection{
		MemoryCollection: &MemoryCollection{},
	}

	// Test with empty collection
	count := emc.GetMemoryCount()
	if count != 0 {
		t.Errorf("expected 0 memories, got %d", count)
	}

	// Test with memories
	emc.Profiles = []*ProfileMemory{{ID: "1"}}
	emc.Preferences = []*PreferenceMemory{{ID: "1"}, {ID: "2"}}

	count = emc.GetMemoryCount()
	if count != 3 {
		t.Errorf("expected 3 memories, got %d", count)
	}
}

func TestEnhancedMemoryCollection_GetSessionMemories(t *testing.T) {
	emc := &EnhancedMemoryCollection{
		MemoryCollection: &MemoryCollection{
			Profiles: []*ProfileMemory{
				{ID: "1"},
				{ID: "2"},
			},
		},
		MemoryMetadata: make(map[string]TieredMemory),
	}

	// Set tier for memories
	emc.MemoryMetadata["1"] = TieredMemory{Tier: MemoryTierSession}
	emc.MemoryMetadata["2"] = TieredMemory{Tier: MemoryTierDaily}

	sessionMemories := emc.GetSessionMemories()
	if sessionMemories == nil {
		t.Error("expected session memories to be returned")
	}

	// Note: GetSessionMemories returns the entire collection, not filtered by tier
	// The tier filtering is done through metadata lookup
}

func TestEnhancedMemoryCollection_GetDailyMemories(t *testing.T) {
	emc := &EnhancedMemoryCollection{
		MemoryCollection: &MemoryCollection{
			Profiles: []*ProfileMemory{
				{ID: "1"},
				{ID: "2"},
			},
		},
		MemoryMetadata: make(map[string]TieredMemory),
	}

	dailyMemories := emc.GetDailyMemories()
	if dailyMemories == nil {
		t.Fatal("expected daily memories to be returned")
	}
	if len(dailyMemories.Profiles) != 2 {
		t.Errorf("expected 2 profiles, got %d", len(dailyMemories.Profiles))
	}
}

func TestEnhancedMemoryCollection_GetLongTermMemories(t *testing.T) {
	emc := &EnhancedMemoryCollection{
		MemoryCollection: &MemoryCollection{
			Profiles: []*ProfileMemory{
				{ID: "1"},
				{ID: "2"},
			},
			Preferences: []*PreferenceMemory{
				{ID: "3"},
			},
		},
		MemoryMetadata: make(map[string]TieredMemory),
	}

	longTermMemories := emc.GetLongTermMemories()
	if longTermMemories == nil {
		t.Fatal("expected long term memories to be returned")
	}
	if len(longTermMemories.Profiles) != 2 {
		t.Errorf("expected 2 profiles, got %d", len(longTermMemories.Profiles))
	}
	if len(longTermMemories.Preferences) != 1 {
		t.Errorf("expected 1 preference, got %d", len(longTermMemories.Preferences))
	}
}
