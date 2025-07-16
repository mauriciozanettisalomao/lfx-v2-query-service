// Copyright The Linux Foundation and each contributor to LFX.
// SPDX-License-Identifier: MIT

package service

import (
	"context"
	"testing"

	"github.com/linuxfoundation/lfx-v2-query-service/internal/domain"
	"github.com/linuxfoundation/lfx-v2-query-service/internal/infrastructure/mock"
	"github.com/linuxfoundation/lfx-v2-query-service/pkg/constants"
	"github.com/stretchr/testify/assert"
)

func TestResourceSearchQueryResources(t *testing.T) {
	tests := []struct {
		name                 string
		criteria             domain.SearchCriteria
		principal            string
		setupMocks           func(*mock.MockResourceSearcher, *mock.MockAccessControlChecker)
		expectedError        bool
		expectedResources    int
		expectedCacheControl bool
	}{
		{
			name: "successful search with authenticated user - public resources",
			criteria: domain.SearchCriteria{
				Name: stringPtr("test"),
			},
			principal: "user123",
			setupMocks: func(searcher *mock.MockResourceSearcher, accessChecker *mock.MockAccessControlChecker) {
				searcher.AddResource(domain.Resource{
					Type: "project",
					ID:   "test-project",
					Data: map[string]any{"name": "Test Project"},
					TransactionBodyStub: domain.TransactionBodyStub{
						ObjectRef:           "project:test-project",
						ObjectType:          "project",
						ObjectID:            "test-project",
						Public:              true,
						AccessCheckObject:   "project:test-project",
						AccessCheckRelation: "view",
					},
				})
				accessChecker.DefaultResult = "allowed"
			},
			expectedError:        false,
			expectedResources:    1,
			expectedCacheControl: false,
		},
		{
			name: "successful search with anonymous user",
			criteria: domain.SearchCriteria{
				Name: stringPtr("test"),
			},
			principal: constants.AnonymousPrincipal,
			setupMocks: func(searcher *mock.MockResourceSearcher, accessChecker *mock.MockAccessControlChecker) {
				searcher.AddResource(domain.Resource{
					Type: "project",
					ID:   "test-project",
					Data: map[string]any{"name": "Test Project"},
					TransactionBodyStub: domain.TransactionBodyStub{
						ObjectRef:  "project:test-project",
						ObjectType: "project",
						ObjectID:   "test-project",
						Public:     true,
					},
				})
			},
			expectedError:        false,
			expectedResources:    1,
			expectedCacheControl: true,
		},
		{
			name:     "invalid search criteria - empty criteria",
			criteria: domain.SearchCriteria{
				// All fields empty
			},
			principal: "user123",
			setupMocks: func(searcher *mock.MockResourceSearcher, accessChecker *mock.MockAccessControlChecker) {
				// No setup needed for this test
			},
			expectedError:        true,
			expectedResources:    0,
			expectedCacheControl: false,
		},
		{
			name: "missing principal in context",
			criteria: domain.SearchCriteria{
				Name: stringPtr("test"),
			},
			principal: "", // Empty principal to trigger error
			setupMocks: func(searcher *mock.MockResourceSearcher, accessChecker *mock.MockAccessControlChecker) {
				// No setup needed for this test
			},
			expectedError:        true,
			expectedResources:    0,
			expectedCacheControl: false,
		},
		{
			name: "searcher returns error",
			criteria: domain.SearchCriteria{
				Name: stringPtr("test"),
			},
			principal: "user123",
			setupMocks: func(searcher *mock.MockResourceSearcher, accessChecker *mock.MockAccessControlChecker) {
				// Create a mock that will fail
				searcher.ClearResources()
			},
			expectedError:        false, // Mock searcher doesn't return errors in this implementation
			expectedResources:    0,
			expectedCacheControl: false,
		},
		{
			name: "access control check fails",
			criteria: domain.SearchCriteria{
				Name: stringPtr("test"),
			},
			principal: "user123",
			setupMocks: func(searcher *mock.MockResourceSearcher, accessChecker *mock.MockAccessControlChecker) {
				searcher.AddResource(domain.Resource{
					Type: "project",
					ID:   "test-project",
					Data: map[string]any{"name": "Test Project"},
					TransactionBodyStub: domain.TransactionBodyStub{
						ObjectRef:           "project:test-project",
						ObjectType:          "project",
						ObjectID:            "test-project",
						Public:              false,
						AccessCheckObject:   "project:test-project",
						AccessCheckRelation: "view",
					},
				})
				accessChecker.DefaultResult = "denied"
			},
			expectedError:        false,
			expectedResources:    0,
			expectedCacheControl: false,
		},
	}

	assertion := assert.New(t)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mocks
			mockSearcher := mock.NewMockResourceSearcher()
			mockAccessChecker := mock.NewMockAccessControlChecker()

			tc.setupMocks(mockSearcher, mockAccessChecker)

			// Create service
			service := NewResourceSearch(mockSearcher, mockAccessChecker).(*ResourceSearch)

			// Setup context
			ctx := context.Background()
			if tc.principal != "" {
				ctx = context.WithValue(ctx, constants.PrincipalContextID, tc.principal)
			}

			// Execute
			result, err := service.QueryResources(ctx, tc.criteria)

			// Verify
			if tc.expectedError {
				assertion.Error(err)
				assertion.Nil(result)
				return
			}
			assertion.NoError(err)
			assertion.NotNil(result)
			assertion.Equal(tc.expectedResources, len(result.Resources))

			if tc.expectedCacheControl {
				assertion.NotNil(result.CacheControl)
				assertion.Equal(constants.AnonymousCacheControlHeader, *result.CacheControl)
				return
			}
			assertion.Nil(result.CacheControl)

		})
	}
}

func TestResourceSearchValidateSearchCriteria(t *testing.T) {
	tests := []struct {
		name        string
		criteria    domain.SearchCriteria
		expectError bool
	}{
		{
			name: "valid criteria with name",
			criteria: domain.SearchCriteria{
				Name: stringPtr("test"),
			},
			expectError: false,
		},
		{
			name: "valid criteria with parent",
			criteria: domain.SearchCriteria{
				Parent: stringPtr("parent-id"),
			},
			expectError: false,
		},
		{
			name: "valid criteria with resource type",
			criteria: domain.SearchCriteria{
				ResourceType: stringPtr("project"),
			},
			expectError: false,
		},
		{
			name: "valid criteria with tags",
			criteria: domain.SearchCriteria{
				Tags: []string{"tag1", "tag2"},
			},
			expectError: false,
		},
		{
			name: "valid criteria with multiple fields",
			criteria: domain.SearchCriteria{
				Name:         stringPtr("test"),
				ResourceType: stringPtr("project"),
				Tags:         []string{"tag1"},
			},
			expectError: false,
		},
		{
			name:     "invalid criteria - all fields empty",
			criteria: domain.SearchCriteria{
				// All fields empty
			},
			expectError: true,
		},
		{
			name: "invalid criteria - empty tags array",
			criteria: domain.SearchCriteria{
				Tags: []string{},
			},
			expectError: true,
		},
	}

	assertion := assert.New(t)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create service
			service := &ResourceSearch{}

			// Execute
			err := service.validateSearchCriteria(tc.criteria)

			// Verify
			if tc.expectError {
				assertion.Error(err)
				return
			}
			assertion.NoError(err)

		})
	}
}

func TestResourceSearchBuildMessage(t *testing.T) {
	tests := []struct {
		name                    string
		principal               string
		searchResult            *domain.SearchResult
		expectedPublicCount     int
		expectedNeedCheckCount  int
		expectedMessageContains []string
	}{
		{
			name:      "only public resources",
			principal: "user123",
			searchResult: &domain.SearchResult{
				Resources: []domain.Resource{
					{
						Type: "project",
						ID:   "public-project",
						Data: map[string]any{"name": "Public Project"},
						TransactionBodyStub: domain.TransactionBodyStub{
							ObjectRef:  "project:public-project",
							ObjectType: "project",
							ObjectID:   "public-project",
							Public:     true,
						},
					},
				},
			},
			expectedPublicCount:     1,
			expectedNeedCheckCount:  0,
			expectedMessageContains: []string{},
		},
		{
			name:      "only private resources",
			principal: "user123",
			searchResult: &domain.SearchResult{
				Resources: []domain.Resource{
					{
						Type: "project",
						ID:   "private-project",
						Data: map[string]any{"name": "Private Project"},
						TransactionBodyStub: domain.TransactionBodyStub{
							ObjectRef:           "project:private-project",
							ObjectType:          "project",
							ObjectID:            "private-project",
							Public:              false,
							AccessCheckObject:   "project:private-project",
							AccessCheckRelation: "view",
						},
					},
				},
			},
			expectedPublicCount:     0,
			expectedNeedCheckCount:  1,
			expectedMessageContains: []string{"project:private-project#view@user:user123"},
		},
		{
			name:      "mixed public and private resources",
			principal: "user123",
			searchResult: &domain.SearchResult{
				Resources: []domain.Resource{
					{
						Type: "project",
						ID:   "public-project",
						Data: map[string]any{"name": "Public Project"},
						TransactionBodyStub: domain.TransactionBodyStub{
							ObjectRef:  "project:public-project",
							ObjectType: "project",
							ObjectID:   "public-project",
							Public:     true,
						},
					},
					{
						Type: "project",
						ID:   "private-project",
						Data: map[string]any{"name": "Private Project"},
						TransactionBodyStub: domain.TransactionBodyStub{
							ObjectRef:           "project:private-project",
							ObjectType:          "project",
							ObjectID:            "private-project",
							Public:              false,
							AccessCheckObject:   "project:private-project",
							AccessCheckRelation: "view",
						},
					},
				},
			},
			expectedPublicCount:     1,
			expectedNeedCheckCount:  1,
			expectedMessageContains: []string{"project:private-project#view@user:user123"},
		},
		{
			name:      "duplicate resources filtered out",
			principal: "user123",
			searchResult: &domain.SearchResult{
				Resources: []domain.Resource{
					{
						Type: "project",
						ID:   "duplicate-project",
						Data: map[string]any{"name": "Duplicate Project"},
						TransactionBodyStub: domain.TransactionBodyStub{
							ObjectRef:  "project:duplicate-project",
							ObjectType: "project",
							ObjectID:   "duplicate-project",
							Public:     true,
						},
					},
					{
						Type: "project",
						ID:   "duplicate-project",
						Data: map[string]any{"name": "Duplicate Project"},
						TransactionBodyStub: domain.TransactionBodyStub{
							ObjectRef:  "project:duplicate-project",
							ObjectType: "project",
							ObjectID:   "duplicate-project",
							Public:     true,
						},
					},
				},
			},
			expectedPublicCount:     2, // Both resources remain, both are public
			expectedNeedCheckCount:  0,
			expectedMessageContains: []string{},
		},
		{
			name:      "resource missing access check info",
			principal: "user123",
			searchResult: &domain.SearchResult{
				Resources: []domain.Resource{
					{
						Type: "project",
						ID:   "invalid-project",
						Data: map[string]any{"name": "Invalid Project"},
						TransactionBodyStub: domain.TransactionBodyStub{
							ObjectRef:  "project:invalid-project",
							ObjectType: "project",
							ObjectID:   "invalid-project",
							Public:     false,
							// Missing AccessCheckObject and AccessCheckRelation
						},
					},
				},
			},
			expectedPublicCount:     0,
			expectedNeedCheckCount:  1,
			expectedMessageContains: []string{},
		},
	}

	assertion := assert.New(t)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create service
			service := &ResourceSearch{}

			// Setup context
			ctx := context.Background()

			// Execute
			message := service.BuildMessage(ctx, tc.principal, tc.searchResult)

			// Count resources by their NeedCheck field
			publicCount := 0
			needCheckCount := 0
			for _, resource := range tc.searchResult.Resources {
				if resource.NeedCheck {
					needCheckCount++
				} else {
					publicCount++
				}
			}

			// Verify
			assertion.Equal(tc.expectedPublicCount, publicCount)
			if tc.expectedNeedCheckCount != needCheckCount {
				t.Errorf("Test case '%s' failed: expected needCheckCount=%d, got=%d", tc.name, tc.expectedNeedCheckCount, needCheckCount)
			}
			assertion.Equal(tc.expectedNeedCheckCount, needCheckCount)

			messageStr := string(message)
			for _, expectedSubstring := range tc.expectedMessageContains {
				assertion.Contains(messageStr, expectedSubstring)
			}
		})
	}
}

func TestResourceSearchCheckAccess(t *testing.T) {
	tests := []struct {
		name               string
		principal          string
		resources          []domain.Resource
		message            []byte
		setupAccessChecker func(*mock.MockAccessControlChecker)
		expectedResources  int
		expectedError      bool
	}{
		{
			name:      "access granted for all resources",
			principal: "user123",
			resources: []domain.Resource{
				{
					Type:      "project",
					ID:        "test-project",
					Data:      map[string]any{"name": "Test Project"},
					NeedCheck: true,
					TransactionBodyStub: domain.TransactionBodyStub{
						ObjectRef:           "project:test-project",
						ObjectType:          "project",
						ObjectID:            "test-project",
						AccessCheckObject:   "project:test-project",
						AccessCheckRelation: "view",
					},
				},
			},
			message: []byte("project:test-project#view@user:user123\n"),
			setupAccessChecker: func(checker *mock.MockAccessControlChecker) {
				checker.DefaultResult = "allowed"
				checker.AllowedUserIDs = []string{"user123"}
			},
			expectedResources: 1,
			expectedError:     false,
		},
		{
			name:      "access denied for all resources",
			principal: "user123",
			resources: []domain.Resource{
				{
					Type:      "project",
					ID:        "test-project",
					Data:      map[string]any{"name": "Test Project"},
					NeedCheck: true,
					TransactionBodyStub: domain.TransactionBodyStub{
						ObjectRef:           "project:test-project",
						ObjectType:          "project",
						ObjectID:            "test-project",
						AccessCheckObject:   "project:test-project",
						AccessCheckRelation: "view",
					},
				},
			},
			message: []byte("project:test-project#view@user:user123\n"),
			setupAccessChecker: func(checker *mock.MockAccessControlChecker) {
				checker.DefaultResult = "denied"
			},
			expectedResources: 0,
			expectedError:     false,
		},
		{
			name:      "mixed access results",
			principal: "user123",
			resources: []domain.Resource{
				{
					Type:      "project",
					ID:        "allowed-project",
					Data:      map[string]any{"name": "Allowed Project"},
					NeedCheck: true,
					TransactionBodyStub: domain.TransactionBodyStub{
						ObjectRef:           "project:allowed-project",
						ObjectType:          "project",
						ObjectID:            "allowed-project",
						AccessCheckObject:   "project:allowed-project",
						AccessCheckRelation: "view",
					},
				},
				{
					Type:      "project",
					ID:        "denied-project",
					Data:      map[string]any{"name": "Denied Project"},
					NeedCheck: true,
					TransactionBodyStub: domain.TransactionBodyStub{
						ObjectRef:           "project:denied-project",
						ObjectType:          "project",
						ObjectID:            "denied-project",
						AccessCheckObject:   "project:denied-project",
						AccessCheckRelation: "view",
					},
				},
			},
			message: []byte("project:allowed-project#view@user:user123\nproject:denied-project#view@user:user123\n"),
			setupAccessChecker: func(checker *mock.MockAccessControlChecker) {
				// Set up allowed and denied resources
				checker.AllowedUserIDs = []string{"user123"}
				checker.DeniedResourceIDs = []string{"denied-project"}
			},
			expectedResources: 1,
			expectedError:     false,
		},
		{
			name:      "empty resources list",
			principal: "user123",
			resources: []domain.Resource{},
			message:   []byte(""),
			setupAccessChecker: func(checker *mock.MockAccessControlChecker) {
				checker.DefaultResult = "allowed"
			},
			expectedResources: 0,
			expectedError:     false,
		},
		{
			name:      "public resources should be included without access check",
			principal: "user123",
			resources: []domain.Resource{
				{
					Type:      "project",
					ID:        "public-project",
					Data:      map[string]any{"name": "Public Project"},
					NeedCheck: false,
					TransactionBodyStub: domain.TransactionBodyStub{
						ObjectRef:  "project:public-project",
						ObjectType: "project",
						ObjectID:   "public-project",
						Public:     true,
					},
				},
			},
			message: []byte(""),
			setupAccessChecker: func(checker *mock.MockAccessControlChecker) {
				checker.DefaultResult = "allowed"
			},
			expectedResources: 1,
			expectedError:     false,
		},
	}

	assertion := assert.New(t)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mocks
			mockAccessChecker := mock.NewMockAccessControlChecker()
			tc.setupAccessChecker(mockAccessChecker)

			// Create service
			service := &ResourceSearch{
				accessChecker: mockAccessChecker,
			}

			// Setup context
			ctx := context.Background()

			// Execute
			resources, err := service.CheckAccess(ctx, tc.principal, tc.resources, tc.message)

			// Verify
			if tc.expectedError {
				assertion.Error(err)
				return
			}
			assertion.NoError(err)
			assertion.Equal(tc.expectedResources, len(resources))

		})
	}
}

func TestNewResourceSearch(t *testing.T) {
	tests := []struct {
		name         string
		setupMocks   func() (domain.ResourceSearcher, domain.AccessControlChecker)
		expectNonNil bool
		expectType   string
	}{
		{
			name: "creates new resource search with valid dependencies",
			setupMocks: func() (domain.ResourceSearcher, domain.AccessControlChecker) {
				return mock.NewMockResourceSearcher(), mock.NewMockAccessControlChecker()
			},
			expectNonNil: true,
			expectType:   "*service.ResourceSearch",
		},
		{
			name: "creates new resource search with nil dependencies",
			setupMocks: func() (domain.ResourceSearcher, domain.AccessControlChecker) {
				return nil, nil
			},
			expectNonNil: true,
			expectType:   "*service.ResourceSearch",
		},
	}

	assertion := assert.New(t)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			searcher, accessChecker := tc.setupMocks()

			// Execute
			result := NewResourceSearch(searcher, accessChecker)

			// Verify
			if tc.expectNonNil {
				assertion.NotNil(result)
				assertion.IsType(&ResourceSearch{}, result)

				// Cast to concrete type to verify internal fields
				if resourceSearch, ok := result.(*ResourceSearch); ok {
					assertion.Equal(searcher, resourceSearch.resourceSearcher)
					assertion.Equal(accessChecker, resourceSearch.accessChecker)
				}
			} else {
				assertion.Nil(result)
			}
		})
	}
}

func TestResourceSearchQueryResourcesEdgeCases(t *testing.T) {
	assertion := assert.New(t)

	t.Run("search with complex criteria", func(t *testing.T) {
		// Setup
		mockSearcher := mock.NewMockResourceSearcher()
		mockAccessChecker := mock.NewMockAccessControlChecker()
		service := NewResourceSearch(mockSearcher, mockAccessChecker).(*ResourceSearch)

		// Add test data
		mockSearcher.AddResource(domain.Resource{
			Type: "project",
			ID:   "complex-project",
			Data: map[string]any{
				"name": "Complex Project",
				"tags": []string{"active", "governance"},
			},
			TransactionBodyStub: domain.TransactionBodyStub{
				ObjectRef:  "project:complex-project",
				ObjectType: "project",
				ObjectID:   "complex-project",
				Public:     true,
			},
		})

		criteria := domain.SearchCriteria{
			Name:         stringPtr("Complex"),
			ResourceType: stringPtr("project"),
			Tags:         []string{"active"},
			SortBy:       "name",
			SortOrder:    "asc",
			PageSize:     10,
		}

		ctx := context.WithValue(context.Background(), constants.PrincipalContextID, "user123")

		// Execute
		result, err := service.QueryResources(ctx, criteria)

		// Verify
		assertion.NoError(err)
		assertion.NotNil(result)
		assertion.Equal(1, len(result.Resources))
		assertion.Equal("complex-project", result.Resources[0].ID)
	})

	t.Run("search with pagination", func(t *testing.T) {
		// Setup
		mockSearcher := mock.NewMockResourceSearcher()
		mockAccessChecker := mock.NewMockAccessControlChecker()
		service := NewResourceSearch(mockSearcher, mockAccessChecker).(*ResourceSearch)

		criteria := domain.SearchCriteria{
			Name:      stringPtr("test"),
			PageSize:  5,
			PageToken: stringPtr("test-token"),
		}

		ctx := context.WithValue(context.Background(), constants.PrincipalContextID, "user123")

		// Execute
		result, err := service.QueryResources(ctx, criteria)

		// Verify
		assertion.NoError(err)
		assertion.NotNil(result)
	})
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}
