// Agent Framework - Context Memory Tests
// Copyright (C) 2025  Agent Framework Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package memory

import (
	"strings"
	"testing"

	beadscontext "AgentFramework/pkg/beads/context"
)

// TestNewExtractor 测试创建提取器
func TestNewExtractor(t *testing.T) {
	e := NewExtractor()
	if e == nil {
		t.Fatal("expected extractor to be created")
	}
	if e.profileExtractor == nil {
		t.Error("expected profile extractor to be initialized")
	}
	if e.preferenceExtractor == nil {
		t.Error("expected preference extractor to be initialized")
	}
	if e.entityExtractor == nil {
		t.Error("expected entity extractor to be initialized")
	}
	if e.eventExtractor == nil {
		t.Error("expected event extractor to be initialized")
	}
	if e.caseExtractor == nil {
		t.Error("expected case extractor to be initialized")
	}
	if e.patternExtractor == nil {
		t.Error("expected pattern extractor to be initialized")
	}
}

// TestExtractFromContext_NilL2 测试从没有 L2 层的上下文提取
func TestExtractFromContext_NilL2(t *testing.T) {
	e := NewExtractor()
	ctx := beadscontext.Context{}

	ctxt := &beadscontext.Context{
		ID:    "test",
		Type:  beadscontext.ContextTypeFile,
		Title: "Test",
		Layers: beadscontext.ContextLayers{
			L2: nil,
		},
	}

	collection, err := e.ExtractFromContext(ctx, ctxt)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if collection == nil {
		t.Error("expected collection to be returned")
	}
}

// TestExtractFromContext_WithL2 测试从有 L2 层的上下文提取
func TestExtractFromContext_WithL2(t *testing.T) {
	e := NewExtractor()
	ctx := beadscontext.Context{}

	ctxt := &beadscontext.Context{
		ID:    "test",
		Type:  beadscontext.ContextTypeFile,
		Title: "Test",
		Layers: beadscontext.ContextLayers{
			L2: &beadscontext.LayerDetails{
				Content: "我是张三，是一名开发者。",
				Format:  "plain",
			},
		},
	}

	collection, err := e.ExtractFromContext(ctx, ctxt)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if collection == nil {
		t.Error("expected collection to be returned")
	}
}

// TestExtractFromText_EmptyText 测试从空文本提取
func TestExtractFromText_EmptyText(t *testing.T) {
	e := NewExtractor()
	ctx := beadscontext.Context{}

	collection, err := e.ExtractFromText(ctx, "", "plain")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if collection == nil {
		t.Error("expected collection to be returned")
	}
	if len(collection.Profiles) != 0 {
		t.Error("expected no profiles from empty text")
	}
}

// TestExtractFromConversation 测试从对话提取
func TestExtractFromConversation(t *testing.T) {
	e := NewExtractor()
	ctx := beadscontext.Context{}

	messages := []beadscontext.Message{
		{Role: "user", Content: "Hello"},
		{Role: "assistant", Content: "Hi there"},
	}

	collection, err := e.ExtractFromConversation(ctx, messages)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if collection == nil {
		t.Error("expected collection to be returned")
	}
}

// TestExtractFromConversation_Empty 测试从空对话提取
func TestExtractFromConversation_Empty(t *testing.T) {
	e := NewExtractor()
	ctx := beadscontext.Context{}

	collection, err := e.ExtractFromConversation(ctx, []beadscontext.Message{})
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if collection == nil {
		t.Error("expected collection to be returned")
	}
}

// TestMergeMemories 测试合并记忆
func TestMergeMemories(t *testing.T) {
	e := NewExtractor()
	ctx := beadscontext.Context{}

	existing := &beadscontext.MemoryCollection{
		Profiles: []*beadscontext.ProfileMemory{
			{ID: "1", Name: "User1"},
		},
	}

	new := &beadscontext.MemoryCollection{
		Profiles: []*beadscontext.ProfileMemory{
			{ID: "2", Name: "User2"},
			{ID: "1", Name: "User1"}, // Duplicate ID
		},
	}

	result, err := e.Merge(ctx, existing, new)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(result.Profiles) != 2 { // Should deduplicate
		t.Errorf("expected 2 profiles after deduplication, got %d", len(result.Profiles))
	}
}

// TestDeduplicate 测试去重
func TestDeduplicate(t *testing.T) {
	e := NewExtractor()
	ctx := beadscontext.Context{}

	memories := &beadscontext.MemoryCollection{
		Profiles: []*beadscontext.ProfileMemory{
			{ID: "1", Name: "User1"},
			{ID: "1", Name: "User1 Updated"},
		},
	}

	result, err := e.Deduplicate(ctx, memories)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(result.Profiles) != 1 {
		t.Errorf("expected 1 profile after deduplication, got %d", len(result.Profiles))
	}
}

// TestDeduplicateSlice 测试切片去重
func TestDeduplicateSlice(t *testing.T) {
	items := []*beadscontext.ProfileMemory{
		{ID: "1", Name: "User1"},
		{ID: "2", Name: "User2"},
		{ID: "1", Name: "User1"}, // Duplicate
	}

	result := deduplicateSlice(items)
	if len(result) != 2 {
		t.Errorf("expected 2 items after deduplication, got %d", len(result))
	}
}

// TestDeduplicateSlice_Empty 测试空切片去重
func TestDeduplicateSlice_Empty(t *testing.T) {
	items := []*beadscontext.ProfileMemory{}
	result := deduplicateSlice(items)
	if len(result) != 0 {
		t.Errorf("expected 0 items, got %d", len(result))
	}
}

// ===== ProfileExtractor Tests =====

// TestNewProfileExtractor 测试创建 Profile 提取器
func TestNewProfileExtractor(t *testing.T) {
	e := NewProfileExtractor()
	if e == nil {
		t.Fatal("expected profile extractor to be created")
	}
}

// TestProfileExtractor_Extract 测试提取 Profile
func TestProfileExtractor_Extract(t *testing.T) {
	e := NewProfileExtractor()

	// Use text where regex captures the name (may include trailing punctuation)
	text := "我是张三。"
	profiles := e.Extract(text)

	if len(profiles) == 0 {
		t.Error("expected at least one profile to be extracted")
	} else {
		// Regex captures everything after "我是" including punctuation
		if profiles[0].Name != "张三。" {
			t.Errorf("expected name '张三。', got '%s'", profiles[0].Name)
		}
		if profiles[0].Role != "user" {
			t.Errorf("expected role 'user', got '%s'", profiles[0].Role)
		}
	}
}

// TestProfileExtractor_Extract_Empty 测试从空文本提取 Profile
func TestProfileExtractor_Extract_Empty(t *testing.T) {
	e := NewProfileExtractor()

	profiles := e.Extract("")
	if len(profiles) != 0 {
		t.Errorf("expected 0 profiles from empty text, got %d", len(profiles))
	}
}

// TestProfileExtractor_DetectRole 测试检测角色
func TestProfileExtractor_DetectRole(t *testing.T) {
	e := NewProfileExtractor()

	tests := []struct {
		text     string
		expected string
	}{
		{"我是一名开发者", "developer"},
		{"我是学生", "student"},
		{"我是老师", "teacher"},
		{"我是用户", "user"},
		{"I am a developer", "user"}, // No Chinese keyword
	}

	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			role := e.detectRole(tt.text)
			if role != tt.expected {
				t.Errorf("expected role '%s', got '%s'", tt.expected, role)
			}
		})
	}
}

// TestProfileExtractor_GenerateID 测试生成 ID
func TestProfileExtractor_GenerateID(t *testing.T) {
	e := NewProfileExtractor()

	id1 := e.generateID("test")
	id2 := e.generateID("test")

	if id1 != id2 {
		t.Error("expected same ID for same content")
	}
	if len(id1) != 16 {
		t.Errorf("expected ID length 16, got %d", len(id1))
	}

	id3 := e.generateID("different")
	if id1 == id3 {
		t.Error("expected different IDs for different content")
	}
}

// ===== PreferenceExtractor Tests =====

// TestNewPreferenceExtractor 测试创建 Preference 提取器
func TestNewPreferenceExtractor(t *testing.T) {
	e := NewPreferenceExtractor()
	if e == nil {
		t.Fatal("expected preference extractor to be created")
	}
}

// TestPreferenceExtractor_Extract 测试提取 Preference
func TestPreferenceExtractor_Extract(t *testing.T) {
	e := NewPreferenceExtractor()

	text := "我喜欢用Go编程，我偏好深色主题。"
	prefs := e.Extract(text)

	if len(prefs) == 0 {
		t.Error("expected at least one preference to be extracted")
	}
}

// TestPreferenceExtractor_Extract_Empty 测试从空文本提取 Preference
func TestPreferenceExtractor_Extract_Empty(t *testing.T) {
	e := NewPreferenceExtractor()

	prefs := e.Extract("")
	if len(prefs) != 0 {
		t.Errorf("expected 0 preferences from empty text, got %d", len(prefs))
	}
}

// TestPreferenceExtractor_GenerateID 测试生成 ID
func TestPreferenceExtractor_GenerateID(t *testing.T) {
	e := NewPreferenceExtractor()

	id := e.generateID("test")
	if len(id) != 16 {
		t.Errorf("expected ID length 16, got %d", len(id))
	}
}

// ===== EntityExtractor Tests =====

// TestNewEntityExtractor 测试创建 Entity 提取器
func TestNewEntityExtractor(t *testing.T) {
	e := NewEntityExtractor()
	if e == nil {
		t.Fatal("expected entity extractor to be created")
	}
}

// TestEntityExtractor_Extract 测试提取 Entity
func TestEntityExtractor_Extract(t *testing.T) {
	e := NewEntityExtractor()

	text := "张三和李四在使用Go和Python开发。"
	entities := e.Extract(text)

	if len(entities) == 0 {
		t.Error("expected at least one entity to be extracted")
	}
}

// TestEntityExtractor_Extract_Empty 测试从空文本提取 Entity
func TestEntityExtractor_Extract_Empty(t *testing.T) {
	e := NewEntityExtractor()

	entities := e.Extract("")
	if len(entities) != 0 {
		t.Errorf("expected 0 entities from empty text, got %d", len(entities))
	}
}

// TestEntityExtractor_ExtractTechTerms 测试提取技术术语
func TestEntityExtractor_ExtractTechTerms(t *testing.T) {
	e := NewEntityExtractor()

	text := "我们使用Go和Python开发，使用React和Vue做前端。"
	terms := e.extractTechTerms(text)

	if len(terms) == 0 {
		t.Error("expected at least one tech term to be extracted")
	}
}

// TestEntityExtractor_GenerateID 测试生成 ID
func TestEntityExtractor_GenerateID(t *testing.T) {
	e := NewEntityExtractor()

	id := e.generateID("test")
	if len(id) != 16 {
		t.Errorf("expected ID length 16, got %d", len(id))
	}
}

// ===== EventExtractor Tests =====

// TestNewEventExtractor 测试创建 Event 提取器
func TestNewEventExtractor(t *testing.T) {
	e := NewEventExtractor()
	if e == nil {
		t.Fatal("expected event extractor to be created")
	}
}

// TestEventExtractor_Extract 测试提取 Event
func TestEventExtractor_Extract(t *testing.T) {
	e := NewEventExtractor()

	text := "我们决定使用Go，今天开始开发。"
	events := e.Extract(text)

	if len(events) == 0 {
		t.Error("expected at least one event to be extracted")
	}
}

// TestEventExtractor_Extract_Empty 测试从空文本提取 Event
func TestEventExtractor_Extract_Empty(t *testing.T) {
	e := NewEventExtractor()

	events := e.Extract("")
	if len(events) != 0 {
		t.Errorf("expected 0 events from empty text, got %d", len(events))
	}
}

// TestEventExtractor_ExtractTitle 测试提取事件标题
func TestEventExtractor_ExtractTitle(t *testing.T) {
	e := NewEventExtractor()

	shortText := "Short text"
	title := e.extractTitle("决定", shortText)
	if title != shortText {
		t.Errorf("expected title '%s', got '%s'", shortText, title)
	}

	longText := strings.Repeat("A", 100)
	title = e.extractTitle("决定", longText)
	if len(title) != 53 { // 50 + "..."
		t.Errorf("expected title length 53, got %d", len(title))
	}
}

// TestEventExtractor_GenerateID 测试生成 ID
func TestEventExtractor_GenerateID(t *testing.T) {
	e := NewEventExtractor()

	id := e.generateID("test")
	if len(id) != 16 {
		t.Errorf("expected ID length 16, got %d", len(id))
	}
}

// ===== CaseExtractor Tests =====

// TestNewCaseExtractor 测试创建 Case 提取器
func TestNewCaseExtractor(t *testing.T) {
	e := NewCaseExtractor()
	if e == nil {
		t.Fatal("expected case extractor to be created")
	}
}

// TestCaseExtractor_Extract 测试提取 Case
func TestCaseExtractor_Extract(t *testing.T) {
	e := NewCaseExtractor()

	// Use longer text to meet the 50+ character requirement
	text := "问题：程序启动时报错，显示无法连接到数据库服务器，导致应用无法正常启动和运行。解决：检查配置文件中的数据库连接字符串，发现端口号配置错误，修正后重新启动应用程序，现在可以正常连接数据库了。"
	cases := e.Extract(text)

	if len(cases) == 0 {
		t.Error("expected at least one case to be extracted")
	}
}

// TestCaseExtractor_Extract_Empty 测试从空文本提取 Case
func TestCaseExtractor_Extract_Empty(t *testing.T) {
	e := NewCaseExtractor()

	cases := e.Extract("")
	if len(cases) != 0 {
		t.Errorf("expected 0 cases from empty text, got %d", len(cases))
	}
}

// TestCaseExtractor_SplitProblemSolution 测试分离问题和解决方案
func TestCaseExtractor_SplitProblemSolution(t *testing.T) {
	e := NewCaseExtractor()

	// The separator "解决" will be used to split, so:
	// parts[0] = "问题：无法连接数据库。"
	// parts[1] = "检查配置。"
	text := "问题：无法连接数据库。解决：检查配置。"
	parts := e.splitProblemSolution(text)

	if parts == nil {
		t.Fatal("expected parts to be returned")
	}
	if len(parts) != 2 {
		t.Errorf("expected 2 parts, got %d", len(parts))
	}
	if !strings.Contains(parts[0], "问题") {
		t.Error("expected first part to contain problem")
	}
	// Second part contains the solution content, not the word "解决" which is the separator
	if !strings.Contains(parts[1], "检查") {
		t.Error("expected second part to contain solution content")
	}
}

// TestCaseExtractor_SplitProblemSolution_NoSeparator 测试没有分隔符
func TestCaseExtractor_SplitProblemSolution_NoSeparator(t *testing.T) {
	e := NewCaseExtractor()

	parts := e.splitProblemSolution("Just some text without separators")
	if parts != nil {
		t.Error("expected nil for text without separators")
	}
}

// TestCaseExtractor_DetectDomain 测试检测领域
func TestCaseExtractor_DetectDomain(t *testing.T) {
	e := NewCaseExtractor()

	tests := []struct {
		text     string
		expected string
	}{
		{"这是一个编程问题", "programming"},
		{"正在调试代码", "debugging"},
		{"系统设计", "design"},
		{"部署到生产环境", "deployment"},
		{"一些普通文本", "general"},
	}

	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			domain := e.detectDomain(tt.text)
			if domain != tt.expected {
				t.Errorf("expected domain '%s', got '%s'", tt.expected, domain)
			}
		})
	}
}

// TestCaseExtractor_ExtractLessons 测试提取教训
func TestCaseExtractor_ExtractLessons(t *testing.T) {
	e := NewCaseExtractor()

	text := "问题：无法连接。解决：修复。教训：记得检查配置。学到了要仔细测试。"
	lessons := e.extractLessons(text)

	if len(lessons) == 0 {
		t.Error("expected at least one lesson to be extracted")
	}
}

// TestCaseExtractor_ExtractTags 测试提取标签
func TestCaseExtractor_ExtractTags(t *testing.T) {
	e := NewCaseExtractor()

	tags := e.extractTags("这是一个bug修复case")
	if len(tags) == 0 {
		t.Error("expected some tags to be extracted")
	}

	hasBugTag := false
	for _, tag := range tags {
		if tag == "bug" {
			hasBugTag = true
		}
	}
	if !hasBugTag {
		t.Error("expected 'bug' tag to be present")
	}
}

// TestCaseExtractor_GenerateID 测试生成 ID
func TestCaseExtractor_GenerateID(t *testing.T) {
	e := NewCaseExtractor()

	id := e.generateID("test")
	if len(id) != 16 {
		t.Errorf("expected ID length 16, got %d", len(id))
	}
}

// ===== PatternExtractor Tests =====

// TestNewPatternExtractor 测试创建 Pattern 提取器
func TestNewPatternExtractor(t *testing.T) {
	e := NewPatternExtractor()
	if e == nil {
		t.Fatal("expected pattern extractor to be created")
	}
}

// TestPatternExtractor_Extract 测试提取 Pattern
func TestPatternExtractor_Extract(t *testing.T) {
	e := NewPatternExtractor()

	// Test with code patterns
	text := "func test() { return 1 } func another() { return 2 }"
	patterns := e.Extract(text)

	if len(patterns) == 0 {
		t.Error("expected at least one pattern to be extracted")
	}
}

// TestPatternExtractor_Extract_Empty 测试从空文本提取 Pattern
func TestPatternExtractor_Extract_Empty(t *testing.T) {
	e := NewPatternExtractor()

	patterns := e.Extract("")
	if len(patterns) != 0 {
		t.Errorf("expected 0 patterns from empty text, got %d", len(patterns))
	}
}

// TestPatternExtractor_DetectRepeatedPatterns 测试检测重复模式
func TestPatternExtractor_DetectRepeatedPatterns(t *testing.T) {
	e := NewPatternExtractor()

	// Use longer sentences (>20 chars) to be detected
	text := "这是一个非常长的模式描述文本。这是一个非常长的模式描述文本。另一个不同的内容描述。"
	patterns := e.detectRepeatedPatterns(text)

	if len(patterns) == 0 {
		t.Error("expected at least one repeated pattern to be detected")
	}
}

// TestPatternExtractor_DetectRepeatedPatterns_NoRepetition 测试没有重复
func TestPatternExtractor_DetectRepeatedPatterns_NoRepetition(t *testing.T) {
	e := NewPatternExtractor()

	text := "这是第一个内容。这是第二个内容。这是第三个内容。"
	patterns := e.detectRepeatedPatterns(text)

	if len(patterns) != 0 {
		t.Errorf("expected 0 patterns for non-repeating text, got %d", len(patterns))
	}
}

// TestPatternExtractor_DetectCodePatterns 测试检测代码模式
func TestPatternExtractor_DetectCodePatterns(t *testing.T) {
	e := NewPatternExtractor()

	text := "func test() { if true { return 1 } }"
	patterns := e.detectCodePatterns(text)

	if len(patterns) == 0 {
		t.Error("expected at least one code pattern to be detected")
	}
}

// TestPatternExtractor_CountPattern 测试计算模式出现次数
func TestPatternExtractor_CountPattern(t *testing.T) {
	e := NewPatternExtractor()

	text := "func func func test"
	count := e.countPattern(text, "func")
	if count != 3 {
		t.Errorf("expected count 3, got %d", count)
	}
}

// TestPatternExtractor_GenerateID 测试生成 ID
func TestPatternExtractor_GenerateID(t *testing.T) {
	e := NewPatternExtractor()

	id := e.generateID("test")
	if len(id) != 16 {
		t.Errorf("expected ID length 16, got %d", len(id))
	}
}

// ===== Integration Tests =====

// TestExtractFromContext_WithExistingMemories 测试从有现有记忆的上下文提取
func TestExtractFromContext_WithExistingMemories(t *testing.T) {
	e := NewExtractor()
	ctx := beadscontext.Context{}

	existingMemories := &beadscontext.MemoryCollection{
		Profiles: []*beadscontext.ProfileMemory{
			{ID: "existing-1", Name: "Existing User"},
		},
	}

	ctxt := &beadscontext.Context{
		ID:    "test",
		Type:  beadscontext.ContextTypeFile,
		Title: "Test",
		Layers: beadscontext.ContextLayers{
			L2: &beadscontext.LayerDetails{
				Content: "我是新用户。",
				Format:  "plain",
			},
		},
		Memories: existingMemories,
	}

	collection, err := e.ExtractFromContext(ctx, ctxt)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if collection == nil {
		t.Fatal("expected collection to be returned")
	}

	// Should have both existing and new profiles
	if len(collection.Profiles) != 2 {
		t.Errorf("expected 2 profiles (existing + new), got %d", len(collection.Profiles))
	}
}

// TestMergeMemories_AllTypes 测试合并所有类型的记忆
func TestMergeMemories_AllTypes(t *testing.T) {
	e := NewExtractor()
	ctx := beadscontext.Context{}

	existing := &beadscontext.MemoryCollection{
		Profiles:     []*beadscontext.ProfileMemory{{ID: "1"}},
		Preferences:  []*beadscontext.PreferenceMemory{{ID: "1"}},
		Entities:     []*beadscontext.EntityMemory{{ID: "1"}},
		Events:       []*beadscontext.EventMemory{{ID: "1"}},
		Cases:        []*beadscontext.CaseMemory{{ID: "1"}},
		Patterns:     []*beadscontext.PatternMemory{{ID: "1"}},
	}

	new := &beadscontext.MemoryCollection{
		Profiles:     []*beadscontext.ProfileMemory{{ID: "2"}},
		Preferences:  []*beadscontext.PreferenceMemory{{ID: "2"}},
		Entities:     []*beadscontext.EntityMemory{{ID: "2"}},
		Events:       []*beadscontext.EventMemory{{ID: "2"}},
		Cases:        []*beadscontext.CaseMemory{{ID: "2"}},
		Patterns:     []*beadscontext.PatternMemory{{ID: "2"}},
	}

	result, err := e.Merge(ctx, existing, new)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Check all types are merged
	if len(result.Profiles) != 2 {
		t.Errorf("expected 2 profiles, got %d", len(result.Profiles))
	}
	if len(result.Preferences) != 2 {
		t.Errorf("expected 2 preferences, got %d", len(result.Preferences))
	}
	if len(result.Entities) != 2 {
		t.Errorf("expected 2 entities, got %d", len(result.Entities))
	}
	if len(result.Events) != 2 {
		t.Errorf("expected 2 events, got %d", len(result.Events))
	}
	if len(result.Cases) != 2 {
		t.Errorf("expected 2 cases, got %d", len(result.Cases))
	}
	if len(result.Patterns) != 2 {
		t.Errorf("expected 2 patterns, got %d", len(result.Patterns))
	}
}

// TestExtractFromText_Comprehensive 测试从复杂文本提取各种记忆
func TestExtractFromText_Comprehensive(t *testing.T) {
	e := NewExtractor()
	ctx := beadscontext.Context{}

	// Use separate sentences to match extraction patterns
	text := "我是张三。我喜欢用Go编程。"

	collection, err := e.ExtractFromText(ctx, text, "plain")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if collection == nil {
		t.Fatal("expected collection to be returned")
	}

	// Verify extraction happened
	if len(collection.Profiles) == 0 {
		t.Error("expected at least one profile to be extracted")
	}
	if len(collection.Preferences) == 0 {
		t.Error("expected at least one preference to be extracted")
	}
}

// TestGetID_Implementations 测试所有类型的 GetID 实现
func TestGetID_Implementations(t *testing.T) {
	// ProfileMemory
	profile := &beadscontext.ProfileMemory{ID: "test-profile"}
	if profile.GetID() != "test-profile" {
		t.Error("ProfileMemory GetID not working")
	}

	// PreferenceMemory
	pref := &beadscontext.PreferenceMemory{ID: "test-pref"}
	if pref.GetID() != "test-pref" {
		t.Error("PreferenceMemory GetID not working")
	}

	// EntityMemory
	entity := &beadscontext.EntityMemory{ID: "test-entity"}
	if entity.GetID() != "test-entity" {
		t.Error("EntityMemory GetID not working")
	}

	// EventMemory
	event := &beadscontext.EventMemory{ID: "test-event"}
	if event.GetID() != "test-event" {
		t.Error("EventMemory GetID not working")
	}

	// CaseMemory
	caseMem := &beadscontext.CaseMemory{ID: "test-case"}
	if caseMem.GetID() != "test-case" {
		t.Error("CaseMemory GetID not working")
	}

	// PatternMemory
	pattern := &beadscontext.PatternMemory{ID: "test-pattern"}
	if pattern.GetID() != "test-pattern" {
		t.Error("PatternMemory GetID not working")
	}
}

// TestExtractFromText_FormatHandling 测试不同内容格式的处理
func TestExtractFromText_FormatHandling(t *testing.T) {
	e := NewExtractor()
	ctx := beadscontext.Context{}

	// Test with markdown format
	markdownText := "# 标题\n\n我是开发者。"
	collection, err := e.ExtractFromText(ctx, markdownText, "markdown")
	if err != nil {
		t.Errorf("expected no error for markdown format, got %v", err)
	}
	if collection == nil {
		t.Error("expected collection to be returned for markdown format")
	}
}

// TestExtractFromText_NestedExtraction 测试嵌套提取
func TestExtractFromText_NestedExtraction(t *testing.T) {
	e := NewExtractor()
	ctx := beadscontext.Context{}

	// Text with multiple potential matches
	text := "我是王五。我喜欢功能A。我喜欢功能B。我习惯使用工具C。"
	collection, err := e.ExtractFromText(ctx, text, "plain")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Should extract multiple preferences
	if len(collection.Preferences) < 2 {
		t.Errorf("expected at least 2 preferences, got %d", len(collection.Preferences))
	}
}
