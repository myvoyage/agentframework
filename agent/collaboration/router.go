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

package collaboration

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"
)

// RoutingStrategy defines the routing strategy
type RoutingStrategy int

const (
	StrategyRoundRobin RoutingStrategy = iota
	StrategyLeastLoaded
	StrategyFastestResponse
	StrategyCostOptimized
	StrategyIntelligent
	StrategyCapabilityMatch
	StrategyPriorityBased
)

func (rs RoutingStrategy) String() string {
	switch rs {
	case StrategyRoundRobin:
		return "round_robin"
	case StrategyLeastLoaded:
		return "least_loaded"
	case StrategyFastestResponse:
		return "fastest_response"
	case StrategyCostOptimized:
		return "cost_optimized"
	case StrategyIntelligent:
		return "intelligent"
	case StrategyCapabilityMatch:
		return "capability_match"
	case StrategyPriorityBased:
		return "priority_based"
	default:
		return "unknown"
	}
}

// RouterConfig configures the intelligent router
type RouterConfig struct {
	DefaultStrategy   RoutingStrategy
	EnableCaching     bool
	CacheTTL          time.Duration
	EnablePredictions bool
	ScoringWeights    ScoringWeights
	MaxRetries        int
	SelectionTimeout  time.Duration
}

// ScoringWeights defines weights for different scoring factors
type ScoringWeights struct {
	Latency  float64 // 0.0-1.0
	Success  float64 // 0.0-1.0
	Load     float64 // 0.0-1.0
	Cost     float64 // 0.0-1.0
	Quality  float64 // 0.0-1.0
	Priority float64 // 0.0-1.0
}

// DefaultScoringWeights returns default scoring weights
func DefaultScoringWeights() ScoringWeights {
	return ScoringWeights{
		Latency:  0.30,
		Success:  0.30,
		Load:     0.20,
		Cost:     0.10,
		Quality:  0.10,
		Priority: 0.00,
	}
}

// IntelligentRouter implements intelligent routing for agent selection
type IntelligentRouter struct {
	config        RouterConfig
	strategies    map[RoutingStrategy]RoutingStrategyFunc
	roundRobinIdx int
	mu            sync.RWMutex
	cache         *RouterCache
	metrics       *RouterMetrics
}

// RoutingStrategyFunc defines the function signature for routing strategies
type RoutingStrategyFunc func(ctx context.Context, members []*TeamMember, task *CollaborativeTask) (*TeamMember, error)

// RouterCache caches routing decisions
type RouterCache struct {
	entries map[string]*CacheEntry
	mu      sync.RWMutex
	ttl     time.Duration
}

// CacheEntry represents a cached routing decision
type CacheEntry struct {
	Member   *TeamMember
	ExpireAt time.Time
}

// RouterMetrics tracks routing metrics
type RouterMetrics struct {
	TotalSelections      int64
	StrategyUsage        map[RoutingStrategy]int64
	AverageSelectionTime time.Duration
	CacheHits            int64
	CacheMisses          int64
	SelectionHistory     []SelectionRecord
	mu                   sync.RWMutex
}

// SelectionRecord records a routing selection
type SelectionRecord struct {
	Timestamp  time.Time
	TaskID     string
	TaskType   string
	MemberName string
	Strategy   RoutingStrategy
	Duration   time.Duration
	Success    bool
}

// NewIntelligentRouter creates a new intelligent router
func NewIntelligentRouter(config RouterConfig) *IntelligentRouter {
	if config.ScoringWeights.Latency == 0 &&
		config.ScoringWeights.Success == 0 &&
		config.ScoringWeights.Load == 0 {
		config.ScoringWeights = DefaultScoringWeights()
	}

	router := &IntelligentRouter{
		config:        config,
		strategies:    make(map[RoutingStrategy]RoutingStrategyFunc),
		roundRobinIdx: 0,
		metrics:       NewRouterMetrics(),
	}

	// Register built-in strategies
	router.registerStrategy(StrategyRoundRobin, router.roundRobinStrategy)
	router.registerStrategy(StrategyLeastLoaded, router.leastLoadedStrategy)
	router.registerStrategy(StrategyFastestResponse, router.fastestResponseStrategy)
	router.registerStrategy(StrategyCostOptimized, router.costOptimizedStrategy)
	router.registerStrategy(StrategyIntelligent, router.intelligentStrategy)
	router.registerStrategy(StrategyCapabilityMatch, router.capabilityMatchStrategy)
	router.registerStrategy(StrategyPriorityBased, router.priorityBasedStrategy)

	// Initialize cache if enabled
	if config.EnableCaching {
		router.cache = &RouterCache{
			entries: make(map[string]*CacheEntry),
			ttl:     config.CacheTTL,
		}
	}

	return router
}

// registerStrategy registers a routing strategy
func (ir *IntelligentRouter) registerStrategy(strategy RoutingStrategy, fn RoutingStrategyFunc) {
	ir.mu.Lock()
	defer ir.mu.Unlock()
	ir.strategies[strategy] = fn
}

// SelectMember selects the best member for a task using the default strategy
func (ir *IntelligentRouter) SelectMember(ctx context.Context, members []*TeamMember, task *CollaborativeTask) (*TeamMember, error) {
	return ir.SelectMemberWithStrategy(ctx, members, task, ir.config.DefaultStrategy)
}

// SelectMemberWithStrategy selects a member using a specific strategy
func (ir *IntelligentRouter) SelectMemberWithStrategy(ctx context.Context, members []*TeamMember, task *CollaborativeTask, strategy RoutingStrategy) (*TeamMember, error) {
	startTime := time.Now()

	// Filter capable members
	capableMembers := ir.filterCapableMembers(members, task)
	if len(capableMembers) == 0 {
		return nil, fmt.Errorf("no capable members found for task %s", task.ID)
	}

	// Check cache
	if ir.cache != nil {
		cacheKey := ir.buildCacheKey(task, strategy)
		if cached := ir.cache.Get(cacheKey); cached != nil {
			ir.metrics.RecordCacheHit()
			ir.metrics.RecordSelection(task.ID, task.Type, cached.Agent.Name(), strategy, time.Since(startTime), true)
			return cached, nil
		}
		ir.metrics.RecordCacheMiss()
	}

	// Get strategy function
	ir.mu.RLock()
	strategyFunc, ok := ir.strategies[strategy]
	ir.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("unknown routing strategy: %s", strategy)
	}

	// Execute strategy
	member, err := strategyFunc(ctx, capableMembers, task)
	if err != nil {
		return nil, err
	}

	// Update metrics
	duration := time.Since(startTime)
	ir.metrics.RecordSelection(task.ID, task.Type, member.Agent.Name(), strategy, duration, true)
	ir.metrics.RecordStrategyUsage(strategy)

	// Cache result
	if ir.cache != nil {
		cacheKey := ir.buildCacheKey(task, strategy)
		ir.cache.Set(cacheKey, member)
	}

	return member, nil
}

// filterCapableMembers filters members that have the required capabilities
func (ir *IntelligentRouter) filterCapableMembers(members []*TeamMember, task *CollaborativeTask) []*TeamMember {
	if len(task.RequiredCapabilities) == 0 {
		return members
	}

	var capable []*TeamMember
	for _, member := range members {
		if ir.hasRequiredCapabilities(member, task.RequiredCapabilities) {
			capable = append(capable, member)
		}
	}

	return capable
}

// hasRequiredCapabilities checks if a member has the required capabilities
func (ir *IntelligentRouter) hasRequiredCapabilities(member *TeamMember, requiredCaps []string) bool {
	capMap := make(map[string]bool)
	for _, cap := range member.Capabilities {
		capMap[cap] = true
	}

	for _, reqCap := range requiredCaps {
		if !capMap[reqCap] {
			return false
		}
	}

	return true
}

// buildCacheKey builds a cache key for a task
func (ir *IntelligentRouter) buildCacheKey(task *CollaborativeTask, strategy RoutingStrategy) string {
	return fmt.Sprintf("%s:%s:%v", task.Type, strategy, task.RequiredCapabilities)
}

// ===== Routing Strategies =====

// roundRobinStrategy implements round-robin routing
func (ir *IntelligentRouter) roundRobinStrategy(ctx context.Context, members []*TeamMember, task *CollaborativeTask) (*TeamMember, error) {
	ir.mu.Lock()
	defer ir.mu.Unlock()

	idx := ir.roundRobinIdx % len(members)
	ir.roundRobinIdx++

	return members[idx], nil
}

// leastLoadedStrategy implements least-loaded routing
func (ir *IntelligentRouter) leastLoadedStrategy(ctx context.Context, members []*TeamMember, task *CollaborativeTask) (*TeamMember, error) {
	var selected *TeamMember
	minLoad := int(math.MaxInt32)

	for _, member := range members {
		member.mu.Lock()
		load := member.ActiveTasks
		member.mu.Unlock()

		if load < minLoad {
			minLoad = load
			selected = member
		}
	}

	if selected == nil {
		return nil, fmt.Errorf("no members available")
	}

	return selected, nil
}

// fastestResponseStrategy implements fastest-response routing
func (ir *IntelligentRouter) fastestResponseStrategy(ctx context.Context, members []*TeamMember, task *CollaborativeTask) (*TeamMember, error) {
	var selected *TeamMember
	bestDuration := time.Duration(math.MaxInt64)

	for _, member := range members {
		member.mu.Lock()
		duration := member.Performance.AvgDuration
		member.mu.Unlock()

		if duration == 0 {
			// No history, give it a chance
			duration = 1 * time.Second
		}

		if duration < bestDuration {
			bestDuration = duration
			selected = member
		}
	}

	if selected == nil {
		return nil, fmt.Errorf("no members available")
	}

	return selected, nil
}

// costOptimizedStrategy implements cost-optimized routing
func (ir *IntelligentRouter) costOptimizedStrategy(ctx context.Context, members []*TeamMember, task *CollaborativeTask) (*TeamMember, error) {
	var selected *TeamMember
	lowestCost := math.MaxFloat64

	for _, member := range members {
		member.mu.Lock()
		cost := member.Performance.AvgCost
		member.mu.Unlock()

		if cost < lowestCost {
			lowestCost = cost
			selected = member
		}
	}

	if selected == nil {
		return nil, fmt.Errorf("no members available")
	}

	return selected, nil
}

// intelligentStrategy implements intelligent multi-factor routing
func (ir *IntelligentRouter) intelligentStrategy(ctx context.Context, members []*TeamMember, task *CollaborativeTask) (*TeamMember, error) {
	var selected *TeamMember
	bestScore := -1.0

	for _, member := range members {
		score := ir.calculateScore(member, task)
		if score > bestScore {
			bestScore = score
			selected = member
		}
	}

	if selected == nil {
		return nil, fmt.Errorf("no members available")
	}

	return selected, nil
}

// capabilityMatchStrategy implements capability-based routing
func (ir *IntelligentRouter) capabilityMatchStrategy(ctx context.Context, members []*TeamMember, task *CollaborativeTask) (*TeamMember, error) {
	// Find the member with the most matching capabilities
	var selected *TeamMember
	maxMatches := 0

	for _, member := range members {
		matches := ir.countCapabilityMatches(member, task.RequiredCapabilities)
		if matches > maxMatches {
			maxMatches = matches
			selected = member
		}
	}

	if selected == nil {
		return nil, fmt.Errorf("no members available")
	}

	return selected, nil
}

// priorityBasedStrategy implements priority-based routing
func (ir *IntelligentRouter) priorityBasedStrategy(ctx context.Context, members []*TeamMember, task *CollaborativeTask) (*TeamMember, error) {
	var selected *TeamMember
	highestPriority := -1

	for _, member := range members {
		if member.Priority > highestPriority {
			highestPriority = member.Priority
			selected = member
		}
	}

	if selected == nil {
		return nil, fmt.Errorf("no members available")
	}

	return selected, nil
}

// ===== Scoring =====

// calculateScore calculates a composite score for a member
func (ir *IntelligentRouter) calculateScore(member *TeamMember, task *CollaborativeTask) float64 {
	member.mu.Lock()
	defer member.mu.Unlock()

	weights := ir.config.ScoringWeights

	// Factor 1: Latency score (lower is better)
	latencyScore := ir.calculateLatencyScore(member)

	// Factor 2: Success rate score (higher is better)
	successScore := ir.calculateSuccessScore(member)

	// Factor 3: Load score (lower is better)
	loadScore := ir.calculateLoadScore(member)

	// Factor 4: Cost score (lower is better)
	costScore := ir.calculateCostScore(member)

	// Factor 5: Quality score (higher is better)
	qualityScore := ir.calculateQualityScore(member)

	// Factor 6: Priority score (higher is better)
	priorityScore := ir.calculatePriorityScore(member)

	// Calculate weighted sum
	totalScore :=
		latencyScore*weights.Latency +
			successScore*weights.Success +
			loadScore*weights.Load +
			costScore*weights.Cost +
			qualityScore*weights.Quality +
			priorityScore*weights.Priority

	return totalScore
}

// calculateLatencyScore calculates latency score (0.0-1.0, higher is better)
func (ir *IntelligentRouter) calculateLatencyScore(member *TeamMember) float64 {
	duration := member.Performance.AvgDuration.Seconds()

	// Use logarithmic scale to handle wide range of values
	// Lower duration = higher score
	return 1.0 / (1.0 + math.Log(1.0+duration))
}

// calculateSuccessScore calculates success rate score (0.0-1.0, higher is better)
func (ir *IntelligentRouter) calculateSuccessScore(member *TeamMember) float64 {
	if member.Performance.TotalTasks == 0 {
		return 0.5 // Neutral score for new members
	}

	return member.Performance.SuccessRate
}

// calculateLoadScore calculates load score (0.0-1.0, higher is better)
func (ir *IntelligentRouter) calculateLoadScore(member *TeamMember) float64 {
	// Lower active tasks = higher score
	return 1.0 / (1.0 + float64(member.ActiveTasks))
}

// calculateCostScore calculates cost score (0.0-1.0, higher is better)
func (ir *IntelligentRouter) calculateCostScore(member *TeamMember) float64 {
	cost := member.Performance.AvgCost

	// Lower cost = higher score
	// Assume max reasonable cost is 1.0
	return math.Max(0, 1.0-cost)
}

// calculateQualityScore calculates quality score (0.0-1.0, higher is better)
func (ir *IntelligentRouter) calculateQualityScore(member *TeamMember) float64 {
	return member.Performance.QualityScore
}

// calculatePriorityScore calculates priority score (0.0-1.0, higher is better)
func (ir *IntelligentRouter) calculatePriorityScore(member *TeamMember) float64 {
	// Normalize priority to 0.0-1.0 range
	return float64(member.Priority) / 9.0
}

// countCapabilityMatches counts how many required capabilities a member has
func (ir *IntelligentRouter) countCapabilityMatches(member *TeamMember, required []string) int {
	capMap := make(map[string]bool)
	for _, cap := range member.Capabilities {
		capMap[cap] = true
	}

	count := 0
	for _, req := range required {
		if capMap[req] {
			count++
		}
	}

	return count
}

// ===== RouterCache Methods =====

// Get gets a cached entry
func (rc *RouterCache) Get(key string) *TeamMember {
	if rc == nil {
		return nil
	}

	rc.mu.RLock()
	defer rc.mu.RUnlock()

	entry, ok := rc.entries[key]
	if !ok {
		return nil
	}

	// Check if expired
	if time.Now().After(entry.ExpireAt) {
		return nil
	}

	return entry.Member
}

// Set sets a cached entry
func (rc *RouterCache) Set(key string, member *TeamMember) {
	if rc == nil {
		return
	}

	rc.mu.Lock()
	defer rc.mu.Unlock()

	rc.entries[key] = &CacheEntry{
		Member:   member,
		ExpireAt: time.Now().Add(rc.ttl),
	}
}

// Clear clears the cache
func (rc *RouterCache) Clear() {
	if rc == nil {
		return
	}

	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.entries = make(map[string]*CacheEntry)
}

// ===== RouterMetrics Methods =====

// NewRouterMetrics creates a new router metrics
func NewRouterMetrics() *RouterMetrics {
	return &RouterMetrics{
		StrategyUsage:    make(map[RoutingStrategy]int64),
		SelectionHistory: make([]SelectionRecord, 0, 1000),
	}
}

// RecordSelection records a routing selection
func (rm *RouterMetrics) RecordSelection(taskID, taskType, memberName string, strategy RoutingStrategy, duration time.Duration, success bool) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rm.TotalSelections++
	rm.SelectionHistory = append(rm.SelectionHistory, SelectionRecord{
		Timestamp:  time.Now(),
		TaskID:     taskID,
		TaskType:   taskType,
		MemberName: memberName,
		Strategy:   strategy,
		Duration:   duration,
		Success:    success,
	})

	// Trim history if needed
	if len(rm.SelectionHistory) > 1000 {
		rm.SelectionHistory = rm.SelectionHistory[len(rm.SelectionHistory)-1000:]
	}
}

// RecordStrategyUsage records strategy usage
func (rm *RouterMetrics) RecordStrategyUsage(strategy RoutingStrategy) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rm.StrategyUsage[strategy]++
}

// RecordCacheHit records a cache hit
func (rm *RouterMetrics) RecordCacheHit() {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.CacheHits++
}

// RecordCacheMiss records a cache miss
func (rm *RouterMetrics) RecordCacheMiss() {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.CacheMisses++
}

// GetStats returns router statistics
func (ir *IntelligentRouter) GetStats() *RouterStats {
	ir.metrics.mu.RLock()
	defer ir.metrics.mu.RUnlock()

	stats := &RouterStats{
		TotalSelections:      ir.metrics.TotalSelections,
		CacheHits:            ir.metrics.CacheHits,
		CacheMisses:          ir.metrics.CacheMisses,
		StrategyUsage:        make(map[string]int64),
		AverageSelectionTime: 0,
	}

	for strategy, count := range ir.metrics.StrategyUsage {
		stats.StrategyUsage[strategy.String()] = count
	}

	// Calculate average selection time
	if len(ir.metrics.SelectionHistory) > 0 {
		var total time.Duration
		for _, record := range ir.metrics.SelectionHistory {
			total += record.Duration
		}
		stats.AverageSelectionTime = total / time.Duration(len(ir.metrics.SelectionHistory))
	}

	return stats
}

// RouterStats represents router statistics
type RouterStats struct {
	TotalSelections      int64
	CacheHits            int64
	CacheMisses          int64
	StrategyUsage        map[string]int64
	AverageSelectionTime time.Duration
}

// ClearCache clears the router cache
func (ir *IntelligentRouter) ClearCache() {
	if ir.cache != nil {
		ir.cache.Clear()
	}
}
