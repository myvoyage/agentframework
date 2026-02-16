package main

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"

	"AgentFramework/agent"
	"AgentFramework/agent/skills"
	"AgentFramework/core"
)

// App struct wraps the core application for desktop usage
type App struct {
	core *core.Application
	ctx  context.Context
}

// InitOpenTelemetry initializes OpenTelemetry tracing
func InitOpenTelemetry(ctx context.Context) (*trace.TracerProvider, error) {
	// Create stdout trace exporter for demo purposes
	exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout trace exporter: %w", err)
	}

	// Create resource with service information
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String("agentframework-desktop"),
		semconv.ServiceVersionKey.String("1.0.0"),
		semconv.ServiceInstanceIDKey.String("desktop"),
	)

	// Create trace provider
	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(res),
	)

	// Set global tracer provider
	otel.SetTracerProvider(tp)

	return tp, nil
}

// NewApp creates a new App application struct
// Refactored to use core.Application (DRY principle)
func NewApp() *App {
	ctx := context.Background()

	// Initialize OpenTelemetry
	_, err := InitOpenTelemetry(ctx)
	if err != nil {
		fmt.Printf("Warning: Failed to initialize OpenTelemetry: %v\n", err)
	}

	// 创建默认的HostConfig
	defaultHostConfig := &agent.HostConfig{
		Models: map[string]agent.ModelConfig{
			"default": {
				Type:  "ollama",
				Model: "llama3",
			},
		},
		SkillSystemDir: ".skills", // 启用技能系统
	}

	// 创建模型工厂
	modelFactory := agent.NewModelFactoryWithConfig(agent.ModelConfig{
		Type:  "ollama",
		Model: "llama3",
	})

	// Create core application (shared between CLI and desktop)
	coreApp, err := core.NewApplication(ctx, defaultHostConfig, modelFactory, nil)
	if err != nil {
		panic(fmt.Errorf("failed to create core application: %w", err))
	}

	// Initialize core application
	if err := coreApp.Initialize(ctx); err != nil {
		panic(fmt.Errorf("failed to initialize core application: %w", err))
	}

	return &App{
		core: coreApp,
		ctx:  ctx,
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	// 初始化工作流管理器
	a.core.GetWorkflowManager().Init(ctx)
	// 初始化文件浏览器
	a.core.GetFileExplorer().Init(ctx)
}

// getSkillSystemBaseDir returns the base directory for the skill system
func (a *App) getSkillSystemBaseDir() string {
	if a.core.GetHost() != nil {
		cfg := a.core.GetHost().Config()
		if cfg != nil && cfg.SkillSystemDir != "" {
			return cfg.SkillSystemDir
		}
	}
	return ".skills" // 默认值
}

// ===== 工作流管理 API =====

// CreateWorkflow creates a new workflow
func (a *App) CreateWorkflow(name string, description string, definition ...string) (string, error) {
	return a.core.GetWorkflowManager().CreateWorkflow(a.ctx, name, description, definition...)
}

// GetWorkflows returns all workflows
func (a *App) GetWorkflows() ([]*agent.WorkflowInfo, error) {
	return a.core.GetWorkflowManager().GetWorkflows(a.ctx)
}

// GetWorkflow returns a workflow by ID
func (a *App) GetWorkflow(id string) (*agent.WorkflowInfo, error) {
	return a.core.GetWorkflowManager().GetWorkflow(a.ctx, id)
}

// UpdateWorkflow updates a workflow
func (a *App) UpdateWorkflow(id string, name string, description string, definition string) error {
	return a.core.GetWorkflowManager().UpdateWorkflow(a.ctx, id, name, description, definition)
}

// DeleteWorkflow deletes a workflow
func (a *App) DeleteWorkflow(id string) error {
	return a.core.GetWorkflowManager().DeleteWorkflow(a.ctx, id)
}

// ExecuteWorkflow executes a workflow
func (a *App) ExecuteWorkflow(id string, input string) (string, error) {
	return a.core.GetWorkflowManager().ExecuteWorkflow(a.ctx, id, input)
}

// GetWorkflowVersions returns all versions of a workflow
func (a *App) GetWorkflowVersions(workflowID string) ([]*agent.WorkflowVersion, error) {
	return a.core.GetWorkflowManager().GetWorkflowVersions(a.ctx, workflowID)
}

// GetWorkflowVersion returns a specific version of a workflow
func (a *App) GetWorkflowVersion(workflowID string, version int) (*agent.WorkflowVersion, error) {
	return a.core.GetWorkflowManager().GetWorkflowVersion(a.ctx, workflowID, version)
}

// RestoreWorkflowVersion restores a workflow to a specific version
func (a *App) RestoreWorkflowVersion(workflowID string, version int) error {
	return a.core.GetWorkflowManager().RestoreWorkflowVersion(a.ctx, workflowID, version)
}

// GetWorkflowExecutionResult gets the execution result of a workflow
func (a *App) GetWorkflowExecutionResult(executionID string) (*agent.WorkflowExecutionResult, error) {
	return a.core.GetWorkflowManager().GetWorkflowExecutionResult(a.ctx, executionID)
}

// GetWorkflowExecutionResults gets all execution results for a workflow
func (a *App) GetWorkflowExecutionResults(workflowID string) ([]*agent.WorkflowExecutionResult, error) {
	return a.core.GetWorkflowManager().GetWorkflowExecutionResults(a.ctx, workflowID)
}

// ===== 技能管理 API =====

// GetSkills returns all skills
func (a *App) GetSkills() (map[string]agent.SkillMetadata, error) {
	skills := a.core.GetSkillLibrary().GetAllSkills(a.ctx)
	result := make(map[string]agent.SkillMetadata)

	for name, skill := range skills {
		metadata := skill.GetMetadata(a.ctx)
		result[name] = metadata
	}

	return result, nil
}

// GetSkill returns a skill by name
func (a *App) GetSkill(name string) (agent.SkillMetadata, error) {
	skill, found := a.core.GetSkillLibrary().GetSkill(a.ctx, name)
	if !found {
		return agent.SkillMetadata{}, fmt.Errorf("skill not found: %s", name)
	}

	return skill.GetMetadata(a.ctx), nil
}

// DeleteSkill deletes a skill
func (a *App) DeleteSkill(name string) error {
	return a.core.GetSkillLibrary().UnregisterSkill(a.ctx, name)
}

// ===== 配置管理 API =====

// GetConfig returns the current configuration
func (a *App) GetConfig() (*agent.HostConfig, error) {
	return a.core.GetHost().Config(), nil
}

// UpdateConfig updates the configuration
func (a *App) UpdateConfig(config *agent.HostConfig) error {
	// For now, we'll just return nil since Host doesn't have an UpdateConfig method yet
	// This will be implemented in a future update
	return nil
}

// ReloadConfig reloads the configuration
func (a *App) ReloadConfig() error {
	// For now, we'll just return nil since Host doesn't have a ReloadConfig method yet
	// This will be implemented in a future update
	return nil
}

// ===== 文件系统 API =====

// ListFiles lists files in a directory
func (a *App) ListFiles(path string) ([]*agent.FileInfo, error) {
	return a.core.GetFileExplorer().ListFiles(a.ctx, path)
}

// CreateFile creates a new file
func (a *App) CreateFile(path string, content string) error {
	return a.core.GetFileExplorer().CreateFile(a.ctx, path, content)
}

// ReadFile reads a file's content
func (a *App) ReadFile(path string) (string, error) {
	return a.core.GetFileExplorer().ReadFile(a.ctx, path)
}

// WriteFile writes to a file
func (a *App) WriteFile(path string, content string) error {
	return a.core.GetFileExplorer().WriteFile(a.ctx, path, content)
}

// DeleteFile deletes a file
func (a *App) DeleteFile(path string) error {
	return a.core.GetFileExplorer().DeleteFile(a.ctx, path)
}

// CreateDirectory creates a new directory
func (a *App) CreateDirectory(path string) error {
	return a.core.GetFileExplorer().CreateDirectory(a.ctx, path)
}

// DeleteDirectory deletes a directory
func (a *App) DeleteDirectory(path string) error {
	return a.core.GetFileExplorer().DeleteDirectory(a.ctx, path)
}

// MoveFile moves a file or directory
func (a *App) MoveFile(src string, dst string) error {
	return a.core.GetFileExplorer().MoveFile(a.ctx, src, dst)
}

// CopyFile copies a file or directory
func (a *App) CopyFile(src string, dst string) error {
	return a.core.GetFileExplorer().CopyFile(a.ctx, src, dst)
}

// GetFileInfo returns information about a file or directory
func (a *App) GetFileInfo(path string) (*agent.FileInfo, error) {
	return a.core.GetFileExplorer().GetFileInfo(a.ctx, path)
}

// UploadFile uploads a file to the specified path
func (a *App) UploadFile(path string, content []byte) error {
	return a.core.GetFileExplorer().UploadFile(a.ctx, path, content)
}

// DownloadFile downloads a file from the specified path
func (a *App) DownloadFile(path string) ([]byte, error) {
	return a.core.GetFileExplorer().DownloadFile(a.ctx, path)
}

// ===== 增强技能系统 API =====

// SkillSystemInfo contains information about the skill system
type SkillSystemInfo struct {
	Initialized bool   `json:"initialized"`
	BaseDir     string `json:"baseDir"`
	TotalSkills int    `json:"totalSkills"`
}

// GetSkillSystemInfo returns basic information about the skill system
func (a *App) GetSkillSystemInfo() (*SkillSystemInfo, error) {
	if a.core.GetSkillSystem() == nil {
		return &SkillSystemInfo{Initialized: false}, nil
	}

	// Get total skill count
	entries := a.core.GetSkillSystem().Registry().ListAll()

	return &SkillSystemInfo{
		Initialized: true,
		BaseDir:     ".skills",
		TotalSkills: len(entries),
	}, nil
}

// SkillListItem represents a skill in the list
type SkillListItem struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	Tags        []string `json:"tags"`
	Version     string   `json:"version"`
	Enabled     bool     `json:"enabled"`
	UseCount    int64    `json:"useCount"`
	LastUsed    string   `json:"lastUsed"`
}

// ListRegisteredSkills lists all registered skills
func (a *App) ListRegisteredSkills() ([]*SkillListItem, error) {
	if a.core.GetSkillSystem() == nil {
		return nil, fmt.Errorf("skill system not initialized")
	}

	entries := a.core.GetSkillSystem().Registry().ListAll()
	result := make([]*SkillListItem, 0, len(entries))

	for _, entry := range entries {
		lastUsed := ""
		if !entry.LastUsed.IsZero() {
			lastUsed = entry.LastUsed.Format("2006-01-02 15:04:05")
		}

		result = append(result, &SkillListItem{
			ID:          entry.ID,
			Name:        entry.Name,
			Description: entry.Description,
			Category:    entry.Category,
			Tags:        entry.Tags,
			Version:     entry.Version,
			Enabled:     true, // For now, all registered skills are enabled
			UseCount:    entry.UsedCount,
			LastUsed:    lastUsed,
		})
	}

	return result, nil
}

// SkillDefinitionInfo contains information about a skill definition
type SkillDefinitionInfo struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Version     string                 `json:"version"`
	Category    string                 `json:"category"`
	Author      string                 `json:"author"`
	License     string                 `json:"license"`
	Workflow    map[string]interface{} `json:"workflow"`
	Config      map[string]interface{} `json:"config"`
}

// ListSkillDefinitions lists all skill definitions
func (a *App) ListSkillDefinitions() ([]*SkillDefinitionInfo, error) {
	if a.core.GetSkillSystem() == nil {
		return nil, fmt.Errorf("skill system not initialized")
	}

	definitions := a.core.GetSkillSystem().DefinitionManager().List()
	result := make([]*SkillDefinitionInfo, 0, len(definitions))

	for _, def := range definitions {
		// Convert workflow to map
		workflow := make(map[string]interface{})
		for _, step := range def.Workflow {
			workflow[step.ID] = map[string]interface{}{
				"name":      step.Name,
				"action":    step.Action,
				"timeout":   step.Timeout.String(),
			}
		}

		result = append(result, &SkillDefinitionInfo{
			ID:          def.ID,
			Name:        def.Name,
			Description: def.Description,
			Version:     def.Version,
			Category:    def.Category,
			Author:      def.Author,
			License:     def.License,
			Workflow:    workflow,
			Config: map[string]interface{}{
				"cache_enabled":   def.Config.EnableCache,
				"cache_ttl":       def.Config.CacheTTL.String(),
				"max_exec_time":   def.Config.MaxExecutionTime.String(),
				"enable_rate_limit": def.Config.EnableRateLimit,
				"rate_limit":      def.Config.RateLimit,
			},
		})
	}

	return result, nil
}

// GetSkillDefinition returns a skill definition by ID
func (a *App) GetSkillDefinition(skillID string) (*SkillDefinitionInfo, error) {
	if a.core.GetSkillSystem() == nil {
		return nil, fmt.Errorf("skill system not initialized")
	}

	definition, err := a.core.GetSkillSystem().DefinitionManager().Load(skillID)
	if err != nil {
		return nil, fmt.Errorf("failed to load skill definition: %w", err)
	}

	// Convert workflow to map
	workflow := make(map[string]interface{})
	for _, step := range definition.Workflow {
		stepData := map[string]interface{}{
			"name":   step.Name,
			"action": step.Action,
		}
		if step.Timeout > 0 {
			stepData["timeout"] = step.Timeout.String()
		}
		if step.SkipIf != "" {
			stepData["skip_if"] = step.SkipIf
		}
		workflow[step.ID] = stepData
	}

	return &SkillDefinitionInfo{
		ID:          definition.ID,
		Name:        definition.Name,
		Description: definition.Description,
		Version:     definition.Version,
		Category:    definition.Category,
		Author:      definition.Author,
		License:     definition.License,
		Workflow:    workflow,
		Config: map[string]interface{}{
			"cache_enabled":    definition.Config.EnableCache,
			"cache_ttl":        definition.Config.CacheTTL.String(),
			"max_exec_time":    definition.Config.MaxExecutionTime.String(),
			"enable_rate_limit": definition.Config.EnableRateLimit,
			"rate_limit":       definition.Config.RateLimit,
		},
	}, nil
}

// ExecuteSkillInput represents the input for executing a skill
type ExecuteSkillInput struct {
	SkillName  string                 `json:"skillName"`
	Input      string                 `json:"input"`
	Workspace  string                 `json:"workspace"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

// ExecuteSkillOutput represents the output from executing a skill
type ExecuteSkillOutput struct {
	Success bool        `json:"success"`
	Result  interface{} `json:"result,omitempty"`
	Error   string      `json:"error,omitempty"`
	Stats   map[string]interface{} `json:"stats,omitempty"`
}

// ExecuteSkillByName executes a skill by name
func (a *App) ExecuteSkillByName(input *ExecuteSkillInput) (*ExecuteSkillOutput, error) {
	if a.core.GetSkillSystem() == nil {
		return &ExecuteSkillOutput{
			Success: false,
			Error:   "skill system not initialized",
		}, nil
	}

	// Create execution context
	execCtx := skills.NewExecutionContext()
	if input.Workspace != "" {
		execCtx.Workspace = input.Workspace
	} else {
		execCtx.Workspace = "/workspace"
	}

	// Add parameters to context
	if len(input.Parameters) > 0 {
		for key, value := range input.Parameters {
			execCtx.SetMetadata(key, value)
		}
	}

	// Execute the skill
	result, err := a.core.GetSkillSystem().ExecuteSkill(a.ctx, input.SkillName, input.Input, execCtx)
	if err != nil {
		return &ExecuteSkillOutput{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &ExecuteSkillOutput{
		Success: true,
		Result:  result,
		Stats: map[string]interface{}{
			"skill_name": input.SkillName,
			"workspace":  execCtx.Workspace,
		},
	}, nil
}

// SkillSystemStats contains statistics about the skill system
type SkillSystemStats struct {
	TotalSkills    int64                  `json:"totalSkills"`
	TotalUses      int64                  `json:"totalUses"`
	Categories     map[string]int         `json:"categories"`
	MostUsedSkills []map[string]interface{} `json:"mostUsedSkills"`
}

// GetSkillSystemStats returns statistics about the skill system
func (a *App) GetSkillSystemStats() (*SkillSystemStats, error) {
	if a.core.GetSkillSystem() == nil {
		return nil, fmt.Errorf("skill system not initialized")
	}

	stats := a.core.GetSkillSystem().Registry().GetStats()
	totalSkills := int64(0)
	totalUses := int64(0)

	if val, ok := stats["total_skills"]; ok {
		totalSkills = val.(int64)
	}
	if val, ok := stats["total_uses"]; ok {
		totalUses = val.(int64)
	}

	// Get most used skills
	mostUsed := []map[string]interface{}{}
	if val, ok := stats["most_used"].([]map[string]interface{}); ok {
		mostUsed = val
	}

	return &SkillSystemStats{
		TotalSkills:    totalSkills,
		TotalUses:      totalUses,
		Categories:     make(map[string]int),
		MostUsedSkills: mostUsed,
	}, nil
}

// CacheStats contains statistics about the cache
type CacheStats struct {
	HitRate      float64 `json:"hitRate"`
	Hits         int64   `json:"hits"`
	Misses       int64   `json:"misses"`
	TotalEntries int     `json:"totalEntries"`
	Size         int64   `json:"size"`
}

// GetCacheStats returns cache statistics from the skill system
func (a *App) GetCacheStats() (*CacheStats, error) {
	if a.core.GetSkillSystem() == nil {
		return nil, fmt.Errorf("skill system not initialized")
	}

	// Get stats from the skill registry
	statsMap := a.core.GetSkillSystem().Registry().GetStats()

	// Extract statistics from map
	stats := &CacheStats{}

	// Extract hits
	if hits, ok := statsMap["hits"].(int64); ok {
		stats.Hits = hits
	} else if hits, ok := statsMap["hits"].(int); ok {
		stats.Hits = int64(hits)
	}

	// Extract misses
	if misses, ok := statsMap["misses"].(int64); ok {
		stats.Misses = misses
	} else if misses, ok := statsMap["misses"].(int); ok {
		stats.Misses = int64(misses)
	}

	// Extract hit rate
	if hitRate, ok := statsMap["hit_rate"].(float64); ok {
		stats.HitRate = hitRate
	} else if hitRate, ok := statsMap["hit_rate"].(float32); ok {
		stats.HitRate = float64(hitRate)
	}

	// Extract total entries (sum of l1 and l2 entries)
	if l1Entries, ok := statsMap["l1_entries"].(int); ok {
		if l2Entries, ok2 := statsMap["l2_entries"].(int); ok2 {
			stats.TotalEntries = l1Entries + l2Entries
		} else {
			stats.TotalEntries = l1Entries
		}
	}

	// Extract total size
	if size, ok := statsMap["total_size"].(int64); ok {
		stats.Size = size
	} else if size, ok := statsMap["total_size"].(int); ok {
		stats.Size = int64(size)
	}

	return stats, nil
}

// PoolStats contains statistics about the connection pool
type PoolStats struct {
	ActiveConnections int     `json:"activeConnections"`
	IdleConnections   int     `json:"idleConnections"`
	MaxConnections    int     `json:"maxConnections"`
	MinConnections    int     `json:"minConnections"`
	UtilizationRate   float64 `json:"utilizationRate"`
}

// GetPoolStats returns pool statistics from the skill system
func (a *App) GetPoolStats() (*PoolStats, error) {
	if a.core.GetSkillSystem() == nil {
		return nil, fmt.Errorf("skill system not initialized")
	}

	// Get stats from the skill registry
	statsMap := a.core.GetSkillSystem().Registry().GetStats()

	stats := &PoolStats{}

	// Extract pool statistics if available
	if activeConns, ok := statsMap["current_active"].(int32); ok {
		stats.ActiveConnections = int(activeConns)
	} else if activeConns, ok := statsMap["current_active"].(int); ok {
		stats.ActiveConnections = activeConns
	}

	if idleConns, ok := statsMap["current_idle"].(int32); ok {
		stats.IdleConnections = int(idleConns)
	} else if idleConns, ok := statsMap["current_idle"].(int); ok {
		stats.IdleConnections = idleConns
	}

	// Get configuration for max/min connections
	if config, ok := statsMap["config"].(map[string]interface{}); ok {
		if maxConns, ok := config["max_connections"].(int32); ok {
			stats.MaxConnections = int(maxConns)
		} else if maxConns, ok := config["max_connections"].(int); ok {
			stats.MaxConnections = maxConns
		}

		if minConns, ok := config["min_connections"].(int32); ok {
			stats.MinConnections = int(minConns)
		} else if minConns, ok := config["min_connections"].(int); ok {
			stats.MinConnections = minConns
		}
	}

	// Calculate utilization rate
	if stats.MaxConnections > 0 {
		stats.UtilizationRate = float64(stats.ActiveConnections) / float64(stats.MaxConnections) * 100
	}

	// If no pool stats available, return default values
	if stats.MaxConnections == 0 {
		stats.MaxConnections = 10
		stats.MinConnections = 2
		stats.UtilizationRate = 0.0
	}

	return stats, nil
}

// ReloadSkillDefinitions reloads skill definitions
func (a *App) ReloadSkillDefinitions() error {
	if a.core.GetSkillSystem() == nil {
		return fmt.Errorf("skill system not initialized")
	}

	// This would need to be implemented in the ProgressiveLoader
	// For now, just return nil
	return nil
}

// ClearCache clears the skill cache
func (a *App) ClearCache() error {
	if a.core.GetSkillSystem() == nil {
		return fmt.Errorf("skill system not initialized")
	}

	// This would need to be implemented in the cache system
	// For now, just return nil
	return nil
}

// EnableSkill enables a skill
func (a *App) EnableSkill(skillID string) error {
	if a.core.GetSkillSystem() == nil {
		return fmt.Errorf("skill system not initialized")
	}

	return a.core.GetSkillSystem().Registry().EnableSkill(skillID)
}

// DisableSkill disables a skill
func (a *App) DisableSkill(skillID string) error {
	if a.core.GetSkillSystem() == nil {
		return fmt.Errorf("skill system not initialized")
	}

	return a.core.GetSkillSystem().Registry().DisableSkill(skillID)
}

// ToggleSkill toggles a skill's enabled status
func (a *App) ToggleSkill(skillID string) (bool, error) {
	if a.core.GetSkillSystem() == nil {
		return false, fmt.Errorf("skill system not initialized")
	}

	return a.core.GetSkillSystem().Registry().ToggleSkill(skillID)
}

// ===== 技能导入 API =====

// ImportSkillInput represents the input for importing a skill
type ImportSkillInput struct {
	SourceType string `json:"sourceType"` // "file", "url", "paste", "git"
	Data       []byte `json:"data,omitempty"`
	URL        string `json:"url,omitempty"`
	Content    string `json:"content,omitempty"`
	AuthToken  string `json:"authToken,omitempty"`
	Branch     string `json:"branch,omitempty"`
	Path       string `json:"path,omitempty"`
}

// ImportSkillOptions represents options for importing a skill
type ImportSkillOptions struct {
	SkillID    string `json:"skillId,omitempty"`
	Overwrite  bool   `json:"overwrite"`
	AutoEnable bool   `json:"autoEnable"`
	Validate   bool   `json:"validate"`
	Workspace  string `json:"workspace,omitempty"`
}

// ImportSkillResult represents the result of importing a skill
type ImportSkillResult struct {
	Success   bool   `json:"success"`
	SkillID   string `json:"skillId"`
	SkillName string `json:"skillName"`
	Message   string `json:"message"`
	Warnings  []string `json:"warnings,omitempty"`
}

// ImportSkillFromFile imports a skill from a file
func (a *App) ImportSkillFromFile(data []byte, format string, options ImportSkillOptions) (*ImportSkillResult, error) {
	if a.core.GetSkillSystem() == nil {
		return nil, fmt.Errorf("skill system not initialized")
	}

	baseDir := a.getSkillSystemBaseDir()
	importer := skills.NewSkillImporter(
		a.core.GetSkillSystem().Registry(),
		a.core.GetSkillSystem().DefinitionManager(),
		baseDir,
	)

	result, err := importer.ImportFromArchive(a.ctx, data, format, skills.ImportOptions{
		SkillID:    options.SkillID,
		Overwrite:  options.Overwrite,
		AutoEnable: options.AutoEnable,
		Validate:   options.Validate,
		Workspace:  options.Workspace,
	})

	if err != nil {
		return &ImportSkillResult{
			Success: false,
			Message: err.Error(),
		}, err
	}

	return &ImportSkillResult{
		Success:   result.Success,
		SkillID:   result.SkillID,
		SkillName: result.SkillName,
		Message:   result.Message,
		Warnings:  result.Warnings,
	}, nil
}

// ImportSkillFromURL imports a skill from a URL
func (a *App) ImportSkillFromURL(url string, options ImportSkillOptions) (*ImportSkillResult, error) {
	if a.core.GetSkillSystem() == nil {
		return nil, fmt.Errorf("skill system not initialized")
	}

	baseDir := a.getSkillSystemBaseDir()
	importer := skills.NewSkillImporter(
		a.core.GetSkillSystem().Registry(),
		a.core.GetSkillSystem().DefinitionManager(),
		baseDir,
	)

	result, err := importer.ImportFromURL(a.ctx, url, skills.ImportOptions{
		SkillID:    options.SkillID,
		Overwrite:  options.Overwrite,
		AutoEnable: options.AutoEnable,
		Validate:   options.Validate,
		Workspace:  options.Workspace,
	})

	if err != nil {
		return &ImportSkillResult{
			Success: false,
			Message: err.Error(),
		}, err
	}

	return &ImportSkillResult{
		Success:   result.Success,
		SkillID:   result.SkillID,
		SkillName: result.SkillName,
		Message:   result.Message,
		Warnings:  result.Warnings,
	}, nil
}

// ImportSkillFromContent imports a skill from content string
func (a *App) ImportSkillFromContent(content string, options ImportSkillOptions) (*ImportSkillResult, error) {
	if a.core.GetSkillSystem() == nil {
		return nil, fmt.Errorf("skill system not initialized")
	}

	baseDir := a.getSkillSystemBaseDir()
	importer := skills.NewSkillImporter(
		a.core.GetSkillSystem().Registry(),
		a.core.GetSkillSystem().DefinitionManager(),
		baseDir,
	)

	result, err := importer.ImportFromContent(a.ctx, content, skills.ImportOptions{
		SkillID:    options.SkillID,
		Overwrite:  options.Overwrite,
		AutoEnable: options.AutoEnable,
		Validate:   options.Validate,
		Workspace:  options.Workspace,
	})

	if err != nil {
		return &ImportSkillResult{
			Success: false,
			Message: err.Error(),
		}, err
	}

	return &ImportSkillResult{
		Success:   result.Success,
		SkillID:   result.SkillID,
		SkillName: result.SkillName,
		Message:   result.Message,
		Warnings:  result.Warnings,
	}, nil
}

// ValidateSkillFile validates a skill file without importing it
func (a *App) ValidateSkillFile(content string) (map[string]interface{}, error) {
	if a.core.GetSkillSystem() == nil {
		return nil, fmt.Errorf("skill system not initialized")
	}

	baseDir := a.getSkillSystemBaseDir()
	importer := skills.NewSkillImporter(
		a.core.GetSkillSystem().Registry(),
		a.core.GetSkillSystem().DefinitionManager(),
		baseDir,
	)

	metadata, err := importer.ValidateSkillFile(content)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"name":        metadata.Name,
		"description": metadata.Description,
		"version":     metadata.Version,
		"category":    metadata.Category,
		"author":      metadata.Author,
		"tags":        metadata.Tags,
	}, nil
}

