// Copyright The Linux Foundation and each contributor to LFX.
// SPDX-License-Identifier: MIT

package design

import (
	. "goa.design/goa/v3/dsl"
)

var JWTAuth = JWTSecurity("jwt", func() {
	Description("Heimdall authorization")
})

var _ = Service("query-svc", func() {
	Description("The query service provides resource and user queries.")

	Method("query-resources", func() {
		Description("Locate resources by their type or parent, or use typeahead search to query resources by a display name or similar alias.")

		Security(JWTAuth)

		Payload(func() {
			Extend(Sortable)
			Token("bearer_token", String, func() {
				Description("JWT token issued by Heimdall")
				Example("eyJhbGci...")
			})
			Attribute("version", String, "Version of the API", func() {
				Enum("1")
				Example("1")
			})
			Attribute("name", String, "Resource name or alias; supports typeahead", func() {
				Example("gov board")
				MinLength(1)
			})
			Attribute("parent", String, "Parent (for navigation; varies by object type)", func() {
				Example("project:123")
			})
			Attribute("type", String, "Resource type to search", func() {
				Example("committee")
			})
			Attribute("tags", ArrayOf(String), "Tags to search (varies by object type)", func() {
				Example([]string{"active"})
			})
			Required("bearer_token", "version")
		})

		Result(func() {
			Attribute("resources", ArrayOf(Resource), "Resources found", func() {})
			Attribute("page_token", String, "Opaque token if more results are available", func() {
				Example("****")
			})
			Attribute("cache_control", String, "Cache control header", func() {
				Example("public, max-age=300")
			})
			Required("resources")
		})

		Error("BadRequest", ErrorResult, "Bad request")

		HTTP(func() {
			GET("/query/resources")
			Param("version:v")
			Param("name")
			Param("parent")
			Param("type")
			Param("tags")
			Param("sort")
			Param("page_token")
			Header("bearer_token:Authorization")
			Response(StatusOK, func() {
				Header("cache_control:Cache-Control")
			})
			Response("BadRequest", StatusBadRequest)
		})
	})

	Method("readyz", func() {
		Description("Check if the service is able to take inbound requests.")
		Result(Bytes, func() {
			Example("OK")
		})
		Error("NotReady", func() {
			Description("Service is not ready yet")
			Temporary()
			Fault()
		})
		HTTP(func() {
			GET("/readyz")
			Response(StatusOK, func() {
				ContentType("text/plain")
			})
			Response("NotReady", StatusServiceUnavailable)
		})
	})

	Method("livez", func() {
		Description("Check if the service is alive.")
		Result(Bytes, func() {
			Example("OK")
		})
		HTTP(func() {
			GET("/livez")
			Response(StatusOK, func() {
				ContentType("text/plain")
			})
		})
	})

	// Serve the file gen/http/openapi3.json for requests sent to /openapi.json.
	Files("/openapi.json", "gen/http/openapi3.json")
})
