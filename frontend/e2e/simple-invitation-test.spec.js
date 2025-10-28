import { test, expect } from '@playwright/test';

test.describe('Simple Invitation Test', () => {
  
  test('debug invitation page content', async ({ page }) => {
    test.setTimeout(30000);
    
    console.log('🔍 Debugging invitation page...');
    
    // Navigate to invitation URL
    const invitationUrl = 'http://localhost:5173/invite/1/test-token';
    console.log('🔗 Going to:', invitationUrl);
    
    await page.goto(invitationUrl);
    await page.waitForLoadState('networkidle');
    
    console.log('📄 Final URL:', page.url());
    
    // Get page title
    const title = await page.title();
    console.log('📝 Page title:', title);
    
    // Get all text content
    const bodyText = await page.locator('body').textContent();
    console.log('📖 Page text:', bodyText.substring(0, 500) + '...');
    
    // Check for common elements
    const h1Elements = await page.locator('h1').count();
    const h2Elements = await page.locator('h2').count();
    const buttonElements = await page.locator('button').count();
    const linkElements = await page.locator('a').count();
    
    console.log(`📊 Page elements: ${h1Elements} h1, ${h2Elements} h2, ${buttonElements} buttons, ${linkElements} links`);
    
    // Get all h1 and h2 text
    if (h1Elements > 0) {
      const h1Texts = await page.locator('h1').allTextContents();
      console.log('📍 H1 texts:', h1Texts);
    }
    
    if (h2Elements > 0) {
      const h2Texts = await page.locator('h2').allTextContents();
      console.log('📍 H2 texts:', h2Texts);
    }
    
    // Check if we're on login page
    const isLoginPage = await page.locator('input[name="email"]').isVisible();
    if (isLoginPage) {
      console.log('🔐 Page appears to be a login page');
    }
    
    // Check if we're on invitation page
    const isInvitationPage = await page.locator('text=invitation').isVisible();
    if (isInvitationPage) {
      console.log('📨 Page contains invitation-related text');
    }
    
    // Check for error messages
    const hasError = await page.locator('text=error').isVisible();
    if (hasError) {
      console.log('❌ Page contains error text');
    }
    
    // Take a screenshot for debugging
    await page.screenshot({ path: 'debug-invitation-page.png', fullPage: true });
    console.log('📸 Screenshot saved as debug-invitation-page.png');
    
    console.log('✅ Debug completed');
  });
});