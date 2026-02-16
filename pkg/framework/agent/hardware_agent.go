// Package agent provides hardware control agent implementation.
package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"AgentFramework/pkg/beads/hardware"
	"AgentFramework/pkg/beads/hardware/drivers"
)

// HardwareAgent manages hardware device connections and operations.
type HardwareAgent struct {
	driverManager *hardware.HardwareDriverManager
	deviceManager *hardware.DeviceManager
	connections   map[string]hardware.HardwareController
	mutex         sync.RWMutex
}

// NewHardwareAgent creates a new HardwareAgent instance.
func NewHardwareAgent() *HardwareAgent {
	return &HardwareAgent{
		driverManager: hardware.NewHardwareDriverManager(),
		deviceManager: hardware.NewDeviceManager(),
		connections:   make(map[string]hardware.HardwareController),
	}
}

// Initialize initializes the hardware agent with default drivers.
func (a *HardwareAgent) Initialize(ctx context.Context) error {
	// Register default serial driver
	a.driverManager.RegisterDriver("serial", &drivers.SerialDriver{})

	// Register default Modbus driver
	a.driverManager.RegisterDriver("modbus", &drivers.ModbusDriver{})

	// Register CAN bus driver
	a.driverManager.RegisterDriver("can", &drivers.CANDriver{})

	// Register GPIO driver
	a.driverManager.RegisterDriver("gpio", &drivers.GPIODriver{})

	return nil
}

// ConnectDevice connects to a hardware device.
func (a *HardwareAgent) ConnectDevice(ctx context.Context, deviceID string, driverType string, config interface{}) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	// Check if device is already connected
	if _, exists := a.connections[deviceID]; exists {
		return errors.New("device already connected")
	}

	// Get driver
	driver, err := a.driverManager.GetDriver(driverType)
	if err != nil {
		return fmt.Errorf("failed to get driver: %w", err)
	}

	// Connect to device
	if err := driver.Connect(ctx, config); err != nil {
		return fmt.Errorf("failed to connect to device: %w", err)
	}

	// Store connection
	a.connections[deviceID] = driver

	// Get device status
	status, err := driver.GetStatus(ctx)
	if err != nil {
		return fmt.Errorf("failed to get device status: %w", err)
	}

	// Add device to device manager
	deviceInfo := &hardware.DeviceInfo{
		ID:          deviceID,
		Name:        deviceID,
		Type:        driverType,
		Status:      "connected",
		Description: fmt.Sprintf("%s device", driverType),
		Properties:  status,
	}

	if err := a.deviceManager.AddDevice(deviceInfo); err != nil {
		return fmt.Errorf("failed to add device to manager: %w", err)
	}

	return nil
}

// DisconnectDevice disconnects from a hardware device.
func (a *HardwareAgent) DisconnectDevice(ctx context.Context, deviceID string) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	// Get connection
	driver, exists := a.connections[deviceID]
	if !exists {
		return errors.New("device not connected")
	}

	// Disconnect from device
	if err := driver.Disconnect(ctx); err != nil {
		return fmt.Errorf("failed to disconnect from device: %w", err)
	}

	// Remove connection
	delete(a.connections, deviceID)

	// Update device status
	if err := a.deviceManager.UpdateDeviceStatus(deviceID, "disconnected"); err != nil {
		return fmt.Errorf("failed to update device status: %w", err)
	}

	return nil
}

// SendCommand sends a command to a hardware device.
func (a *HardwareAgent) SendCommand(ctx context.Context, deviceID string, command string, params map[string]interface{}) (interface{}, error) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	// Get connection
	driver, exists := a.connections[deviceID]
	if !exists {
		return nil, errors.New("device not connected")
	}

	// Send command
	return driver.SendCommand(ctx, command, params)
}

// ReceiveData receives data from a hardware device.
func (a *HardwareAgent) ReceiveData(ctx context.Context, deviceID string, timeout time.Duration) (interface{}, error) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	// Get connection
	driver, exists := a.connections[deviceID]
	if !exists {
		return nil, errors.New("device not connected")
	}

	// Receive data
	return driver.ReceiveData(ctx, timeout)
}

// GetDeviceStatus retrieves the status of a hardware device.
func (a *HardwareAgent) GetDeviceStatus(ctx context.Context, deviceID string) (map[string]interface{}, error) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	// Get connection
	driver, exists := a.connections[deviceID]
	if !exists {
		return nil, errors.New("device not connected")
	}

	// Get status
	return driver.GetStatus(ctx)
}

// ListDevices lists all connected devices.
func (a *HardwareAgent) ListDevices(ctx context.Context) ([]*hardware.DeviceInfo, error) {
	return a.deviceManager.ListDevices(), nil
}

// SubscribeEvents subscribes to hardware device events.
func (a *HardwareAgent) SubscribeEvents(ctx context.Context, deviceID string, eventTypes []string, handler hardware.EventHandler) error {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	// Get connection
	driver, exists := a.connections[deviceID]
	if !exists {
		return errors.New("device not connected")
	}

	// Subscribe to events
	return driver.SubscribeEvents(ctx, eventTypes, handler)
}

// UnsubscribeEvents unsubscribes from hardware device events.
func (a *HardwareAgent) UnsubscribeEvents(ctx context.Context, deviceID string, eventTypes []string) error {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	// Get connection
	driver, exists := a.connections[deviceID]
	if !exists {
		return errors.New("device not connected")
	}

	// Unsubscribe from events
	return driver.UnsubscribeEvents(ctx, eventTypes)
}

// RegisterDriver registers a custom hardware driver.
func (a *HardwareAgent) RegisterDriver(name string, driver hardware.HardwareController) error {
	return a.driverManager.RegisterDriver(name, driver)
}

// UnregisterDriver unregisters a hardware driver.
func (a *HardwareAgent) UnregisterDriver(name string) error {
	return a.driverManager.UnregisterDriver(name)
}

// ListDrivers lists all registered drivers.
func (a *HardwareAgent) ListDrivers(ctx context.Context) []string {
	return a.driverManager.ListDrivers()
}

// GetDevice retrieves device information.
func (a *HardwareAgent) GetDevice(ctx context.Context, deviceID string) (*hardware.DeviceInfo, error) {
	return a.deviceManager.GetDevice(deviceID)
}

// ExecuteCommand executes a complex command sequence.
func (a *HardwareAgent) ExecuteCommand(ctx context.Context, commandSequence *CommandSequence) (*CommandResult, error) {
	results := make([]map[string]interface{}, 0)

	for _, cmd := range commandSequence.Commands {
		result, err := a.SendCommand(ctx, cmd.DeviceID, cmd.Command, cmd.Params)
		if err != nil {
			return &CommandResult{
				Success: false,
				Error:   err.Error(),
				Results: results,
			}, err
		}

		if resultMap, ok := result.(map[string]interface{}); ok {
			results = append(results, resultMap)
		} else {
			results = append(results, map[string]interface{}{
				"result": result,
			})
		}

		// Wait before next command if delay is specified
		if cmd.Delay > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(cmd.Delay):
			}
		}
	}

	return &CommandResult{
		Success: true,
		Results: results,
	}, nil
}

// CommandSequence represents a sequence of commands to execute.
type CommandSequence struct {
	Commands []Command `json:"commands"`
}

// Command represents a single command in a sequence.
type Command struct {
	DeviceID string                 `json:"device_id"`
	Command  string                 `json:"command"`
	Params   map[string]interface{} `json:"params"`
	Delay    time.Duration          `json:"delay"`
}

// CommandResult represents the result of a command sequence execution.
type CommandResult struct {
	Success bool                   `json:"success"`
	Error   string                 `json:"error,omitempty"`
	Results []map[string]interface{} `json:"results"`
}

// MarshalJSON implements JSON marshaling for CommandResult.
func (r *CommandResult) MarshalJSON() ([]byte, error) {
	type Alias CommandResult
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(r),
	})
}

// Close closes all connections and cleans up resources.
func (a *HardwareAgent) Close(ctx context.Context) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	var errs []error

	// Disconnect all devices
	for deviceID := range a.connections {
		if err := a.DisconnectDevice(ctx, deviceID); err != nil {
			errs = append(errs, fmt.Errorf("failed to disconnect device %s: %w", deviceID, err))
		}
	}

	// Clear connections
	a.connections = make(map[string]hardware.HardwareController)

	// Clear devices
	a.deviceManager = hardware.NewDeviceManager()

	if len(errs) > 0 {
		return fmt.Errorf("encountered %d errors during cleanup: %v", len(errs), errs)
	}

	return nil
}