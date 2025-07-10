// Copyright The Linux Foundation and each contributor to LFX.
// SPDX-License-Identifier: MIT

package opensearch

import "encoding/json"

// Config represents OpenSearch configuration
type Config struct {
	URL   string `json:"url"`
	Index string `json:"index"`
}

// SearchResponse represents the OpenSearch search response
type SearchResponse struct {
	Hits `json:"hits"`
}

// Hits represents the hits in the search response
type Hits struct {
	Total `json:"total"`
	Hits  []Hit `json:"hits"`
}

// Total represents the total number of hits
type Total struct {
	Value int `json:"value"`
}

// Hit represents a single search result hit
type Hit struct {
	ID     string          `json:"_id"`
	Score  float64         `json:"_score"`
	Source json.RawMessage `json:"_source"`
}
