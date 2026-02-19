// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later


// Package context provides enhanced real-time context management.
package context

import (
	"context"
	"errors"
	"sync"
	"time"
)

// RealTimeContext manages real-time data with high-speed storage and retrieval.
type RealTimeContext struct {
	dataStore    map[string]interface{}
	metadata     map[string]map[string]interface{}
	index        map[string][]string // reverse index for fast lookup
	mu           sync.RWMutex
	maxSize      int
	ttl          time.Duration
	cleanupTimer *time.Timer
}

// NewRealTimeContext creates a new real-time context instance.
func NewRealTimeContext(maxSize int, ttl time.Duration) *RealTimeContext {
	rtc := &RealTimeContext{
		dataStore: make(map[string]interface{}),
		metadata:  make(map[string]map[string]interface{}),
		index:     make(map[string][]string),
		maxSize:   maxSize,
		ttl:       ttl,
	}

	// Start cleanup timer
	rtc.startCleanup()

	return rtc
}

// Set stores a value in the real-time context.
func (rtc *RealTimeContext) Set(ctx context.Context, key string, value interface{}) error {
	rtc.mu.Lock()
	defer rtc.mu.Unlock()

	// Check max size
	if len(rtc.dataStore) >= rtc.maxSize {
		// Remove oldest entry
		var oldestKey string
		var oldestTime time.Time

		for key, meta := range rtc.metadata {
			timestamp, ok := meta["timestamp"].(time.Time)
			if !ok {
				continue
			}

			if oldestKey == "" || timestamp.Before(oldestTime) {
				oldestKey = key
				oldestTime = timestamp
			}
		}

		if oldestKey != "" {
			rtc.removeLocked(oldestKey)
		}
	}

	// Store value
	rtc.dataStore[key] = value

	// Update metadata
	if rtc.metadata[key] == nil {
		rtc.metadata[key] = make(map[string]interface{})
	}
	rtc.metadata[key]["timestamp"] = time.Now()
	rtc.metadata[key]["size"] = calculateSize(value)

	// Update index
	rtc.updateIndexLocked(key, value)

	return nil
}

// Get retrieves a value from the real-time context.
func (rtc *RealTimeContext) Get(ctx context.Context, key string) (interface{}, error) {
	rtc.mu.RLock()
	defer rtc.mu.RUnlock()

	value, exists := rtc.dataStore[key]
	if !exists {
		return nil, ErrKeyNotFound
	}

	// Check TTL
	if meta, ok := rtc.metadata[key]; ok {
		if timestamp, ok := meta["timestamp"].(time.Time); ok {
			if time.Since(timestamp) > rtc.ttl {
				return nil, ErrKeyExpired
			}
		}
	}

	return value, nil
}

// Delete removes a value from the real-time context.
func (rtc *RealTimeContext) Delete(ctx context.Context, key string) error {
	rtc.mu.Lock()
	defer rtc.mu.Unlock()

	return rtc.removeLocked(key)
}

// removeLocked removes a value (must be called with lock held).
func (rtc *RealTimeContext) removeLocked(key string) error {
	if _, exists := rtc.dataStore[key]; !exists {
		return ErrKeyNotFound
	}

	delete(rtc.dataStore, key)
	delete(rtc.metadata, key)

	// Remove from index
	for indexedKey, keys := range rtc.index {
		for i, k := range keys {
			if k == key {
				rtc.index[indexedKey] = append(keys[:i], keys[i+1:]...)
				break
			}
		}
	}

	return nil
}

// Query performs a query on the real-time context.
func (rtc *RealTimeContext) Query(ctx context.Context, query Query) ([]*QueryResult, error) {
	rtc.mu.RLock()
	defer rtc.mu.RUnlock()

	results := make([]*QueryResult, 0)

	for key, value := range rtc.dataStore {
		// Check TTL
		if meta, ok := rtc.metadata[key]; ok {
			if timestamp, ok := meta["timestamp"].(time.Time); ok {
				if time.Since(timestamp) > rtc.ttl {
					continue
				}
			}
		}

		// Apply query filter
		if query.Matches(value) {
			result := &QueryResult{
				Key:   key,
				Value: value,
			}

			if meta, ok := rtc.metadata[key]; ok {
				result.Metadata = meta
			}

			results = append(results, result)
		}

		// Check limit
		if query.Limit > 0 && len(results) >= query.Limit {
			break
		}
	}

	return results, nil
}

// Search performs a full-text search on the real-time context.
func (rtc *RealTimeContext) Search(ctx context.Context, searchTerm string, limit int) ([]*SearchResult, error) {
	rtc.mu.RLock()
	defer rtc.mu.RUnlock()

	results := make([]*SearchResult, 0)

	for key, value := range rtc.dataStore {
		// Check TTL
		if meta, ok := rtc.metadata[key]; ok {
			if timestamp, ok := meta["timestamp"].(time.Time); ok {
				if time.Since(timestamp) > rtc.ttl {
					continue
				}
			}
		}

		// Simple text matching (can be enhanced with more sophisticated search)
		if matchesSearch(value, searchTerm) {
			result := &SearchResult{
				Key:   key,
				Value: value,
				Score: calculateRelevance(value, searchTerm),
			}

			results = append(results, result)

			// Check limit
			if limit > 0 && len(results) >= limit {
				break
			}
		}
	}

	return results, nil
}

// GetStats returns statistics about the real-time context.
func (rtc *RealTimeContext) GetStats(ctx context.Context) *RealTimeStats {
	rtc.mu.RLock()
	defer rtc.mu.RUnlock()

	stats := &RealTimeStats{
		TotalEntries: len(rtc.dataStore),
		TotalSize:    0,
	}

	for _, meta := range rtc.metadata {
		if size, ok := meta["size"].(int); ok {
			stats.TotalSize += size
		}
	}

	return stats
}

// Clear clears all data from the real-time context.
func (rtc *RealTimeContext) Clear(ctx context.Context) error {
	rtc.mu.Lock()
	defer rtc.mu.Unlock()

	rtc.dataStore = make(map[string]interface{})
	rtc.metadata = make(map[string]map[string]interface{})
	rtc.index = make(map[string][]string)

	return nil
}

// Close closes the real-time context and cleans up resources.
func (rtc *RealTimeContext) Close() error {
	if rtc.cleanupTimer != nil {
		rtc.cleanupTimer.Stop()
	}
	return nil
}

// startCleanup starts the background cleanup timer.
func (rtc *RealTimeContext) startCleanup() {
	rtc.cleanupTimer = time.AfterFunc(rtc.ttl, func() {
		rtc.cleanup()

		// Restart timer
		rtc.cleanupTimer.Reset(rtc.ttl)
	})
}

// cleanup removes expired entries.
func (rtc *RealTimeContext) cleanup() {
	rtc.mu.Lock()
	defer rtc.mu.Unlock()

	now := time.Now()

	for key, meta := range rtc.metadata {
		if timestamp, ok := meta["timestamp"].(time.Time); ok {
			if now.Sub(timestamp) > rtc.ttl {
				rtc.removeLocked(key)
			}
		}
	}
}

// updateIndexLocked updates the reverse index (must be called with lock held).
func (rtc *RealTimeContext) updateIndexLocked(key string, value interface{}) {
	// Extract indexable terms from the value
	terms := extractTerms(value)

	for _, term := range terms {
		rtc.index[term] = append(rtc.index[term], key)
	}
}

// Query represents a query on the real-time context.
type Query struct {
	Filter func(interface{}) bool
	Limit  int
}

// Matches checks if a value matches the query filter.
func (q *Query) Matches(value interface{}) bool {
	if q.Filter == nil {
		return true
	}
	return q.Filter(value)
}

// QueryResult represents a query result.
type QueryResult struct {
	Key      string
	Value    interface{}
	Metadata map[string]interface{}
}

// SearchResult represents a search result.
type SearchResult struct {
	Key   string
	Value interface{}
	Score float64
}

// RealTimeStats contains statistics about the real-time context.
type RealTimeStats struct {
	TotalEntries int
	TotalSize    int
}

// calculateSize estimates the size of a value.
func calculateSize(value interface{}) int {
	switch v := value.(type) {
	case string:
		return len(v)
	case []byte:
		return len(v)
	case map[string]interface{}:
		size := 0
		for _, val := range v {
			size += calculateSize(val)
		}
		return size
	case []interface{}:
		size := 0
		for _, val := range v {
			size += calculateSize(val)
		}
		return size
	default:
		return 8 // approximate size for other types
	}
}

// matchesSearch checks if a value matches the search term.
func matchesSearch(value interface{}, searchTerm string) bool {
	switch v := value.(type) {
	case string:
		return contains(v, searchTerm)
	case map[string]interface{}:
		for _, val := range v {
			if matchesSearch(val, searchTerm) {
				return true
			}
		}
		return false
	case []interface{}:
		for _, val := range v {
			if matchesSearch(val, searchTerm) {
				return true
			}
		}
		return false
	default:
		return false
	}
}

// calculateRelevance calculates the relevance score of a value.
func calculateRelevance(value interface{}, searchTerm string) float64 {
	// Simple relevance calculation (can be enhanced)
	switch v := value.(type) {
	case string:
		if v == searchTerm {
			return 1.0
		}
		if contains(v, searchTerm) {
			return 0.5
		}
		return 0.0
	default:
		return 0.0
	}
}

// contains checks if a string contains a substring (case-insensitive).
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsIgnoreCase(s, substr))
}

// containsIgnoreCase performs case-insensitive substring matching.
func containsIgnoreCase(s, substr string) bool {
	// Simple implementation (can be enhanced with more sophisticated matching)
	sLower := toLower(s)
	substrLower := toLower(substr)

	for i := 0; i <= len(sLower)-len(substrLower); i++ {
		if sLower[i:i+len(substrLower)] == substrLower {
			return true
		}
	}
	return false
}

// toLower converts a string to lowercase.
func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			result[i] = c + 32
		} else {
			result[i] = c
		}
	}
	return string(result)
}

// extractTerms extracts indexable terms from a value.
func extractTerms(value interface{}) []string {
	terms := make([]string, 0)

	switch v := value.(type) {
	case string:
		terms = append(terms, v)
	case map[string]interface{}:
		for key, val := range v {
			terms = append(terms, key)
			terms = append(terms, extractTerms(val)...)
		}
	case []interface{}:
		for _, val := range v {
			terms = append(terms, extractTerms(val)...)
		}
	}

	return terms
}

// Errors
var (
	ErrKeyNotFound = errors.New("key not found")
	ErrKeyExpired = errors.New("key expired")
)