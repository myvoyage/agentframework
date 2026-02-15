//go:build !eino
// +build !eino

package einobridge

import "AgentFramework/internal/pipelineengine"

// Fallback bridge when the 'eino' build tag is not enabled. No-Op for MVP.
func InitBridge() error { return nil }

func SetBridgeEngine(e *pipelineengine.PipelineEngine) {}
