// Copyright The Linux Foundation and each contributor to LFX.
// SPDX-License-Identifier: MIT

package nats

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/linuxfoundation/lfx-v2-query-service/pkg/constants"
	"github.com/linuxfoundation/lfx-v2-query-service/pkg/errors"

	"github.com/nats-io/nats.go"
)

// NATSClient wraps the NATS connection and provides access control operations
type NATSClient struct {
	conn    *nats.Conn
	config  Config
	timeout time.Duration
}

// NATSClientInterface defines the interface for NATS operations
// This allows for easy mocking and testing
type NATSClientInterface interface {
	CheckAccess(ctx context.Context, request *AccessCheckNATSRequest) (AccessCheckNATSResponse, error)
	Close() error
	IsReady(ctx context.Context) error
}

// CheckAccess sends an access control request via NATS and waits for the response
func (c *NATSClient) CheckAccess(ctx context.Context, request *AccessCheckNATSRequest) (AccessCheckNATSResponse, error) {

	if request == nil {
		return nil, fmt.Errorf("invalid NATS access check request: request cannot be nil")
	}

	if request.Subject == "" || request.Message == nil || len(request.Message) == 0 {
		return nil, fmt.Errorf("invalid NATS access check request: subject and message must be set")
	}

	// Send the request and wait for response
	natsResponse, errRequest := c.conn.Request(request.Subject, request.Message, request.Timeout)
	if errRequest != nil {
		return nil, fmt.Errorf("NATS request failed: %w", errRequest)
	}

	slog.DebugContext(ctx, "received NATS response",
		"subject", request.Subject,
		"message", string(natsResponse.Data),
		"timeout", request.Timeout,
	)

	response := make(map[string]string)
	// Deserialize the response
	// Parse the response.
	lines := bytes.Split(natsResponse.Data, []byte("\n"))
	for _, line := range lines {
		// Split the relation from the "allowed" result.
		var relationPart, allowedPart []byte
		var found bool
		if relationPart, allowedPart, found = bytes.Cut(line, []byte("\t")); !found {
			slog.ErrorContext(ctx, "invalid NATS response format",
				"message", string(line),
			)
			return nil, errors.NewUnexpected("invalid NATS response format")
		}
		// Add the response to our map so we can look it up on the corresponding hit.
		response[string(relationPart)] = string(allowedPart)
	}

	return response, nil
}

// Close gracefully closes the NATS connection
func (c *NATSClient) Close() error {
	if c.conn != nil {
		c.conn.Close()
	}
	return nil
}

// IsReady checks if the NATS client is ready
func (c *NATSClient) IsReady(ctx context.Context) error {
	if c.conn == nil {
		return errors.NewServiceUnavailable("NATS client is not initialized or not connected")
	}
	if !c.conn.IsConnected() || c.conn.IsDraining() {
		return errors.NewServiceUnavailable("NATS client is not ready, connection is not established or is draining")
	}
	return nil
}

// NewClient creates a new NATS client with the given configuration
func NewClient(ctx context.Context, config Config) (*NATSClient, error) {
	slog.InfoContext(ctx, "creating NATS client",
		"url", config.URL,
		"timeout", config.Timeout,
	)

	// Validate configuration
	if config.URL == "" {
		return nil, errors.NewUnexpected("NATS URL is required")
	}

	// Configure NATS connection options
	opts := []nats.Option{
		nats.Name(constants.ServiceName),
		nats.Timeout(config.Timeout),
		nats.MaxReconnects(config.MaxReconnect),
		nats.ReconnectWait(config.ReconnectWait),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			slog.WarnContext(ctx, "NATS disconnected", "error", err)
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			slog.InfoContext(ctx, "NATS reconnected", "url", nc.ConnectedUrl())
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			slog.InfoContext(ctx, "NATS connection closed")
		}),
	}

	// Establish connection
	conn, err := nats.Connect(config.URL, opts...)
	if err != nil {
		return nil, errors.NewServiceUnavailable("failed to connect to NATS", err)
	}

	client := &NATSClient{
		conn:    conn,
		config:  config,
		timeout: config.Timeout,
	}

	slog.InfoContext(ctx, "NATS client created successfully",
		"connected_url", conn.ConnectedUrl(),
		"status", conn.Status(),
	)

	return client, nil
}
