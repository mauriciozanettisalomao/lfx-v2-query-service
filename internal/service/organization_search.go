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
