# sergiogimenez.com
Personal website / blog based on Hugo

## Agent-readiness

The site exposes machine-readable interfaces for AI agents/crawlers:

- `/llms.txt` and `/llms-full.txt` ([llmstxt.org](https://llmstxt.org) standard)
- Markdown version of every page: append `index.md` to its URL
- Full-text RSS (`/index.xml`), JSON-LD `Person`/`Article` schema, AI-crawler-friendly `robots.txt`

Audit: `scripts/agent-readiness.sh [base_url]` (default `http://localhost:1313`) — exits non-zero if any check fails.
