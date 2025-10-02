#!/bin/bash

# Script to update research publications from Google Scholar
# Usage: ./update-research.sh [scholar_user_id]

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WEBSITE_ROOT="$(dirname "$SCRIPT_DIR")"

echo "Updating research publications from Google Scholar..."

# Change to scripts directory to run the Go program
cd "$SCRIPT_DIR"

# Run the scraper (pass scholar ID if provided)
if [ "$1" != "" ]; then
    go run scholar-scraper.go "$1"
else
    go run scholar-scraper.go
fi

if [ $? -eq 0 ]; then
    echo "Research page updated successfully!"
    echo "You can now build your Hugo site with: hugo --minify"
else
    echo "Error updating research page"
    exit 1
fi