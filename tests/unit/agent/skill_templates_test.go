// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
// Copyright (C) 2025  Agent Framework Contributors

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.

// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

// Additional permission under GNU Affero General Public License version 3 section 7
// If you modify this Program, or any covered work, by linking or combining it
// with other code, such other code is not for that reason alone subject to any
// of the GNU Affero GPL version 3 as long as you maintain
// the separation between the Program and the other code.

// For network interaction purposes, when this Program is used over a network,
// the source code of the Program must be made available to users of the network.
// You can comply with this requirement by providing a link to the source code
// repository in your user interface or documentation.

// SPDX-License-Identifier: AGPL-3.0-or-later

package agent

import (
	"context"
	"encoding/json"
	"testing"
)

// TestBaseSkill functionality is now tested in skills package
// BaseSkill is now an internal type of the skills package

// TestHTTPRequestSkill tests the HTTPRequestSkill functionality
func TestHTTPRequestSkill(t *testing.T) {
	ctx := context.Background()

	// Create a HTTPRequestSkill
	skill := NewHTTPRequestSkill()
	if skill == nil {
		t.Fatal("Expected HTTPRequestSkill instance, got nil")
	}

	// Test Info method
	info, err := skill.Info(ctx)
	if err != nil {
		t.Fatalf("Expected no error from Info method, got: %v", err)
	}
	if info == nil {
		t.Fatal("Expected skill info, got nil")
	}
	if info.Name != "http_request" {
		t.Errorf("Expected skill name 'http_request', got '%s'", info.Name)
	}

	// Test IsEnabled method
	if !skill.IsEnabled(ctx) {
		t.Error("Expected skill to be enabled by default")
	}
}

// TestFileOperationSkill tests the FileOperationSkill functionality
func TestFileOperationSkill(t *testing.T) {
	ctx := context.Background()

	// Create a FileOperationSkill
	skill := NewFileOperationSkill()
	if skill == nil {
		t.Fatal("Expected FileOperationSkill instance, got nil")
	}

	// Test Info method
	info, err := skill.Info(ctx)
	if err != nil {
		t.Fatalf("Expected no error from Info method, got: %v", err)
	}
	if info == nil {
		t.Fatal("Expected skill info, got nil")
	}
	if info.Name != "file_operation" {
		t.Errorf("Expected skill name 'file_operation', got '%s'", info.Name)
	}

	// Test IsEnabled method
	if !skill.IsEnabled(ctx) {
		t.Error("Expected skill to be enabled by default")
	}

	// Test GetMetadata method
	metadata := skill.GetMetadata(ctx)
	if metadata.Category != "file" {
		t.Errorf("Expected metadata category 'file', got '%s'", metadata.Category)
	}
}

// TestCodeExecutionSkill tests the CodeExecutionSkill functionality
func TestCodeExecutionSkill(t *testing.T) {
	ctx := context.Background()

	// Create a CodeExecutionSkill
	skill := NewCodeExecutionSkill()
	if skill == nil {
		t.Fatal("Expected CodeExecutionSkill instance, got nil")
	}

	// Test Info method
	info, err := skill.Info(ctx)
	if err != nil {
		t.Fatalf("Expected no error from Info method, got: %v", err)
	}
	if info == nil {
		t.Fatal("Expected skill info, got nil")
	}
	if info.Name != "code_execution" {
		t.Errorf("Expected skill name 'code_execution', got '%s'", info.Name)
	}

	// Test IsEnabled method
	if !skill.IsEnabled(ctx) {
		t.Error("Expected skill to be enabled by default")
	}

	// Test GetMetadata method
	metadata := skill.GetMetadata(ctx)
	if metadata.Category != "code" {
		t.Errorf("Expected metadata category 'code', got '%s'", metadata.Category)
	}
}

// TestDataProcessingSkill tests the DataProcessingSkill functionality
func TestDataProcessingSkill(t *testing.T) {
	ctx := context.Background()

	// Create a DataProcessingSkill
	skill := NewDataProcessingSkill()
	if skill == nil {
		t.Fatal("Expected DataProcessingSkill instance, got nil")
	}

	// Test Info method
	info, err := skill.Info(ctx)
	if err != nil {
		t.Fatalf("Expected no error from Info method, got: %v", err)
	}
	if info == nil {
		t.Fatal("Expected skill info, got nil")
	}
	if info.Name != "data_processing" {
		t.Errorf("Expected skill name 'data_processing', got '%s'", info.Name)
	}

	// Test IsEnabled method
	if !skill.IsEnabled(ctx) {
		t.Error("Expected skill to be enabled by default")
	}

	// Test GetMetadata method
	metadata := skill.GetMetadata(ctx)
	if metadata.Category != "data" {
		t.Errorf("Expected metadata category 'data', got '%s'", metadata.Category)
	}

	// Test some data processing operations
	testData := []interface{}{1, 2, 3, 4, 5, 1, 2, 3}

	// Test unique operation
	uniqueInput := map[string]interface{}{
		"operation": "unique",
		"data":      testData,
	}
	uniqueInputJSON, _ := json.Marshal(uniqueInput)
	uniqueResult, err := skill.Invoke(ctx, string(uniqueInputJSON))
	if err != nil {
		t.Fatalf("Expected no error from unique operation, got: %v", err)
	}

	// Parse unique result
	var uniqueOutput map[string]interface{}
	if err := json.Unmarshal([]byte(uniqueResult), &uniqueOutput); err != nil {
		t.Fatalf("Expected valid JSON from unique operation, got: %v", err)
	}

	// Test count operation
	countInput := map[string]interface{}{
		"operation": "count",
		"data":      testData,
	}
	countInputJSON, _ := json.Marshal(countInput)
	countResult, err := skill.Invoke(ctx, string(countInputJSON))
	if err != nil {
		t.Fatalf("Expected no error from count operation, got: %v", err)
	}

	// Parse count result
	var countOutput map[string]interface{}
	if err := json.Unmarshal([]byte(countResult), &countOutput); err != nil {
		t.Fatalf("Expected valid JSON from count operation, got: %v", err)
	}

	// Test sum operation
	sumInput := map[string]interface{}{
		"operation": "sum",
		"data":      testData,
	}
	sumInputJSON, _ := json.Marshal(sumInput)
	sumResult, err := skill.Invoke(ctx, string(sumInputJSON))
	if err != nil {
		t.Fatalf("Expected no error from sum operation, got: %v", err)
	}

	// Parse sum result
	var sumOutput map[string]interface{}
	if err := json.Unmarshal([]byte(sumResult), &sumOutput); err != nil {
		t.Fatalf("Expected valid JSON from sum operation, got: %v", err)
	}
}

// TestFileExtensionFunction is now tested in skills package
// getFileExtension is now an internal function of the skills package

// TestConvertToNumber is now tested in skills package
// convertToNumber is now an internal function of the skills package
