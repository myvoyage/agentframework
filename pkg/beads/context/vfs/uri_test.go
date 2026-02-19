// Agent Framework - VFS URI Tests
// Copyright (C) 2025  Agent Framework Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package vfs

import (
	"testing"

	"AgentFramework/pkg/beads/context"
)

func TestNewVikingURI(t *testing.T) {
	// Test valid URI
	uri, err := NewVikingURI("viking://workspace/path/to/file.txt")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if uri == nil {
		t.Fatal("expected URI to be created")
	}
	if uri.Scheme() != "viking" {
		t.Errorf("expected scheme 'viking', got '%s'", uri.Scheme())
	}
	if uri.Workspace() != "workspace" {
		t.Errorf("expected workspace 'workspace', got '%s'", uri.Workspace())
	}
	if uri.Path() != "/path/to/file.txt" {
		t.Errorf("expected path '/path/to/file.txt', got '%s'", uri.Path())
	}

	// Test invalid URI (no viking:// prefix)
	_, err = NewVikingURI("http://example.com")
	if err == nil {
		t.Error("expected error for non-viking URI")
	}
}

func TestVikingURI_LayerFromQuery(t *testing.T) {
	uri, _ := NewVikingURI("viking://workspace/path?layer=l0")
	if uri.Layer() != context.LayerTypeL0 {
		t.Errorf("expected LayerTypeL0, got %s", uri.Layer())
	}
}

func TestVikingURI_QueryParams(t *testing.T) {
	uri, _ := NewVikingURI("viking://workspace/path?key1=value1&key2=value2")
	query := uri.Query()
	if query["key1"] != "value1" {
		t.Errorf("expected key1 to be 'value1', got '%s'", query["key1"])
	}
	if query["key2"] != "value2" {
		t.Errorf("expected key2 to be 'value2', got '%s'", query["key2"])
	}
}

func TestVikingURI_DefaultLayer(t *testing.T) {
	uri, _ := NewVikingURI("viking://workspace/path")
	if uri.Layer() != context.LayerAuto {
		t.Errorf("expected LayerAuto when no layer specified, got %s", uri.Layer())
	}
}

func TestVikingURI_Setters(t *testing.T) {
	uri, _ := NewVikingURI("viking://workspace/path")

	// Test SetWorkspace
	newURI := uri.SetWorkspace("new-workspace")
	if newURI.Workspace() != "new-workspace" {
		t.Errorf("expected workspace 'new-workspace', got '%s'", newURI.Workspace())
	}

	// Test SetPath
	newURI = uri.SetPath("/new/path")
	if newURI.Path() != "/new/path" {
		t.Errorf("expected path '/new/path', got '%s'", newURI.Path())
	}

	// Test SetLayer
	newURI = uri.SetLayer(context.LayerTypeL1)
	if newURI.Layer() != context.LayerTypeL1 {
		t.Errorf("expected LayerTypeL1, got %s", newURI.Layer())
	}

	// Test SetQuery
	newURI = newURI.SetQuery("test", "value")
	query := newURI.Query()
	if query["test"] != "value" {
		t.Errorf("expected query test to be 'value', got '%s'", query["test"])
	}
}

func TestVikingURI_String(t *testing.T) {
	uri, _ := NewVikingURI("viking://workspace/path/to/file")

	str := uri.String()
	if str != "viking://workspace/path/to/file" {
		t.Errorf("expected 'viking://workspace/path/to/file', got '%s'", str)
	}

	// Test with query params
	uri.SetQuery("key", "value")
	str = uri.String()
	if str != "viking://workspace/path/to/file?key=value" {
		t.Errorf("expected URI with query params, got '%s'", str)
	}
}

func TestVikingURI_FullPath(t *testing.T) {
	uri, _ := NewVikingURI("viking://workspace/path/to/file")

	fullPath := uri.FullPath()
	if fullPath != "workspace/path/to/file" {
		t.Errorf("expected 'workspace/path/to/file', got '%s'", fullPath)
	}
}

func TestVikingURI_Parent(t *testing.T) {
	// Test parent of nested path
	uri, _ := NewVikingURI("viking://workspace/path/to/file.txt")
	parent := uri.Parent()
	if parent == nil {
		t.Error("expected parent to be returned")
	} else {
		if parent.Path() != "/path/to" {
			t.Errorf("expected parent path '/path/to', got '%s'", parent.Path())
		}
	}

	// Test parent of root path
	rootURI, _ := NewVikingURI("viking://workspace/")
	parent = rootURI.Parent()
	if parent != nil {
		t.Error("expected nil parent for root path")
	}
}

func TestVikingURI_Join(t *testing.T) {
	uri, _ := NewVikingURI("viking://workspace/base")
	joined := uri.Join("subdir/file.txt")
	if joined.Path() != "/base/subdir/file.txt" {
		t.Errorf("expected '/base/subdir/file.txt', got '%s'", joined.Path())
	}
}

func TestVikingURI_WithLayer(t *testing.T) {
	uri, _ := NewVikingURI("viking://workspace/path")
	newURI := uri.WithLayer(context.LayerTypeL2)
	if newURI.Layer() != context.LayerTypeL2 {
		t.Errorf("expected LayerTypeL2, got %s", newURI.Layer())
	}
	if newURI.Path() != "/path" {
		t.Errorf("expected path to remain '/path', got '%s'", newURI.Path())
	}
}

func TestVikingURI_IsRoot(t *testing.T) {
	// Test empty path
	uri, _ := NewVikingURI("viking://workspace")
	if !uri.IsRoot() {
		t.Error("expected empty path to be root")
	}

	// Test root path
	uri, _ = NewVikingURI("viking://workspace/")
	if !uri.IsRoot() {
		t.Error("expected '/' to be root")
	}

	// Test non-root path
	uri, _ = NewVikingURI("viking://workspace/path")
	if uri.IsRoot() {
		t.Error("expected non-root path")
	}
}

func TestVikingURI_IsValid(t *testing.T) {
	// Test valid URI
	uri, _ := NewVikingURI("viking://workspace/path")
	if !uri.IsValid() {
		t.Error("expected valid URI to be valid")
	}

	// Test invalid URI (no workspace)
	invalidURI := &VikingURI{
		scheme:    "viking",
		workspace: "",
		path:      "/path",
		layer:     context.LayerAuto,
		query:     make(map[string]string),
	}
	if invalidURI.IsValid() {
		t.Error("expected URI with no workspace to be invalid")
	}
}

func TestBuildURI(t *testing.T) {
	// Test basic build
	uri := BuildURI("viking", "/path/to/file")
	if uri != "viking:///path/to/file" {
		t.Errorf("expected 'viking:///path/to/file', got '%s'", uri)
	}

	// Test with options
	uri = BuildURI("viking", "/path",
		WithWorkspaceOption("my-workspace"),
		WithLayerOption(context.LayerTypeL1),
		WithQueryOption("format", "json"),
	)
	if uri != "viking://my-workspace/path?format=json&layer=l1" {
		t.Errorf("expected URI with options, got '%s'", uri)
	}
}

func TestParseURI(t *testing.T) {
	uri, err := ParseURI("viking://workspace/path/file.txt")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if uri.Workspace() != "workspace" {
		t.Errorf("expected workspace 'workspace', got '%s'", uri.Workspace())
	}
	if uri.Path() != "/path/file.txt" {
		t.Errorf("expected path '/path/file.txt', got '%s'", uri.Path())
	}
}

func TestIsValidVikingURI(t *testing.T) {
	if !IsValidVikingURI("viking://workspace/path") {
		t.Error("expected viking URI to be valid")
	}
	if IsValidVikingURI("http://example.com") {
		t.Error("expected non-viking URI to be invalid")
	}
	if IsValidVikingURI("") {
		t.Error("expected empty string to be invalid")
	}
}

func TestExtractWorkspace(t *testing.T) {
	ws, err := ExtractWorkspace("viking://my-workspace/path/file.txt")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if ws != "my-workspace" {
		t.Errorf("expected 'my-workspace', got '%s'", ws)
	}
}

func TestExtractPath(t *testing.T) {
	p, err := ExtractPath("viking://workspace/path/to/file.txt")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if p != "/path/to/file.txt" {
		t.Errorf("expected '/path/to/file.txt', got '%s'", p)
	}
}

func TestExtractLayer(t *testing.T) {
	// Test with explicit layer
	layer, err := ExtractLayer("viking://workspace/path?layer=l1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if layer != context.LayerTypeL1 {
		t.Errorf("expected LayerTypeL1, got %s", layer)
	}

	// Test without layer (should default to auto)
	layer, err = ExtractLayer("viking://workspace/path")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if layer != context.LayerAuto {
		t.Errorf("expected LayerAuto, got %s", layer)
	}
}

func TestNormalizeURI(t *testing.T) {
	// Test with messy path - path.Clean resolves ./ and ../
	// /path/./to/../file.txt -> /path/file.txt -> path/file.txt
	normalized, err := NormalizeURI("viking://workspace/path/./to/../file.txt")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if normalized != "viking://workspace/path/file.txt" {
		t.Errorf("expected 'viking://workspace/path/file.txt', got '%s'", normalized)
	}

	// Test with leading slash
	// //path -> /path -> path
	normalized, err = NormalizeURI("viking://workspace//path")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if normalized != "viking://workspace/path" {
		t.Errorf("expected 'viking://workspace/path', got '%s'", normalized)
	}

	// Test with complex path
	// /a/b/../c/./d -> /a/c/d -> a/c/d
	normalized, err = NormalizeURI("viking://workspace/a/b/../c/./d")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if normalized != "viking://workspace/a/c/d" {
		t.Errorf("expected 'viking://workspace/a/c/d', got '%s'", normalized)
	}
}

func TestJoinURIs(t *testing.T) {
	// Test joining base URI with parts
	joined, err := JoinURIs("viking://workspace/base", "subdir", "file.txt")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if joined != "viking://workspace/base/subdir/file.txt" {
		t.Errorf("expected joined URI, got '%s'", joined)
	}
}
