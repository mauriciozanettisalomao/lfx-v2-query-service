// Copyright The Linux Foundation and each contributor to LFX.
// SPDX-License-Identifier: MIT

package model

// Organization represents an organization entity
type Organization struct {
	// Organization name
	Name string `json:"name"`
	// Organization domain
	Domain string `json:"domain"`
	// Organization industry classification
	Industry string `json:"industry"`
	// Business sector classification
	Sector string `json:"sector"`
	// Employee count or range
	Employees string `json:"employees"`
}

// OrganizationSuggestion represents a suggested organization for typeahead search
type OrganizationSuggestion struct {
	// Organization name
	Name string `json:"name"`
	// Organization domain
	Domain string `json:"domain"`
	// Organization logo URL
	Logo *string `json:"logo,omitempty"`
}

// OrganizationSuggestionsResult contains the results of an organization suggestions search
type OrganizationSuggestionsResult struct {
	// Suggestions found
	Suggestions []OrganizationSuggestion `json:"suggestions"`
}
