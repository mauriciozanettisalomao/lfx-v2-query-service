// Copyright The Linux Foundation and each contributor to LFX.
// SPDX-License-Identifier: MIT

package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/linuxfoundation/lfx-v2-query-service/internal/domain"
)

// ResourceSearch handles resource-related business operations
// It depends on abstractions (interfaces) rather than concrete implementations
type ResourceSearch struct {
	resourceSearcher domain.ResourceSearcher
}

// QueryResources performs resource search with business logic validation
func (s *ResourceSearch) QueryResources(ctx context.Context, criteria domain.SearchCriteria) (*domain.SearchResult, error) {

	slog.DebugContext(ctx, "starting resource search",
		"name", criteria.Name,
		"type", criteria.Type,
		"parent", criteria.Parent,
	)

	// Validate business rules
	if err := s.validateSearchCriteria(criteria); err != nil {
		return nil, fmt.Errorf("invalid search criteria: %w", err)
	}

	// Apply default sorting if not provided
	if criteria.Sort == "" {
		criteria.Sort = "name_asc"
	}

	// Log the search operation
	slog.DebugContext(ctx, "validated search criteria, proceeding with search")

	// Delegate to the search implementation
	result, err := s.resourceSearcher.QueryResources(ctx, criteria)
	if err != nil {
		return nil, fmt.Errorf("search operation failed: %w", err)
	}

	// Apply business rules to the results
	s.processSearchResults(ctx, result)

	return result, nil
}

// validateSearchCriteria validates the search criteria according to business rules
func (s *ResourceSearch) validateSearchCriteria(criteria domain.SearchCriteria) error {
	// At least one search parameter must be provided
	if criteria.Name == nil && criteria.Parent == nil && criteria.Type == nil && len(criteria.Tags) == 0 {
		return fmt.Errorf("at least one search parameter must be provided")
	}

	// Validate name length if provided
	if criteria.Name != nil && len(*criteria.Name) < 1 {
		return fmt.Errorf("name must be at least 1 character long")
	}

	// Validate sort parameter
	validSortValues := []string{"name_asc", "name_desc", "updated_asc", "updated_desc"}
	if criteria.Sort != "" {
		valid := false
		for _, v := range validSortValues {
			if criteria.Sort == v {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid sort value: %s", criteria.Sort)
		}
	}

	return nil
}

// processSearchResults applies business logic to search results
func (s *ResourceSearch) processSearchResults(ctx context.Context, result *domain.SearchResult) {
	// Apply cache control policy
	if result.CacheControl == nil {
		cacheControl := "public, max-age=300"
		result.CacheControl = &cacheControl
	}

	// Log the results count
	slog.InfoContext(ctx, "search completed successfully", "resources_found", len(result.Resources))
}

// NewResourceSearch creates a new ResourceSearch instance
func NewResourceSearch(resourceSearcher domain.ResourceSearcher) domain.ResourceSearcher {
	return &ResourceSearch{
		resourceSearcher: resourceSearcher,
	}
}
