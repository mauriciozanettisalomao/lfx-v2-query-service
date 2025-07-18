// Copyright The Linux Foundation and each contributor to LFX.
// SPDX-License-Identifier: MIT

package design

import (
	"goa.design/goa/v3/dsl"
)

var SortValues = []any{
	"name_asc",
	"name_desc",
	// Note, "created_at" sorting is not currently possible, because we only
	// include it on the "created" trasaction, to better distinguish it from an
	// "updated" transaction. Adding it would slow down the indexing service
	// (perhaps it could be asyncronously added by the janitor?) since
	// propogating attributes from earlier revisions is not currently supported.
	"updated_asc",
	"updated_desc",
}

var Sortable = dsl.Type("Sortable", func() {
	dsl.Attribute("sort", dsl.String, "Sort order for results", func() {
		dsl.Enum(SortValues...)
		dsl.Default("name_asc")
		dsl.Example("updated_desc")
	})
	dsl.Attribute("page_token", dsl.String, "Opaque token for pagination", func() {
		dsl.Example("****")
	})
})

var Resource = dsl.Type("Resource", func() {
	dsl.Description("A resource is a universal representation of an LFX API resource for indexing.")

	dsl.Attribute("type", dsl.String, "Resource type", func() {
		dsl.Example("committee")
	})
	dsl.Attribute("id", dsl.String, "Resource ID (within its resource collection)", func() {
		dsl.Example("123")
	})
	dsl.Attribute("data", dsl.Any, "Resource data snapshot", func() {
		dsl.Example(CommitteeExampleStub{
			ID:          "123",
			Name:        "My committee",
			Description: "a committee",
		})
	})
})

// BadRequestError is the DSL type for a bad request error.
var BadRequestError = dsl.Type("BadRequestError", func() {
	dsl.Attribute("message", dsl.String, "Error message", func() {
		dsl.Example("The request was invalid.")
	})
	dsl.Required("message")
})

// InternalServerError is the DSL type for an internal server error.
var InternalServerError = dsl.Type("InternalServerError", func() {
	dsl.Attribute("message", dsl.String, "Error message", func() {
		dsl.Example("An internal server error occurred.")
	})
	dsl.Required("message")
})

// ServiceUnavailableError is the DSL type for a service unavailable error.
var ServiceUnavailableError = dsl.Type("ServiceUnavailableError", func() {
	dsl.Attribute("message", dsl.String, "Error message", func() {
		dsl.Example("The service is unavailable.")
	})
	dsl.Required("message")
})

// Define an example cached LFX resource for the nested "data" attribute for
// resource searches. This example happens to be a committee to match the
// example value of "committee" for the "type" attribute of Resource.
type CommitteeExampleStub struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}
