//go:build eino
// +build eino

package einobridge

import (
	"context"
	"fmt"
	"net"
	"net/rpc"
	"os"
	"strconv"
	"sync"
	"time"
)

// BridgeConfig holds the configuration for the Eino bridge
type BridgeConfig struct {
	Protocol   string  // "http", "tcp", "auto"
	Host       string  // RPC server host (for tcp client mode)
	Port       int     // RPC server port (for http server mode)
	EnableHTTP bool    // Enable HTTP RPC server
	HealthCheckInterval time.Duration // Health check interval
}

// DefaultBridgeConfig returns default configuration
func DefaultBridgeConfig() *BridgeConfig {
	port, _ := strconv.Atoi(os.Getenv("EINO_BRIDGE_PORT"))
	if port == 0 {
		port = 8080
	}

	return &BridgeConfig{
		Protocol:   getEnv("EINO_BRIDGE_PROTOCOL", "auto"),
		Host:       getEnv("EINO_BRIDGE_HOST", "localhost:8080"),
		Port:       port,
		EnableHTTP: getEnv("EINO_BRIDGE_ENABLE_HTTP", "true") == "true",
		HealthCheckInterval: 30 * time.Second,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

var (
	bridgeConfig     *BridgeConfig
	bridgeConfigOnce sync.Once
	bridgeInitialized bool
	bridgeMutex      sync.RWMutex
	cancelHealthCheck context.CancelFunc
)

// SetBridgeConfig sets the bridge configuration
func SetBridgeConfig(config *BridgeConfig) {
	bridgeConfigOnce.Do(func() {
		bridgeConfig = config
	})
}

// InitBridge initializes the Eino bridge with full configuration support
// It supports multiple protocols and includes health checking
func InitBridge() error {
	bridgeMutex.Lock()
	defer bridgeMutex.Unlock()

	// Check if already initialized
	if bridgeInitialized {
		return nil
	}

	// Load configuration
	if bridgeConfig == nil {
		bridgeConfig = DefaultBridgeConfig()
	}

	// Determine initialization strategy based on protocol
	var err error
	switch bridgeConfig.Protocol {
	case "http", "auto":
		if bridgeEngine != nil && bridgeConfig.EnableHTTP {
			err = initHTTPBridge()
		}
	case "tcp":
		err = initTCPBridge()
	default:
		return fmt.Errorf("unsupported bridge protocol: %s", bridgeConfig.Protocol)
	}

	if err != nil {
		return fmt.Errorf("failed to initialize bridge: %w", err)
	}

	// Start health check if configured
	if bridgeConfig.HealthCheckInterval > 0 {
		startHealthCheck()
	}

	bridgeInitialized = true
	fmt.Printf("Eino bridge initialized successfully (protocol=%s, port=%d)\n",
		bridgeConfig.Protocol, bridgeConfig.Port)

	return nil
}

// initHTTPBridge initializes the HTTP RPC bridge
func initHTTPBridge() error {
	if bridgeEngine == nil {
		return fmt.Errorf("bridge engine not set, cannot initialize HTTP bridge")
	}

	err := StartHTTPBridge(bridgeConfig.Port, bridgeEngine)
	if err != nil {
		return fmt.Errorf("failed to start HTTP bridge: %w", err)
	}

	return nil
}

// initTCPBridge initializes a TCP RPC client connection
func initTCPBridge() error {
	if bridgeConfig.Host == "" {
		return fmt.Errorf("bridge host not configured for TCP protocol")
	}

	// Attempt to connect to RPC server
	client, err := rpc.Dial("tcp", bridgeConfig.Host)
	if err != nil {
		return fmt.Errorf("failed to connect to RPC server at %s: %w", bridgeConfig.Host, err)
	}

	// Verify connection with a simple ping (if supported)
	// This is a placeholder for actual RPC health check
	_ = client

	return nil
}

// startHealthCheck starts a background health check routine
func startHealthCheck() {
	ctx, cancel := context.WithCancel(context.Background())
	cancelHealthCheck = cancel

	go func() {
		ticker := time.NewTicker(bridgeConfig.HealthCheckInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := healthCheck(); err != nil {
					fmt.Printf("Bridge health check failed: %v\n", err)
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}

// healthCheck performs a health check on the bridge
func healthCheck() error {
	if bridgeEngine == nil {
		return fmt.Errorf("bridge engine not set")
	}

	// For HTTP bridge, we could check if the server is still listening
	if bridgeConfig.Protocol == "http" || bridgeConfig.Protocol == "auto" {
		// Simple connection check to the HTTP port
		address := fmt.Sprintf("localhost:%d", bridgeConfig.Port)
		conn, err := net.DialTimeout("tcp", address, 5*time.Second)
		if err != nil {
			return fmt.Errorf("HTTP bridge not responding: %w", err)
		}
		conn.Close()
	}

	return nil
}

// ShutdownBridge gracefully shuts down the bridge
func ShutdownBridge() error {
	bridgeMutex.Lock()
	defer bridgeMutex.Unlock()

	// Stop health check
	if cancelHealthCheck != nil {
		cancelHealthCheck()
		cancelHealthCheck = nil
	}

	// Stop HTTP bridge if running
	if globalHTTPServer != nil {
		if err := StopHTTPBridge(); err != nil {
			return fmt.Errorf("failed to stop HTTP bridge: %w", err)
		}
	}

	bridgeInitialized = false
	fmt.Println("Eino bridge shutdown complete")

	return nil
}

// IsBridgeInitialized returns whether the bridge has been initialized
func IsBridgeInitialized() bool {
	bridgeMutex.RLock()
	defer bridgeMutex.RUnlock()
	return bridgeInitialized
}

// GetBridgeConfig returns the current bridge configuration
func GetBridgeConfig() *BridgeConfig {
	if bridgeConfig == nil {
		bridgeConfig = DefaultBridgeConfig()
	}
	return bridgeConfig
}
