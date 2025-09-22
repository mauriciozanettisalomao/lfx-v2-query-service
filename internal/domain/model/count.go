// Copyright The Linux Foundation and each contributor to LFX.
// SPDX-License-Identifier: MIT

package model

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
