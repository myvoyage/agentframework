// Agent Framework - Enhanced Core Application with Security, Performance, and Quality improvements
// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package core

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/cloudwego/eino/components/tool"

	"AgentFramework/agent"
	"AgentFramework/internal/auth"
	"AgentFramework/pkg/cache"
	"AgentFramework/pkg/validation"
	"AgentFramework/pkg/errors"
	"AgentFramework/pkg/lockfree"
	"AgentFramework/pkg/pool"
	"AgentFramework/pkg/rbac"
)

// EnhancedApplication represents the enhanced core application with security, performance, and quality improvements
type EnhancedApplication struct {
	ctx                context.Context
	host               *agent.Host
	skillLibrary       agent.SkillLibrary
	skillSystem        *agent.SkillSystem
	fileExplorer       *agent.FileExplorer
	eventBus           agent.EventBus
	workflowManager    *agent.WorkflowManager
	config             *agent.HostConfig

	// Security enhancements
	jwtValidator       *auth.JWTValidator
	rbacManager        *rbac.RBACManager
	validator          *validation.InputValidator
	errHandler         *errors.Handler

	// Performance enhancements
	messagePool        *pool.MessagePool
	eventPool          *pool.EventPool
	contextPool        *pool.ContextPool
	multiLevelCache     *cache.MultiLevelCache
	agentRegistry       *lockfree.AgentRegistry
	metrics             *lockfree.Metrics

	// Thread safety
	mu                  sync.RWMutex
}

// NewEnhancedApplication creates a new enhanced application with all optimizations
func NewEnhancedApplication(ctx context.Context, cfg *agent.HostConfig, modelFactory agent.ModelFactory, toolRegistry map[string]tool.BaseTool) (*EnhancedApplication, error) {
	// Initialize OpenTelemetry
	tp, err := InitOpenTelemetry(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize OpenTelemetry: %w", err)
	}
	defer tp.Shutdown(ctx)

	// Create Host instance
	host, err := agent.NewHost(ctx, cfg, modelFactory, toolRegistry)
	if err != nil {
		return nil, fmt.Errorf("failed to create host: %w", err)
	}

	// Create skill library and register built-in skills
	skillLibrary := agent.NewSkillLibrary()

	// Register built-in skills
	builtinSkills := []agent.Skill{
		agent.NewHTTPRequestSkill(),
		agent.NewFileOperationSkill(),
		agent.NewCodeExecutionSkill(),
		agent.NewDataProcessingSkill(),
	}

	for _, skill := range builtinSkills {
		skillLibrary.RegisterSkill(ctx, skill)
	}

	// Create workflow manager
	workflowManager := agent.NewWorkflowManager(skillLibrary, modelFactory)

	// Initialize skill system
	var skillSystem *agent.SkillSystem
	if cfg.SkillSystemDir != "" {
		skillSystem, err = agent.NewSkillSystem(cfg.SkillSystemDir)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize skill system: %w", err)
		}
	}

	// ===== Security Enhancements =====
	// Initialize JWT validator
	secretKey := "default-secret-key-change-in-production"
	log.Println("Warning: Using default JWT secret key, please configure JWT properly in production")
	jwtValidator := auth.NewJWTValidator(secretKey, "HS256")

	// Initialize RBAC manager
	rbacManager := rbac.NewRBACManager()
	rbacManager.InitializeDefaultRoles()

	// Assign default admin role if configured
	// TODO: Load admin user ID from configuration
	adminUserID := ""
	if adminUserID != "" {
		rbacManager.AssignRole(adminUserID, "admin")
	}

	// Initialize input validator
	validator := validation.RequiredStringValidation(1000)

	// Initialize error handler with logger
	errHandler := errors.NewHandler("core", nil)

	// ===== Performance Enhancements =====
	// Initialize object pools
	messagePool := pool.DefaultMessagePool
	eventPool := pool.DefaultEventPool
	contextPool := pool.DefaultContextPool

	// Initialize multi-level cache
	// L1: In-memory cache
	l1Cache := cache.NewInMemoryCache()
	// L2: Redis cache (if configured)
	var l2Cache cache.Cache
	// TODO: Check Redis configuration from environment or config file
	redisEnabled := false
	if redisEnabled {
		// TODO: Initialize Redis cache
		l2Cache = nil // Placeholder
	}
	// L3: Persistent storage (using file storage for now)
	l3Cache := cache.NewInMemoryCache() // Placeholder

	multiLevelCache := cache.NewMultiLevelCache(
		l1Cache,
		l2Cache,
		l3Cache,
		5*time.Minute,
		1*time.Hour,
	)

	// Initialize lock-free structures
	agentRegistry := lockfree.NewAgentRegistry()
	metrics := lockfree.NewMetrics()

	return &EnhancedApplication{
		ctx:             ctx,
		host:            host,
		skillLibrary:    skillLibrary,
		skillSystem:     skillSystem,
		fileExplorer:    agent.NewFileExplorer(),
		eventBus:        agent.NewMemoryEventBus(),
		workflowManager: workflowManager,
		config:          cfg,

		// Security
		jwtValidator:  jwtValidator,
		rbacManager:   rbacManager,
		validator:     validator,
		errHandler:    errHandler,

		// Performance
		messagePool:    messagePool,
		eventPool:      eventPool,
		contextPool:    contextPool,
		multiLevelCache: multiLevelCache,
		agentRegistry:  agentRegistry,
		metrics:        metrics,
	}, nil
}

// Initialize initializes the enhanced application with all security and performance features
func (app *EnhancedApplication) Initialize(ctx context.Context) error {
	// Initialize core components
	app.workflowManager.Init(ctx)
	app.fileExplorer.Init(ctx)

	// Initialize security systems
	if err := app.initializeSecurity(ctx); err != nil {
		return err
	}

	// Initialize performance systems
	if err := app.initializePerformance(ctx); err != nil {
		return err
	}

	log.Println("✅ Enhanced application initialized successfully")
	return nil
}

// initializeSecurity initializes all security components
func (app *EnhancedApplication) initializeSecurity(ctx context.Context) error {
	// Assign default roles to system users
	defaultUsers := map[string]string{
		"system": "admin",
		"admin":  "admin",
	}

	for userID, roleName := range defaultUsers {
		if err := app.rbacManager.AssignRole(userID, roleName); err != nil {
			log.Printf("Warning: Failed to assign role %s to user %s: %v", roleName, userID, err)
		}
	}

	log.Println("✅ Security components initialized")
	return nil
}

// initializePerformance initializes all performance components
func (app *EnhancedApplication) initializePerformance(ctx context.Context) error {
	// Pre-warm object pools
	for i := 0; i < 100; i++ {
		msg := app.messagePool.Get()
		app.messagePool.Put(msg)
	}

	evt := app.eventPool.Get()
	app.eventPool.Put(evt)

	ctxObj := app.contextPool.Get()
	app.contextPool.Put(ctxObj)

	log.Println("✅ Performance components initialized")
	return nil
}

// ===== Enhanced Security Methods =====

// ValidateAndSanitizeInput validates and sanitizes user input
func (app *EnhancedApplication) ValidateAndSanitizeInput(input string) (string, error) {
	return app.validator.ValidateAndSanitize(input)
}

// ValidateJWT validates a JWT token and returns the subject
func (app *EnhancedApplication) ValidateJWT(token string) (string, error) {
	// TODO: Store secret key in EnhancedApplication for proper JWT validation
	// For now, use a default implementation
	secretKey := "default-secret-key-change-in-production"
	subject, err := auth.ValidateJWTWithSecret(token, "", secretKey, "HS256")
	if err != nil {
		return "", app.errHandler.Wrap(err, "JWT validation failed")
	}
	return subject, nil
}

// CheckPermission checks if a user has permission to perform an action
func (app *EnhancedApplication) CheckPermission(userID, resource, action string) bool {
	return app.rbacManager.CheckPermission(userID, resource, action)
}

// RequirePermission checks permission and returns an error if denied
func (app *EnhancedApplication) RequirePermission(userID, resource, action string) error {
	if !app.CheckPermission(userID, resource, action) {
		return app.errHandler.Forbidden(fmt.Sprintf("cannot %s %s", action, resource))
	}
	return nil
}

// AssignRole assigns a role to a user
func (app *EnhancedApplication) AssignRole(userID, roleName string) error {
	return app.rbacManager.AssignRole(userID, roleName)
}

// ===== Enhanced Performance Methods =====

// GetFromCache retrieves a value from the multi-level cache
func (app *EnhancedApplication) GetFromCache(key string) (interface{}, error) {
	return app.multiLevelCache.Get(app.ctx, key)
}

// SetInCache stores a value in the multi-level cache
func (app *EnhancedApplication) SetInCache(key string, value interface{}, ttl time.Duration) error {
	return app.multiLevelCache.Set(app.ctx, key, value, ttl)
}

// RegisterAgent registers an agent in the lock-free registry
func (app *EnhancedApplication) RegisterAgent(id string, agent interface{}) {
	app.agentRegistry.Register(id, agent)
}

// GetAgent retrieves an agent from the registry
func (app *EnhancedApplication) GetAgent(id string) (interface{}, bool) {
	return app.agentRegistry.Get(id)
}

// IncrementRequestCount atomically increments the request counter
func (app *EnhancedApplication) IncrementRequestCount() {
	app.metrics.IncrementRequestCount()
}

// IncrementErrorCount atomically increments the error counter
func (app *EnhancedApplication) IncrementErrorCount() {
	app.metrics.IncrementErrorCount()
}

// AddLatency records latency for performance monitoring
func (app *EnhancedApplication) AddLatency(latency uint64) {
	app.metrics.AddLatency(latency)
}

// ===== Resource Management =====

// AcquireMessage gets a message from the pool
func (app *EnhancedApplication) AcquireMessage() *pool.PooledMessage {
	return pool.NewPooledMessage(app.messagePool)
}

// AcquireEvent gets an event from the pool
func (app *EnhancedApplication) AcquireEvent() *pool.PooledEvent {
	return pool.NewPooledEvent(app.eventPool)
}

// AcquireContext gets a context from the pool
func (app *EnhancedApplication) AcquireContext() *pool.PooledContext {
	return pool.NewPooledContext(app.contextPool)
}

// ===== Getters (maintain backward compatibility) =====

func (app *EnhancedApplication) GetContext() context.Context {
	return app.ctx
}

func (app *EnhancedApplication) GetHost() *agent.Host {
	return app.host
}

func (app *EnhancedApplication) GetSkillLibrary() agent.SkillLibrary {
	return app.skillLibrary
}

func (app *EnhancedApplication) GetSkillSystem() *agent.SkillSystem {
	return app.skillSystem
}

func (app *EnhancedApplication) GetFileExplorer() *agent.FileExplorer {
	return app.fileExplorer
}

func (app *EnhancedApplication) GetWorkflowManager() *agent.WorkflowManager {
	return app.workflowManager
}

func (app *EnhancedApplication) GetConfig() *agent.HostConfig {
	return app.config
}

// GetJWTValidator returns the JWT validator
func (app *EnhancedApplication) GetJWTValidator() *auth.JWTValidator {
	return app.jwtValidator
}

// GetRBACManager returns the RBAC manager
func (app *EnhancedApplication) GetRBACManager() *rbac.RBACManager {
	return app.rbacManager
}

// GetInputValidator returns the input validator
func (app *EnhancedApplication) GetInputValidator() *validation.InputValidator {
	return app.validator
}

// GetErrorHandler returns the error handler
func (app *EnhancedApplication) GetErrorHandler() *errors.Handler {
	return app.errHandler
}

// GetMessagePool returns the message pool
func (app *EnhancedApplication) GetMessagePool() *pool.MessagePool {
	return app.messagePool
}

// GetEventPool returns the event pool
func (app *EnhancedApplication) GetEventPool() *pool.EventPool {
	return app.eventPool
}

// GetContextPool returns the context pool
func (app *EnhancedApplication) GetContextPool() *pool.ContextPool {
	return app.contextPool
}

// GetMultiLevelCache returns the multi-level cache
func (app *EnhancedApplication) GetMultiLevelCache() *cache.MultiLevelCache {
	return app.multiLevelCache
}

// GetAgentRegistry returns the agent registry
func (app *EnhancedApplication) GetAgentRegistry() *lockfree.AgentRegistry {
	return app.agentRegistry
}

// GetMetrics returns the metrics
func (app *EnhancedApplication) GetMetrics() *lockfree.Metrics {
	return app.metrics
}

// Cleanup performs cleanup of resources
func (app *EnhancedApplication) Cleanup(ctx context.Context) error {
	log.Println("🧹 Cleaning up enhanced application resources...")

	// Cleanup caches
	if err := app.multiLevelCache.Clear(ctx); err != nil {
		log.Printf("Warning: Failed to clear cache: %v", err)
	}

	// Print performance statistics
	metrics := pool.GetAllMetrics()
	log.Printf("📊 Pool Statistics:")
	log.Printf("   Messages: Allocated=%d, Pooled=%d, Reused=%d, Reuse Rate=%.2f%%",
		metrics.MessageAllocated,
		metrics.MessagePooled,
		metrics.MessageReused,
		metrics.ReusedRate,
	)

	// Print metrics
	log.Printf("📊 Performance Metrics:")
	log.Printf("   Requests: %d", app.metrics.GetRequestCount())
	log.Printf("   Errors: %d", app.metrics.GetErrorCount())
	if app.metrics.GetRequestCount() > 0 {
		log.Printf("   Avg Latency: %d units", app.metrics.GetAverageLatency())
	}

	log.Println("✅ Cleanup completed")
	return nil
}
