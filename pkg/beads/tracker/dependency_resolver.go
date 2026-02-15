package tracker

import (
	"context"
	"fmt"

	"AgentFramework/pkg/beads"
)

// DependencyResolverImpl implements the beads.DependencyResolver interface
// It computes task ready states and validates dependency graphs using DFS cycle detection
type DependencyResolverImpl struct {
	store beads.SQLiteStore
}

// NewDependencyResolver creates a new DependencyResolver instance
func NewDependencyResolver(store beads.SQLiteStore) beads.DependencyResolver {
	return &DependencyResolverImpl{
		store: store,
	}
}

// ComputeReadyState determines if a task is ready to execute by checking
// if all blocking dependencies are completed
func (dr *DependencyResolverImpl) ComputeReadyState(ctx context.Context, taskID string) (bool, error) {
	// Get all incoming dependencies (tasks that block this one)
	incomingDeps, err := dr.store.ReadDependencies(ctx, taskID, beads.DirectionIncoming)
	if err != nil {
		return false, fmt.Errorf("failed to read dependencies: %w", err)
	}

	// Filter for blocking dependencies (blocks and parent-child types)
	blockingDeps := make([]*beads.Dependency, 0)
	for _, dep := range incomingDeps {
		if dep.Type == beads.DependencyTypeBlocks || dep.Type == beads.DependencyTypeParentChild {
			blockingDeps = append(blockingDeps, dep)
		}
	}

	// If no blocking dependencies, task is ready
	if len(blockingDeps) == 0 {
		return true, nil
	}

	// Check if all blocking tasks are completed
	for _, dep := range blockingDeps {
		task, err := dr.store.ReadTask(ctx, dep.FromTaskID)
		if err != nil {
			return false, fmt.Errorf("failed to read blocking task %s: %w", dep.FromTaskID, err)
		}

		// Task is ready only if all blocking tasks are completed
		if task.Status != beads.StatusCompleted && task.Status != beads.StatusCancelled {
			return false, nil
		}
	}

	return true, nil
}

// ValidateNoCycles checks if adding a dependency would create a cycle
// Uses Depth-First Search (DFS) algorithm for cycle detection
func (dr *DependencyResolverImpl) ValidateNoCycles(ctx context.Context, fromID, toID string, depType beads.DependencyType) error {
	// Only blocks and parent-child dependencies need cycle validation
	if depType != beads.DependencyTypeBlocks && depType != beads.DependencyTypeParentChild {
		return nil
	}

	// Special case: self-reference is always a cycle
	if fromID == toID {
		return fmt.Errorf("self-referential dependency detected: %s -> %s", fromID, toID)
	}

	// Use DFS to detect cycles
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	return dr.dfs(ctx, toID, fromID, visited, recStack)
}

// dfs performs depth-first search to detect cycles in the dependency graph
func (dr *DependencyResolverImpl) dfs(ctx context.Context, current, target string, visited, recStack map[string]bool) error {
	// Mark current node as visited and add to recursion stack
	visited[current] = true
	recStack[current] = true

	// Check outgoing dependencies from current node
	outgoingDeps, err := dr.store.ReadDependencies(ctx, current, beads.DirectionOutgoing)
	if err != nil {
		return fmt.Errorf("failed to read dependencies for %s: %w", current, err)
	}

	for _, dep := range outgoingDeps {
		// Only follow blocks and parent-child edges
		if dep.Type != beads.DependencyTypeBlocks && dep.Type != beads.DependencyTypeParentChild {
			continue
		}

		// If we've reached the target node in the recursion stack, we found a cycle
		if dep.ToTaskID == target {
			return fmt.Errorf("cycle detected: %s -> ... -> %s -> %s", target, current, dep.ToTaskID)
		}

		// If neighbor hasn't been visited, recurse
		if !visited[dep.ToTaskID] {
			if err := dr.dfs(ctx, dep.ToTaskID, target, visited, recStack); err != nil {
				return err
			}
		} else if recStack[dep.ToTaskID] {
			// Neighbor is in recursion stack - cycle detected
			return fmt.Errorf("cycle detected: %s -> ... -> %s", current, dep.ToTaskID)
		}
	}

	// Remove current node from recursion stack
	recStack[current] = false

	return nil
}

// GetBlockingTasks returns all tasks that are blocking the given task
// These are tasks with "blocks" or "parent-child" dependencies that are not completed
func (dr *DependencyResolverImpl) GetBlockingTasks(ctx context.Context, taskID string) ([]*beads.Task, error) {
	// Get all incoming dependencies
	incomingDeps, err := dr.store.ReadDependencies(ctx, taskID, beads.DirectionIncoming)
	if err != nil {
		return nil, fmt.Errorf("failed to read dependencies: %w", err)
	}

	// Collect blocking tasks
	blockingTasks := make([]*beads.Task, 0)
	for _, dep := range incomingDeps {
		// Only consider blocks and parent-child types
		if dep.Type != beads.DependencyTypeBlocks && dep.Type != beads.DependencyTypeParentChild {
			continue
		}

		task, err := dr.store.ReadTask(ctx, dep.FromTaskID)
		if err != nil {
			return nil, fmt.Errorf("failed to read task %s: %w", dep.FromTaskID, err)
		}

		// Task is blocking if not completed or cancelled
		if task.Status != beads.StatusCompleted && task.Status != beads.StatusCancelled {
			blockingTasks = append(blockingTasks, task)
		}
	}

	return blockingTasks, nil
}

// GetDependencyChain returns the full dependency chain for a task
// This includes all transitive dependencies (both incoming and outgoing)
func (dr *DependencyResolverImpl) GetDependencyChain(ctx context.Context, taskID string) ([]*beads.Task, error) {
	// Collect all related tasks using BFS
	queue := []string{taskID}
	visited := make(map[string]bool)
	taskIDs := make(map[string]bool)

	taskIDs[taskID] = true
	visited[taskID] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		// Get all dependencies (both incoming and outgoing)
		for _, direction := range []beads.Direction{beads.DirectionIncoming, beads.DirectionOutgoing} {
			deps, err := dr.store.ReadDependencies(ctx, current, direction)
			if err != nil {
				return nil, fmt.Errorf("failed to read dependencies for %s: %w", current, err)
			}

			for _, dep := range deps {
				var relatedID string
				if direction == beads.DirectionIncoming {
					relatedID = dep.FromTaskID
				} else {
					relatedID = dep.ToTaskID
				}

				// Add to queue if not visited
				if !visited[relatedID] {
					visited[relatedID] = true
					taskIDs[relatedID] = true
					queue = append(queue, relatedID)
				}
			}
		}
	}

	// Fetch all tasks in the chain
	tasks := make([]*beads.Task, 0, len(taskIDs))
	for id := range taskIDs {
		task, err := dr.store.ReadTask(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("failed to read task %s: %w", id, err)
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}
