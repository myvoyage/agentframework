// Package channels provides channel adapter interfaces
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package channels

import (
	"context"
	"io"
	"sync"
)

// ChannelAdapter defines the interface for channel adapters
//
// SOLID - Interface Segregation Principle (ISP):
// Focused interface with only essential methods for channel operations
//
// SOLID - Open/Closed Principle (OCP):
// New channels can be added by implementing this interface without modifying existing code
type ChannelAdapter interface {
	// Lifecycle management

	// Initialize initializes the channel adapter with configuration
	Initialize(ctx context.Context, config ChannelConfig) error

	// Connect establishes connection to the channel platform
	Connect(ctx context.Context) error

	// Disconnect gracefully closes the channel connection
	Disconnect(ctx context.Context) error

	// IsConnected returns the current connection status
	IsConnected() bool

	// GetStatus returns detailed status information
	GetStatus(ctx context.Context) (ChannelStatus, error)

	// GetStats returns statistics about the channel
	GetStats(ctx context.Context) (*ChannelStats, error)

	// Messaging operations

	// SendMessage sends a message to the channel
	// Returns the message ID assigned by the platform
	SendMessage(ctx context.Context, msg *Message, opts MessageSendOptions) (string, error)

	// EditMessage edits an existing message
	EditMessage(ctx context.Context, messageID string, msg *Message) error

	// DeleteMessage deletes a message
	DeleteMessage(ctx context.Context, messageID string) error

	// UploadFile uploads a file and returns an attachment
	UploadFile(ctx context.Context, filename string, content io.Reader, mimeType string) (*Attachment, error)

	// Event handling

	// SetMessageHandler sets the handler for incoming messages
	SetMessageHandler(handler MessageHandler)

	// SetEventHandler sets the handler for channel events
	SetEventHandler(handler EventHandler)

	// GetChannelID returns the unique identifier for this channel instance
	GetChannelID() string

	// GetChannelType returns the type of this channel
	GetChannelType() ChannelType

	// Capabilities

	// Supports returns true if the channel supports a specific feature
	Supports(feature string) bool
}

// ChannelCapability defines supported features
type ChannelCapability string

const (
	// CapabilityThreads indicates support for threaded messages
	CapabilityThreads ChannelCapability = "threads"
	// CapabilityEdits indicates support for editing messages
	CapabilityEdits ChannelCapability = "edits"
	// CapabilityReactions indicates support for message reactions
	CapabilityReactions ChannelCapability = "reactions"
	// CapabilityTypingIndicator indicates support for typing indicators
	CapabilityTypingIndicator ChannelCapability = "typing_indicator"
	// CapabilityReadReceipt indicates support for read receipts
	CapabilityReadReceipt ChannelCapability = "read_receipt"
	// CapabilityWebhooks indicates support for webhook mode
	CapabilityWebhooks ChannelCapability = "webhooks"
	// CapabilityPolling indicates support for polling mode
	CapabilityPolling ChannelCapability = "polling"
	// CapabilityRichText indicates support for rich text formatting
	CapabilityRichText ChannelCapability = "rich_text"
	// CapabilityInlineKeyboard indicates support for inline keyboards/buttons
	CapabilityInlineKeyboard ChannelCapability = "inline_keyboard"
)

// BaseAdapter provides common functionality for all adapters
//
// SOLID - DRY Principle:
// Common functionality is implemented once and reused by all adapters
//
// SOLID - Template Method Pattern:
// Provides default implementations that can be overridden
type BaseAdapter struct {
	config         ChannelConfig
	channelID      string
	channelType    ChannelType
	messageHandler MessageHandler
	eventHandler   EventHandler
	status         ChannelStatus
	capabilities   map[ChannelCapability]bool
}

// NewBaseAdapter creates a new base adapter
func NewBaseAdapter(channelID string, channelType ChannelType) *BaseAdapter {
	return &BaseAdapter{
		channelID:    channelID,
		channelType:  channelType,
		status:       ChannelStatusDisconnected,
		capabilities: make(map[ChannelCapability]bool),
	}
}

// GetChannelID returns the channel ID
func (b *BaseAdapter) GetChannelID() string {
	return b.channelID
}

// GetChannelType returns the channel type
func (b *BaseAdapter) GetChannelType() ChannelType {
	return b.channelType
}

// SetMessageHandler sets the message handler
func (b *BaseAdapter) SetMessageHandler(handler MessageHandler) {
	b.messageHandler = handler
}

// SetEventHandler sets the event handler
func (b *BaseAdapter) SetEventHandler(handler EventHandler) {
	b.eventHandler = handler
}

// IsConnected returns the connection status
func (b *BaseAdapter) IsConnected() bool {
	return b.status == ChannelStatusConnected
}

// Supports checks if a capability is supported
func (b *BaseAdapter) Supports(feature string) bool {
	cap := ChannelCapability(feature)
	supported, exists := b.capabilities[cap]
	return exists && supported
}

// SetCapability sets a capability
func (b *BaseAdapter) SetCapability(capability ChannelCapability, supported bool) {
	b.capabilities[capability] = supported
}

// setStatus updates the adapter status and emits an event
func (b *BaseAdapter) setStatus(ctx context.Context, status ChannelStatus) {
	oldStatus := b.status
	b.status = status

	if b.eventHandler != nil {
		event := Event{
			Type:        EventType(status),
			ChannelID:   b.channelID,
			ChannelType: b.channelType,
		}

		if status == ChannelStatusError {
			event.Type = EventTypeError
		}

		// Emit event in a non-blocking way
		go b.eventHandler(event)
	}

	_ = ctx // Avoid unused parameter warning
	_ = oldStatus
}

// handleMessage processes an incoming message through the registered handler
func (b *BaseAdapter) handleMessage(ctx context.Context, msg *Message) error {
	if b.messageHandler == nil {
		return nil // No handler registered, ignore message
	}

	// Set channel info if not already set
	if msg.ChannelID == "" {
		msg.ChannelID = b.channelID
	}
	if msg.ChannelType == "" {
		msg.ChannelType = b.channelType
	}

	return b.messageHandler(msg)
}

// emitEvent emits a channel event
func (b *BaseAdapter) emitEvent(ctx context.Context, eventType EventType, data map[string]interface{}, err error) {
	if b.eventHandler == nil {
		return
	}

	event := Event{
		Type:        eventType,
		ChannelID:   b.channelID,
		ChannelType: b.channelType,
		Data:        data,
		Error:       err,
	}

	// Emit event in a non-blocking way
	go b.eventHandler(event)
}

// AdapterFactory creates channel adapters
//
// SOLID - Factory Pattern:
// Encapsulates adapter creation logic
//
// SOLID - Open/Closed Principle:
// New adapter types can be added by registering them without modifying factory code
type AdapterFactory interface {
	// CreateAdapter creates a new channel adapter
	CreateAdapter(channelID string, channelType ChannelType) (ChannelAdapter, error)

	// SupportedTypes returns the list of supported channel types
	SupportedTypes() []ChannelType
}

// DefaultAdapterFactory is the default factory implementation
type DefaultAdapterFactory struct {
	adapters map[ChannelType]func(string) ChannelAdapter
}

// NewAdapterFactory creates a new adapter factory
// Note: Built-in adapters need to be registered separately using RegisterAdapter
// to avoid import cycles. See pkg/channels/adapters for built-in adapter implementations.
func NewAdapterFactory() *DefaultAdapterFactory {
	factory := &DefaultAdapterFactory{
		adapters: make(map[ChannelType]func(string) ChannelAdapter),
	}

	// Built-in adapters are registered in adapters package init() functions
	// to avoid circular import dependencies.
	// Import the adapters package to enable built-in adapters:
	//   import _ "AgentFramework/pkg/channels/adapters"

	return factory
}

// RegisterAdapter registers a new adapter type
func (f *DefaultAdapterFactory) RegisterAdapter(channelType ChannelType, constructor func(string) ChannelAdapter) {
	f.adapters[channelType] = constructor
}

// CreateAdapter creates a new channel adapter
func (f *DefaultAdapterFactory) CreateAdapter(channelID string, channelType ChannelType) (ChannelAdapter, error) {
	constructor, exists := f.adapters[channelType]
	if !exists {
		return nil, ErrUnsupportedChannelType
	}

	return constructor(channelID), nil
}

// SupportedTypes returns the list of supported channel types
func (f *DefaultAdapterFactory) SupportedTypes() []ChannelType {
	types := make([]ChannelType, 0, len(f.adapters))
	for t := range f.adapters {
		types = append(types, t)
	}
	return types
}

// Error definitions
var (
	// ErrUnsupportedChannelType is returned when an unsupported channel type is requested
	ErrUnsupportedChannelType = &ChannelError{
		Code:    "UNSUPPORTED_CHANNEL_TYPE",
		Message: "the requested channel type is not supported",
	}

	// ErrNotConnected is returned when trying to send a message while disconnected
	ErrNotConnected = &ChannelError{
		Code:    "NOT_CONNECTED",
		Message: "the channel is not connected",
	}

	// ErrInvalidConfiguration is returned when the channel configuration is invalid
	ErrInvalidConfiguration = &ChannelError{
		Code:    "INVALID_CONFIGURATION",
		Message: "the channel configuration is invalid",
	}

	// ErrRateLimited is returned when the rate limit is exceeded
	ErrRateLimited = &ChannelError{
		Code:    "RATE_LIMITED",
		Message: "the rate limit has been exceeded",
	}

	// ErrMessageTooLarge is returned when a message is too large
	ErrMessageTooLarge = &ChannelError{
		Code:    "MESSAGE_TOO_LARGE",
		Message: "the message exceeds the maximum size limit",
	}

	// ErrUnsupportedMessageType is returned when a message type is not supported
	ErrUnsupportedMessageType = &ChannelError{
		Code:    "UNSUPPORTED_MESSAGE_TYPE",
		Message: "the message type is not supported by this channel",
	}
)

// ChannelError represents a channel-specific error
type ChannelError struct {
	Code    string
	Message string
	Err     error
}

// Error returns the error message
func (e *ChannelError) Error() string {
	if e.Err != nil {
		return e.Code + ": " + e.Message + ": " + e.Err.Error()
	}
	return e.Code + ": " + e.Message
}

// Unwrap returns the underlying error
func (e *ChannelError) Unwrap() error {
	return e.Err
}

// builtinRegistry holds the global registry for built-in adapters
// This allows adapters to register themselves without creating circular imports
var builtinRegistry = struct {
	sync.RWMutex
	constructors map[ChannelType]func(string) ChannelAdapter
}{
	constructors: make(map[ChannelType]func(string) ChannelAdapter),
}

// RegisterBuiltinAdapter registers a built-in adapter constructor
// This function is called by adapters in their init() functions
func RegisterBuiltinAdapter(channelType ChannelType, constructor func(string) ChannelAdapter) {
	builtinRegistry.Lock()
	defer builtinRegistry.Unlock()
	builtinRegistry.constructors[channelType] = constructor
}

// RegisterBuiltinAdaptersToFactory registers all built-in adapters to a factory instance
// This should be called after creating a factory to enable built-in adapter support
func RegisterBuiltinAdaptersToFactory(factory *DefaultAdapterFactory) {
	builtinRegistry.RLock()
	defer builtinRegistry.RUnlock()
	
	for channelType, constructor := range builtinRegistry.constructors {
		factory.RegisterAdapter(channelType, constructor)
	}
}
