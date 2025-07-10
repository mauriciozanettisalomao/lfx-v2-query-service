// Copyright The Linux Foundation and each contributor to LFX.
// SPDX-License-Identifier: MIT

package opensearch

import (
	"bytes"
	"fmt"
	"text/template"
)

// SearchTemplates contains all OpenSearch query templates
type SearchTemplates struct {
	resourceSearchTemplate *template.Template
	typeaheadTemplate      *template.Template
}

// NewSearchTemplates creates a new instance of SearchTemplates
func NewSearchTemplates() (*SearchTemplates, error) {
	resourceSearchTemplate, err := template.New("resourceSearch").Parse(resourceSearchQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to parse resource search template: %w", err)
	}

	typeaheadTemplate, err := template.New("typeahead").Parse(typeaheadQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to parse typeahead template: %w", err)
	}

	return &SearchTemplates{
		resourceSearchTemplate: resourceSearchTemplate,
		typeaheadTemplate:      typeaheadTemplate,
	}, nil
}

// TemplateData represents the data structure for template rendering
type TemplateData struct {
	Name      string
	Type      string
	Parent    string
	Tags      []string
	Sort      string
	Size      int
	From      int
	PageToken string
}

// RenderResourceSearchQuery renders the resource search query template
func (st *SearchTemplates) RenderResourceSearchQuery(data TemplateData) (string, error) {
	var buf bytes.Buffer
	if err := st.resourceSearchTemplate.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to render resource search query: %w", err)
	}
	return buf.String(), nil
}

// RenderTypeaheadQuery renders the typeahead search query template
func (st *SearchTemplates) RenderTypeaheadQuery(data TemplateData) (string, error) {
	var buf bytes.Buffer
	if err := st.typeaheadTemplate.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to render typeahead query: %w", err)
	}
	return buf.String(), nil
}

// resourceSearchQuery is the main search query template
const resourceSearchQuery = `{
  "query": {
    "bool": {
      "must": [
        {{if .Name}}
        {
          "multi_match": {
            "query": "{{.Name}}",
            "fields": ["name^3", "description^2", "tags"],
            "type": "phrase_prefix",
            "fuzziness": "AUTO"
          }
        },
        {{end}}
        {{if .Type}}
        {
          "term": {
            "type": "{{.Type}}"
          }
        },
        {{end}}
        {{if .Parent}}
        {
          "term": {
            "parent": "{{.Parent}}"
          }
        },
        {{end}}
        {{if .Tags}}
        {
          "terms": {
            "tags": [{{range $i, $tag := .Tags}}{{if $i}}, {{end}}"{{$tag}}"{{end}}]
          }
        },
        {{end}}
        {
          "range": {
            "updated_at": {
              "gte": "now-1y"
            }
          }
        }
      ],
      "must_not": [
        {
          "term": {
            "status": "deleted"
          }
        }
      ]
    }
  },
  "sort": [
    {{if eq .Sort "name_asc"}}
    {
      "name.keyword": {
        "order": "asc"
      }
    }
    {{else if eq .Sort "name_desc"}}
    {
      "name.keyword": {
        "order": "desc"
      }
    }
    {{else if eq .Sort "updated_asc"}}
    {
      "updated_at": {
        "order": "asc"
      }
    }
    {{else if eq .Sort "updated_desc"}}
    {
      "updated_at": {
        "order": "desc"
      }
    }
    {{else}}
    {
      "name.keyword": {
        "order": "asc"
      }
    }
    {{end}},
    {
      "_score": {
        "order": "desc"
      }
    }
  ],
  "size": {{.Size}},
  "from": {{.From}},
  "_source": {
    "includes": ["type", "id", "name", "description", "tags", "updated_at", "data"]
  },
  "highlight": {
    "fields": {
      "name": {},
      "description": {}
    }
  }
}`

// typeaheadQuery is the typeahead search query template
const typeaheadQuery = `{
  "query": {
    "bool": {
      "must": [
        {
          "multi_match": {
            "query": "{{.Name}}",
            "fields": ["name^3", "name.ngram^2"],
            "type": "phrase_prefix"
          }
        },
        {{if .Type}}
        {
          "term": {
            "type": "{{.Type}}"
          }
        },
        {{end}}
        {
          "term": {
            "status": "active"
          }
        }
      ]
    }
  },
  "sort": [
    {
      "_score": {
        "order": "desc"
      }
    },
    {
      "name.keyword": {
        "order": "asc"
      }
    }
  ],
  "size": {{.Size}},
  "_source": {
    "includes": ["type", "id", "name", "description"]
  },
  "highlight": {
    "fields": {
      "name": {}
    }
  }
}`
