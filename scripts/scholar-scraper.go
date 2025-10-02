package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
)

type Publication struct {
	Title     string
	Authors   string
	Journal   string
	Year      string
	URL       string
	Citations int
}

func main() {
	scholarID := "o9sbhDUAAAAJ"
	if len(os.Args) > 1 {
		scholarID = os.Args[1]
	}

	publications, err := scrapeGoogleScholar(scholarID)
	if err != nil {
		log.Fatalf("Error scraping Google Scholar: %v", err)
	}

	err = generateMarkdown(publications)
	if err != nil {
		log.Fatalf("Error generating markdown: %v", err)
	}

	fmt.Printf("Successfully scraped %d publications\n", len(publications))
}

func scrapeGoogleScholar(userID string) ([]Publication, error) {
	var publications []Publication

	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"),
	)

	// Add delay to avoid being blocked
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*scholar.google.*",
		Parallelism: 1,
		Delay:       2 * time.Second,
	})

	c.OnHTML("tr.gsc_a_tr", func(e *colly.HTMLElement) {
		titleElement := e.DOM.Find("a.gsc_a_at")
		title := strings.TrimSpace(titleElement.Text())
		if title == "" {
			return
		}

		// Get the publication URL
		href, exists := titleElement.Attr("href")
		var url string
		if exists {
			url = "https://scholar.google.com" + href
		}

		// Get authors and journal info
		authorsJournal := strings.TrimSpace(e.DOM.Find("div.gs_gray").First().Text())
		
		// Get year
		yearText := strings.TrimSpace(e.DOM.Find("span.gsc_a_h").Text())
		
		// Get citations
		citationsText := strings.TrimSpace(e.DOM.Find("a.gsc_a_ac").Text())
		citations := 0
		if citationsText != "" {
			if c, err := strconv.Atoi(citationsText); err == nil {
				citations = c
			}
		}

		// Split authors and journal
		authors, journal := parseAuthorsJournal(authorsJournal)

		publication := Publication{
			Title:     title,
			Authors:   authors,
			Journal:   journal,
			Year:      yearText,
			URL:       url,
			Citations: citations,
		}

		publications = append(publications, publication)
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Printf("Error: %v", err)
	})

	url := fmt.Sprintf("https://scholar.google.com/citations?hl=en&user=%s&view_op=list_works&sortby=pubdate", userID)
	err := c.Visit(url)
	if err != nil {
		return nil, fmt.Errorf("failed to visit URL: %v", err)
	}

	return publications, nil
}

func parseAuthorsJournal(text string) (authors, journal string) {
	// Common separators between authors and journal
	separators := []string{" - ", " – ", " — "}
	
	for _, sep := range separators {
		if strings.Contains(text, sep) {
			parts := strings.SplitN(text, sep, 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
			}
		}
	}
	
	// If no separator found, assume it's all authors
	return text, ""
}

func generateMarkdown(publications []Publication) error {
	// Group publications by year
	yearGroups := make(map[string][]Publication)
	
	for _, pub := range publications {
		year := pub.Year
		if year == "" {
			year = "Unknown"
		}
		yearGroups[year] = append(yearGroups[year], pub)
	}

	// Generate markdown content
	var content strings.Builder
	
	content.WriteString(`---
showDate : false
showAuthor : false
showDateOnlyInArticle : false
showDateUpdated : false
showHeadingAnchors : false
showPagination : false
showReadingTime : false
showTableOfContents : true
showTaxonomies : false 
showWordCount : false
showSummary : false
sharingLinks : false
showEdit: false
showViews: false
showLikes: false
layoutBackgroundHeaderSpace: false
---

# Publications

*Last updated: ` + time.Now().Format("January 2, 2006") + `*

`)

	// Sort years in descending order
	years := make([]string, 0, len(yearGroups))
	for year := range yearGroups {
		years = append(years, year)
	}
	
	// Simple sort for years (descending)
	for i := 0; i < len(years); i++ {
		for j := i + 1; j < len(years); j++ {
			if years[i] < years[j] {
				years[i], years[j] = years[j], years[i]
			}
		}
	}

	for _, year := range years {
		pubs := yearGroups[year]
		if year == "Unknown" {
			continue // Skip unknown years for now
		}
		
		content.WriteString(fmt.Sprintf("\n## %s\n\n", year))
		
		for _, pub := range pubs {
			// Format the publication entry
			entry := formatPublication(pub)
			content.WriteString(entry + "\n\n")
		}
	}

	// Write to file (relative to parent directory since script runs from scripts/)
	err := os.WriteFile("../content/research.md", []byte(content.String()), 0644)
	if err != nil {
		return fmt.Errorf("failed to write markdown file: %v", err)
	}

	return nil
}

func formatPublication(pub Publication) string {
	var entry strings.Builder
	
	// Title with link if available
	if pub.URL != "" {
		entry.WriteString(fmt.Sprintf("- [**%s**](%s)", pub.Title, pub.URL))
	} else {
		entry.WriteString(fmt.Sprintf("- **%s**", pub.Title))
	}
	
	// Authors
	if pub.Authors != "" {
		entry.WriteString(fmt.Sprintf("  \n  *%s*", pub.Authors))
	}
	
	// Journal/Conference
	if pub.Journal != "" {
		entry.WriteString(fmt.Sprintf("  \n  Published in *%s*", pub.Journal))
		if pub.Year != "" {
			entry.WriteString(fmt.Sprintf(", %s", pub.Year))
		}
	} else if pub.Year != "" {
		entry.WriteString(fmt.Sprintf("  \n  Published in %s", pub.Year))
	}
	
	entry.WriteString(".")
	
	return entry.String()
}