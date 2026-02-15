package agent

import (
	"context"
	"runtime"
	"sync"
	"time"

	"AgentFramework/agent/errors"
)

// ModelCacheConfig contains configuration options for the model cache
type ModelCacheConfig struct {
	MaxSize               int           `json:"max_size"`                // Maximum number of models to cache
	TTL                   time.Duration `json:"ttl"`                     // Time to live for cached models
	CleanupInterval       time.Duration `json:"cleanup_interval"`        // Interval for cleaning up expired models
	EnableDynamicWeights  bool          `json:"enable_dynamic_weights"`  // Enable dynamic weight adjustment for eviction scoring
	InitialLRUWeight      float64       `json:"initial_lru_weight"`      // Initial weight for LRU component (0.0-1.0)
	InitialLFUWeight      float64       `json:"initial_lfu_weight"`      // Initial weight for LFU component (0.0-1.0)
	InitialPriorityWeight float64       `json:"initial_priority_weight"` // Initial weight for priority component (0.0-1.0)
}

// DefaultModelCacheConfig returns a default ModelCacheConfig
func DefaultModelCacheConfig() ModelCacheConfig {
	return ModelCacheConfig{
		MaxSize:               100,
		TTL:                   1 * time.Hour,
		CleanupInterval:       10 * time.Minute,
		EnableDynamicWeights:  true,
		InitialLRUWeight:      0.4,
		InitialLFUWeight:      0.3,
		InitialPriorityWeight: 0.3,
	}
}

// ModelCacheItem represents an item stored in the model cache
type ModelCacheItem struct {
	model     ChatModel
	createdAt time.Time
	lastUsed  time.Time
	useCount  int64 // Number of times this item has been used
	priority  int   // Priority of this item (higher = better chance to stay)
}

// ModelCacheInterface defines the interface for model caching
type ModelCacheInterface interface {
	Get(key string) ChatModel
	Put(key string, model ChatModel)
	PutWithPriority(key string, model ChatModel, priority int)
	Delete(key string)
	Clear()
	Len() int
	StopCleanup()
	GetStats() CacheStats
	ResetStats()
}

// ModelCache provides caching for ChatModel instances
type ModelCache struct {
	config      ModelCacheConfig
	shards      []*shard
	numShards   int
	stopCleanup chan struct{}

	// Dynamic weights for eviction scoring
	lruWeight      float64
	lfuWeight      float64
	priorityWeight float64
	muWeights      sync.RWMutex

	// Statistics
	hits        int64
	misses      int64
	total       int64
	evictions   int64
	loadTime    time.Duration
	lastCleanup time.Time
	muStats     sync.RWMutex

	// Per-model statistics
	modelStats   map[string]*ModelCacheStats
	muModelStats sync.RWMutex
}

// shard represents a single cache shard
type shard struct {
	cache map[string]*ModelCacheItem
	mu    sync.RWMutex
}

// NewModelCache creates a new ModelCache with the given config
func NewModelCache(config ModelCacheConfig) *ModelCache {
	// Set default values if not provided
	if config.MaxSize <= 0 {
		config.MaxSize = 100
	}
	if config.TTL <= 0 {
		config.TTL = 1 * time.Hour
	}
	if config.CleanupInterval <= 0 {
		config.CleanupInterval = 10 * time.Minute
	}
	// Set default values for dynamic weights
	if !config.EnableDynamicWeights {
		// If dynamic weights are disabled, ensure we have valid initial weights
		if config.InitialLRUWeight <= 0 {
			config.InitialLRUWeight = 0.4
		}
		if config.InitialLFUWeight <= 0 {
			config.InitialLFUWeight = 0.3
		}
		if config.InitialPriorityWeight <= 0 {
			config.InitialPriorityWeight = 0.3
		}
	}

	// Determine number of shards (based on CPU cores for good parallelism)
	numShards := runtime.GOMAXPROCS(0)
	if numShards < 1 {
		numShards = 1
	}

	// Initialize shards
	shards := make([]*shard, numShards)
	for i := range shards {
		shards[i] = &shard{
			cache: make(map[string]*ModelCacheItem),
		}
	}

	cache := &ModelCache{
		config:         config,
		shards:         shards,
		numShards:      numShards,
		stopCleanup:    make(chan struct{}),
		lruWeight:      config.InitialLRUWeight,
		lfuWeight:      config.InitialLFUWeight,
		priorityWeight: config.InitialPriorityWeight,
		modelStats:     make(map[string]*ModelCacheStats),
	}

	// Start periodic cleanup
	go cache.periodicCleanup()

	return cache
}

// Get retrieves a model from the cache or returns nil if not found
func (c *ModelCache) Get(key string) ChatModel {
	shard := c.getShard(key)
	shard.mu.RLock()
	defer shard.mu.RUnlock()

	c.muStats.RLock()
	c.total++
	c.muStats.RUnlock()

	item, ok := shard.cache[key]
	if !ok {
		c.muStats.Lock()
		c.misses++
		c.muStats.Unlock()

		// Update model miss stats
		c.muModelStats.Lock()
		if stats, exists := c.modelStats[key]; exists {
			stats.Misses++
			stats.Total++
		} else {
			c.modelStats[key] = &ModelCacheStats{
				Hits:     0,
				Misses:   1,
				Total:    1,
				HitRate:  0.0,
				LoadTime: 0,
				Size:     0,
				Priority: 0,
				LastUsed: time.Time{},
			}
		}
		c.muModelStats.Unlock()

		return nil
	}

	c.muStats.Lock()
	c.hits++
	c.muStats.Unlock()

	// Update access statistics
	item.lastUsed = time.Now()
	item.useCount++

	// Update model hit stats
	c.muModelStats.Lock()
	if stats, exists := c.modelStats[key]; exists {
		stats.Hits++
		stats.Total++
		stats.LastUsed = time.Now()
	} else {
		c.modelStats[key] = &ModelCacheStats{
			Hits:     1,
			Misses:   0,
			Total:    1,
			HitRate:  1.0,
			LoadTime: 0,
			Size:     0,
			Priority: item.priority,
			LastUsed: time.Now(),
		}
	}
	c.muModelStats.Unlock()

	return item.model
}

// Put adds a model to the cache
func (c *ModelCache) Put(key string, model ChatModel) {
	c.putWithPriority(key, model, 0, time.Now())
}

// putWithPriority adds a model to the cache with a specified priority and creation time
func (c *ModelCache) putWithPriority(key string, model ChatModel, priority int, loadStartTime time.Time) {
	shard := c.getShard(key)
	shard.mu.Lock()
	defer shard.mu.Unlock()

	// Calculate load time
	loadTime := time.Since(loadStartTime)

	// If the item already exists, update it
	if existingItem, ok := shard.cache[key]; ok {
		existingItem.model = model
		existingItem.lastUsed = time.Now()
		existingItem.useCount++
		existingItem.createdAt = time.Now()
		existingItem.priority = priority

		// Update model stats
		c.muModelStats.Lock()
		if stats, ok := c.modelStats[key]; ok {
			stats.LastUsed = time.Now()
			stats.Priority = priority
		}
		c.muModelStats.Unlock()

		// Update total load time
		c.muStats.Lock()
		c.loadTime += loadTime
		c.muStats.Unlock()

		return
	}

	// Calculate shard max size (divide total max size by number of shards)
	shardMaxSize := c.config.MaxSize / c.numShards
	if shardMaxSize < 1 {
		shardMaxSize = 1
	}

	// Check if shard is full
	if len(shard.cache) >= shardMaxSize {
		// Remove the lowest scored item using smart eviction policy
		lowestScoreKey := ""
		lowestScore := 1.0 // Initialize with a high value

		now := time.Now()

		for k, v := range shard.cache {
			// Calculate eviction score (lower = better to evict)
			score := c.calculateEvictionScore(v, now)

			if lowestScoreKey == "" || score < lowestScore {
				lowestScoreKey = k
				lowestScore = score
			}
		}

		if lowestScoreKey != "" {
			delete(shard.cache, lowestScoreKey)
			// Increment eviction count
			c.muStats.Lock()
			c.evictions++
			c.muStats.Unlock()
		}
	}

	// Add new item
	shard.cache[key] = &ModelCacheItem{
		model:     model,
		createdAt: time.Now(),
		lastUsed:  time.Now(),
		useCount:  1,
		priority:  priority,
	}

	// Update model stats
	c.muModelStats.Lock()
	c.modelStats[key] = &ModelCacheStats{
		Hits:     0,
		Misses:   0,
		Total:    0,
		HitRate:  0.0,
		LoadTime: loadTime,
		Size:     0, // Estimated size, implement actual size calculation later
		Priority: priority,
		LastUsed: time.Now(),
	}
	c.muModelStats.Unlock()

	// Update total load time
	c.muStats.Lock()
	c.loadTime += loadTime
	c.muStats.Unlock()
}

// calculateEvictionScore calculates a score for cache eviction
// Lower score means better to evict
func (c *ModelCache) calculateEvictionScore(item *ModelCacheItem, now time.Time) float64 {
	// Calculate time since last use in hours
	timeSinceLastUse := now.Sub(item.lastUsed).Hours()

	// Get current weights
	c.muWeights.RLock()
	lruWeight := c.lruWeight
	lfuWeight := c.lfuWeight
	priorityWeight := c.priorityWeight
	c.muWeights.RUnlock()

	// Normalize use count (log scale to prevent very high counts from dominating)
	useCountScore := 1.0
	if item.useCount > 0 {
		useCountScore = 1.0 / (float64(item.useCount) + 1.0)
	}

	// Normalize priority (higher priority means better to keep)
	priorityScore := 1.0 - (float64(item.priority) / 100.0)

	// Calculate recency score (more recent = higher score)
	recencyScore := 1.0 / (timeSinceLastUse + 1.0)

	// Calculate total score (higher = better to keep)
	totalScore := (lruWeight * recencyScore) +
		(lfuWeight * (1.0 - useCountScore)) +
		(priorityWeight * (1.0 - priorityScore))

	// Invert the score so lower score means better to evict
	return 1.0 - totalScore
}

// adjustDynamicWeights dynamically adjusts the weights based on usage patterns
func (c *ModelCache) adjustDynamicWeights() {
	if !c.config.EnableDynamicWeights {
		return
	}

	// This is a simplified implementation of dynamic weight adjustment
	// In a real implementation, we would analyze usage patterns and adjust weights accordingly
	// For now, we'll implement a simple algorithm that adjusts weights based on hit/miss ratio

	c.muStats.RLock()
	total := c.total
	hitRate := 0.0
	if total > 0 {
		hitRate = float64(c.hits) / float64(total)
	}
	c.muStats.RUnlock()

	// Adjust weights based on hit rate
	// If hit rate is low, increase LFU weight to keep frequently used models
	// If hit rate is high, maintain current weights
	c.muWeights.Lock()
	defer c.muWeights.Unlock()

	if hitRate < 0.7 {
		// Low hit rate, increase LFU weight
		c.lfuWeight = c.lfuWeight*0.9 + 0.1*0.5
		c.lruWeight = c.lruWeight*0.9 + 0.1*0.3
	} else {
		// High hit rate, maintain balanced weights
		c.lfuWeight = c.lfuWeight*0.9 + 0.1*0.3
		c.lruWeight = c.lruWeight*0.9 + 0.1*0.4
	}

	// Ensure weights sum to approximately 1.0
	totalWeight := c.lruWeight + c.lfuWeight + c.priorityWeight
	if totalWeight > 0 {
		c.lruWeight = c.lruWeight / totalWeight
		c.lfuWeight = c.lfuWeight / totalWeight
		c.priorityWeight = c.priorityWeight / totalWeight
	}
}

// getShard returns the shard for a given key
func (c *ModelCache) getShard(key string) *shard {
	// Simple hash function to distribute keys across shards
	hash := 0
	for _, char := range key {
		hash = 31*hash + int(char)
	}
	// Ensure non-negative hash
	if hash < 0 {
		hash = -hash
	}
	return c.shards[hash%c.numShards]
}

// Delete removes a model from the cache
func (c *ModelCache) Delete(key string) {
	shard := c.getShard(key)
	shard.mu.Lock()
	defer shard.mu.Unlock()

	delete(shard.cache, key)
}

// Clear removes all models from the cache
func (c *ModelCache) Clear() {
	for _, shard := range c.shards {
		shard.mu.Lock()
		shard.cache = make(map[string]*ModelCacheItem)
		shard.mu.Unlock()
	}
}

// Len returns the number of models in the cache
func (c *ModelCache) Len() int {
	total := 0
	for _, shard := range c.shards {
		shard.mu.RLock()
		total += len(shard.cache)
		shard.mu.RUnlock()
	}
	return total
}

// StopCleanup stops the periodic cleanup goroutine
func (c *ModelCache) StopCleanup() {
	close(c.stopCleanup)
}

// CacheStats represents cache statistics
type CacheStats struct {
	Hits           int64                      `json:"hits"`            // Number of cache hits
	Misses         int64                      `json:"misses"`          // Number of cache misses
	Total          int64                      `json:"total"`           // Total number of cache requests
	HitRate        float64                    `json:"hit_rate"`        // Cache hit rate (hits / total)
	Size           int                        `json:"size"`            // Current cache size
	MaxSize        int                        `json:"max_size"`        // Maximum cache size
	TTL            time.Duration              `json:"ttl"`             // Time-to-live for cached items
	NumShards      int                        `json:"num_shards"`      // Number of cache shards
	Evictions      int64                      `json:"evictions"`       // Number of cache evictions
	LoadTime       time.Duration              `json:"load_time"`       // Total time spent loading models into cache
	LastCleanup    time.Time                  `json:"last_cleanup"`    // Last time cleanup was performed
	LRUWeight      float64                    `json:"lru_weight"`      // Current LRU weight for eviction scoring
	LFUWeight      float64                    `json:"lfu_weight"`      // Current LFU weight for eviction scoring
	PriorityWeight float64                    `json:"priority_weight"` // Current priority weight for eviction scoring
	ModelStats     map[string]ModelCacheStats `json:"model_stats"`     // Per-model cache statistics
}

// ModelCacheStats represents cache statistics for a specific model
type ModelCacheStats struct {
	Hits     int64         `json:"hits"`      // Number of hits for this model
	Misses   int64         `json:"misses"`    // Number of misses for this model
	Total    int64         `json:"total"`     // Total requests for this model
	HitRate  float64       `json:"hit_rate"`  // Hit rate for this model
	LoadTime time.Duration `json:"load_time"` // Time spent loading this model
	Size     int           `json:"size"`      // Estimated size of this model in bytes
	Priority int           `json:"priority"`  // Priority of this model
	LastUsed time.Time     `json:"last_used"` // Last time this model was used
}

// GetStats returns the current cache statistics
func (c *ModelCache) GetStats() CacheStats {
	c.muStats.RLock()
	hits := c.hits
	misses := c.misses
	total := c.total
	evictions := c.evictions
	loadTime := c.loadTime
	lastCleanup := c.lastCleanup
	c.muStats.RUnlock()

	hitRate := 0.0
	if total > 0 {
		hitRate = float64(hits) / float64(total)
	}

	// Get current weights
	c.muWeights.RLock()
	lruWeight := c.lruWeight
	lfuWeight := c.lfuWeight
	priorityWeight := c.priorityWeight
	c.muWeights.RUnlock()

	// Calculate current cache size
	size := c.Len()

	// Get per-model statistics
	c.muModelStats.RLock()
	modelStats := make(map[string]ModelCacheStats, len(c.modelStats))
	for name, stats := range c.modelStats {
		// Calculate hit rate for each model
		hitRate := 0.0
		if stats.Total > 0 {
			hitRate = float64(stats.Hits) / float64(stats.Total)
		}
		stats.HitRate = hitRate
		modelStats[name] = *stats
	}
	c.muModelStats.RUnlock()

	return CacheStats{
		Hits:           hits,
		Misses:         misses,
		Total:          total,
		HitRate:        hitRate,
		Size:           size,
		MaxSize:        c.config.MaxSize,
		TTL:            c.config.TTL,
		NumShards:      c.numShards,
		Evictions:      evictions,
		LoadTime:       loadTime,
		LastCleanup:    lastCleanup,
		LRUWeight:      lruWeight,
		LFUWeight:      lfuWeight,
		PriorityWeight: priorityWeight,
		ModelStats:     modelStats,
	}
}

// ResetStats resets the cache statistics
func (c *ModelCache) ResetStats() {
	c.muStats.Lock()
	defer c.muStats.Unlock()

	c.hits = 0
	c.misses = 0
	c.total = 0
	c.evictions = 0
	c.loadTime = 0
	c.lastCleanup = time.Time{}

	c.muModelStats.Lock()
	defer c.muModelStats.Unlock()

	c.modelStats = make(map[string]*ModelCacheStats)
}

// periodicCleanup removes expired models from the cache at regular intervals
func (c *ModelCache) periodicCleanup() {
	ticker := time.NewTicker(c.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.cleanupExpired()
			// Adjust dynamic weights based on usage patterns
			c.adjustDynamicWeights()
		case <-c.stopCleanup:
			return
		}
	}
}

// cleanupExpired removes expired models from the cache
func (c *ModelCache) cleanupExpired() {
	for _, shard := range c.shards {
		shard.mu.Lock()

		now := time.Now()
		for k, v := range shard.cache {
			// Check TTL based on creation time, not last used time
			// This ensures cache items expire even if they are frequently accessed
			if now.Sub(v.createdAt) > c.config.TTL {
				delete(shard.cache, k)
				// Increment eviction count for expired models
				c.muStats.Lock()
				c.evictions++
				c.muStats.Unlock()

				// Remove per-model stats
				c.muModelStats.Lock()
				delete(c.modelStats, k)
				c.muModelStats.Unlock()
			}
		}

		shard.mu.Unlock()
	}

	// Update last cleanup time
	c.muStats.Lock()
	c.lastCleanup = time.Now()
	c.muStats.Unlock()
}

// CachedModelFactory wraps a ModelFactory to provide caching functionality
func CachedModelFactory(factory ModelFactory, cache *ModelCache) ModelFactory {
	return func(ctx context.Context, modelName string) (ChatModel, error) {
		// Try to get from cache first
		if cache != nil {
			if model := cache.Get(modelName); model != nil {
				return model, nil
			}
		}

		// Create new model if not in cache, record load time
		loadStartTime := time.Now()
		model, err := factory(ctx, modelName)
		if err != nil {
			return nil, err
		}

		// Cache the model with load time
		if cache != nil {
			cache.putWithPriority(modelName, model, 0, loadStartTime)
		}

		return model, nil
	}
}

// PreheatCacheOption defines options for preheating the cache
type PreheatCacheOption func(*preheatOptions)

// preheatOptions contains options for preheating the cache
type preheatOptions struct {
	async    bool
	priority int
	parallel int
}

// WithAsyncPreheat enables asynchronous preheating
func WithAsyncPreheat() PreheatCacheOption {
	return func(opts *preheatOptions) {
		opts.async = true
	}
}

// WithPreheatPriority sets the priority for preheated models
func WithPreheatPriority(priority int) PreheatCacheOption {
	return func(opts *preheatOptions) {
		opts.priority = priority
	}
}

// WithPreheatParallelism sets the maximum number of parallel preheat operations
func WithPreheatParallelism(parallel int) PreheatCacheOption {
	return func(opts *preheatOptions) {
		opts.parallel = parallel
	}
}

// PreheatCache preloads models into the cache
// This improves initial cache hit rates and reduces cold start latency
func PreheatCache(ctx context.Context, factory ModelFactory, cache *ModelCache, modelNames []string, opts ...PreheatCacheOption) error {
	if cache == nil || factory == nil || len(modelNames) == 0 {
		return nil
	}

	// Set default options
	options := &preheatOptions{
		async:    false,
		priority: 50,                    // Default medium priority
		parallel: runtime.GOMAXPROCS(0), // Default to CPU cores
	}

	// Apply options
	for _, opt := range opts {
		opt(options)
	}

	// Create a semaphore to limit parallelism
	sem := make(chan struct{}, options.parallel)

	var wg sync.WaitGroup
	errCh := make(chan error, len(modelNames))

	// Preheat all models
	for _, modelName := range modelNames {
		wg.Add(1)
		if !options.async {
			// Synchronous preheat
			sem <- struct{}{}
			model, err := factory(ctx, modelName)
			<-sem
			if err != nil {
				errCh <- errors.Wrapf(err, errors.ErrCodeModelCreation, "failed to preheat model %s", modelName)
				wg.Done()
				continue
			}
			// Set specified priority for preheated models
			cache.PutWithPriority(modelName, model, options.priority)
			wg.Done()
		} else {
			// Asynchronous preheat
			go func(name string) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()

				// Create model and add to cache
				model, err := factory(ctx, name)
				if err != nil {
					errCh <- errors.Wrapf(err, errors.ErrCodeModelCreation, "failed to preheat model %s", name)
					return
				}

				// Set specified priority for preheated models
				cache.PutWithPriority(name, model, options.priority)
			}(modelName)
		}
	}

	// Wait for all goroutines to complete
	go func() {
		wg.Wait()
		close(errCh)
	}()

	// Collect any errors
	var errList []error
	for err := range errCh {
		errList = append(errList, err)
	}

	if len(errList) > 0 {
		return errors.Newf(errors.ErrCodeModelCreation, "failed to preheat %d models: %v", len(errList), errList)
	}

	return nil
}

// PutWithPriority adds a model to the cache with a specified priority
func (c *ModelCache) PutWithPriority(key string, model ChatModel, priority int) {
	c.putWithPriority(key, model, priority, time.Now())
}
