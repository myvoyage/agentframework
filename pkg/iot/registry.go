// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later


package iot

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// DeviceRegistry manages persistent device registry.
type DeviceRegistry struct {
	devices map[string]*DeviceInfo
	db      *sql.DB
	mutex   sync.RWMutex
}

// NewDeviceRegistry creates a new device registry.
func NewDeviceRegistry() *DeviceRegistry {
	return &DeviceRegistry{
		devices: make(map[string]*DeviceInfo),
	}
}

// NewDeviceRegistryWithDB creates a new device registry with SQLite persistence.
func NewDeviceRegistryWithDB(dbPath string) (*DeviceRegistry, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	registry := &DeviceRegistry{
		devices: make(map[string]*DeviceInfo),
		db:      db,
	}

	// Initialize database schema
	if err := registry.initDB(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Load devices from database
	if err := registry.loadDevices(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to load devices: %w", err)
	}

	return registry, nil
}

// initDB initializes the database schema.
func (r *DeviceRegistry) initDB() error {
	if r.db == nil {
		return nil
	}

	query := `
	CREATE TABLE IF NOT EXISTS devices (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		type TEXT NOT NULL,
		protocol TEXT NOT NULL,
		manufacturer TEXT,
		model TEXT,
		version TEXT,
		status TEXT NOT NULL,
		capabilities TEXT,
		properties TEXT,
		last_seen TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_protocol ON devices(protocol);
	CREATE INDEX IF NOT EXISTS idx_type ON devices(type);
	CREATE INDEX IF NOT EXISTS idx_status ON devices(status);
	`

	_, err := r.db.Exec(query)
	return err
}

// loadDevices loads devices from the database.
func (r *DeviceRegistry) loadDevices() error {
	if r.db == nil {
		return nil
	}

	query := `SELECT id, name, type, protocol, manufacturer, model, version, status, capabilities, properties, last_seen FROM devices`

	rows, err := r.db.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var info DeviceInfo
		var capabilitiesJSON, propertiesJSON string
		var lastSeen sql.NullTime

		err := rows.Scan(
			&info.ID,
			&info.Name,
			&info.Type,
			&info.Protocol,
			&info.Manufacturer,
			&info.Model,
			&info.Version,
			&info.Status,
			&capabilitiesJSON,
			&propertiesJSON,
			&lastSeen,
		)
		if err != nil {
			return err
		}

		// Parse JSON fields (simplified - in production, use proper JSON parsing)
		info.Properties = make(map[string]interface{})
		info.Capabilities = []DeviceCapability{}

		if lastSeen.Valid {
			info.LastSeen = lastSeen.Time
		}

		r.devices[info.ID] = &info
	}

	return rows.Err()
}

// Register registers a device in the registry.
func (r *DeviceRegistry) Register(info *DeviceInfo) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.devices[info.ID] = info

	// Persist to database if available
	if r.db != nil {
		return r.saveDevice(info)
	}

	return nil
}

// Unregister removes a device from the registry.
func (r *DeviceRegistry) Unregister(deviceID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.devices[deviceID]; !exists {
		return ErrDeviceNotFound
	}

	delete(r.devices, deviceID)

	// Remove from database if available
	if r.db != nil {
		_, err := r.db.Exec("DELETE FROM devices WHERE id = ?", deviceID)
		return err
	}

	return nil
}

// Get retrieves a device by ID.
func (r *DeviceRegistry) Get(deviceID string) (*DeviceInfo, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	device, exists := r.devices[deviceID]
	if !exists {
		return nil, ErrDeviceNotFound
	}

	return device, nil
}

// List lists all devices in the registry.
func (r *DeviceRegistry) List() []*DeviceInfo {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	devices := make([]*DeviceInfo, 0, len(r.devices))
	for _, device := range r.devices {
		devices = append(devices, device)
	}
	return devices
}

// Query queries devices based on criteria.
func (r *DeviceRegistry) Query(criteria QueryCriteria) []*DeviceInfo {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	results := make([]*DeviceInfo, 0)

	for _, device := range r.devices {
		if r.matchCriteria(device, criteria) {
			results = append(results, device)
		}
	}

	return results
}

// matchCriteria checks if a device matches query criteria.
func (r *DeviceRegistry) matchCriteria(device *DeviceInfo, criteria QueryCriteria) bool {
	// Filter by protocol
	if criteria.Protocol != "" && device.Protocol != criteria.Protocol {
		return false
	}

	// Filter by type
	if criteria.Type != "" && device.Type != criteria.Type {
		return false
	}

	// Filter by status
	if criteria.Status != "" && device.Status != criteria.Status {
		return false
	}

	// Filter by manufacturer
	if criteria.Manufacturer != "" && device.Manufacturer != criteria.Manufacturer {
		return false
	}

	// Filter by model
	if criteria.Model != "" && device.Model != criteria.Model {
		return false
	}

	// Filter by capability
	if len(criteria.Capabilities) > 0 {
		hasCapability := false
		for _, cap := range criteria.Capabilities {
			for _, deviceCap := range device.Capabilities {
				if deviceCap == cap {
					hasCapability = true
					break
				}
			}
			if hasCapability {
				break
			}
		}
		if !hasCapability {
			return false
		}
	}

	return true
}

// UpdateStatus updates device status.
func (r *DeviceRegistry) UpdateStatus(deviceID string, status DeviceStatus) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	device, exists := r.devices[deviceID]
	if !exists {
		return ErrDeviceNotFound
	}

	device.Status = status

	// Update in database if available
	if r.db != nil {
		_, err := r.db.Exec("UPDATE devices SET status = ? WHERE id = ?", status, deviceID)
		return err
	}

	return nil
}

// UpdateLastSeen updates device last seen timestamp.
func (r *DeviceRegistry) UpdateLastSeen(deviceID string, lastSeen time.Time) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	device, exists := r.devices[deviceID]
	if !exists {
		return ErrDeviceNotFound
	}

	device.LastSeen = lastSeen

	// Update in database if available
	if r.db != nil {
		_, err := r.db.Exec("UPDATE devices SET last_seen = ? WHERE id = ?", lastSeen, deviceID)
		return err
	}

	return nil
}

// UpdateProperties updates device properties.
func (r *DeviceRegistry) UpdateProperties(deviceID string, properties map[string]interface{}) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	device, exists := r.devices[deviceID]
	if !exists {
		return ErrDeviceNotFound
	}

	// Merge properties
	if device.Properties == nil {
		device.Properties = make(map[string]interface{})
	}
	for k, v := range properties {
		device.Properties[k] = v
	}

	// Update in database if available (simplified)
	if r.db != nil {
		// In production, serialize properties to JSON
		_, err := r.db.Exec("UPDATE devices SET properties = ? WHERE id = ?", "", deviceID)
		return err
	}

	return nil
}

// saveDevice saves a device to the database.
func (r *DeviceRegistry) saveDevice(info *DeviceInfo) error {
	if r.db == nil {
		return nil
	}

	// Simplified - in production, serialize capabilities and properties to JSON
	query := `
	INSERT OR REPLACE INTO devices
	(id, name, type, protocol, manufacturer, model, version, status, capabilities, properties, last_seen)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.Exec(
		query,
		info.ID,
		info.Name,
		info.Type,
		info.Protocol,
		info.Manufacturer,
		info.Model,
		info.Version,
		info.Status,
		"", // capabilities JSON
		"", // properties JSON
		info.LastSeen,
	)

	return err
}

// Count returns the number of devices in the registry.
func (r *DeviceRegistry) Count() int {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return len(r.devices)
}

// CountByProtocol returns the number of devices per protocol.
func (r *DeviceRegistry) CountByProtocol() map[ProtocolType]int {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	counts := make(map[ProtocolType]int)
	for _, device := range r.devices {
		counts[device.Protocol]++
	}
	return counts
}

// CountByStatus returns the number of devices per status.
func (r *DeviceRegistry) CountByStatus() map[DeviceStatus]int {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	counts := make(map[DeviceStatus]int)
	for _, device := range r.devices {
		counts[device.Status]++
	}
	return counts
}

// Close closes the registry and releases resources.
func (r *DeviceRegistry) Close() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.db != nil {
		return r.db.Close()
	}
	return nil
}

// QueryCriteria defines query criteria for device registry queries.
type QueryCriteria struct {
	Protocol     ProtocolType
	Type         DeviceType
	Status       DeviceStatus
	Manufacturer string
	Model        string
	Capabilities []DeviceCapability
}

// DeviceStats represents device statistics.
type DeviceStats struct {
	Total            int
	ByProtocol       map[ProtocolType]int
	ByType          map[DeviceType]int
	ByStatus        map[DeviceStatus]int
	OnlineCount     int
	OfflineCount    int
}

// GetStats returns device statistics.
func (r *DeviceRegistry) GetStats() *DeviceStats {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	stats := &DeviceStats{
		Total:      len(r.devices),
		ByProtocol: make(map[ProtocolType]int),
		ByType:     make(map[DeviceType]int),
		ByStatus:   make(map[DeviceStatus]int),
	}

	for _, device := range r.devices {
		stats.ByProtocol[device.Protocol]++
		stats.ByType[device.Type]++
		stats.ByStatus[device.Status]++

		if device.Status == DeviceStatusOnline {
			stats.OnlineCount++
		} else if device.Status == DeviceStatusOffline {
			stats.OfflineCount++
		}
	}

	return stats
}
