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

	"goa.design/goa/v3/security"
)

// query-svc service implementation using clean architecture.
type querySvcsrvc struct {
	resourceService domain.ResourceSearcher
}

// NewQuerySvc returns the query-svc service implementation.
func NewQuerySvc(resourceSearcher domain.ResourceSearcher) querysvc.Service {
	resourceService := service.NewResourceSearch(resourceSearcher)
	return &querySvcsrvc{
		resourceService: resourceService,
	}
}

// JWTAuth implements the authorization logic for service "query-svc" for the
// "jwt" security scheme.
func (s *querySvcsrvc) JWTAuth(ctx context.Context, token string, scheme *security.JWTScheme) (context.Context, error) {
	//
	// TBD: add authorization logic.
	//
	// In case of authorization failure this function should return
	// one of the generated error structs, e.g.:
	//
	//    return ctx, myservice.MakeUnauthorizedError("invalid token")
	//
	// Alternatively this function may return an instance of
	// goa.ServiceError with a Name field value that matches one of
	// the design error names, e.g:
	//
	//    return ctx, goa.PermanentError("unauthorized", "invalid token")
	//
	//return ctx, fmt.Errorf("not implemented")

	return ctx, nil // No authorization logic implemented yet
}

// Locate resources by their type or parent, or use typeahead search to query
// resources by a display name or similar alias.
func (s *querySvcsrvc) QueryResources(ctx context.Context, p *querysvc.QueryResourcesPayload) (res *querysvc.QueryResourcesResult, err error) {

	slog.DebugContext(ctx, "querySvc.query-resources",
		"name", p.Name,
	)

	// Convert payload to domain criteria
	criteria := s.payloadToCriteria(p)

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
func (s *querySvcsrvc) payloadToCriteria(p *querysvc.QueryResourcesPayload) domain.SearchCriteria {
	criteria := domain.SearchCriteria{
		Name:      p.Name,
		Parent:    p.Parent,
		Type:      p.Type,
		Tags:      p.Tags,
		Sort:      p.Sort,
		PageToken: p.PageToken,
	}
	return criteria
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
