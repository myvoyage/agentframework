// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later


package drivers

import (
	"context"
	"testing"
	"time"
)

func TestNewGPIODriver(t *testing.T) {
	config := &GPIODeviceConfig{
		Chip:     "gpiochip0",
		PinCount: 28,
	}

	driver := NewGPIODriver(config)
	if driver == nil {
		t.Fatal("NewGPIODriver returned nil")
	}

	if driver.config.Chip != "gpiochip0" {
		t.Errorf("Expected chip gpiochip0, got %s", driver.config.Chip)
	}

	if driver.config.PinCount != 28 {
		t.Errorf("Expected pin count 28, got %d", driver.config.PinCount)
	}
}

func TestNewGPIODriverNilConfig(t *testing.T) {
	driver := NewGPIODriver(nil)
	if driver == nil {
		t.Fatal("NewGPIODriver with nil config returned nil")
	}

	if driver.config.Chip != "gpiochip0" {
		t.Errorf("Expected default chip gpiochip0, got %s", driver.config.Chip)
	}

	if driver.config.PinCount != 28 {
		t.Errorf("Expected default pin count 28, got %d", driver.config.PinCount)
	}
}

func TestGPIODriverGetStatus(t *testing.T) {
	config := &GPIODeviceConfig{
		Chip:     "gpiochip0",
		PinCount: 28,
	}

	driver := NewGPIODriver(config)

	status, err := driver.GetStatus(context.Background())
	if err != nil {
		t.Fatalf("GetStatus failed: %v", err)
	}

	if status["chip"] != "gpiochip0" {
		t.Errorf("Expected chip gpiochip0 in status, got %v", status["chip"])
	}

	if status["pin_count"] != 28 {
		t.Errorf("Expected pin count 28 in status, got %v", status["pin_count"])
	}

	if status["connected"] != false {
		t.Errorf("Expected connected false in status, got %v", status["connected"])
	}
}

func TestGPIODriverConnectDisconnect(t *testing.T) {
	config := &GPIODeviceConfig{
		Chip:     "gpiochip0",
		PinCount: 28,
	}

	driver := NewGPIODriver(config)

	// Connect
	err := driver.Connect(context.Background(), nil)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	// Check status after connect
	status, _ := driver.GetStatus(context.Background())
	if status["connected"] != true {
		t.Error("Expected connected true after Connect")
	}

	// Disconnect
	err = driver.Disconnect(context.Background())
	if err != nil {
		t.Fatalf("Disconnect failed: %v", err)
	}

	// Check status after disconnect
	status, _ = driver.GetStatus(context.Background())
	if status["connected"] != false {
		t.Error("Expected connected false after Disconnect")
	}
}

func TestGPIODriverDoubleConnect(t *testing.T) {
	config := &GPIODeviceConfig{
		Chip:     "gpiochip0",
		PinCount: 28,
	}

	driver := NewGPIODriver(config)

	// First connect
	err := driver.Connect(context.Background(), nil)
	if err != nil {
		t.Fatalf("First connect failed: %v", err)
	}

	// Second connect should fail
	err = driver.Connect(context.Background(), nil)
	if err == nil {
		t.Error("Expected double Connect to fail")
	}
}

func TestGPIODriverDisconnectNotConnected(t *testing.T) {
	config := &GPIODeviceConfig{
		Chip:     "gpiochip0",
		PinCount: 28,
	}

	driver := NewGPIODriver(config)

	err := driver.Disconnect(context.Background())
	if err == nil {
		t.Error("Expected Disconnect to fail when not connected")
	}
}

func TestGPIODriverSendCommandNotConnected(t *testing.T) {
	config := &GPIODeviceConfig{
		Chip:     "gpiochip0",
		PinCount: 28,
	}

	driver := NewGPIODriver(config)

	params := map[string]interface{}{
		"pin":       1,
		"direction": "out",
	}

	_, err := driver.SendCommand(context.Background(), "setup_pin", params)
	if err == nil {
		t.Error("Expected SendCommand to fail when not connected")
	}
}

func TestGPIODriverUnsupportedCommand(t *testing.T) {
	config := &GPIODeviceConfig{
		Chip:     "gpiochip0",
		PinCount: 28,
	}

	driver := NewGPIODriver(config)

	// Connect
	err := driver.Connect(context.Background(), nil)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	_, err = driver.SendCommand(context.Background(), "unsupported_command", nil)
	if err == nil {
		t.Error("Expected unsupported command to fail")
	}
}

func TestGPIODriverPinOperations(t *testing.T) {
	config := &GPIODeviceConfig{
		Chip:     "gpiochip0",
		PinCount: 28,
	}

	driver := NewGPIODriver(config)

	// Connect
	err := driver.Connect(context.Background(), nil)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer driver.Disconnect(context.Background())

	// Setup pin as output
	setupParams := map[string]interface{}{
		"pin":       float64(1),
		"direction": "out",
	}

	_, err = driver.SendCommand(context.Background(), "setup_pin", setupParams)
	if err != nil {
		t.Fatalf("Setup pin failed: %v", err)
	}

	// Write pin
	writeParams := map[string]interface{}{
		"pin":   float64(1),
		"value": float64(1),
	}

	_, err = driver.SendCommand(context.Background(), "write_pin", writeParams)
	if err != nil {
		t.Fatalf("Write pin failed: %v", err)
	}

	// Read pin
	readParams := map[string]interface{}{
		"pin": float64(1),
	}

	result, err := driver.SendCommand(context.Background(), "read_pin", readParams)
	if err != nil {
		t.Fatalf("Read pin failed: %v", err)
	}

	// Check result type
	readResult, ok := result.(*GPIOReadResult)
	if !ok {
		t.Fatalf("Expected *GPIOReadResult, got %T", result)
	}

	if readResult.Pin != 1 {
		t.Errorf("Expected pin 1, got %d", readResult.Pin)
	}
}

func TestGPIODriverListPins(t *testing.T) {
	config := &GPIODeviceConfig{
		Chip:     "gpiochip0",
		PinCount: 28,
	}

	driver := NewGPIODriver(config)

	// Connect
	err := driver.Connect(context.Background(), nil)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer driver.Disconnect(context.Background())

	// Setup a pin
	setupParams := map[string]interface{}{
		"pin":       float64(1),
		"direction": "out",
	}

	_, err = driver.SendCommand(context.Background(), "setup_pin", setupParams)
	if err != nil {
		t.Fatalf("Setup pin failed: %v", err)
	}

	// List pins
	result, err := driver.SendCommand(context.Background(), "list_pins", nil)
	if err != nil {
		t.Fatalf("List pins failed: %v", err)
	}

	// Check result
	listResult, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected map[string]interface{}, got %T", result)
	}

	if listResult["success"] != true {
		t.Error("Expected success true")
	}

	if listResult["count"] != 1 {
		t.Errorf("Expected count 1, got %v", listResult["count"])
	}
}

func TestGPIODriverPWM(t *testing.T) {
	config := &GPIODeviceConfig{
		Chip:     "gpiochip0",
		PinCount: 28,
	}

	driver := NewGPIODriver(config)

	// Connect
	err := driver.Connect(context.Background(), nil)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer driver.Disconnect(context.Background())

	// Set PWM
	pwmParams := map[string]interface{}{
		"pin":        float64(1),
		"period":     float64(1000000), // 1ms
		"duty_cycle": float64(500000),  // 50%
	}

	result, err := driver.SendCommand(context.Background(), "set_pwm", pwmParams)
	if err != nil {
		t.Fatalf("Set PWM failed: %v", err)
	}

	// Check result
	pwmResult, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected map[string]interface{}, got %T", result)
	}

	if pwmResult["success"] != true {
		t.Error("Expected success true")
	}
}

func TestGPIODriverReceiveDataNotConnected(t *testing.T) {
	config := &GPIODeviceConfig{
		Chip:     "gpiochip0",
		PinCount: 28,
	}

	driver := NewGPIODriver(config)

	_, err := driver.ReceiveData(context.Background(), 1*time.Second)
	if err == nil {
		t.Error("Expected ReceiveData to fail when not connected")
	}
}

func TestGPIODriverGetIntParam(t *testing.T) {
	tests := []struct {
		name     string
		params   map[string]interface{}
		key      string
		expected int
		wantErr  bool
	}{
		{
			name:     "float64 value",
			params:   map[string]interface{}{"pin": float64(5)},
			key:      "pin",
			expected: 5,
			wantErr:  false,
		},
		{
			name:     "int value",
			params:   map[string]interface{}{"pin": int(5)},
			key:      "pin",
			expected: 5,
			wantErr:  false,
		},
		{
			name:     "int64 value",
			params:   map[string]interface{}{"pin": int64(5)},
			key:      "pin",
			expected: 5,
			wantErr:  false,
		},
		{
			name:     "missing key",
			params:   map[string]interface{}{},
			key:      "pin",
			expected: 0,
			wantErr:  true,
		},
		{
			name:     "invalid type",
			params:   map[string]interface{}{"pin": "invalid"},
			key:      "pin",
			expected: 0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := getIntParam(tt.params, tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("getIntParam() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if result != tt.expected {
				t.Errorf("getIntParam() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGPIODriverGetStringParam(t *testing.T) {
	tests := []struct {
		name     string
		params   map[string]interface{}
		key      string
		expected string
		wantErr  bool
	}{
		{
			name:     "string value",
			params:   map[string]interface{}{"direction": "out"},
			key:      "direction",
			expected: "out",
			wantErr:  false,
		},
		{
			name:     "missing key",
			params:   map[string]interface{}{},
			key:      "direction",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "invalid type",
			params:   map[string]interface{}{"direction": 123},
			key:      "direction",
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := getStringParam(tt.params, tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("getStringParam() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if result != tt.expected {
				t.Errorf("getStringParam() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// Test MockGPIOChip
func TestMockGPIOChip(t *testing.T) {
	chip := NewMockGPIOChip("mockchip0", 28)
	if chip == nil {
		t.Fatal("NewMockGPIOChip returned nil")
	}

	// Open a pin
	pin, err := chip.OpenPin(5, "out")
	if err != nil {
		t.Fatalf("OpenPin failed: %v", err)
	}

	// Write to pin
	err = pin.Write(1)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Read from pin
	value, err := pin.Read()
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if value != 1 {
		t.Errorf("Expected value 1, got %d", value)
	}

	// Close pin
	err = pin.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Close chip
	err = chip.Close()
	if err != nil {
		t.Fatalf("Chip Close failed: %v", err)
	}
}

func TestMockGPIOChipInvalidPin(t *testing.T) {
	chip := NewMockGPIOChip("mockchip0", 28)

	// Try to open invalid pin
	_, err := chip.OpenPin(100, "out")
	if err == nil {
		t.Error("Expected OpenPin to fail for invalid pin")
	}
}

func TestMockGPIOChipWriteInputPin(t *testing.T) {
	chip := NewMockGPIOChip("mockchip0", 28)

	// Open pin as input
	pin, err := chip.OpenPin(5, "in")
	if err != nil {
		t.Fatalf("OpenPin failed: %v", err)
	}

	// Try to write to input pin
	err = pin.Write(1)
	if err == nil {
		t.Error("Expected Write to fail for input pin")
	}
}

func TestMockPinSimulateInterrupt(t *testing.T) {
	chip := NewMockGPIOChip("mockchip0", 28)

	// Open pin as input
	pin, err := chip.OpenPin(5, "in")
	if err != nil {
		t.Fatalf("OpenPin failed: %v", err)
	}

	mockPin := pin.(*MockPin)

	// Simulate interrupt
	mockPin.SimulateInterrupt(1)

	// Read value
	value, _ := pin.Read()
	if value != 1 {
		t.Errorf("Expected value 1 after interrupt, got %d", value)
	}
}

func TestGPIOPinStateStruct(t *testing.T) {
	state := &GPIOPinState{
		Pin:       5,
		Direction: "out",
		Value:     1,
		Edge:      "rising",
		ActiveLow: false,
		Pull:      "up",
	}

	if state.Pin != 5 {
		t.Errorf("Expected Pin 5, got %d", state.Pin)
	}

	if state.Direction != "out" {
		t.Errorf("Expected Direction out, got %s", state.Direction)
	}

	if state.Value != 1 {
		t.Errorf("Expected Value 1, got %d", state.Value)
	}
}

func TestGPIOEventStruct(t *testing.T) {
	event := GPIOEvent{
		Pin:       5,
		Value:     1,
		Timestamp: time.Now().UnixNano(),
		Type:      "rising",
	}

	if event.Pin != 5 {
		t.Errorf("Expected Pin 5, got %d", event.Pin)
	}

	if event.Type != "rising" {
		t.Errorf("Expected Type rising, got %s", event.Type)
	}
}
