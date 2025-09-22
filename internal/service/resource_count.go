// Copyright The Linux Foundation and each contributor to LFX.
// SPDX-License-Identifier: MIT

package service

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

// ResourceCountSearcher defines the interface for resource search operations
// This abstraction allows different search implementations (OpenSearch, etc.)
// without the domain layer knowing about specific implementations
type ResourceCountSearcher interface {
	// QueryResourcesCount searches for resources based on the provided criteria
	QueryResourcesCount(ctx context.Context, countCriteria model.SearchCriteria, aggregationCriteria model.SearchCriteria) (*model.CountResult, error)

	// IsReady checks if the search service is ready
	IsReady(ctx context.Context) error
}

// ResourceCount handles resource-related business operations
// It depends on abstractions (interfaces) rather than concrete implementations
type ResourceCount struct {
	resourceSearcher port.ResourceSearcher
	accessChecker    port.AccessControlChecker
}

func (s *ResourceCount) QueryResourcesCount(
	ctx context.Context,
	publicCountCriteria model.SearchCriteria,
	aggregationCriteria model.SearchCriteria,
) (*model.CountResult, error) {

	slog.DebugContext(ctx, "starting resource count search",
		"count_criteria", publicCountCriteria,
		"aggregation_criteria", aggregationCriteria,
	)

	// Grab the principal which was stored into the context by the security handler.
	principal, ok := ctx.Value(constants.PrincipalContextID).(string)
	if !ok {
		// This should not happen; the Auther always sets this or errors.
		return nil, errors.NewValidation("missing principal in context")
	}

	// Log the search operation
	slog.DebugContext(ctx, "validated search criteria, proceeding with count search")

	// Delegate to the search implementation
	publicOnly := principal == constants.AnonymousPrincipal
	result, err := s.resourceSearcher.QueryResourcesCount(ctx, publicCountCriteria, aggregationCriteria, publicOnly)
	if err != nil {
		slog.ErrorContext(ctx, "search operation failed while executing query resources",
			"error", err,
		)
		return nil, fmt.Errorf("search operation failed: %w", err)
	}

	// If the principal is anonymous, we can return the result immediately without checking access control
	// since we already retrieved the public-only count.
	if principal == constants.AnonymousPrincipal {
		slog.DebugContext(ctx, "returning anonymous count result",
			"count", result.Count,
		)
		// Set a cache control header for anonymous users.
		cacheControl := constants.AnonymousCacheControlHeader
		result.CacheControl = &cacheControl
		return result, nil
	}

	slog.DebugContext(ctx, "checking access control for private resources",
		"aggregations", result.Aggregation,
	)

	messageCheckAccess := s.BuildMessage(ctx, principal, result, aggregationCriteria)

	// Check access control for the resources to determine the authorized response count
	privateCount, err := s.CheckAccess(ctx, principal, result, messageCheckAccess)
	if err != nil {
		slog.ErrorContext(ctx, "access control check failed",
			"error", err,
		)
		return nil, fmt.Errorf("access control check failed: %w", err)
	}
	// The count already contains the count of public resources, so we need to add the count of private resources.
	result.Count += int(privateCount)

	// Check for bucket overflow.
	// There could be more buckets than the page size, and therefore more results.
	if result.Aggregation.SumOtherDocCount > 0 {
		result.HasMore = true
	}

	return result, nil
}

func (s *ResourceCount) BuildMessage(ctx context.Context, principal string, result *model.CountResult, aggregationCriteria model.SearchCriteria) []byte {

	// Create a map to store the "doc_count" of each aggregation bucket.
	docCountMap := make(map[string]uint64, aggregationCriteria.PageSize)

	// estimate the size of each line in the access check message
	accessCheckMessage := make([]byte, 0, 80*aggregationCriteria.PageSize)

	for _, bucket := range result.Aggregation.Buckets {
		docCountMap[bucket.Key] = bucket.DocCount
		accessCheckMessage = append(accessCheckMessage, bucket.Key...)
		accessCheckMessage = append(accessCheckMessage, []byte("@user:")...)
		accessCheckMessage = append(accessCheckMessage, []byte(principal)...)
		accessCheckMessage = append(accessCheckMessage, '\n')
	}

	return accessCheckMessage
}

func (s *ResourceCount) CheckAccess(ctx context.Context, principal string, result *model.CountResult, accessCheckMessage []byte) (uint64, error) {

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
			return 0, fmt.Errorf("access control check failed: %w", errCheckAccess)
		}
		accessCheckResponses = accessCheckResult
	}
	slog.DebugContext(ctx, "access check responses", "responses", accessCheckResponses)

	var count uint64
	for _, bucket := range result.Aggregation.Buckets {
		// The bucket.Key already contains the full access check query including the principal
		// e.g.: "committee:830513f8-0e77-4a48-a8e4-ede4c1a61f98#viewer@user:project_super_admin"
		// The BuildMessage function appends "@user:" + principal to create the access check key
		// So we need to use the same format here
		accessCheckKey := bucket.Key + "@user:" + principal
		slog.DebugContext(ctx, "checking access control for bucket",
			"bucket", bucket.Key,
			"access_check_key", accessCheckKey,
		)
		if allowed, ok := accessCheckResponses[accessCheckKey]; ok && allowed == "true" {
			count += bucket.DocCount
		}
	}

	return count, nil

}

func (s *ResourceCount) IsReady(ctx context.Context) error {
	if err := s.resourceSearcher.IsReady(ctx); err != nil {
		return err
	}

	if err := s.accessChecker.IsReady(ctx); err != nil {
		return err
	}

	return nil
}

// NewResourceCount creates a new ResourceCount instance
func NewResourceCount(resourceSearcher port.ResourceSearcher, accessChecker port.AccessControlChecker) ResourceCountSearcher {
	return &ResourceCount{
		resourceSearcher: resourceSearcher,
		accessChecker:    accessChecker,
	}
}
