// Copyright The Linux Foundation and each contributor to LFX.
// SPDX-License-Identifier: MIT

package opensearch

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
)

type httpClient struct {
	baseURL    string
	httpClient *http.Client
	client     *opensearchapi.Client
}

func (c *httpClient) Search(ctx context.Context, index string, query []byte) (*SearchResponse, error) {

	slog.DebugContext(ctx, "executing opensearch search",
		"index", index,
		"query", string(query),
	)

	searchRequest := opensearchapi.SearchReq{
		Indices: []string{index},
		Body:    bytes.NewReader(query),
		Params: opensearchapi.SearchParams{
			Source: true,
			SourceIncludes: []string{
				"object_ref",
				"object_type",
				"object_id",
				"public",
				"access_check_object",
				"access_check_relation",
				"data",
			},
		},
	}

	searchResponse, errSearchResponse := c.client.Search(ctx, &searchRequest)
	if errSearchResponse != nil {
		slog.ErrorContext(ctx, "failed to execute search", "error", errSearchResponse)
		return nil, fmt.Errorf("failed to execute search: %w", errSearchResponse)
	}

	// Check for errors in the response
	if searchResponse.Errors {
		slog.ErrorContext(ctx, "opensearch search returned errors")
		return nil, fmt.Errorf("opensearch search returned errors")
	}

	result := &SearchResponse{
		Hits: Hits{
			Total: Total{
				Value: searchResponse.Hits.Total.Value,
			},
			Hits: make([]Hit, len(searchResponse.Hits.Hits)),
		},
	}
	for i, hit := range searchResponse.Hits.Hits {
		result.Hits.Hits[i] = Hit{
			ID:     hit.ID,
			Source: hit.Source,
		}
	}
	return result, nil
}
