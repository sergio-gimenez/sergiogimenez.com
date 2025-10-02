# Research Publication Scripts

This directory contains scripts to automatically update the research publications page by scraping Google Scholar.

## Files

- `scholar-scraper.go` - Main Go script that scrapes Google Scholar and generates the research.md file
- `update-research.sh` - Shell script wrapper to run the scraper
- `go.mod` / `go.sum` - Go module dependencies

## Usage

### Manual Update

To manually update the research publications:

```bash
# From the website root directory
./scripts/update-research.sh

# Or with a specific Google Scholar user ID
./scripts/update-research.sh YOUR_SCHOLAR_ID
```

### Building the Site

After updating research, build your Hugo site:

```bash
# Update research publications
./scripts/update-research.sh

# Build the Hugo site
hugo --minify

# Or serve locally for development
hugo serve -D
```

### Automatic Updates

The GitHub Action in `.github/workflows/update-research.yml` automatically:

- Runs weekly on Sundays at 6 AM UTC
- Updates the research page when the scraper code changes
- Can be triggered manually from the GitHub Actions tab

## Configuration

The default Google Scholar user ID is hardcoded in `scholar-scraper.go`. To change it:

1. Edit the `scholarID` variable in `scholar-scraper.go`
2. Or pass it as an argument to the script: `./update-research.sh YOUR_ID`

## Output

The script generates a markdown file at `content/research.md` with:

- Publications grouped by year (newest first)
- Links to Google Scholar citation pages
- Author information
- Publication venues
- Last updated timestamp