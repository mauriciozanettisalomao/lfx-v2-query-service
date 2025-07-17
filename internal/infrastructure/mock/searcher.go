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
				TransactionBodyStub: domain.TransactionBodyStub{
					ObjectRef:           "committee:123",
					ObjectType:          "committee",
					ObjectID:            "123",
					Public:              true,
					AccessCheckObject:   "committee:123",
					AccessCheckRelation: "view",
				},
			},
			{
				Type: "project",
				ID:   "123",
				Data: map[string]any{
					"name":               "Cloud Native Computing Foundation",
					"slug":               "cncf",
					"description":        "The Cloud Native Computing Foundation (CNCF) hosts critical components of the global technology infrastructure. CNCF brings together the world’s top developers, end users, and vendors and runs the largest open source developer conferences.",
					"status":             "active",
					"logo":               "https://lf-master-project-logos-prod.s3.us-east-2.amazonaws.com/cncf.svg",
					"tags":               []string{"active", "platform"},
					"committees_count":   9,
					"meetings_count":     10,
					"mailing_list_count": 11,
				},
				TransactionBodyStub: domain.TransactionBodyStub{
					ObjectRef:           "project:123",
					ObjectType:          "project",
					ObjectID:            "123",
					Public:              true,
					AccessCheckObject:   "project:123",
					AccessCheckRelation: "view",
				},
			},
			{
				Type: "project",
				ID:   "456",
				Data: map[string]any{
					"name":               "The Linux Foundation",
					"slug":               "tlf",
					"description":        "The Linux Foundation is dedicated to building sustainable ecosystems around open source projects to accelerate technology development and industry adoption. Founded in 2000, the Linux Foundation provides unparalleled support for open source communities through financial and intellectual resources, infrastructure, services, events, and training. Working together, the Linux Foundation and its projects form the most ambitious and successful investment in the creation of shared technology.",
					"status":             "active",
					"logo":               "https://lf-master-project-logos-prod.s3.us-east-2.amazonaws.com/thelinuxfoundation-color.svg",
					"tags":               []string{"active", "platform"},
					"committees_count":   3,
					"meetings_count":     3,
					"mailing_list_count": 3,
				},
				TransactionBodyStub: domain.TransactionBodyStub{
					ObjectRef:           "project:456",
					ObjectType:          "project",
					ObjectID:            "456",
					Public:              true,
					AccessCheckObject:   "project:456",
					AccessCheckRelation: "view",
				},
			},
			{
				Type: "project",
				ID:   "789",
				Data: map[string]any{
					"name":               "Academy Software Foundation",
					"slug":               "aswf",
					"description":        "The mission of the Academy Software Foundation (ASWF) is to increase the quality and quantity of contributions to the content creation industry’s open source software base; to provide a neutral forum to coordinate cross-project efforts; to provide a common build and test infrastructure; and to provide individuals and organizations a clear path to participation in advancing our open source ecosystem.",
					"status":             "active",
					"logo":               "https://lf-master-project-logos-prod.s3.us-east-2.amazonaws.com/aswf.svg",
					"tags":               []string{"active", "platform"},
					"committees_count":   4,
					"meetings_count":     5,
					"mailing_list_count": 6,
				},
				TransactionBodyStub: domain.TransactionBodyStub{
					ObjectRef:           "project:789",
					ObjectType:          "project",
					ObjectID:            "789",
					Public:              true,
					AccessCheckObject:   "project:789",
					AccessCheckRelation: "view",
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
				TransactionBodyStub: domain.TransactionBodyStub{
					ObjectRef:           "committee:789",
					ObjectType:          "committee",
					ObjectID:            "789",
					Public:              true,
					AccessCheckObject:   "committee:789",
					AccessCheckRelation: "view",
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
				TransactionBodyStub: domain.TransactionBodyStub{
					ObjectRef:           "meeting:101",
					ObjectType:          "meeting",
					ObjectID:            "101",
					Public:              true,
					AccessCheckObject:   "meeting:101",
					AccessCheckRelation: "view",
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

	// Filter by slug (exact match, case-insensitive)
	if criteria.Slug != nil {
		var slugFilteredResources []domain.Resource
		searchSlug := strings.ToLower(*criteria.Slug)

		for _, resource := range filteredResources {
			if data, ok := resource.Data.(map[string]interface{}); ok {
				if slug, ok := data["slug"].(string); ok {
					if strings.ToLower(slug) == searchSlug {
						slugFilteredResources = append(slugFilteredResources, resource)
					}
				}
			}
		}
		filteredResources = slugFilteredResources
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
