// Copyright The Linux Foundation and each contributor to LFX.
// SPDX-License-Identifier: MIT

package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/linuxfoundation/lfx-v2-query-service/internal/domain"
	"github.com/linuxfoundation/lfx-v2-query-service/pkg/constants"
)

// ResourceSearch handles resource-related business operations
// It depends on abstractions (interfaces) rather than concrete implementations
type ResourceSearch struct {
	resourceSearcher domain.ResourceSearcher
	accessChecker    domain.AccessControlChecker
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

	slog.DebugContext(ctx, "checking access control for resources",
		"resource_count", len(result.Resources),
	)

	messageCheckAccess := s.BuildMessage(ctx, principal, result)

	searchResult := &domain.SearchResult{
		PageToken: result.PageToken,
	}

	// Check access control for the resources if needed
	checkedResources, errCheckAccess := s.CheckAccess(ctx, principal, result.Resources, messageCheckAccess)
	if errCheckAccess != nil {
		slog.ErrorContext(ctx, "access control check failed",
			"error", errCheckAccess,
			"message", string(messageCheckAccess),
		)
		return nil, fmt.Errorf("access control check failed: %w", errCheckAccess)
	}
	searchResult.Resources = checkedResources

	slog.DebugContext(ctx, "resource search completed",
		"query_count", len(result.Resources),
		"response_after_access_check", len(searchResult.Resources),
	)

	if principal == constants.AnonymousPrincipal {
		// Set a cache control header for anonymous users.
		cacheControl := constants.AnonymousCacheControlHeader
		searchResult.CacheControl = &cacheControl
	}

	return searchResult, nil
}

// validateSearchCriteria validates the search criteria according to business rules
func (s *ResourceSearch) validateSearchCriteria(criteria domain.SearchCriteria) error {
	// At least one search parameter must be provided
	if criteria.Name == nil && criteria.Parent == nil && criteria.ResourceType == nil && len(criteria.Tags) == 0 {
		return fmt.Errorf("at least one search parameter must be provided")
	}

	return nil
}

func (s *ResourceSearch) BuildMessage(ctx context.Context, principal string, result *domain.SearchResult) []byte {

	// avoid duplicate resource references in the result
	seenRefs := make(map[string]struct{}, len(result.Resources))

	// estimate the size of each line in the access check message
	accessCheckMessage := make([]byte, 0, 80*len(result.Resources))
	for idx := range result.Resources {

		if _, seen := seenRefs[result.Resources[idx].ObjectRef]; seen {
			// Skip this result.
			continue
		}
		seenRefs[result.Resources[idx].ObjectRef] = struct{}{}

		if result.Resources[idx].Public {
			result.Resources[idx].NeedCheck = false
			continue
		}

		if result.Resources[idx].AccessCheckObject == "" || result.Resources[idx].AccessCheckRelation == "" {
			// Unable to perform access check without these fields.
			slog.WarnContext(ctx, "resource missing access control information, skipping",
				"object_ref", result.Resources[idx].ObjectRef,
				"object_type", result.Resources[idx].ObjectType,
				"object_id", result.Resources[idx].ObjectID,
			)
			result.Resources[idx].NeedCheck = true
			continue
		}
		result.Resources[idx].NeedCheck = true
		// make the access check message
		accessCheckMessage = append(accessCheckMessage, result.Resources[idx].AccessCheckObject...)
		accessCheckMessage = append(accessCheckMessage, byte('#'))
		accessCheckMessage = append(accessCheckMessage, result.Resources[idx].AccessCheckRelation...)
		accessCheckMessage = append(accessCheckMessage, []byte("@user:")...)
		accessCheckMessage = append(accessCheckMessage, []byte(principal)...)
		accessCheckMessage = append(accessCheckMessage, '\n')

	}
	return accessCheckMessage
}

func (s *ResourceSearch) CheckAccess(ctx context.Context, principal string, resourceList []domain.Resource, accessCheckMessage []byte) ([]domain.Resource, error) {

	var accessCheckResponses map[string]string
	if len(accessCheckMessage) > 0 {

		slog.DebugContext(ctx, "performing access control checks",
			"message", string(accessCheckMessage),
		)

		// Trim trailing newline.
		accessCheckMessage = accessCheckMessage[:len(accessCheckMessage)-1]
		accessCheckResult, errCheckAccess := s.accessChecker.CheckAccess(ctx, constants.AccessCheckSubject, accessCheckMessage, 15*time.Second)
		if errCheckAccess != nil {
			slog.ErrorContext(ctx, "access control check failed",
				"error", errCheckAccess,
				"message", string(accessCheckMessage),
			)
			return nil, fmt.Errorf("access control check failed: %w", errCheckAccess)
		}
		accessCheckResponses = accessCheckResult
	}

	var resources []domain.Resource
	// ensuring the ori
	for _, resource := range resourceList {
		addToList := false
		if resource.NeedCheck && resource.AccessCheckObject != "" && resource.AccessCheckRelation != "" {
			relationKey := resource.AccessCheckObject + "#" + resource.AccessCheckRelation + "@user:" + principal
			if allowed, ok := accessCheckResponses[relationKey]; ok && allowed == "true" {
				addToList = true
			}
		}
		if !resource.NeedCheck || addToList {
			resources = append(resources, resource)
		}
	}

	return resources, nil

}

// NewResourceSearch creates a new ResourceSearch instance
func NewResourceSearch(resourceSearcher domain.ResourceSearcher, accessChecker domain.AccessControlChecker) domain.ResourceSearcher {
	return &ResourceSearch{
		resourceSearcher: resourceSearcher,
		accessChecker:    accessChecker,
	}
}
