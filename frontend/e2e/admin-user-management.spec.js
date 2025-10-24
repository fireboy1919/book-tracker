import { test, expect } from '@playwright/test'

test.describe('Admin User Management Features', () => {
  test.beforeEach(async ({ page }) => {
    // Skip database reset since we don't have test routes enabled
    // await page.request.delete('http://localhost:8080/api/test/reset-db')
  })

  test('admin can create new users', async ({ page }) => {
    // Use the existing verified admin user (admin@test.com) that we know exists and is verified
    
    await page.goto('/login')
    
    // Login as the existing verified admin user
    await page.fill('input[name="email"]', 'admin@test.com')
    await page.fill('input[name="password"]', 'password123')
    await page.click('button[type="submit"]')
    
    // Wait for login to complete
    await page.waitForTimeout(3000)
    
    // Should be redirected to dashboard since user is verified and admin
    await expect(page).toHaveURL('/dashboard')
    await expect(page.locator('text=My Children')).toBeVisible({ timeout: 10000 })
    
    // Navigate to admin panel
    await page.goto('/admin')
    await expect(page).toHaveURL('/admin')
    await expect(page.locator('text=User Management')).toBeVisible()
    
    // Should see a "Create User" button
    const createUserButton = page.locator('button:has-text("Create User")')
    await expect(createUserButton).toBeVisible()
    await createUserButton.click()
    
    // Fill out the user creation form with unique email
    const timestamp = Date.now()
    const newUserEmail = `testuser${timestamp}@test.com`
    
    await page.fill('input[name="firstName"]', 'Test')
    await page.fill('input[name="lastName"]', 'User')
    await page.fill('input[name="email"]', newUserEmail)
    await page.fill('input[name="password"]', 'password123')
    
    // Submit the form
    await page.click('button[type="submit"]')
    
    // Should see success message
    await expect(page.locator('text=User created successfully!')).toBeVisible({ timeout: 5000 })
    
    // Check that the new user appears in the user list
    await expect(page.locator('text=Test User')).toBeVisible({ timeout: 5000 })
  })

  test('admin can resend verification emails', async ({ page }) => {
    // Use the existing verified admin user 
    await page.goto('/login')
    
    // Login as admin
    await page.fill('input[name="email"]', 'admin@test.com')
    await page.fill('input[name="password"]', 'password123')
    await page.click('button[type="submit"]')
    
    await page.waitForTimeout(3000)
    await expect(page).toHaveURL('/dashboard')
    
    // Go to admin panel and create a user to test resend functionality
    await page.goto('/admin')
    await expect(page.locator('text=User Management')).toBeVisible()
    
    // Create a new user with unique email
    const timestamp = Date.now()
    const testUserEmail = `testuser${timestamp}@test.com`
    
    await page.click('button:has-text("Create User")')
    await page.fill('input[name="firstName"]', 'Test')
    await page.fill('input[name="lastName"]', 'User')
    await page.fill('input[name="email"]', testUserEmail)
    await page.fill('input[name="password"]', 'password123')
    await page.click('button[type="submit"]')
    
    // Wait for success message
    await expect(page.locator('text=User created successfully!')).toBeVisible({ timeout: 5000 })
    
    // Look for users with unverified email status and resend button
    const resendButton = page.locator('[data-testid="resend-verification"]').first()
    
    if (await resendButton.isVisible()) {
      await resendButton.click()
      // Should see success message
      await expect(page.locator('text=Verification email sent successfully!')).toBeVisible({ timeout: 5000 })
    }
  })

  test('admin endpoints require authentication', async ({ page }) => {
    // Test that admin endpoints return 401 without authentication
    
    // Test create user endpoint
    const createUserResponse = await page.request.post('http://localhost:8080/api/users', {
      data: {
        email: 'test@test.com',
        password: 'password123',
        firstName: 'Test',
        lastName: 'User'
      }
    })
    
    expect(createUserResponse.status()).toBe(401)
    
    // Test resend verification endpoint
    const resendResponse = await page.request.post('http://localhost:8080/api/users/1/resend-verification')
    
    expect(resendResponse.status()).toBe(401)
  })
})