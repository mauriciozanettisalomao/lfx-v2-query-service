// Copyright The Linux Foundation and each contributor to LFX.
// SPDX-License-Identifier: MIT

package clearbit

// ClearbitCompany represents the company data structure returned by Clearbit API
type ClearbitCompany struct {
	// Name is the company name
	Name string `json:"name"`

	// LegalName is the legal name of the company
	LegalName string `json:"legalName"`

	// Domain is the company's primary domain
	Domain string `json:"domain"`

	// Site contains the company's website information
	Site *ClearbitSite `json:"site,omitempty"`

	// Category contains industry classification
	Category *ClearbitCategory `json:"category,omitempty"`

	// Metrics contains company size and other metrics
	Metrics *ClearbitMetrics `json:"metrics,omitempty"`

	// Description is a description of the company
	Description string `json:"description"`
}

// ClearbitSite contains website information
type ClearbitSite struct {
	URL string `json:"url"`
}

// ClearbitCategory contains industry classification
type ClearbitCategory struct {
	Sector        string `json:"sector"`
	IndustryGroup string `json:"industryGroup"`
	Industry      string `json:"industry"`
	SubIndustry   string `json:"subIndustry"`
}

// ClearbitMetrics contains company metrics
type ClearbitMetrics struct {
	Employees      *int   `json:"employees,omitempty"`
	EmployeesRange string `json:"employeesRange"`
}

// ClearbitErrorResponse represents an error response from Clearbit API
type ClearbitErrorResponse struct {
	Error *ClearbitError `json:"error,omitempty"`
}

// ClearbitError represents error details
type ClearbitError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// ClearbitCompanySuggestion represents a company suggestion from the autocomplete API
type ClearbitCompanySuggestion struct {
	// Name is the company name
	Name string `json:"name"`

	// Domain is the company's primary domain
	Domain string `json:"domain"`

	// Logo is the URL to the company's logo (can be null)
	Logo *string `json:"logo"`
}
