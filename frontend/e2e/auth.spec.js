import { test, expect } from '@playwright/test'

test.describe('Authentication Flow', () => {
  test.beforeEach(async ({ page }) => {
    // Reset database state before each test
    await page.request.delete('http://localhost:8080/api/test/reset-db')
  })

  test('user can register and login', async ({ page }) => {
    // Go to register page
    await page.goto('/register')
    
    // Fill registration form
    await page.fill('input[name="firstName"]', 'Test')
    await page.fill('input[name="lastName"]', 'User')
    await page.fill('input[name="email"]', 'test@example.com')
    await page.fill('input[name="password"]', 'password123')
    await page.fill('input[name="confirmPassword"]', 'password123')
    
    // Submit registration
    await page.click('button[type="submit"]')
    
    // Should redirect to login page
    await expect(page).toHaveURL('/login')
    await expect(page.locator('text=create a new account')).toBeVisible()
    
    // Fill login form
    await page.fill('input[name="email"]', 'test@example.com')
    await page.fill('input[name="password"]', 'password123')
    
    // Submit login
    await page.click('button[type="submit"]')
    
    // Should redirect to dashboard
    await expect(page).toHaveURL('/dashboard')
    await expect(page.locator('text=My Children')).toBeVisible()
  })

  test('user cannot login with wrong credentials', async ({ page }) => {
    // First register a user
    await page.goto('/register')
    await page.fill('input[name="firstName"]', 'Test')
    await page.fill('input[name="lastName"]', 'User')
    await page.fill('input[name="email"]', 'test@example.com')
    await page.fill('input[name="password"]', 'password123')
    await page.fill('input[name="confirmPassword"]', 'password123')
    await page.click('button[type="submit"]')
    
    // Wait for redirect to login page
    await expect(page).toHaveURL('/login')
    
    // Try to login with wrong password
    await page.fill('input[name="email"]', 'test@example.com')
    await page.fill('input[name="password"]', 'wrongpassword')
    await page.click('button[type="submit"]')
    
    // Should show error and stay on login page
    await expect(page.locator('text=Invalid credentials')).toBeVisible()
    await expect(page).toHaveURL('/login')
  })

  test('protected routes redirect to login', async ({ page }) => {
    // Try to access dashboard without login
    await page.goto('/dashboard')
    
    // Should redirect to login
    await expect(page).toHaveURL('/login')
  })

  test('user can logout', async ({ page }) => {
    // Register and login first
    await page.goto('/register')
    await page.fill('input[name="firstName"]', 'Test')
    await page.fill('input[name="lastName"]', 'User')
    await page.fill('input[name="email"]', 'test@example.com')
    await page.fill('input[name="password"]', 'password123')
    await page.fill('input[name="confirmPassword"]', 'password123')
    await page.click('button[type="submit"]')
    
    await page.fill('input[name="email"]', 'test@example.com')
    await page.fill('input[name="password"]', 'password123')
    await page.click('button[type="submit"]')
    
    // Should be on dashboard
    await expect(page).toHaveURL('/dashboard')
    
    // Click logout button
    await page.click('[title="Logout"]')
    
    // Should redirect to login
    await expect(page).toHaveURL('/login')
  })
})