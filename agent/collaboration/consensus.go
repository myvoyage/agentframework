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
	"sort"
	"sync"
	"time"
)

// ConsensusStrategy defines the consensus strategy
type ConsensusStrategy int

const (
	ConsensusMajority ConsensusStrategy = iota
	ConsensusUnanimous
	ConsensusWeighted
	ConsensusBestN
)

// ConsensusManager manages consensus among multiple agents
type ConsensusManager struct {
	strategy ConsensusStrategy
	team     *AgentTeam
	timeout  time.Duration
}

// NewConsensusManager creates a new consensus manager
func NewConsensusManager(strategy ConsensusStrategy, team *AgentTeam, timeout time.Duration) *ConsensusManager {
	return &ConsensusManager{
		strategy: strategy,
		team:     team,
		timeout:  timeout,
	}
}

// ReachConsensus reaches consensus among agents for a task
func (cm *ConsensusManager) ReachConsensus(ctx context.Context, task *CollaborativeTask) (*ConsensusResult, error) {
	// Get all capable members
	members := cm.team.ListMembers()
	capableMembers := filterCapableMembers(members, task)

	if len(capableMembers) == 0 {
		return nil, fmt.Errorf("no capable members for consensus")
	}

	// Execute task with all capable members
	results := make([]*TaskResult, len(capableMembers))
	var wg sync.WaitGroup
	errChan := make(chan error, len(capableMembers))

	for i, member := range capableMembers {
		wg.Add(1)
		go func(idx int, m *TeamMember) {
			defer wg.Done()

			result, err := cm.team.AssignTask(ctx, task)
			if err != nil {
				errChan <- err
				return
			}

			results[idx] = result
		}(i, member)
	}

	// Wait for all tasks to complete or timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// All tasks completed
	case <-time.After(cm.timeout):
		return nil, fmt.Errorf("consensus timeout")
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	close(errChan)
	for err := range errChan {
		if err != nil {
			return nil, fmt.Errorf("task execution error: %w", err)
		}
	}

	// Apply consensus strategy
	switch cm.strategy {
	case ConsensusMajority:
		return cm.majorityConsensus(results)
	case ConsensusUnanimous:
		return cm.unanimousConsensus(results)
	case ConsensusWeighted:
		return cm.weightedConsensus(results)
	case ConsensusBestN:
		return cm.bestNConsensus(results)
	default:
		return nil, fmt.Errorf("unknown consensus strategy: %d", cm.strategy)
	}
}

// majorityConsensus implements majority voting
func (cm *ConsensusManager) majorityConsensus(results []*TaskResult) (*ConsensusResult, error) {
	// Group results by output
	outputGroups := make(map[string][]*TaskResult)
	for _, result := range results {
		if result.Success {
			outputGroups[result.Output] = append(outputGroups[result.Output], result)
		}
	}

	// Find the majority output
	var majorityOutput string
	var maxCount int
	for output, group := range outputGroups {
		if len(group) > maxCount {
			maxCount = len(group)
			majorityOutput = output
		}
	}

	if maxCount == 0 {
		return nil, fmt.Errorf("no successful results")
	}

	// Calculate confidence
	confidence := float64(maxCount) / float64(len(results))

	return &ConsensusResult{
		Output:       majorityOutput,
		Strategy:     "majority",
		Confidence:   confidence,
		Agreement:    float64(maxCount) / float64(len(results)),
		VoteCounts:   len(outputGroups),
		TotalVotes:   len(results),
		WinningVotes: maxCount,
		Results:      results,
	}, nil
}

// unanimousConsensus implements unanimous voting
func (cm *ConsensusManager) unanimousConsensus(results []*TaskResult) (*ConsensusResult, error) {
	// Check if all results are the same
	var firstOutput string
	allSame := true

	for _, result := range results {
		if !result.Success {
			return nil, fmt.Errorf("not all tasks succeeded")
		}

		if firstOutput == "" {
			firstOutput = result.Output
		} else if result.Output != firstOutput {
			allSame = false
			break
		}
	}

	if !allSame {
		return nil, fmt.Errorf("no unanimous agreement")
	}

	return &ConsensusResult{
		Output:       firstOutput,
		Strategy:     "unanimous",
		Confidence:   1.0,
		Agreement:    1.0,
		VoteCounts:   1,
		TotalVotes:   len(results),
		WinningVotes: len(results),
		Results:      results,
	}, nil
}

// weightedConsensus implements weighted voting based on member priority
func (cm *ConsensusManager) weightedConsensus(results []*TaskResult) (*ConsensusResult, error) {
	// Group results by output and accumulate weights
	outputWeights := make(map[string]float64)
	outputResults := make(map[string][]*TaskResult)

	for _, result := range results {
		if !result.Success {
			continue
		}

		// Get member priority
		member, err := cm.team.GetMember(result.AgentName)
		if err != nil {
			continue
		}

		weight := float64(member.Priority + 1) // +1 to ensure minimum weight of 1
		outputWeights[result.Output] += weight
		outputResults[result.Output] = append(outputResults[result.Output], result)
	}

	// Find the output with highest weight
	var bestOutput string
	var maxWeight float64
	for output, weight := range outputWeights {
		if weight > maxWeight {
			maxWeight = weight
			bestOutput = output
		}
	}

	if bestOutput == "" {
		return nil, fmt.Errorf("no successful results")
	}

	// Calculate confidence (normalized weight)
	totalWeight := 0.0
	for _, weight := range outputWeights {
		totalWeight += weight
	}
	confidence := maxWeight / totalWeight

	return &ConsensusResult{
		Output:       bestOutput,
		Strategy:     "weighted",
		Confidence:   confidence,
		Agreement:    confidence,
		VoteCounts:   len(outputWeights),
		TotalVotes:   len(results),
		WinningVotes: len(outputResults[bestOutput]),
		Results:      outputResults[bestOutput],
	}, nil
}

// bestNConsensus selects the best N results and aggregates them
func (cm *ConsensusManager) bestNConsensus(results []*TaskResult) (*ConsensusResult, error) {
	// Sort results by quality (duration, success, etc.)
	sortedResults := make([]*TaskResult, len(results))
	copy(sortedResults, results)

	sort.Slice(sortedResults, func(i, j int) bool {
		// Prefer successful results
		if sortedResults[i].Success != sortedResults[j].Success {
			return sortedResults[i].Success
		}

		// Prefer faster results
		return sortedResults[i].Duration < sortedResults[j].Duration
	})

	// Take top 3 results (or fewer if not enough)
	n := 3
	if len(sortedResults) < n {
		n = len(sortedResults)
	}

	bestResults := sortedResults[:n]

	// Aggregate outputs
	outputs := make([]string, 0, n)
	for _, result := range bestResults {
		if result.Success {
			outputs = append(outputs, result.Output)
		}
	}

	if len(outputs) == 0 {
		return nil, fmt.Errorf("no successful results")
	}

	// Combine outputs (simplified - just concatenate)
	combinedOutput := ""
	for i, output := range outputs {
		if i > 0 {
			combinedOutput += "\n\n"
		}
		combinedOutput += output
	}

	// Calculate confidence based on success rate
	confidence := float64(len(outputs)) / float64(len(bestResults))

	return &ConsensusResult{
		Output:       combinedOutput,
		Strategy:     "best_n",
		Confidence:   confidence,
		Agreement:    confidence,
		VoteCounts:   1,
		TotalVotes:   len(results),
		WinningVotes: len(outputs),
		Results:      bestResults,
	}, nil
}

// ConsensusResult represents the result of consensus
type ConsensusResult struct {
	Output       string
	Strategy     string
	Confidence   float64 // 0.0-1.0
	Agreement    float64 // 0.0-1.0
	VoteCounts   int
	TotalVotes   int
	WinningVotes int
	Results      []*TaskResult
	Timestamp    time.Time
}

// filterCapableMembers filters members that have required capabilities
func filterCapableMembers(members []*TeamMember, task *CollaborativeTask) []*TeamMember {
	var capable []*TeamMember

	for _, member := range members {
		if hasAllCapabilities(member, task.RequiredCapabilities) {
			capable = append(capable, member)
		}
	}

	return capable
}

// hasAllCapabilities checks if a member has all required capabilities
func hasAllCapabilities(member *TeamMember, required []string) bool {
	if len(required) == 0 {
		return true
	}

	capMap := make(map[string]bool)
	for _, cap := range member.Capabilities {
		capMap[cap] = true
	}

	for _, req := range required {
		if !capMap[req] {
			return false
		}
	}

	return true
}

// ConsensusMetrics tracks consensus metrics
type ConsensusMetrics struct {
	TotalConsensuses      int64
	SuccessfulConsensuses int64
	FailedConsensuses     int64
	AverageConfidence     float64
	AverageAgreement      float64
	AverageDuration       time.Duration
	StrategyUsage         map[string]int64
	mu                    sync.RWMutex
}

// RecordConsensus records a consensus operation
func (cm *ConsensusMetrics) RecordConsensus(result *ConsensusResult, duration time.Duration, success bool) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.TotalConsensuses++

	if success {
		cm.SuccessfulConsensuses++
	} else {
		cm.FailedConsensuses++
	}

	// Update averages
	n := float64(cm.TotalConsensuses)
	cm.AverageConfidence = (cm.AverageConfidence*(n-1) + result.Confidence) / n
	cm.AverageAgreement = (cm.AverageAgreement*(n-1) + result.Agreement) / n
	cm.AverageDuration = time.Duration((int64(cm.AverageDuration)*(int64(n)-1) + int64(duration)) / int64(n))

	// Update strategy usage
	if cm.StrategyUsage == nil {
		cm.StrategyUsage = make(map[string]int64)
	}
	cm.StrategyUsage[result.Strategy]++
}

// GetStats returns consensus statistics
func (cm *ConsensusMetrics) GetStats() *ConsensusStats {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return &ConsensusStats{
		TotalConsensuses:      cm.TotalConsensuses,
		SuccessfulConsensuses: cm.SuccessfulConsensuses,
		FailedConsensuses:     cm.FailedConsensuses,
		AverageConfidence:     cm.AverageConfidence,
		AverageAgreement:      cm.AverageAgreement,
		AverageDuration:       cm.AverageDuration,
		StrategyUsage:         cm.StrategyUsage,
	}
}

// ConsensusStats represents consensus statistics
type ConsensusStats struct {
	TotalConsensuses      int64
	SuccessfulConsensuses int64
	FailedConsensuses     int64
	AverageConfidence     float64
	AverageAgreement      float64
	AverageDuration       time.Duration
	StrategyUsage         map[string]int64
}

// String returns a string representation of the consensus result
func (cr *ConsensusResult) String() string {
	return fmt.Sprintf(
		"ConsensusResult{Strategy=%s, Confidence=%.2f, Agreement=%.2f, Votes=%d/%d}",
		cr.Strategy,
		cr.Confidence,
		cr.Agreement,
		cr.WinningVotes,
		cr.TotalVotes,
	)
}

// IsSuccessful returns true if consensus was reached successfully
func (cr *ConsensusResult) IsSuccessful() bool {
	return cr.Confidence >= 0.5 && len(cr.Results) > 0
}

// GetOutput returns the consensus output
func (cr *ConsensusResult) GetOutput() string {
	return cr.Output
}

// GetConfidence returns the confidence level
func (cr *ConsensusResult) GetConfidence() float64 {
	return cr.Confidence
}

// GetAgreement returns the agreement level
func (cr *ConsensusResult) GetAgreement() float64 {
	return cr.Agreement
}

// GetResults returns all results that contributed to consensus
func (cr *ConsensusResult) GetResults() []*TaskResult {
	return cr.Results
}
