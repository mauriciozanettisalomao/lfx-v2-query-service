// Copyright The Linux Foundation and each contributor to LFX.
// SPDX-License-Identifier: MIT

package clearbit

import (
	"fmt"
	"time"
)

// Config holds the configuration for Clearbit API client
type Config struct {
	// APIKey is the Clearbit API key for authentication
	APIKey string

	// BaseURL is the base URL for Clearbit API (default: https://company.clearbit.com)
	BaseURL string

	// Timeout is the HTTP client timeout for API requests
	Timeout time.Duration

	// MaxRetries is the maximum number of retry attempts for failed requests
	MaxRetries int

	// RetryDelay is the delay between retry attempts
	RetryDelay time.Duration
}

// DefaultConfig returns a Config with sensible defaults
func DefaultConfig() Config {
	return Config{
		BaseURL:    "https://company.clearbit.com",
		Timeout:    30 * time.Second,
		MaxRetries: 3,
		RetryDelay: 1 * time.Second,
	}
}

// NewConfig creates a new Clearbit configuration with the provided parameters
func NewConfig(apiKey, baseURL, timeout string, maxRetries int, retryDelay string) (Config, error) {
	// Validate required parameters
	if apiKey == "" {
		return Config{}, fmt.Errorf("API key is required for Clearbit configuration")
	}

	// Set defaults for optional parameters
	if baseURL == "" {
		return Config{}, fmt.Errorf("base URL is required for Clearbit configuration")
	}

	if timeout == "" {
		timeout = "10s"
	}
	timeoutDuration, err := time.ParseDuration(timeout)
	if err != nil {
		return Config{}, fmt.Errorf("invalid timeout duration: %w", err)
	}

	if maxRetries <= 0 {
		maxRetries = 3
	}

	if retryDelay == "" {
		retryDelay = "1s"
	}
	retryDelayDuration, err := time.ParseDuration(retryDelay)
	if err != nil {
		return Config{}, fmt.Errorf("invalid retry delay duration: %w", err)
	}

	return Config{
		APIKey:     apiKey,
		BaseURL:    baseURL,
		Timeout:    timeoutDuration,
		MaxRetries: maxRetries,
		RetryDelay: retryDelayDuration,
	}, nil
}
