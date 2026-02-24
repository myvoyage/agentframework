// Agent Framework - API Handlers
// SPDX-License-Identifier: AGPL-3.0-or-later

package api

import (
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/mux"
)

// Ensure core package is available for type references
import _ "AgentFramework/core"

// ========== 健康检查 ==========

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"status":  "healthy",
		"time":    s.getTime(),
		"version": "1.0.0",
		"services": map[string]bool{
			"workflow": s.core.GetWorkflowManager() != nil,
			"host":     s.core.GetHost() != nil,
			"skills":   s.core.GetSkillSystem() != nil,
		},
	}

	s.respondSuccess(w, status)
}

// ========== 工作流 API 处理器 ==========

type CreateWorkflowRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Definition  string `json:"definition,omitempty"`
	Type        string `json:"type,omitempty"` // sequential, parallel, dag, graph
}

type UpdateWorkflowRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Definition  string `json:"definition,omitempty"`
}

type ExecuteWorkflowRequest struct {
	Input string `json:"input"`
}

func (s *Server) listWorkflowsHandler(w http.ResponseWriter, r *http.Request) {
	workflows, err := s.core.GetWorkflowManager().GetWorkflows(r.Context())
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "Failed to list workflows", err)
		return
	}

	s.respondSuccess(w, workflows)
}

func (s *Server) createWorkflowHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateWorkflowRequest
	if err := s.parseJSONBody(r, &req); err != nil {
		s.respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if req.Name == "" {
		s.respondError(w, http.StatusBadRequest, "Workflow name is required", nil)
		return
	}

	// 创建工作流
	id, err := s.core.GetWorkflowManager().CreateWorkflow(
		r.Context(),
		req.Name,
		req.Description,
		req.Definition,
	)

	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "Failed to create workflow", err)
		return
	}

	s.respondSuccess(w, map[string]interface{}{
		"id":   id,
		"name": req.Name,
	})
}

func (s *Server) getWorkflowHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	workflow, err := s.core.GetWorkflowManager().GetWorkflow(r.Context(), id)
	if err != nil {
		s.respondError(w, http.StatusNotFound, "Workflow not found", err)
		return
	}

	s.respondSuccess(w, workflow)
}

func (s *Server) updateWorkflowHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req UpdateWorkflowRequest
	if err := s.parseJSONBody(r, &req); err != nil {
		s.respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// 更新工作流
	err := s.core.GetWorkflowManager().UpdateWorkflow(
		r.Context(),
		id,
		req.Name,
		req.Description,
		req.Definition,
	)

	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "Failed to update workflow", err)
		return
	}

	s.respondSuccess(w, map[string]interface{}{
		"id":      id,
		"updated": true,
	})
}

func (s *Server) deleteWorkflowHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	err := s.core.GetWorkflowManager().DeleteWorkflow(r.Context(), id)
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "Failed to delete workflow", err)
		return
	}

	s.respondSuccess(w, map[string]interface{}{
		"id":      id,
		"deleted": true,
	})
}

func (s *Server) executeWorkflowHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req ExecuteWorkflowRequest
	if err := s.parseJSONBody(r, &req); err != nil {
		s.respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// 执行工作流
	result, err := s.core.GetWorkflowManager().ExecuteWorkflow(r.Context(), id, req.Input)
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "Failed to execute workflow", err)
		return
	}

	// 通知所有 WebSocket 客户端
	s.Broadcast(map[string]interface{}{
		"type":      "workflow_executed",
		"id":        id,
		"result":    result,
		"timestamp": s.getTime(),
	})

	s.respondSuccess(w, map[string]interface{}{
		"executionId": result,
		"workflowId":  id,
		"output":      result,
	})
}

func (s *Server) getWorkflowVersionsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	workflowID := vars["id"]

	versions, err := s.core.GetWorkflowManager().GetWorkflowVersions(r.Context(), workflowID)
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "Failed to get workflow versions", err)
		return
	}

	s.respondSuccess(w, versions)
}

// ========== 技能 API 处理器 ==========

type RegisterSkillRequest struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Version     string                 `json:"version"`
	Handler     string                 `json:"handler"` // JavaScript 代码或内置处理器名称
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Enabled     bool                   `json:"enabled"`
}

func (s *Server) listSkillsHandler(w http.ResponseWriter, r *http.Request) {
	skills := s.core.GetSkillLibrary().GetAllSkills(r.Context())

	// 获取技能详情
	skillDetails := make([]map[string]interface{}, 0)
	for skillID, skill := range skills {
		info, err := skill.Info(r.Context())
		if err == nil {
			skillDetails = append(skillDetails, map[string]interface{}{
				"id":          skillID,
				"name":        info.Name,
				"description": info.Desc,
				"version":     "1.0.0",
				"enabled":     skill.IsEnabled(r.Context()),
				"metadata":    map[string]interface{}{},
			})
		}
	}

	s.respondSuccess(w, skillDetails)
}

func (s *Server) registerSkillHandler(w http.ResponseWriter, r *http.Request) {
	var req RegisterSkillRequest
	if err := s.parseJSONBody(r, &req); err != nil {
		s.respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// 这里需要创建一个动态技能
	// 实际实现中需要根据 handler 类型创建相应的技能

	s.respondSuccess(w, map[string]interface{}{
		"id":   req.ID,
		"name": req.Name,
	})
}

func (s *Server) getSkillHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	skill, found := s.core.GetSkillLibrary().GetSkill(r.Context(), id)
	if !found {
		s.respondError(w, http.StatusNotFound, "Skill not found", nil)
		return
	}

	info, err := skill.Info(r.Context())
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "Failed to get skill info", err)
		return
	}

	s.respondSuccess(w, map[string]interface{}{
		"info":    info,
		"enabled": skill.IsEnabled(r.Context()),
	})
}

func (s *Server) updateSkillHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var data map[string]interface{}
	if err := s.parseJSONBody(r, &data); err != nil {
		s.respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// 更新技能元数据
	// 实际实现需要根据 data 内容更新技能

	s.respondSuccess(w, map[string]interface{}{
		"id":      id,
		"updated": true,
	})
}

func (s *Server) deleteSkillHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	err := s.core.GetSkillLibrary().UnregisterSkill(r.Context(), id)
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "Failed to delete skill", err)
		return
	}

	s.respondSuccess(w, map[string]interface{}{
		"id":      id,
		"deleted": true,
	})
}

func (s *Server) enableSkillHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := s.core.GetSkillLibrary().EnableSkill(r.Context(), id); err != nil {
		s.respondError(w, http.StatusInternalServerError, "Failed to enable skill", err)
		return
	}

	s.respondSuccess(w, map[string]interface{}{
		"id":      id,
		"enabled": true,
	})
}

func (s *Server) disableSkillHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := s.core.GetSkillLibrary().DisableSkill(r.Context(), id); err != nil {
		s.respondError(w, http.StatusInternalServerError, "Failed to disable skill", err)
		return
	}

	s.respondSuccess(w, map[string]interface{}{
		"id":      id,
		"enabled": false,
	})
}

// ========== 配置 API 处理器 ==========

func (s *Server) getConfigHandler(w http.ResponseWriter, r *http.Request) {
	config := s.core.GetHost().Config()

	s.respondSuccess(w, config)
}

func (s *Server) updateConfigHandler(w http.ResponseWriter, r *http.Request) {
	var config map[string]interface{}
	if err := s.parseJSONBody(r, &config); err != nil {
		s.respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// 更新配置
	// 实际实现需要保存到文件并重新加载

	s.respondSuccess(w, map[string]interface{}{
		"updated": true,
	})
}

// ========== 文件系统 API 处理器 ==========

type ListFilesRequest struct {
	Path  string `json:"path"`
	Depth int    `json:"depth,omitempty"`
}

type ReadFileRequest struct {
	Path string `json:"path"`
}

type WriteFileRequest struct {
	Path    string `json:"path"`
	Content string `json:"content"`
	Mode    string `json:"mode,omitempty"` // "overwrite", "append"
}

type CopyFileRequest struct {
	Src       string `json:"src"`
	Dst       string `json:"dst"`
	Overwrite bool   `json:"overwrite,omitempty"`
}

type MoveFileRequest struct {
	Src string `json:"src"`
	Dst string `json:"dst"`
}

type CreateDirectoryRequest struct {
	Path string `json:"path"`
	Mode string `json:"mode,omitempty"`
}

func (s *Server) listFilesHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		path = "."
	}

	// 使用文件浏览器获取文件列表
	files, err := s.core.GetFileExplorer().ListFiles(r.Context(), path)
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "Failed to list files", err)
		return
	}

	s.respondSuccess(w, files)
}

func (s *Server) readFileHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		s.respondError(w, http.StatusBadRequest, "Path parameter is required", nil)
		return
	}

	content, err := s.core.GetFileExplorer().ReadFile(r.Context(), path)
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "Failed to read file", err)
		return
	}

	s.respondSuccess(w, map[string]interface{}{
		"path":    path,
		"content": content,
	})
}

func (s *Server) writeFileHandler(w http.ResponseWriter, r *http.Request) {
	var req WriteFileRequest
	if err := s.parseJSONBody(r, &req); err != nil {
		s.respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	err := s.core.GetFileExplorer().WriteFile(r.Context(), req.Path, req.Content)
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "Failed to write file", err)
		return
	}

	// 广播文件变化
	s.Broadcast(map[string]interface{}{
		"type":      "file_changed",
		"path":      req.Path,
		"action":    "written",
		"timestamp": s.getTime(),
	})

	s.respondSuccess(w, map[string]interface{}{
		"path":    req.Path,
		"written": true,
	})
}

func (s *Server) deleteFileHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		s.respondError(w, http.StatusBadRequest, "Path parameter is required", nil)
		return
	}

	err := s.core.GetFileExplorer().DeleteFile(r.Context(), path)
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "Failed to delete file", err)
		return
	}

	s.respondSuccess(w, map[string]interface{}{
		"path":    path,
		"deleted": true,
	})
}

func (s *Server) copyFileHandler(w http.ResponseWriter, r *http.Request) {
	var req CopyFileRequest
	if err := s.parseJSONBody(r, &req); err != nil {
		s.respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	err := s.core.GetFileExplorer().CopyFile(r.Context(), req.Src, req.Dst)
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "Failed to copy file", err)
		return
	}

	s.respondSuccess(w, map[string]interface{}{
		"src":    req.Src,
		"dst":    req.Dst,
		"copied": true,
	})
}

func (s *Server) moveFileHandler(w http.ResponseWriter, r *http.Request) {
	var req MoveFileRequest
	if err := s.parseJSONBody(r, &req); err != nil {
		s.respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	err := os.Rename(req.Src, req.Dst)
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "Failed to move file", err)
		return
	}

	s.respondSuccess(w, map[string]interface{}{
		"src":   req.Src,
		"dst":   req.Dst,
		"moved": true,
	})
}

func (s *Server) createDirectoryHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateDirectoryRequest
	if err := s.parseJSONBody(r, &req); err != nil {
		s.respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	err := os.MkdirAll(req.Path, 0755)
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "Failed to create directory", err)
		return
	}

	s.respondSuccess(w, map[string]interface{}{
		"path":    req.Path,
		"created": true,
	})
}

// ========== Agent API 处理器 ==========

type ChatRequest struct {
	Message string                 `json:"message"`
	Context string                 `json:"context,omitempty"`
	Options map[string]interface{} `json:"options,omitempty"`
	Stream  bool                   `json:"stream,omitempty"`
}

func (s *Server) listAgentsHandler(w http.ResponseWriter, r *http.Request) {
	agents := s.core.GetHost().ListAgents()

	s.respondSuccess(w, agents)
}

func (s *Server) chatHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	agentID := vars["id"]

	var req ChatRequest
	if err := s.parseJSONBody(r, &req); err != nil {
		s.respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// 获取 agent 并执行对话
	agent, err := s.core.GetHost().GetAgent(agentID)
	if err != nil {
		s.respondError(w, http.StatusNotFound, "Agent not found", err)
		return
	}

	response, err := agent.Run(r.Context(), req.Message)
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "Chat failed", err)
		return
	}

	s.respondSuccess(w, map[string]interface{}{
		"agentId":  agentID,
		"response": response,
	})
}

// ========== 辅助方法 ==========

func (s *Server) getTime() int64 {
	return time.Now().Unix()
}

func (s *Server) resolvePath(path string) (string, error) {
	if filepath.IsAbs(path) {
		return path, nil
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	return filepath.Join(cwd, path), nil
}
