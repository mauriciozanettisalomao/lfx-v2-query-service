// Copyright The Linux Foundation and each contributor to LFX.
// SPDX-License-Identifier: MIT

package mock

import (
	"context"
	"log/slog"
	"strings"

	"github.com/linuxfoundation/lfx-v2-query-service/internal/domain"
)

// MockResourceSearcher is a mock implementation of ResourceSearcher for testing
// This demonstrates how the clean architecture allows easy swapping of implementations
type MockResourceSearcher struct {
	resources []domain.Resource
}

// NewMockResourceSearcher creates a new mock searcher with some sample data
func NewMockResourceSearcher() *MockResourceSearcher {
	return &MockResourceSearcher{
		resources: []domain.Resource{
			{
				Type: "committee",
				ID:   "123",
				Data: map[string]any{
					"name":        "Technical Advisory Committee",
					"description": "Main technical governance body",
					"status":      "active",
					"tags":        []string{"active", "governance"},
				},
			},
			{
				Type: "project",
				ID:   "456",
				Data: map[string]any{
					"name":        "LFX Platform Project",
					"description": "Core platform development project",
					"status":      "active",
					"tags":        []string{"active", "platform"},
				},
			},
			{
				Type: "committee",
				ID:   "789",
				Data: map[string]any{
					"name":        "Security Committee",
					"description": "Handles security-related matters",
					"status":      "active",
					"tags":        []string{"active", "security"},
				},
			},
			{
				Type: "meeting",
				ID:   "101",
				Data: map[string]any{
					"name":        "Monthly Board Meeting",
					"description": "Regular board meeting for project governance",
					"status":      "active",
					"tags":        []string{"active", "governance"},
				},
			},
		},
	}
}

// QueryResources implements the ResourceSearcher interface with mock data
func (m *MockResourceSearcher) QueryResources(ctx context.Context, criteria domain.SearchCriteria) (*domain.SearchResult, error) {
	slog.DebugContext(ctx, "executing mock search", "criteria", criteria)

	var filteredResources []domain.Resource

	// Filter by type
	if criteria.ResourceType != nil {
		for _, resource := range m.resources {
			if resource.Type == *criteria.ResourceType {
				filteredResources = append(filteredResources, resource)
			}
		}
	} else {
		filteredResources = m.resources
	}

	// Filter by name (case-insensitive substring search)
	if criteria.Name != nil {
		var nameFilteredResources []domain.Resource
		searchName := strings.ToLower(*criteria.Name)

		for _, resource := range filteredResources {
			if data, ok := resource.Data.(map[string]interface{}); ok {
				if name, ok := data["name"].(string); ok {
					if strings.Contains(strings.ToLower(name), searchName) {
						nameFilteredResources = append(nameFilteredResources, resource)
					}
				}
			}
		}
		filteredResources = nameFilteredResources
	}

	// Filter by tags
	if len(criteria.Tags) > 0 {
		var tagFilteredResources []domain.Resource

		for _, resource := range filteredResources {
			if data, ok := resource.Data.(map[string]interface{}); ok {
				if resourceTags, ok := data["tags"].([]string); ok {
					// Check if resource has any of the requested tags
					for _, requestedTag := range criteria.Tags {
						for _, resourceTag := range resourceTags {
							if requestedTag == resourceTag {
								tagFilteredResources = append(tagFilteredResources, resource)
								goto nextResource
							}
						}
					}
				}
			}
		nextResource:
		}
		filteredResources = tagFilteredResources
	}

	// Sort results (simplified implementation)
	m.sortResources(filteredResources, criteria.SortBy)

	result := &domain.SearchResult{
		Resources: filteredResources,
	}

	slog.DebugContext(ctx, "mock search completed", "results_count", len(result.Resources))
	return result, nil
}

// sortResources sorts the resources based on the sort criteria
func (m *MockResourceSearcher) sortResources(resources []domain.Resource, sort string) {
	// This is a simplified sorting implementation
	// In a real implementation, you'd use proper sorting algorithms

	if sort == "name_desc" {
		// Reverse the order for descending sort
		for i := len(resources)/2 - 1; i >= 0; i-- {
			opp := len(resources) - 1 - i
			resources[i], resources[opp] = resources[opp], resources[i]
		}
	}
}

// AddResource adds a resource to the mock data (useful for testing)
func (m *MockResourceSearcher) AddResource(resource domain.Resource) {
	m.resources = append(m.resources, resource)
}

// ClearResources clears all resources (useful for testing)
func (m *MockResourceSearcher) ClearResources() {
	m.resources = []domain.Resource{}
}

// GetResourceCount returns the total number of resources
func (m *MockResourceSearcher) GetResourceCount() int {
	return len(m.resources)
}
