# {{ .Site.Title }}

> {{ .Site.Params.description | plainify | chomp }}

Personal site and blog of {{ .Site.Params.author.name }}: PhD researcher in mobile networks (6G), researcher at the i2cat Foundation, part-time lecturer at Universitat Pompeu Fabra, and community-networks volunteer.

## Pages

- [About]({{ "about/" | absURL }}): Background, research, work and volunteering
- [Now]({{ "now/" | absURL }}): What I am doing now
- [Research]({{ "research/" | absURL }}): Publications and research activity
- [Projects]({{ "projects/" | absURL }}): Projects I lead or contribute to
- [CV]({{ "cv/" | absURL }}): Curriculum vitae

## Posts
{{ range first 15 (where .Site.RegularPages "Section" "posts") }}
- [{{ .Title }}]({{ .Permalink }}){{ with .Summary }}: {{ . | plainify | chomp | truncate 140 }}{{ end }}
{{- end }}

## Feeds & machine-readable

- [llms-full.txt]({{ "llms-full.txt" | absURL }}): Full plaintext of every page and post
- [RSS]({{ "index.xml" | absURL }}): Full-text RSS feed
- [Sitemap]({{ "sitemap.xml" | absURL }}): XML sitemap

## Optional

- Every page has a Markdown version: append `index.md` to its URL (e.g. {{ "about/index.md" | absURL }})
