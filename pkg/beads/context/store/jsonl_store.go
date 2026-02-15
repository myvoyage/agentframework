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

package store

import (
	"bufio"
	stdcontext "context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"AgentFramework/pkg/beads/context"
)

// JSONLStore JSONL 事件日志存储实现
// 提供事件溯源能力，所有上下文变更都记录为不可变事件
type JSONLStore struct {
	filePath    string
	file        *os.File
	writer      *bufio.Writer
	mu          sync.Mutex
	started     bool
	ctx         stdcontext.Context // 使用标准库的 context
	cancel      stdcontext.CancelFunc
	flushTicker *time.Ticker
}

// JSONLStoreConfig JSONL 存储配置
type JSONLStoreConfig struct {
	FilePath  string        // JSONL 文件路径
	FlushInterval time.Duration // 刷新间隔
}

// NewJSONLStore 创建新的 JSONL 存储
func NewJSONLStore(config *JSONLStoreConfig) (*JSONLStore, error) {
	if config == nil {
		config = &JSONLStoreConfig{}
	}

	// 默认文件路径
	if config.FilePath == "" {
		config.FilePath = "./contexts.jsonl"
	}

	// 默认刷新间隔
	if config.FlushInterval == 0 {
		config.FlushInterval = 5 * time.Second
	}

	return &JSONLStore{
		filePath: config.FilePath,
	}, nil
}

// Start 启动 JSONL 存储
func (s *JSONLStore) Start(ctx stdcontext.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.started {
		return fmt.Errorf("store already started")
	}

	// 创建上下文
	s.ctx, s.cancel = stdcontext.WithCancel(ctx)

	// 确保目录存在
	dir := filepath.Dir(s.filePath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("create directory: %w", err)
		}
	}

	// 打开文件（追加模式）
	file, err := os.OpenFile(s.filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}

	s.file = file
	s.writer = bufio.NewWriter(file)
	s.started = true

	// 启动定时刷新
	s.flushTicker = time.NewTicker(5 * time.Second)
	go s.flushLoop()

	return nil
}

// Stop 停止 JSONL 存储
func (s *JSONLStore) Stop(ctx stdcontext.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.stop()
}

// stop 内部停止方法（不加锁）
func (s *JSONLStore) stop() error {
	if !s.started {
		return nil
	}

	if s.flushTicker != nil {
		s.flushTicker.Stop()
	}

	if s.cancel != nil {
		s.cancel()
	}

	// 刷新缓冲区
	if s.writer != nil {
		s.writer.Flush()
	}

	if s.file != nil {
		if err := s.file.Close(); err != nil {
			return fmt.Errorf("close file: %w", err)
		}
	}

	s.started = false
	return nil
}

// flushLoop 定时刷新循环
func (s *JSONLStore) flushLoop() {
	for {
		select {
		case <-s.flushTicker.C:
			s.mu.Lock()
			if s.writer != nil {
				s.writer.Flush()
			}
			s.mu.Unlock()
		case <-s.ctx.Done(): // 这里使用 stdcontext 的 Done() 方法
			return
		}
	}
}

// ForceFlush 强制刷新（用于测试）
func (s *JSONLStore) ForceFlush() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.writer != nil {
		return s.writer.Flush()
	}
	return nil
}

// ===== 事件操作 =====

// AppendEvent 添加事件
func (s *JSONLStore) AppendEvent(ctx stdcontext.Context, event *context.ContextEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.started {
		return fmt.Errorf("store not started")
	}

	// 设置时间戳（如果还没有）
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// 编码为 JSON
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	// 写入文件
	if _, err := s.writer.Write(data); err != nil {
		return fmt.Errorf("write event: %w", err)
	}

	// 写入换行符
	if _, err := s.writer.Write([]byte("\n")); err != nil {
		return fmt.Errorf("write newline: %w", err)
	}

	return nil
}

// ReadEvents 读取事件
func (s *JSONLStore) ReadEvents(ctx stdcontext.Context, since time.Time) ([]*context.ContextEvent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.started {
		return nil, fmt.Errorf("store not started")
	}

	// 打开文件读取
	file, err := os.Open(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []*context.ContextEvent{}, nil
		}
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var events []*context.ContextEvent

	for scanner.Scan() {
		line := scanner.Bytes()

		var event context.ContextEvent
		if err := json.Unmarshal(line, &event); err != nil {
			// 跳过无法解析的行
			continue
		}

		// 过滤时间
		if !since.IsZero() && event.Timestamp.Before(since) {
			continue
		}

		events = append(events, &event)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan file: %w", err)
	}

	return events, nil
}

// ReadAllEvents 读取所有事件
func (s *JSONLStore) ReadAllEvents(ctx stdcontext.Context) ([]*context.ContextEvent, error) {
	return s.ReadEvents(ctx, time.Time{})
}

// GetLatestTimestamp 获取最新事件的时间戳
func (s *JSONLStore) GetLatestTimestamp(ctx stdcontext.Context) (time.Time, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.started {
		return time.Time{}, fmt.Errorf("store not started")
	}

	file, err := os.Open(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return time.Time{}, nil
		}
		return time.Time{}, fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	// 从文件末尾开始读取
	var latestTimestamp time.Time
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Bytes()

		var event context.ContextEvent
		if err := json.Unmarshal(line, &event); err != nil {
			continue
		}

		if event.Timestamp.After(latestTimestamp) {
			latestTimestamp = event.Timestamp
		}
	}

	return latestTimestamp, scanner.Err()
}

// ReplayEvents 回放事件以重建状态
func (s *JSONLStore) ReplayEvents(ctx stdcontext.Context, contextID string, fromTime time.Time) ([]*context.ContextEvent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.started {
		return nil, fmt.Errorf("store not started")
	}

	file, err := os.Open(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []*context.ContextEvent{}, nil
		}
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var events []*context.ContextEvent

	for scanner.Scan() {
		line := scanner.Bytes()

		var event context.ContextEvent
		if err := json.Unmarshal(line, &event); err != nil {
			continue
		}

		// 过滤上下文 ID
		if contextID != "" && event.ContextID != contextID {
			continue
		}

		// 过滤时间
		if !fromTime.IsZero() && event.Timestamp.Before(fromTime) {
			continue
		}

		events = append(events, &event)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan file: %w", err)
	}

	return events, nil
}

// GetLatest 获取最新事件
func (s *JSONLStore) GetLatest(ctx stdcontext.Context, contextID string, limit int) ([]*context.ContextEvent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.started {
		return nil, fmt.Errorf("store not started")
	}

	file, err := os.Open(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []*context.ContextEvent{}, nil
		}
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var allEvents []*context.ContextEvent

	for scanner.Scan() {
		line := scanner.Bytes()

		var event context.ContextEvent
		if err := json.Unmarshal(line, &event); err != nil {
			continue
		}

		// 过滤上下文 ID
		if contextID != "" && event.ContextID != contextID {
			continue
		}

		allEvents = append(allEvents, &event)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan file: %w", err)
	}

	// 返回最后的 N 个事件
	if limit > 0 && len(allEvents) > limit {
		start := len(allEvents) - limit
		return allEvents[start:], nil
	}

	return allEvents, nil
}

// Compact 压缩事件
func (s *JSONLStore) Compact(ctx stdcontext.Context, beforeTime time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.started {
		return fmt.Errorf("store not started")
	}

	// 读取所有事件
	allEvents, err := s.ReadAllEvents(ctx)
	if err != nil {
		return fmt.Errorf("read events: %w", err)
	}

	// 分离需要保留的和需要压缩的
	var keepEvents []*context.ContextEvent
	var compactEvents []*context.ContextEvent

	for _, event := range allEvents {
		if event.Timestamp.Before(beforeTime) {
			compactEvents = append(compactEvents, event)
		} else {
			keepEvents = append(keepEvents, event)
		}
	}

	// 创建压缩文件
	compactPath := s.filePath + ".compact"
	compactFile, err := os.Create(compactPath)
	if err != nil {
		return fmt.Errorf("create compact file: %w", err)
	}
	defer compactFile.Close()

	// 写入压缩事件摘要
	summary := map[string]interface{}{
		"compacted_at": time.Now(),
		"before_time":  beforeTime,
		"event_count":  len(compactEvents),
		"events":       compactEvents,
	}

	summaryData, _ := json.Marshal(summary)
	compactFile.Write(summaryData)
	compactFile.Write([]byte("\n"))

	// 写入保留的事件
	for _, event := range keepEvents {
		data, _ := json.Marshal(event)
		compactFile.Write(data)
		compactFile.Write([]byte("\n"))
	}

	// 关闭当前文件
	s.writer.Flush()
	s.file.Close()

	// 备份原文件
	backupPath := s.filePath + ".backup"
	if err := os.Rename(s.filePath, backupPath); err != nil {
		// 尝试恢复
		s.file, _ = os.OpenFile(s.filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		s.writer = bufio.NewWriter(s.file)
		return fmt.Errorf("backup file: %w", err)
	}

	// 使用压缩文件
	if err := os.Rename(compactPath, s.filePath); err != nil {
		// 尝试恢复备份
		os.Rename(backupPath, s.filePath)
		s.file, _ = os.OpenFile(s.filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		s.writer = bufio.NewWriter(s.file)
		return fmt.Errorf("replace with compact file: %w", err)
	}

	// 删除备份
	os.Remove(backupPath)

	// 重新打开文件
	file, err := os.OpenFile(s.filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("reopen file: %w", err)
	}

	s.file = file
	s.writer = bufio.NewWriter(file)

	return nil
}

// Close 关闭存储
func (s *JSONLStore) Close() error {
	return s.Stop(stdcontext.Background())
}
