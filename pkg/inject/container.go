// Agent Framework - Dependency Injection Container
// Copyright (C) 2025  Agent Framework Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package inject

import (
	"fmt"
	"reflect"
	"sync"
)

// Provider is a function that creates an instance
type Provider func() (interface{}, error)

// Container manages dependency injection
type Container struct {
	providers  map[string]Provider
	singletons map[string]interface{}
	mu         sync.RWMutex
}

// New creates a new dependency injection container
func New() *Container {
	return &Container{
		providers:  make(map[string]Provider),
		singletons: make(map[string]interface{}),
	}
}

// Register registers a provider function for a service
func (c *Container) Register(name string, provider Provider) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.providers[name]; exists {
		return fmt.Errorf("provider %s already registered", name)
	}

	c.providers[name] = provider
	return nil
}

// RegisterSingleton registers a singleton instance
func (c *Container) RegisterSingleton(name string, instance interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.singletons[name] = instance
}

// Resolve resolves a dependency by name
func (c *Container) Resolve(name string) (interface{}, error) {
	c.mu.RLock()
	
	// Check singletons first
	if instance, ok := c.singletons[name]; ok {
		c.mu.RUnlock()
		return instance, nil
	}

	// Check providers
	provider, ok := c.providers[name]
	c.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("service %s not found", name)
	}

	// Call provider to create instance
	instance, err := provider()
	if err != nil {
		return nil, fmt.Errorf("failed to create %s: %w", name, err)
	}

	return instance, nil
}

// ResolveTo resolves a dependency and assigns it to the target
func (c *Container) ResolveTo(name string, target interface{}) error {
	instance, err := c.Resolve(name)
	if err != nil {
		return err
	}

	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr {
		return fmt.Errorf("target must be a pointer")
	}

	instanceValue := reflect.ValueOf(instance)
	targetType := targetValue.Elem().Type()
	instanceType := instanceValue.Type()

	// Check if types are compatible
	if !instanceType.AssignableTo(targetType) {
		// Try to convert if possible
		if instanceValue.CanConvert(targetType) {
			instanceValue = instanceValue.Convert(targetType)
		} else {
			return fmt.Errorf("type mismatch: %s cannot be assigned to %s", instanceType, targetType)
		}
	}

	targetValue.Elem().Set(instanceValue)
	return nil
}

// MustResolve resolves a dependency and panics on error
func (c *Container) MustResolve(name string) interface{} {
	instance, err := c.Resolve(name)
	if err != nil {
		panic(fmt.Sprintf("failed to resolve %s: %v", name, err))
	}
	return instance
}

// IsRegistered checks if a service is registered
func (c *Container) IsRegistered(name string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	_, hasSingleton := c.singletons[name]
	_, hasProvider := c.providers[name]
	return hasSingleton || hasProvider
}

// Unregister removes a service registration
func (c *Container) Unregister(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	delete(c.providers, name)
	delete(c.singletons, name)
}

// Clear removes all registrations
func (c *Container) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.providers = make(map[string]Provider)
	c.singletons = make(map[string]interface{})
}

// ListServices returns a list of all registered service names
func (c *Container) ListServices() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	services := make(map[string]bool)
	for name := range c.providers {
		services[name] = true
	}
	for name := range c.singletons {
		services[name] = true
	}
	
	result := make([]string, 0, len(services))
	for name := range services {
		result = append(result, name)
	}
	
	return result
}
