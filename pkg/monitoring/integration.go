// Agent Framework - Monitoring Integration Example
// Copyright (C) 2025  Agent Framework Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package monitoring

import (
	"context"
	"database/sql"
	"time"

	"AgentFramework/pkg/health"
)

// RegisterDefaultHealthChecks registers default health checks for common components
func RegisterDefaultHealthChecks(checker *health.Checker, components map[string]interface{}) {
	// Database health check
	if db, ok := components["database"].(*sql.DB); ok {
		checker.Register("database", func(ctx context.Context) error {
			return db.PingContext(ctx)
		})
	}

	// Redis health check (if Redis client is available)
	if redisClient, ok := components["redis"].(interface {
		Ping(context.Context) error
	}); ok {
		checker.Register("redis", func(ctx context.Context) error {
			return redisClient.Ping(ctx)
		})
	}

	// System health check (always available)
	checker.Register("system", func(ctx context.Context) error {
		// Basic system check - always healthy
		return nil
	})
}

// SetupMonitoring sets up monitoring with default configuration
func SetupMonitoring(ctx context.Context, config *MonitoringConfig, healthChecker *health.Checker) (*MonitoringManager, *Server, error) {
	// Create monitoring manager
	manager, err := NewMonitoringManager(config)
	if err != nil {
		return nil, nil, err
	}

	// Start system metrics collection
	systemMetrics := manager.GetSystemMetrics(15 * time.Second)
	systemMetrics.Start(ctx)

	// Create health checker if not provided
	if healthChecker == nil {
		healthChecker = health.New()
	}

	// Create and start monitoring server
	port := config.Port
	if port == 0 {
		port = 9090
	}

	server := NewServer(manager, healthChecker, port)
	if err := server.Start(ctx); err != nil {
		return nil, nil, err
	}

	return manager, server, nil
}
