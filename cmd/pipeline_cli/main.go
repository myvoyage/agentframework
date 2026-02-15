package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	einobridge "AgentFramework/internal/eino_bridge"
	pe "AgentFramework/internal/pipelineengine"
	reg "AgentFramework/internal/registry"
	hello "AgentFramework/internal/tools/hello"
	timeutil "AgentFramework/internal/tools/time"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: pipeline_cli <pipeline.yaml>")
		os.Exit(1)
	}
	path := os.Args[1]
	data, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read pipeline: %v\n", err)
		os.Exit(1)
	}

	// Registry with two demo tools
	regImpl := reg.NewInMemoryToolRegistry()
	// Register Hello and Time tools via adapters
	regImpl.RegisterTool(&helloToolAdapter{})
	regImpl.RegisterTool(&timeToolAdapter{})
	regImpl.ListTools()

	// Initialize pipeline engine
	engine := pe.NewPipelineEngine(regImpl)

	// Initialize Eino bridge (enable via build tag 'eino' if needed)
	_ = einobridge.InitBridge()
	einobridge.SetBridgeEngine(engine)
	// Expose engine to Eino bridge for future RPC routing
	einobridge.SetBridgeEngine(engine)

	p, err := engine.LoadPipeline(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load pipeline: %v\n", err)
		os.Exit(1)
	}
	ctx := context.Background()
	exec, err := engine.RunPipeline(ctx, p)
	if err != nil {
		fmt.Fprintf(os.Stderr, "execution failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Pipeline %s finished. Outputs: %v\n", p.Id, exec.Outputs)
}

// helloToolAdapter and timeToolAdapter are minimal adapters to satisfy the Tool interface
type helloToolAdapter struct{}

func (a *helloToolAdapter) Name() string    { return "hello" }
func (a *helloToolAdapter) Version() string { return "0.1" }
func (a *helloToolAdapter) Execute(ctx interface{}, inputs map[string]interface{}) (map[string]interface{}, error) {
	// simple wrapper to reuse actual tool implementation located in internal/tools/hello
	st := hello.NewHelloTool()
	return st.Execute(ctx, inputs)
}

// time adapter
type timeToolAdapter struct{}

func (a *timeToolAdapter) Name() string    { return "time_now" }
func (a *timeToolAdapter) Version() string { return "0.1" }
func (a *timeToolAdapter) Execute(ctx interface{}, inputs map[string]interface{}) (map[string]interface{}, error) {
	tt := timeutil.NewTimeTool()
	return tt.Execute(ctx, inputs)
}
