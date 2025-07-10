// Copyright The Linux Foundation and each contributor to LFX.
// SPDX-License-Identifier: MIT

package opensearch

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"text/template"

	"github.com/Masterminds/sprig/v3"
)

const queryResourceSource = `{
  "size": {{ .PageSize }},
  "query": {
    "bool": {
      "must": [
        {
          "term": {"latest": true}
        }
        {{- if .PublicOnly }},
        {
          "term": {"public": true}
        }
        {{- end }}
        {{- if .ResourceType }},
        {
          "term": {
            "object_type": {{ .ResourceType | quote }}
          }
        }
        {{- end }}
        {{- if .ParentRef }},
        {
          "term": {
            "parent_refs": {{ .ParentRef | quote }}
          }
        }
        {{- end }}
        {{- if .Name }},
        {
          "multi_match": {
            "query": {{ .Name | quote }},
            "type": "bool_prefix",
            "fields": [
              "name_and_aliases",
              "name_and_aliases._2gram",
              "name_and_aliases._3gram"
            ]
          }
        }
        {{- end }}
      ]
      {{- if .Tags }},
      "minimum_should_match": 1,
      "should": [
        {{- $first := true -}}
        {{- range .Tags -}}
        {{- if $first -}}
        {{- $first = false -}}
        {{- else }},
        {{- end }}
        {
          "term": {
            "tags": {{ . | quote }}
          }
        }
        {{- end }}
      ]
      {{- end }}
    }
  }
  {{- if .SearchAfter }},
  "search_after": {{ .SearchAfter }}
  {{- end }},
  "sort": [
    {
      {{ .SortBy | quote }}: {
        "order": {{ .SortOrder | quote }}
      }
    },
    {"_id": "asc"}
  ]
}`

var queryResourceTemplate = template.Must(
	template.New("queryResource").
		Funcs(sprig.FuncMap()).
		Parse(queryResourceSource))

// TemplateData represents the data structure for template rendering
type TemplateData struct {
	Tags         []string
	Name         string
	ResourceType string
	ParentRef    string
	SortBy       string
	SortOrder    string
	SearchAfter  string
	PageSize     int
	PublicOnly   bool
}

func (t *TemplateData) Render(ctx context.Context) ([]byte, error) {
	var buf bytes.Buffer
	if err := queryResourceTemplate.Execute(&buf, t); err != nil {
		slog.ErrorContext(ctx, "failed to render query template", "error", err)
		return nil, err
	}
	query := json.RawMessage(buf.Bytes())
	parsed, err := json.Marshal(query)
	if err != nil {
		slog.ErrorContext(ctx, "failed to marshal rendered query", "error", err)
		return nil, err
	}
	return parsed, nil
}
