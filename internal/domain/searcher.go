// Copyright The Linux Foundation and each contributor to LFX.
// SPDX-License-Identifier: MIT

package domain

import (
	"context"
)

// ResourceSearcher defines the interface for resource search operations
// This abstraction allows different search implementations (OpenSearch, etc.)
// without the domain layer knowing about specific implementations
type ResourceSearcher interface {
	// QueryResources searches for resources based on the provided criteria
	QueryResources(ctx context.Context, criteria SearchCriteria) (*SearchResult, error)
}

// SearchCriteria encapsulates all possible search parameters
type SearchCriteria struct {
	// Resource name or alias; supports typeahead
	Name *string
	// Parent (for navigation; varies by object type)
	Parent *string
	// Resource type to search
	Type *string
	// Tags to search (varies by object type)
	Tags []string
	// Sort order for results
	Sort string
	// Opaque token for pagination
	PageToken *string
}

// SearchResult contains the results of a resource search
type SearchResult struct {
	// Resources found
	Resources []Resource
	// Opaque token if more results are available
	PageToken *string
	// Cache control header
	CacheControl *string
}

// Resource represents a domain resource entity
type Resource struct {
	// Resource type
	Type string
	// Resource ID (within its resource collection)
	ID string
	// Resource data snapshot
	Data any
}
