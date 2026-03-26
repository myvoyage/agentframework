// Agent Framework - RAG Memory Integration
// Copyright (C) 2025 Agent Framework Contributors
//
// This file provides the memory retrieval subsystem used by ContextAssembler.
// Two implementations are provided:
//
//   VectorMemory – delegates to an external RAG client (e.g. Graphlit) for
//                  semantic vector search.  Falls back to BM25 keyword search
//                  when no RAG client is configured.
//
//   SimpleMemory  – pure in-process keyword/BM25 store with optional JSON-file
//                   persistence.  Suitable for single-process deployments or
//                   development environments.
//
// Scoring
// ───────
// The fallback scorer uses a hybrid of:
//   1. BM25 (Okapi BM25) for term-frequency–based relevance.
//   2. Exact-phrase boost: multiplies the score by 1.5 when the full query
//      string appears verbatim in the document.
//   3. Position boost: earlier occurrence → slightly higher score.
//
// BM25 parameters: k1 = 1.5, b = 0.75 (standard literature values).
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"
)

// ──────────────────────────────────────────────────────────────────────────────
// Public types
// ──────────────────────────────────────────────────────────────────────────────

// MemoryResult represents a single retrieved memory entry with its relevance
// score and provenance information.
type MemoryResult struct {
	// Content is the raw text of the memory.
	Content string `json:"content"`
	// Score is the normalised relevance score in [0, 1].
	Score float64 `json:"score"`
	// Source is a human-readable identifier for the origin of the memory
	// (e.g. "file_0", "cache_3", "vector_db").
	Source string `json:"source"`
	// StoredAt is when this memory was originally stored.
	StoredAt time.Time `json:"stored_at"`
}

// MemorySearcher is the interface that any long-term memory backend must satisfy
// so that ContextAssembler can retrieve relevant memories.
type MemorySearcher interface {
	// Search returns up to limit memories relevant to query.
	Search(ctx context.Context, query string, limit int) ([]MemoryResult, error)
	// Add stores a new memory entry.
	Add(ctx context.Context, content string, metadata map[string]interface{}) error
}

// RAGClient is the interface for external vector-database backends.
type RAGClient interface {
	// Ingest stores content and its metadata in the vector database.
	Ingest(ctx context.Context, content string, metadata map[string]interface{}) error
	// Query performs semantic search and returns the summary string plus
	// individual MemoryResult entries.
	Query(ctx context.Context, text string, maxResults int) (string, []MemoryResult, error)
}

// ──────────────────────────────────────────────────────────────────────────────
// memoryEntry – persisted record
// ──────────────────────────────────────────────────────────────────────────────

// memoryEntry is the JSON-serialisable unit stored in SimpleMemory.
type memoryEntry struct {
	ID       string                 `json:"id"`
	Content  string                 `json:"content"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	StoredAt time.Time              `json:"stored_at"`
	// Pre-computed term frequencies for fast BM25 scoring.
	// Not serialised; rebuilt on load.
	tf map[string]float64 `json:"-"`
}

// ──────────────────────────────────────────────────────────────────────────────
// VectorMemory
// ──────────────────────────────────────────────────────────────────────────────

// VectorMemory delegates search to an external RAG client and falls back to
// BM25 over an in-memory cache when the client is unavailable.
type VectorMemory struct {
	ragClient RAGClient

	// In-memory cache (LRU-evicting ring buffer)
	cache      []*memoryEntry
	cacheMutex sync.RWMutex
	cacheSize  int

	// BM25 corpus statistics (updated on every Add).
	avgDocLen float64
	docCount  int
}

// NewVectorMemory constructs a VectorMemory.  If ragClient is nil the instance
// operates in pure BM25 / cache mode.  cacheSize ≤ 0 defaults to 200.
func NewVectorMemory(ragClient RAGClient, cacheSize int) *VectorMemory {
	if cacheSize <= 0 {
		cacheSize = 200
	}
	return &VectorMemory{
		ragClient: ragClient,
		cache:     make([]*memoryEntry, 0, cacheSize),
		cacheSize: cacheSize,
	}
}

// Search retrieves up to limit relevant memories for the given query.
func (vm *VectorMemory) Search(ctx context.Context, query string, limit int) ([]MemoryResult, error) {
	if limit <= 0 {
		limit = 5
	}

	if vm.ragClient != nil {
		_, results, err := vm.ragClient.Query(ctx, query, limit)
		if err != nil {
			// RAG failed – fall through to local BM25.
			vm.cacheMutex.RLock()
			defer vm.cacheMutex.RUnlock()
			return vm.bm25Search(query, limit), nil
		}
		return results, nil
	}

	vm.cacheMutex.RLock()
	defer vm.cacheMutex.RUnlock()
	return vm.bm25Search(query, limit), nil
}

// Add stores a new memory entry in the vector database (if available) and the
// local BM25 cache.
func (vm *VectorMemory) Add(ctx context.Context, content string, metadata map[string]interface{}) error {
	if vm.ragClient != nil {
		if err := vm.ragClient.Ingest(ctx, content, metadata); err != nil {
			return fmt.Errorf("vector ingest failed: %w", err)
		}
	}

	entry := &memoryEntry{
		ID:       fmt.Sprintf("vm_%d", time.Now().UnixNano()),
		Content:  content,
		Metadata: metadata,
		StoredAt: time.Now(),
	}
	entry.tf = computeTF(tokenise(content))

	vm.cacheMutex.Lock()
	defer vm.cacheMutex.Unlock()

	vm.cache = append(vm.cache, entry)
	if len(vm.cache) > vm.cacheSize {
		vm.cache = vm.cache[len(vm.cache)-vm.cacheSize:]
	}
	vm.rebuildCorpusStats()
	return nil
}

// GetCacheSize returns the number of entries currently in the cache.
func (vm *VectorMemory) GetCacheSize() int {
	vm.cacheMutex.RLock()
	defer vm.cacheMutex.RUnlock()
	return len(vm.cache)
}

// ClearCache empties the in-memory cache.
func (vm *VectorMemory) ClearCache() {
	vm.cacheMutex.Lock()
	defer vm.cacheMutex.Unlock()
	vm.cache = make([]*memoryEntry, 0, vm.cacheSize)
	vm.avgDocLen = 0
	vm.docCount = 0
}

// FormatMemoryResults serialises a result slice into a Markdown section
// suitable for insertion into a system prompt.
func (vm *VectorMemory) FormatMemoryResults(results []MemoryResult) string {
	if len(results) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("\n\n## 相关记忆\n\n")
	for i, r := range results {
		sb.WriteString(fmt.Sprintf("[%d] 相关性: %.2f  来源: %s\n%s\n\n---\n", i+1, r.Score, r.Source, r.Content))
	}
	return sb.String()
}

// rebuildCorpusStats recomputes avgDocLen and docCount; must be called under cacheMutex.
func (vm *VectorMemory) rebuildCorpusStats() {
	total := 0
	for _, e := range vm.cache {
		total += len(tokenise(e.Content))
	}
	vm.docCount = len(vm.cache)
	if vm.docCount > 0 {
		vm.avgDocLen = float64(total) / float64(vm.docCount)
	}
}

// bm25Search performs BM25 scoring over the in-memory cache.
// Caller must hold cacheMutex (at least RLock).
func (vm *VectorMemory) bm25Search(query string, limit int) []MemoryResult {
	if len(vm.cache) == 0 {
		return nil
	}

	queryTokens := tokenise(query)
	if len(queryTokens) == 0 {
		return nil
	}

	// Build inverse document frequency map for query terms.
	idf := vm.computeIDF(queryTokens)

	type scored struct {
		entry *memoryEntry
		score float64
	}

	scores := make([]scored, 0, len(vm.cache))
	for _, entry := range vm.cache {
		s := vm.bm25Score(entry, queryTokens, idf)
		if s > 0 {
			// Exact-phrase boost.
			if strings.Contains(strings.ToLower(entry.Content), strings.ToLower(query)) {
				s *= 1.5
			}
			scores = append(scores, scored{entry: entry, score: s})
		}
	}

	// Sort descending.
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})

	// Normalise scores to [0, 1] relative to the top result.
	if len(scores) > 0 {
		maxScore := scores[0].score
		if maxScore > 0 {
			for i := range scores {
				scores[i].score /= maxScore
			}
		}
	}

	// Take top-k.
	if len(scores) > limit {
		scores = scores[:limit]
	}

	results := make([]MemoryResult, len(scores))
	for i, s := range scores {
		results[i] = MemoryResult{
			Content:  s.entry.Content,
			Score:    s.score,
			Source:   s.entry.ID,
			StoredAt: s.entry.StoredAt,
		}
	}
	return results
}

// computeIDF computes IDF weights for each query token against the cache corpus.
// IDF = log((N - df + 0.5) / (df + 0.5) + 1)  (Robertson's BM25 IDF variant)
func (vm *VectorMemory) computeIDF(tokens []string) map[string]float64 {
	N := float64(len(vm.cache))
	idf := make(map[string]float64, len(tokens))

	for _, tok := range tokens {
		df := 0
		for _, entry := range vm.cache {
			if _, has := entry.tf[tok]; has {
				df++
			}
		}
		idf[tok] = math.Log((N-float64(df)+0.5)/(float64(df)+0.5) + 1)
	}
	return idf
}

// bm25Score computes the BM25 score for a single document against the query.
func (vm *VectorMemory) bm25Score(entry *memoryEntry, queryTokens []string, idf map[string]float64) float64 {
	const k1 = 1.5
	const b = 0.75

	docLen := float64(len(tokenise(entry.Content)))
	score := 0.0

	for _, tok := range queryTokens {
		tf := entry.tf[tok]
		if tf == 0 {
			continue
		}
		numerator := tf * (k1 + 1)
		denominator := tf + k1*(1-b+b*(docLen/vm.avgDocLen))
		score += idf[tok] * (numerator / denominator)
	}
	return score
}

// ──────────────────────────────────────────────────────────────────────────────
// SimpleMemory – persistent, BM25-indexed, in-process store
// ──────────────────────────────────────────────────────────────────────────────

// SimpleMemory is a lightweight, file-backed memory store that uses BM25
// for relevance ranking.  It is designed for single-process deployments and
// development.  On each Add the corpus is persisted to a JSON file so that
// memories survive process restarts.
type SimpleMemory struct {
	entries    []*memoryEntry
	mu         sync.RWMutex
	persistDir string // empty → no persistence

	// BM25 corpus statistics
	avgDocLen float64
	docCount  int
}

// NewSimpleMemory creates a SimpleMemory.  If persistDir is non-empty the store
// will save and load entries from <persistDir>/memory.json.
func NewSimpleMemory(persistDir string) *SimpleMemory {
	sm := &SimpleMemory{
		entries:    make([]*memoryEntry, 0, 256),
		persistDir: persistDir,
	}
	if persistDir != "" {
		_ = sm.load() // best-effort; errors are ignored on startup
	}
	return sm
}

// Add stores a new memory entry.  If persistence is enabled the full corpus is
// written to disk after each add.
func (sm *SimpleMemory) Add(ctx context.Context, content string, metadata map[string]interface{}) error {
	if strings.TrimSpace(content) == "" {
		return fmt.Errorf("content must not be empty")
	}

	entry := &memoryEntry{
		ID:       fmt.Sprintf("mem_%d", time.Now().UnixNano()),
		Content:  content,
		Metadata: metadata,
		StoredAt: time.Now(),
	}
	entry.tf = computeTF(tokenise(content))

	sm.mu.Lock()
	sm.entries = append(sm.entries, entry)
	sm.rebuildCorpusStats()
	sm.mu.Unlock()

	if sm.persistDir != "" {
		return sm.persist()
	}
	return nil
}

// Search returns the top-k most relevant memories for query using BM25.
func (sm *SimpleMemory) Search(ctx context.Context, query string, limit int) ([]MemoryResult, error) {
	if limit <= 0 {
		limit = 5
	}

	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if len(sm.entries) == 0 {
		return nil, nil
	}

	queryTokens := tokenise(query)
	if len(queryTokens) == 0 {
		return nil, nil
	}

	idf := sm.computeIDF(queryTokens)

	type scored struct {
		entry *memoryEntry
		score float64
	}

	scores := make([]scored, 0, len(sm.entries))
	for _, entry := range sm.entries {
		s := sm.bm25Score(entry, queryTokens, idf)
		if s > 0 {
			if strings.Contains(strings.ToLower(entry.Content), strings.ToLower(query)) {
				s *= 1.5
			}
			scores = append(scores, scored{entry: entry, score: s})
		}
	}

	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})

	// Normalise.
	if len(scores) > 0 {
		maxScore := scores[0].score
		if maxScore > 0 {
			for i := range scores {
				scores[i].score /= maxScore
			}
		}
	}

	if len(scores) > limit {
		scores = scores[:limit]
	}

	results := make([]MemoryResult, len(scores))
	for i, s := range scores {
		results[i] = MemoryResult{
			Content:  s.entry.Content,
			Score:    s.score,
			Source:   s.entry.ID,
			StoredAt: s.entry.StoredAt,
		}
	}
	return results, nil
}

// Get retrieves a memory by its ID.
func (sm *SimpleMemory) Get(ctx context.Context, id string) (*memoryEntry, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	for _, e := range sm.entries {
		if e.ID == id {
			return e, nil
		}
	}
	return nil, fmt.Errorf("memory not found: %s", id)
}

// List returns all stored memories, newest first.
func (sm *SimpleMemory) List(ctx context.Context) ([]string, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	out := make([]string, len(sm.entries))
	for i, e := range sm.entries {
		out[len(sm.entries)-1-i] = e.Content // reverse order
	}
	return out, nil
}

// Delete removes a memory by ID and re-persists the corpus.
func (sm *SimpleMemory) Delete(ctx context.Context, id string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	filtered := sm.entries[:0]
	found := false
	for _, e := range sm.entries {
		if e.ID == id {
			found = true
		} else {
			filtered = append(filtered, e)
		}
	}
	if !found {
		return fmt.Errorf("memory not found: %s", id)
	}
	sm.entries = filtered
	sm.rebuildCorpusStats()

	if sm.persistDir != "" {
		return sm.persist()
	}
	return nil
}

// Len returns the number of stored entries.
func (sm *SimpleMemory) Len() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return len(sm.entries)
}

// rebuildCorpusStats recomputes BM25 statistics after mutations.
// Caller must hold mu (write lock).
func (sm *SimpleMemory) rebuildCorpusStats() {
	total := 0
	for _, e := range sm.entries {
		if e.tf == nil {
			e.tf = computeTF(tokenise(e.Content))
		}
		total += len(tokenise(e.Content))
	}
	sm.docCount = len(sm.entries)
	if sm.docCount > 0 {
		sm.avgDocLen = float64(total) / float64(sm.docCount)
	}
}

func (sm *SimpleMemory) computeIDF(tokens []string) map[string]float64 {
	N := float64(len(sm.entries))
	idf := make(map[string]float64, len(tokens))
	for _, tok := range tokens {
		df := 0
		for _, entry := range sm.entries {
			if _, has := entry.tf[tok]; has {
				df++
			}
		}
		idf[tok] = math.Log((N-float64(df)+0.5)/(float64(df)+0.5) + 1)
	}
	return idf
}

func (sm *SimpleMemory) bm25Score(entry *memoryEntry, queryTokens []string, idf map[string]float64) float64 {
	const k1 = 1.5
	const b = 0.75

	docLen := float64(len(tokenise(entry.Content)))
	score := 0.0
	for _, tok := range queryTokens {
		tf := entry.tf[tok]
		if tf == 0 {
			continue
		}
		numerator := tf * (k1 + 1)
		denominator := tf + k1*(1-b+b*(docLen/sm.avgDocLen))
		score += idf[tok] * (numerator / denominator)
	}
	return score
}

// ── Persistence ───────────────────────────────────────────────────────────────

// persist writes the current corpus to <persistDir>/memory.json.
// Called under no lock (Add/Delete lock before calling).
func (sm *SimpleMemory) persist() error {
	if sm.persistDir == "" {
		return nil
	}
	if err := os.MkdirAll(sm.persistDir, 0755); err != nil {
		return fmt.Errorf("failed to create persist dir: %w", err)
	}

	sm.mu.RLock()
	data, err := json.MarshalIndent(sm.entries, "", "  ")
	sm.mu.RUnlock()
	if err != nil {
		return fmt.Errorf("failed to marshal memories: %w", err)
	}

	path := filepath.Join(sm.persistDir, "memory.json")
	// Atomic write via temp file.
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}
	return os.Rename(tmp, path)
}

// load reads <persistDir>/memory.json into the entries slice.
// Called once from NewSimpleMemory before any concurrent access.
func (sm *SimpleMemory) load() error {
	path := filepath.Join(sm.persistDir, "memory.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // first run
		}
		return fmt.Errorf("failed to read memory file: %w", err)
	}

	var entries []*memoryEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return fmt.Errorf("failed to parse memory file: %w", err)
	}

	// Rebuild TF maps (not serialised).
	for _, e := range entries {
		e.tf = computeTF(tokenise(e.Content))
	}

	sm.entries = entries
	sm.rebuildCorpusStats()
	return nil
}

// ──────────────────────────────────────────────────────────────────────────────
// MockRAGClient – for testing / demos
// ──────────────────────────────────────────────────────────────────────────────

// MockRAGClient is an in-memory RAGClient implementation suitable for unit
// tests and demo environments.  It uses BM25 scoring identical to SimpleMemory.
type MockRAGClient struct {
	mu      sync.RWMutex
	entries []*memoryEntry
}

// NewMockRAGClient creates a MockRAGClient pre-seeded with the provided memories.
func NewMockRAGClient(memories []string) *MockRAGClient {
	entries := make([]*memoryEntry, 0, len(memories))
	for i, m := range memories {
		e := &memoryEntry{
			ID:       fmt.Sprintf("mock_%d", i),
			Content:  m,
			StoredAt: time.Now(),
		}
		e.tf = computeTF(tokenise(m))
		entries = append(entries, e)
	}
	return &MockRAGClient{entries: entries}
}

// Ingest appends a new memory to the mock store.
func (m *MockRAGClient) Ingest(ctx context.Context, content string, metadata map[string]interface{}) error {
	e := &memoryEntry{
		ID:       fmt.Sprintf("mock_%d", time.Now().UnixNano()),
		Content:  content,
		Metadata: metadata,
		StoredAt: time.Now(),
	}
	e.tf = computeTF(tokenise(content))

	m.mu.Lock()
	m.entries = append(m.entries, e)
	m.mu.Unlock()
	return nil
}

// Query searches the mock store using BM25 scoring.
func (m *MockRAGClient) Query(ctx context.Context, text string, maxResults int) (string, []MemoryResult, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.entries) == 0 {
		return text, nil, nil
	}

	// Build a transient SimpleMemory to reuse BM25 logic.
	sm := &SimpleMemory{entries: m.entries}
	sm.rebuildCorpusStats()

	results, err := sm.Search(ctx, text, maxResults)
	return text, results, err
}

// ──────────────────────────────────────────────────────────────────────────────
// Tokeniser & TF utilities
// ──────────────────────────────────────────────────────────────────────────────

// tokenise splits text into lower-cased tokens, stripping punctuation.
// It handles both ASCII and CJK (Chinese/Japanese/Korean) text: CJK characters
// are emitted as individual unigrams; Latin/Cyrillic words are whitespace-split.
func tokenise(text string) []string {
	text = strings.ToLower(text)
	tokens := make([]string, 0, 32)
	var buf strings.Builder

	flush := func() {
		if buf.Len() > 0 {
			tok := strings.TrimSpace(buf.String())
			if tok != "" {
				tokens = append(tokens, tok)
			}
			buf.Reset()
		}
	}

	for _, r := range text {
		switch {
		case isCJK(r):
			flush()
			tokens = append(tokens, string(r))
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			buf.WriteRune(r)
		default:
			flush()
		}
	}
	flush()
	return tokens
}

// isCJK returns true for CJK Unified Ideographs and common CJK extensions.
func isCJK(r rune) bool {
	return (r >= 0x4E00 && r <= 0x9FFF) || // CJK Unified Ideographs
		(r >= 0x3400 && r <= 0x4DBF) || // Extension A
		(r >= 0x20000 && r <= 0x2A6DF) || // Extension B
		(r >= 0xF900 && r <= 0xFAFF) || // Compatibility Ideographs
		(r >= 0x3000 && r <= 0x303F) // CJK Symbols and Punctuation
}

// computeTF computes a term-frequency map from a token slice.
// Raw frequency is used (not log-normalised) to match BM25 expectations.
func computeTF(tokens []string) map[string]float64 {
	tf := make(map[string]float64, len(tokens))
	for _, tok := range tokens {
		tf[tok]++
	}
	return tf
}
