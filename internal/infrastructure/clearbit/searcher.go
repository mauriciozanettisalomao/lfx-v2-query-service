// Copyright The Linux Foundation and each contributor to LFX.
// SPDX-License-Identifier: MIT

package clearbit

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/linuxfoundation/lfx-v2-query-service/internal/domain/model"
	"github.com/linuxfoundation/lfx-v2-query-service/pkg/errors"
)

// OrganizationSearcher implements the port.OrganizationSearcher interface using Clearbit API
type OrganizationSearcher struct {
	client *Client
}

// QueryOrganizations searches for organizations using Clearbit API
func (s *OrganizationSearcher) QueryOrganizations(ctx context.Context, criteria model.OrganizationSearchCriteria) (*model.Organization, error) {
	slog.DebugContext(ctx, "searching organization via Clearbit API",
		"name", criteria.Name,
		"domain", criteria.Domain,
	)

	var (
		clearbitCompany *ClearbitCompany
		err             error
	)

	// Search by domain first if provided (more accurate)
	if criteria.Domain != nil {
		slog.DebugContext(ctx, "searching by domain", "domain", *criteria.Domain)
		clearbitCompany, err = s.client.FindCompanyByDomain(ctx, *criteria.Domain)
		if err == nil {
			slog.DebugContext(ctx, "found organization by domain", "name", clearbitCompany.Name)
		}
	}

	// If domain search failed or wasn't provided, try name search
	if clearbitCompany == nil && criteria.Name != nil {
		slog.DebugContext(ctx, "searching by name", "name", *criteria.Name)
		clearbitCompany, err = s.client.FindCompanyByName(ctx, *criteria.Name)
		if err == nil {
			slog.DebugContext(ctx, "found organization by name", "name", clearbitCompany.Name)
		}
		// search by domain again to enrich the organization
		if clearbitCompany != nil && clearbitCompany.Domain != "" {
			clearbitCompanyEnriched, errFindCompanyByDomain := s.client.FindCompanyByDomain(ctx, clearbitCompany.Domain)
			if errFindCompanyByDomain == nil {
				slog.DebugContext(ctx, "found organization by domain", "name", clearbitCompany.Name)
				clearbitCompany = clearbitCompanyEnriched
			}
		}
	}

	if err != nil {
		slog.ErrorContext(ctx, "error searching organization", "error", err)
		return nil, err
	}

	if clearbitCompany == nil {
		slog.ErrorContext(ctx, "organization not found", "error", err)
		return nil, errors.NewNotFound("organization not found")
	}

	// Convert Clearbit company to domain model
	org := s.convertToDomainModel(clearbitCompany)

	slog.DebugContext(ctx, "successfully found and converted organization",
		"name", org.Name,
		"domain", org.Domain,
		"industry", org.Industry,
	)

	return org, nil
}

// SuggestOrganizations returns organization suggestions using Clearbit Autocomplete API
func (s *OrganizationSearcher) SuggestOrganizations(ctx context.Context, criteria model.OrganizationSuggestionCriteria) (*model.OrganizationSuggestionsResult, error) {
	slog.DebugContext(ctx, "searching organization suggestions via Clearbit Autocomplete API",
		"query", criteria.Query,
	)

	// Call the Clearbit Autocomplete API
	clearbitSuggestions, err := s.client.SuggestCompanies(ctx, criteria.Query)
	if err != nil {
		slog.ErrorContext(ctx, "error searching organization suggestions", "error", err)
		return nil, err
	}

	// Convert to domain model
	suggestions := make([]model.OrganizationSuggestion, len(clearbitSuggestions))
	for i, suggestion := range clearbitSuggestions {
		suggestions[i] = model.OrganizationSuggestion{
			Name:   suggestion.Name,
			Domain: suggestion.Domain,
			Logo:   suggestion.Logo,
		}
	}

	result := &model.OrganizationSuggestionsResult{
		Suggestions: suggestions,
	}

	slog.DebugContext(ctx, "successfully found organization suggestions",
		"query", criteria.Query,
		"count", len(suggestions),
	)

	return result, nil
}

// convertToDomainModel converts a Clearbit company to the domain model
func (s *OrganizationSearcher) convertToDomainModel(company *ClearbitCompany) *model.Organization {
	org := &model.Organization{
		Name:   company.Name,
		Domain: company.Domain,
	}

	// Map industry information
	if company.Category != nil {
		if company.Category.Industry != "" {
			org.Industry = company.Category.Industry
		} else if company.Category.Sector != "" {
			org.Industry = company.Category.Sector
		}

		if company.Category.SubIndustry != "" {
			org.Sector = company.Category.SubIndustry
		} else if company.Category.IndustryGroup != "" {
			org.Sector = company.Category.IndustryGroup
		}
	}

	// Map employee information
	if company.Metrics != nil {
		if company.Metrics.EmployeesRange != "" {
			org.Employees = company.Metrics.EmployeesRange
		} else if company.Metrics.Employees != nil {
			org.Employees = strconv.Itoa(*company.Metrics.Employees)
		}
	}

	return org
}

// IsReady checks if the Clearbit API is ready to serve requests
func (s *OrganizationSearcher) IsReady(ctx context.Context) error {
	return s.client.IsReady(ctx)
}

// NewOrganizationSearcher creates a new Clearbit-based organization searcher
func NewOrganizationSearcher(ctx context.Context, config Config) (*OrganizationSearcher, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("clearbit API key is required")
	}

	client := NewClient(config)

	// Test the connection
	if err := client.IsReady(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to Clearbit API: %w", err)
	}

	slog.InfoContext(ctx, "Clearbit organization searcher initialized successfully")

	return &OrganizationSearcher{
		client: client,
	}, nil
}
