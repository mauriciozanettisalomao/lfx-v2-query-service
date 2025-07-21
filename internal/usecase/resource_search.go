// Copyright The Linux Foundation and each contributor to LFX.
// SPDX-License-Identifier: MIT

package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/linuxfoundation/lfx-v2-query-service/internal/domain/model"
	"github.com/linuxfoundation/lfx-v2-query-service/internal/domain/port"
	"github.com/linuxfoundation/lfx-v2-query-service/pkg/constants"
	"github.com/linuxfoundation/lfx-v2-query-service/pkg/errors"
)

// ResourceSearcher defines the interface for resource search operations
// This abstraction allows different search implementations (OpenSearch, etc.)
// without the domain layer knowing about specific implementations
type ResourceSearcher interface {
	// QueryResources searches for resources based on the provided criteria
	QueryResources(ctx context.Context, criteria model.SearchCriteria) (*model.SearchResult, error)

	// IsReady checks if the search service is ready
	IsReady(ctx context.Context) error
}

// ResourceSearch handles resource-related business operations
// It depends on abstractions (interfaces) rather than concrete implementations
type ResourceSearch struct {
	resourceSearcher port.ResourceSearcher
	accessChecker    port.AccessControlChecker
}

// QueryResources performs resource search with business logic validation
func (s *ResourceSearch) QueryResources(ctx context.Context, criteria model.SearchCriteria) (*model.SearchResult, error) {

	slog.DebugContext(ctx, "starting resource search",
		"name", criteria.Name,
		"type", criteria.ResourceType,
		"parent", criteria.Parent,
	)

	// It seems that Goa v3 does not natively support complex conditional validations
	// like “at least one of these fields must be set"
	if err := s.validateSearchCriteria(criteria); err != nil {
		slog.ErrorContext(ctx, "search criteria validation failed", "error", err)
		return nil, errors.NewValidation(
			"search criteria validation failed",
			err,
		)
	}

	// Grab the principal which was stored into the context by the security handler.
	principal, ok := ctx.Value(constants.PrincipalContextID).(string)
	if !ok {
		// This should not happen; the Auther always sets this or errors.
		return nil, errors.NewValidation("missing principal in context")
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

	searchResult := &model.SearchResult{
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
func (s *ResourceSearch) validateSearchCriteria(criteria model.SearchCriteria) error {
	// At least one search parameter must be provided
	if criteria.Name == nil && criteria.Parent == nil && criteria.ResourceType == nil && len(criteria.Tags) == 0 {
		return fmt.Errorf("at least one search parameter must be provided: name, parent, type, or tags")
	}

	return nil
}

func (s *ResourceSearch) BuildMessage(ctx context.Context, principal string, result *model.SearchResult) []byte {

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

func (s *ResourceSearch) CheckAccess(ctx context.Context, principal string, resourceList []model.Resource, accessCheckMessage []byte) ([]model.Resource, error) {

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

	var resources []model.Resource
	// ensuring the original order of resources
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

func (s *ResourceSearch) IsReady(ctx context.Context) error {
	if err := s.resourceSearcher.IsReady(ctx); err != nil {
		return err
	}

	if err := s.accessChecker.IsReady(ctx); err != nil {
		return err
	}

	return nil
}

// NewResourceSearch creates a new ResourceSearch instance
func NewResourceSearch(resourceSearcher port.ResourceSearcher, accessChecker port.AccessControlChecker) ResourceSearcher {
	return &ResourceSearch{
		resourceSearcher: resourceSearcher,
		accessChecker:    accessChecker,
	}
}
