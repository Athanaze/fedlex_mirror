#!/bin/bash

set -e

echo "================================================"
echo "Fedlex Scraper - Setup Script"
echo "================================================"
echo ""

# Check Go
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed"
    echo "Install with: sudo apt install golang-go"
    exit 1
else
    echo "✓ Go found: $(go version)"
fi

# Check Node.js
if ! command -v node &> /dev/null; then
    echo "❌ Node.js is not installed"
    echo "Install with: curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash - && sudo apt install -y nodejs"
    exit 1
else
    echo "✓ Node.js found: $(node --version)"
fi

# Check disk space
available=$(df -BG . | tail -1 | awk '{print $4}' | sed 's/G//')
if [ "$available" -lt 35 ]; then
    echo "⚠️  Warning: Only ${available}GB disk space available"
    echo "   Recommended: 35GB+ for complete scrape"
    read -p "Continue anyway? (y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
else
    echo "✓ Disk space: ${available}GB available"
fi

echo ""
echo "Installing dependencies..."
echo ""

# Go dependencies
echo "→ Installing Go dependencies..."
if [ ! -f "go.mod" ]; then
    go mod init fedlex-crawler
fi
go get github.com/gocolly/colly/v2
echo "✓ Go dependencies installed"

# Node dependencies
echo "→ Installing Node.js dependencies..."
npm install
echo "✓ Node.js dependencies installed"

# Playwright browsers
echo "→ Installing Playwright Chromium..."
npx playwright install chromium
echo "✓ Chromium installed"

echo ""
echo "================================================"
echo "Setup Complete!"
echo "================================================"
echo ""
echo "To start crawling:"
echo "  1. Main crawler:     go run main.go"
echo "  2. Link extractor:   node extract-links.js"
echo ""
echo "To monitor progress:"
echo "  watch -n 5 'wc -l progress.txt links-progress.txt edges.tsv && du -sh mirror/'"
echo ""
echo "See README.md for detailed documentation"
echo ""
