// Copyright The Linux Foundation and each contributor to LFX.
// SPDX-License-Identifier: MIT

package model

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
