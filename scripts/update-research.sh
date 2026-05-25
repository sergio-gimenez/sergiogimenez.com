#!/bin/bash

# Update research publications from Google Scholar
#
# Live mode (may get rate-limited):
#   ./update-research.sh
#   ./update-research.sh --id YOUR_SCHOLAR_ID
#
# Offline mode (recommended):
#   1. Visit https://scholar.google.com/citations?user=YOUR_ID&view_op=list_works&sortby=pubdate
#   2. Save page as HTML
#   3. ./update-research.sh --file saved_page.html
#
# Flags are passed through to the Go scraper.
#   --id     Google Scholar user ID
#   --file   Path to saved Google Scholar HTML file (offline mode)
#   --output Output markdown path (default: ../content/research.md)

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WEBSITE_ROOT="$(dirname "$SCRIPT_DIR")"

echo "Updating research publications from Google Scholar..."
cd "$SCRIPT_DIR"

go run scholar-scraper.go --output "$WEBSITE_ROOT/content/research.md" "$@"

if [ $? -eq 0 ]; then
    echo "Research page updated successfully!"
    echo "Build Hugo site: hugo --minify"
else
    echo "Error updating research page"
    exit 1
fi
