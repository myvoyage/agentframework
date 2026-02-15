package hello

import (
	"fmt"
)

type HelloTool struct{}

func (t *HelloTool) Name() string    { return "hello" }
func (t *HelloTool) Version() string { return "0.1" }
func (t *HelloTool) Execute(ctx interface{}, inputs map[string]interface{}) (map[string]interface{}, error) {
	name := "World"
	if v, ok := inputs["name"]; ok {
		if s, ok := v.(string); ok && s != "" {
			name = s
		}
	}
	greeting := fmt.Sprintf("Hello, %s!", name)
	return map[string]interface{}{"greeting": greeting}, nil
}

func NewHelloTool() *HelloTool { return &HelloTool{} }
