// Copyright The Linux Foundation and each contributor to LFX.
// SPDX-License-Identifier: MIT

package service

import (
	"context"
	"log/slog"

	"github.com/linuxfoundation/lfx-v2-query-service/internal/domain/model"
	"github.com/linuxfoundation/lfx-v2-query-service/internal/domain/port"
)

// OrganizationSearcher defines the interface for organization search operations
// This abstraction allows different search implementations (OpenSearch, etc.)
// without the domain layer knowing about specific implementations
type OrganizationSearcher interface {
	// QueryOrganizations searches for organizations based on the provided criteria
	QueryOrganizations(ctx context.Context, criteria model.OrganizationSearchCriteria) (*model.Organization, error)

	// SuggestOrganizations returns organization suggestions for typeahead search
	SuggestOrganizations(ctx context.Context, criteria model.OrganizationSuggestionCriteria) (*model.OrganizationSuggestionsResult, error)

	// IsReady checks if the search service is ready
	IsReady(ctx context.Context) error
}

// OrganizationSearch handles organization-related business operations
// It depends on abstractions (interfaces) rather than concrete implementations
type OrganizationSearch struct {
	organizationSearcher port.OrganizationSearcher
}

// QueryOrganizations performs organization search with business logic validation
func (s *OrganizationSearch) QueryOrganizations(ctx context.Context, criteria model.OrganizationSearchCriteria) (*model.Organization, error) {

	slog.DebugContext(ctx, "starting organization search",
		"name", criteria.Name,
		"domain", criteria.Domain,
	)

	// Delegate to the search implementation
	result, err := s.organizationSearcher.QueryOrganizations(ctx, criteria)
	if err != nil {
		slog.ErrorContext(ctx, "organization search operation failed while executing query organizations",
			"error", err,
		)
		return nil, err
	}

	var orgName, orgDomain string
	if result != nil {
		orgName = result.Name
		orgDomain = result.Domain
	}

	slog.DebugContext(ctx, "organization search completed",
		"organization_name", orgName,
		"organization_domain", orgDomain,
	)

	return result, nil
}

// SuggestOrganizations performs organization suggestions with business logic validation
func (s *OrganizationSearch) SuggestOrganizations(ctx context.Context, criteria model.OrganizationSuggestionCriteria) (*model.OrganizationSuggestionsResult, error) {

	slog.DebugContext(ctx, "starting organization suggestions search",
		"query", criteria.Query,
	)

	// Delegate to the search implementation
	result, err := s.organizationSearcher.SuggestOrganizations(ctx, criteria)
	if err != nil {
		slog.ErrorContext(ctx, "organization suggestions search operation failed",
			"error", err,
		)
		return nil, err
	}

	var suggestionCount int
	if result != nil {
		suggestionCount = len(result.Suggestions)
	}

	slog.DebugContext(ctx, "organization suggestions search completed",
		"query", criteria.Query,
		"suggestion_count", suggestionCount,
	)

	return result, nil
}

func (s *OrganizationSearch) IsReady(ctx context.Context) error {
	if err := s.organizationSearcher.IsReady(ctx); err != nil {
		return err
	}

	return nil
}

// NewOrganizationSearch creates a new OrganizationSearch instance
func NewOrganizationSearch(organizationSearcher port.OrganizationSearcher) OrganizationSearcher {
	return &OrganizationSearch{
		organizationSearcher: organizationSearcher,
	}
}
