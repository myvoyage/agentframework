//go:build linux
// +build linux

package drivers

import (
	"runtime"
	"testing"
)

// TestLinuxGPIOChip tests the Linux GPIO chip implementation.
func TestLinuxGPIOChip(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Skipping Linux GPIO test on non-Linux platform")
	}

	// This test may fail if no GPIO chip is available
	// It's mainly for checking that the code compiles
	chip, err := NewLinuxGPIOChip("gpiochip0")
	if err != nil {
		t.Skipf("No GPIO chip available: %v", err)
	}

	if chip == nil {
		t.Fatal("NewLinuxGPIOChip returned nil")
	}

	// Just close it
	chip.Close()
}
