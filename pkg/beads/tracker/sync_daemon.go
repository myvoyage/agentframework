// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later


package tracker

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"AgentFramework/pkg/beads"
)

// SyncDaemonImpl implements the beads.SyncDaemon interface
// It manages bidirectional synchronization between JSONL and SQLite stores
type SyncDaemonImpl struct {
	sqliteStore    beads.SQLiteStore
	jsonlStore     beads.JSONLStore
	eventProcessor beads.EventProcessor
	config         *beads.Config

	running       bool
	lastSyncTime  time.Time
	errorCount    int
	lastError     string
	mu            sync.RWMutex
	stopChan      chan struct{}
	wg            sync.WaitGroup
}

// NewSyncDaemon creates a new SyncDaemon instance
func NewSyncDaemon(
	sqliteStore beads.SQLiteStore,
	jsonlStore beads.JSONLStore,
	eventProcessor beads.EventProcessor,
	config *beads.Config,
) beads.SyncDaemon {
	return &SyncDaemonImpl{
		sqliteStore:    sqliteStore,
		jsonlStore:     jsonlStore,
		eventProcessor: eventProcessor,
		config:         config,
		stopChan:       make(chan struct{}),
	}
}

// Start begins the sync daemon with background goroutine
func (sd *SyncDaemonImpl) Start(ctx context.Context) error {
	sd.mu.Lock()
	if sd.running {
		sd.mu.Unlock()
		return fmt.Errorf("sync daemon already running")
	}
	sd.running = true
	sd.mu.Unlock()

	sd.wg.Add(1)
	go sd.syncLoop(ctx)

	return nil
}

// Stop gracefully shuts down the sync daemon
func (sd *SyncDaemonImpl) Stop(ctx context.Context) error {
	sd.mu.Lock()
	if !sd.running {
		sd.mu.Unlock()
		return fmt.Errorf("sync daemon not running")
	}

	// Signal stop and wait for goroutine to finish
	close(sd.stopChan)
	sd.mu.Unlock()

	doneChan := make(chan struct{})
	go func() {
		sd.wg.Wait()
		close(doneChan)
	}()

	select {
	case <-doneChan:
		sd.mu.Lock()
		sd.running = false
		sd.mu.Unlock()
		return nil
	case <-ctx.Done():
		return fmt.Errorf("timeout waiting for sync daemon to stop: %w", ctx.Err())
	}
}

// TriggerSync manually triggers a synchronization
func (sd *SyncDaemonImpl) TriggerSync(ctx context.Context) error {
	return sd.syncOnce(ctx)
}

// GetStatus returns the current sync status
func (sd *SyncDaemonImpl) GetStatus() beads.SyncStatus {
	sd.mu.RLock()
	defer sd.mu.RUnlock()

	return beads.SyncStatus{
		Running:      sd.running,
		LastSyncTime: sd.lastSyncTime,
		ErrorCount:   sd.errorCount,
		LastError:    sd.lastError,
	}
}

// syncLoop runs the periodic synchronization
func (sd *SyncDaemonImpl) syncLoop(ctx context.Context) {
	defer sd.wg.Done()

	// Initial synchronization
	if err := sd.syncOnce(ctx); err != nil {
		sd.mu.Lock()
		sd.errorCount++
		sd.lastError = err.Error()
		sd.mu.Unlock()
	}

	// Periodic synchronization
	ticker := time.NewTicker(sd.config.SyncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := sd.syncOnce(ctx); err != nil {
				sd.mu.Lock()
				sd.errorCount++
				sd.lastError = err.Error()
				sd.mu.Unlock()
			}
		case <-sd.stopChan:
			return
		case <-ctx.Done():
			return
		}
	}
}

// syncOnce performs a single synchronization from JSONL to SQLite
func (sd *SyncDaemonImpl) syncOnce(ctx context.Context) error {
	// Get last sync timestamp
	lastSyncTime := sd.getLastSyncTime()

	// Read events since last sync
	events, err := sd.jsonlStore.ReadEvents(ctx, lastSyncTime)
	if err != nil {
		return fmt.Errorf("failed to read events: %w", err)
	}

	if len(events) == 0 {
		return nil // No new events to sync
	}

	// Replay events to SQLite
	if err := sd.eventProcessor.ReplayEvents(ctx, events); err != nil {
		return fmt.Errorf("failed to replay events: %w", err)
	}

	// Update last sync time
	sd.mu.Lock()
	sd.lastSyncTime = time.Now()
	sd.mu.Unlock()

	return nil
}

// getLastSyncTime returns the last synchronization timestamp
func (sd *SyncDaemonImpl) getLastSyncTime() time.Time {
	sd.mu.RLock()
	defer sd.mu.RUnlock()

	return sd.lastSyncTime
}

// SyncDaemonWithFileWatcher extends SyncDaemonImpl with file watcher functionality
type SyncDaemonWithFileWatcher struct {
	*SyncDaemonImpl
	watcher chan struct{}
}

// NewSyncDaemonWithFileWatcher creates a new SyncDaemon with file watcher functionality
func NewSyncDaemonWithFileWatcher(
	sqliteStore beads.SQLiteStore,
	jsonlStore beads.JSONLStore,
	eventProcessor beads.EventProcessor,
	config *beads.Config,
) beads.SyncDaemon {
	daemon := &SyncDaemonWithFileWatcher{
		SyncDaemonImpl: NewSyncDaemon(sqliteStore, jsonlStore, eventProcessor, config).(*SyncDaemonImpl),
		watcher:        make(chan struct{}),
	}

	return daemon
}

// Start begins the sync daemon with file watcher
func (sd *SyncDaemonWithFileWatcher) Start(ctx context.Context) error {
	if err := sd.SyncDaemonImpl.Start(ctx); err != nil {
		return err
	}

	// Start file watcher in a goroutine
	go sd.watchForChanges(ctx)

	return nil
}

// watchForChanges monitors the JSONL directory for changes
func (sd *SyncDaemonWithFileWatcher) watchForChanges(ctx context.Context) {
	// Get initial file state
	fileState, err := sd.getDirectoryState()
	if err != nil {
		return
	}

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Check for directory changes
			newState, err := sd.getDirectoryState()
			if err != nil {
				continue
			}

			if sd.stateChanged(fileState, newState) {
				fileState = newState
				sd.triggerSync()
			}
		case <-sd.stopChan:
			return
		case <-ctx.Done():
			return
		}
	}
}

// getDirectoryState gets the current state of the JSONL directory
func (sd *SyncDaemonWithFileWatcher) getDirectoryState() (map[string]os.FileInfo, error) {
	entries, err := os.ReadDir(sd.config.JSONLPath)
	if err != nil {
		return nil, err
	}

	state := make(map[string]os.FileInfo)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		state[entry.Name()] = info
	}

	return state, nil
}

// stateChanged checks if directory state has changed
func (sd *SyncDaemonWithFileWatcher) stateChanged(
	oldState, newState map[string]os.FileInfo,
) bool {
	if len(oldState) != len(newState) {
		return true
	}

	for filename, oldInfo := range oldState {
		newInfo, exists := newState[filename]
		if !exists {
			return true
		}

		if newInfo.ModTime() != oldInfo.ModTime() ||
			newInfo.Size() != oldInfo.Size() {
			return true
		}
	}

	return false
}

// triggerSync triggers a manual synchronization
func (sd *SyncDaemonWithFileWatcher) triggerSync() {
	// Create a background context for the sync
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := sd.SyncDaemonImpl.TriggerSync(ctx); err != nil {
		sd.mu.Lock()
		sd.errorCount++
		sd.lastError = err.Error()
		sd.mu.Unlock()
	}
}

// DefaultSyncDaemonConfig returns a default synchronization configuration
func DefaultSyncDaemonConfig() *beads.Config {
	return &beads.Config{
		StoragePath:  ".beads",
		GitEnabled:   true,
		SyncInterval: 30 * time.Second,
		MaxTasks:     1000,
		DBPath:       filepath.Join(".beads", "tasks.db"),
		JSONLPath:    filepath.Join(".beads", "events"),
	}
}
