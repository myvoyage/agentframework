// Package iot provides unified IoT protocol abstraction layer.
//
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
	"time"
)

// ProtocolType defines supported IoT protocol types.
type ProtocolType string

const (
	ProtocolZigbee   ProtocolType = "zigbee"
	ProtocolZWave    ProtocolType = "zwave"
	ProtocolThread   ProtocolType = "thread"
	ProtocolNearLink ProtocolType = "nearlink"
	ProtocolMQTT     ProtocolType = "mqtt"
	ProtocolHTTP     ProtocolType = "http"
	ProtocolUnknown  ProtocolType = "unknown"
)

// DeviceType defines IoT device types.
type DeviceType string

const (
	DeviceTypeSensor     DeviceType = "sensor"
	DeviceTypeActuator   DeviceType = "actuator"
	DeviceTypeController DeviceType = "controller"
	DeviceTypeGateway    DeviceType = "gateway"
	DeviceTypeUnknown    DeviceType = "unknown"
)

// DeviceStatus defines IoT device status.
type DeviceStatus string

const (
	DeviceStatusOnline   DeviceStatus = "online"
	DeviceStatusOffline  DeviceStatus = "offline"
	DeviceStatusPairing  DeviceStatus = "pairing"
	DeviceStatusError    DeviceStatus = "error"
	DeviceStatusSleep    DeviceStatus = "sleep"
)

// DeviceCapability defines device capabilities.
type DeviceCapability string

const (
	CapabilityOnOff           DeviceCapability = "on_off"
	CapabilityLevelControl    DeviceCapability = "level_control"
	CapabilityColorControl    DeviceCapability = "color_control"
	CapabilityColorTemp       DeviceCapability = "color_temperature"
	CapabilitySensor          DeviceCapability = "sensor"
	CapabilitySwitch          DeviceCapability = "switch"
	CapabilityLock            DeviceCapability = "lock"
	CapabilityThermostat      DeviceCapability = "thermostat"
	CapabilityNotification    DeviceCapability = "notification"
	CapabilityBinarySensor    DeviceCapability = "binary_sensor"
	CapabilityMultistateInput DeviceCapability = "multistate_input"
)

// EventType defines IoT event types.
type EventType string

const (
	EventDeviceDiscovered    EventType = "device_discovered"
	EventDeviceJoined        EventType = "device_joined"
	EventDeviceLeft          EventType = "device_left"
	EventDeviceStatusChanged EventType = "device_status_changed"
	EventDataReceived        EventType = "data_received"
	EventError               EventType = "error"
)

// DeviceInfo contains information about an IoT device.
type DeviceInfo struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Type         DeviceType             `json:"type"`
	Protocol     ProtocolType           `json:"protocol"`
	Manufacturer string                 `json:"manufacturer"`
	Model        string                 `json:"model"`
	Version      string                 `json:"version"`
	Status       DeviceStatus           `json:"status"`
	Capabilities []DeviceCapability     `json:"capabilities"`
	Properties   map[string]interface{} `json:"properties"`
	LastSeen     time.Time              `json:"last_seen"`
}

// ProtocolConfig defines protocol configuration.
type ProtocolConfig struct {
	Type     ProtocolType       `json:"type"`
	Hardware HardwareConfig     `json:"hardware"`
	Network  NetworkConfig      `json:"network"`
	Metadata  map[string]string  `json:"metadata"`
}

// HardwareConfig defines hardware configuration.
type HardwareConfig struct {
	Type     string `json:"type"`     // "usb", "serial", "tcp", "mqtt", "websocket"
	Port     string `json:"port"`     // "/dev/ttyUSB0", "192.168.1.100:5555"
	BaudRate int    `json:"baud_rate"`
	DataBits int    `json:"data_bits"`
	Parity   string `json:"parity"`
	StopBits int    `json:"stop_bits"`
	Timeout  int    `json:"timeout"` // milliseconds
}

// NetworkConfig defines network configuration.
type NetworkConfig struct {
	NetworkName      string   `json:"network_name"`       // Thread network name
	PanID            uint16   `json:"pan_id"`             // Zigbee/Thread PAN ID
	PanIDExt         uint64   `json:"pan_id_ext"`         // Extended PAN ID
	Channel          uint8    `json:"channel"`            // RF channel
	NetworkKey       string   `json:"network_key"`        // Encryption key
	PermitJoin       bool     `json:"permit_join"`        // Allow devices to join
	MeshLocalPrefix  string   `json:"mesh_local_prefix"`  // Thread mesh-local prefix
	OnMeshPrefix     []string `json:"on_mesh_prefix"`     // Thread on-mesh prefixes
	BorderRouterIP   string   `json:"border_router_ip"`   // Thread border router IP
}

// PairingResult contains device pairing result.
type PairingResult struct {
	Success bool        `json:"success"`
	Device  *DeviceInfo `json:"device,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// NetworkInfo contains network information.
type NetworkInfo struct {
	PanID        uint16            `json:"pan_id"`
	Channel      uint8             `json:"channel"`
	DeviceCount  int               `json:"device_count"`
	Status       string            `json:"status"`
	PermitJoin   bool              `json:"permit_join"`
	Properties   map[string]string `json:"properties"`
}

// Message represents an IoT message.
type Message struct {
	ID        string                 `json:"id"`
	DeviceID  string                 `json:"device_id"`
	Protocol  ProtocolType           `json:"protocol"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// Command represents an IoT device command.
type Command struct {
	DeviceID string                 `json:"device_id"`
	Action   string                 `json:"action"`
	Value    interface{}            `json:"value"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}
