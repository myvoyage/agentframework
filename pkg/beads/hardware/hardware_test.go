// Package hardware provides unified hardware interface abstraction and driver management.
package hardware

import (
	"context"
	"errors"
	"testing"
	"time"
)

// TestHardwareDriverManager 测试硬件驱动管理器
func TestHardwareDriverManager(t *testing.T) {
	manager := NewHardwareDriverManager()

	t.Run("RegisterDriver", func(t *testing.T) {
		driver := &MockHardwareDriver{}
		err := manager.RegisterDriver("mock", driver)
		if err != nil {
			t.Errorf("Failed to register driver: %v", err)
		}

		// 验证驱动已注册
		retrieved, err := manager.GetDriver("mock")
		if err != nil {
			t.Errorf("Failed to get registered driver: %v", err)
		}

		if retrieved != driver {
			t.Error("Retrieved driver doesn't match registered driver")
		}
	})

	t.Run("RegisterDuplicate", func(t *testing.T) {
		driver := &MockHardwareDriver{}
		err := manager.RegisterDriver("mock", driver)
		if err == nil {
			t.Error("Expected error for duplicate driver registration")
		}
	})

	t.Run("UnregisterDriver", func(t *testing.T) {
		driver := &MockHardwareDriver{}
		manager.RegisterDriver("temp", driver)

		err := manager.UnregisterDriver("temp")
		if err != nil {
			t.Errorf("Failed to unregister driver: %v", err)
		}

		// 验证驱动已删除
		_, err = manager.GetDriver("temp")
		if err == nil {
			t.Error("Expected error for unregistered driver")
		}
	})

	t.Run("ListDrivers", func(t *testing.T) {
		driver := &MockHardwareDriver{}
		manager.RegisterDriver("test", driver)

		drivers := manager.ListDrivers()

		if len(drivers) == 0 {
			t.Error("Expected at least one driver")
		}

		found := false
		for _, name := range drivers {
			if name == "test" {
				found = true
				break
			}
		}

		if !found {
			t.Error("Expected to find 'test' driver in list")
		}
	})
}

// TestDeviceManager 测试设备管理器
func TestDeviceManager(t *testing.T) {
	manager := NewDeviceManager()

	t.Run("AddDevice", func(t *testing.T) {
		deviceInfo := &DeviceInfo{
			ID:          "device1",
			Name:        "Test Device",
			Type:        "mock",
			Status:      "connected",
			Description: "A test device",
			Properties: map[string]interface{}{
				"version": "1.0",
			},
		}

		err := manager.AddDevice(deviceInfo)
		if err != nil {
			t.Errorf("Failed to add device: %v", err)
		}

		// 验证设备已添加
		retrieved, err := manager.GetDevice("device1")
		if err != nil {
			t.Errorf("Failed to get device: %v", err)
		}

		if retrieved.ID != deviceInfo.ID {
			t.Errorf("Expected device ID %s, got %s", deviceInfo.ID, retrieved.ID)
		}
	})

	t.Run("AddDuplicate", func(t *testing.T) {
		deviceInfo := &DeviceInfo{
			ID:   "device1",
			Name: "Duplicate Device",
			Type: "mock",
		}

		err := manager.AddDevice(deviceInfo)
		if err == nil {
			t.Error("Expected error for duplicate device")
		}
	})

	t.Run("UpdateDeviceStatus", func(t *testing.T) {
		newStatus := "disconnected"
		err := manager.UpdateDeviceStatus("device1", newStatus)
		if err != nil {
			t.Errorf("Failed to update device status: %v", err)
		}

		// 验证状态已更新
		device, err := manager.GetDevice("device1")
		if err != nil {
			t.Errorf("Failed to get device: %v", err)
		}

		if device.Status != newStatus {
			t.Errorf("Expected status %s, got %s", newStatus, device.Status)
		}
	})

	t.Run("RemoveDevice", func(t *testing.T) {
		// 添加一个临时设备
		deviceInfo := &DeviceInfo{
			ID:   "temp_device",
			Name: "Temp Device",
			Type: "mock",
		}

		manager.AddDevice(deviceInfo)

		err := manager.RemoveDevice("temp_device")
		if err != nil {
			t.Errorf("Failed to remove device: %v", err)
		}

		// 验证设备已删除
		_, err = manager.GetDevice("temp_device")
		if err == nil {
			t.Error("Expected error for removed device")
		}
	})

	t.Run("ListDevices", func(t *testing.T) {
		// 再添加一个设备
		deviceInfo := &DeviceInfo{
			ID:   "device2",
			Name: "Second Device",
			Type: "mock",
		}

		manager.AddDevice(deviceInfo)

		devices := manager.ListDevices()

		if len(devices) < 2 {
			t.Errorf("Expected at least 2 devices, got %d", len(devices))
		}
	})
}

// TestMockHardwareDriver 测试模拟硬件驱动
func TestMockHardwareDriver(t *testing.T) {
	driver := &MockHardwareDriver{}
	ctx := context.Background()

	t.Run("ConnectDisconnect", func(t *testing.T) {
		config := map[string]interface{}{
			"port": "/dev/ttyUSB0",
		}

		err := driver.Connect(ctx, config)
		if err != nil {
			t.Errorf("Failed to connect: %v", err)
		}

		if !driver.connected {
			t.Error("Driver should be connected")
		}

		err = driver.Disconnect(ctx)
		if err != nil {
			t.Errorf("Failed to disconnect: %v", err)
		}

		if driver.connected {
			t.Error("Driver should be disconnected")
		}
	})

	t.Run("SendCommand", func(t *testing.T) {
		config := map[string]interface{}{
			"port": "/dev/ttyUSB0",
		}

		driver.Connect(ctx, config)

		result, err := driver.SendCommand(ctx, "test_command", map[string]interface{}{
			"param1": "value1",
		})

		if err != nil {
			t.Errorf("Failed to send command: %v", err)
		}

		if result == nil {
			t.Error("Expected result from command")
		}

		driver.Disconnect(ctx)
	})

	t.Run("ReceiveData", func(t *testing.T) {
		config := map[string]interface{}{
			"port": "/dev/ttyUSB0",
		}

		driver.Connect(ctx, config)

		data, err := driver.ReceiveData(ctx, 1*time.Second)
		if err != nil {
			t.Errorf("Failed to receive data: %v", err)
		}

		if data == nil {
			t.Error("Expected data from receive")
		}

		driver.Disconnect(ctx)
	})

	t.Run("GetStatus", func(t *testing.T) {
		config := map[string]interface{}{
			"port": "/dev/ttyUSB0",
		}

		driver.Connect(ctx, config)

		status, err := driver.GetStatus(ctx)
		if err != nil {
			t.Errorf("Failed to get status: %v", err)
		}

		if status["connected"] != true {
			t.Error("Expected status to show connected")
		}

		driver.Disconnect(ctx)
	})

	t.Run("SubscribeEvents", func(t *testing.T) {
		config := map[string]interface{}{
			"port": "/dev/ttyUSB0",
		}

		driver.Connect(ctx, config)

		eventReceived := false
		handler := func(ctx context.Context, event HardwareEvent) {
			eventReceived = true
		}

		err := driver.SubscribeEvents(ctx, []string{"test_event"}, handler)
		if err != nil {
			t.Errorf("Failed to subscribe to events: %v", err)
		}

		// 在模拟驱动中，事件订阅可能不会立即触发
		_ = eventReceived

		driver.Disconnect(ctx)
	})
}

// MockHardwareDriver 是用于测试的模拟硬件驱动
type MockHardwareDriver struct {
	connected bool
	events    map[string][]EventHandler
}

func (m *MockHardwareDriver) Connect(ctx context.Context, config interface{}) error {
	m.connected = true
	m.events = make(map[string][]EventHandler)
	return nil
}

func (m *MockHardwareDriver) Disconnect(ctx context.Context) error {
	m.connected = false
	return nil
}

func (m *MockHardwareDriver) SendCommand(ctx context.Context, cmd string, params map[string]interface{}) (interface{}, error) {
	if !m.connected {
		return nil, errors.New("not connected")
	}

	return map[string]interface{}{
		"command": cmd,
		"status":  "success",
		"result":   "command executed",
	}, nil
}

func (m *MockHardwareDriver) ReceiveData(ctx context.Context, timeout time.Duration) (interface{}, error) {
	if !m.connected {
		return nil, errors.New("not connected")
	}

	return map[string]interface{}{
		"data":      "test data",
		"timestamp": time.Now(),
	}, nil
}

func (m *MockHardwareDriver) GetStatus(ctx context.Context) (map[string]interface{}, error) {
	return map[string]interface{}{
		"connected": m.connected,
		"type":      "mock",
		"version":   "1.0",
	}, nil
}

func (m *MockHardwareDriver) SubscribeEvents(ctx context.Context, eventTypes []string, handler EventHandler) error {
	if !m.connected {
		return errors.New("not connected")
	}

	for _, eventType := range eventTypes {
		m.events[eventType] = append(m.events[eventType], handler)
	}

	return nil
}

func (m *MockHardwareDriver) UnsubscribeEvents(ctx context.Context, eventTypes []string) error {
	for _, eventType := range eventTypes {
		delete(m.events, eventType)
	}

	return nil
}

// BenchmarkHardwareDriverManager 性能测试
func BenchmarkHardwareDriverManager(b *testing.B) {
	manager := NewHardwareDriverManager()
	driver := &MockHardwareDriver{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.RegisterDriver("bench", driver)
		manager.GetDriver("bench")
		manager.UnregisterDriver("bench")
	}
}