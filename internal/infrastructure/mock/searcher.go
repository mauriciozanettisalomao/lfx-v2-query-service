// Copyright The Linux Foundation and each contributor to LFX.
// SPDX-License-Identifier: MIT

package mock

import (
	"context"
	"log/slog"
	"strings"

	"github.com/linuxfoundation/lfx-v2-query-service/internal/domain/model"
)

// MockResourceSearcher is a mock implementation of ResourceSearcher for testing
// This demonstrates how the clean architecture allows easy swapping of implementations
type MockResourceSearcher struct {
	resources []model.Resource
}

// NewMockResourceSearcher creates a new mock searcher with some sample data
func NewMockResourceSearcher() *MockResourceSearcher {
	return &MockResourceSearcher{
		resources: []model.Resource{
			{
				Type: "committee",
				ID:   "123",
				Data: map[string]any{
					"name":        "Technical Advisory Committee",
					"description": "Main technical governance body",
					"status":      "active",
					"tags":        []string{"active", "governance"},
				},
				TransactionBodyStub: model.TransactionBodyStub{
					ObjectRef:            "committee:123",
					ObjectType:           "committee",
					ObjectID:             "123",
					Public:               false,
					AccessCheckObject:    "committee:123",
					AccessCheckRelation:  "member",
					HistoryCheckObject:   "committee:123",
					HistoryCheckRelation: "viewer",
				},
				NeedCheck: true,
			},
			{
				Type: "project",
				ID:   "456",
				Data: map[string]any{
					"name":        "LFX Platform Project",
					"slug":        "lfx-platform-project",
					"description": "Core platform development project",
					"status":      "active",
					"tags":        []string{"active", "platform"},
				},
				TransactionBodyStub: model.TransactionBodyStub{
					ObjectRef:            "project:456",
					ObjectType:           "project",
					ObjectID:             "456",
					Public:               true,
					AccessCheckObject:    "project:456",
					AccessCheckRelation:  "viewer",
					HistoryCheckObject:   "project:456",
					HistoryCheckRelation: "viewer",
				},
				NeedCheck: false,
			},
			{
				Type: "committee",
				ID:   "567",
				Data: map[string]any{
					"name":        "Security Committee",
					"description": "Handles security-related matters",
					"status":      "active",
					"tags":        []string{"active", "security"},
				},
				TransactionBodyStub: model.TransactionBodyStub{
					ObjectRef:            "committee:567",
					ObjectType:           "committee",
					ObjectID:             "567",
					Public:               false,
					AccessCheckObject:    "committee:567",
					AccessCheckRelation:  "member",
					HistoryCheckObject:   "committee:567",
					HistoryCheckRelation: "viewer",
				},
				NeedCheck: true,
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
				TransactionBodyStub: model.TransactionBodyStub{
					ObjectRef:            "meeting:101",
					ObjectType:           "meeting",
					ObjectID:             "101",
					Public:               false,
					AccessCheckObject:    "", // Empty to simulate missing access control info
					AccessCheckRelation:  "",
					HistoryCheckObject:   "meeting:101",
					HistoryCheckRelation: "viewer",
				},
				NeedCheck: true,
			},
			{
				Type: "project",
				ID:   "789",
				Data: map[string]any{
					"name":        "Internal Security Project",
					"slug":        "internal-security-project",
					"description": "Private security-focused project",
					"status":      "active",
					"tags":        []string{"active", "security", "private"},
				},
				TransactionBodyStub: model.TransactionBodyStub{
					ObjectRef:            "project:789",
					ObjectType:           "project",
					ObjectID:             "789",
					Public:               false,
					AccessCheckObject:    "project:789",
					AccessCheckRelation:  "contributor",
					HistoryCheckObject:   "project:789",
					HistoryCheckRelation: "viewer",
				},
				NeedCheck: true,
			},
		},
	}
}

// QueryResources implements the ResourceSearcher interface with mock data
func (m *MockResourceSearcher) QueryResources(ctx context.Context, criteria model.SearchCriteria) (*model.SearchResult, error) {
	slog.DebugContext(ctx, "executing mock search", "criteria", criteria)

	var filteredResources []model.Resource

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
		var nameFilteredResources []model.Resource
		searchName := strings.ToLower(*criteria.Name)

		for _, resource := range filteredResources {
			if data, ok := resource.Data.(map[string]interface{}); ok {
				// Check name field
				nameMatch := false
				if name, ok := data["name"].(string); ok {
					if strings.Contains(strings.ToLower(name), searchName) {
						nameMatch = true
					}
				}

				// For projects, also check slug field
				if !nameMatch && resource.Type == "project" {
					if slug, ok := data["slug"].(string); ok {
						if strings.Contains(strings.ToLower(slug), searchName) {
							nameMatch = true
						}
					}
				}

				if nameMatch {
					nameFilteredResources = append(nameFilteredResources, resource)
				}
			}
		}
		filteredResources = nameFilteredResources
	}

	// Filter by tags
	if len(criteria.Tags) > 0 {
		var tagFilteredResources []model.Resource

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

	result := &model.SearchResult{
		Resources: filteredResources,
	}

	slog.DebugContext(ctx, "mock search completed", "results_count", len(result.Resources))
	return result, nil
}

// IsReady implements the ResourceSearcher interface (always ready for mock)
func (m *MockResourceSearcher) IsReady(ctx context.Context) error {
	return nil
}

// sortResources sorts the resources based on the sort criteria
func (m *MockResourceSearcher) sortResources(resources []model.Resource, sort string) {
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
func (m *MockResourceSearcher) AddResource(resource model.Resource) {
	// Ensure the resource has proper access control fields if not already set
	if resource.ObjectRef == "" {
		resource.ObjectRef = resource.Type + ":" + resource.ID
	}
	if resource.ObjectType == "" {
		resource.ObjectType = resource.Type
	}
	if resource.ObjectID == "" {
		resource.ObjectID = resource.ID
	}

	// Set default access control values if not specified
	if resource.AccessCheckObject == "" && resource.AccessCheckRelation == "" {
		// Default to requiring access check with reasonable defaults
		resource.AccessCheckObject = resource.Type + ":" + resource.ID
		resource.AccessCheckRelation = "viewer"
		resource.NeedCheck = true
	}

	m.resources = append(m.resources, resource)
}

// NewResourceWithDefaults creates a new resource with proper default access control fields
func NewResourceWithDefaults(resourceType, id string, data map[string]any, isPublic bool) model.Resource {
	// For projects, ensure slug is included if not present
	if resourceType == "project" {
		if _, hasSlug := data["slug"]; !hasSlug {
			if name, hasName := data["name"].(string); hasName {
				// Generate slug from name: lowercase, replace spaces with hyphens
				slug := strings.ToLower(strings.ReplaceAll(name, " ", "-"))
				data["slug"] = slug
			}
		}
	}

	resource := model.Resource{
		Type: resourceType,
		ID:   id,
		Data: data,
		TransactionBodyStub: model.TransactionBodyStub{
			ObjectRef:            resourceType + ":" + id,
			ObjectType:           resourceType,
			ObjectID:             id,
			Public:               isPublic,
			AccessCheckObject:    resourceType + ":" + id,
			AccessCheckRelation:  "viewer",
			HistoryCheckObject:   resourceType + ":" + id,
			HistoryCheckRelation: "viewer",
		},
		NeedCheck: !isPublic,
	}

	// If not public, set appropriate access check defaults
	if !isPublic {
		switch resourceType {
		case "committee":
			resource.AccessCheckRelation = "member"
		case "project":
			resource.AccessCheckRelation = "viewer"
		case "meeting":
			resource.AccessCheckRelation = "attendee"
		default:
			resource.AccessCheckRelation = "viewer"
		}
	}

	return resource
}

// ClearResources clears all resources (useful for testing)
func (m *MockResourceSearcher) ClearResources() {
	m.resources = []model.Resource{}
}

// GetResourceCount returns the total number of resources
func (m *MockResourceSearcher) GetResourceCount() int {
	return len(m.resources)
}
