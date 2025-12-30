package config

import (
	"crypto/tls"
	"fmt"
)

// TLSConfig holds TLS/HTTPS configuration
type TLSConfig struct {
	Enabled  bool
	CertFile string
	KeyFile  string
	MinVersion string // TLS12, TLS13
}

// LoadTLSConfig loads TLS configuration from environment or returns default
func LoadTLSConfig() *TLSConfig {
	return &TLSConfig{
		Enabled: getEnv("TLS_ENABLED", "false") == "true",
		CertFile: getEnv("TLS_CERT_FILE", ""),
		KeyFile: getEnv("TLS_KEY_FILE", ""),
		MinVersion: getEnv("TLS_MIN_VERSION", "TLS12"),
	}
}

// GetTLSConfig returns the Go TLS configuration
func (t *TLSConfig) GetTLSConfig() (*tls.Config, error) {
	if !t.Enabled {
		return nil, nil
	}

	if t.CertFile == "" || t.KeyFile == "" {
		return nil, fmt.Errorf("TLS enabled but certificate or key file not provided")
	}

	// Parse minimum version
	var minVersion uint16
	switch t.MinVersion {
	case "TLS12":
		minVersion = tls.VersionTLS12
	case "TLS13":
		minVersion = tls.VersionTLS13
	default:
		minVersion = tls.VersionTLS12
	}

	return &tls.Config{
		MinVersion:               minVersion,
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			// TLS 1.3
			tls.TLS_AES_256_GCM_SHA384,
			tls.TLS_CHACHA20_POLY1305_SHA256,
			tls.TLS_AES_128_GCM_SHA256,
			
			// TLS 1.2
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		},
	}, nil
}

// IsEnabled checks if TLS is enabled
func (t *TLSConfig) IsEnabled() bool {
	return t.Enabled
}
