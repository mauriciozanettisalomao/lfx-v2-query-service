// Copyright The Linux Foundation and each contributor to LFX.
// SPDX-License-Identifier: MIT

package opensearch

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/linuxfoundation/lfx-v2-query-service/internal/domain"
)

// OpenSearchSearcher implements the ResourceSearcher interface for OpenSearch
type OpenSearchSearcher struct {
	client    OpenSearchClient
	templates *SearchTemplates
	index     string
}

// OpenSearchClient defines the interface for OpenSearch operations
// This allows for easy mocking and testing
type OpenSearchClient interface {
	Search(ctx context.Context, index string, query string) (*SearchResponse, error)
	IsHealthy(ctx context.Context) error
}

// QueryResources implements the ResourceSearcher interface
func (os *OpenSearchSearcher) QueryResources(ctx context.Context, criteria domain.SearchCriteria) (*domain.SearchResult, error) {
	slog.DebugContext(ctx, "executing opensearch query for criteria",
		"criteria", criteria,
	)

	// Prepare template data
	templateData := os.prepareTemplateData(criteria)

	// Render the appropriate query template
	query, err := os.renderQuery(templateData)
	if err != nil {
		return nil, fmt.Errorf("failed to render query: %w", err)
	}

	// Execute the search
	response, err := os.client.Search(ctx, os.index, query)
	if err != nil {
		return nil, fmt.Errorf("opensearch search failed: %w", err)
	}

	// Convert response to domain objects
	result, err := os.convertResponse(ctx, response)
	if err != nil {
		return nil, fmt.Errorf("failed to convert search response: %w", err)
	}

	slog.DebugContext(ctx, "opensearch search completed",
		"results_count", len(result.Resources),
	)
	return result, nil
}

// prepareTemplateData converts domain search criteria to template data
func (os *OpenSearchSearcher) prepareTemplateData(criteria domain.SearchCriteria) TemplateData {
	data := TemplateData{
		Sort: criteria.Sort,
		Size: 50, // Default page size
		From: 0,
	}

	if criteria.Name != nil {
		data.Name = *criteria.Name
	}
	if criteria.Type != nil {
		data.Type = *criteria.Type
	}
	if criteria.Parent != nil {
		data.Parent = *criteria.Parent
	}
	if criteria.Tags != nil {
		data.Tags = criteria.Tags
	}
	if criteria.PageToken != nil {
		data.PageToken = *criteria.PageToken
		// Parse page token to set pagination offset
		// This is a simplified implementation - in production you'd use a more sophisticated approach
		if offset := parsePageToken(*criteria.PageToken); offset > 0 {
			data.From = offset
		}
	}

	return data
}

// renderQuery renders the appropriate query template based on the search criteria
func (os *OpenSearchSearcher) renderQuery(data TemplateData) (string, error) {
	// If this is a typeahead search (name only with no other filters), use typeahead template
	if data.Name != "" && data.Type == "" && data.Parent == "" && len(data.Tags) == 0 {
		return os.templates.RenderTypeaheadQuery(data)
	}

	// Otherwise, use the full resource search template
	return os.templates.RenderResourceSearchQuery(data)
}

// convertResponse converts OpenSearch response to domain objects
func (os *OpenSearchSearcher) convertResponse(ctx context.Context, response *SearchResponse) (*domain.SearchResult, error) {
	result := &domain.SearchResult{
		Resources: make([]domain.Resource, 0, len(response.Hits.Hits)),
	}

	for _, hit := range response.Hits.Hits {
		resource, err := os.convertHit(hit)
		if err != nil {
			// Log error but continue processing other hits
			slog.ErrorContext(ctx, "failed to convert hit", "hit_id", hit.ID, "error", err)
			continue
		}
		result.Resources = append(result.Resources, resource)
	}

	// Generate next page token if there are more results
	if len(response.Hits.Hits) > 0 && response.Hits.Total.Value > int64(len(result.Resources)) {
		nextPageToken := generatePageToken(len(result.Resources))
		result.PageToken = &nextPageToken
	}

	return result, nil
}

// convertHit converts a single OpenSearch hit to a domain resource
func (os *OpenSearchSearcher) convertHit(hit Hit) (domain.Resource, error) {
	resource := domain.Resource{
		ID: hit.ID,
	}

	// Parse the source data
	if hit.Source != nil {
		sourceData := make(map[string]interface{})
		if err := json.Unmarshal(hit.Source, &sourceData); err != nil {
			return resource, fmt.Errorf("failed to unmarshal source data: %w", err)
		}

		// Extract type
		if typeVal, ok := sourceData["type"].(string); ok {
			resource.Type = typeVal
		}

		// Extract data
		if dataVal, ok := sourceData["data"]; ok {
			resource.Data = dataVal
		} else {
			// If no separate data field, use the entire source as data
			resource.Data = sourceData
		}
	}

	return resource, nil
}

// parsePageToken parses the page token to extract pagination offset
func parsePageToken(token string) int {
	// This is a simplified implementation
	// In production, you'd use a more sophisticated approach like base64 encoding
	if strings.HasPrefix(token, "offset_") {
		offsetStr := strings.TrimPrefix(token, "offset_")
		var offset int
		if _, err := fmt.Sscanf(offsetStr, "%d", &offset); err == nil {
			return offset
		}
	}
	return 0
}

// generatePageToken generates a page token for pagination
func generatePageToken(offset int) string {
	// This is a simplified implementation
	// In production, you'd use a more sophisticated approach
	return fmt.Sprintf("offset_%d", offset)
}

// NewOpenSearchSearcher creates a new OpenSearch searcher
func NewOpenSearchSearcher(client OpenSearchClient, index string) (*OpenSearchSearcher, error) {
	templates, err := NewSearchTemplates()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize search templates: %w", err)
	}

	return &OpenSearchSearcher{
		client:    client,
		templates: templates,
		index:     index,
	}, nil
}
