# Quick Start Guide

## For the Colleague Running This on a Fresh Machine

### Prerequisites
- Ubuntu/Debian Linux (or macOS)
- 35GB+ free disk space
- 4GB+ RAM (8GB recommended)
- Internet connection

### Installation (5 minutes)

1. **Copy project files to machine:**
   ```bash
   # Copy these files to a new directory:
   # - main.go
   # - extract-links.js
   # - package.json
   # - setup.sh
   # - README.md
   ```

2. **Run setup script:**
   ```bash
   chmod +x setup.sh
   ./setup.sh
   ```

   This will:
   - Check Go and Node.js are installed
   - Install dependencies
   - Download Chromium browser

### Running (3-5 hours)

**Option A: Run both crawlers simultaneously (faster)**
```bash
# Terminal 1 - Main crawler
go run main.go

# Terminal 2 - Link extractor (start after 5 minutes)
node extract-links.js
```

**Option B: Run sequentially (simpler)**
```bash
# Step 1: Download website (~2 hours)
go run main.go

# Step 2: Extract links (~2 hours)
node extract-links.js
```

### Monitoring Progress

**Check status:**
```bash
# Quick check
wc -l progress.txt        # Pages downloaded
wc -l edges.tsv           # Links extracted
du -sh mirror/            # Disk usage

# Live monitoring (updates every 5 seconds)
watch -n 5 'wc -l progress.txt links-progress.txt edges.tsv && du -sh mirror/'
```

**Calculate percentage:**
```bash
# Main crawler (out of 377,158 total pages)
echo "scale=1; $(wc -l < progress.txt) / 377158 * 100" | bc

# Link extractor
echo "scale=1; $(wc -l < links-progress.txt) / $(wc -l < progress.txt) * 100" | bc
```

### Output Files

After completion, you'll have:

1. **`mirror/`** - Complete offline website (~25-30GB)
   - Organized by URL structure
   - All HTML pages in German, French, Italian, English

2. **`edges.tsv`** - Link graph (~500MB-1GB)
   - Format: `source_url<TAB>target_url`
   - ~5-10 million edges
   - Ready for NetworkX, Neo4j, etc.

3. **Progress files** (for resume):
   - `progress.txt` - Downloaded pages
   - `links-progress.txt` - Processed pages
   - `urls.txt` - All 377K URLs (cached)

### Resume After Interruption

If stopped (Ctrl+C, crash, network issue), just restart:

```bash
# Resume main crawler
go run main.go

# Resume link extractor
node extract-links.js
```

Both automatically skip completed work.

### Troubleshooting

**"No space left on device"**
- Need 35GB+ free
- Check: `df -h`

**"go: command not found"**
- Install Go: `sudo apt install golang-go`

**"node: command not found"**
- Install Node: `curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash - && sudo apt install -y nodejs`

**Crawler stops with errors**
- Most errors are non-critical (some pages unavailable)
- Just restart - it will resume automatically

**Out of memory (link extractor)**
- Edit `extract-links.js`
- Change `const CONCURRENCY = 20;` to `10`

### When Complete

You'll have a complete offline mirror of fedlex.admin.ch!

**Compress for transfer:**
```bash
# Compress mirror (takes ~30 mins, reduces to ~10-15GB)
tar -czf fedlex-mirror.tar.gz mirror/

# Compress everything
tar -czf fedlex-complete.tar.gz mirror/ edges.tsv urls.txt
```

**Verify completeness:**
```bash
# Should be ~377,158
wc -l progress.txt

# Should be close to progress.txt count
wc -l links-progress.txt

# Should be millions
wc -l edges.tsv
```

### Questions?

Read the detailed **README.md** for:
- Performance tuning
- Graph analysis examples
- Detailed troubleshooting
- Data formats

---

**Expected Timeline:**
- Setup: 5 minutes
- Main crawler: 1-2 hours
- Link extractor: 2-3 hours
- **Total: 3-5 hours**

Good luck! 🚀
