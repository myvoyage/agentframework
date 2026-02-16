// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
// Copyright (C) 2025  Agent Framework Contributors

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Aggero General Public License for more details.

// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

// Additional permission under GNU Affero General Public License version 3 section 7
// If you modify this Program, or any covered work, by linking or combining it
// with other code, such other code is not for that reason alone subject to any
// of the requirements of the GNU Affero GPL version 3 as long as you maintain
// the separation between the Program and the other code.

// For network interaction purposes, when this Program is used over a network,
// the source code of the Program must be made available to users of the network.
// You can comply with this requirement by providing a link to the source code
// repository in your user interface or documentation.

// SPDX-License-Identifier: AGPL-3.0-or-later

package agent

import (
	"context"
	"sync"
)

// LoggerProvider manages logger instances for dependency injection
// This eliminates the need for global logger variables
// It provides thread-safe access to logger instances

type LoggerProvider struct {
	mu     sync.RWMutex
	logger Logger
}

// NewLoggerProvider creates a new logger provider with default logger
func NewLoggerProvider() *LoggerProvider {
	return &LoggerProvider{
		logger: NewLogger(DefaultLoggerConfig()),
	}
}

// NewLoggerProviderWithLogger creates a new logger provider with specified logger
func NewLoggerProviderWithLogger(logger Logger) *LoggerProvider {
	return &LoggerProvider{
		logger: logger,
	}
}

// GetLogger returns the current logger instance
func (p *LoggerProvider) GetLogger() Logger {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.logger
}

// SetLogger sets the logger instance (thread-safe)
func (p *LoggerProvider) SetLogger(logger Logger) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.logger = logger
}

// WithLogger returns a new context with the logger provider
func (p *LoggerProvider) WithLogger(ctx context.Context) context.Context {
	return context.WithValue(ctx, loggerProviderKey{}, p)
}

// FromContext retrieves the logger provider from context
func LoggerFromContext(ctx context.Context) *LoggerProvider {
	if provider, ok := ctx.Value(loggerProviderKey{}).(*LoggerProvider); ok {
		return provider
	}
	// Return a default provider if not found in context
	return NewLoggerProvider()
}

// LoggerFromCtx is a shorter alias for LoggerFromContext
func LoggerFromCtx(ctx context.Context) *LoggerProvider {
	return LoggerFromContext(ctx)
}

// Get returns the logger from context (convenience method)
func Get(ctx context.Context) Logger {
	return LoggerFromContext(ctx).GetLogger()
}

// loggerProviderKey is a private key for context value storage
type loggerProviderKey struct{}

// DefaultLoggerProvider is the default logger provider instance
// This is used for backward compatibility during migration
// TODO: Remove this once all code is migrated to dependency injection
var DefaultLoggerProvider = NewLoggerProvider()