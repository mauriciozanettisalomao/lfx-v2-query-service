// Copyright The Linux Foundation and each contributor to LFX.
// SPDX-License-Identifier: MIT

package clearbit

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/linuxfoundation/lfx-v2-query-service/internal/domain/model"
)

func TestNewOrganizationSearcher(t *testing.T) {
	ctx := context.Background()

	// Local test server to satisfy IsReady without internet access.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	tests := []struct {
		name        string
		config      Config
		expectError bool
	}{
		{
			name: "valid config with fake API key",
			config: Config{
				APIKey:     "test-api-key",
				BaseURL:    ts.URL,
				Timeout:    5 * time.Second,
				MaxRetries: 1,
				RetryDelay: 100 * time.Millisecond,
			},
			expectError: false,
		},
		{
			name: "missing API key",
			config: Config{
				BaseURL:    ts.URL,
				Timeout:    30 * time.Second,
				MaxRetries: 3,
				RetryDelay: 1 * time.Second,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewOrganizationSearcher(ctx, tt.config)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestConvertToDomainModel(t *testing.T) {
	searcher := &OrganizationSearcher{}

	tests := []struct {
		name     string
		company  *ClearbitCompany
		expected *model.Organization
	}{
		{
			name: "complete company data",
			company: &ClearbitCompany{
				Name:   "Test Company",
				Domain: "test.com",
				Category: &ClearbitCategory{
					Industry:    "Technology",
					SubIndustry: "Software",
				},
				Metrics: &ClearbitMetrics{
					EmployeesRange: "100-500",
				},
			},
			expected: &model.Organization{
				Name:      "Test Company",
				Domain:    "test.com",
				Industry:  "Technology",
				Sector:    "Software",
				Employees: "100-500",
			},
		},
		{
			name: "minimal company data",
			company: &ClearbitCompany{
				Name:   "Minimal Company",
				Domain: "minimal.com",
			},
			expected: &model.Organization{
				Name:   "Minimal Company",
				Domain: "minimal.com",
			},
		},
		{
			name: "company with employee count instead of range",
			company: &ClearbitCompany{
				Name:   "Employee Count Company",
				Domain: "employees.com",
				Metrics: &ClearbitMetrics{
					Employees: intPtr(250),
				},
			},
			expected: &model.Organization{
				Name:      "Employee Count Company",
				Domain:    "employees.com",
				Employees: "250",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := searcher.convertToDomainModel(tt.company)

			if result.Name != tt.expected.Name {
				t.Errorf("Expected name '%s', got '%s'", tt.expected.Name, result.Name)
			}
			if result.Domain != tt.expected.Domain {
				t.Errorf("Expected domain '%s', got '%s'", tt.expected.Domain, result.Domain)
			}
			if result.Industry != tt.expected.Industry {
				t.Errorf("Expected industry '%s', got '%s'", tt.expected.Industry, result.Industry)
			}
			if result.Sector != tt.expected.Sector {
				t.Errorf("Expected sector '%s', got '%s'", tt.expected.Sector, result.Sector)
			}
			if result.Employees != tt.expected.Employees {
				t.Errorf("Expected employees '%s', got '%s'", tt.expected.Employees, result.Employees)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.BaseURL != "https://company.clearbit.com" {
		t.Errorf("Expected default BaseURL 'https://company.clearbit.com', got '%s'", config.BaseURL)
	}
	if config.Timeout != 30*time.Second {
		t.Errorf("Expected default timeout 30s, got %v", config.Timeout)
	}
	if config.MaxRetries != 3 {
		t.Errorf("Expected default max retries 3, got %d", config.MaxRetries)
	}
	if config.RetryDelay != 1*time.Second {
		t.Errorf("Expected default retry delay 1s, got %v", config.RetryDelay)
	}
}

// Helper function to create int pointer
func intPtr(i int) *int {
	return &i
}
