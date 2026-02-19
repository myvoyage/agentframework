// Agent Framework - Dependency Injection Example Usage
// Copyright (C) 2025  Agent Framework Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package inject

import (
	"fmt"
)

// ExampleService demonstrates how to use the dependency injection container
func ExampleService() {
	// Create container
	container := New()

	// Register a singleton (configuration)
	container.RegisterSingleton("config", map[string]string{
		"log_level": "info",
		"env":       "production",
	})

	// Register a provider function (logger)
	container.Register("logger", func() (interface{}, error) {
		config, _ := container.Resolve("config")
		cfg := config.(map[string]string)
		return fmt.Sprintf("Logger initialized with level: %s", cfg["log_level"]), nil
	})

	// Register a service that depends on logger
	container.Register("service", func() (interface{}, error) {
		logger, err := container.Resolve("logger")
		if err != nil {
			return nil, err
		}
		return fmt.Sprintf("Service using %v", logger), nil
	})

	// Resolve dependencies
	service, err := container.Resolve("service")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Resolved service: %v\n", service)
}
