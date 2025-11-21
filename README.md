# Fedlex.admin.ch Complete Website Scraper

A high-performance web scraper for creating a complete offline mirror of [fedlex.admin.ch](https://www.fedlex.admin.ch) (Swiss Federal Law database) with link graph extraction.

## Overview

This project consists of two components:

1. **Main Crawler** (Go/Colly): Downloads all HTML pages from fedlex.admin.ch
2. **Link Extractor** (Node.js/Playwright): Extracts link relationships to build a directed graph

## Features

- **Complete Coverage**: Discovers all 377,158+ URLs from sitemaps
- **Resume Capability**: Can stop/restart without losing progress
- **High Performance**:
  - Main crawler: ~100 pages/second
  - Link extractor: ~30 pages/second (renders local files with Playwright)
- **Link Graph**: Builds directed graph of all internal links (TSV format)
- **Efficient Storage**: ~50MB per 1000 pages

## Requirements

### System Requirements
- **Disk Space**: 30-40GB minimum (final size ~25-30GB)
- **RAM**: 4GB minimum, 8GB recommended
- **CPU**: Multi-core recommended for parallel processing

### Software Requirements
- **Go**: 1.19 or later
- **Node.js**: 16.x or later
- **npm**: 7.x or later

## Setup Instructions

### 1. Install Dependencies

#### Ubuntu/Debian
```bash
# Install Go
sudo apt update
sudo apt install golang-go

# Install Node.js and npm
curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
sudo apt install -y nodejs

# Verify installations
go version
node --version
npm --version
```

#### macOS
```bash
# Install with Homebrew
brew install go node

# Verify installations
go version
node --version
```

### 2. Clone/Copy Project

```bash
# Create project directory
mkdir fedlex-scraper
cd fedlex-scraper

# Copy all project files here
# (main.go, extract-links.js, package.json)
```

### 3. Install Go Dependencies

```bash
go mod init fedlex-crawler
go get github.com/gocolly/colly/v2
```

### 4. Install Node.js Dependencies

```bash
npm install playwright
npx playwright install chromium
```

## Project Structure

```
fedlex-scraper/
├── main.go                    # Main crawler (Go)
├── extract-links.js           # Link extractor (Node.js)
├── go.mod                     # Go dependencies
├── package.json               # Node dependencies
├── progress.txt               # Main crawler progress (resume file)
├── links-progress.txt         # Link extractor progress (resume file)
├── urls.txt                   # Cached list of all URLs (377K URLs)
├── edges.tsv                  # Output: link graph edges
├── mirror/                    # Output: downloaded website
│   └── www.fedlex.admin.ch/
│       ├── de/
│       ├── fr/
│       ├── it/
│       ├── en/
│       └── eli/
└── README.md
```

## Usage

### Running the Main Crawler

The main crawler downloads all HTML pages:

```bash
go run main.go
```

**Features:**
- Downloads from sitemaps (377,158 URLs)
- 100 parallel connections
- 20ms delay between requests
- Auto-resume from `progress.txt`
- Saves to `mirror/` directory

**Output:**
```
Total URLs: 377158, Already done: 0, Pending: 377158
Progress: 100/377158 (0.03%)
Progress: 200/377158 (0.05%)
...
Done! Downloaded 377158 pages in 10542s (35.8 pages/sec)
```

### Running the Link Extractor

The link extractor builds the link graph by rendering local HTML files:

```bash
node extract-links.js
```

**Features:**
- Loads local HTML files (no network requests)
- Renders with Playwright (handles JavaScript)
- 20 concurrent browser tabs
- Auto-resume from `links-progress.txt`
- Outputs to `edges.tsv`

**Output:**
```
URLs to process: 214429
Progress: 100/214429 (0.0%) | Edges: 45234 | 29.5/s | ETA: 121m
Progress: 200/214429 (0.1%) | Edges: 90821 | 30.1/s | ETA: 118m
...
Done! Processed 214429 pages, found 9847234 edges in 7245s (29.6 pages/sec)
```

## Output Formats

### 1. Website Mirror (`mirror/`)

Directory structure mirrors the website:
```
mirror/www.fedlex.admin.ch/eli/treaty/1852/0001/de/index.html
mirror/www.fedlex.admin.ch/eli/treaty/1852/0001/fr/index.html
mirror/www.fedlex.admin.ch/eli/treaty/1852/0001/it/index.html
```

### 2. Link Graph (`edges.tsv`)

Tab-separated values (source → target):
```tsv
https://www.fedlex.admin.ch/de/home	https://www.fedlex.admin.ch/de/cc/internal-law
https://www.fedlex.admin.ch/de/home	https://www.fedlex.admin.ch/de/cc/international-law
https://www.fedlex.admin.ch/eli/treaty/1852/0001/de	https://www.fedlex.admin.ch/eli/treaty/1852/0001/fr
```

**Import into graph databases:**
```python
# NetworkX (Python)
import networkx as nx
G = nx.read_edgelist('edges.tsv', create_using=nx.DiGraph(), delimiter='\t')

# Neo4j
LOAD CSV FROM 'file:///edges.tsv' AS row FIELDTERMINATOR '\t'
MERGE (a:Page {url: row[0]})
MERGE (b:Page {url: row[1]})
MERGE (a)-[:LINKS_TO]->(b)
```

## Resume Capability

Both crawlers support resume - you can stop and restart anytime:

**Main Crawler:**
- Tracks completed URLs in `progress.txt`
- On restart, skips already-downloaded pages
- Safe to Ctrl+C and restart

**Link Extractor:**
- Tracks completed URLs in `links-progress.txt`
- Skips already-processed pages
- Safe to Ctrl+C and restart

**To start fresh:**
```bash
# Remove progress files to start over
rm progress.txt links-progress.txt edges.tsv
rm -rf mirror/
```

## Performance Tuning

### Main Crawler (`main.go`)

Adjust these parameters in the code:

```go
c.Limit(&colly.LimitRule{
    DomainGlob:  "*",
    Delay:       20 * time.Millisecond,  // Delay between requests
    Parallelism: 100,                     // Concurrent connections
})
```

**Recommendations:**
- **Fast network**: Increase `Parallelism` to 200+, reduce `Delay` to 10ms
- **Slow network**: Decrease `Parallelism` to 50, increase `Delay` to 50ms
- **Rate limiting**: If you get 429 errors, increase `Delay`

### Link Extractor (`extract-links.js`)

Adjust concurrency:

```javascript
const CONCURRENCY = 20;  // Concurrent browser tabs
```

**Recommendations:**
- **More RAM**: Increase to 40-50 tabs
- **Less RAM**: Decrease to 10-15 tabs
- Each tab uses ~50-100MB RAM

## Monitoring Progress

### Check current status
```bash
# Main crawler
wc -l progress.txt        # Pages downloaded
du -sh mirror/            # Disk usage

# Link extractor
wc -l links-progress.txt  # Pages processed
wc -l edges.tsv           # Edges found
```

### Calculate percentage
```bash
# Main crawler progress
echo "scale=2; $(wc -l < progress.txt) / 377158 * 100" | bc

# Link extractor progress
echo "scale=2; $(wc -l < links-progress.txt) / $(wc -l < progress.txt) * 100" | bc
```

### Live monitoring
```bash
# Watch progress in real-time
watch -n 5 'wc -l progress.txt links-progress.txt edges.tsv && du -sh mirror/'
```

## Estimated Completion Times

### Main Crawler
- **100 pages/sec**: ~1 hour for 377K pages
- **50 pages/sec**: ~2 hours
- **25 pages/sec**: ~4 hours

### Link Extractor
- **30 pages/sec**: ~2-3 hours for 200K+ pages
- **20 pages/sec**: ~3-4 hours
- **10 pages/sec**: ~6-8 hours

**Total time**: 3-5 hours for complete mirror + graph

## Running Both Simultaneously

You can run both crawlers at the same time:

```bash
# Terminal 1: Main crawler
go run main.go

# Terminal 2: Link extractor (wait for main to download some pages first)
node extract-links.js
```

The link extractor only processes pages that exist in `mirror/`, so it's safe to run both concurrently.

## Troubleshooting

### "Error 0" in main crawler
- Network connection issues
- Server temporarily down
- Non-critical - crawler continues

### Link extractor timeouts
- Increase timeout in `extract-links.js`:
  ```javascript
  await page.goto(fileUrl, { timeout: 10000 })  // Increase to 10s
  ```

### Out of disk space
- Main mirror: ~25-30GB
- Check space: `df -h`
- Clean up: `rm -rf mirror/` to start fresh

### Out of memory (link extractor)
- Reduce `CONCURRENCY` in `extract-links.js`
- Close other applications
- Monitor: `htop` or `top`

### Chromium download fails
- Manual install: `npx playwright install chromium --with-deps`
- Or install system chromium: `sudo apt install chromium-browser`

## Data Analysis Examples

### Graph statistics
```python
import networkx as nx

G = nx.read_edgelist('edges.tsv', create_using=nx.DiGraph(), delimiter='\t')

print(f"Nodes: {G.number_of_nodes()}")
print(f"Edges: {G.number_of_edges()}")
print(f"Average degree: {sum(dict(G.degree()).values()) / G.number_of_nodes():.2f}")

# Most linked pages
in_degree = sorted(G.in_degree(), key=lambda x: x[1], reverse=True)
print("Most referenced pages:", in_degree[:10])
```

### Language distribution
```bash
# Count pages by language
find mirror/ -name "index.html" | grep -o '/de/\|/fr/\|/it/\|/en/' | sort | uniq -c
```

### Document types
```bash
# Count by document type
grep -o '/eli/[^/]*' urls.txt | sort | uniq -c
```

## License

This scraper is for archival and research purposes. The content from fedlex.admin.ch belongs to the Swiss Federal Government and is subject to their terms of use.

## Credits

- Built with [Colly](https://github.com/gocolly/colly) (Go web scraping framework)
- Built with [Playwright](https://playwright.dev/) (Browser automation)
- Data source: [Fedlex](https://www.fedlex.admin.ch) - Swiss Federal Law Platform

## Support

For issues or questions:
1. Check Troubleshooting section above
2. Review code comments in `main.go` and `extract-links.js`
3. Monitor console output for specific error messages
