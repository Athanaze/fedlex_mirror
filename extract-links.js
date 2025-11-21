const { chromium } = require('playwright');
const fs = require('fs');
const path = require('path');
const readline = require('readline');

const EDGES_FILE = 'edges.tsv';
const PROGRESS_FILE = 'links-progress.txt';
const URLS_FILE = 'progress.txt';
const MIRROR_DIR = 'mirror';
const CONCURRENCY = 20;

function urlToPath(url) {
  let urlPath = url.replace(/^https?:\/\//, '');
  urlPath = urlPath.replace(/\?/g, '_').replace(/&/g, '_');

  if (urlPath.endsWith('/') || !path.basename(urlPath).includes('.')) {
    urlPath = path.join(urlPath, 'index.html');
  }
  return path.join(MIRROR_DIR, urlPath);
}

async function main() {
  // Load completed URLs
  const completed = new Set();
  if (fs.existsSync(PROGRESS_FILE)) {
    const data = fs.readFileSync(PROGRESS_FILE, 'utf-8');
    data.split('\n').filter(Boolean).forEach(url => completed.add(url));
    console.log(`Loaded ${completed.size} already-processed URLs`);
  }

  // Load URLs to process
  const urls = [];
  const rl = readline.createInterface({
    input: fs.createReadStream(URLS_FILE),
    crlfDelay: Infinity
  });
  for await (const line of rl) {
    if (line && !completed.has(line)) {
      const filePath = urlToPath(line);
      if (fs.existsSync(filePath)) {
        urls.push({ url: line, filePath });
      }
    }
  }
  console.log(`URLs to process: ${urls.length}`);

  if (urls.length === 0) {
    console.log('No URLs to process!');
    return;
  }

  // Open files for appending
  const edgesFd = fs.openSync(EDGES_FILE, 'a');
  const progressFd = fs.openSync(PROGRESS_FILE, 'a');

  // Launch browser ONCE and reuse
  const browser = await chromium.launch({ headless: true });
  const context = await browser.newContext();

  let processed = 0;
  let edgeCount = 0;
  const startTime = Date.now();

  // Process in batches with concurrent pages (tabs)
  for (let i = 0; i < urls.length; i += CONCURRENCY) {
    const batch = urls.slice(i, i + CONCURRENCY);

    await Promise.all(batch.map(async ({ url, filePath }) => {
      const page = await context.newPage();

      try {
        // Load LOCAL file instead of visiting URL
        const fileUrl = 'file://' + path.resolve(filePath);
        await page.goto(fileUrl, { timeout: 5000, waitUntil: 'domcontentloaded' });

        // Extract all links
        const links = await page.evaluate((baseUrl) => {
          const anchors = document.querySelectorAll('a[href]');
          return Array.from(anchors).map(a => {
            try {
              // Resolve relative URLs against the original URL (not file://)
              const url = new URL(a.getAttribute('href'), baseUrl);
              return url.href;
            } catch {
              return null;
            }
          }).filter(href =>
            href && href.includes('fedlex.admin.ch') && !href.startsWith('javascript:')
          );
        }, url);

        // Write edges
        const uniqueLinks = [...new Set(links)];
        for (const target of uniqueLinks) {
          if (target !== url) {
            fs.writeSync(edgesFd, `${url}\t${target}\n`);
            edgeCount++;
          }
        }

        // Mark as done
        fs.writeSync(progressFd, url + '\n');
        processed++;

        if (processed % 100 === 0) {
          const elapsed = (Date.now() - startTime) / 1000;
          const rate = processed / elapsed;
          const eta = (urls.length - processed) / rate;
          console.log(`Progress: ${processed}/${urls.length} (${(processed/urls.length*100).toFixed(1)}%) | Edges: ${edgeCount} | ${rate.toFixed(1)}/s | ETA: ${(eta/60).toFixed(0)}m`);
        }
      } catch (err) {
        console.error(`Error ${url}: ${err.message}`);
      } finally {
        await page.close();
      }
    }));
  }

  await context.close();
  await browser.close();
  fs.closeSync(edgesFd);
  fs.closeSync(progressFd);

  const elapsed = (Date.now() - startTime) / 1000;
  console.log(`\nDone! Processed ${processed} pages, found ${edgeCount} edges in ${elapsed.toFixed(0)}s (${(processed/elapsed).toFixed(1)} pages/sec)`);
}

main().catch(console.error);
