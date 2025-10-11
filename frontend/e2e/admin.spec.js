import { test, expect } from '@playwright/test'

test.describe('Admin Features', () => {
  test.beforeEach(async ({ page }) => {
    // Reset database state
    await page.request.delete('http://localhost:8080/api/test/reset-db')
  })

  test('admin can manage users', async ({ page }) => {
    // Register admin user
    await page.goto('/register')
    await page.fill('input[name="firstName"]', 'Admin')
    await page.fill('input[name="lastName"]', 'User')
    await page.fill('input[name="email"]', 'admin@example.com')
    await page.fill('input[name="password"]', 'password123')
    await page.fill('input[name="confirmPassword"]', 'password123')
    
    // Check admin checkbox
    await page.check('input[name="isAdmin"]')
    await page.click('button[type="submit"]')
    
    // Wait for redirect to login page
    await expect(page).toHaveURL('/login')
    
    // Login as admin
    await page.fill('input[name="email"]', 'admin@example.com')
    await page.fill('input[name="password"]', 'password123')
    await page.click('button[type="submit"]')
    
    // Wait for redirect to dashboard
    await expect(page).toHaveURL('/dashboard')
    
    // Wait for network requests to complete
    await page.waitForLoadState('networkidle')
    
    // Wait for dashboard to load
    await expect(page.locator('text=My Children')).toBeVisible({ timeout: 10000 })
    
    // Navigate directly to admin panel (admin user should have access)
    await page.goto('/admin')
    await expect(page).toHaveURL('/admin')
    await expect(page.locator('text=User Management')).toBeVisible()
    
    // Should see admin user in the list (use more specific selectors)
    await expect(page.locator('table').getByText('Admin User')).toBeVisible()
    await expect(page.locator('table').getByText('admin@example.com')).toBeVisible()
    await expect(page.locator('table').getByText('Admin').first()).toBeVisible()
  })

  test('regular user cannot access admin panel', async ({ page }) => {
    // Register regular user
    await page.goto('/register')
    await page.fill('input[name="firstName"]', 'Regular')
    await page.fill('input[name="lastName"]', 'User')
    await page.fill('input[name="email"]', 'user@example.com')
    await page.fill('input[name="password"]', 'password123')
    await page.fill('input[name="confirmPassword"]', 'password123')
    await page.click('button[type="submit"]')
    
    // Wait for redirect to login page
    await expect(page).toHaveURL('/login')
    
    // Login as regular user
    await page.fill('input[name="email"]', 'user@example.com')
    await page.fill('input[name="password"]', 'password123')
    await page.click('button[type="submit"]')
    
    // Wait for redirect to dashboard
    await expect(page).toHaveURL('/dashboard')
    
    // Wait for network requests to complete
    await page.waitForLoadState('networkidle')
    
    // Wait for dashboard to load
    await expect(page.locator('text=My Children')).toBeVisible({ timeout: 10000 })
    
    // Should NOT see admin panel link
    await expect(page.locator('text=Admin Panel')).not.toBeVisible()
    
    // Try to access admin panel directly
    await page.goto('/admin')
    
    // Should redirect to dashboard
    await expect(page).toHaveURL('/dashboard')
  })

  test('admin can promote and demote users', async ({ page }) => {
    // Register admin
    await page.goto('/register')
    await page.fill('input[name="firstName"]', 'Admin')
    await page.fill('input[name="lastName"]', 'User')
    await page.fill('input[name="email"]', 'admin@example.com')
    await page.fill('input[name="password"]', 'password123')
    await page.fill('input[name="confirmPassword"]', 'password123')
    await page.check('input[name="isAdmin"]')
    await page.click('button[type="submit"]')
    
    // Wait for redirect to login page
    await expect(page).toHaveURL('/login')
    
    // Register regular user
    await page.goto('/register')
    await page.fill('input[name="firstName"]', 'Regular')
    await page.fill('input[name="lastName"]', 'User')
    await page.fill('input[name="email"]', 'user@example.com')
    await page.fill('input[name="password"]', 'password123')
    await page.fill('input[name="confirmPassword"]', 'password123')
    await page.click('button[type="submit"]')
    
    // Wait for redirect to login page
    await expect(page).toHaveURL('/login')
    
    // Login as admin
    await page.fill('input[name="email"]', 'admin@example.com')
    await page.fill('input[name="password"]', 'password123')
    await page.click('button[type="submit"]')
    
    // Wait for redirect to dashboard
    await expect(page).toHaveURL('/dashboard')
    
    // Wait for network requests to complete
    await page.waitForLoadState('networkidle')
    
    // Wait for dashboard to load
    await expect(page.locator('text=My Children')).toBeVisible({ timeout: 10000 })
    
    // Set mobile viewport to ensure hamburger menu is visible
    await page.setViewportSize({ width: 375, height: 667 })
    
    // Go to admin panel - open mobile menu first, then click Admin Panel
    await page.locator('nav button:not([title])').click() // Open mobile menu (hamburger icon - button without title)
    await page.getByRole('link', { name: 'Admin Panel' }).click()
    
    // Should see both users in the table
    await expect(page.getByRole('table').getByText('Admin User')).toBeVisible()
    await expect(page.getByRole('table').getByText('Regular User')).toBeVisible()
    
    // Find regular user row and promote to admin
    const regularUserRow = page.locator('tr:has-text("Regular User")')
    const roleButton = regularUserRow.locator('button:has-text("User")')
    await roleButton.click()
    
    // Should now show as Admin
    await expect(regularUserRow.locator('button:has-text("Admin")')).toBeVisible()
    
    // Demote back to user
    await regularUserRow.locator('button:has-text("Admin")').click()
    await expect(regularUserRow.locator('button:has-text("User")')).toBeVisible()
  })

  test('admin can delete users', async ({ page }) => {
    // Register admin
    await page.goto('/register')
    await page.fill('input[name="firstName"]', 'Admin')
    await page.fill('input[name="lastName"]', 'User')
    await page.fill('input[name="email"]', 'admin@example.com')
    await page.fill('input[name="password"]', 'password123')
    await page.fill('input[name="confirmPassword"]', 'password123')
    await page.check('input[name="isAdmin"]')
    await page.click('button[type="submit"]')
    
    // Wait for redirect to login page
    await expect(page).toHaveURL('/login')
    
    // Register user to delete
    await page.goto('/register')
    await page.fill('input[name="firstName"]', 'To Delete')
    await page.fill('input[name="lastName"]', 'User')
    await page.fill('input[name="email"]', 'delete@example.com')
    await page.fill('input[name="password"]', 'password123')
    await page.fill('input[name="confirmPassword"]', 'password123')
    await page.click('button[type="submit"]')
    
    // Wait for redirect to login page
    await expect(page).toHaveURL('/login')
    
    // Login as admin
    await page.fill('input[name="email"]', 'admin@example.com')
    await page.fill('input[name="password"]', 'password123')
    await page.click('button[type="submit"]')
    
    // Wait for redirect to dashboard
    await expect(page).toHaveURL('/dashboard')
    
    // Wait for network requests to complete
    await page.waitForLoadState('networkidle')
    
    // Wait for dashboard to load
    await expect(page.locator('text=My Children')).toBeVisible({ timeout: 10000 })
    
    // Go to admin panel - try visible link first, fallback to direct navigation
    try {
      await page.click('text=Admin Panel', { timeout: 5000 })
    } catch {
      await page.goto('/admin')
    }
    
    // Should see user to delete
    await expect(page.locator('text=To Delete User')).toBeVisible()
    
    // Delete user
    const userRow = page.locator('tr:has-text("To Delete User")')
    
    // Mock the confirm dialog to return true
    page.on('dialog', dialog => dialog.accept())
    
    await userRow.locator('button').last().click() // Delete button
    
    // User should be removed
    await expect(page.locator('text=To Delete User')).not.toBeVisible()
  })
})