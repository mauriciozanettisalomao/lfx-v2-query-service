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

	publicResources, resourcesToCheck, message := s.BuildMessage(ctx, principal, result)

	searchResult := &domain.SearchResult{
		Resources: publicResources,
	}

	// Check access control for the resources
	if len(resourcesToCheck) > 0 {
		checkedResources, errCheckAccess := s.CheckAccess(ctx, principal, resourcesToCheck, message)
		if errCheckAccess != nil {
			slog.ErrorContext(ctx, "access control check failed",
				"error", errCheckAccess,
				"message", string(message),
			)
			return nil, fmt.Errorf("access control check failed: %w", errCheckAccess)
		}
		searchResult.Resources = append(searchResult.Resources, checkedResources...)
	}

	slog.DebugContext(ctx, "resource search completed",
		"query_count", len(result.Resources),
		"public_resources_count", len(publicResources),
		"checked_resources_count", len(resourcesToCheck),
		"response_count", len(searchResult.Resources),
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

func (s *ResourceSearch) BuildMessage(ctx context.Context, principal string, result *domain.SearchResult) (public []domain.Resource, needCheck []domain.Resource, msg []byte) {

	// avoid duplicate resource references in the result
	seenRefs := make(map[string]struct{}, len(result.Resources))

	accessCheckNeededHits := make([]domain.Resource, 0, len(result.Resources))
	// estimate the size of each line in the access check message
	accessCheckMessage := make([]byte, 0, 80*len(result.Resources))

	// TODO - evaluate if it's necessary to execute concurrent access control checks (GOROUTINEs)
	resources := make([]domain.Resource, 0, len(result.Resources))
	for _, resource := range result.Resources {

		if _, seen := seenRefs[resource.ObjectRef]; seen {
			// Skip this result.
			continue
		}
		seenRefs[resource.ObjectRef] = struct{}{}

		if resource.Public {
			resources = append(resources, domain.Resource{
				ID:   resource.ID,
				Type: resource.Type,
				Data: resource.Data,
			})
			continue
		}

		if resource.AccessCheckObject == "" || resource.AccessCheckRelation == "" {
			// Unable to perform access check without these fields.
			slog.WarnContext(ctx, "resource missing access control information, skipping",
				"object_ref", resource.ObjectRef,
				"object_type", resource.ObjectType,
				"object_id", resource.ObjectID,
			)
			continue
		}
		accessCheckNeededHits = append(accessCheckNeededHits, resource)
		// make the access check message
		accessCheckMessage = append(accessCheckMessage, resource.AccessCheckObject...)
		accessCheckMessage = append(accessCheckMessage, byte('#'))
		accessCheckMessage = append(accessCheckMessage, resource.AccessCheckRelation...)
		accessCheckMessage = append(accessCheckMessage, []byte("@user:")...)
		accessCheckMessage = append(accessCheckMessage, []byte(principal)...)
		accessCheckMessage = append(accessCheckMessage, '\n')

	}
	return resources, accessCheckNeededHits, accessCheckMessage
}

func (s *ResourceSearch) CheckAccess(ctx context.Context, principal string, accessCheckNeededHits []domain.Resource, accessCheckMessage []byte) ([]domain.Resource, error) {

	accessCheckResponses := make(map[string]string, len(accessCheckNeededHits))
	if len(accessCheckNeededHits) > 0 {

		slog.DebugContext(ctx, "performing access control checks",
			"resource_count", len(accessCheckNeededHits),
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
	for _, resource := range accessCheckNeededHits {
		relationKey := resource.AccessCheckObject + "#" + resource.AccessCheckRelation + "@user:" + principal
		if allowed, ok := accessCheckResponses[relationKey]; ok && allowed == "true" {
			resources = append(resources, domain.Resource{
				ID:   resource.ObjectID,
				Type: resource.ObjectType,
				Data: resource.Data,
			})
			continue
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
