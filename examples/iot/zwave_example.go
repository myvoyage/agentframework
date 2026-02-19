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
	"AgentFramework/pkg/iot/adapters"
)

func main() {
	ctx := context.Background()

	// Create IoT device manager
	manager := iot.NewIoTDeviceManager()
	defer manager.Close(ctx)

	// Create Z-Wave adapter
	zwaveAdapter := adapters.NewZWaveAdapter()

	// Configure Z-Wave adapter
	config := iot.ProtocolConfig{
		Type: iot.ProtocolZWave,
		Hardware: iot.HardwareConfig{
			Type:    "websocket",
			Timeout: 5000,
		},
		Metadata: map[string]string{
			"ws_url": "ws://localhost:3000",
		},
	}

	// Initialize adapter
	if err := zwaveAdapter.Initialize(ctx, config); err != nil {
		log.Fatalf("Failed to initialize Z-Wave adapter: %v", err)
	}

	// Register adapter with manager
	if err := manager.RegisterAdapter(zwaveAdapter); err != nil {
		log.Fatalf("Failed to register Z-Wave adapter: %v", err)
	}

	// Start adapter
	if err := zwaveAdapter.Start(ctx); err != nil {
		log.Fatalf("Failed to start Z-Wave adapter: %v", err)
	}
	defer zwaveAdapter.Stop(ctx)

	fmt.Println("Z-Wave adapter started successfully")

	// Example 1: Discover devices
	fmt.Println("\n=== Discovering Z-Wave devices ===")
	devices, err := zwaveAdapter.DiscoverDevices(ctx, 10*time.Second)
	if err != nil {
		log.Printf("Device discovery failed: %v", err)
	} else {
		fmt.Printf("Found %d Z-Wave devices\n", len(devices))
		for _, device := range devices {
			fmt.Printf("- %s (Node ID: %v) - %s\n",
				device.Name,
				device.Properties["node_id"],
				device.Status)
		}
	}

	// Example 2: Get network information
	fmt.Println("\n=== Z-Wave Network Information ===")
	networkInfo, err := zwaveAdapter.GetNetworkInfo(ctx)
	if err != nil {
		log.Printf("Failed to get network info: %v", err)
	} else {
		fmt.Printf("Protocol: %s\n", networkInfo.Status)
		fmt.Printf("Device Count: %d\n", networkInfo.DeviceCount)
		fmt.Printf("Status: %s\n", networkInfo.Status)

		// Print additional properties
		for key, value := range networkInfo.Properties {
			fmt.Printf("%s: %v\n", key, value)
		}
	}

	// Example 3: Pair a new Z-Wave device
	fmt.Println("\n=== Pairing Z-Wave Device ===")
	fmt.Println("Put your Z-Wave device in inclusion mode...")
	fmt.Println("Press Enter when ready")
	// fmt.Scanln() // Uncomment to actually wait for user input

	pairingResult, err := zwaveAdapter.StartPairing(ctx, 60*time.Second)
	if err != nil {
		log.Printf("Pairing failed: %v", err)
	} else if pairingResult.Success {
		fmt.Printf("Device paired successfully: %s\n", pairingResult.Device.Name)
		fmt.Printf("Device ID: %s\n", pairingResult.Device.ID)
		fmt.Printf("Node ID: %v\n", pairingResult.Device.Properties["node_id"])

		// Get the device instance
		deviceID := pairingResult.Device.ID
		device, err := zwaveAdapter.GetDevice(ctx, deviceID)
		if err != nil {
			log.Printf("Failed to get device: %v", err)
		} else {
			// Example 4: Read from device
			fmt.Println("\n=== Reading from Device ===")
			if value, err := device.Read(ctx, "state"); err == nil {
				fmt.Printf("State: %v\n", value)
			}

			if value, err := device.Read(ctx, "battery_level"); err == nil {
				fmt.Printf("Battery: %v%%\n", value)
			}

			// Example 5: Control device (if it's an actuator)
			if device.Type() == iot.DeviceTypeActuator {
				fmt.Println("\n=== Controlling Device ===")

				// Turn on device
				if zwaveDev, ok := device.(*adapters.ZWaveDevice); ok {
					err = zwaveDev.TurnOn(ctx)
					if err == nil {
						fmt.Println("Device turned on")
					}

					// Wait a bit
					time.Sleep(2 * time.Second)

					// Turn off device
					err = zwaveDev.TurnOff(ctx)
					if err == nil {
						fmt.Println("Device turned off")
					}
				}
			}

			// Example 6: Get device information
			fmt.Println("\n=== Device Information ===")
			if zwaveDev, ok := device.(*adapters.ZWaveDevice); ok {
				info, err := zwaveDev.GetNodeInfo(ctx)
				if err == nil {
					fmt.Printf("Node Info: %v\n", info)
				}

				// Get diagnostic info
				diag, err := zwaveDev.GetDiagnosticInfo(ctx)
				if err == nil {
					fmt.Printf("Diagnostic Info: %v\n", diag)
				}
			}

			// Example 7: Batch operations
			fmt.Println("\n=== Batch Operations ===")
			if zwaveDev, ok := device.(*adapters.ZWaveDevice); ok {
				values, err := zwaveDev.BatchRead(ctx, []string{
					"state",
					"battery_level",
					"location",
				})
				if err == nil {
					fmt.Printf("Batch read results: %v\n", values)
				}
			}
		}
	} else {
		fmt.Printf("Pairing failed: %s\n", pairingResult.Error)
	}

	// Example 8: List all devices
	fmt.Println("\n=== All Z-Wave Devices ===")
	allDevices, err := zwaveAdapter.ListDevices(ctx)
	if err != nil {
		log.Printf("Failed to list devices: %v", err)
	} else {
		for _, dev := range allDevices {
			fmt.Printf("- %s (%s) - %s\n",
				dev.Name(),
				dev.ID(),
				dev.Status())
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

	// Example 10: Heal network
	fmt.Println("\n=== Healing Z-Wave Network ===")
	if err := zwaveAdapter.ResetNetwork(ctx); err != nil {
		log.Printf("Network heal failed: %v", err)
	} else {
		fmt.Println("Network healing initiated")
	}

	fmt.Println("\n=== Examples Complete ===")
}
