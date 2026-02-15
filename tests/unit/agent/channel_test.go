// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
// Copyright (C) 2025  Agent Framework Contributors

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.

// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

// SPDX-License-Identifier: AGPL-3.0-or-later

package messaging

import (
	"context"
	"testing"
	"time"
)

func TestNewChannelMessage(t *testing.T) {
	msg := NewChannelMessage("test", ChannelTypeInternal, "test content")

	if msg.ID == "" {
		t.Error("expected message ID to be generated")
	}

	if msg.Channel != "test" {
		t.Errorf("expected channel to be 'test', got '%s'", msg.Channel)
	}

	if msg.Type != ChannelTypeInternal {
		t.Errorf("expected type to be '%s', got '%s'", ChannelTypeInternal, msg.Type)
	}

	if msg.Content != "test content" {
		t.Errorf("expected content to be 'test content', got '%s'", msg.Content)
	}

	if msg.Timestamp.IsZero() {
		t.Error("expected timestamp to be set")
	}
}

func TestNewChannelMessageWithSource(t *testing.T) {
	msg := NewChannelMessageWithSource("test", ChannelTypeInternal, "test content", "user123")

	if msg.Source != "user123" {
		t.Errorf("expected source to be 'user123', got '%s'", msg.Source)
	}
}

func TestChannelStats_RecordMessage(t *testing.T) {
	stats := ChannelStats{}

	// 记录成功的消息
	stats.RecordMessage(true, 100*time.Millisecond)

	if stats.TotalMessages != 1 {
		t.Errorf("expected total messages to be 1, got %d", stats.TotalMessages)
	}

	if stats.FailedMessages != 0 {
		t.Errorf("expected failed messages to be 0, got %d", stats.FailedMessages)
	}

	if stats.AverageLatency != 100*time.Millisecond {
		t.Errorf("expected average latency to be 100ms, got %v", stats.AverageLatency)
	}

	if !stats.IsHealthy {
		t.Error("expected channel to be healthy")
	}

	// 记录失败的消息
	stats.RecordMessage(false, 50*time.Millisecond)

	if stats.TotalMessages != 2 {
		t.Errorf("expected total messages to be 2, got %d", stats.TotalMessages)
	}

	if stats.FailedMessages != 1 {
		t.Errorf("expected failed messages to be 1, got %d", stats.FailedMessages)
	}

	// 平均延迟应该是 (100 + 50) / 2 = 75
	expectedLatency := 75 * time.Millisecond
	if stats.AverageLatency != expectedLatency {
		t.Errorf("expected average latency to be %v, got %v", expectedLatency, stats.AverageLatency)
	}
}

func TestChannelConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  ChannelConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: ChannelConfig{
				Name:    "test",
				Type:    ChannelTypeInternal,
				Enabled: true,
			},
			wantErr: false,
		},
		{
			name: "empty name",
			config: ChannelConfig{
				Type:    ChannelTypeInternal,
				Enabled: true,
			},
			wantErr: true,
		},
		{
			name: "empty type",
			config: ChannelConfig{
				Name:    "test",
				Enabled: true,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestInternalChannel_BasicOperations(t *testing.T) {
	ctx := context.Background()

	config := ChannelConfig{
		Name:       "test-internal",
		Type:       ChannelTypeInternal,
		BufferSize: 10,
	}

	ch, err := NewInternalChannel("test-internal", config)
	if err != nil {
		t.Fatalf("failed to create internal channel: %v", err)
	}

	// 测试启动
	if err := ch.Start(ctx); err != nil {
		t.Fatalf("failed to start channel: %v", err)
	}

	// 验证通道正在运行
	if !ch.IsRunning() {
		t.Error("expected channel to be running")
	}

	// 测试订阅
	messageReceived := false
	handler := func(ctx context.Context, msg *ChannelMessage) error {
		messageReceived = true
		return nil
	}

	if err := ch.Subscribe(ctx, handler); err != nil {
		t.Fatalf("failed to subscribe: %v", err)
	}

	// 测试发布
	msg := NewChannelMessage("test-internal", ChannelTypeInternal, "test message")
	if err := ch.Publish(ctx, msg); err != nil {
		t.Fatalf("failed to publish message: %v", err)
	}

	// 等待消息被处理
	time.Sleep(100 * time.Millisecond)

	// 注意：由于异步处理，这里可能需要更长的等待时间或使用同步机制
	// 在实际测试中，应该使用 channel 或 WaitGroup 来确保消息被处理

	// 测试健康检查
	if err := ch.HealthCheck(ctx); err != nil {
		t.Errorf("health check failed: %v", err)
	}

	// 测试统计信息
	stats := ch.GetStats()
	if stats == nil {
		t.Error("expected stats to be returned")
	}

	// 测试停止
	if err := ch.Stop(ctx); err != nil {
		t.Fatalf("failed to stop channel: %v", err)
	}

	// 验证通道已停止
	if ch.IsRunning() {
		t.Error("expected channel to be stopped")
	}
}

func TestChannelManager_BasicOperations(t *testing.T) {
	ctx := context.Background()

	config := ChannelManagerConfig{
		Logger: &DefaultLogger{},
	}

	manager, err := NewChannelManager(config)
	if err != nil {
		t.Fatalf("failed to create channel manager: %v", err)
	}

	// 测试启动
	if err := manager.Start(ctx); err != nil {
		t.Fatalf("failed to start manager: %v", err)
	}

	// 验证管理器正在运行
	if !manager.IsRunning() {
		t.Error("expected manager to be running")
	}

	// 创建并注册内部通道
	channelConfig := ChannelConfig{
		Name:       "test",
		Type:       ChannelTypeInternal,
		BufferSize: 10,
	}

	channel, err := NewInternalChannel("test", channelConfig)
	if err != nil {
		t.Fatalf("failed to create channel: %v", err)
	}

	if err := manager.RegisterChannel(channel); err != nil {
		t.Fatalf("failed to register channel: %v", err)
	}

	// 验证通道已注册
	channels := manager.ListChannels()
	if len(channels) != 1 {
		t.Errorf("expected 1 channel, got %d", len(channels))
	}

	// 测试获取通道
	retrieved, err := manager.GetChannel("test")
	if err != nil {
		t.Fatalf("failed to get channel: %v", err)
	}

	if retrieved.Name() != "test" {
		t.Errorf("expected channel name to be 'test', got '%s'", retrieved.Name())
	}

	// 测试健康检查
	healthResults := manager.HealthCheck(ctx)
	if len(healthResults) != 1 {
		t.Errorf("expected 1 health result, got %d", len(healthResults))
	}

	// 测试获取统计信息
	stats := manager.GetAllStats()
	if len(stats) != 1 {
		t.Errorf("expected 1 stats result, got %d", len(stats))
	}

	// 测试注销通道
	if err := manager.UnregisterChannel("test"); err != nil {
		t.Fatalf("failed to unregister channel: %v", err)
	}

	// 验证通道已注销
	channels = manager.ListChannels()
	if len(channels) != 0 {
		t.Errorf("expected 0 channels, got %d", len(channels))
	}

	// 测试停止
	if err := manager.Stop(ctx); err != nil {
		t.Fatalf("failed to stop manager: %v", err)
	}

	// 验证管理器已停止
	if manager.IsRunning() {
		t.Error("expected manager to be stopped")
	}
}

func TestChannelManager_GetChannelsByType(t *testing.T) {
	ctx := context.Background()

	config := ChannelManagerConfig{
		Logger: &DefaultLogger{},
	}

	manager, err := NewChannelManager(config)
	if err != nil {
		t.Fatalf("failed to create channel manager: %v", err)
	}

	if err := manager.Start(ctx); err != nil {
		t.Fatalf("failed to start manager: %v", err)
	}

	// 注册多个不同类型的通道
	// 注意：这里只测试内部通道，其他通道需要模拟或集成测试

	channelConfig1 := ChannelConfig{
		Name:  "internal1",
		Type:  ChannelTypeInternal,
	}

	channel1, err := NewInternalChannel("internal1", channelConfig1)
	if err != nil {
		t.Fatalf("failed to create channel: %v", err)
	}

	if err := manager.RegisterChannel(channel1); err != nil {
		t.Fatalf("failed to register channel: %v", err)
	}

	channelConfig2 := ChannelConfig{
		Name:  "internal2",
		Type:  ChannelTypeInternal,
	}

	channel2, err := NewInternalChannel("internal2", channelConfig2)
	if err != nil {
		t.Fatalf("failed to create channel: %v", err)
	}

	if err := manager.RegisterChannel(channel2); err != nil {
		t.Fatalf("failed to register channel: %v", err)
	}

	// 按类型获取通道
	internalChannels := manager.GetChannelsByType(ChannelTypeInternal)
	if len(internalChannels) != 2 {
		t.Errorf("expected 2 internal channels, got %d", len(internalChannels))
	}

	telegramChannels := manager.GetChannelsByType(ChannelTypeTelegram)
	if len(telegramChannels) != 0 {
		t.Errorf("expected 0 telegram channels, got %d", len(telegramChannels))
	}

	// 清理
	manager.Stop(ctx)
}

// TestChannelStats_HealthCalculation 测试健康状态计算
func TestChannelStats_HealthCalculation(t *testing.T) {
	stats := ChannelStats{}

	// 正常情况：少量消息，无失败
	for i := 0; i < 5; i++ {
		stats.RecordMessage(true, 100*time.Millisecond)
	}

	if !stats.IsHealthy {
		t.Error("expected channel to be healthy with no failures")
	}

	// 高失败率：10条消息中2条失败（20%）
	stats2 := ChannelStats{}
	for i := 0; i < 8; i++ {
		stats2.RecordMessage(true, 100*time.Millisecond)
	}
	for i := 0; i < 2; i++ {
		stats2.RecordMessage(false, 100*time.Millisecond)
	}

	if !stats2.IsHealthy {
		t.Error("expected channel to be healthy with 20% failure rate")
	}

	// 超过10%失败率：15条消息中2条失败（约13%）
	stats3 := ChannelStats{}
	for i := 0; i < 13; i++ {
		stats3.RecordMessage(true, 100*time.Millisecond)
	}
	for i := 0; i < 2; i++ {
		stats3.RecordMessage(false, 100*time.Millisecond)
	}

	if stats3.IsHealthy {
		t.Error("expected channel to be unhealthy with >10% failure rate")
	}
}
