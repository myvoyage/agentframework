package token

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTokenCounter(t *testing.T) {
	t.Parallel()

	counter := NewDefaultTokenCounter()

	t.Run("Empty text", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, 0, counter.CountText(""))
	})

	t.Run("English text", func(t *testing.T) {
		t.Parallel()
		text := "Hello world! This is a test of the token counter."
		tokens := counter.CountText(text)
		assert.Greater(t, tokens, 0)
		assert.Less(t, tokens, 20) // Should be around 10-15 tokens
	})

	t.Run("Chinese text", func(t *testing.T) {
		t.Parallel()
		text := "你好，世界！这是一个 Token 计数器的测试。"
		tokens := counter.CountText(text)
		assert.Greater(t, tokens, 0)
		assert.Less(t, tokens, 15) // Should be around 8-12 tokens
	})
}

func TestTokenCompression(t *testing.T) {
	t.Parallel()

	config := DefaultCompressConfig()
	compressor := NewMessageCompressor(config, mockLLMFunc)

	t.Run("Compress empty", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		result, err := compressor.CompressMessages(ctx, []interface{}{}, 100)
		assert.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("Compress small", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		messages := []interface{}{
			map[string]interface{}{"role": "system", "content": "system prompt"},
			map[string]interface{}{"role": "user", "content": "user message"},
			map[string]interface{}{"role": "assistant", "content": "assistant response"},
		}

		result, err := compressor.CompressMessages(ctx, messages, 100)
		assert.NoError(t, err)
		assert.Len(t, result, len(messages)) // Should not compress small messages
	})

	t.Run("Compress large", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		messages := generateTestMessages()
		compressor.SetStrategy(StrategyHybrid)

		result, err := compressor.CompressMessages(ctx, messages, 500)
		assert.NoError(t, err)

		counter := NewDefaultTokenCounter()
		totalTokens := counter.CountMessages(result)
		assert.LessOrEqual(t, totalTokens, 500)
	})
}

func TestCompressionStats(t *testing.T) {
	t.Parallel()

	config := DefaultCompressConfig()
	compressor := NewMessageCompressor(config, mockLLMFunc)

	ctx := context.Background()
	messages := generateTestMessages()

	compressor.CompressMessages(ctx, messages, 500)

	stats := compressor.GetStats()
	assert.Equal(t, int64(1), stats.TotalCompressions)
	assert.Greater(t, stats.TotalInputTokens, int64(0))
	assert.Greater(t, stats.TotalOutputTokens, int64(0))
	assert.Less(t, stats.AverageRatio, 1.0) // Should compress to smaller size

	assert.NotNil(t, stats.LastCompression)
	assert.Equal(t, StrategyHybrid, stats.LastCompression.Strategy)
	assert.Greater(t, stats.LastCompression.InputTokens, 0)
	assert.Less(t, stats.LastCompression.OutputTokens, stats.LastCompression.InputTokens)
	assert.Less(t, stats.LastCompression.CompressionRatio, 1.0)
	assert.True(t, stats.LastCompression.Success)
}

func mockLLMFunc(ctx context.Context, prompt string, maxTokens int) (string, error) {
	time.Sleep(100 * time.Millisecond)
	return "[Conversation summary: user asks about Go concurrency and gets explanations]", nil
}

func generateTestMessages() []interface{} {
	messages := []interface{}{
		map[string]interface{}{
			"role":    "system",
			"content": "你是一个专业的软件开发工程师助手。请帮我解答技术问题。",
		},
	}

	for i := 0; i < 10; i++ {
		userMsg := map[string]interface{}{
			"role":    "user",
			"content": "这是第 " + fmt.Sprintf("%d", i+1) + " 个用户消息。请解释一下 Go 语言中的并发编程模型。",
		}

		assistantMsg := map[string]interface{}{
			"role":    "assistant",
			"content": "这是第 " + fmt.Sprintf("%d", i+1) + " 个助手回复。Go 语言使用 goroutine 和 channel 进行并发编程。Goroutine 是轻量级线程，由 Go 运行时管理，而不是操作系统。Channel 用于在 goroutine 之间通信和同步。",
		}

		messages = append(messages, userMsg, assistantMsg)
	}

	return messages
}
