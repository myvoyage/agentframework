// Package channels provides message routing for multi-channel system
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
	"fmt"
	"regexp"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Router handles message routing based on rules
//
// SOLID - Single Responsibility Principle (SRP):
// Responsible only for routing decisions
//
// SOLID - Open/Closed Principle (OCP):
// New routing actions can be added without modifying the router
type Router struct {
	rules      map[string]*RoutingRule
	rulesByPri []*RoutingRule // Sorted by priority (highest first)
	mu         sync.RWMutex
	tracer     trace.Tracer
	// Rate limiting trackers
	rateTrackers map[string]*rateTracker
}

// RouterConfig represents router configuration
type RouterConfig struct {
	// DefaultRule is applied when no other rules match
	DefaultRule *RoutingRule

	// EnableRateLimiting enables rate limiting per rule
	EnableRateLimiting bool
}

// rateTracker tracks rate limit for a rule
type rateTracker struct {
	count     int
	windowEnd time.Time
	mu        sync.Mutex
}

// NewRouter creates a new message router
func NewRouter(config *RouterConfig) (*Router, error) {
	if config == nil {
		config = &RouterConfig{}
	}

	r := &Router{
		rules:        make(map[string]*RoutingRule),
		rulesByPri:   make([]*RoutingRule, 0),
		tracer:       otel.Tracer("channels-router"),
		rateTrackers: make(map[string]*rateTracker),
	}

	// Add default rule if provided
	if config.DefaultRule != nil {
		if err := r.AddRule(config.DefaultRule); err != nil {
			return nil, err
		}
	}

	return r, nil
}

// AddRule adds a routing rule
func (r *Router) AddRule(rule *RoutingRule) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if rule.ID == "" {
		return fmt.Errorf("rule ID cannot be empty")
	}

	if _, exists := r.rules[rule.ID]; exists {
		return fmt.Errorf("rule %s already exists", rule.ID)
	}

	// Validate rule
	if rule.Pattern != "" {
		if _, err := regexp.Compile(rule.Pattern); err != nil {
			return fmt.Errorf("invalid pattern in rule %s: %w", rule.ID, err)
		}
	}

	r.rules[rule.ID] = rule

	// Insert sorted by priority
	inserted := false
	for i, existing := range r.rulesByPri {
		if rule.Priority > existing.Priority {
			r.rulesByPri = append(r.rulesByPri[:i], append([]*RoutingRule{rule}, r.rulesByPri[i:]...)...)
			inserted = true
			break
		}
	}
	if !inserted {
		r.rulesByPri = append(r.rulesByPri, rule)
	}

	// Initialize rate tracker if rate limiting is enabled
	if rule.RateLimit > 0 {
		r.rateTrackers[rule.ID] = &rateTracker{}
	}

	return nil
}

// RemoveRule removes a routing rule
func (r *Router) RemoveRule(ruleID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.rules, ruleID)
	delete(r.rateTrackers, ruleID)

	// Rebuild priority list
	r.rulesByPri = make([]*RoutingRule, 0, len(r.rules))
	for _, rule := range r.rules {
		r.rulesByPri = append(r.rulesByPri, rule)
	}

	// Sort by priority
	for i := 0; i < len(r.rulesByPri); i++ {
		for j := i + 1; j < len(r.rulesByPri); j++ {
			if r.rulesByPri[j].Priority > r.rulesByPri[i].Priority {
				r.rulesByPri[i], r.rulesByPri[j] = r.rulesByPri[j], r.rulesByPri[i]
			}
		}
	}
}

// GetRule returns a rule by ID
func (r *Router) GetRule(ruleID string) (*RoutingRule, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	rule, exists := r.rules[ruleID]
	if !exists {
		return nil, fmt.Errorf("rule %s not found", ruleID)
	}

	return rule, nil
}

// ListRules returns all rules
func (r *Router) ListRules() []*RoutingRule {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*RoutingRule, len(r.rulesByPri))
	copy(result, r.rulesByPri)
	return result
}

// Route routes a message through the routing rules
func (r *Router) Route(ctx context.Context, msg *Message, sender MessageSender) error {
	ctx, span := r.tracer.Start(ctx, "Router.Route",
		trace.WithAttributes(
			attribute.String("message_id", msg.ID),
			attribute.String("channel_type", string(msg.ChannelType)),
		),
	)
	defer span.End()

	r.mu.RLock()
	defer r.mu.RUnlock()

	// Try rules in priority order
	for _, rule := range r.rulesByPri {
		matched, err := r.matchRule(ctx, msg, rule)
		if err != nil {
			span.RecordError(err)
			continue
		}

		if !matched {
			continue
		}

		// Check rate limit
		if rule.RateLimit > 0 && !r.checkRateLimit(rule.ID, rule.RateLimit, rule.RateWindow) {
			span.AddEvent("rate_limit_exceeded",
				trace.WithAttributes(attribute.String("rule_id", rule.ID)))
			continue
		}

		// Execute action
		if err := r.executeAction(ctx, msg, rule, sender); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "action execution failed")
			return fmt.Errorf("failed to execute action for rule %s: %w", rule.ID, err)
		}

		span.SetAttributes(attribute.String("matched_rule", rule.ID))
		return nil
	}

	// No rule matched - this is okay, message is simply not routed
	return nil
}

// matchRule checks if a message matches a rule
func (r *Router) matchRule(ctx context.Context, msg *Message, rule *RoutingRule) (bool, error) {
	// Check channel type
	if len(rule.ChannelType) > 0 {
		found := false
		for _, ct := range rule.ChannelType {
			if ct == msg.ChannelType {
				found = true
				break
			}
		}
		if !found {
			return false, nil
		}
	}

	// Check channel ID
	if len(rule.ChannelID) > 0 {
		found := false
		for _, cid := range rule.ChannelID {
			if cid == msg.ChannelID {
				found = true
				break
			}
		}
		if !found {
			return false, nil
		}
	}

	// Check chat ID
	if len(rule.ChatID) > 0 {
		found := false
		for _, chatID := range rule.ChatID {
			if chatID == msg.ChatID {
				found = true
				break
			}
		}
		if !found {
			return false, nil
		}
	}

	// Check user ID
	if len(rule.UserID) > 0 && msg.From != nil {
		found := false
		for _, uid := range rule.UserID {
			if uid == msg.From.ID || uid == msg.From.ChannelUserID {
				found = true
				break
			}
		}
		if !found {
			return false, nil
		}
	}

	// Check message type
	if len(rule.MessageType) > 0 {
		found := false
		for _, mt := range rule.MessageType {
			if mt == msg.Type {
				found = true
				break
			}
		}
		if !found {
			return false, nil
		}
	}

	// Check pattern
	if rule.Pattern != "" && msg.Text != "" {
		matched, err := regexp.MatchString(rule.Pattern, msg.Text)
		if err != nil {
			return false, fmt.Errorf("pattern match failed: %w", err)
		}
		if !matched {
			return false, nil
		}
	}

	// Check custom predicate
	if rule.Predicate != "" {
		// Custom predicates would be registered separately
		// This is a placeholder for future extension
	}

	return true, nil
}

// executeAction executes the action for a matched rule
func (r *Router) executeAction(ctx context.Context, msg *Message, rule *RoutingRule, sender MessageSender) error {
	switch rule.Action {
	case RoutingActionAccept:
		// Message is accepted for normal processing
		return nil

	case RoutingActionReject:
		// Message is rejected
		return fmt.Errorf("message rejected by rule %s", rule.ID)

	case RoutingActionForward:
		// Forward to another channel
		targetChannelID, ok := rule.ActionData["target_channel"]
		if !ok {
			return fmt.Errorf("forward action missing target_channel")
		}

		// Create a copy of the message for forwarding
		forwardMsg := *msg
		forwardMsg.ID = "" // Will get new ID
		forwardMsg.ChannelID = targetChannelID

		opts := MessageSendOptions{}
		if _, err := sender.SendMessage(ctx, targetChannelID, &forwardMsg, opts); err != nil {
			return fmt.Errorf("failed to forward message: %w", err)
		}

	case RoutingActionTransform:
		// Transform message before processing
		// This would require a transform function to be registered
		return fmt.Errorf("transform action not implemented")

	case RoutingActionDelay:
		// Delay processing of the message
		delayStr, ok := rule.ActionData["delay_ms"]
		if !ok {
			return fmt.Errorf("delay action missing delay_ms")
		}

		delay, err := time.ParseDuration(delayStr + "ms")
		if err != nil {
			return fmt.Errorf("invalid delay duration: %w", err)
		}

		select {
		case <-time.After(delay):
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}

	default:
		return fmt.Errorf("unknown action: %s", rule.Action)
	}

	return nil
}

// checkRateLimit checks if a rule has exceeded its rate limit
func (r *Router) checkRateLimit(ruleID string, limit int, window time.Duration) bool {
	tracker, exists := r.rateTrackers[ruleID]
	if !exists {
		return true
	}

	tracker.mu.Lock()
	defer tracker.mu.Unlock()

	now := time.Now()

	// Reset if window expired
	if now.After(tracker.windowEnd) {
		tracker.count = 1
		tracker.windowEnd = now.Add(window)
		return true
	}

	// Check limit
	if tracker.count >= limit {
		return false
	}

	tracker.count++
	return true
}

// MessageSender is an interface for sending messages
// This allows the router to send messages without depending on the Manager
type MessageSender interface {
	SendMessage(ctx context.Context, channelID string, msg *Message, opts MessageSendOptions) (string, error)
}

// ClearRules removes all rules
func (r *Router) ClearRules() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.rules = make(map[string]*RoutingRule)
	r.rulesByPri = make([]*RoutingRule, 0)
	r.rateTrackers = make(map[string]*rateTracker)
}

// UpdateRule updates an existing rule
func (r *Router) UpdateRule(rule *RoutingRule) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.rules[rule.ID]; !exists {
		return fmt.Errorf("rule %s not found", rule.ID)
	}

	// Remove old rule
	delete(r.rules, rule.ID)

	// Add updated rule
	r.rules[rule.ID] = rule

	// Rebuild priority list
	r.rulesByPri = make([]*RoutingRule, 0, len(r.rules))
	for _, existingRule := range r.rules {
		r.rulesByPri = append(r.rulesByPri, existingRule)
	}

	// Sort by priority
	for i := 0; i < len(r.rulesByPri); i++ {
		for j := i + 1; j < len(r.rulesByPri); j++ {
			if r.rulesByPri[j].Priority > r.rulesByPri[i].Priority {
				r.rulesByPri[i], r.rulesByPri[j] = r.rulesByPri[j], r.rulesByPri[i]
			}
		}
	}

	return nil
}

// EnableRule enables or disables a rule
func (r *Router) EnableRule(ruleID string, enabled bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	rule, exists := r.rules[ruleID]
	if !exists {
		return fmt.Errorf("rule %s not found", ruleID)
	}

	// Update metadata
	if rule.Metadata == nil {
		rule.Metadata = make(map[string]string)
	}

	if enabled {
		delete(rule.Metadata, "disabled")
	} else {
		rule.Metadata["disabled"] = "true"
	}

	return nil
}

// IsRuleEnabled checks if a rule is enabled
func (r *Router) IsRuleEnabled(ruleID string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	rule, exists := r.rules[ruleID]
	if !exists {
		return false
	}

	if rule.Metadata == nil {
		return true
	}

	_, disabled := rule.Metadata["disabled"]
	return !disabled
}
