package network

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// TLSConfig represents TLS configuration options
type TLSConfig struct {
	CertFile   string `yaml:"cert_file"`
	KeyFile    string `yaml:"key_file"`
	CAFile     string `yaml:"ca_file"`
	Insecure   bool   `yaml:"insecure"`
	ServerName string `yaml:"server_name"`
}

// GetTLSConfig returns a TLS configuration with loaded certificates if provided
func GetTLSConfig(cfg *TLSConfig) (*tls.Config, error) {
	if cfg == nil {
		return nil, nil
	}

	if cfg.Insecure {
		return &tls.Config{
			InsecureSkipVerify: true,
		}, nil
	}

	tlsConfig := &tls.Config{}

	// Load client certificate if specified
	if cfg.CertFile != "" && cfg.KeyFile != "" {
		cert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load client certificate: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	// Load CA certificate if specified
	if cfg.CAFile != "" {
		caCert, err := os.ReadFile(cfg.CAFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA certificate: %w", err)
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to parse CA certificate")
		}

		tlsConfig.RootCAs = caCertPool
	}

	// Set server name for SNI if specified
	if cfg.ServerName != "" {
		tlsConfig.ServerName = cfg.ServerName
	}

	// Enable TLS 1.3 by default
	tlsConfig.MinVersion = tls.VersionTLS12
	tlsConfig.MaxVersion = tls.VersionTLS13

	return tlsConfig, nil
}

// LoadTLSConfigFromFile loads TLS configuration from a YAML file
func LoadTLSConfigFromFile(filePath string) (*TLSConfig, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read TLS config file: %w", err)
	}

	var cfg TLSConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse TLS config: %w", err)
	}

	return &cfg, nil
}
