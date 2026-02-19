// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later


// Example: GPIO Usage
// This example demonstrates how to use the GPIO driver to control pins.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"AgentFramework/pkg/beads/hardware/drivers"
)

func main() {
	ctx := context.Background()

	// Create GPIO driver configuration
	config := &drivers.GPIODeviceConfig{
		Chip:     "gpiochip0", // GPIO chip name
		PinCount: 28,          // Number of pins
	}

	// Create GPIO driver
	driver := drivers.NewGPIODriver(config)

	// Connect to GPIO device
	fmt.Println("Connecting to GPIO device...")
	if err := driver.Connect(ctx, nil); err != nil {
		log.Fatalf("Failed to connect to GPIO: %v", err)
	}
	defer driver.Disconnect(ctx)

	fmt.Println("Connected to GPIO device successfully!")

	// Get device status
	status, err := driver.GetStatus(ctx)
	if err != nil {
		log.Printf("Failed to get status: %v", err)
	} else {
		fmt.Printf("GPIO Status: %+v\n", status)
	}

	// Example 1: Configure pin as output and toggle
	fmt.Println("\n--- Pin Output Example ---")
	pinNumber := float64(17) // GPIO17 (commonly available on Raspberry Pi)

	// Setup pin as output
	setupParams := map[string]interface{}{
		"pin":       pinNumber,
		"direction": "out",
	}

	result, err := driver.SendCommand(ctx, "setup_pin", setupParams)
	if err != nil {
		log.Printf("Failed to setup pin: %v", err)
	} else {
		fmt.Printf("Pin configured: %+v\n", result)
	}

	// Toggle pin 5 times
	fmt.Println("Toggling pin 5 times...")
	for i := 0; i < 5; i++ {
		// Write HIGH
		writeParams := map[string]interface{}{
			"pin":   pinNumber,
			"value": float64(1),
		}
		_, err := driver.SendCommand(ctx, "write_pin", writeParams)
		if err != nil {
			log.Printf("Failed to write HIGH: %v", err)
		}
		fmt.Print("HIGH ")
		time.Sleep(500 * time.Millisecond)

		// Write LOW
		writeParams["value"] = float64(0)
		_, err = driver.SendCommand(ctx, "write_pin", writeParams)
		if err != nil {
			log.Printf("Failed to write LOW: %v", err)
		}
		fmt.Print("LOW ")
		time.Sleep(500 * time.Millisecond)
	}
	fmt.Println()

	// Example 2: Configure pin as input and read
	fmt.Println("\n--- Pin Input Example ---")
	inputPin := float64(18)

	// Setup pin as input
	setupParams = map[string]interface{}{
		"pin":       inputPin,
		"direction": "in",
		"pull":      "up", // Enable pull-up resistor
	}

	result, err = driver.SendCommand(ctx, "setup_pin", setupParams)
	if err != nil {
		log.Printf("Failed to setup input pin: %v", err)
	} else {
		fmt.Printf("Input pin configured: %+v\n", result)
	}

	// Read pin value
	readParams := map[string]interface{}{
		"pin": inputPin,
	}

	readResult, err := driver.SendCommand(ctx, "read_pin", readParams)
	if err != nil {
		log.Printf("Failed to read pin: %v", err)
	} else {
		if gpioResult, ok := readResult.(*drivers.GPIOReadResult); ok {
			fmt.Printf("Pin value: %d (timestamp: %d)\n", gpioResult.Value, gpioResult.Timestamp)
		}
	}

	// Example 3: Configure pin with edge detection
	fmt.Println("\n--- Edge Detection Example ---")
	eventPin := float64(22)

	// Setup pin as input with edge detection
	setupParams = map[string]interface{}{
		"pin":       eventPin,
		"direction": "in",
		"edge":      "both", // Detect both rising and falling edges
		"pull":      "down",
	}

	result, err = driver.SendCommand(ctx, "setup_pin", setupParams)
	if err != nil {
		log.Printf("Failed to setup event pin: %v", err)
	} else {
		fmt.Printf("Event pin configured: %+v\n", result)
	}

	// Try to receive events
	fmt.Println("Waiting for edge events (5 seconds)...")
	done := make(chan bool)
	go func() {
		for i := 0; i < 3; i++ {
			data, err := driver.ReceiveData(ctx, 2*time.Second)
			if err != nil {
				fmt.Printf("No event received: %v\n", err)
				continue
			}
			if event, ok := data.(drivers.GPIOEvent); ok {
				fmt.Printf("Event received: Pin=%d, Value=%d, Type=%s\n",
					event.Pin, event.Value, event.Type)
			}
		}
		done <- true
	}()

	select {
	case <-done:
		fmt.Println("Event monitoring completed")
	case <-time.After(6 * time.Second):
		fmt.Println("Event monitoring timeout")
	}

	// Example 4: Configure PWM
	fmt.Println("\n--- PWM Example ---")
	pwmPin := float64(12)

	pwmParams := map[string]interface{}{
		"pin":        pwmPin,
		"period":     float64(20000),   // 20ms period (50Hz)
		"duty_cycle": float64(10000),   // 50% duty cycle (10ms)
	}

	result, err = driver.SendCommand(ctx, "set_pwm", pwmParams)
	if err != nil {
		log.Printf("Failed to set PWM: %v", err)
	} else {
		fmt.Printf("PWM configured: %+v\n", result)
	}

	// Example 5: List configured pins
	fmt.Println("\n--- List Configured Pins ---")
	result, err = driver.SendCommand(ctx, "list_pins", nil)
	if err != nil {
		log.Printf("Failed to list pins: %v", err)
	} else {
		fmt.Printf("Configured pins: %+v\n", result)
	}

	// Example 6: Get pin info
	fmt.Println("\n--- Get Pin Info ---")
	infoParams := map[string]interface{}{
		"pin": pinNumber,
	}

	result, err = driver.SendCommand(ctx, "get_pin_info", infoParams)
	if err != nil {
		log.Printf("Failed to get pin info: %v", err)
	} else {
		fmt.Printf("Pin info: %+v\n", result)
	}

	fmt.Println("\n--- GPIO Example Completed ---")
}
