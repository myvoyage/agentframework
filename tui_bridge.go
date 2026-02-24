// Agent Framework - TUI Bridge
// Copyright (C) 2025 Agent Framework Contributors
//
// SPDX-License-Identifier: AGPL-3.0-or-later

//go:build !tui_standalone

package main

import (
	"fmt"
	"os"
	"os/exec"
)

// runTUI executes the TUI application as a subprocess
// This avoids circular dependencies between main and tui packages
func runTUI() error {
	// Get the path to the current executable
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Create a command to run the TUI
	// We use the tui.exe if it exists, otherwise build and run it
	tuiExe := exePath + "-tui"
	if _, err := os.Stat(tuiExe); os.IsNotExist(err) {
		// Try to find tui.exe in the same directory
		tuiExe = "tui.exe"
	}

	// Execute the TUI
	cmd := exec.Command(tuiExe)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("TUI exited with error: %w", err)
	}

	return nil
}
