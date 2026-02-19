// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later


// Package drivers provides specific hardware driver implementations.
package drivers

import (
	"fmt"
	"sync"
	"time"
)

// MockGPIOChip implements GPIOChip for non-Linux platforms (Windows, macOS).
// This is a mock implementation for development and testing purposes.
type MockGPIOChip struct {
	name     string
	pinCount int
	pins     map[int]*MockPin
	mu       sync.RWMutex
}

// MockPin implements Pin for mock GPIO.
type MockPin struct {
	pin       int
	direction string
	value     int
	activeLow bool
	edge      string
	chip      *MockGPIOChip
}

// NewMockGPIOChip creates a new mock GPIO chip.
func NewMockGPIOChip(name string, pinCount int) *MockGPIOChip {
	if pinCount <= 0 {
		pinCount = 28
	}
	return &MockGPIOChip{
		name:     name,
		pinCount: pinCount,
		pins:     make(map[int]*MockPin),
	}
}

// OpenPin opens a mock GPIO pin.
func (c *MockGPIOChip) OpenPin(pin int, direction string, options ...PinOption) (Pin, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if pin < 0 || pin >= c.pinCount {
		return nil, fmt.Errorf("pin %d out of range (0-%d)", pin, c.pinCount-1)
	}

	// Apply options
	cfg := &pinConfig{}
	for _, opt := range options {
		opt(cfg)
	}

	mockPin := &MockPin{
		pin:       pin,
		direction: direction,
		value:     0,
		activeLow: cfg.activeLow,
		edge:      cfg.edge,
		chip:      c,
	}

	c.pins[pin] = mockPin

	return mockPin, nil
}

// Close closes the mock GPIO chip.
func (c *MockGPIOChip) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.pins = make(map[int]*MockPin)
	return nil
}

// GetPin returns a mock pin by number.
func (c *MockGPIOChip) GetPin(pin int) (*MockPin, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	mockPin, exists := c.pins[pin]
	return mockPin, exists
}

// Read reads the mock pin value.
func (p *MockPin) Read() (int, error) {
	p.chip.mu.RLock()
	defer p.chip.mu.RUnlock()

	value := p.value
	if p.activeLow {
		value = 1 - value
	}

	return value, nil
}

// Write writes a value to the mock pin.
func (p *MockPin) Write(value int) error {
	p.chip.mu.Lock()
	defer p.chip.mu.Unlock()

	if p.direction != "out" {
		return fmt.Errorf("pin %d is not configured as output", p.pin)
	}

	p.value = value

	// Simulate some delay
	time.Sleep(1 * time.Millisecond)

	return nil
}

// SetDirection sets the mock pin direction.
func (p *MockPin) SetDirection(direction string) error {
	p.chip.mu.Lock()
	defer p.chip.mu.Unlock()

	validDirections := []string{"in", "out", "high", "low"}
	valid := false
	for _, d := range validDirections {
		if d == direction {
			valid = true
			break
		}
	}

	if !valid {
		return fmt.Errorf("invalid direction: %s", direction)
	}

	p.direction = direction

	// Set initial value for high/low directions
	if direction == "high" {
		p.direction = "out"
		p.value = 1
	} else if direction == "low" {
		p.direction = "out"
		p.value = 0
	}

	return nil
}

// SetEdge sets the mock pin edge trigger.
func (p *MockPin) SetEdge(edge string) error {
	p.chip.mu.Lock()
	defer p.chip.mu.Unlock()

	validEdges := []string{"none", "rising", "falling", "both"}
	valid := false
	for _, e := range validEdges {
		if e == edge {
			valid = true
			break
		}
	}

	if !valid {
		return fmt.Errorf("invalid edge: %s", edge)
	}

	p.edge = edge
	return nil
}

// Close closes the mock pin.
func (p *MockPin) Close() error {
	p.chip.mu.Lock()
	defer p.chip.mu.Unlock()

	delete(p.chip.pins, p.pin)
	return nil
}

// SimulateInterrupt simulates an interrupt on the pin (for testing).
func (p *MockPin) SimulateInterrupt(value int) {
	p.chip.mu.Lock()
	defer p.chip.mu.Unlock()

	p.value = value
}

// GetInfo returns information about the mock pin.
func (p *MockPin) GetInfo() map[string]interface{} {
	p.chip.mu.RLock()
	defer p.chip.mu.RUnlock()

	return map[string]interface{}{
		"pin":        p.pin,
		"direction":  p.direction,
		"value":      p.value,
		"active_low": p.activeLow,
		"edge":       p.edge,
	}
}
