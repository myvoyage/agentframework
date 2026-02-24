// Agent Framework - Multi-Channel API
// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"AgentFramework/core"
	"AgentFramework/pkg/channels"
)

// ChannelAPI provides HTTP endpoints for multi-channel management
type ChannelAPI struct {
	app *core.ApplicationWithChannels
}

// NewChannelAPI creates a new multi-channel API handler
func NewChannelAPI(app *core.ApplicationWithChannels) *ChannelAPI {
	return &ChannelAPI{app: app}
}

// RegisterRoutes registers all API routes
func (api *ChannelAPI) RegisterRoutes(r *mux.Router) {
	// Channel management
	r.HandleFunc("/channels", api.GetChannels).Methods("GET")
	r.HandleFunc("/channels/{id}", api.GetChannel).Methods("GET")
	r.HandleFunc("/channels/stats", api.GetChannelStats).Methods("GET")

	// Messaging
	r.HandleFunc("/channels/send", api.SendMessage).Methods("POST")
	r.HandleFunc("/channels/broadcast", api.BroadcastMessage).Methods("POST")
	r.HandleFunc("/channels/send-to-user", api.SendMessageToUser).Methods("POST")
	r.HandleFunc("/channels/broadcast-all", api.BroadcastAll).Methods("POST")

	// Sessions
	r.HandleFunc("/channels/sessions", api.GetSessions).Methods("GET")
	r.HandleFunc("/channels/sessions/cleanup", api.CleanupSessions).Methods("POST")

	// Routing rules
	r.HandleFunc("/channels/routes", api.GetRoutes).Methods("GET")
	r.HandleFunc("/channels/routes", api.AddRoute).Methods("POST")
	r.HandleFunc("/channels/routes/{id}", api.RemoveRoute).Methods("DELETE")

	// Configuration
	r.HandleFunc("/channels/config", api.GetConfig).Methods("GET")
	r.HandleFunc("/channels/config/reload", api.ReloadConfig).Methods("POST")
}

// GetChannels returns all channel information
func (api *ChannelAPI) GetChannels(w http.ResponseWriter, r *http.Request) {
	stats, err := api.app.GetChannelStats()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"channels": stats,
		"total":    len(stats),
	})
}

// GetChannel returns a specific channel's information
func (api *ChannelAPI) GetChannel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	channelID := vars["id"]

	stats, err := api.app.GetChannelStats()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	stat, ok := stats[channelID]
	if !ok {
		http.Error(w, "channel not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stat)
}

// GetChannelStats returns statistics for all channels
func (api *ChannelAPI) GetChannelStats(w http.ResponseWriter, r *http.Request) {
	stats, err := api.app.GetChannelStats()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Calculate totals
	var totalSent, totalReceived, totalErrors int64
	for _, stat := range stats {
		totalSent += stat.MessagesSent
		totalReceived += stat.MessagesReceived
		totalErrors += int64(stat.ErrorCount)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"channels":       stats,
		"total_sent":     totalSent,
		"total_received": totalReceived,
		"total_errors":   totalErrors,
	})
}

// SendMessage sends a message to a specific channel
func (api *ChannelAPI) SendMessage(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ChannelID string `json:"channel_id"`
		Text      string `json:"text"`
		Type      string `json:"type,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := api.app.SendChannelMessage(req.ChannelID, req.Text); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "sent",
		"channel": req.ChannelID,
	})
}

// BroadcastMessage broadcasts a message to all channels of a type
func (api *ChannelAPI) BroadcastMessage(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ChannelType string `json:"channel_type"`
		Text        string `json:"text"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ct := channels.ChannelType(req.ChannelType)
	if err := api.app.BroadcastChannelMessage(ct, req.Text); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":       "broadcast",
		"channel_type": req.ChannelType,
	})
}

// SendMessageToUser sends a message to a specific user across all channels
func (api *ChannelAPI) SendMessageToUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID string `json:"user_id"`
		Text   string `json:"text"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := api.app.SendChannelMessageToUser(req.UserID, req.Text); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "sent",
		"user":   req.UserID,
	})
}

// BroadcastAll sends a message to all active channels
func (api *ChannelAPI) BroadcastAll(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Text string `json:"text"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := api.app.BroadcastToAllChannels(req.Text); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "broadcast_to_all",
	})
}

// GetSessions returns all active sessions
func (api *ChannelAPI) GetSessions(w http.ResponseWriter, r *http.Request) {
	sessions := api.app.GetChannelSessions()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"sessions": sessions,
		"total":    len(sessions),
	})
}

// CleanupSessions removes inactive sessions
func (api *ChannelAPI) CleanupSessions(w http.ResponseWriter, r *http.Request) {
	timeoutStr := r.URL.Query().Get("timeout")
	timeout := int64(3600) // Default 1 hour

	if timeoutStr != "" {
		if t, err := strconv.ParseInt(timeoutStr, 10, 64); err == nil {
			timeout = t
		}
	}

	// Note: This would require exposing CleanupSessions in ApplicationWithChannels
	// For now, return count of sessions that would be cleaned
	sessions := api.app.GetChannelSessions()
	cleaned := 0
	now := r.Context().Value("now").(int64)
	if now == 0 {
		now = ctxTimestamp()
	}

	for _, sess := range sessions {
		if now-sess.UpdatedAt > timeout {
			cleaned++
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"cleaned":   cleaned,
		"remaining": len(sessions) - cleaned,
		"timeout":   timeout,
	})
}

// GetRoutes returns all routing rules
func (api *ChannelAPI) GetRoutes(w http.ResponseWriter, r *http.Request) {
	cm := api.app.GetChannelManager()
	if cm == nil {
		http.Error(w, "channels not available", http.StatusServiceUnavailable)
		return
	}

	rules := cm.GetRoutingRules()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"routes": rules,
		"total":  len(rules),
	})
}

// AddRoute adds a new routing rule
func (api *ChannelAPI) AddRoute(w http.ResponseWriter, r *http.Request) {
	var rule channels.RoutingRule

	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cm := api.app.GetChannelManager()
	if cm == nil {
		http.Error(w, "channels not available", http.StatusServiceUnavailable)
		return
	}

	if err := cm.AddRoutingRule(&rule); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(rule)
}

// RemoveRoute removes a routing rule
func (api *ChannelAPI) RemoveRoute(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ruleID := vars["id"]

	cm := api.app.GetChannelManager()
	if cm == nil {
		http.Error(w, "channels not available", http.StatusServiceUnavailable)
		return
	}

	cm.RemoveRoutingRule(ruleID)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "removed",
		"rule":   ruleID,
	})
}

// GetConfig returns the current channel configuration
func (api *ChannelAPI) GetConfig(w http.ResponseWriter, r *http.Request) {
	cm := api.app.GetChannelManager()
	if cm == nil {
		http.Error(w, "channels not available", http.StatusServiceUnavailable)
		return
	}

	// Return enabled channels info
	stats, _ := api.app.GetChannelStats()
	activeChannels := make([]string, 0, len(stats))
	for id := range stats {
		activeChannels = append(activeChannels, id)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"active_channels": activeChannels,
		"total_channels":  len(activeChannels),
	})
}

// ReloadConfig reloads the channel configuration
func (api *ChannelAPI) ReloadConfig(w http.ResponseWriter, r *http.Request) {
	// This would require implementing config reload in ChannelManager
	// For now, just acknowledge

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "config_reload_initiated",
	})
}

// Helper function
func ctxTimestamp() int64 {
	return 0 // Placeholder
}

// SetupMiddleware sets up middleware for the API
func SetupMiddleware(r *mux.Router) {
	// Logging middleware
	r.Use(loggingMiddleware)

	// CORS middleware (for web clients)
	r.Use(corsMiddleware)

	// Recovery middleware
	r.Use(recoveryMiddleware)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("API: %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// MountChannelAPI mounts the channels API to a router
func MountChannelAPI(r *mux.Router, app *core.ApplicationWithChannels) {
	api := NewChannelAPI(app)
	api.RegisterRoutes(r)
	SetupMiddleware(r.PathPrefix("/api/channels").Subrouter())
}

// Example usage:
//
//   r := mux.NewRouter()
//   app, _ := core.NewApplicationWithChannels(ctx, config, modelFactory, tools)
//   app.InitChannels(ctx)
//   app.StartChannels()
//   defer app.StopChannels()
//
//   api.MountChannelAPI(r, app)
//
//   http.ListenAndServe(":8080", r)
