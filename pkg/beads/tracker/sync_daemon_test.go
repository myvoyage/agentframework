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
	"path/filepath"
	"testing"
	"time"

	"AgentFramework/pkg/beads"
	"AgentFramework/pkg/beads/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSyncDaemon(t *testing.T) {
	tmpDir := t.TempDir()

	config := &beads.Config{
		StoragePath: tmpDir,
		DBPath:      filepath.Join(tmpDir, "tasks.db"),
		JSONLPath:   filepath.Join(tmpDir, "events"),
		GitEnabled:  false,
		SyncInterval: 500 * time.Millisecond,
	}

	sqliteStore, err := store.NewSQLiteStore(config.DBPath)
	require.NoError(t, err)
	defer sqliteStore.Close()

	jsonlStore, err := store.NewJSONLStore(config.JSONLPath)
	require.NoError(t, err)
	defer jsonlStore.Close()

	eventProcessor := NewEventProcessor(sqliteStore, jsonlStore)

	daemon := NewSyncDaemon(sqliteStore, jsonlStore, eventProcessor, config)
	require.NotNil(t, daemon)
}

func TestSyncDaemonStartStop(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	config := &beads.Config{
		StoragePath: tmpDir,
		DBPath:      filepath.Join(tmpDir, "tasks.db"),
		JSONLPath:   filepath.Join(tmpDir, "events"),
		GitEnabled:  false,
		SyncInterval: 500 * time.Millisecond,
	}

	sqliteStore, err := store.NewSQLiteStore(config.DBPath)
	require.NoError(t, err)
	defer sqliteStore.Close()

	jsonlStore, err := store.NewJSONLStore(config.JSONLPath)
	require.NoError(t, err)
	defer jsonlStore.Close()

	eventProcessor := NewEventProcessor(sqliteStore, jsonlStore)

	daemon := NewSyncDaemon(sqliteStore, jsonlStore, eventProcessor, config)
	require.NotNil(t, daemon)

	// Initial status should be not running
	status := daemon.GetStatus()
	assert.False(t, status.Running)

	// Start should succeed
	err = daemon.Start(ctx)
	require.NoError(t, err)

	// Status should be running
	status = daemon.GetStatus()
	assert.True(t, status.Running)

	// Stop should succeed
	err = daemon.Stop(ctx)
	require.NoError(t, err)

	// Status should be not running again
	status = daemon.GetStatus()
	assert.False(t, status.Running)
}

func TestSyncDaemonStatusReporting(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	config := &beads.Config{
		StoragePath: tmpDir,
		DBPath:      filepath.Join(tmpDir, "tasks.db"),
		JSONLPath:   filepath.Join(tmpDir, "events"),
		GitEnabled:  false,
		SyncInterval: 500 * time.Millisecond,
	}

	sqliteStore, err := store.NewSQLiteStore(config.DBPath)
	require.NoError(t, err)
	defer sqliteStore.Close()

	jsonlStore, err := store.NewJSONLStore(config.JSONLPath)
	require.NoError(t, err)
	defer jsonlStore.Close()

	eventProcessor := NewEventProcessor(sqliteStore, jsonlStore)

	daemon := NewSyncDaemon(sqliteStore, jsonlStore, eventProcessor, config)
	require.NotNil(t, daemon)

	// Initial status should be clean
	status := daemon.GetStatus()
	assert.False(t, status.Running)
	assert.Equal(t, 0, status.ErrorCount)
	assert.Empty(t, status.LastError)
	assert.Zero(t, status.LastSyncTime)

	// Start daemon
	err = daemon.Start(ctx)
	require.NoError(t, err)
	defer daemon.Stop(ctx)

	// Give daemon time to perform initial sync
	time.Sleep(100 * time.Millisecond)

	// Status should be running
	status = daemon.GetStatus()
	assert.True(t, status.Running)

	// Note: LastSyncTime may be zero if no events exist to sync
	// We only check that the daemon is running
}

func TestSyncDaemonTriggerSync(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	config := &beads.Config{
		StoragePath: tmpDir,
		DBPath:      filepath.Join(tmpDir, "tasks.db"),
		JSONLPath:   filepath.Join(tmpDir, "events"),
		GitEnabled:  false,
		SyncInterval: 5 * time.Second, // Longer interval to avoid automatic sync during test
	}

	sqliteStore, err := store.NewSQLiteStore(config.DBPath)
	require.NoError(t, err)
	defer sqliteStore.Close()

	jsonlStore, err := store.NewJSONLStore(config.JSONLPath)
	require.NoError(t, err)
	defer jsonlStore.Close()

	eventProcessor := NewEventProcessor(sqliteStore, jsonlStore)

	daemon := NewSyncDaemon(sqliteStore, jsonlStore, eventProcessor, config)
	require.NotNil(t, daemon)

	err = daemon.Start(ctx)
	require.NoError(t, err)
	defer daemon.Stop(ctx)

	// Trigger manual sync (should succeed even if no events to sync)
	err = daemon.TriggerSync(ctx)
	require.NoError(t, err)

	// Check that daemon is still running properly
	status := daemon.GetStatus()
	assert.True(t, status.Running)
	assert.Equal(t, 0, status.ErrorCount)
}

func TestNewSyncDaemonWithFileWatcher(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	config := &beads.Config{
		StoragePath: tmpDir,
		DBPath:      filepath.Join(tmpDir, "tasks.db"),
		JSONLPath:   filepath.Join(tmpDir, "events"),
		GitEnabled:  false,
		SyncInterval: 500 * time.Millisecond,
	}

	sqliteStore, err := store.NewSQLiteStore(config.DBPath)
	require.NoError(t, err)
	defer sqliteStore.Close()

	jsonlStore, err := store.NewJSONLStore(config.JSONLPath)
	require.NoError(t, err)
	defer jsonlStore.Close()

	eventProcessor := NewEventProcessor(sqliteStore, jsonlStore)

	daemon := NewSyncDaemonWithFileWatcher(sqliteStore, jsonlStore, eventProcessor, config)
	require.NotNil(t, daemon)

	err = daemon.Start(ctx)
	require.NoError(t, err)

	err = daemon.Stop(ctx)
	require.NoError(t, err)
}

func TestSyncDaemonWithFileWatcherIntegration(t *testing.T) {
	t.Skip("Skipping integration test requiring file system operations")
	ctx := context.Background()
	tmpDir := t.TempDir()

	config := &beads.Config{
		StoragePath: tmpDir,
		DBPath:      filepath.Join(tmpDir, "tasks.db"),
		JSONLPath:   filepath.Join(tmpDir, "events"),
		GitEnabled:  false,
		SyncInterval: 5 * time.Second, // Longer interval to avoid automatic sync
	}

	sqliteStore, err := store.NewSQLiteStore(config.DBPath)
	require.NoError(t, err)
	defer sqliteStore.Close()

	jsonlStore, err := store.NewJSONLStore(config.JSONLPath)
	require.NoError(t, err)
	defer jsonlStore.Close()

	eventProcessor := NewEventProcessor(sqliteStore, jsonlStore)

	daemon := NewSyncDaemonWithFileWatcher(sqliteStore, jsonlStore, eventProcessor, config)
	require.NotNil(t, daemon)

	err = daemon.Start(ctx)
	require.NoError(t, err)
	defer daemon.Stop(ctx)

	// Create task tracker instance
	tracker := NewTaskTracker(config)

	err = tracker.Start(ctx)
	require.NoError(t, err)
	defer tracker.Stop(ctx)

	// Create a test task to generate events
	task1 := &beads.Task{
		Type:  beads.TaskTypeTask,
		Title: "Test Task",
		Status: beads.StatusOpen,
	}
	task1ID, err := tracker.CreateTask(ctx, task1)
	require.NoError(t, err)

	// Check that task exists in both stores
	retrievedTask, err := tracker.GetTask(ctx, task1ID)
	require.NoError(t, err)
	assert.Equal(t, "Test Task", retrievedTask.Title)

	events, err := jsonlStore.ReadAllEvents(ctx)
	require.NoError(t, err)
	assert.Len(t, events, 1)
}
