---
url: {{ .Permalink }}
{{- if not .Date.IsZero }}
date: {{ .Date.Format "2006-01-02" }}
{{- end }}
{{- with .Params.tags }}
tags: [{{ delimit . ", " }}]
{{- end }}
---

# {{ .Title }}

{{ .RawContent }}
