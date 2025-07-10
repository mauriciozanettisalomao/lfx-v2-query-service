// Copyright The Linux Foundation and each contributor to LFX.
// SPDX-License-Identifier: MIT

package design

import (
	. "goa.design/goa/v3/dsl"
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

var Sortable = Type("Sortable", func() {
	Attribute("sort", String, "Sort order for results", func() {
		Enum(SortValues...)
		Default("name_asc")
		Example("updated_desc")
	})
	Attribute("page_token", String, "Opaque token for pagination", func() {
		Example("****")
	})
})

/*
var Contact = Type("Contact", func() {
	Description("A contact is a per-resource representation of a user or bot that is associated with that resource.")

	Attribute("parent_refs", ArrayOf(string), "LFX object references this profile was found on", func() {
		Example([]string{"committee:123", "meeting:456"})
	})
	Attribute("lfx_principal", String, "LFX principal (username)", func() {
		Example("jane_doe")
	})
	Attribute("name", String, "Contact full name", func() {
		Example("Jane Doe")
	})
	Attribute("emails", ArrayOf(String), "Contact email addresses", func() {})
	Attribute("bot", Boolean, "Contact is a bot", func() {})
	Attribute("profile", Any, "Contact profile data", func() {})
})
*/

var Resource = Type("Resource", func() {
	Description("A resource is a universal representation of an LFX API resource for indexing.")

	Attribute("type", String, "Resource type", func() {
		Example("committee")
	})
	Attribute("id", String, "Resource ID (within its resource collection)", func() {
		Example("123")
	})
	Attribute("data", Any, "Resource data snapshot", func() {
		Example(CommitteeExampleStub{
			ID:          "123",
			Name:        "My committee",
			Description: "a committee",
		})
	})
})

// Define an example cached LFX resource for the nested "data" attribute for
// resource searches. This example happens to be a committee to match the
// example value of "committee" for the "type" attribute of Resource.
type CommitteeExampleStub struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}
