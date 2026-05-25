package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
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
	var (
		scholarID string
		filePath  string
		output    string
	)
	if len(os.Args) > 1 {
		scholarID = os.Args[1]
	}

	flag.StringVar(&scholarID, "id", "o9sbhDUAAAAJ", "Google Scholar user ID")
	flag.StringVar(&filePath, "file", "", "Path to saved Google Scholar HTML file (offline mode)")
	flag.StringVar(&output, "output", "", "Output markdown file path (default: ../content/research.md)")
	flag.Parse()

	if output == "" {
		output = "../content/research.md"
	}

	var publications []Publication
	var err error

	if filePath != "" {
		publications, err = parseFromHTMLFile(filePath)
		if err != nil {
			log.Fatalf("Error parsing HTML file: %v", err)
		}
	} else {
		publications, err = scrapeGoogleScholar(scholarID)
		if err != nil {
			log.Fatalf("Error scraping Google Scholar: %v", err)
		}
	}

	err = generateMarkdown(publications, output)
	if err != nil {
		log.Fatalf("Error generating markdown: %v", err)
	}

	fmt.Printf("Successfully scraped %d publications\n", len(publications))
	if filePath != "" {
		fmt.Println("Warning: offline mode - publication count limited to what was saved in the HTML file")
	}
}

func scrapeGoogleScholar(userID string) ([]Publication, error) {
	var publications []Publication
	lastStatus := 0

	maxRetries := 6
	if v := strings.TrimSpace(os.Getenv("SCHOLAR_SCRAPER_MAX_RETRIES")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			maxRetries = n
		}
	}

	cookieJarFile := "scholar_cookies.json"
	proxyURL := strings.TrimSpace(os.Getenv("SCHOLAR_SCRAPER_PROXY"))

	c := colly.NewCollector(
		colly.UserAgent(pickUA()),
	)
	c.AllowURLRevisit = true

	jar, err := loadCookieJar(cookieJarFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load cookie jar: %v", err)
	}
	c.SetCookieJar(jar)

	if proxyURL != "" {
		if err := c.SetProxy(proxyURL); err != nil {
			return nil, fmt.Errorf("failed to set proxy: %v", err)
		}
	}

	c.Limit(&colly.LimitRule{
		DomainGlob:  "*scholar.google.*",
		Parallelism: 1,
		Delay:       5 * time.Second,
		RandomDelay: 5 * time.Second,
	})

	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("User-Agent", pickUA())
		r.Headers.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
		r.Headers.Set("Accept-Language", "en-US,en;q=0.9")
		r.Headers.Set("Cache-Control", "no-cache")
		r.Headers.Set("Pragma", "no-cache")
		r.Headers.Set("Sec-Ch-Ua", `"Google Chrome";v="131", "Chromium";v="131", "Not_A Brand";v="24"`)
		r.Headers.Set("Sec-Ch-Ua-Mobile", "?0")
		r.Headers.Set("Sec-Ch-Ua-Platform", `"Windows"`)
		r.Headers.Set("Upgrade-Insecure-Requests", "1")
		r.Headers.Set("DNT", "1")
	})

	c.OnResponse(func(r *colly.Response) {
		lastStatus = r.StatusCode
		saveCookieJar(cookieJarFile, c)
	})

	c.OnHTML("tr.gsc_a_tr", func(e *colly.HTMLElement) {
		titleElement := e.DOM.Find("a.gsc_a_at")
		title := strings.TrimSpace(titleElement.Text())
		if title == "" {
			return
		}

		href, exists := titleElement.Attr("href")
		var pubURL string
		if exists {
			pubURL = "https://scholar.google.com" + href
		}

		authorsJournal := strings.TrimSpace(e.DOM.Find("div.gs_gray").First().Text())
		yearText := strings.TrimSpace(e.DOM.Find("span.gsc_a_h").Text())

		citationsText := strings.TrimSpace(e.DOM.Find("a.gsc_a_ac").Text())
		citations := 0
		if citationsText != "" {
			if c, err := strconv.Atoi(citationsText); err == nil {
				citations = c
			}
		}

		authors, journal := parseAuthorsJournal(authorsJournal)

		publication := Publication{
			Title:     title,
			Authors:   authors,
			Journal:   journal,
			Year:      yearText,
			URL:       pubURL,
			Citations: citations,
		}

		publications = append(publications, publication)
	})

	c.OnError(func(r *colly.Response, err error) {
		if r != nil {
			lastStatus = r.StatusCode
		}
		log.Printf("Error: %v", err)
	})

	pagesize := 20
	for cstart := 0; ; cstart += pagesize {
		publicationsBefore := len(publications)

		pageURL := fmt.Sprintf(
			"https://scholar.google.com/citations?hl=en&user=%s&view_op=list_works&sortby=pubdate&cstart=%d&pagesize=%d",
			userID,
			cstart,
			pagesize,
		)

		lastStatus = 0
		for attempt := 0; attempt <= maxRetries; attempt++ {
			err = c.Visit(pageURL)
			if err == nil {
				break
			}

			tooMany := lastStatus == 429 || strings.Contains(strings.ToLower(err.Error()), "too many requests")
			if !tooMany || attempt == maxRetries {
				return nil, fmt.Errorf("failed to visit URL: %v", err)
			}

			backoff := time.Duration(10*(1<<attempt)) * time.Second
			if backoff > 5*time.Minute {
				backoff = 5 * time.Minute
			}
			jitter := time.Duration(rand.IntN(2000)) * time.Millisecond
			log.Printf("Rate limited (HTTP 429). Retrying in %s...", (backoff+jitter).Round(time.Second))
			time.Sleep(backoff + jitter)
		}

		c.Wait()

		if len(publications) == publicationsBefore {
			break
		}
	}

	if len(publications) == 0 {
		return nil, fmt.Errorf("no publications found — likely blocked by Google Scholar. Try saving the page HTML and using -file flag")
	}

	saveCookieJar(cookieJarFile, c)
	return publications, nil
}

func parseFromHTMLFile(filePath string) ([]Publication, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("reading file: %v", err)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("parsing HTML: %v", err)
	}

	var publications []Publication

	doc.Find("tr.gsc_a_tr").Each(func(i int, s *goquery.Selection) {
		titleSel := s.Find("a.gsc_a_at")
		title := strings.TrimSpace(titleSel.Text())
		if title == "" {
			return
		}

		var pubURL string
		if href, exists := titleSel.Attr("href"); exists {
			pubURL = "https://scholar.google.com" + href
		}

		authorsJournal := strings.TrimSpace(s.Find("div.gs_gray").First().Text())
		yearText := strings.TrimSpace(s.Find("span.gsc_a_h").Text())

		citationsText := strings.TrimSpace(s.Find("a.gsc_a_ac").Text())
		citations := 0
		if citationsText != "" {
			if c, err := strconv.Atoi(citationsText); err == nil {
				citations = c
			}
		}

		authors, journal := parseAuthorsJournal(authorsJournal)

		publications = append(publications, Publication{
			Title:     title,
			Authors:   authors,
			Journal:   journal,
			Year:      yearText,
			URL:       pubURL,
			Citations: citations,
		})
	})

	if len(publications) == 0 {
		return nil, fmt.Errorf("no publication rows found in HTML file")
	}

	return publications, nil
}

func parseAuthorsJournal(text string) (authors, journal string) {
	separators := []string{" - ", " – ", " — "}

	for _, sep := range separators {
		if strings.Contains(text, sep) {
			parts := strings.SplitN(text, sep, 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
			}
		}
	}

	return text, ""
}

func generateMarkdown(publications []Publication, outputPath string) error {
	yearGroups := make(map[string][]Publication)

	for _, pub := range publications {
		year := pub.Year
		if year == "" {
			year = "Unknown"
		}
		yearGroups[year] = append(yearGroups[year], pub)
	}

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

	years := make([]string, 0, len(yearGroups))
	for year := range yearGroups {
		years = append(years, year)
	}

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
			continue
		}

		content.WriteString(fmt.Sprintf("\n## %s\n\n", year))

		for _, pub := range pubs {
			entry := formatPublication(pub)
			content.WriteString(entry + "\n\n")
		}
	}

	absPath, err := filepath.Abs(outputPath)
	if err != nil {
		absPath = outputPath
	}

	if err := os.MkdirAll(filepath.Dir(absPath), 0755); err != nil {
		return fmt.Errorf("creating output directory: %v", err)
	}

	err = os.WriteFile(absPath, []byte(content.String()), 0644)
	if err != nil {
		return fmt.Errorf("writing markdown file: %v", err)
	}

	fmt.Printf("Wrote %s\n", absPath)
	return nil
}

func formatPublication(pub Publication) string {
	var entry strings.Builder

	if pub.URL != "" {
		entry.WriteString(fmt.Sprintf("- [**%s**](%s)", pub.Title, pub.URL))
	} else {
		entry.WriteString(fmt.Sprintf("- **%s**", pub.Title))
	}

	if pub.Authors != "" {
		entry.WriteString(fmt.Sprintf("  \n  *%s*", pub.Authors))
	}

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

var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:136.0) Gecko/20100101 Firefox/136.0",
}

func pickUA() string {
	return userAgents[rand.IntN(len(userAgents))]
}

func loadCookieJar(path string) (http.CookieJar, error) {
	jar, err := readCookieJar(path)
	if err != nil {
		log.Printf("Cookie jar not loaded (%v), starting fresh", err)
		return newCookieJar(), nil
	}
	return jar, nil
}

func newCookieJar() http.CookieJar {
	return NewMemoryCookieJar()
}

func readCookieJar(path string) (http.CookieJar, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	jar := NewMemoryCookieJar()

	scholarURL, _ := url.Parse("https://scholar.google.com")
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(line, ";")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if idx := strings.Index(part, "="); idx > 0 {
				name := strings.TrimSpace(part[:idx])
				value := strings.TrimSpace(part[idx+1:])
				jar.SetCookies(scholarURL, []*http.Cookie{{
					Name:  name,
					Value: value,
				}})
			}
		}
	}

	return jar, nil
}

func saveCookieJar(path string, c *colly.Collector) {
	cookies := c.Cookies("https://scholar.google.com")
	if len(cookies) == 0 {
		return
	}
	var lines []string
	lines = append(lines, "# Netscape HTTP Cookie File")
	for _, ck := range cookies {
		domain := "scholar.google.com"
		if ck.Domain != "" {
			domain = ck.Domain
		}
		flag := "TRUE"
		secure := "TRUE"
		pathCookie := ck.Path
		if pathCookie == "" {
			pathCookie = "/"
		}
		tail := fmt.Sprintf("%s\t%s\t%s\t%s\t%d\t%s\t%s", domain, flag, pathCookie, secure, ck.Expires.Unix(), ck.Name, ck.Value)
		lines = append(lines, tail)
	}
	_ = os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0600)
}

type MemoryCookieJar struct {
	cookies map[string][]*http.Cookie
}

func NewMemoryCookieJar() *MemoryCookieJar {
	return &MemoryCookieJar{cookies: make(map[string][]*http.Cookie)}
}

func (j *MemoryCookieJar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	j.cookies[u.Host] = append(j.cookies[u.Host], cookies...)
}

func (j *MemoryCookieJar) Cookies(u *url.URL) []*http.Cookie {
	return j.cookies[u.Host]
}
