package registry

import (
	"fmt"

	"AgentFramework/internal/types"
)

// InMemoryToolRegistry is a minimal in-memory registry that stores both Tool implementations
// and their corresponding ToolSpec (JSON-like schemas) for MVP validation.
type InMemoryToolRegistry struct {
	tools map[string]types.Tool
	specs map[string]types.ToolSpec
}

func NewInMemoryToolRegistry() *InMemoryToolRegistry {
	return &InMemoryToolRegistry{tools: make(map[string]types.Tool), specs: make(map[string]types.ToolSpec)}
}

// RegisterTool registers a Tool implementation (no schema yet).
func (r *InMemoryToolRegistry) RegisterTool(t types.Tool) error {
	if t == nil {
		return fmt.Errorf("tool is nil")
	}
	r.tools[t.Name()] = t
	return nil
}

// SetToolSpec associates a ToolSpec (schema) with a tool name.
func (r *InMemoryToolRegistry) SetToolSpec(name string, spec types.ToolSpec) error {
	r.specs[name] = spec
	return nil
}

// GetTool retrieves a Tool by name. Version field is reserved for future use.
func (r *InMemoryToolRegistry) GetTool(name string, version string) (types.Tool, error) {
	t, ok := r.tools[name]
	if !ok {
		return nil, fmt.Errorf("tool not found: %s", name)
	}
	return t, nil
}

func (r *InMemoryToolRegistry) GetToolSpec(name string) (types.ToolSpec, error) {
	s, ok := r.specs[name]
	if !ok {
		return types.ToolSpec{}, fmt.Errorf("tool spec not found: %s", name)
	}
	return s, nil
}

// ListTools returns all registered ToolSpecs
func (r *InMemoryToolRegistry) ListTools() []types.ToolSpec {
	out := make([]types.ToolSpec, 0, len(r.specs))
	for _, s := range r.specs {
		out = append(out, s)
	}
	return out
}
