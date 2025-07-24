// Copyright The Linux Foundation and each contributor to LFX.
// SPDX-License-Identifier: MIT

package port

import (
	"context"

	"github.com/linuxfoundation/lfx-v2-query-service/internal/domain/model"
)

// ResourceSearcher defines the interface for resource search operations
// This abstraction allows different search implementations (OpenSearch, etc.)
// without the domain layer knowing about specific implementations
type ResourceSearcher interface {
	// QueryResources searches for resources based on the provided criteria
	QueryResources(ctx context.Context, criteria model.SearchCriteria) (*model.SearchResult, error)

	// IsReady checks if the search service is ready
	IsReady(ctx context.Context) error
}
