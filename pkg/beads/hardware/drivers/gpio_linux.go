// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later


//go:build linux
// +build linux

package drivers

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// LinuxGPIOChip implements GPIOChip for Linux using sysfs.
type LinuxGPIOChip struct {
	name   string
	base   int
	ngpio  int
	path   string
}

// LinuxPin implements Pin for Linux GPIO.
type LinuxPin struct {
	pin       int
	direction string
	activeLow bool
	path      string
	valueFile *os.File
}

// NewLinuxGPIOChip creates a new Linux GPIO chip.
func NewLinuxGPIOChip(name string) (*LinuxGPIOChip, error) {
	chipPath := filepath.Join("/sys/class/gpio", name)

	// Check if chip exists
	if _, err := os.Stat(chipPath); os.IsNotExist(err) {
		// Try to find chip
		chipPath = filepath.Join("/dev", name)
		if _, err := os.Stat(chipPath); os.IsNotExist(err) {
			return nil, fmt.Errorf("GPIO chip %s not found", name)
		}
	}

	chip := &LinuxGPIOChip{
		name: name,
		path: chipPath,
	}

	// Read chip info
	chip.readChipInfo()

	return chip, nil
}

// readChipInfo reads the chip's base and ngpio values.
func (c *LinuxGPIOChip) readChipInfo() {
	// Try to read base
	basePath := filepath.Join(c.path, "base")
	if data, err := os.ReadFile(basePath); err == nil {
		if base, err := strconv.Atoi(strings.TrimSpace(string(data))); err == nil {
			c.base = base
		}
	}

	// Try to read ngpio
	ngpioPath := filepath.Join(c.path, "ngpio")
	if data, err := os.ReadFile(ngpioPath); err == nil {
		if ngpio, err := strconv.Atoi(strings.TrimSpace(string(data))); err == nil {
			c.ngpio = ngpio
		}
	}

	// Default values
	if c.ngpio == 0 {
		c.ngpio = 28 // Default to 28 pins
	}
}

// OpenPin opens a GPIO pin.
func (c *LinuxGPIOChip) OpenPin(pin int, direction string, options ...PinOption) (Pin, error) {
	// Apply options
	cfg := &pinConfig{}
	for _, opt := range options {
		opt(cfg)
	}

	// Calculate absolute pin number
	absolutePin := c.base + pin

	// Export pin
	if err := c.exportPin(absolutePin); err != nil {
		return nil, fmt.Errorf("failed to export pin %d: %w", pin, err)
	}

	pinPath := filepath.Join("/sys/class/gpio", fmt.Sprintf("gpio%d", absolutePin))

	// Wait for pin directory to exist
	for i := 0; i < 100; i++ {
		if _, err := os.Stat(pinPath); !os.IsNotExist(err) {
			break
		}
	}

	// Set direction
	directionPath := filepath.Join(pinPath, "direction")
	if err := os.WriteFile(directionPath, []byte(direction), 0644); err != nil {
		// Try to unexport and re-export
		c.unexportPin(absolutePin)
		if err := c.exportPin(absolutePin); err != nil {
			return nil, fmt.Errorf("failed to set direction for pin %d: %w", pin, err)
		}
		if err := os.WriteFile(directionPath, []byte(direction), 0644); err != nil {
			return nil, fmt.Errorf("failed to set direction for pin %d: %w", pin, err)
		}
	}

	// Set active_low if needed
	if cfg.activeLow {
		activeLowPath := filepath.Join(pinPath, "active_low")
		os.WriteFile(activeLowPath, []byte("1"), 0644)
	}

	// Open value file for reading/writing
	valuePath := filepath.Join(pinPath, "value")
	var valueFile *os.File
	if direction == "in" {
		valueFile, _ = os.Open(valuePath)
	} else {
		valueFile, _ = os.OpenFile(valuePath, os.O_WRONLY, 0644)
	}

	return &LinuxPin{
		pin:       pin,
		direction: direction,
		activeLow: cfg.activeLow,
		path:      pinPath,
		valueFile: valueFile,
	}, nil
}

// Close closes the GPIO chip.
func (c *LinuxGPIOChip) Close() error {
	// Unexport all exported pins
	for pin := 0; pin < c.ngpio; pin++ {
		c.unexportPin(c.base + pin)
	}
	return nil
}

// exportPin exports a GPIO pin.
func (c *LinuxGPIOChip) exportPin(pin int) error {
	exportPath := "/sys/class/gpio/export"
	pinStr := strconv.Itoa(pin)

	// Check if already exported
	pinPath := filepath.Join("/sys/class/gpio", fmt.Sprintf("gpio%d", pin))
	if _, err := os.Stat(pinPath); !os.IsNotExist(err) {
		return nil // Already exported
	}

	return os.WriteFile(exportPath, []byte(pinStr), 0644)
}

// unexportPin unexports a GPIO pin.
func (c *LinuxGPIOChip) unexportPin(pin int) error {
	unexportPath := "/sys/class/gpio/unexport"
	pinStr := strconv.Itoa(pin)
	return os.WriteFile(unexportPath, []byte(pinStr), 0644)
}

// Read reads the pin value.
func (p *LinuxPin) Read() (int, error) {
	if p.valueFile != nil {
		p.valueFile.Seek(0, 0)
		buf := make([]byte, 1)
		_, err := p.valueFile.Read(buf)
		if err != nil {
			return 0, err
		}
		if buf[0] == '1' {
			return 1, nil
		}
		return 0, nil
	}

	// Fallback to reading from file
	valuePath := filepath.Join(p.path, "value")
	data, err := os.ReadFile(valuePath)
	if err != nil {
		return 0, err
	}

	value := strings.TrimSpace(string(data))
	if value == "1" {
		return 1, nil
	}
	return 0, nil
}

// Write writes a value to the pin.
func (p *LinuxPin) Write(value int) error {
	valueStr := "0"
	if value != 0 {
		valueStr = "1"
	}

	if p.valueFile != nil {
		_, err := p.valueFile.WriteString(valueStr)
		return err
	}

	valuePath := filepath.Join(p.path, "value")
	return os.WriteFile(valuePath, []byte(valueStr), 0644)
}

// SetDirection sets the pin direction.
func (p *LinuxPin) SetDirection(direction string) error {
	directionPath := filepath.Join(p.path, "direction")
	return os.WriteFile(directionPath, []byte(direction), 0644)
}

// SetEdge sets the pin edge trigger.
func (p *LinuxPin) SetEdge(edge string) error {
	edgePath := filepath.Join(p.path, "edge")
	return os.WriteFile(edgePath, []byte(edge), 0644)
}

// Close closes the pin.
func (p *LinuxPin) Close() error {
	if p.valueFile != nil {
		p.valueFile.Close()
	}

	// Extract pin number from path
	// gpio123 -> 123
	pinStr := strings.TrimPrefix(filepath.Base(p.path), "gpio")
	if pinNum, err := strconv.Atoi(pinStr); err == nil {
		unexportPath := "/sys/class/gpio/unexport"
		os.WriteFile(unexportPath, []byte(strconv.Itoa(pinNum)), 0644)
	}

	return nil
}
