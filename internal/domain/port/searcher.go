// Copyright The Linux Foundation and each contributor to LFX.
// SPDX-License-Identifier: MIT

package port

import (
	"context"

	"github.com/linuxfoundation/lfx-v2-query-service/internal/domain/model"
)

// ResourceSearcher defines the behavior for resource search operations
// This abstraction allows different search implementations (OpenSearch, etc.)
// without the domain layer knowing about specific implementations
type ResourceSearcher interface {
	// QueryResources searches for resources based on the provided criteria
	QueryResources(ctx context.Context, criteria model.SearchCriteria) (*model.SearchResult, error)

	// IsReady checks if the search service is ready
	IsReady(ctx context.Context) error
}

// OrganizationSearcher defines the behavior for organization search operations
// This abstraction allows different search implementations (External API, etc.)
// without the domain layer knowing about specific implementations
type OrganizationSearcher interface {
	// QueryOrganizations searches for organizations based on the provided criteria
	QueryOrganizations(ctx context.Context, criteria model.OrganizationSearchCriteria) (*model.Organization, error)

	// IsReady checks if the search service is ready
	IsReady(ctx context.Context) error
}
