const { test, expect } = require('@playwright/test');

test('simple page test', async ({ page }) => {
  console.log('Navigating to base URL...');
  await page.goto('http://localhost:5174/');
  
  console.log('Waiting for page to load...');
  await page.waitForLoadState('networkidle');
  
  const title = await page.title();
  console.log('Page title:', title);
  
  const content = await page.textContent('body');
  console.log('Body content length:', content ? content.length : 'null');
  console.log('First 500 chars of body:', content ? content.substring(0, 500) : 'null');
  
  // Check for any errors in console
  page.on('console', msg => console.log('CONSOLE:', msg.text()));
  page.on('pageerror', err => console.log('PAGE ERROR:', err.message));
  
  await page.screenshot({ path: 'debug-screenshot.png', fullPage: true });
});
