#!/usr/bin/env bash
# agent-readiness.sh — score how ready the site is for AI agents/crawlers.
# Usage: scripts/agent-readiness.sh [base_url]   (default: http://localhost:1313)
set -u
BASE="${1:-http://localhost:1313}"
BASE="${BASE%/}"
PASS=0; FAIL=0

ok()   { PASS=$((PASS+1)); printf "  PASS  %s\n" "$1"; }
bad()  { FAIL=$((FAIL+1)); printf "  FAIL  %s\n" "$1"; }
hdr()  { printf "\n[%s]\n" "$1"; }

body() { curl -sfL --max-time 10 "$1" 2>/dev/null; }
code() { curl -sL -o /dev/null -w '%{http_code}' --max-time 10 "$1" 2>/dev/null; }

contains() { # contains <url> <pattern> <label>
  if body "$1" | grep -qi -- "$2"; then ok "$3"; else bad "$3 ($1 ~ $2)"; fi
}
status_is() { # status_is <url> <label>
  local c; c=$(code "$1")
  if [ "$c" = "200" ]; then ok "$2"; else bad "$2 ($1 -> $c)"; fi
}

FIRST_POST=$(body "$BASE/sitemap.xml" | grep -o '<loc>[^<]*/posts/[^<]\+</loc>' | head -1 | sed -e 's/<[^>]*>//g')

hdr "1. Crawler access (robots.txt)"
R=$(body "$BASE/robots.txt")
[ -n "$R" ] && ok "robots.txt exists" || bad "robots.txt missing"
for bot in GPTBot ClaudeBot PerplexityBot OAI-SearchBot Google-Extended; do
  echo "$R" | grep -q "User-agent: $bot" && ok "explicit rule: $bot" || bad "no rule for $bot"
done
echo "$R" | grep -qi "Sitemap:" && ok "sitemap declared" || bad "no sitemap line"

hdr "2. llms.txt standard"
contains "$BASE/llms.txt" '^# ' "llms.txt has H1 title"
contains "$BASE/llms.txt" '^> ' "llms.txt has blockquote summary"
contains "$BASE/llms.txt" '^- \[' "llms.txt has link list"
L=$(body "$BASE/llms-full.txt" | wc -c)
[ "$L" -gt 5000 ] && ok "llms-full.txt substantial (${L} bytes)" || bad "llms-full.txt too small (${L} bytes)"

hdr "3. Machine-readable content"
status_is "$BASE/index.xml" "RSS feed exists"
status_is "$BASE/index.json" "JSON index exists"
contains "$BASE/index.xml" '<content:encoded' "RSS is full-text"
if [ -n "$FIRST_POST" ]; then
  status_is "${FIRST_POST}index.md" "post markdown output ($FIRST_POST)"
  contains "$FIRST_POST" 'rel="alternate" type="text/markdown"' "HTML links to markdown alternate"
else
  bad "no post found in sitemap"
fi

hdr "4. Structured data (JSON-LD)"
H=$(body "$BASE/")
echo "$H" | grep -q 'application/ld+json' && ok "home has JSON-LD" || bad "home has no JSON-LD"
echo "$H" | grep -q '"@type": *"Person"\|"@type":"Person"' && ok "Person entity on home" || bad "no Person JSON-LD on home"
echo "$H" | grep -q 'sameAs' && ok "Person has sameAs links" || bad "Person missing sameAs"
[ -n "$FIRST_POST" ] && contains "$FIRST_POST" '"@type": *"Article"\|"@type":"Article"' "Article JSON-LD on posts"

hdr "5. Crawl hygiene"
status_is "$BASE/sitemap.xml" "sitemap.xml exists"
echo "$H" | grep -q 'rel="canonical"' && ok "canonical link on home" || bad "no canonical on home"
H1=$(echo "$H" | grep -o '<h1' | wc -l)
[ "$H1" -eq 1 ] && ok "exactly one h1 on home" || bad "home has $H1 h1 elements"

hdr "6. Entity clarity"
status_is "$BASE/about/" "about page exists"
echo "$H" | grep -q 'rel="me"' && ok "rel=me links present" || bad "no rel=me links"
status_is "$BASE/cv/" "cv page exists"

printf "\n== Score: %d pass / %d fail ==\n" "$PASS" "$FAIL"
[ "$FAIL" -eq 0 ]
