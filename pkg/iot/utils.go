// Agent Framework - IoT Package Utilities
// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package iot

// getStringFromInterface safely extracts a string value from a map.
// If the key doesn't exist or the value is not a string, returns the default value.
func getStringFromInterface(m map[string]interface{}, key, defaultVal string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultVal
}

// getIntFromInterface safely extracts an int value from a map.
// If the key doesn't exist or the value is not an int, returns the default value.
func getIntFromInterface(m map[string]interface{}, key string, defaultVal int) int {
	if val, ok := m[key]; ok {
		// Handle float64 (JSON numbers are parsed as float64)
		if f, ok := val.(float64); ok {
			return int(f)
		}
		if i, ok := val.(int); ok {
			return i
		}
	}
	return defaultVal
}

// getBoolFromInterface safely extracts a bool value from a map.
// If the key doesn't exist or the value is not a bool, returns the default value.
func getBoolFromInterface(m map[string]interface{}, key string, defaultVal bool) bool {
	if val, ok := m[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return defaultVal
}
