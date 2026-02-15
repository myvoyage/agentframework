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

// Additional permission under GNU Affero General Public License version 3 section 7
// If you modify this Program, or any covered work, by linking or combining it
// with other code, such other code is not for that reason alone subject to any
// of the requirements of the GNU Affero GPL version 3 as long as you maintain
// the separation between the Program and the other code.

// For network interaction purposes, when this Program is used over a network,
// the source code of the Program must be made available to users of the network.
// You can comply with this requirement by providing a link to the source code
// repository in your user interface or documentation.

// SPDX-License-Identifier: AGPL-3.0-or-later

package workflow

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

// AggregatingParallelWorkflow represents a workflow that runs multiple agents in parallel and aggregates their results.
type AggregatingParallelWorkflow struct {
	name           string
	agents         []Agent
	aggregator     Agent
	maxConcurrency int // Maximum number of concurrent agents to run
	timeout        int // Timeout in seconds per agent
}

// NewAggregatingParallelWorkflow creates a new AggregatingParallelWorkflow instance.
func NewAggregatingParallelWorkflow(name string, aggregator Agent, agents ...Agent) *AggregatingParallelWorkflow {
	return &AggregatingParallelWorkflow{
		name:           name,
		agents:         agents,
		aggregator:     aggregator,
		maxConcurrency: 10, // Default to 10 concurrent agents
		timeout:        30, // Default to 30 seconds timeout per agent
	}
}

func (w *AggregatingParallelWorkflow) Name() string {
	return w.name
}

func (w *AggregatingParallelWorkflow) Run(ctx context.Context, input string, opts ...model.Option) (*schema.Message, error) {
	type result struct {
		index int
		msg   *schema.Message
		err   error
	}

	var (
		wg          sync.WaitGroup
		results     = make([]*schema.Message, len(w.agents))
		errChan     = make(chan error, 1)
		mutex       sync.Mutex
		err         error
		total       = len(w.agents)
		concurrency = w.maxConcurrency
	)

	// Limit concurrency to number of agents if it's smaller
	if concurrency > total {
		concurrency = total
	}

	// Create a semaphore to control concurrency
	sem := make(chan struct{}, concurrency)

	for i, ag := range w.agents {
		wg.Add(1)
		i := i
		ag := ag

		go func() {
			defer wg.Done()

			// Acquire semaphore
			sem <- struct{}{}
			defer func() {
				// Release semaphore
				<-sem
			}()

			// Create timeout context
			timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(w.timeout)*time.Second)
			defer cancel()

			// Run agent with timeout
			msg, agentErr := ag.Run(timeoutCtx, input, opts...)

			mutex.Lock()
			defer mutex.Unlock()

			if agentErr != nil {
				// Only send first error
				if err == nil {
					err = fmt.Errorf("agent %d failed: %w", i, agentErr)
					errChan <- err
				}
				return
			}

			results[i] = msg
		}()
	}

	// Wait for all agents to complete or an error occurs
	waitDone := make(chan struct{})
	go func() {
		wg.Wait()
		close(waitDone)
	}()

	select {
	case <-waitDone:
		// All agents completed successfully
		if err != nil {
			return nil, err
		}
	case agentErr := <-errChan:
		// An agent failed
		return nil, agentErr
	case <-ctx.Done():
		// Context canceled
		return nil, ctx.Err()
	}

	if w.aggregator == nil {
		for _, m := range results {
			if m != nil {
				return m, nil
			}
		}
		return &schema.Message{Role: schema.Assistant, Content: ""}, nil
	}

	var combined string

	for _, m := range results {
		if m == nil {
			continue
		}
		if combined != "" {
			combined += "\n\n"
		}
		combined += m.Content
	}

	return w.aggregator.Run(ctx, combined, opts...)
}

func (w *AggregatingParallelWorkflow) Resume(ctx context.Context, runID string, input string, opts ...model.Option) (*schema.Message, error) {
	return nil, fmt.Errorf("resume not supported for aggregating parallel workflow")
}
