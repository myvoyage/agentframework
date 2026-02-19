// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later


//go:build !eino
// +build !eino

package einobridge

import "AgentFramework/internal/pipelineengine"

// Fallback bridge when the 'eino' build tag is not enabled. No-Op for MVP.
func InitBridge() error { return nil }

func SetBridgeEngine(e *pipelineengine.PipelineEngine) {}
