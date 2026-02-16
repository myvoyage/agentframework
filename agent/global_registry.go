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

package agent

import (
	"context"
	"sync"
)

// Global tool registry instance
var (
	globalToolRegistry *DynamicToolRegistry
	globalRegistryOnce  sync.Once
)

// GetGlobalToolRegistry returns the global tool registry instance
// This replaces the TODO comment about loading tools from a registry
func GetGlobalToolRegistry() *DynamicToolRegistry {
	globalRegistryOnce.Do(func() {
		// Initialize with default loaders
		config := DynamicToolRegistryConfig{
			EnableCache:      true,
			EnableHotReload:  false, // Disable hot reload by default for simplicity
			HotReloadInterval: 30 * time.Second,
			InitialLoaders: []ToolLoader{
				NewHTTPToolLoader(),
				NewFileToolLoader(),
				NewPluginToolLoader(),
				NewBuiltinToolLoader(),
			},
		}
		
		registry, err := NewDynamicToolRegistry(config)
		if err != nil {
			panic(fmt.Sprintf("failed to create global tool registry: %v", err))
		}
		
		globalToolRegistry = registry
		
		// Load built-in tools
		if err := registry.LoadBuiltinTools(context.Background()); err != nil {
			// Log warning but don't fail
			println(fmt.Sprintf("Warning: failed to load built-in tools: %v", err))
		}
	})
	
	return globalToolRegistry
}

// SetGlobalToolRegistry sets the global tool registry (mainly for testing)
func SetGlobalToolRegistry(registry *DynamicToolRegistry) {
	globalToolRegistry = registry
}