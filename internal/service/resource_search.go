// Copyright The Linux Foundation and each contributor to LFX.
// SPDX-License-Identifier: MIT

package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/linuxfoundation/lfx-v2-query-service/internal/domain"
	"github.com/linuxfoundation/lfx-v2-query-service/pkg/constants"
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
		"type", criteria.ResourceType,
		"parent", criteria.Parent,
	)

	// It seems that Goa v3 does not natively support complex conditional validations
	// like “at least one of these fields must be set"
	if err := s.validateSearchCriteria(criteria); err != nil {
		slog.With("error", err).ErrorContext(ctx, "search criteria validation failed")
		return nil, fmt.Errorf("invalid search criteria: %w", err)
	}

	// Grab the principal which was stored into the context by the security handler.
	principal, ok := ctx.Value(constants.PrincipalContextID).(string)
	if !ok {
		// This should not happen; the Auther always sets this or errors.
		return nil, errors.New("authenticated principal is missing")
	}
	if principal == constants.AnonymousPrincipal {
		// For an anonymous use, we will use the "public:true" OpenSearch term
		// filter, instead of OpenFGA, to filter results for performance.
		slog.DebugContext(ctx, "anonymous user detected, applying public-only filter")
		criteria.PublicOnly = true
	}

	// Log the search operation
	slog.DebugContext(ctx, "validated search criteria, proceeding with search")

	// Delegate to the search implementation
	result, err := s.resourceSearcher.QueryResources(ctx, criteria)
	if err != nil {
		return nil, fmt.Errorf("search operation failed: %w", err)
	}

	// TODO check access NATS implementation
	//
	//

	return result, nil
}

// validateSearchCriteria validates the search criteria according to business rules
func (s *ResourceSearch) validateSearchCriteria(criteria domain.SearchCriteria) error {
	// At least one search parameter must be provided
	if criteria.Name == nil && criteria.Parent == nil && criteria.ResourceType == nil && len(criteria.Tags) == 0 {
		return fmt.Errorf("at least one search parameter must be provided")
	}

	return nil
}

// NewResourceSearch creates a new ResourceSearch instance
func NewResourceSearch(resourceSearcher domain.ResourceSearcher) domain.ResourceSearcher {
	return &ResourceSearch{
		resourceSearcher: resourceSearcher,
	}
}
