// Agent Framework - Core Application Layer
// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package core

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"

	"github.com/cloudwego/eino/components/tool"

	"AgentFramework/agent"
)

// Application represents the core application logic that can be shared
// between desktop and CLI interfaces
type Application struct {
	ctx             context.Context
	host            *agent.Host
	skillLibrary    agent.SkillLibrary
	skillSystem     *agent.SkillSystem
	fileExplorer    *agent.FileExplorer
	eventBus        agent.EventBus
	workflowManager *agent.WorkflowManager
	config          *agent.HostConfig
}

// NewApplication creates a new core application instance
// This follows the Dependency Injection principle (SOLID - D)
func NewApplication(ctx context.Context, config *agent.HostConfig, modelFactory agent.ModelFactory, toolRegistry map[string]tool.BaseTool) (*Application, error) {
	// Initialize OpenTelemetry
	tp, err := InitOpenTelemetry(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize OpenTelemetry: %w", err)
	}
	defer tp.Shutdown(ctx)

	// Create Host instance
	host, err := agent.NewHost(ctx, config, modelFactory, toolRegistry)
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
	if config.SkillSystemDir != "" {
		skillSystem, err = agent.NewSkillSystem(config.SkillSystemDir)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize skill system: %w", err)
		}
	}

	return &Application{
		ctx:             ctx,
		host:            host,
		skillLibrary:    skillLibrary,
		skillSystem:     skillSystem,
		fileExplorer:    agent.NewFileExplorer(),
		eventBus:        agent.NewMemoryEventBus(),
		workflowManager: workflowManager,
		config:          config,
	}, nil
}

// InitOpenTelemetry initializes OpenTelemetry tracing
func InitOpenTelemetry(ctx context.Context) (*trace.TracerProvider, error) {
	exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout trace exporter: %w", err)
	}

	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String("agentframework-core"),
		semconv.ServiceVersionKey.String("1.0.0"),
	)

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(res),
	)

	otel.SetTracerProvider(tp)

	return tp, nil
}

// Initialize initializes the application components
func (app *Application) Initialize(ctx context.Context) error {
	// Initialize workflow manager
	app.workflowManager.Init(ctx)

	// Initialize file explorer
	app.fileExplorer.Init(ctx)

	return nil
}

// GetContext returns the application context
func (app *Application) GetContext() context.Context {
	return app.ctx
}

// GetHost returns the agent host
func (app *Application) GetHost() *agent.Host {
	return app.host
}

// GetSkillLibrary returns the skill library
func (app *Application) GetSkillLibrary() agent.SkillLibrary {
	return app.skillLibrary
}

// GetSkillSystem returns the skill system
func (app *Application) GetSkillSystem() *agent.SkillSystem {
	return app.skillSystem
}

// GetFileExplorer returns the file explorer
func (app *Application) GetFileExplorer() *agent.FileExplorer {
	return app.fileExplorer
}

// GetEventBus returns the event bus
func (app *Application) GetEventBus() agent.EventBus {
	return app.eventBus
}

// GetWorkflowManager returns the workflow manager
func (app *Application) GetWorkflowManager() *agent.WorkflowManager {
	return app.workflowManager
}

// GetConfig returns the host configuration
func (app *Application) GetConfig() *agent.HostConfig {
	return app.config
}

// Shutdown gracefully shuts down the application
func (app *Application) Shutdown(ctx context.Context) error {
	// Cleanup resources
	if app.workflowManager != nil {
		// Add workflow manager cleanup if needed
	}

	if app.fileExplorer != nil {
		// Add file explorer cleanup if needed
	}

	return nil
}
