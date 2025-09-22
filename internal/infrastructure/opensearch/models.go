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
	Hits      `json:"hits"`
	PageToken *string `json:"last_item_id,omitempty"`
}

type CountResponse struct {
	Count int `json:"count"`
}

// AggregationBucket represents a single aggregation bucket.
type AggregationBucket struct {
	Key      string `json:"key"`
	DocCount uint64 `json:"doc_count"`
}

// TermsAggregation represents a terms aggregation response.
type TermsAggregation struct {
	DocCountErrorUpperBound uint64              `json:"doc_count_error_upper_bound"`
	SumOtherDocCount        uint64              `json:"sum_other_doc_count"`
	Buckets                 []AggregationBucket `json:"buckets"`
}

// AggregationResponse represents the aggregations in a search response.
type AggregationResponse struct {
	GroupBy TermsAggregation `json:"group_by"`
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
