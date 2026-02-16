// Example: CAN Bus Usage
// This example demonstrates how to use the CAN bus driver to send and receive messages.
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

	// Create CAN driver configuration
	config := &drivers.CANDeviceConfig{
		Interface: "can0",    // CAN interface name
		BaudRate:  500000,    // 500K baud rate
		Timeout:   5000,      // 5 second timeout
		EnableFD:  false,     // Standard CAN (not CAN FD)
	}

	// Create CAN driver
	driver := drivers.NewCANDriver(config)

	// Connect to CAN bus
	fmt.Println("Connecting to CAN bus...")
	if err := driver.Connect(ctx, nil); err != nil {
		log.Fatalf("Failed to connect to CAN bus: %v", err)
	}
	defer driver.Disconnect(ctx)

	fmt.Println("Connected to CAN bus successfully!")

	// Get device status
	status, err := driver.GetStatus(ctx)
	if err != nil {
		log.Printf("Failed to get status: %v", err)
	} else {
		fmt.Printf("CAN Status: %+v\n", status)
	}

	// Example 1: Send a standard CAN frame
	fmt.Println("\n--- Sending Standard Frame ---")
	sendParams := map[string]interface{}{
		"id":          uint32(0x123), // Standard 11-bit ID
		"data":        []byte{0x01, 0x02, 0x03, 0x04, 0x05},
		"is_extended": false,
		"is_remote":   false,
	}

	result, err := driver.SendCommand(ctx, "send_frame", sendParams)
	if err != nil {
		log.Printf("Failed to send frame: %v", err)
	} else {
		fmt.Printf("Frame sent successfully: %+v\n", result)
	}

	// Example 2: Send an extended CAN frame
	fmt.Println("\n--- Sending Extended Frame ---")
	extendedParams := map[string]interface{}{
		"id":          uint32(0x12345678), // Extended 29-bit ID
		"data":        []byte{0xAA, 0xBB, 0xCC},
		"is_extended": true,
		"is_remote":   false,
	}

	result, err = driver.SendCommand(ctx, "send_frame", extendedParams)
	if err != nil {
		log.Printf("Failed to send extended frame: %v", err)
	} else {
		fmt.Printf("Extended frame sent successfully: %+v\n", result)
	}

	// Example 3: Send multiple frames in batch
	fmt.Println("\n--- Sending Batch Frames ---")
	batchParams := map[string]interface{}{
		"frames": []interface{}{
			map[string]interface{}{
				"id":   uint32(0x100),
				"data": []byte{0x01},
			},
			map[string]interface{}{
				"id":   uint32(0x101),
				"data": []byte{0x02},
			},
			map[string]interface{}{
				"id":   uint32(0x102),
				"data": []byte{0x03},
			},
		},
	}

	result, err = driver.SendCommand(ctx, "send_batch", batchParams)
	if err != nil {
		log.Printf("Failed to send batch: %v", err)
	} else {
		fmt.Printf("Batch sent successfully: %+v\n", result)
	}

	// Example 4: Set CAN filters
	fmt.Println("\n--- Setting CAN Filters ---")
	filterParams := map[string]interface{}{
		"id":   uint32(0x100),
		"mask": uint32(0x700), // Filter for IDs 0x100-0x1FF
	}

	result, err = driver.SendCommand(ctx, "set_filter", filterParams)
	if err != nil {
		log.Printf("Failed to set filter: %v", err)
	} else {
		fmt.Printf("Filter set successfully: %+v\n", result)
	}

	// Example 5: Receive CAN frames
	fmt.Println("\n--- Receiving CAN Frames ---")
	fmt.Println("Waiting for incoming frames (timeout: 5 seconds)...")

	// Receive with timeout
	data, err := driver.ReceiveData(ctx, 5*time.Second)
	if err != nil {
		fmt.Printf("No frame received: %v\n", err)
	} else {
		if receiveResult, ok := data.(*drivers.CANReceiveResult); ok && receiveResult.Success {
			frame := receiveResult.Frame
			fmt.Printf("Received frame:\n")
			fmt.Printf("  ID: 0x%X\n", frame.ID)
			fmt.Printf("  Extended: %v\n", frame.IsExtended)
			fmt.Printf("  Remote: %v\n", frame.IsRemote)
			fmt.Printf("  Data: %v\n", frame.Data)
		}
	}

	// Example 6: Get CAN statistics
	fmt.Println("\n--- Getting Statistics ---")
	statsParams := map[string]interface{}{}

	result, err = driver.SendCommand(ctx, "get_stats", statsParams)
	if err != nil {
		log.Printf("Failed to get stats: %v", err)
	} else {
		fmt.Printf("CAN Statistics: %+v\n", result)
	}

	fmt.Println("\n--- CAN Example Completed ---")
}
