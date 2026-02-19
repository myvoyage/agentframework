// Agent Framework - Monitoring HTTP Server
// Copyright (C) 2025  Agent Framework Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"AgentFramework/pkg/health"
)

// Server provides HTTP endpoints for monitoring
type Server struct {
	manager     *MonitoringManager
	healthChecker *health.Checker
	server     *http.Server
	mu         sync.RWMutex
	port       int
	started    bool
}

// NewServer creates a new monitoring HTTP server
func NewServer(manager *MonitoringManager, healthChecker *health.Checker, port int) *Server {
	if port == 0 {
		port = 9090
	}

	return &Server{
		manager:      manager,
		healthChecker: healthChecker,
		port:         port,
		started:      false,
	}
}

// Start starts the monitoring HTTP server
func (s *Server) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.started {
		return fmt.Errorf("monitoring server already started")
	}

	mux := http.NewServeMux()

	// Prometheus metrics endpoint
	mux.Handle("/metrics", promhttp.HandlerFor(
		s.manager.GetRegistry(),
		promhttp.HandlerOpts{
			EnableOpenMetrics: true,
		},
	))

	// Health check endpoint
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/health/live", s.handleLiveness)
	mux.HandleFunc("/health/ready", s.handleReadiness)

	// Status endpoint
	mux.HandleFunc("/status", s.handleStatus)

	s.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.port),
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	s.started = true

	// Start server in goroutine
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Monitoring server error: %v\n", err)
		}
	}()

	return nil
}

// Stop stops the monitoring HTTP server
func (s *Server) Stop(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.started || s.server == nil {
		return nil
	}

	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := s.server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("failed to shutdown monitoring server: %w", err)
	}

	s.started = false
	return nil
}

// handleHealth handles the /health endpoint
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	status := s.healthChecker.GetStatus(ctx)
	overallStatus, ok := status["status"].(health.Status)
	if !ok {
		// Fallback if type assertion fails
		if statusStr, ok := status["status"].(string); ok {
			overallStatus = health.Status(statusStr)
		} else {
			overallStatus = health.StatusUnhealthy
		}
	}

	w.Header().Set("Content-Type", "application/json")

	if overallStatus == health.StatusHealthy {
		w.WriteHeader(http.StatusOK)
	} else if overallStatus == health.StatusDegraded {
		w.WriteHeader(http.StatusOK) // Still OK but degraded
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	json.NewEncoder(w).Encode(status)
}

// handleLiveness handles the /health/live endpoint (liveness probe)
func (s *Server) handleLiveness(w http.ResponseWriter, r *http.Request) {
	// Simple liveness check - just verify server is running
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "alive",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// handleReadiness handles the /health/ready endpoint (readiness probe)
func (s *Server) handleReadiness(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	status := s.healthChecker.GetStatus(ctx)
	overallStatus, ok := status["status"].(health.Status)
	if !ok {
		if statusStr, ok := status["status"].(string); ok {
			overallStatus = health.Status(statusStr)
		} else {
			overallStatus = health.StatusUnhealthy
		}
	}

	w.Header().Set("Content-Type", "application/json")

	if overallStatus == health.StatusHealthy || overallStatus == health.StatusDegraded {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	json.NewEncoder(w).Encode(status)
}

// handleStatus handles the /status endpoint
func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	healthStatus := s.healthChecker.GetStatus(ctx)
	metrics := s.manager.GetMetrics()

	status := map[string]interface{}{
		"health":  healthStatus,
		"metrics": map[string]interface{}{
			"agent_id":     s.manager.agentID,
			"version":      s.manager.version,
			"environment":  s.manager.environment,
		},
		"timestamp": time.Now().Format(time.RFC3339),
	}

	// Add basic metric counts if available
	if metrics != nil {
		status["metrics"].(map[string]interface{})["available"] = true
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(status)
}
