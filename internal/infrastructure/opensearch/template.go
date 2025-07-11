// Copyright The Linux Foundation and each contributor to LFX.
// SPDX-License-Identifier: MIT

package opensearch

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
