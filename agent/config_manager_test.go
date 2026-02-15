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
	"testing"
)

func TestConfigManager(t *testing.T) {
	// Create a sample HostConfig
	cfg := &HostConfig{
		Name:         "test-host",
		DefaultModel: "test-model",
		Models: map[string]ModelConfig{
			"test-model": {
				Type:  "ollama",
				Model: "test-model",
			},
			"another-model": {
				Type:  "ollama",
				Model: "another-model",
			},
		},
		Agents: []AgentSpec{
			{
				Name:  "test-agent",
				Kind:  "chat",
				Model: "test-model",
			},
			{
				Name:  "another-agent",
				Kind:  "react",
				Model: "another-model",
			},
		},
		Workflows: []WorkflowSpec{
			{
				Name: "test-workflow",
				Kind: "sequential",
			},
			{
				Name: "another-workflow",
				Kind: "dag",
			},
		},
		Memory: MemoryManagementSpec{
			ModelCache: ModelCacheSpec{
				Enabled:         true,
				MaxSize:         100,
				TTL:             3600,
				CleanupInterval: 600,
			},
			MemoryMonitor: MemoryMonitorSpec{
				Enabled:        true,
				Interval:       5,
				HistorySize:    100,
				AlertThreshold: 512,
				AlertInterval:  60,
			},
		},
		ThreadStore: ThreadStoreSpec{
			Type: "memory",
		},
	}

	// Create a ConfigManager
	configMgr := NewConfigManager(cfg)

	// Test GetHostConfig
	hostCfg := configMgr.GetHostConfig()
	if hostCfg.Name != "test-host" {
		t.Errorf("Expected host name 'test-host', got '%s'", hostCfg.Name)
	}

	// Test GetModelConfig
	modelCfg, ok := configMgr.GetModelConfig("test-model")
	if !ok {
		t.Error("Expected to find 'test-model' config")
	}
	if modelCfg.Model != "test-model" {
		t.Errorf("Expected model name 'test-model', got '%s'", modelCfg.Model)
	}

	// Test GetModelConfig for non-existent model
	_, ok = configMgr.GetModelConfig("non-existent-model")
	if ok {
		t.Error("Expected not to find 'non-existent-model' config")
	}

	// Test GetAgentConfig
	agentCfg, ok := configMgr.GetAgentConfig("test-agent")
	if !ok {
		t.Error("Expected to find 'test-agent' config")
	}
	if agentCfg.Name != "test-agent" {
		t.Errorf("Expected agent name 'test-agent', got '%s'", agentCfg.Name)
	}

	// Test GetAgentConfig for non-existent agent
	_, ok = configMgr.GetAgentConfig("non-existent-agent")
	if ok {
		t.Error("Expected not to find 'non-existent-agent' config")
	}

	// Test GetWorkflowConfig
	workflowCfg, ok := configMgr.GetWorkflowConfig("test-workflow")
	if !ok {
		t.Error("Expected to find 'test-workflow' config")
	}
	if workflowCfg.Name != "test-workflow" {
		t.Errorf("Expected workflow name 'test-workflow', got '%s'", workflowCfg.Name)
	}

	// Test GetWorkflowConfig for non-existent workflow
	_, ok = configMgr.GetWorkflowConfig("non-existent-workflow")
	if ok {
		t.Error("Expected not to find 'non-existent-workflow' config")
	}

	// Test GetModelCacheConfig
	modelCacheCfg := configMgr.GetModelCacheConfig()
	if !modelCacheCfg.Enabled {
		t.Error("Expected model cache to be enabled")
	}
	if modelCacheCfg.MaxSize != 100 {
		t.Errorf("Expected model cache max size 100, got %d", modelCacheCfg.MaxSize)
	}

	// Test GetMemoryMonitorConfig
	memoryMonitorCfg := configMgr.GetMemoryMonitorConfig()
	if !memoryMonitorCfg.Enabled {
		t.Error("Expected memory monitor to be enabled")
	}
	if memoryMonitorCfg.Interval != 5 {
		t.Errorf("Expected memory monitor interval 5, got %d", memoryMonitorCfg.Interval)
	}

	// Test GetThreadStoreConfig
	threadStoreCfg := configMgr.GetThreadStoreConfig()
	if threadStoreCfg.Type != "memory" {
		t.Errorf("Expected thread store type 'memory', got '%s'", threadStoreCfg.Type)
	}
}
