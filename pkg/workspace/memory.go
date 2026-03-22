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

package workspace

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

// MemoryEntry represents a single memory entry
type MemoryEntry struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	Tags      []string  `json:"tags"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Score     float64   `json:"score,omitempty"` // Relevance score for retrieval
}

// MemoryStore manages long-term memory storage and retrieval
type MemoryStore struct {
	mu      sync.RWMutex
	root    string
	entries map[string]*MemoryEntry
}

// NewMemoryStore creates a new memory store
func NewMemoryStore(root string) (*MemoryStore, error) {
	store := &MemoryStore{
		root:    root,
		entries: make(map[string]*MemoryEntry),
	}

	// Ensure memory directory exists
	memDir := filepath.Join(root, ".memory")
	if err := os.MkdirAll(memDir, 0755); err != nil {
		return nil, err
	}

	// Load existing memories
	if err := store.Load(); err != nil {
		return nil, err
	}

	return store, nil
}

// Load loads memories from disk
func (s *MemoryStore) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	memFile := filepath.Join(s.root, FileMEMORY)
	file, err := os.Open(memFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var current *MemoryEntry

	for scanner.Scan() {
		line := scanner.Text()

		// New memory entry (starts with timestamp)
		if strings.HasPrefix(line, "## ") {
			if current != nil {
				s.entries[current.ID] = current
			}
			// Parse timestamp from header
			header := strings.TrimPrefix(line, "## ")
			parts := strings.SplitN(header, " ", 2)
			if len(parts) >= 1 {
				t, err := time.Parse("2006-01-02", parts[0])
				if err == nil {
					current = &MemoryEntry{
						ID:        generateID(),
						CreatedAt: t,
						UpdatedAt: t,
						Tags:      []string{},
					}
				}
			}
			continue
		}

		// Tags line
		if strings.HasPrefix(line, "Tags:") && current != nil {
			tags := strings.TrimPrefix(line, "Tags:")
			current.Tags = parseTags(tags)
			continue
		}

		// Content
		if current != nil {
			if current.Content != "" {
				current.Content += "\n" + line
			} else {
				current.Content = line
			}
		}
	}

	if current != nil {
		s.entries[current.ID] = current
	}

	return scanner.Err()
}

// Save saves memories to disk
func (s *MemoryStore) Save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	memFile := filepath.Join(s.root, FileMEMORY)
	file, err := os.Create(memFile)
	if err != nil {
		return err
	}
	defer file.Close()

	// Sort entries by date (newest first)
	entries := make([]*MemoryEntry, 0, len(s.entries))
	for _, e := range s.entries {
		entries = append(entries, e)
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].CreatedAt.After(entries[j].CreatedAt)
	})

	for _, e := range entries {
		file.WriteString("## ")
		file.WriteString(e.CreatedAt.Format("2006-01-02"))
		file.WriteString(" ")
		file.WriteString(e.ID)
		file.WriteString("\n\n")

		if len(e.Tags) > 0 {
			file.WriteString("Tags: ")
			file.WriteString(strings.Join(e.Tags, ", "))
			file.WriteString("\n\n")
		}

		file.WriteString(e.Content)
		file.WriteString("\n\n---\n\n")
	}

	return nil
}

// Add adds a new memory entry
func (s *MemoryStore) Add(content string, tags ...string) error {
	entry := &MemoryEntry{
		ID:        generateID(),
		Content:   content,
		Tags:      tags,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	s.mu.Lock()
	s.entries[entry.ID] = entry
	s.mu.Unlock()

	return s.Save()
}

// Search searches memories by keyword
func (s *MemoryStore) Search(query string, limit int) []*MemoryEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query = strings.ToLower(query)
	var results []*MemoryEntry

	for _, e := range s.entries {
		score := s.calculateRelevance(e, query)
		if score > 0 {
			entryCopy := *e
			entryCopy.Score = score
			results = append(results, &entryCopy)
		}
	}

	// Sort by relevance score
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	return results
}

// SemanticSearch performs semantic search (placeholder for vector embedding)
func (s *MemoryStore) SemanticSearch(query string, limit int) []*MemoryEntry {
	// Fallback to keyword search if no vector embedding is available
	return s.Search(query, limit)
}

// calculateRelevance calculates relevance score between entry and query
func (s *MemoryStore) calculateRelevance(entry *MemoryEntry, query string) float64 {
	content := strings.ToLower(entry.Content)
	tags := strings.Join(entry.Tags, " ")

	var score float64

	// Exact match in content
	if strings.Contains(content, query) {
		score += 1.0
	}

	// Tag match
	if strings.Contains(strings.ToLower(tags), query) {
		score += 0.5
	}

	// Word-level matches
	queryWords := strings.Fields(query)
	contentWords := strings.Fields(content)
	for _, qw := range queryWords {
		for _, cw := range contentWords {
			if strings.Contains(cw, qw) || strings.Contains(qw, cw) {
				score += 0.1
			}
		}
	}

	return score
}

// Get retrieves a memory entry by ID
func (s *MemoryStore) Get(id string) *MemoryEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.entries[id]
}

// Delete deletes a memory entry
func (s *MemoryStore) Delete(id string) error {
	s.mu.Lock()
	delete(s.entries, id)
	s.mu.Unlock()
	return s.Save()
}

// GetByTag retrieves memories with a specific tag
func (s *MemoryStore) GetByTag(tag string) []*MemoryEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []*MemoryEntry
	tag = strings.ToLower(tag)

	for _, e := range s.entries {
		for _, t := range e.Tags {
			if strings.ToLower(t) == tag {
				results = append(results, e)
				break
			}
		}
	}

	return results
}

// Count returns the total number of memory entries
func (s *MemoryStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.entries)
}

// Clear removes all memory entries
func (s *MemoryStore) Clear() error {
	s.mu.Lock()
	s.entries = make(map[string]*MemoryEntry)
	s.mu.Unlock()
	return s.Save()
}

// Helper functions

func parseTags(tagStr string) []string {
	tagStr = strings.TrimSpace(tagStr)
	if tagStr == "" {
		return nil
	}

	tags := strings.Split(tagStr, ",")
	for i := range tags {
		tags[i] = strings.TrimSpace(tags[i])
	}
	return tags
}

var idPattern = regexp.MustCompile(`[a-z0-9]+`)

func generateID() string {
	return idPattern.FindString(strings.ToLower(generateRandomString(8)))
}

func generateRandomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
		time.Sleep(time.Nanosecond)
	}
	return string(b)
}
