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
	"context"
	"errors"
	"time"

	"AgentFramework/pkg/beads/hardware"
	"github.com/goburrow/modbus"
)

// ModbusDriver implements HardwareController for Modbus devices.
type ModbusDriver struct {
	config *ModbusDeviceConfig
	client modbus.Client
}

// ModbusDeviceConfig contains configuration for a Modbus device.
type ModbusDeviceConfig struct {
	Address   string `json:"address"`
	Port      int    `json:"port"`
	SlaveID   int    `json:"slave_id"`
	Timeout   int    `json:"timeout"`
	Retries   int    `json:"retries"`
}

// NewModbusDriver creates a new ModbusDriver instance.
func NewModbusDriver(config *ModbusDeviceConfig) *ModbusDriver {
	return &ModbusDriver{
		config: config,
	}
}

// Connect establishes a connection to the Modbus device.
func (d *ModbusDriver) Connect(ctx context.Context, config interface{}) error {
	if d.client != nil {
		return errors.New("already connected")
	}

	// Use provided config if available
	if cfg, ok := config.(*ModbusDeviceConfig); ok && cfg != nil {
		d.config = cfg
	}

	// Create TCP client
	handler := modbus.NewTCPClientHandler(d.config.Address)
	handler.Timeout = time.Duration(d.config.Timeout) * time.Millisecond
	handler.SlaveId = byte(d.config.SlaveID)

	err := handler.Connect()
	if err != nil {
		return err
	}

	d.client = modbus.NewClient(handler)
	return nil
}

// Disconnect closes the Modbus connection.
func (d *ModbusDriver) Disconnect(ctx context.Context) error {
	if d.client == nil {
		return errors.New("not connected")
	}

	// Close connection
	if handler, ok := d.client.(interface {
		Close() error
	}); ok {
		err := handler.Close()
		if err != nil {
			return err
		}
	}

	d.client = nil
	return nil
}

// SendCommand sends a command to the Modbus device.
func (d *ModbusDriver) SendCommand(ctx context.Context, cmd string, params map[string]interface{}) (interface{}, error) {
	if d.client == nil {
		return nil, errors.New("not connected")
	}

	switch cmd {
	case "read_coils":
		return d.readCoils(params)
	case "read_discrete_inputs":
		return d.readDiscreteInputs(params)
	case "read_holding_registers":
		return d.readHoldingRegisters(params)
	case "read_input_registers":
		return d.readInputRegisters(params)
	case "write_single_coil":
		return d.writeSingleCoil(params)
	case "write_single_register":
		return d.writeSingleRegister(params)
	case "write_multiple_coils":
		return d.writeMultipleCoils(params)
	case "write_multiple_registers":
		return d.writeMultipleRegisters(params)
	default:
		return nil, errors.New("unsupported command: " + cmd)
	}
}

func (d *ModbusDriver) readCoils(params map[string]interface{}) (interface{}, error) {
	address, err := getUint16Param(params, "address")
	if err != nil {
		return nil, err
	}

	count, err := getUint16Param(params, "count")
	if err != nil {
		return nil, err
	}

	results, err := d.client.ReadCoils(address, count)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"results": results,
		"address": address,
		"count":   count,
	}, nil
}

func (d *ModbusDriver) readDiscreteInputs(params map[string]interface{}) (interface{}, error) {
	address, err := getUint16Param(params, "address")
	if err != nil {
		return nil, err
	}

	count, err := getUint16Param(params, "count")
	if err != nil {
		return nil, err
	}

	results, err := d.client.ReadDiscreteInputs(address, count)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"results": results,
		"address": address,
		"count":   count,
	}, nil
}

func (d *ModbusDriver) readHoldingRegisters(params map[string]interface{}) (interface{}, error) {
	address, err := getUint16Param(params, "address")
	if err != nil {
		return nil, err
	}

	count, err := getUint16Param(params, "count")
	if err != nil {
		return nil, err
	}

	results, err := d.client.ReadHoldingRegisters(address, count)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"results": results,
		"address": address,
		"count":   count,
	}, nil
}

func (d *ModbusDriver) readInputRegisters(params map[string]interface{}) (interface{}, error) {
	address, err := getUint16Param(params, "address")
	if err != nil {
		return nil, err
	}

	count, err := getUint16Param(params, "count")
	if err != nil {
		return nil, err
	}

	results, err := d.client.ReadInputRegisters(address, count)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"results": results,
		"address": address,
		"count":   count,
	}, nil
}

func (d *ModbusDriver) writeSingleCoil(params map[string]interface{}) (interface{}, error) {
	address, err := getUint16Param(params, "address")
	if err != nil {
		return nil, err
	}

	value, err := GetBoolParam(params, "value")
	if err != nil {
		return nil, err
	}

	// Convert bool to uint16 (0 or 1)
	var coilValue uint16 = 0
	if value {
		coilValue = 1
	}

	result, err := d.client.WriteSingleCoil(address, coilValue)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"result":  result,
		"address": address,
		"value":   value,
	}, nil
}

func (d *ModbusDriver) writeSingleRegister(params map[string]interface{}) (interface{}, error) {
	address, err := getUint16Param(params, "address")
	if err != nil {
		return nil, err
	}

	value, err := getUint16Param(params, "value")
	if err != nil {
		return nil, err
	}

	result, err := d.client.WriteSingleRegister(address, value)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"result":  result,
		"address": address,
		"value":   value,
	}, nil
}

func (d *ModbusDriver) writeMultipleCoils(params map[string]interface{}) (interface{}, error) {
	address, err := getUint16Param(params, "address")
	if err != nil {
		return nil, err
	}

	values, err := GetBoolArrayParam(params, "values")
	if err != nil {
		return nil, err
	}

	// Convert []bool to []byte for WriteMultipleCoils
	data := make([]byte, len(values))
	for i, v := range values {
		if v {
			data[i] = 1
		} else {
			data[i] = 0
		}
	}

	result, err := d.client.WriteMultipleCoils(address, uint16(len(values)), data)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"result":  result,
		"address": address,
		"values":  values,
	}, nil
}

func (d *ModbusDriver) writeMultipleRegisters(params map[string]interface{}) (interface{}, error) {
	address, err := getUint16Param(params, "address")
	if err != nil {
		return nil, err
	}

	values, err := getUint16ArrayParam(params, "values")
	if err != nil {
		return nil, err
	}

	// Convert []uint16 to []byte for WriteMultipleRegisters
	data := make([]byte, len(values)*2)
	for i, v := range values {
		data[i*2] = byte(v >> 8)
		data[i*2+1] = byte(v & 0xFF)
	}

	result, err := d.client.WriteMultipleRegisters(address, uint16(len(values)), data)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"result":  result,
		"address": address,
		"values":  values,
	}, nil
}

// ReceiveData receives data from the Modbus device.
func (d *ModbusDriver) ReceiveData(ctx context.Context, timeout time.Duration) (interface{}, error) {
	// Modbus typically uses request/response pattern rather than passive data reception
	return nil, errors.New("not implemented for Modbus")
}

// GetStatus retrieves the current status of the Modbus device.
func (d *ModbusDriver) GetStatus(ctx context.Context) (map[string]interface{}, error) {
	if d.client == nil {
		return map[string]interface{}{
			"connected": false,
			"address":   d.config.Address,
		}, nil
	}

	return map[string]interface{}{
		"connected": true,
		"address":   d.config.Address,
		"port":      d.config.Port,
		"slave_id":  d.config.SlaveID,
		"status":    "ok",
	}, nil
}

// SubscribeEvents subscribes to Modbus device events (not implemented in basic version).
func (d *ModbusDriver) SubscribeEvents(ctx context.Context, eventTypes []string, handler hardware.EventHandler) error {
	return errors.New("not implemented")
}

// UnsubscribeEvents unsubscribes from Modbus device events.
func (d *ModbusDriver) UnsubscribeEvents(ctx context.Context, eventTypes []string) error {
	return errors.New("not implemented")
}

func getUint16Param(params map[string]interface{}, key string) (uint16, error) {
	value, ok := params[key]
	if !ok {
		return 0, errors.New("missing parameter: " + key)
	}

	f, ok := value.(float64)
	if !ok {
		return 0, errors.New("invalid parameter type for " + key)
	}

	return uint16(f), nil
}

func getUint16ArrayParam(params map[string]interface{}, key string) ([]uint16, error) {
	value, ok := params[key]
	if !ok {
		return nil, errors.New("missing parameter: " + key)
	}

	arr, ok := value.([]interface{})
	if !ok {
		return nil, errors.New("invalid parameter type for " + key)
	}

	uints := make([]uint16, len(arr))
	for i, v := range arr {
		f, ok := v.(float64)
		if !ok {
			return nil, errors.New("invalid element type at index " + string(i))
		}
		uints[i] = uint16(f)
	}

	return uints, nil
}