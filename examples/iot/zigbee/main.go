// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later


// Package main provides an example of using Zigbee devices with AgentFramework.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"AgentFramework/pkg/iot"
	"AgentFramework/pkg/iot"
)

func main() {
	fmt.Println("=== Zigbee IoT Device Example ===")

	ctx := context.Background()

	// Create IoT device manager
	manager := iot.NewIoTDeviceManager()
	defer manager.Close(ctx)

	// Create Zigbee adapter
	zigbeeAdapter := adapters.NewZigbeeAdapter()

	// Configure Zigbee adapter (for Zigbee2MQTT)
	config := iot.ProtocolConfig{
		Type: iot.ProtocolZigbee,
		Hardware: iot.HardwareConfig{
			Type:     "mqtt",
			Port:     "localhost:1883",
			Timeout:  5000,
		},
		Network: iot.NetworkConfig{
			Channel:    11,
			PermitJoin: false,
		},
		Metadata: map[string]string{
			"broker_url":    "tcp://localhost:1883",
			"topic_prefix":  "zigbee2mqtt",
		},
	}

	// Initialize adapter
	fmt.Println("\n1. Initializing Zigbee adapter...")
	if err := zigbeeAdapter.Initialize(ctx, config); err != nil {
		log.Fatalf("Failed to initialize adapter: %v", err)
	}

	// Register adapter
	if err := manager.RegisterAdapter(zigbeeAdapter); err != nil {
		log.Fatalf("Failed to register adapter: %v", err)
	}

	// Start adapter
	fmt.Println("2. Starting Zigbee adapter...")
	if err := zigbeeAdapter.Start(ctx); err != nil {
		log.Fatalf("Failed to start adapter: %v", err)
	}
	defer zigbeeAdapter.Stop(ctx)

	// Discover devices
	fmt.Println("\n3. Discovering Zigbee devices...")
	devices, err := zigbeeAdapter.DiscoverDevices(ctx, 10*time.Second)
	if err != nil {
		log.Printf("Warning: Failed to discover devices: %v\n", err)
	} else {
		fmt.Printf("   Found %d devices:\n", len(devices))
		for _, device := range devices {
			fmt.Printf("   - %s (%s)\n", device.Name, device.Type)
		}
	}

	// Example: Pair a new device
	fmt.Println("\n4. Example: Enable pairing mode (60 seconds)...")
	fmt.Println("   Press and hold the pairing button on your Zigbee device now")

	result, err := zigbeeAdapter.StartPairing(ctx, 60*time.Second)
	if err != nil {
		log.Printf("Pairing failed: %v\n", err)
	} else if result.Success {
		fmt.Printf("   Device paired successfully: %s\n", result.Device.Name)
	} else {
		fmt.Printf("   Pairing failed: %s\n", result.Error)
	}

	// List all devices
	fmt.Println("\n5. Listing all devices...")
	allDevices, err := manager.ListDevices(ctx)
	if err != nil {
		log.Printf("Failed to list devices: %v\n", err)
	} else {
		fmt.Printf("   Total devices: %d\n", len(allDevices))
		for _, device := range allDevices {
			fmt.Printf("   - %s (%s)\n", device.Name(), device.Type())
		}
	}

	// Example: Control a device
	fmt.Println("\n6. Example: Controlling a device...")
	devices, _ = zigbeeAdapter.ListDevices(ctx)
	if len(devices) > 0 {
		device := devices[0]
		fmt.Printf("   Controlling device: %s\n", device.Name())

		// Turn on
		if err := device.Write(ctx, "state", "ON"); err != nil {
			log.Printf("   Failed to turn on: %v\n", err)
		} else {
			fmt.Println("   ✓ Device turned on")

			// Wait a bit
			time.Sleep(2 * time.Second)

			// Turn off
			if err := device.Write(ctx, "state", "OFF"); err != nil {
				log.Printf("   Failed to turn off: %v\n", err)
			} else {
				fmt.Println("   ✓ Device turned off")
			}
		}
	} else {
		fmt.Println("   No devices available for control")
	}

	// Get network info
	fmt.Println("\n7. Getting network information...")
	networkInfo, err := zigbeeAdapter.GetNetworkInfo(ctx)
	if err != nil {
		log.Printf("Failed to get network info: %v\n", err)
	} else {
		fmt.Printf("   Network: PAN ID=0x%04X, Channel=%d, Devices=%d\n",
			networkInfo.PanID,
			networkInfo.Channel,
			networkInfo.DeviceCount)
	}

	fmt.Println("\n=== Example completed ===")
}
