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

// SPDX-License-Identifier: AGPL-3.0-or-later

package vfs

import (
	"fmt"
	"net/url"
	"path"
	"strings"

	"AgentFramework/pkg/beads/context"
)

// VikingURI viking:// URI 解析和构建
type VikingURI struct {
	scheme    string
	workspace string
	path      string
	layer     context.LayerType
	query     map[string]string
}

// NewVikingURI 创建新的 Viking URI
func NewVikingURI(rawURI string) (*VikingURI, error) {
	if !strings.HasPrefix(rawURI, "viking://") {
		return nil, fmt.Errorf("invalid viking URI: must start with viking://")
	}

	// 解析 URL
	u, err := url.Parse(rawURI)
	if err != nil {
		return nil, fmt.Errorf("parse URI: %w", err)
	}

	vuri := &VikingURI{
		scheme: u.Scheme,
		query:  make(map[string]string),
	}

	// 提取 host (workspace)
	vuri.workspace = u.Host

	// 提取路径
	vuri.path = u.Path

	// 解析查询参数
	for key, values := range u.Query() {
		if len(values) > 0 {
			vuri.query[key] = values[0]
		}
	}

	// 从查询参数中提取层级
	if layerStr, ok := vuri.query["layer"]; ok {
		vuri.layer = context.LayerType(layerStr)
	} else {
		vuri.layer = context.LayerAuto
	}

	return vuri, nil
}

// Scheme 返回 URI scheme
func (v *VikingURI) Scheme() string {
	return v.scheme
}

// Workspace 返回工作区
func (v *VikingURI) Workspace() string {
	return v.workspace
}

// Path 返回路径
func (v *VikingURI) Path() string {
	return v.path
}

// Layer 返回层级
func (v *VikingURI) Layer() context.LayerType {
	return v.layer
}

// Query 返回查询参数
func (v *VikingURI) Query() map[string]string {
	return v.query
}

// SetWorkspace 设置工作区
func (v *VikingURI) SetWorkspace(workspace string) *VikingURI {
	v.workspace = workspace
	return v
}

// SetPath 设置路径
func (v *VikingURI) SetPath(p string) *VikingURI {
	v.path = p
	return v
}

// SetLayer 设置层级
func (v *VikingURI) SetLayer(layer context.LayerType) *VikingURI {
	v.layer = layer
	if v.query == nil {
		v.query = make(map[string]string)
	}
	v.query["layer"] = string(layer)
	return v
}

// SetQuery 设置查询参数
func (v *VikingURI) SetQuery(key, value string) *VikingURI {
	if v.query == nil {
		v.query = make(map[string]string)
	}
	v.query[key] = value
	return v
}

// String 返回完整的 URI 字符串
func (v *VikingURI) String() string {
	u := &url.URL{
		Scheme: v.scheme,
		Host:   v.workspace,
		Path:   v.path,
	}

	// 构建查询参数
	if len(v.query) > 0 {
		query := u.Query()
		for key, value := range v.query {
			query.Set(key, value)
		}
		u.RawQuery = query.Encode()
	}

	return u.String()
}

// FullPath 返回完整路径 (workspace + path)
func (v *VikingURI) FullPath() string {
	return path.Join(v.workspace, v.path)
}

// Parent 返回父目录的 URI
func (v *VikingURI) Parent() *VikingURI {
	if v.path == "" || v.path == "/" {
		return nil
	}

	parentPath := path.Dir(v.path)
	if parentPath == "." {
		parentPath = ""
	}

	return &VikingURI{
		scheme:    v.scheme,
		workspace: v.workspace,
		path:      parentPath,
		layer:     v.layer,
		query:     make(map[string]string),
	}
}

// Join 拼接子路径
func (v *VikingURI) Join(subPath string) *VikingURI {
	newPath := path.Join(v.path, subPath)
	return &VikingURI{
		scheme:    v.scheme,
		workspace: v.workspace,
		path:      newPath,
		layer:     v.layer,
		query:     make(map[string]string),
	}
}

// WithLayer 返回带指定层级的 URI
func (v *VikingURI) WithLayer(layer context.LayerType) *VikingURI {
	return &VikingURI{
		scheme:    v.scheme,
		workspace: v.workspace,
		path:      v.path,
		layer:     layer,
		query:     make(map[string]string),
	}
}

// IsRoot 检查是否为根路径
func (v *VikingURI) IsRoot() bool {
	return v.path == "" || v.path == "/"
}

// IsValid 检查 URI 是否有效
func (v *VikingURI) IsValid() bool {
	return v.scheme == "viking" && v.workspace != ""
}

// URIOption URI 构建选项函数
type URIOption func(*VikingURI)

// WithWorkspaceOption 设置工作区的选项
func WithWorkspaceOption(workspace string) URIOption {
	return func(v *VikingURI) {
		v.SetWorkspace(workspace)
	}
}

// WithLayerOption 设置层级的选项
func WithLayerOption(layer context.LayerType) URIOption {
	return func(v *VikingURI) {
		v.SetLayer(layer)
	}
}

// WithQueryOption 设置查询参数的选项
func WithQueryOption(key, value string) URIOption {
	return func(v *VikingURI) {
		v.SetQuery(key, value)
	}
}

// BuildURI 构建 viking:// URI
func BuildURI(scheme, path string, opts ...URIOption) string {
	if scheme != "viking" {
		scheme = "viking"
	}

	vuri := &VikingURI{
		scheme: scheme,
		path:   path,
		query:  make(map[string]string),
	}

	for _, opt := range opts {
		opt(vuri)
	}

	return vuri.String()
}

// ParseURI 解析 URI 字符串
func ParseURI(uri string) (*VikingURI, error) {
	return NewVikingURI(uri)
}

// IsValidVikingURI 检查是否为有效的 viking URI
func IsValidVikingURI(uri string) bool {
	return strings.HasPrefix(uri, "viking://")
}

// ExtractWorkspace 从 URI 中提取工作区
func ExtractWorkspace(uri string) (string, error) {
	vuri, err := NewVikingURI(uri)
	if err != nil {
		return "", err
	}
	return vuri.Workspace(), nil
}

// ExtractPath 从 URI 中提取路径
func ExtractPath(uri string) (string, error) {
	vuri, err := NewVikingURI(uri)
	if err != nil {
		return "", err
	}
	return vuri.Path(), nil
}

// ExtractLayer 从 URI 中提取层级
func ExtractLayer(uri string) (context.LayerType, error) {
	vuri, err := NewVikingURI(uri)
	if err != nil {
		return context.LayerAuto, err
	}
	return vuri.Layer(), nil
}

// NormalizeURI 规范化 URI
func NormalizeURI(uri string) (string, error) {
	vuri, err := NewVikingURI(uri)
	if err != nil {
		return "", err
	}

	// 清理路径
	vuri.path = path.Clean(vuri.path)

	// 移除开头的斜杠（如果有）
	if strings.HasPrefix(vuri.path, "/") {
		vuri.path = strings.TrimPrefix(vuri.path, "/")
	}

	return vuri.String(), nil
}

// JoinURIs 拼接多个 URI 部分
func JoinURIs(base string, parts ...string) (string, error) {
	vuri, err := NewVikingURI(base)
	if err != nil {
		return "", err
	}

	result := vuri
	for _, part := range parts {
		result = result.Join(part)
	}

	return result.String(), nil
}
