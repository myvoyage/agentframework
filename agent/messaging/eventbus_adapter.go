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
	"fmt"
	"sync"
	"sync/atomic"
)

// SimpleEventBus 简单的内存事件总线实现
// 完全独立于 agent 包，避免循环依赖
type SimpleEventBus struct {
	mu           sync.RWMutex
	subscribers  map[string][]subscriberEntry
	nextSubID    uint64
	running      bool
	wg           sync.WaitGroup
	stopChan     chan struct{}
}

type subscriberEntry struct {
	id      string
	handler EventHandler
}

// NewSimpleEventBus 创建新的简单事件总线
func NewSimpleEventBus() *SimpleEventBus {
	return &SimpleEventBus{
		subscribers: make(map[string][]subscriberEntry),
		nextSubID:   1,
		running:     true,
		stopChan:    make(chan struct{}),
	}
}

// Start 启动事件总线
func (bus *SimpleEventBus) Start() error {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	if bus.running {
		return fmt.Errorf("event bus already started")
	}

	bus.running = true
	return nil
}

// Stop 停止事件总线
func (bus *SimpleEventBus) Stop() {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	if !bus.running {
		return
	}

	close(bus.stopChan)
	bus.running = false
}

// Publish 发布事件
func (bus *SimpleEventBus) Publish(ctx context.Context, event *Event) error {
	bus.mu.RLock()
	defer bus.mu.RUnlock()

	if !bus.running {
		return fmt.Errorf("event bus is not running")
	}

	subs, exists := bus.subscribers[event.Topic]
	if !exists {
		return nil // 没有订阅者，不是错误
	}

	// 异步调用所有订阅者
	for _, sub := range subs {
		bus.wg.Add(1)
		go func(handler EventHandler) {
			defer bus.wg.Done()
			_ = handler(ctx, event)
		}(sub.handler)
	}

	return nil
}

// Subscribe 订阅事件
func (bus *SimpleEventBus) Subscribe(topic string, handler EventHandler) string {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	subID := fmt.Sprintf("sub-%d", atomic.AddUint64(&bus.nextSubID, 1))

	entry := subscriberEntry{
		id:      subID,
		handler: handler,
	}

	bus.subscribers[topic] = append(bus.subscribers[topic], entry)

	return subID
}

// Unsubscribe 取消订阅
func (bus *SimpleEventBus) Unsubscribe(subID string) {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	for topic, subs := range bus.subscribers {
		for i, sub := range subs {
			if sub.id == subID {
				// 删除订阅者
				bus.subscribers[topic] = append(subs[:i], subs[i+1:]...)
				return
			}
		}
	}
}

