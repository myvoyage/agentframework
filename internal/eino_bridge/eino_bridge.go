//go:build eino
// +build eino

package einobridge

import (
	"context"
	"fmt"
	"AgentFramework/internal/pipelineengine"
)

var bridgeEngine *pipelineengine.PipelineEngine

// SetBridgeEngine allows MVP to route calls through a bridging Engine instance.
func SetBridgeEngine(e *pipelineengine.PipelineEngine) {
	bridgeEngine = e
}

// InitBridge initializes the Eino bridge (currently a no-op until real wiring).
func InitBridge() error {
	// In MVP the bridge can be a no-op; if a bridge engine is set later, it will be used.
	if bridgeEngine == nil {
		// Not strictly required to fail; returning nil to allow builds without Eino
		return nil
	}
	return nil
}

// BridgeClient provides a minimal API surface for invoking tools via the bridge.
type BridgeClient struct{}

func (c *BridgeClient) InvokeTool(ctx context.Context, name string, params map[string]interface{}, contextData map[string]interface{}) (map[string]interface{}, error) {
	if bridgeEngine == nil {
		return nil, fmt.Errorf("bridge engine not set")
	}
	return bridgeEngine.InvokeTool(ctx, name, params, contextData)
}

func (c *BridgeClient) RunPipeline(ctx context.Context, p *pipelineengine.PipelineSpec) (*pipelineengine.ExecutionContext, error) {
	if bridgeEngine == nil {
		return nil, fmt.Errorf("bridge engine not set")
	}
	return bridgeEngine.RunPipeline(ctx, p)
}
