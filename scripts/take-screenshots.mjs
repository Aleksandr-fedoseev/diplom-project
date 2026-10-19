import puppeteer from 'puppeteer-core';
import fs from 'fs';
import path from 'path';

const CHROME_PATH = 'C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe';
const EDGE_PATH = 'C:\\Program Files (x86)\\Microsoft\\Edge\\Application\\msedge.exe';

const executablePath = fs.existsSync(CHROME_PATH) ? CHROME_PATH : EDGE_PATH;

const OUTPUT_DIR = path.resolve('screenshots_for_client');
if (fs.existsSync(OUTPUT_DIR)) {
  const existingFiles = fs.readdirSync(OUTPUT_DIR);
  for (const file of existingFiles) {
    fs.unlinkSync(path.join(OUTPUT_DIR, file));
  }
} else {
  fs.mkdirSync(OUTPUT_DIR, { recursive: true });
}

const PAGES = [
  { name: '01_main_page.png', url: 'http://localhost:3000/', fullPage: true },
  { name: '02_projects_catalog.png', url: 'http://localhost:3000/projects', fullPage: true },
  { name: '03_interactive_map.png', url: 'http://localhost:3000/map', fullPage: false },
  { name: '04_map_object_card.png', url: 'http://localhost:3000/map?object=birzha', fullPage: false },
  { name: '05_wooden_museum.png', url: 'http://localhost:3000/projects/wooden-museum', fullPage: true },
  { name: '06_stories_catalog.png', url: 'http://localhost:3000/stories', fullPage: true },
  { name: '07_story_detail.png', url: 'http://localhost:3000/stories/nikolskaya-ulitsa-i-kosa', fullPage: true },
  { name: '08_routes_catalog.png', url: 'http://localhost:3000/routes', fullPage: true },
  { name: '09_shop_catalog.png', url: 'http://localhost:3000/shop', fullPage: true },
];

async function capture() {
  console.log('Launching browser from:', executablePath);
  const browser = await puppeteer.launch({
    executablePath,
    headless: true,
    args: ['--no-sandbox', '--disable-setuid-sandbox', '--window-size=1440,900'],
  });

  const page = await browser.newPage();
  await page.setViewport({ width: 1440, height: 900, deviceScaleFactor: 1.5 });

  for (const item of PAGES) {
    const filePath = path.join(OUTPUT_DIR, item.name);
    console.log(`Capturing: ${item.name} (${item.url})...`);
    try {
      await page.goto(item.url, { waitUntil: 'networkidle0', timeout: 30000 });
      // Пауза для плавной отрисовки шрифтов, изображений и слоев карты
      await new Promise(r => setTimeout(r, 1800));

      if (item.name === '04_map_object_card.png') {
        try {
          await page.waitForSelector('.cursor-pointer', { timeout: 5000 });
          await page.click('.cursor-pointer');
          await new Promise(r => setTimeout(r, 1200));
        } catch (e) {
          console.log('Sidebar click fallback:', e.message);
        }
      }

      await page.screenshot({ path: filePath, fullPage: item.fullPage });
      console.log(`Saved: ${filePath}`);
    } catch (err) {
      console.error(`Error capturing ${item.name}:`, err.message);
    }
  }

  await browser.close();
  console.log('\nAll screenshots captured successfully in:', OUTPUT_DIR);
}

capture().catch(err => {
  console.error('Fatal error:', err);
  process.exit(1);
});
