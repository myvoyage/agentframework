package drivers

import (
	"context"
	"runtime"
	"testing"
	"time"
)

func TestNewCANDriver(t *testing.T) {
	config := &CANDeviceConfig{
		Interface: "can0",
		BaudRate:  500000,
		Timeout:   5000,
		EnableFD:  false,
	}

	driver := NewCANDriver(config)
	if driver == nil {
		t.Fatal("NewCANDriver returned nil")
	}

	if driver.config.Interface != "can0" {
		t.Errorf("Expected interface can0, got %s", driver.config.Interface)
	}

	if driver.config.BaudRate != 500000 {
		t.Errorf("Expected baud rate 500000, got %d", driver.config.BaudRate)
	}
}

func TestNewCANDriverNilConfig(t *testing.T) {
	driver := NewCANDriver(nil)
	if driver == nil {
		t.Fatal("NewCANDriver with nil config returned nil")
	}

	if driver.config.Interface != "can0" {
		t.Errorf("Expected default interface can0, got %s", driver.config.Interface)
	}

	if driver.config.BaudRate != 500000 {
		t.Errorf("Expected default baud rate 500000, got %d", driver.config.BaudRate)
	}
}

func TestCANDriverGetStatus(t *testing.T) {
	config := &CANDeviceConfig{
		Interface: "can0",
		BaudRate:  500000,
	}

	driver := NewCANDriver(config)

	status, err := driver.GetStatus(context.Background())
	if err != nil {
		t.Fatalf("GetStatus failed: %v", err)
	}

	if status["interface"] != "can0" {
		t.Errorf("Expected interface can0 in status, got %v", status["interface"])
	}

	if status["baud_rate"] != 500000 {
		t.Errorf("Expected baud rate 500000 in status, got %v", status["baud_rate"])
	}

	if status["connected"] != false {
		t.Errorf("Expected connected false in status, got %v", status["connected"])
	}

	if runtime.GOOS != "linux" {
		if _, ok := status["warning"]; !ok {
			t.Error("Expected warning in status for non-Linux platform")
		}
	}
}

func TestCANDriverConnectDisconnectNotLinux(t *testing.T) {
	if runtime.GOOS == "linux" {
		t.Skip("Skipping test on Linux platform")
	}

	config := &CANDeviceConfig{
		Interface: "can0",
		BaudRate:  500000,
	}

	driver := NewCANDriver(config)

	// Connect should fail on non-Linux platforms
	err := driver.Connect(context.Background(), nil)
	if err == nil {
		t.Error("Expected Connect to fail on non-Linux platform")
	}
}

func TestCANDriverDoubleConnect(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Skipping test on non-Linux platform")
	}

	config := &CANDeviceConfig{
		Interface: "can0",
		BaudRate:  500000,
	}

	driver := NewCANDriver(config)

	// First connect (will fail if no CAN interface)
	_ = driver.Connect(context.Background(), nil)

	// Second connect should fail
	err := driver.Connect(context.Background(), nil)
	if err == nil {
		t.Error("Expected double Connect to fail")
	}
}

func TestCANDriverDisconnectNotConnected(t *testing.T) {
	config := &CANDeviceConfig{
		Interface: "can0",
		BaudRate:  500000,
	}

	driver := NewCANDriver(config)

	err := driver.Disconnect(context.Background())
	if err == nil {
		t.Error("Expected Disconnect to fail when not connected")
	}
}

func TestCANDriverSendCommandNotConnected(t *testing.T) {
	config := &CANDeviceConfig{
		Interface: "can0",
		BaudRate:  500000,
	}

	driver := NewCANDriver(config)

	params := map[string]interface{}{
		"id":   uint32(0x123),
		"data": []byte{0x01, 0x02, 0x03},
	}

	_, err := driver.SendCommand(context.Background(), "send_frame", params)
	if err == nil {
		t.Error("Expected SendCommand to fail when not connected")
	}
}

func TestCANDriverUnsupportedCommand(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Skipping test on non-Linux platform")
	}

	config := &CANDeviceConfig{
		Interface: "can0",
		BaudRate:  500000,
	}

	driver := NewCANDriver(config)

	// Connect (will fail if no CAN interface)
	_ = driver.Connect(context.Background(), nil)

	_, err := driver.SendCommand(context.Background(), "unsupported_command", nil)
	if err == nil {
		t.Error("Expected unsupported command to fail")
	}
}

func TestCANDriverReceiveDataNotConnected(t *testing.T) {
	config := &CANDeviceConfig{
		Interface: "can0",
		BaudRate:  500000,
	}

	driver := NewCANDriver(config)

	_, err := driver.ReceiveData(context.Background(), 1*time.Second)
	if err == nil {
		t.Error("Expected ReceiveData to fail when not connected")
	}
}

func TestCANFrameStruct(t *testing.T) {
	frame := &CANFrame{
		ID:         0x123,
		IsExtended: false,
		IsRemote:   false,
		Data:       []byte{0x01, 0x02, 0x03, 0x04},
		Timestamp:  time.Now().UnixNano(),
	}

	if frame.ID != 0x123 {
		t.Errorf("Expected ID 0x123, got 0x%x", frame.ID)
	}

	if frame.IsExtended {
		t.Error("Expected IsExtended to be false")
	}

	if len(frame.Data) != 4 {
		t.Errorf("Expected 4 data bytes, got %d", len(frame.Data))
	}
}

func TestCANFilterStruct(t *testing.T) {
	filter := CANFilter{
		ID:   0x123,
		Mask: 0x7FF,
	}

	if filter.ID != 0x123 {
		t.Errorf("Expected ID 0x123, got 0x%x", filter.ID)
	}

	if filter.Mask != 0x7FF {
		t.Errorf("Expected Mask 0x7FF, got 0x%x", filter.Mask)
	}
}

func TestCANDriverGetUint32Param(t *testing.T) {
	tests := []struct {
		name     string
		params   map[string]interface{}
		key      string
		expected uint32
		wantErr  bool
	}{
		{
			name:     "float64 value",
			params:   map[string]interface{}{"id": float64(0x123)},
			key:      "id",
			expected: 0x123,
			wantErr:  false,
		},
		{
			name:     "int value",
			params:   map[string]interface{}{"id": int(0x123)},
			key:      "id",
			expected: 0x123,
			wantErr:  false,
		},
		{
			name:     "uint32 value",
			params:   map[string]interface{}{"id": uint32(0x123)},
			key:      "id",
			expected: 0x123,
			wantErr:  false,
		},
		{
			name:     "missing key",
			params:   map[string]interface{}{},
			key:      "id",
			expected: 0,
			wantErr:  true,
		},
		{
			name:     "invalid type",
			params:   map[string]interface{}{"id": "invalid"},
			key:      "id",
			expected: 0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := getUint32Param(tt.params, tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("getUint32Param() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if result != tt.expected {
				t.Errorf("getUint32Param() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCANDriverGetByteArrayParam(t *testing.T) {
	tests := []struct {
		name     string
		params   map[string]interface{}
		key      string
		expected []byte
		wantErr  bool
	}{
		{
			name:     "interface array",
			params:   map[string]interface{}{"data": []interface{}{float64(1), float64(2), float64(3)}},
			key:      "data",
			expected: []byte{1, 2, 3},
			wantErr:  false,
		},
		{
			name:     "missing key",
			params:   map[string]interface{}{},
			key:      "data",
			expected: []byte{},
			wantErr:  false,
		},
		{
			name:     "string value",
			params:   map[string]interface{}{"data": "test"},
			key:      "data",
			expected: []byte("test"),
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := getByteArrayParam(tt.params, tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("getByteArrayParam() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if string(result) != string(tt.expected) {
				t.Errorf("getByteArrayParam() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCANDriverGetBoolParam(t *testing.T) {
	tests := []struct {
		name     string
		params   map[string]interface{}
		key      string
		expected bool
		wantErr  bool
	}{
		{
			name:     "bool true",
			params:   map[string]interface{}{"flag": true},
			key:      "flag",
			expected: true,
			wantErr:  false,
		},
		{
			name:     "bool false",
			params:   map[string]interface{}{"flag": false},
			key:      "flag",
			expected: false,
			wantErr:  false,
		},
		{
			name:     "float64 nonzero",
			params:   map[string]interface{}{"flag": float64(1)},
			key:      "flag",
			expected: true,
			wantErr:  false,
		},
		{
			name:     "missing key",
			params:   map[string]interface{}{},
			key:      "flag",
			expected: false,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := getBoolParam(tt.params, tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("getBoolParam() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if result != tt.expected {
				t.Errorf("getBoolParam() = %v, want %v", result, tt.expected)
			}
		})
	}
}
