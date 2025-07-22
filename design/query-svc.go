// Copyright The Linux Foundation and each contributor to LFX.
// SPDX-License-Identifier: MIT

package design

import (
	"goa.design/goa/v3/dsl"
)

var JWTAuth = dsl.JWTSecurity("jwt", func() {
	dsl.Description("Heimdall authorization")
})

var _ = dsl.Service("query-svc", func() {
	dsl.Description("The query service provides resource and user queries.")

	dsl.Error("BadRequest", BadRequestError, "Bad request")
	dsl.Error("InternalServerError", InternalServerError, "Internal server error")
	dsl.Error("ServiceUnavailable", ServiceUnavailableError, "Service unavailable")

	dsl.Method("query-resources", func() {
		dsl.Description("Locate resources by their type or parent, or use typeahead search to query resources by a display name or similar alias.")

		dsl.Security(JWTAuth)

		dsl.Payload(func() {
			dsl.Extend(Sortable)
			dsl.Token("bearer_token", dsl.String, func() {
				dsl.Description("JWT token issued by Heimdall")
				dsl.Example("eyJhbGci...")
			})
			dsl.Attribute("version", dsl.String, "Version of the API", func() {
				dsl.Enum("1")
				dsl.Example("1")
			})
			dsl.Attribute("name", dsl.String, "Resource name or alias; supports typeahead", func() {
				dsl.Example("gov board")
				dsl.MinLength(1)
			})
			dsl.Attribute("parent", dsl.String, "Parent (for navigation; varies by object type)", func() {
				dsl.Example("project:123")
				dsl.Pattern(`^[a-zA-Z]+:[a-zA-Z0-9_-]+$`)
			})
			dsl.Attribute("type", dsl.String, "Resource type to search", func() {
				dsl.Example("committee")
			})
			dsl.Attribute("tags", dsl.ArrayOf(dsl.String), "Tags to search (varies by object type)", func() {
				dsl.Example([]string{"active"})
			})
			dsl.Required("bearer_token", "version")
		})

		dsl.Result(func() {
			dsl.Attribute("resources", dsl.ArrayOf(Resource), "Resources found", func() {})
			dsl.Attribute("page_token", dsl.String, "Opaque token if more results are available", func() {
				dsl.Example("****")
			})
			dsl.Attribute("cache_control", dsl.String, "Cache control header", func() {
				dsl.Example("public, max-age=300")
			})
			dsl.Required("resources")
		})

		dsl.HTTP(func() {
			dsl.GET("/query/resources")
			dsl.Param("version:v")
			dsl.Param("name")
			dsl.Param("parent")
			dsl.Param("type")
			dsl.Param("tags")
			dsl.Param("sort")
			dsl.Param("page_token")
			dsl.Header("bearer_token:Authorization")
			dsl.Response(dsl.StatusOK, func() {
				dsl.Header("cache_control:Cache-Control")
			})
			dsl.Response("BadRequest", dsl.StatusBadRequest)
			dsl.Response("InternalServerError", dsl.StatusInternalServerError)
			dsl.Response("ServiceUnavailable", dsl.StatusServiceUnavailable)
		})
	})

	dsl.Method("readyz", func() {
		dsl.Description("Check if the service is able to take inbound requests.")
		dsl.Result(dsl.Bytes, func() {
			dsl.Example("OK")
		})
		dsl.Error("NotReady", func() {
			dsl.Description("Service is not ready yet")
			dsl.Temporary()
			dsl.Fault()
		})
		dsl.HTTP(func() {
			dsl.GET("/readyz")
			dsl.Response(dsl.StatusOK, func() {
				dsl.ContentType("text/plain")
			})
			dsl.Response("NotReady", dsl.StatusServiceUnavailable)
		})
	})

	dsl.Method("livez", func() {
		dsl.Description("Check if the service is alive.")
		dsl.Result(dsl.Bytes, func() {
			dsl.Example("OK")
		})
		dsl.HTTP(func() {
			dsl.GET("/livez")
			dsl.Response(dsl.StatusOK, func() {
				dsl.ContentType("text/plain")
			})
		})
	})

	// Serve the file gen/http/openapi3.json for requests sent to /openapi.json.
	dsl.Files("/openapi.json", "gen/http/openapi3.json")
})
