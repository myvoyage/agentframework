// Gateway Configuration
// SPDX-License-Identifier: AGPL-3.0-or-later

package gateway

import (
	"os"
	"path/filepath"
	"time"
)

// Config represents the gateway configuration
type Config struct {
	// Gateway settings
	Gateway GatewaySettings `json:"gateway"`

	// Auth settings
	Auth AuthSettings `json:"auth"`

	// Canvas file server settings
	CanvasHost CanvasHostSettings `json:"canvasHost"`

	// Reload settings
	Reload ReloadSettings `json:"reload"`

	// Agent defaults
	Agents AgentDefaults `json:"agents"`

	// Config file path (not persisted, runtime only)
	ConfigPath string `json:"-"`
}

// GatewaySettings gateway core settings
type GatewaySettings struct {
	Port    int    `json:"port"`
	Host    string `json:"host"`
	Verbose bool   `json:"verbose"`
	Dev     bool   `json:"dev"`
	Force   bool   `json:"force"`
}

// AuthSettings authentication settings
type AuthSettings struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

// CanvasHostSettings canvas file server settings
type CanvasHostSettings struct {
	Enabled bool   `json:"enabled"`
	Port    int    `json:"port"`
	Dir     string `json:"dir"`
}

// ReloadSettings hot-reload settings
type ReloadSettings struct {
	Mode string `json:"mode"` // "off", "hot", "restart", "hybrid"
}

// AgentDefaults default agent settings
type AgentDefaults struct {
	Workspace string `json:"workspace"`
	Model     string `json:"model,omitempty"`
	Timeout   int64  `json:"timeout"` // seconds
}

// DefaultConfig returns the default gateway configuration
func DefaultConfig() *Config {
	return &Config{
		Gateway: GatewaySettings{
			Port:    18640,
			Host:    "127.0.0.1",
			Verbose: false,
			Dev:     false,
			Force:   false,
		},
		Auth: AuthSettings{
			Token:    "",
			Password: "",
		},
		CanvasHost: CanvasHostSettings{
			Enabled: true,
			Port:    18793,
			Dir:     "~/.openclaw/workspace/canvas",
		},
		Reload: ReloadSettings{
			Mode: "hybrid",
		},
		Agents: AgentDefaults{
			Workspace: "~/.openclaw/workspace",
			Timeout:   120,
		},
	}
}

// LoadConfig loads config from a file, with env overrides
func LoadConfig(path string) (*Config, error) {
	cfg := DefaultConfig()

	// Apply env overrides
	if port := os.Getenv("OPENCLAW_GATEWAY_PORT"); port != "" {
		var p int
		if _, err := parsePort(port, &p); err == nil {
			cfg.Gateway.Port = p
		}
	}
	if token := os.Getenv("OPENCLAW_GATEWAY_TOKEN"); token != "" {
		cfg.Auth.Token = token
	}
	if skipCanvas := os.Getenv("OPENCLAW_SKIP_CANVAS_HOST"); skipCanvas == "1" {
		cfg.CanvasHost.Enabled = false
	}
	if devPath := os.Getenv("OPENCLAW_CONFIG_PATH"); devPath != "" {
		// Dev mode config path
		_ = devPath
	}
	if devState := os.Getenv("OPENCLAW_STATE_DIR"); devState != "" {
		// Dev mode state dir
		_ = devState
	}

	return cfg, nil
}

// ApplyConfigPath applies command-line overrides to config
func (c *Config) ApplyFlags(port int, verbose, dev, force bool) {
	if port > 0 {
		c.Gateway.Port = port
	}
	c.Gateway.Verbose = verbose
	c.Gateway.Dev = dev
	c.Gateway.Force = force
}

// DevConfig returns config for dev mode
func DevConfig() *Config {
	cfg := DefaultConfig()
	cfg.Gateway.Port = 19001
	cfg.Gateway.Dev = true
	cfg.CanvasHost.Port = 19005
	cfg.Agents.Workspace = "~/.openclaw/workspace-dev"
	cfg.Auth.Token = ""
	return cfg
}

// Policy returns the gateway policy
func (c *Config) Policy() *Policy {
	return &Policy{
		MaxPayload:        4 * 1024 * 1024, // 4MB
		MaxBufferedBytes:  10 * 1024 * 1024, // 10MB
		TickIntervalMs:    30000, // 30s
	}
}

// TickInterval returns the tick interval as duration
func (c *Config) TickInterval() time.Duration {
	return 30 * time.Second
}

// AgentTimeout returns the agent timeout as duration
func (c *Config) AgentTimeout() time.Duration {
	if c.Agents.Timeout <= 0 {
		return 120 * time.Second
	}
	return time.Duration(c.Agents.Timeout) * time.Second
}

// StateDir returns the state directory path
func StateDir() string {
	if dir := os.Getenv("OPENCLAW_STATE_DIR"); dir != "" {
		return dir
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".openclaw")
}

// ConfigPath returns the config file path
func ConfigPath() string {
	if path := os.Getenv("OPENCLAW_CONFIG_PATH"); path != "" {
		return path
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".openclaw", "openclaw.json")
}

func parsePort(s string, port *int) (bool, error) {
	var p int
	for _, c := range s {
		if c < '0' || c > '9' {
			return false, nil
		}
		p = p*10 + int(c-'0')
	}
	*port = p
	return true, nil
}
