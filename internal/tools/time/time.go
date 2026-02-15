package timeutil

import (
	"time"
)

type TimeTool struct{}

func (t *TimeTool) Name() string    { return "time_now" }
func (t *TimeTool) Version() string { return "0.1" }
func (t *TimeTool) Execute(ctx interface{}, inputs map[string]interface{}) (map[string]interface{}, error) {
	now := time.Now().Format(time.RFC3339)
	return map[string]interface{}{"now": now}, nil
}

func NewTimeTool() *TimeTool { return &TimeTool{} }
