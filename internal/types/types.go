package types

// Tool interface and registry abstraction for MVP
type Tool interface {
	Name() string
	Version() string
	Execute(ctx interface{}, inputs map[string]interface{}) (map[string]interface{}, error)
}

type ToolSpec struct {
	Name          string                 `yaml:"name"`
	Version       string                 `yaml:"version"`
	Description   string                 `yaml:"description"`
	InputsSchema  map[string]interface{} `yaml:"inputs_schema"`
	OutputsSchema map[string]interface{} `yaml:"outputs_schema"`
	HandlerRef    string                 `yaml:"handler_ref"`
}

type ToolRegistry interface {
	RegisterTool(t Tool) error
	GetTool(name string, version string) (Tool, error)
	ListTools() []ToolSpec
	GetToolSpec(name string) (ToolSpec, error)
}