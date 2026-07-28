# sergiogimenez.com
Personal website / blog based on Hugo

## Agent-readiness

The site exposes machine-readable interfaces for AI agents/crawlers:

- `/llms.txt` and `/llms-full.txt` ([llmstxt.org](https://llmstxt.org) standard)
- Markdown version of every page: append `index.md` to its URL, or send `Accept: text/markdown` (content negotiation via Cloudflare Transform Rule)
- Full-text RSS (`/index.xml`), JSON-LD `Person`/`Article` schema, AI-crawler-friendly `robots.txt` with Content Signals
- `Link` HTTP header on the homepage pointing to `llms.txt`/`llms-full.txt`/sitemap (Cloudflare Transform Rule)

Note: content negotiation and the `Link` header are Cloudflare Transform Rules
(zone `sergiogimenez.com`, dashboard → Rules → Transform Rules), not in this repo.

Audit: `scripts/agent-readiness.sh [base_url]` (default `http://localhost:1313`) — exits non-zero if any check fails.
