const { test, expect } = require('@playwright/test')

test('debug admin navigation', async ({ page }) => {
  // Reset database
  await fetch('http://localhost:8080/api/test/reset-db', { method: 'DELETE' })
  
  // Register admin
  await page.goto('http://localhost:5173/register')
  await page.fill('input[name="firstName"]', 'Admin')
  await page.fill('input[name="lastName"]', 'User')
  await page.fill('input[name="email"]', 'admin@example.com')
  await page.fill('input[name="password"]', 'password123')
  await page.fill('input[name="confirmPassword"]', 'password123')
  await page.check('input[name="isAdmin"]')
  await page.click('button[type="submit"]')
  
  // Wait for redirect to login
  await expect(page).toHaveURL('/login')
  
  // Login as admin
  await page.fill('input[name="email"]', 'admin@example.com')
  await page.fill('input[name="password"]', 'password123')
  await page.click('button[type="submit"]')
  
  // Wait for dashboard
  await expect(page).toHaveURL('/dashboard')
  await page.waitForLoadState('networkidle')
  
  // Check localStorage
  const userData = await page.evaluate(() => {
    return {
      user: JSON.parse(localStorage.getItem('user') || '{}'),
      token: localStorage.getItem('token')
    }
  })
  console.log('User data in localStorage:', userData)
  
  // Check if admin panel is in DOM
  const adminPanelElements = await page.locator('text=Admin Panel').count()
  console.log('Admin Panel elements found:', adminPanelElements)
  
  // Check visibility
  const isVisible = await page.locator('text=Admin Panel').first().isVisible()
  console.log('First Admin Panel visible:', isVisible)
  
  // Take a screenshot
  await page.screenshot({ path: 'debug-admin-screenshot.png', fullPage: true })
})