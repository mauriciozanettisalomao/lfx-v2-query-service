// Copyright The Linux Foundation and each contributor to LFX.
// SPDX-License-Identifier: MIT

package domain

import (
	"context"
	"time"
)

// AccessControlChecker defines the interface for access control operations
// This abstraction allows different access control implementations (NATS, etc.)
// without the domain layer knowing about specific implementations
type AccessControlChecker interface {
	// CheckAccess verifies if a user has permission to access specific resources
	CheckAccess(ctx context.Context, subj string, data []byte, timeout time.Duration) (AccessCheckResult, error)

	// Close gracefully closes the access control checker connection
	Close() error
}

// AccessCheckResult contains the results of access verification
type AccessCheckResult map[string]string
