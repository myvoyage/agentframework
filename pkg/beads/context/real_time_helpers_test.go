// Agent Framework - RealTimeContext Helper Functions Tests
// Copyright (C) 2025  Agent Framework Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package context

import (
	"testing"
)

// TestCalculateSize tests the calculateSize function
func TestCalculateSize(t *testing.T) {
	// Test string
	size := calculateSize("hello")
	if size != 5 {
		t.Errorf("expected size 5 for string, got %d", size)
	}

	// Test []byte
	size = calculateSize([]byte{1, 2, 3, 4, 5})
	if size != 5 {
		t.Errorf("expected size 5 for []byte, got %d", size)
	}

	// Test map[string]interface{}
	mapValue := map[string]interface{}{
		"key1": "value1",
		"key2": []byte{1, 2, 3},
	}
	size = calculateSize(mapValue)
	if size != 9 { // len("value1") + 3
		t.Errorf("expected size 9 for map, got %d", size)
	}

	// Test []interface{}
	sliceValue := []interface{}{"hello", []byte{1, 2, 3}}
	size = calculateSize(sliceValue)
	if size != 8 { // 5 + 3
		t.Errorf("expected size 8 for slice, got %d", size)
	}

	// Test default type (int)
	size = calculateSize(42)
	if size != 8 {
		t.Errorf("expected size 8 for default type, got %d", size)
	}
}

// TestContains tests the contains function
func TestContains(t *testing.T) {
	// Test exact match
	if !contains("hello", "hello") {
		t.Error("expected exact match to return true")
	}

	// Test substring match
	if !contains("hello world", "world") {
		t.Error("expected substring match to return true")
	}

	// Test case-insensitive match
	if !contains("Hello World", "world") {
		t.Error("expected case-insensitive match to return true")
	}

	// Test no match
	if contains("hello", "xyz") {
		t.Error("expected no match to return false")
	}

	// Test empty substring
	if !contains("hello", "") {
		t.Error("expected empty substring to match")
	}

	// Test longer substring than string
	if contains("hi", "hello") {
		t.Error("expected longer substring to not match")
	}
}

// TestContainsIgnoreCase tests the containsIgnoreCase function
func TestContainsIgnoreCase(t *testing.T) {
	// Test exact match
	if !containsIgnoreCase("hello", "hello") {
		t.Error("expected exact match")
	}

	// Test case-insensitive match
	if !containsIgnoreCase("HELLO", "hello") {
		t.Error("expected case-insensitive match")
	}

	if !containsIgnoreCase("Hello", "HELLO") {
		t.Error("expected case-insensitive match")
	}

	// Test no match
	if containsIgnoreCase("hello", "xyz") {
		t.Error("expected no match")
	}

	// Test empty substring
	if !containsIgnoreCase("hello", "") {
		t.Error("expected empty substring to match")
	}
}

// TestToLower tests the toLower function
func TestToLower(t *testing.T) {
	// Test lowercase
	result := toLower("hello")
	if result != "hello" {
		t.Errorf("expected 'hello', got '%s'", result)
	}

	// Test uppercase
	result = toLower("HELLO")
	if result != "hello" {
		t.Errorf("expected 'hello', got '%s'", result)
	}

	// Test mixed case
	result = toLower("HeLLo")
	if result != "hello" {
		t.Errorf("expected 'hello', got '%s'", result)
	}

	// Test with non-alphabetic characters
	result = toLower("Hello123!")
	if result != "hello123!" {
		t.Errorf("expected 'hello123!', got '%s'", result)
	}

	// Test empty string
	result = toLower("")
	if result != "" {
		t.Errorf("expected empty string, got '%s'", result)
	}
}

// TestExtractTerms tests the extractTerms function
func TestExtractTerms(t *testing.T) {
	// Test string
	terms := extractTerms("hello")
	if len(terms) != 1 || terms[0] != "hello" {
		t.Errorf("expected ['hello'], got %v", terms)
	}

	// Test map
	mapValue := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
	}
	terms = extractTerms(mapValue)
	// Should extract keys and values
	if len(terms) < 2 {
		t.Errorf("expected at least 2 terms from map, got %d", len(terms))
	}

	// Test slice
	sliceValue := []interface{}{"item1", "item2"}
	terms = extractTerms(sliceValue)
	if len(terms) != 2 {
		t.Errorf("expected 2 terms from slice, got %d", len(terms))
	}

	// Test nested structures
	nested := map[string]interface{}{
		"key": []interface{}{"value1", "value2"},
	}
	terms = extractTerms(nested)
	if len(terms) < 1 {
		t.Error("expected terms from nested structure")
	}
}

// TestQueryMatches tests the Query.Matches method
func TestQueryMatches(t *testing.T) {
	// Test with nil filter (should match everything)
	query := &Query{}
	if !query.Matches("anything") {
		t.Error("expected nil filter to match everything")
	}

	// Test with filter that returns true
	query = &Query{
		Filter: func(v interface{}) bool {
			return v == "target"
		},
	}
	if !query.Matches("target") {
		t.Error("expected filter to match target")
	}
	if query.Matches("other") {
		t.Error("expected filter to not match other")
	}
}

// TestCalculateRelevance tests the calculateRelevance function
func TestCalculateRelevance(t *testing.T) {
	// Test exact match
	score := calculateRelevance("hello", "hello")
	if score != 1.0 {
		t.Errorf("expected score 1.0 for exact match, got %v", score)
	}

	// Test substring match
	score = calculateRelevance("hello world", "world")
	if score != 0.5 {
		t.Errorf("expected score 0.5 for substring match, got %v", score)
	}

	// Test no match
	score = calculateRelevance("hello", "xyz")
	if score != 0.0 {
		t.Errorf("expected score 0.0 for no match, got %v", score)
	}

	// Test non-string value
	score = calculateRelevance(123, "123")
	if score != 0.0 {
		t.Errorf("expected score 0.0 for non-string value, got %v", score)
	}
}

// TestMatchesSearch tests the matchesSearch function
func TestMatchesSearch(t *testing.T) {
	// Test string match
	if !matchesSearch("hello world", "world") {
		t.Error("expected string match")
	}

	// Test string no match
	if matchesSearch("hello", "xyz") {
		t.Error("expected no match for string")
	}

	// Test map with matching value
	mapValue := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
	}
	if !matchesSearch(mapValue, "value2") {
		t.Error("expected map match")
	}

	// Test map with no matching value
	if matchesSearch(mapValue, "xyz") {
		t.Error("expected no match for map")
	}

	// Test slice with matching value
	sliceValue := []interface{}{"item1", "item2", "item3"}
	if !matchesSearch(sliceValue, "item2") {
		t.Error("expected slice match")
	}

	// Test slice with no matching value
	if matchesSearch(sliceValue, "xyz") {
		t.Error("expected no match for slice")
	}

	// Test non-matching type
	if matchesSearch(123, "123") {
		t.Error("expected no match for non-string type")
	}
}

// TestQueryStruct tests the Query struct
func TestQueryStruct(t *testing.T) {
	// Test creating query with filter and limit
	query := Query{
		Filter: func(v interface{}) bool {
			s, ok := v.(string)
			return ok && len(s) > 5
		},
		Limit: 10,
	}

	// Test that filter works
	if !query.Matches("longstring") {
		t.Error("expected filter to match long string")
	}
	if query.Matches("short") {
		t.Error("expected filter to not match short string")
	}
}

// TestQueryResultStruct tests the QueryResult struct
func TestQueryResultStruct(t *testing.T) {
	result := QueryResult{
		Key:   "test-key",
		Value: "test-value",
		Metadata: map[string]interface{}{
			"size": 10,
		},
	}

	if result.Key != "test-key" {
		t.Errorf("expected key 'test-key', got '%s'", result.Key)
	}
	if result.Value != "test-value" {
		t.Errorf("expected value 'test-value', got '%v'", result.Value)
	}
	if result.Metadata["size"] != 10 {
		t.Error("expected metadata size to be 10")
	}
}

// TestSearchResultStruct tests the SearchResult struct
func TestSearchResultStruct(t *testing.T) {
	result := SearchResult{
		Key:   "search-key",
		Value: "search-value",
		Score: 0.85,
	}

	if result.Key != "search-key" {
		t.Errorf("expected key 'search-key', got '%s'", result.Key)
	}
	if result.Value != "search-value" {
		t.Errorf("expected value 'search-value', got '%v'", result.Value)
	}
	if result.Score != 0.85 {
		t.Errorf("expected score 0.85, got %v", result.Score)
	}
}

// TestRealTimeStatsStruct tests the RealTimeStats struct
func TestRealTimeStatsStruct(t *testing.T) {
	stats := RealTimeStats{
		TotalEntries: 100,
		TotalSize:    10000,
	}

	if stats.TotalEntries != 100 {
		t.Errorf("expected TotalEntries 100, got %d", stats.TotalEntries)
	}
	if stats.TotalSize != 10000 {
		t.Errorf("expected TotalSize 10000, got %d", stats.TotalSize)
	}
}

// TestErrorVariables tests the error variables
func TestErrorVariables(t *testing.T) {
	// Test ErrKeyNotFound
	if ErrKeyNotFound == nil {
		t.Error("expected ErrKeyNotFound to be defined")
	}
	if ErrKeyNotFound.Error() != "key not found" {
		t.Errorf("expected 'key not found', got '%s'", ErrKeyNotFound.Error())
	}

	// Test ErrKeyExpired
	if ErrKeyExpired == nil {
		t.Error("expected ErrKeyExpired to be defined")
	}
	if ErrKeyExpired.Error() != "key expired" {
		t.Errorf("expected 'key expired', got '%s'", ErrKeyExpired.Error())
	}
}
