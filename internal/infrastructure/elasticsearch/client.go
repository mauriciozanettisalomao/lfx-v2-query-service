// Copyright The Linux Foundation and each contributor to LFX.
// SPDX-License-Identifier: MIT

package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

// HTTPClient implements the ElasticsearchClient interface using HTTP
type HTTPClient struct {
	baseURL    string
	httpClient *http.Client
	username   string
	password   string
}

// NewHTTPClient creates a new HTTP client for Elasticsearch
func NewHTTPClient(baseURL, username, password string) *HTTPClient {
	return &HTTPClient{
		baseURL:  baseURL,
		username: username,
		password: password,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Search executes a search query against Elasticsearch
func (c *HTTPClient) Search(ctx context.Context, index string, query string) (*SearchResponse, error) {
	url := fmt.Sprintf("%s/%s/_search", c.baseURL, index)

	slog.DebugContext(ctx, "executing elasticsearch search", "index", index)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBufferString(query))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.username != "" && c.password != "" {
		req.SetBasicAuth(c.username, c.password)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("elasticsearch returned status %d: %s", resp.StatusCode, string(body))
	}

	var searchResponse SearchResponse
	if err := json.Unmarshal(body, &searchResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &searchResponse, nil
}

// IsHealthy checks if Elasticsearch is healthy
func (c *HTTPClient) IsHealthy(ctx context.Context) error {
	url := fmt.Sprintf("%s/_cluster/health", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	if c.username != "" && c.password != "" {
		req.SetBasicAuth(c.username, c.password)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute health check: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("elasticsearch health check failed with status %d", resp.StatusCode)
	}

	var healthResponse struct {
		Status string `json:"status"`
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read health check response: %w", err)
	}

	if err := json.Unmarshal(body, &healthResponse); err != nil {
		return fmt.Errorf("failed to unmarshal health check response: %w", err)
	}

	if healthResponse.Status != "green" && healthResponse.Status != "yellow" {
		return fmt.Errorf("elasticsearch cluster status is %s", healthResponse.Status)
	}

	return nil
}

// Config represents Elasticsearch configuration
type Config struct {
	URL      string `json:"url"`
	Username string `json:"username"`
	Password string `json:"password"`
	Index    string `json:"index"`
}

// NewElasticsearchSearcherFromConfig creates a new Elasticsearch searcher from configuration
func NewElasticsearchSearcherFromConfig(config Config) (*ElasticsearchSearcher, error) {
	if config.URL == "" {
		return nil, fmt.Errorf("elasticsearch URL is required")
	}
	if config.Index == "" {
		return nil, fmt.Errorf("elasticsearch index is required")
	}

	client := NewHTTPClient(config.URL, config.Username, config.Password)
	return NewElasticsearchSearcher(client, config.Index)
}
