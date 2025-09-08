// Copyright The Linux Foundation and each contributor to LFX.
// SPDX-License-Identifier: MIT

package service

import (
	"context"
	"log/slog"

	querysvc "github.com/linuxfoundation/lfx-v2-query-service/gen/query_svc"
	"github.com/linuxfoundation/lfx-v2-query-service/internal/domain/model"
	"github.com/linuxfoundation/lfx-v2-query-service/internal/domain/port"
	usecase "github.com/linuxfoundation/lfx-v2-query-service/internal/service"
	"github.com/linuxfoundation/lfx-v2-query-service/pkg/constants"
	"github.com/linuxfoundation/lfx-v2-query-service/pkg/global"
	"github.com/linuxfoundation/lfx-v2-query-service/pkg/log"
	"github.com/linuxfoundation/lfx-v2-query-service/pkg/paging"

	"goa.design/goa/v3/security"
)

// query-svc service implementation using clean architecture.
type querySvcsrvc struct {
	resourceService     usecase.ResourceSearcher
	organizationService usecase.OrganizationSearcher
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
		return nil, wrapError(ctx, errCriteria)
	}

	// Execute search using the service layer
	result, errQueryResources := s.resourceService.QueryResources(ctx, criteria)
	if errQueryResources != nil {
		return nil, wrapError(ctx, errQueryResources)
	}

	// Convert domain result to response
	res = s.domainResultToResponse(result)
	return res, nil
}

// Locate a single organization by name or domain.
func (s *querySvcsrvc) QueryOrgs(ctx context.Context, p *querysvc.QueryOrgsPayload) (res *querysvc.Organization, err error) {

	slog.DebugContext(ctx, "querySvc.query-orgs",
		"name", p.Name,
		"domain", p.Domain,
	)

	// Convert payload to domain criteria
	criteria := s.payloadToOrganizationCriteria(ctx, p)

	// Execute search using the service layer
	result, errQueryOrgs := s.organizationService.QueryOrganizations(ctx, criteria)
	if errQueryOrgs != nil {
		return nil, wrapError(ctx, errQueryOrgs)
	}

	// Convert domain result to response
	res = s.domainOrganizationToResponse(result)
	return res, nil
}

// Check if the service is able to take inbound requests.
func (s *querySvcsrvc) Readyz(ctx context.Context) (res []byte, err error) {
	errIsReady := s.resourceService.IsReady(ctx)
	if errIsReady != nil {
		slog.ErrorContext(ctx, "querySvc.readyz failed", "error", errIsReady)
		return nil, wrapError(ctx, errIsReady)
	}

	errIsOrgReady := s.organizationService.IsReady(ctx)
	if errIsOrgReady != nil {
		slog.ErrorContext(ctx, "querySvc.readyz organization service failed", "error", errIsOrgReady)
		return nil, wrapError(ctx, errIsOrgReady)
	}

	return []byte("OK\n"), nil
}

// Check if the service is alive.
func (s *querySvcsrvc) Livez(ctx context.Context) (res []byte, err error) {
	// This always returns as long as the service is still running. As this
	// endpoint is expected to be used as a Kubernetes liveness check, this
	// service must likewise self-detect non-recoverable errors and
	// self-terminate.
	return []byte("OK\n"), nil
}

// payloadToCriteria converts the generated payload to domain search criteria
func (s *querySvcsrvc) payloadToCriteria(ctx context.Context, p *querysvc.QueryResourcesPayload) (model.SearchCriteria, error) {

	criteria := model.SearchCriteria{
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
		pageToken, errPageToken := paging.DecodePageToken(ctx, *criteria.PageToken, global.PageTokenSecret(ctx))
		if errPageToken != nil {
			slog.ErrorContext(ctx, "failed to decode page token", "error", errPageToken)
			return criteria, wrapError(ctx, errPageToken)
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
func (s *querySvcsrvc) domainResultToResponse(result *model.SearchResult) *querysvc.QueryResourcesResult {
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

// payloadToOrganizationCriteria converts the generated payload to domain organization search criteria
func (s *querySvcsrvc) payloadToOrganizationCriteria(ctx context.Context, p *querysvc.QueryOrgsPayload) model.OrganizationSearchCriteria {
	criteria := model.OrganizationSearchCriteria{
		Name:   p.Name,
		Domain: p.Domain,
	}
	return criteria
}

// domainOrganizationToResponse converts domain organization result to generated response
func (s *querySvcsrvc) domainOrganizationToResponse(org *model.Organization) *querysvc.Organization {
	return &querysvc.Organization{
		Name:      &org.Name,
		Domain:    &org.Domain,
		Industry:  &org.Industry,
		Sector:    &org.Sector,
		Employees: &org.Employees,
	}
}

// NewQuerySvc returns the query-svc service implementation.
func NewQuerySvc(resourceSearcher port.ResourceSearcher,
	accessControlChecker port.AccessControlChecker,
	organizationSearcher port.OrganizationSearcher,
) querysvc.Service {
	resourceService := usecase.NewResourceSearch(resourceSearcher, accessControlChecker)
	organizationService := usecase.NewOrganizationSearch(organizationSearcher)
	return &querySvcsrvc{
		resourceService:     resourceService,
		organizationService: organizationService,
	}
}
