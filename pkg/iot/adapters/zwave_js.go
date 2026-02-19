// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later


package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// ZWaveJSClient handles WebSocket communication with Z-Wave JS.
type ZWaveJSClient struct {
	conn         *websocket.Conn
	url          string
	messageID    int
	messages     chan []byte
	handler      ZWaveMessageHandler
	handlerMutex sync.RWMutex
	mutex        sync.Mutex
	connected    bool
}

// ZWaveMessageHandler handles Z-Wave JS messages.
type ZWaveMessageHandler func(message map[string]interface{})

// NewZWaveJSClient creates a new Z-Wave JS client instance.
func NewZWaveJSClient(url string) *ZWaveJSClient {
	return &ZWaveJSClient{
		url:       url,
		messages:  make(chan []byte, 100),
		messageID: 0,
	}
}

// Connect connects to Z-Wave JS server via WebSocket.
func (c *ZWaveJSClient) Connect(ctx context.Context) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.connected {
		return fmt.Errorf("already connected")
	}

	// Create WebSocket dialer
	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	// Connect to Z-Wave JS WebSocket server
	conn, _, err := dialer.Dial(c.url, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to Z-Wave JS: %w", err)
	}

	c.conn = conn
	c.connected = true

	// Start message receiver
	go c.receiveMessages()

	return nil
}

// Disconnect disconnects from Z-Wave JS server.
func (c *ZWaveJSClient) Disconnect(ctx context.Context) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.connected {
		return nil
	}

	if err := c.conn.Close(); err != nil {
		return fmt.Errorf("failed to close connection: %w", err)
	}

	c.connected = false
	return nil
}

// IsConnected returns whether the client is connected.
func (c *ZWaveJSClient) IsConnected() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.connected
}

// SendMessage sends a message to Z-Wave JS server.
func (c *ZWaveJSClient) SendMessage(ctx context.Context, message map[string]interface{}) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.connected {
		return fmt.Errorf("not connected")
	}

	// Add message ID
	c.messageID++
	message["messageId"] = c.messageID

	// Marshal message to JSON
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Send message
	if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

// receiveMessages receives messages from Z-Wave JS server.
func (c *ZWaveJSClient) receiveMessages() {
	defer close(c.messages)

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if !websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				// Log error
			}
			break
		}

		c.messages <- message
	}
}

// SetMessageHandler sets the message handler.
func (c *ZWaveJSClient) SetMessageHandler(handler ZWaveMessageHandler) {
	c.handlerMutex.Lock()
	defer c.handlerMutex.Unlock()
	c.handler = handler
}

// GetNodes retrieves all nodes from Z-Wave network.
func (c *ZWaveJSClient) GetNodes(ctx context.Context) ([]map[string]interface{}, error) {
	resultChan := make(chan []map[string]interface{}, 1)
	errorChan := make(chan error, 1)

	// Set temporary handler
	handler := func(message map[string]interface{}) {
		if msgType, ok := message["type"].(string); ok {
			if msgType == "result" {
				if nodes, ok := message["nodes"].([]map[string]interface{}); ok {
					resultChan <- nodes
				} else {
					errorChan <- fmt.Errorf("invalid nodes format")
				}
			} else if msgType == "error" {
				errorChan <- fmt.Errorf("Z-Wave JS error: %v", message["message"])
			}
		}
	}

	c.SetMessageHandler(handler)
	defer c.SetMessageHandler(nil)

	// Send getNodes command
	message := map[string]interface{}{
		"command": "node.get_nodes",
	}

	if err := c.SendMessage(ctx, message); err != nil {
		return nil, err
	}

	// Wait for response
	select {
	case nodes := <-resultChan:
		return nodes, nil
	case err := <-errorChan:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(10 * time.Second):
		return nil, fmt.Errorf("timeout waiting for response")
	}
}

// GetNodeInfo retrieves information about a specific node.
func (c *ZWaveJSClient) GetNodeInfo(ctx context.Context, nodeID uint8) (map[string]interface{}, error) {
	resultChan := make(chan map[string]interface{}, 1)
	errorChan := make(chan error, 1)

	// Set temporary handler
	handler := func(message map[string]interface{}) {
		if msgType, ok := message["type"].(string); ok {
			if msgType == "result" {
				if node, ok := message["node"].(map[string]interface{}); ok {
					resultChan <- node
				} else {
					errorChan <- fmt.Errorf("invalid node format")
				}
			} else if msgType == "error" {
				errorChan <- fmt.Errorf("Z-Wave JS error: %v", message["message"])
			}
		}
	}

	c.SetMessageHandler(handler)
	defer c.SetMessageHandler(nil)

	// Send getNodeInfo command
	message := map[string]interface{}{
		"command": "node.get_info",
		"nodeId":  nodeID,
	}

	if err := c.SendMessage(ctx, message); err != nil {
		return nil, err
	}

	// Wait for response
	select {
	case node := <-resultChan:
		return node, nil
	case err := <-errorChan:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(5 * time.Second):
		return nil, fmt.Errorf("timeout waiting for response")
	}
}

// SetValue sets a value on a Z-Wave device.
func (c *ZWaveJSClient) SetValue(ctx context.Context, nodeID uint8, commandClass string, value interface{}) error {
	message := map[string]interface{}{
		"command": "node.set_value",
		"nodeId":  nodeID,
		"commandClass": commandClass,
		"value":   value,
	}

	return c.SendMessage(ctx, message)
}

// GetValue gets a value from a Z-Wave device.
func (c *ZWaveJSClient) GetValue(ctx context.Context, nodeID uint8, commandClass string) (interface{}, error) {
	resultChan := make(chan interface{}, 1)
	errorChan := make(chan error, 1)

	// Set temporary handler
	handler := func(message map[string]interface{}) {
		if msgType, ok := message["type"].(string); ok {
			if msgType == "result" {
				resultChan <- message["value"]
			} else if msgType == "error" {
				errorChan <- fmt.Errorf("Z-Wave JS error: %v", message["message"])
			}
		}
	}

	c.SetMessageHandler(handler)
	defer c.SetMessageHandler(nil)

	// Send getValue command
	message := map[string]interface{}{
		"command": "node.get_value",
		"nodeId":  nodeID,
		"commandClass": commandClass,
	}

	if err := c.SendMessage(ctx, message); err != nil {
		return nil, err
	}

	// Wait for response
	select {
	case value := <-resultChan:
		return value, nil
	case err := <-errorChan:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(5 * time.Second):
		return nil, fmt.Errorf("timeout waiting for response")
	}
}

// StartInclusion starts inclusion mode (pairing).
func (c *ZWaveJSClient) StartInclusion(ctx context.Context, includeNonSecure bool) error {
	message := map[string]interface{}{
		"command": "controller.start_inclusion",
		"options": map[string]interface{}{
			"includeNonSecure": includeNonSecure,
		},
	}

	return c.SendMessage(ctx, message)
}

// StopInclusion stops inclusion mode.
func (c *ZWaveJSClient) StopInclusion(ctx context.Context) error {
	message := map[string]interface{}{
		"command": "controller.stop_inclusion",
	}

	return c.SendMessage(ctx, message)
}

// Exclusion excludes a node from the Z-Wave network.
func (c *ZWaveJSClient) Exclusion(ctx context.Context, nodeID uint8) error {
	message := map[string]interface{}{
		"command": "controller.exclusion",
		"nodeId":  nodeID,
	}

	return c.SendMessage(ctx, message)
}

// HealNetwork heals the Z-Wave network.
func (c *ZWaveJSClient) HealNetwork(ctx context.Context) error {
	message := map[string]interface{}{
		"command": "controller.heal_network",
	}

	return c.SendMessage(ctx, message)
}

// GetDriverInfo retrieves Z-Wave driver information.
func (c *ZWaveJSClient) GetDriverInfo(ctx context.Context) (map[string]interface{}, error) {
	resultChan := make(chan map[string]interface{}, 1)
	errorChan := make(chan error, 1)

	// Set temporary handler
	handler := func(message map[string]interface{}) {
		if msgType, ok := message["type"].(string); ok {
			if msgType == "result" {
				if info, ok := message["driver"].(map[string]interface{}); ok {
					resultChan <- info
				} else {
					errorChan <- fmt.Errorf("invalid driver info format")
				}
			} else if msgType == "error" {
				errorChan <- fmt.Errorf("Z-Wave JS error: %v", message["message"])
			}
		}
	}

	c.SetMessageHandler(handler)
	defer c.SetMessageHandler(nil)

	// Send getDriverInfo command
	message := map[string]interface{}{
		"command": "controller.get_driver_info",
	}

	if err := c.SendMessage(ctx, message); err != nil {
		return nil, err
	}

	// Wait for response
	select {
	case info := <-resultChan:
		return info, nil
	case err := <-errorChan:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(5 * time.Second):
		return nil, fmt.Errorf("timeout waiting for response")
	}
}
