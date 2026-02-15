package memory_test

import (
	"testing"
	"time"

	"AgentFramework/pkg/beads/context"
)

// TestMemoryTierTypes 测试记忆分层类型
func TestMemoryTierTypes(t *testing.T) {
	tiers := []context.MemoryTier{
		context.MemoryTierSession,
		context.MemoryTierDaily,
		context.MemoryTierLongTerm,
	}

	for _, tier := range tiers {
		if tier == "" {
			t.Errorf("MemoryTier should not be empty")
		}
	}
}

// TestMemoryCompressionConfig 测试记忆压缩配置
func TestMemoryCompressionConfig(t *testing.T) {
	config := context.DefaultMemoryCompressionConfig()

	if config == nil {
		t.Fatalf("DefaultMemoryCompressionConfig should not return nil")
	}

	if config.SessionRetentionDuration != 24*time.Hour {
		t.Errorf("Expected SessionRetentionDuration to be 24h, got %v", config.SessionRetentionDuration)
	}

	if config.MaxSessionMemories != 100 {
		t.Errorf("Expected MaxSessionMemories to be 100, got %d", config.MaxSessionMemories)
	}
}

// TestEnhancedMemoryCollection 测试增强记忆集合
func TestEnhancedMemoryCollection(t *testing.T) {
	collection := context.NewEnhancedMemoryCollection()

	if collection == nil {
		t.Fatalf("NewEnhancedMemoryCollection should not return nil")
	}

	if collection.IsEmpty() == false {
		t.Errorf("New collection should be empty")
	}

	if collection.GetMemoryCount() != 0 {
		t.Errorf("New collection should have 0 memories")
	}

	// 测试设置层级
	collection.SetTier("test-id", context.MemoryTierSession, time.Now().Add(24*time.Hour), 0.5)

	tier := collection.GetTier("test-id")
	if tier != context.MemoryTierSession {
		t.Errorf("Expected tier to be Session, got %v", tier)
	}
}
