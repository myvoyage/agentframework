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
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// ZigbeeMQTTClient handles MQTT communication with Zigbee2MQTT.
type ZigbeeMQTTClient struct {
	client      mqtt.Client
	brokerURL   string
	topicPrefix string
	messages    chan mqtt.Message
	handler     ZigbeeMessageHandler
	mutex       sync.RWMutex
	connected   bool
}

// ZigbeeMessageHandler handles incoming MQTT messages.
type ZigbeeMessageHandler func(topic string, payload []byte)

// ZigbeeMQTTDevice represents a Zigbee device from Zigbee2MQTT.
type ZigbeeMQTTDevice struct {
	ID                  string                 `json:"ieeeAddr"`
	FriendlyName         string                 `json:"friendly_name"`
	Type                string                 `json:"type"`
	Definition          string                 `json:"definition"`
	Description         string                 `json:"description"`
	Supported          bool                   `json:"supported"`
	PowerSource        string                 `json:"power_source"`
	DateCode            string                 `json:"date_code"`
	LastSeen            time.Time              `json:"last_seen"`
	Endpoints           []string               `json:"endpoints"`
	SoftwareBuildID     string                 `json:"software_build_id"`
	Icon                string                 `json:"icon"`
	Scanned             bool                   `json:"scanned"`
	Interviewing        bool                   `json:"interviewing"`
	Parameters          map[string]interface{} `json:"parameters"`
}

// NewZigbeeMQTTClient creates a new Zigbee2MQTT client.
func NewZigbeeMQTTClient(brokerURL string, topicPrefix string) *ZigbeeMQTTClient {
	return &ZigbeeMQTTClient{
		brokerURL:   brokerURL,
		topicPrefix: topicPrefix,
		messages:    make(chan mqtt.Message, 100),
		connected:   false,
	}
}

// Connect connects to the MQTT broker.
func (c *ZigbeeMQTTClient) Connect(ctx context.Context, clientID string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	opts := mqtt.NewClientOptions()
	opts.AddBroker(c.brokerURL)
	opts.SetClientID(clientID)
	opts.SetAutoReconnect(true)
	opts.SetOnConnectHandler(func(client mqtt.Client) {
		c.onConnected()
	})
	opts.SetConnectionLostHandler(func(client mqtt.Client, err error) {
		c.onConnectionLost(err)
	})

	c.client = mqtt.NewClient(opts)
	if token := c.client.Connect(); token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to connect to MQTT broker: %w", token.Error())
	}

	// Wait for connection to be established
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(5 * time.Second):
		if !c.connected {
			return fmt.Errorf("connection timeout")
		}
	}

	return nil
}

// Disconnect disconnects from the MQTT broker.
func (c *ZigbeeMQTTClient) Disconnect(ctx context.Context) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.client != nil {
		c.client.Disconnect(250)
		c.connected = false
	}
	return nil
}

// Subscribe subscribes to MQTT topics.
func (c *ZigbeeMQTTClient) Subscribe(ctx context.Context, topic string, qos byte) error {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if c.client == nil || !c.connected {
		return fmt.Errorf("not connected")
	}

	token := c.client.Subscribe(topic, qos, c.onMessage)
	if token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to subscribe to %s: %w", topic, token.Error())
	}

	return nil
}

// SubscribeToAllDevices subscribes to all device topics.
func (c *ZigbeeMQTTClient) SubscribeToAllDevices(ctx context.Context) error {
	// Subscribe to device state updates
	topic := fmt.Sprintf("%s/#", c.topicPrefix)
	return c.Subscribe(ctx, topic, 0)
}

// Publish publishes a message to an MQTT topic.
func (c *ZigbeeMQTTClient) Publish(ctx context.Context, topic string, payload interface{}) error {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if c.client == nil || !c.connected {
		return fmt.Errorf("not connected")
	}

	var bytes []byte
	var err error

	switch v := payload.(type) {
	case string:
		bytes = []byte(v)
	case []byte:
		bytes = v
	default:
		bytes, err = json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %w", err)
		}
	}

	token := c.client.Publish(topic, 0, false, bytes)
	if token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to publish to %s: %w", topic, token.Error())
	}

	return nil
}

// SetDeviceState sets a device state.
func (c *ZigbeeMQTTClient) SetDeviceState(ctx context.Context, deviceID string, state map[string]interface{}) error {
	topic := fmt.Sprintf("%s/%s/set", c.topicPrefix, deviceID)
	return c.Publish(ctx, topic, state)
}

// GetDeviceState requests device state.
func (c *ZigbeeMQTTClient) GetDeviceState(ctx context.Context, deviceID string) error {
	topic := fmt.Sprintf("%s/%s/get", c.topicPrefix, deviceID)
	return c.Publish(ctx, topic, "")
}

// GetDevices retrieves all devices from the bridge.
func (c *ZigbeeMQTTClient) GetDevices(ctx context.Context) ([]*ZigbeeMQTTDevice, error) {
	topic := fmt.Sprintf("%s/bridge/request/devices", c.topicPrefix)

	responseTopic := fmt.Sprintf("%s/bridge/response/devices", c.topicPrefix)

	// Subscribe to response
	if err := c.Subscribe(ctx, responseTopic, 0); err != nil {
		return nil, err
	}

	// Publish request
	if err := c.Publish(ctx, topic, ""); err != nil {
		return nil, err
	}

	// Wait for response (with timeout)
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	responseChan := make(chan []*ZigbeeMQTTDevice, 1)

	// Temporary handler for this request
	handler := func(topic string, payload []byte) {
		if topic == responseTopic {
			var devices []*ZigbeeMQTTDevice
			if err := json.Unmarshal(payload, &devices); err == nil {
				select {
				case responseChan <- devices:
				case <-ctx.Done():
				}
			}
		}
	}

	// Set temporary handler (in production, use a better mechanism)
	oldHandler := c.handler
	c.handler = handler
	defer func() { c.handler = oldHandler }()

	// Wait for response
	select {
	case devices := <-responseChan:
		return devices, nil
	case <-ctx.Done():
		return nil, fmt.Errorf("timeout waiting for device list")
	}
}

// PermitJoin enables or disables device joining.
func (c *ZigbeeMQTTClient) PermitJoin(ctx context.Context, enable bool) error {
	topic := fmt.Sprintf("%s/bridge/request/permit_join", c.topicPrefix)
	payload := map[string]interface{}{"value": enable}
	return c.Publish(ctx, topic, payload)
}

// SetMessageHandler sets the message handler.
func (c *ZigbeeMQTTClient) SetMessageHandler(handler ZigbeeMessageHandler) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.handler = handler
}

// IsConnected returns connection status.
func (c *ZigbeeMQTTClient) IsConnected() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.connected
}

// onConnected is called when MQTT connection is established.
func (c *ZigbeeMQTTClient) onConnected() {
	c.mutex.Lock()
	c.connected = true
	c.mutex.Unlock()

	// Subscribe to all device topics
	c.client.Subscribe(fmt.Sprintf("%s/#", c.topicPrefix), 0, c.onMessage)
}

// onConnectionLost is called when MQTT connection is lost.
func (c *ZigbeeMQTTClient) onConnectionLost(err error) {
	c.mutex.Lock()
	c.connected = false
	c.mutex.Unlock()
}

// onMessage handles incoming MQTT messages.
func (c *ZigbeeMQTTClient) onMessage(client mqtt.Client, msg mqtt.Message) {
	if c.handler != nil {
		c.handler(msg.Topic(), msg.Payload())
	}
}

// GetDeviceTopic returns the topic for a specific device.
func (c *ZigbeeMQTTClient) GetDeviceTopic(deviceID string) string {
	return fmt.Sprintf("%s/%s", c.topicPrefix, deviceID)
}

// GetDeviceSetTopic returns the set topic for a device.
func (c *ZigbeeMQTTClient) GetDeviceSetTopic(deviceID string) string {
	return fmt.Sprintf("%s/%s/set", c.topicPrefix, deviceID)
}

// GetDeviceGetTopic returns the get topic for a device.
func (c *ZigbeeMQTTClient) GetDeviceGetTopic(deviceID string) string {
	return fmt.Sprintf("%s/%s/get", c.topicPrefix, deviceID)
}
