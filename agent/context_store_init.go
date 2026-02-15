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

package agent

import (
	"context"
	"fmt"

	beadscontext "AgentFramework/pkg/beads/context"
)

// initContextStore initializes a context store based on the provided configuration
func initContextStore(ctx context.Context, spec *ContextStoreSpec) (beadscontext.ContextStore, error) {
	if spec == nil || !spec.Enabled {
		return nil, nil
	}

	switch spec.Type {
	case "openviking":
		return initOpenVikingStore(ctx, &spec.OpenViking)
	case "memory":
		return initMemoryContextStore(ctx)
	case "none", "":
		return nil, nil
	default:
		return nil, fmt.Errorf("unknown context store type: %s", spec.Type)
	}
}

// initOpenVikingStore initializes an OpenViking context store
func initOpenVikingStore(ctx context.Context, spec *OpenVikingStoreSpec) (beadscontext.ContextStore, error) {
	// TODO: OpenViking integration not yet implemented
	// This is a placeholder for future OpenViking context store implementation
	return nil, fmt.Errorf("OpenViking context store not yet implemented")
}

// initMemoryContextStore initializes an in-memory context store
func initMemoryContextStore(ctx context.Context) (beadscontext.ContextStore, error) {
	// For now, return an error as memory context store is not yet implemented
	// This can be implemented later as a simple map-based store
	return nil, fmt.Errorf("memory context store not yet implemented")
}

// InitContextStoreWithConfig initializes a context store with custom configuration
// This is a convenience function that allows overriding the default configuration
func InitContextStoreWithConfig(
	ctx context.Context,
	storeType string,
	config map[string]interface{},
) (beadscontext.ContextStore, error) {
	spec := &ContextStoreSpec{
		Enabled: true,
		Type:    storeType,
	}

	// Parse custom configuration
	if endpoint, ok := config["endpoint"].(string); ok {
		spec.OpenViking.Endpoint = endpoint
	}
	if apiKey, ok := config["apiKey"].(string); ok {
		spec.OpenViking.APIKey = apiKey
	}
	if workspace, ok := config["workspace"].(string); ok {
		spec.OpenViking.Workspace = workspace
	}
	if timeout, ok := config["timeout"].(int); ok {
		spec.OpenViking.Timeout = timeout
	}
	if maxRetries, ok := config["maxRetries"].(int); ok {
		spec.OpenViking.MaxRetries = maxRetries
	}
	if autoSync, ok := config["autoSync"].(bool); ok {
		spec.OpenViking.AutoSync = autoSync
	}

	return initContextStore(ctx, spec)
}
