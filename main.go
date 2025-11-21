package main

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gocolly/colly/v2"
)

type Sitemap struct {
	XMLName xml.Name `xml:"urlset"`
	URLs    []struct {
		Loc string `xml:"loc"`
	} `xml:"url"`
}

type SitemapIndex struct {
	XMLName  xml.Name `xml:"sitemapindex"`
	Sitemaps []struct {
		Loc string `xml:"loc"`
	} `xml:"sitemap"`
}

const (
	progressFile = "progress.txt"
	urlsFile     = "urls.txt"
	edgesFile    = "edges.tsv"
)

var (
	completed     = make(map[string]bool)
	completedMu   sync.RWMutex
	savedCount    int64
	edgeCount     int64
	progressFd    *os.File
	edgesFd       *os.File
	edgesMu       sync.Mutex
)

func main() {
	// Load completed URLs from progress file
	loadProgress()

	// Get all URLs (from cache or fetch)
	allURLs := getAllURLs()

	// Filter out completed
	var pending []string
	for _, url := range allURLs {
		completedMu.RLock()
		done := completed[url]
		completedMu.RUnlock()
		if !done {
			pending = append(pending, url)
		}
	}

	fmt.Printf("Total URLs: %d, Already done: %d, Pending: %d\n", len(allURLs), len(completed), len(pending))

	if len(pending) == 0 {
		fmt.Println("All URLs already downloaded!")
		return
	}

	// Open progress file for appending
	var err error
	progressFd, err = os.OpenFile(progressFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer progressFd.Close()

	// Open edges file for appending (TSV: source \t target)
	edgesFd, err = os.OpenFile(edgesFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer edgesFd.Close()

	// Create collector with higher parallelism
	c := colly.NewCollector(
		colly.AllowedDomains("www.fedlex.admin.ch", "fedlex.admin.ch"),
		colly.Async(true),
	)

	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Delay:       20 * time.Millisecond,
		Parallelism: 100,
	})

	c.OnResponse(func(r *colly.Response) {
		url := r.Request.URL.String()
		savePath := urlToPath(url)
		os.MkdirAll(filepath.Dir(savePath), 0755)
		r.Save(savePath)

		// Mark as done
		completedMu.Lock()
		completed[url] = true
		progressFd.WriteString(url + "\n")
		completedMu.Unlock()

		count := atomic.AddInt64(&savedCount, 1)
		if count%100 == 0 {
			fmt.Printf("Progress: %d/%d (%.2f%%)\n", count, len(pending), float64(count)/float64(len(pending))*100)
		}
	})

	c.OnError(func(r *colly.Response, err error) {
		if r.StatusCode == 429 {
			fmt.Printf("Rate limited! Slowing down...\n")
			time.Sleep(5 * time.Second)
			r.Request.Retry()
		} else {
			fmt.Printf("Error %d: %s\n", r.StatusCode, r.Request.URL)
		}
	})

	// Record all links as edges and follow PDF/XML links
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		sourceURL := e.Request.URL.String()
		link := e.Attr("href")
		targetURL := e.Request.AbsoluteURL(link)

		// Only record edges within the domain
		if strings.Contains(targetURL, "fedlex.admin.ch") && targetURL != "" && targetURL != sourceURL {
			edgesMu.Lock()
			edgesFd.WriteString(sourceURL + "\t" + targetURL + "\n")
			atomic.AddInt64(&edgeCount, 1)
			edgesMu.Unlock()
		}

		// Follow PDF/XML links
		if strings.HasSuffix(link, ".pdf") || strings.HasSuffix(link, ".xml") {
			e.Request.Visit(link)
		}
	})

	// Start crawling
	start := time.Now()
	for _, url := range pending {
		c.Visit(url)
	}

	c.Wait()

	elapsed := time.Since(start)
	fmt.Printf("\nDone! Downloaded %d pages, recorded %d edges in %v (%.1f pages/sec)\n",
		atomic.LoadInt64(&savedCount), atomic.LoadInt64(&edgeCount), elapsed, float64(savedCount)/elapsed.Seconds())
}

func loadProgress() {
	f, err := os.Open(progressFile)
	if err != nil {
		return // No progress file yet
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		completed[scanner.Text()] = true
	}
	fmt.Printf("Loaded %d completed URLs from progress file\n", len(completed))
}

func getAllURLs() []string {
	// Try to load from cache
	if urls := loadURLsCache(); len(urls) > 0 {
		return urls
	}

	fmt.Println("Fetching sitemaps...")
	sitemapURLs := []string{
		"https://www.fedlex.admin.ch/sitemap-index.xml",
		"https://www.fedlex.admin.ch/sitemap-consultations-1.xml",
		"https://www.fedlex.admin.ch/sitemap-treaty-1.xml",
		"https://www.fedlex.admin.ch/sitemap-treaty-2.xml",
		"https://www.fedlex.admin.ch/sitemap-treaty-3.xml",
		"https://www.fedlex.admin.ch/sitemap-treaty-4.xml",
		"https://www.fedlex.admin.ch/sitemap-treaty-5.xml",
		"https://www.fedlex.admin.ch/sitemap-treaty-6.xml",
		"https://www.fedlex.admin.ch/sitemap-treaty-7.xml",
		"https://www.fedlex.admin.ch/sitemap-treaty-8.xml",
		"https://www.fedlex.admin.ch/sitemap-treaty-9.xml",
		"https://www.fedlex.admin.ch/sitemap-treaty-10.xml",
		"https://www.fedlex.admin.ch/sitemap-act-1.xml",
		"https://www.fedlex.admin.ch/sitemap-act-2.xml",
		"https://www.fedlex.admin.ch/sitemap-act-3.xml",
		"https://www.fedlex.admin.ch/sitemap-act-4.xml",
		"https://www.fedlex.admin.ch/sitemap-act-5.xml",
		"https://www.fedlex.admin.ch/sitemap-act-6.xml",
		"https://www.fedlex.admin.ch/sitemap-act-7.xml",
		"https://www.fedlex.admin.ch/sitemap-act-8.xml",
		"https://www.fedlex.admin.ch/sitemap-act-9.xml",
		"https://www.fedlex.admin.ch/sitemap-act-10.xml",
		"https://www.fedlex.admin.ch/sitemap-act-11.xml",
		"https://www.fedlex.admin.ch/sitemap-act-12.xml",
		"https://www.fedlex.admin.ch/sitemap-act-13.xml",
		"https://www.fedlex.admin.ch/sitemap-act-14.xml",
		"https://www.fedlex.admin.ch/sitemap-act-15.xml",
		"https://www.fedlex.admin.ch/sitemap-act-16.xml",
		"https://www.fedlex.admin.ch/sitemap-act-17.xml",
		"https://www.fedlex.admin.ch/sitemap-act-18.xml",
		"https://www.fedlex.admin.ch/sitemap-act-19.xml",
		"https://www.fedlex.admin.ch/sitemap-act-20.xml",
		"https://www.fedlex.admin.ch/sitemap-act-21.xml",
		"https://www.fedlex.admin.ch/sitemap-act-22.xml",
		"https://www.fedlex.admin.ch/sitemap-act-23.xml",
		"https://www.fedlex.admin.ch/sitemap-act-24.xml",
		"https://www.fedlex.admin.ch/sitemap-act-25.xml",
		"https://www.fedlex.admin.ch/sitemap-act-26.xml",
		"https://www.fedlex.admin.ch/sitemap-act-27.xml",
		"https://www.fedlex.admin.ch/sitemap-cc1-1.xml",
		"https://www.fedlex.admin.ch/sitemap-cc1-2.xml",
	}

	var allPageURLs []string
	for _, sitemapURL := range sitemapURLs {
		fmt.Printf("Parsing: %s\n", sitemapURL)
		urls := parseSitemap(sitemapURL)
		allPageURLs = append(allPageURLs, urls...)
	}

	// Deduplicate
	seen := make(map[string]bool)
	var uniqueURLs []string
	for _, url := range allPageURLs {
		if !seen[url] {
			seen[url] = true
			uniqueURLs = append(uniqueURLs, url)
		}
	}

	// Cache URLs
	saveURLsCache(uniqueURLs)

	return uniqueURLs
}

func loadURLsCache() []string {
	f, err := os.Open(urlsFile)
	if err != nil {
		return nil
	}
	defer f.Close()

	var urls []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		urls = append(urls, scanner.Text())
	}
	fmt.Printf("Loaded %d URLs from cache\n", len(urls))
	return urls
}

func saveURLsCache(urls []string) {
	f, _ := os.Create(urlsFile)
	defer f.Close()
	for _, url := range urls {
		f.WriteString(url + "\n")
	}
	fmt.Printf("Cached %d URLs\n", len(urls))
}

func parseSitemap(url string) []string {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("Error fetching sitemap %s: %v\n", url, err)
		return nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading sitemap %s: %v\n", url, err)
		return nil
	}

	// Try as sitemap index first
	var index SitemapIndex
	if xml.Unmarshal(body, &index) == nil && len(index.Sitemaps) > 0 {
		var urls []string
		for _, s := range index.Sitemaps {
			urls = append(urls, parseSitemap(s.Loc)...)
		}
		return urls
	}

	// Parse as regular sitemap
	var sitemap Sitemap
	if err := xml.Unmarshal(body, &sitemap); err != nil {
		fmt.Printf("Error parsing sitemap %s: %v\n", url, err)
		return nil
	}

	var urls []string
	for _, u := range sitemap.URLs {
		urls = append(urls, u.Loc)
	}
	fmt.Printf("  -> Found %d URLs\n", len(urls))
	return urls
}

func urlToPath(urlStr string) string {
	path := strings.TrimPrefix(urlStr, "https://")
	path = strings.TrimPrefix(path, "http://")
	path = strings.ReplaceAll(path, "?", "_")
	path = strings.ReplaceAll(path, "&", "_")

	if strings.HasSuffix(path, "/") || !strings.Contains(filepath.Base(path), ".") {
		path = filepath.Join(path, "index.html")
	}
	return filepath.Join("mirror", path)
}
