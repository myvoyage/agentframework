// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later


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
	ctx := context.Background()

	// Create IoT device manager
	manager := iot.NewIoTDeviceManager()
	defer manager.Close(ctx)

	// Create Thread adapter
	threadAdapter := adapters.NewThreadAdapter()

	// Configure Thread adapter
	config := iot.ProtocolConfig{
		Type: iot.ProtocolThread,
		Hardware: iot.HardwareConfig{
			Type:    "border_router",
			Timeout: 5000,
		},
		Network: iot.NetworkConfig{
			Channel: 15,
		},
		Metadata: map[string]interface{}{
			"interface":   "wpan0",
			"coap_port":   5683,
			"network": map[string]interface{}{
				"network_name":      "HomeThread",
				"pan_id":            uint16(0x1234),
				"channel":           uint8(15),
				"mesh_local_prefix": "fd00:abcd::/64",
				"on_mesh_prefix":    "2001:db8:1234::/64",
			},
		},
	}

	// Initialize adapter
	if err := threadAdapter.Initialize(ctx, config); err != nil {
		log.Fatalf("Failed to initialize Thread adapter: %v", err)
	}

	// Register adapter with manager
	if err := manager.RegisterAdapter(threadAdapter); err != nil {
		log.Fatalf("Failed to register Thread adapter: %v", err)
	}

	// Start adapter
	if err := threadAdapter.Start(ctx); err != nil {
		log.Fatalf("Failed to start Thread adapter: %v", err)
	}
	defer threadAdapter.Stop(ctx)

	fmt.Println("Thread adapter started successfully")

	// Example 1: Discover devices
	fmt.Println("\n=== Discovering Thread devices ===")
	devices, err := threadAdapter.DiscoverDevices(ctx, 10*time.Second)
	if err != nil {
		log.Printf("Device discovery failed: %v", err)
	} else {
		fmt.Printf("Found %d Thread devices\n", len(devices))
		for _, device := range devices {
			fmt.Printf("- %s (%s) - %s\n", device.Name, device.ID, device.Status)
		}
	}

	// Example 2: Get network information
	fmt.Println("\n=== Thread Network Information ===")
	networkInfo, err := threadAdapter.GetNetworkInfo(ctx)
	if err != nil {
		log.Printf("Failed to get network info: %v", err)
	} else {
		fmt.Printf("Protocol: %s\n", networkInfo.Protocol)
		fmt.Printf("PAN ID: 0x%04X\n", networkInfo.PanID)
		fmt.Printf("Channel: %d\n", networkInfo.Channel)
		fmt.Printf("Device Count: %d\n", networkInfo.DeviceCount)
		fmt.Printf("Running: %v\n", networkInfo.IsRunning)

		// Print additional properties
		for key, value := range networkInfo.Properties {
			fmt.Printf("%s: %v\n", key, value)
		}
	}

	// Example 3: Pair a new Thread device
	fmt.Println("\n=== Pairing Thread Device ===")
	fmt.Println("Put your Thread device in commissioning mode...")
	fmt.Println("Press Enter when ready")
	// fmt.Scanln() // Uncomment to actually wait for user input

	pairingResult, err := threadAdapter.StartPairing(ctx, 60*time.Second)
	if err != nil {
		log.Printf("Pairing failed: %v", err)
	} else if pairingResult.Success {
		fmt.Printf("Device paired successfully: %s\n", pairingResult.Device.Name)
		fmt.Printf("Device ID: %s\n", pairingResult.Device.ID)
		fmt.Printf("Device IPv6: %v\n", pairingResult.Device.Properties["ipv6"])

		// Get the device instance
		device, err := threadAdapter.GetDevice(ctx, pairingResult.Device.ID)
		if err != nil {
			log.Printf("Failed to get device: %v", err)
		} else {
			// Example 4: Read from device
			fmt.Println("\n=== Reading from Device ===")
			if value, err := device.Read(ctx, "temperature"); err == nil {
				fmt.Printf("Temperature: %v\n", value)
			}

			if value, err := device.Read(ctx, "humidity"); err == nil {
				fmt.Printf("Humidity: %v\n", value)
			}

			// Example 5: Subscribe to device events
			fmt.Println("\n=== Subscribing to Device Events ===")
			handler := func(ctx context.Context, event iot.DeviceEvent) {
				fmt.Printf("Device Event: %s\n", event.Type)
				fmt.Printf("Data: %v\n", event.Data)
			}

			cancel := device.Subscribe(ctx, []string{"state_changed", "attribute_changed"}, handler)
			defer cancel()

			fmt.Println("Subscribed to device events. Waiting for events...")

			// Example 6: Control device (if it's an actuator)
			if device.Type() == iot.DeviceTypeActuator {
				fmt.Println("\n=== Controlling Device ===")

				// Turn on device
				if err := device.Write(ctx, "state", "on"); err != nil {
					log.Printf("Failed to turn on device: %v", err)
				} else {
					fmt.Println("Device turned on")
				}

				// Wait a bit
				time.Sleep(2 * time.Second)

				// Turn off device
				if err := device.Write(ctx, "state", "off"); err != nil {
					log.Printf("Failed to turn off device: %v", err)
				} else {
					fmt.Println("Device turned off")
				}
			}

			// Example 7: Batch operations
			fmt.Println("\n=== Batch Operations ===")

			// Batch read
			if threadDev, ok := device.(*adapters.ThreadDevice); ok {
				values, err := threadDev.BatchRead(ctx, []string{"temperature", "humidity", "pressure"})
				if err == nil {
					fmt.Printf("Batch read results: %v\n", values)
				}

				// Batch write
				writeValues := map[string]interface{}{
					"config_interval": 60,
					"config_threshold": 25.5,
				}
				if err := threadDev.BatchWrite(ctx, writeValues); err == nil {
					fmt.Println("Batch write successful")
				}

				// Stream data
				fmt.Println("\n=== Streaming Data ===")
				dataChan, err := threadDev.Stream(ctx, "temperature", 5*time.Second)
				if err == nil {
					fmt.Println("Streaming temperature data...")
					count := 0
					for value := range dataChan {
						fmt.Printf("Temperature: %v\n", value)
						count++
						if count >= 3 {
							break
						}
					}
				}

				// Get diagnostic info
				fmt.Println("\n=== Diagnostic Information ===")
				diag, err := threadDev.GetDiagnosticInfo(ctx)
				if err == nil {
					for key, value := range diag {
						fmt.Printf("%s: %v\n", key, value)
					}
				}
			}
		}
	} else {
		fmt.Printf("Pairing failed: %s\n", pairingResult.Error)
	}

	// Example 8: List all devices
	fmt.Println("\n=== All Thread Devices ===")
	allDevices, err := threadAdapter.ListDevices(ctx)
	if err != nil {
		log.Printf("Failed to list devices: %v", err)
	} else {
		for _, dev := range allDevices {
			fmt.Printf("- %s (%s) - %s\n", dev.Name(), dev.ID(), dev.Status())
		}
	}

	// Example 9: Subscribe to manager events
	fmt.Println("\n=== Subscribing to Manager Events ===")
	managerHandler := func(event iot.Event) {
		fmt.Printf("Manager Event: %s\n", event.Type)
		fmt.Printf("Source: %s\n", event.Source)
		fmt.Printf("Data: %v\n", event.Data)
	}

	unsubscribe := manager.SubscribeToEvents("*", managerHandler)
	defer unsubscribe()

	// Wait for events
	fmt.Println("Waiting for events (30 seconds)...")
	time.Sleep(30 * time.Second)

	// Example 10: Remove a device
	fmt.Println("\n=== Removing Device ===")
	// Note: This would remove an actual device, use with caution
	// if len(allDevices) > 0 {
	//     deviceID := allDevices[0].ID()
	//     if err := threadAdapter.RemoveDevice(ctx, deviceID); err != nil {
	//         log.Printf("Failed to remove device: %v", err)
	//     } else {
	//         fmt.Printf("Device %s removed\n", deviceID)
	//     }
	// }

	fmt.Println("\n=== Examples Complete ===")
}
