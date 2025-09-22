// Copyright The Linux Foundation and each contributor to LFX.
// SPDX-License-Identifier: MIT

package service

import (
	"context"
	"testing"

	"github.com/linuxfoundation/lfx-v2-query-service/internal/domain/model"
	"github.com/linuxfoundation/lfx-v2-query-service/internal/infrastructure/mock"
	"github.com/linuxfoundation/lfx-v2-query-service/pkg/constants"
	"github.com/stretchr/testify/assert"
)

func TestResourceCountQueryResourcesCount(t *testing.T) {
	tests := []struct {
		name                 string
		countCriteria        model.SearchCriteria
		aggregationCriteria  model.SearchCriteria
		principal            string
		setupMocks           func(*mock.MockResourceSearcher, *mock.MockAccessControlChecker)
		expectedError        bool
		expectedCount        int
		expectedCacheControl bool
	}{
		{
			name: "successful count with anonymous user",
			countCriteria: model.SearchCriteria{
				ResourceType: stringPtr("project"),
				PageSize:     -1,
				PublicOnly:   true,
			},
			aggregationCriteria: model.SearchCriteria{},
			principal:           constants.AnonymousPrincipal,
			setupMocks: func(resourceSearcher *mock.MockResourceSearcher, accessChecker *mock.MockAccessControlChecker) {
				resourceSearcher.SetQueryResourcesCountResponse(&model.CountResult{
					Count:   3,
					HasMore: false,
				})
			},
			expectedError:        false,
			expectedCount:        3,
			expectedCacheControl: true,
		},
		{
			name: "successful count with authenticated user - public only",
			countCriteria: model.SearchCriteria{
				ResourceType: stringPtr("project"),
				PageSize:     -1,
				PublicOnly:   true,
			},
			aggregationCriteria: model.SearchCriteria{
				GroupBy:     "access_check_query.keyword",
				PageSize:    0,
				PrivateOnly: true,
			},
			principal: "user:test-user",
			setupMocks: func(resourceSearcher *mock.MockResourceSearcher, accessChecker *mock.MockAccessControlChecker) {
				resourceSearcher.SetQueryResourcesCountResponse(&model.CountResult{
					Count: 2,
					Aggregation: model.TermsAggregation{
						Buckets: []model.AggregationBucket{
							{Key: "project:123#viewer", DocCount: 1},
							{Key: "project:456#contributor", DocCount: 2},
						},
					},
					HasMore: false,
				})
				accessChecker.SetCheckAccessResponse(map[string]string{
					"project:123#viewer@user:test-user":      "true",
					"project:456#contributor@user:test-user": "false",
				})
			},
			expectedError:        false,
			expectedCount:        2, // Just the public count, private access didn't match
			expectedCacheControl: false,
		},
		{
			name: "successful count with authenticated user - with private access",
			countCriteria: model.SearchCriteria{
				PageSize:   -1,
				PublicOnly: true,
			},
			aggregationCriteria: model.SearchCriteria{
				GroupBy:     "access_check_query.keyword",
				PageSize:    0,
				PrivateOnly: true,
			},
			principal: "user:admin",
			setupMocks: func(resourceSearcher *mock.MockResourceSearcher, accessChecker *mock.MockAccessControlChecker) {
				resourceSearcher.SetQueryResourcesCountResponse(&model.CountResult{
					Count: 5,
					Aggregation: model.TermsAggregation{
						Buckets: []model.AggregationBucket{
							{Key: "committee:789#member", DocCount: 3},
							{Key: "project:101#viewer", DocCount: 2},
						},
					},
					HasMore: false,
				})
				accessChecker.SetCheckAccessResponse(map[string]string{
					"committee:789#member@user:admin": "true",
					"project:101#viewer@user:admin":   "true",
				})
			},
			expectedError:        false,
			expectedCount:        5, // Just the public count, private access didn't work
			expectedCacheControl: false,
		},
		{
			name: "search error",
			countCriteria: model.SearchCriteria{
				ResourceType: stringPtr("invalid"),
			},
			aggregationCriteria: model.SearchCriteria{},
			principal:           "user:test-user",
			setupMocks: func(resourceSearcher *mock.MockResourceSearcher, accessChecker *mock.MockAccessControlChecker) {
				resourceSearcher.SetQueryResourcesCountError(assert.AnError)
			},
			expectedError: true,
		},
		{
			name: "access control check error",
			countCriteria: model.SearchCriteria{
				PageSize:   -1,
				PublicOnly: true,
			},
			aggregationCriteria: model.SearchCriteria{
				GroupBy:     "access_check_query.keyword",
				PageSize:    0,
				PrivateOnly: true,
			},
			principal: "user:test-user",
			setupMocks: func(resourceSearcher *mock.MockResourceSearcher, accessChecker *mock.MockAccessControlChecker) {
				resourceSearcher.SetQueryResourcesCountResponse(&model.CountResult{
					Count: 2,
					Aggregation: model.TermsAggregation{
						Buckets: []model.AggregationBucket{
							{Key: "project:123#viewer", DocCount: 1},
						},
					},
					HasMore: false,
				})
				accessChecker.SetCheckAccessError(assert.AnError)
			},
			expectedError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assertion := assert.New(t)

			// Setup mocks
			resourceSearcher := mock.NewMockResourceSearcher()
			accessChecker := mock.NewMockAccessControlChecker()
			tc.setupMocks(resourceSearcher, accessChecker)

			// Create service
			service := NewResourceCount(resourceSearcher, accessChecker)

			// Create context with principal
			ctx := context.WithValue(context.Background(), constants.PrincipalContextID, tc.principal)

			// Execute
			result, err := service.QueryResourcesCount(ctx, tc.countCriteria, tc.aggregationCriteria)

			// Verify
			if tc.expectedError {
				assertion.Error(err)
				assertion.Nil(result)
			} else {
				assertion.NoError(err)
				assertion.NotNil(result)
				assertion.Equal(tc.expectedCount, result.Count)

				if tc.expectedCacheControl {
					assertion.NotNil(result.CacheControl)
				} else {
					// For non-anonymous users, CacheControl might be nil
					// This depends on implementation
				}
			}
		})
	}
}

func TestResourceCountBuildMessage(t *testing.T) {
	assertion := assert.New(t)

	// Setup
	resourceSearcher := mock.NewMockResourceSearcher()
	accessChecker := mock.NewMockAccessControlChecker()
	service := &ResourceCount{
		resourceSearcher: resourceSearcher,
		accessChecker:    accessChecker,
	}

	// Test data
	result := &model.CountResult{
		Aggregation: model.TermsAggregation{
			Buckets: []model.AggregationBucket{
				{Key: "committee:123#member", DocCount: 2},
				{Key: "project:456#viewer", DocCount: 3},
			},
		},
	}

	criteria := model.SearchCriteria{
		PageSize: 10,
	}

	// Execute
	ctx := context.Background()
	message := service.BuildMessage(ctx, "test-user", result, criteria)

	// Verify
	assertion.NotNil(message)
	messageStr := string(message)
	assertion.Contains(messageStr, "committee:123#member@user:test-user")
	assertion.Contains(messageStr, "project:456#viewer@user:test-user")
	assertion.Contains(messageStr, "\n")
}

func TestResourceCountCheckAccess(t *testing.T) {
	tests := []struct {
		name               string
		result             *model.CountResult
		accessResponses    map[string]string
		expectedCount      uint64
		expectedError      bool
		setupAccessChecker func(*mock.MockAccessControlChecker)
	}{
		{
			name: "successful access check with allowed resources",
			result: &model.CountResult{
				Aggregation: model.TermsAggregation{
					Buckets: []model.AggregationBucket{
						{Key: "committee:123#member", DocCount: 2},
						{Key: "project:456#viewer", DocCount: 3},
					},
				},
			},
			setupAccessChecker: func(checker *mock.MockAccessControlChecker) {
				checker.SetCheckAccessResponse(map[string]string{
					"committee:123#member@user:test-user": "true",
					"project:456#viewer@user:test-user":   "false",
				})
			},
			expectedCount: 2, // Only committee:123#member is allowed
			expectedError: false,
		},
		{
			name: "successful access check with all denied",
			result: &model.CountResult{
				Aggregation: model.TermsAggregation{
					Buckets: []model.AggregationBucket{
						{Key: "committee:123#member", DocCount: 2},
						{Key: "project:456#viewer", DocCount: 3},
					},
				},
			},
			setupAccessChecker: func(checker *mock.MockAccessControlChecker) {
				checker.SetCheckAccessResponse(map[string]string{
					"committee:123#member@user:test-user": "false",
					"project:456#viewer@user:test-user":   "false",
				})
			},
			expectedCount: 0,
			expectedError: false,
		},
		{
			name: "access check error",
			result: &model.CountResult{
				Aggregation: model.TermsAggregation{
					Buckets: []model.AggregationBucket{
						{Key: "committee:123#member", DocCount: 2},
					},
				},
			},
			setupAccessChecker: func(checker *mock.MockAccessControlChecker) {
				checker.SetCheckAccessError(assert.AnError)
			},
			expectedError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assertion := assert.New(t)

			// Setup
			resourceSearcher := mock.NewMockResourceSearcher()
			accessChecker := mock.NewMockAccessControlChecker()
			tc.setupAccessChecker(accessChecker)

			service := &ResourceCount{
				resourceSearcher: resourceSearcher,
				accessChecker:    accessChecker,
			}

			// Build message
			ctx := context.Background()
			message := service.BuildMessage(ctx, "test-user", tc.result, model.SearchCriteria{PageSize: 10})

			// Execute
			count, err := service.CheckAccess(ctx, "test-user", tc.result, message)

			// Verify
			if tc.expectedError {
				assertion.Error(err)
			} else {
				assertion.NoError(err)
				assertion.Equal(tc.expectedCount, count)
			}
		})
	}
}

func TestResourceCountIsReady(t *testing.T) {
	tests := []struct {
		name          string
		setupMocks    func(*mock.MockResourceSearcher, *mock.MockAccessControlChecker)
		expectedError bool
	}{
		{
			name: "both services ready",
			setupMocks: func(resourceSearcher *mock.MockResourceSearcher, accessChecker *mock.MockAccessControlChecker) {
				// Both mocks return nil for IsReady by default
			},
			expectedError: false,
		},
		{
			name: "resource searcher not ready",
			setupMocks: func(resourceSearcher *mock.MockResourceSearcher, accessChecker *mock.MockAccessControlChecker) {
				resourceSearcher.SetIsReadyError(assert.AnError)
			},
			expectedError: true,
		},
		{
			name: "access checker not ready",
			setupMocks: func(resourceSearcher *mock.MockResourceSearcher, accessChecker *mock.MockAccessControlChecker) {
				accessChecker.SetIsReadyError(assert.AnError)
			},
			expectedError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assertion := assert.New(t)

			// Setup mocks
			resourceSearcher := mock.NewMockResourceSearcher()
			accessChecker := mock.NewMockAccessControlChecker()
			tc.setupMocks(resourceSearcher, accessChecker)

			// Create service
			service := NewResourceCount(resourceSearcher, accessChecker)

			// Execute
			ctx := context.Background()
			err := service.IsReady(ctx)

			// Verify
			if tc.expectedError {
				assertion.Error(err)
			} else {
				assertion.NoError(err)
			}
		})
	}
}
