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

	// IsReady checks if the search service is ready
	IsReady(ctx context.Context) error
}

// SearchCriteria encapsulates all possible search parameters
type SearchCriteria struct {
	// Tags to filter resources
	Tags []string
	// Resource name or alias; supports typeahead
	Name *string
	// Parent (for navigation; varies by object type)
	Parent *string
	// ParentRef is a reference to the parent resource
	ParentRef *string
	// ResourceType to search
	ResourceType *string
	// SearchAfter is used for pagination
	SearchAfter *string
	// Sortby order for results
	SortBy string
	// SortOrder for results
	SortOrder string
	// Opaque token for pagination
	PageToken *string
	// Pagesize for pagination
	PageSize int
	// PublicOnly indicates if only public resources should be returned
	PublicOnly bool
}

// SearchResult contains the results of a resource search
type SearchResult struct {
	// Resources found
	Resources []Resource
	// Opaque token if more results are available
	PageToken *string
	// Cache control header
	CacheControl *string
	// Total number of resources found
	Total int
}

// Resource represents a domain resource entity
type Resource struct {
	// Resource type
	Type string
	// Resource ID (within its resource collection)
	ID string
	// Resource data snapshot
	Data any
	// Metadata about the resource
	TransactionBodyStub
	// NeedCheck indicates if access control check is needed
	NeedCheck bool
}

// TransactionBodyStub is used to decode the response's "source".
// **Ensure the fields here align to the relevant `SourceIncludes`
// parameters**.
type TransactionBodyStub struct {
	ObjectRef            string `json:"object_ref"`
	ObjectType           string `json:"object_type"`
	ObjectID             string `json:"object_id"`
	Public               bool   `json:"public"`
	AccessCheckObject    string `json:"access_check_object"`
	AccessCheckRelation  string `json:"access_check_relation"`
	HistoryCheckObject   string `json:"history_check_object"`
	HistoryCheckRelation string `json:"history_check_relation"`
}
