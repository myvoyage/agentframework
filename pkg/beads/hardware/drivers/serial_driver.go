// Package drivers provides specific hardware driver implementations.
package drivers

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"AgentFramework/pkg/beads/hardware"
	"github.com/tarm/serial"
)

// SerialDriver implements HardwareController for serial port devices.
type SerialDriver struct {
	config *SerialDeviceConfig
	port   *serial.Port
}

// SerialDeviceConfig contains configuration for a serial device.
type SerialDeviceConfig struct {
	Port     string `json:"port"`
	BaudRate int    `json:"baud_rate"`
	DataBits int    `json:"data_bits"`
	Parity   string `json:"parity"`
	StopBits int    `json:"stop_bits"`
	Timeout  int    `json:"timeout"`
}

// NewSerialDriver creates a new SerialDriver instance.
func NewSerialDriver(config *SerialDeviceConfig) *SerialDriver {
	return &SerialDriver{
		config: config,
	}
}

// Connect establishes a connection to the serial port.
func (d *SerialDriver) Connect(ctx context.Context, config interface{}) error {
	if d.port != nil {
		return errors.New("already connected")
	}

	// Use provided config if available
	if cfg, ok := config.(*SerialDeviceConfig); ok && cfg != nil {
		d.config = cfg
	}

	// Configure serial port
	cfg := &serial.Config{
		Name:        d.config.Port,
		Baud:        d.config.BaudRate,
		ReadTimeout: time.Duration(d.config.Timeout) * time.Millisecond,
	}

	var err error
	d.port, err = serial.OpenPort(cfg)
	if err != nil {
		return err
	}

	return nil
}

// Disconnect closes the serial port connection.
func (d *SerialDriver) Disconnect(ctx context.Context) error {
	if d.port == nil {
		return errors.New("not connected")
	}

	err := d.port.Close()
	if err != nil {
		return err
	}

	d.port = nil
	return nil
}

// SendCommand sends a command to the serial device.
func (d *SerialDriver) SendCommand(ctx context.Context, cmd string, params map[string]interface{}) (interface{}, error) {
	if d.port == nil {
		return nil, errors.New("not connected")
	}

	// Prepare command data
	commandData := map[string]interface{}{
		"command": cmd,
		"params":  params,
	}

	data, err := json.Marshal(commandData)
	if err != nil {
		return nil, err
	}

	// Send data
	_, err = d.port.Write(data)
	if err != nil {
		return nil, err
	}

	// Wait for response
	response := make([]byte, 1024)
	n, err := d.port.Read(response)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	err = json.Unmarshal(response[:n], &result)
	if err != nil {
		return string(response[:n]), nil // Return raw response if JSON parsing fails
	}

	return result, nil
}

// ReceiveData receives data from serial device.
func (d *SerialDriver) ReceiveData(ctx context.Context, timeout time.Duration) (interface{}, error) {
	if d.port == nil {
		return nil, errors.New("not connected")
	}

	// Read data
	data := make([]byte, 1024)
	n, err := d.port.Read(data)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	err = json.Unmarshal(data[:n], &result)
	if err != nil {
		return string(data[:n]), nil // Return raw response if JSON parsing fails
	}

	return result, nil
}

// GetStatus retrieves the current status of the serial device.
func (d *SerialDriver) GetStatus(ctx context.Context) (map[string]interface{}, error) {
	if d.port == nil {
		return map[string]interface{}{
			"connected": false,
			"port":      d.config.Port,
		}, nil
	}

	return map[string]interface{}{
		"connected": true,
		"port":      d.config.Port,
		"baud_rate": d.config.BaudRate,
		"status":    "ok",
	}, nil
}

// SubscribeEvents subscribes to serial device events (not implemented in basic version).
func (d *SerialDriver) SubscribeEvents(ctx context.Context, eventTypes []string, handler hardware.EventHandler) error {
	// Serial port event subscription not implemented in basic version
	return errors.New("not implemented")
}

// UnsubscribeEvents unsubscribes from serial device events.
func (d *SerialDriver) UnsubscribeEvents(ctx context.Context, eventTypes []string) error {
	// Serial port event unsubscription not implemented in basic version
	return errors.New("not implemented")
}
