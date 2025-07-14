// Copyright The Linux Foundation and each contributor to LFX.
// SPDX-License-Identifier: MIT

package nats

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/linuxfoundation/lfx-v2-query-service/internal/domain"
)

// NATSAccessControlChecker implements the AccessControlChecker interface for NATS
type NATSAccessControlChecker struct {
	client NATSClientInterface
}

// CheckAccess implements the AccessControlChecker interface
func (n *NATSAccessControlChecker) CheckAccess(ctx context.Context, subj string, data []byte, timeout time.Duration) (domain.AccessCheckResult, error) {
	slog.DebugContext(ctx, "executing NATS access control check",
		"subject", subj,
		"timeout", timeout,
		"message", string(data),
	)

	// Send request via NATS
	response, err := n.client.CheckAccess(ctx, &AccessCheckNATSRequest{
		Subject: subj,
		Message: data,
		Timeout: timeout,
	})
	if err != nil {
		slog.ErrorContext(ctx, "NATS access control check failed", "error", err)
		return nil, fmt.Errorf("NATS access control check failed: %w", err)
	}

	// Convert NATS response to domain response
	result := n.convertFromNATSResponse(response)

	slog.DebugContext(ctx, "NATS access control check completed",
		"subject", subj,
		"result", result,
	)

	return result, nil
}

// Close gracefully closes the NATS connection
func (n *NATSAccessControlChecker) Close() error {
	return n.client.Close()
}

// convertFromNATSResponse converts NATS response to domain response
func (n *NATSAccessControlChecker) convertFromNATSResponse(response AccessCheckNATSResponse) domain.AccessCheckResult {
	return domain.AccessCheckResult(response)
}

// NewAccessControlChecker creates a new NATS access control checker
func NewAccessControlChecker(ctx context.Context, config Config) (domain.AccessControlChecker, error) {
	slog.InfoContext(ctx, "creating NATS access control checker",
		"url", config.URL,
	)

	client, err := NewClient(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create NATS client: %w", err)
	}

	return &NATSAccessControlChecker{
		client: client,
	}, nil
}
