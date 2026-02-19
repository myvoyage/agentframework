// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later


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
