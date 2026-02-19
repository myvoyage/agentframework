// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later


// Agent Framework Heartbeat Service
package heartbeat

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type HeartbeatService interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	SendBeat(ctx context.Context) error
	RegisterTarget(target string, config TargetConfig) error
	UnregisterTarget(target string) error
	GetStatus(target string) (*TargetStatus, error)
	GetAllStatuses() map[string]*TargetStatus
}

type MemoryHeartbeatService struct {
	config  *HeartbeatConfig
	targets map[string]*TargetState
	mu      sync.RWMutex
	running bool
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	ticker  *time.Ticker
}

type TargetState struct {
	Name       string
	Config     TargetConfig
	Status     TargetStatus
	LastBeatAt time.Time
	FailCount  int
	IsHealthy  bool
}

type TargetConfig struct {
	Interval    time.Duration
	Timeout     time.Duration
	MaxFailures int
	Required    bool
	OnFailure   FailureHandler
	OnRecovery  RecoveryHandler
	Metadata    map[string]string
}

type TargetStatus struct {
	Name            string
	IsHealthy       bool
	LastBeatAt      time.Time
	FailCount       int
	ConsecutiveFails int
	Message         string
}

type FailureHandler  func(ctx context.Context, target string, status *TargetStatus)
type RecoveryHandler func(ctx context.Context, target string, status *TargetStatus)

type HeartbeatConfig struct {
	Interval      time.Duration
	Timeout       time.Duration
	Logger        Logger
	EnableMetrics bool
}

type HeartbeatStats struct {
	TotalTargets     int
	HealthyTargets   int
	UnhealthyTargets int
	TotalBeats       int64
	TotalFailures    int64
}

type Logger interface {
	Info(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Debug(msg string, args ...interface{})
}

type DefaultLogger struct{}

func (l *DefaultLogger) Info(msg string, args ...interface{}) { fmt.Printf("[INFO] %s %v\n", msg, args) }
func (l *DefaultLogger) Error(msg string, args ...interface{}) { fmt.Printf("[ERROR] %s %v\n", msg, args) }
func (l *DefaultLogger) Warn(msg string, args ...interface{}) { fmt.Printf("[WARN] %s %v\n", msg, args) }
func (l *DefaultLogger) Debug(msg string, args ...interface{}) { fmt.Printf("[DEBUG] %s %v\n", msg, args) }

func NewMemoryHeartbeatService(config *HeartbeatConfig) *MemoryHeartbeatService {
	if config == nil {
		config = &HeartbeatConfig{
			Interval: 30 * time.Second,
			Timeout:  10 * time.Second,
			Logger:   &DefaultLogger{},
		}
	}
	return &MemoryHeartbeatService{
		config:  config,
		targets: make(map[string]*TargetState),
	}
}

func (s *MemoryHeartbeatService) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.running {
		return fmt.Errorf("already running")
	}
	s.ctx, s.cancel = context.WithCancel(ctx)
	s.running = true
	s.ticker = time.NewTicker(s.config.Interval)
	s.wg.Add(1)
	go s.runLoop()
	s.config.Logger.Info("heartbeat service started")
	return nil
}

func (s *MemoryHeartbeatService) Stop(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.running {
		return nil
	}
	s.cancel()
	s.running = false
	if s.ticker != nil {
		s.ticker.Stop()
	}
	s.wg.Wait()
	s.config.Logger.Info("heartbeat service stopped")
	return nil
}

func (s *MemoryHeartbeatService) SendBeat(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	for _, target := range s.targets {
		target.LastBeatAt = now
		target.IsHealthy = true
		target.FailCount = 0
		target.Status = TargetStatus{
			Name:       target.Name,
			IsHealthy:  true,
			LastBeatAt: now,
			FailCount:  0,
			Message:    "OK",
		}
	}
	s.config.Logger.Debug("heartbeat sent", "targets", len(s.targets))
	return nil
}

func (s *MemoryHeartbeatService) RegisterTarget(target string, config TargetConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.targets[target]; exists {
		return fmt.Errorf("target already registered: %s", target)
	}
	s.targets[target] = &TargetState{
		Name:      target,
		Config:    config,
		IsHealthy: true,
		LastBeatAt: time.Now(),
		Status: TargetStatus{
			Name:       target,
			IsHealthy:  true,
			LastBeatAt: time.Now(),
			Message:    "Registered",
		},
	}
	s.config.Logger.Info("target registered", "name", target)
	return nil
}

func (s *MemoryHeartbeatService) UnregisterTarget(target string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.targets[target]; !exists {
		return fmt.Errorf("target not found: %s", target)
	}
	delete(s.targets, target)
	s.config.Logger.Info("target unregistered", "name", target)
	return nil
}

func (s *MemoryHeartbeatService) GetStatus(target string) (*TargetStatus, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	state, exists := s.targets[target]
	if !exists {
		return nil, fmt.Errorf("target not found: %s", target)
	}
	statusCopy := state.Status
	return &statusCopy, nil
}

func (s *MemoryHeartbeatService) GetAllStatuses() map[string]*TargetStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make(map[string]*TargetStatus)
	for _, state := range s.targets {
		statusCopy := state.Status
		result[state.Name] = &statusCopy
	}
	return result
}

func (s *MemoryHeartbeatService) GetStats() *HeartbeatStats {
	s.mu.RLock()
	defer s.mu.RUnlock()
	stats := &HeartbeatStats{TotalTargets: len(s.targets)}
	for _, state := range s.targets {
		if state.IsHealthy {
			stats.HealthyTargets++
		} else {
			stats.UnhealthyTargets++
		}
	}
	return stats
}

func (s *MemoryHeartbeatService) runLoop() {
	defer s.wg.Done()
	for {
		select {
		case <-s.ticker.C:
			s.checkTargets()
		case <-s.ctx.Done():
			return
		}
	}
}

func (s *MemoryHeartbeatService) checkTargets() {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	for _, target := range s.targets {
		elapsed := now.Sub(target.LastBeatAt)
		if elapsed > target.Config.Timeout {
			target.FailCount++
			target.IsHealthy = false
			status := &TargetStatus{
				Name:       target.Name,
				IsHealthy:  false,
				LastBeatAt: target.LastBeatAt,
				FailCount:  target.FailCount,
				Message:    fmt.Sprintf("Timeout: %v", elapsed),
			}
			target.Status = *status
			s.config.Logger.Warn("target unhealthy", "name", target.Name, "elapsed", elapsed)
			if target.Config.OnFailure != nil {
				go target.Config.OnFailure(s.ctx, target.Name, status)
			}
		} else if !target.IsHealthy {
			target.IsHealthy = true
			target.FailCount = 0
			status := &TargetStatus{
				Name:       target.Name,
				IsHealthy:  true,
				LastBeatAt: target.LastBeatAt,
				FailCount:  0,
				Message:    "Recovered",
			}
			target.Status = *status
			s.config.Logger.Info("target recovered", "name", target.Name)
			if target.Config.OnRecovery != nil {
				go target.Config.OnRecovery(s.ctx, target.Name, status)
			}
		}
	}
}
