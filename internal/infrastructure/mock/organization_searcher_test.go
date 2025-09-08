// Copyright The Linux Foundation and each contributor to LFX.
// SPDX-License-Identifier: MIT

package mock

import (
	"context"
	"testing"

	"github.com/linuxfoundation/lfx-v2-query-service/internal/domain/model"
)

func TestMockOrganizationSearcher_QueryOrganizations_ByName(t *testing.T) {
	searcher := NewMockOrganizationSearcher()
	ctx := context.Background()

	tests := []struct {
		name         string
		searchName   string
		expectedName string
		shouldFind   bool
	}{
		{
			name:         "Find Linux Foundation by exact name",
			searchName:   "The Linux Foundation",
			expectedName: "The Linux Foundation",
			shouldFind:   true,
		},
		{
			name:         "Find Linux Foundation by case insensitive name",
			searchName:   "the linux foundation",
			expectedName: "The Linux Foundation",
			shouldFind:   true,
		},
		{
			name:         "Find Zyx-42 Quantum Widgets LLC by exact name",
			searchName:   "Zyx-42 Quantum Widgets LLC",
			expectedName: "Zyx-42 Quantum Widgets LLC",
			shouldFind:   true,
		},
		{
			name:         "Find Blorbtech by case insensitive name",
			searchName:   "blorbtech intergalactic solutions",
			expectedName: "Blorbtech Intergalactic Solutions",
			shouldFind:   true,
		},
		{
			name:       "Not found - non-existent organization",
			searchName: "Non-existent Organization",
			shouldFind: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			criteria := model.OrganizationSearchCriteria{
				Name: &tt.searchName,
			}

			result, err := searcher.QueryOrganizations(ctx, criteria)

			if tt.shouldFind {
				if err != nil {
					t.Errorf("Expected to find organization, but got error: %v", err)
					return
				}
				if result == nil {
					t.Error("Expected to find organization, but got nil result")
					return
				}
				if result.Name != tt.expectedName {
					t.Errorf("Expected organization name '%s', got '%s'", tt.expectedName, result.Name)
				}
			} else {
				if err == nil {
					t.Error("Expected error for non-existent organization, but got nil")
				}
				if result != nil {
					t.Error("Expected nil result for non-existent organization")
				}
			}
		})
	}
}

func TestMockOrganizationSearcher_QueryOrganizations_ByDomain(t *testing.T) {
	searcher := NewMockOrganizationSearcher()
	ctx := context.Background()

	tests := []struct {
		name           string
		searchDomain   string
		expectedName   string
		expectedDomain string
		shouldFind     bool
	}{
		{
			name:           "Find Linux Foundation by domain",
			searchDomain:   "linuxfoundation.org",
			expectedName:   "The Linux Foundation",
			expectedDomain: "linuxfoundation.org",
			shouldFind:     true,
		},
		{
			name:           "Find Zyx-42 by case insensitive domain",
			searchDomain:   "ZYX42-QUANTUM-WIDGETS.FAKE",
			expectedName:   "Zyx-42 Quantum Widgets LLC",
			expectedDomain: "zyx42-quantum-widgets.fake",
			shouldFind:     true,
		},
		{
			name:           "Find Whizbang by domain",
			searchDomain:   "whizbang-doodads.fake",
			expectedName:   "Whizbang Doodad Corporation",
			expectedDomain: "whizbang-doodads.fake",
			shouldFind:     true,
		},
		{
			name:         "Not found - non-existent domain",
			searchDomain: "nonexistent.com",
			shouldFind:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			criteria := model.OrganizationSearchCriteria{
				Domain: &tt.searchDomain,
			}

			result, err := searcher.QueryOrganizations(ctx, criteria)

			if tt.shouldFind {
				if err != nil {
					t.Errorf("Expected to find organization, but got error: %v", err)
					return
				}
				if result == nil {
					t.Error("Expected to find organization, but got nil result")
					return
				}
				if result.Name != tt.expectedName {
					t.Errorf("Expected organization name '%s', got '%s'", tt.expectedName, result.Name)
				}
				if result.Domain != tt.expectedDomain {
					t.Errorf("Expected organization domain '%s', got '%s'", tt.expectedDomain, result.Domain)
				}
			} else {
				if err == nil {
					t.Error("Expected error for non-existent organization, but got nil")
				}
				if result != nil {
					t.Error("Expected nil result for non-existent organization")
				}
			}
		})
	}
}

func TestMockOrganizationSearcher_QueryOrganizations_NoCriteria(t *testing.T) {
	searcher := NewMockOrganizationSearcher()
	ctx := context.Background()

	criteria := model.OrganizationSearchCriteria{}

	result, err := searcher.QueryOrganizations(ctx, criteria)

	if err == nil {
		t.Error("Expected error for no search criteria, but got nil")
	}
	if result != nil {
		t.Error("Expected nil result for no search criteria")
	}
}

func TestMockOrganizationSearcher_HelperMethods(t *testing.T) {
	searcher := NewMockOrganizationSearcher()

	// Test GetOrganizationCount
	initialCount := searcher.GetOrganizationCount()
	if initialCount == 0 {
		t.Error("Expected initial organization count to be greater than 0")
	}

	// Test GetOrganizationByName
	org := searcher.GetOrganizationByName("The Linux Foundation")
	if org == nil {
		t.Error("Expected to find Linux Foundation")
	} else if org.Domain != "linuxfoundation.org" {
		t.Errorf("Expected domain 'linuxfoundation.org', got '%s'", org.Domain)
	}

	// Test GetOrganizationByDomain
	org = searcher.GetOrganizationByDomain("zyx42-quantum-widgets.fake")
	if org == nil {
		t.Error("Expected to find Zyx-42 Quantum Widgets LLC")
	} else if org.Name != "Zyx-42 Quantum Widgets LLC" {
		t.Errorf("Expected name 'Zyx-42 Quantum Widgets LLC', got '%s'", org.Name)
	}

	// Test AddOrganization
	newOrg := model.Organization{
		Name:      "Test Organization",
		Domain:    "test.org",
		Industry:  "Testing",
		Sector:    "Quality Assurance",
		Employees: "1-10",
	}
	searcher.AddOrganization(newOrg)

	newCount := searcher.GetOrganizationCount()
	if newCount != initialCount+1 {
		t.Errorf("Expected count to increase by 1, got %d -> %d", initialCount, newCount)
	}

	// Verify the new organization can be found
	foundOrg := searcher.GetOrganizationByName("Test Organization")
	if foundOrg == nil {
		t.Error("Expected to find newly added organization")
	}

	// Test GetAllOrganizations
	allOrgs := searcher.GetAllOrganizations()
	if len(allOrgs) != newCount {
		t.Errorf("Expected GetAllOrganizations to return %d organizations, got %d", newCount, len(allOrgs))
	}

	// Test ClearOrganizations
	searcher.ClearOrganizations()
	if searcher.GetOrganizationCount() != 0 {
		t.Error("Expected organization count to be 0 after clearing")
	}
}

func TestMockOrganizationSearcher_IsReady(t *testing.T) {
	searcher := NewMockOrganizationSearcher()
	ctx := context.Background()

	err := searcher.IsReady(ctx)
	if err != nil {
		t.Errorf("Expected mock searcher to always be ready, got error: %v", err)
	}
}

func TestMockOrganizationSearcher_FictionalCompanies(t *testing.T) {
	searcher := NewMockOrganizationSearcher()
	ctx := context.Background()

	// Test searching for various fictional companies
	fictionalCompanies := []struct {
		name   string
		domain string
	}{
		{"Zyx-42 Quantum Widgets LLC", "zyx42-quantum-widgets.fake"},
		{"Blorbtech Intergalactic Solutions", "blorbtech-solutions.notreal"},
		{"Fizzlebottom & Associates Pty", "fizzlebottom-associates.example"},
		{"Whizbang Doodad Corporation", "whizbang-doodads.fake"},
		{"Sproinkel Digital Dynamics", "sproinkel-digital.test"},
		{"Flibber-Jib Environmental Corp", "flibber-jib-env.localhost"},
		{"Quibblesnort Cybersecurity Ltd", "quibblesnort-cyber.mock"},
	}

	for _, company := range fictionalCompanies {
		t.Run("Find "+company.name+" by name", func(t *testing.T) {
			criteria := model.OrganizationSearchCriteria{
				Name: &company.name,
			}

			result, err := searcher.QueryOrganizations(ctx, criteria)
			if err != nil {
				t.Errorf("Expected to find %s, but got error: %v", company.name, err)
				return
			}
			if result == nil {
				t.Errorf("Expected to find %s, but got nil result", company.name)
				return
			}
			if result.Name != company.name {
				t.Errorf("Expected name '%s', got '%s'", company.name, result.Name)
			}
			if result.Domain != company.domain {
				t.Errorf("Expected domain '%s', got '%s'", company.domain, result.Domain)
			}
		})

		t.Run("Find "+company.name+" by domain", func(t *testing.T) {
			criteria := model.OrganizationSearchCriteria{
				Domain: &company.domain,
			}

			result, err := searcher.QueryOrganizations(ctx, criteria)
			if err != nil {
				t.Errorf("Expected to find %s by domain, but got error: %v", company.name, err)
				return
			}
			if result == nil {
				t.Errorf("Expected to find %s by domain, but got nil result", company.name)
				return
			}
			if result.Name != company.name {
				t.Errorf("Expected name '%s', got '%s'", company.name, result.Name)
			}
		})
	}
}
