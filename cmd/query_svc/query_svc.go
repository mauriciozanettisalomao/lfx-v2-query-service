// Copyright The Linux Foundation and each contributor to LFX.
// SPDX-License-Identifier: MIT

package querysvcapi

import (
	"context"
	"fmt"
	"log/slog"

	querysvc "github.com/linuxfoundation/lfx-v2-query-service/gen/query_svc"
	"github.com/linuxfoundation/lfx-v2-query-service/internal/domain"
	"github.com/linuxfoundation/lfx-v2-query-service/internal/service"
	"github.com/linuxfoundation/lfx-v2-query-service/pkg/constants"
	"github.com/linuxfoundation/lfx-v2-query-service/pkg/global"
	"github.com/linuxfoundation/lfx-v2-query-service/pkg/log"
	"github.com/linuxfoundation/lfx-v2-query-service/pkg/paging"

	"goa.design/goa/v3/security"
)

// query-svc service implementation using clean architecture.
type querySvcsrvc struct {
	resourceService           domain.ResourceSearcher
	accessControlCheckService domain.AccessControlChecker
}

// JWTAuth implements the authorization logic for service "query-svc" for the
// "jwt" security scheme.
func (s *querySvcsrvc) JWTAuth(ctx context.Context, token string, scheme *security.JWTScheme) (context.Context, error) {

	// Parse the Heimdall-authorized principal from the token.
	principal, err := ParsePrincipal(ctx, token)
	if err != nil {
		return ctx, err
	}

	// Log the principal for debugging purposes in all logs for this request.
	ctx = log.AppendCtx(ctx, slog.String(string(constants.PrincipalAttribute), principal))

	// Return a new context containing the principal as a value.
	return context.WithValue(ctx, constants.PrincipalContextID, principal), nil
}

// Locate resources by their type or parent, or use typeahead search to query
// resources by a display name or similar alias.
func (s *querySvcsrvc) QueryResources(ctx context.Context, p *querysvc.QueryResourcesPayload) (res *querysvc.QueryResourcesResult, err error) {

	slog.DebugContext(ctx, "querySvc.query-resources",
		"name", p.Name,
	)

	// Convert payload to domain criteria
	criteria, errCriteria := s.payloadToCriteria(ctx, p)
	if errCriteria != nil {
		slog.ErrorContext(ctx, "failed to convert payload to criteria", "error", errCriteria)
		return nil, fmt.Errorf("failed to convert payload to criteria: %w", errCriteria)
	}

	// Execute search using the service layer
	result, err := s.resourceService.QueryResources(ctx, criteria)
	if err != nil {
		return nil, fmt.Errorf("resource search failed: %w", err)
	}

	// Convert domain result to response
	res = s.domainResultToResponse(result)
	return res, nil
}

// Check if the service is able to take inbound requests.
func (s *querySvcsrvc) Readyz(ctx context.Context) (res []byte, err error) {
	slog.DebugContext(ctx, "querySvc.readyz")
	return
}

// Check if the service is alive.
func (s *querySvcsrvc) Livez(ctx context.Context) (res []byte, err error) {
	slog.DebugContext(ctx, "querySvc.livez")
	return
}

// payloadToCriteria converts the generated payload to domain search criteria
func (s *querySvcsrvc) payloadToCriteria(ctx context.Context, p *querysvc.QueryResourcesPayload) (domain.SearchCriteria, error) {

	criteria := domain.SearchCriteria{
		Name:         p.Name,
		Parent:       p.Parent,
		ResourceType: p.Type,
		Tags:         p.Tags,
		SortBy:       p.Sort,
		PageToken:    p.PageToken,
		PageSize:     constants.DefaultPageSize,
	}
	switch p.Sort {
	case "name_asc":
		criteria.SortBy = "sort_name"
		criteria.SortOrder = "asc"
	case "name_desc":
		criteria.SortBy = "sort_name"
		criteria.SortOrder = "desc"
	case "updated_asc":
		criteria.SortBy = "updated_at"
		criteria.SortOrder = "asc"
	case "updated_desc":
		criteria.SortBy = "updated_at"
		criteria.SortOrder = "desc"
	}

	if criteria.PageToken != nil {
		pageToken, err := paging.DecodePageToken(ctx, *criteria.PageToken, global.PageTokenSecret(ctx))
		if err != nil {
			slog.ErrorContext(ctx, "failed to decode page token", "error", err)
			return criteria, fmt.Errorf("failed to decode page token: %w", err)
		}
		criteria.SearchAfter = &pageToken
		slog.DebugContext(ctx, "decoded page token",
			"page_token", *criteria.PageToken,
			"decoded", pageToken,
		)
	}

	return criteria, nil
}

// domainResultToResponse converts domain search result to generated response
func (s *querySvcsrvc) domainResultToResponse(result *domain.SearchResult) *querysvc.QueryResourcesResult {
	response := &querysvc.QueryResourcesResult{
		Resources:    make([]*querysvc.Resource, len(result.Resources)),
		PageToken:    result.PageToken,
		CacheControl: result.CacheControl,
	}

	for i, domainResource := range result.Resources {
		response.Resources[i] = &querysvc.Resource{
			Type: &domainResource.Type,
			ID:   &domainResource.ID,
			Data: domainResource.Data,
		}
	}

	return response
}

// NewQuerySvc returns the query-svc service implementation.
func NewQuerySvc(resourceSearcher domain.ResourceSearcher, accessControlChecker domain.AccessControlChecker) querysvc.Service {
	resourceService := service.NewResourceSearch(resourceSearcher, accessControlChecker)
	return &querySvcsrvc{
		resourceService: resourceService,
	}
}
