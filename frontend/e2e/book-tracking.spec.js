import { test, expect } from '@playwright/test'

test.describe('Book Tracking Flow', () => {
  test.beforeEach(async ({ page }) => {
    // Reset database state
    await page.request.delete('http://localhost:8080/api/test/reset-db')
    
    // Register and login a user
    await page.goto('/register')
    await page.fill('input[name="firstName"]', 'Test')
    await page.fill('input[name="lastName"]', 'User')
    await page.fill('input[name="email"]', 'test@example.com')
    await page.fill('input[name="password"]', 'password123')
    await page.fill('input[name="confirmPassword"]', 'password123')
    await page.click('button[type="submit"]')
    
    // Wait for redirect to login page
    await expect(page).toHaveURL('/login')
    
    await page.fill('input[name="email"]', 'test@example.com')
    await page.fill('input[name="password"]', 'password123')
    await page.click('button[type="submit"]')
    
    await expect(page).toHaveURL('/dashboard')
  })

  test('user can add a child and track books', async ({ page }) => {
    // Should see empty state
    await expect(page.locator('text=No children added yet')).toBeVisible()
    
    // Add first child
    await page.click('text=Add Your First Child')
    
    // Fill child form
    await page.fill('input[name="name"]', 'Alice')
    await page.fill('input[name="age"]', '8')
    await page.click('button[type="submit"]')
    
    // Should see child card
    await expect(page.locator('text=Alice')).toBeVisible()
    await expect(page.locator('text=Age 8')).toBeVisible()
    await expect(page.locator('text=0 books read')).toBeVisible()
    
    // Add a book
    await page.click('text=Add Book')
    
    // Fill book form
    await page.fill('input[name="title"]', 'The Cat in the Hat')
    await page.fill('input[name="author"]', 'Dr. Seuss')
    
    // Date should default to today - verify it's filled
    const dateInput = page.locator('input[name="dateRead"]')
    await expect(dateInput).toHaveValue(/\d{4}-\d{2}-\d{2}/)
    
    await page.click('button[type="submit"]')
    
    // Should see updated book count
    await expect(page.locator('text=1 books read')).toBeVisible()
    await expect(page.locator('text=Recent: The Cat in the Hat')).toBeVisible()
  })

  test('user can add multiple children and books', async ({ page }) => {
    // Add first child
    await page.click('text=Add Your First Child')
    await page.fill('input[name="name"]', 'Alice')
    await page.fill('input[name="age"]', '8')
    await page.click('button[type="submit"]')
    
    // Add second child using the header button
    await page.click('button:has-text("Add Child")')
    await page.fill('input[name="name"]', 'Bob')
    await page.fill('input[name="age"]', '10')
    await page.click('button[type="submit"]')
    
    // Should see both children
    await expect(page.locator('text=Alice')).toBeVisible()
    await expect(page.locator('text=Bob')).toBeVisible()
    
    // Add books to each child
    const aliceCard = page.locator('.grid > div:has-text("Alice")')
    await aliceCard.locator('text=Add Book').click()
    await page.fill('input[name="title"]', 'Green Eggs and Ham')
    await page.fill('input[name="author"]', 'Dr. Seuss')
    await page.click('button[type="submit"]')
    
    const bobCard = page.locator('.grid > div:has-text("Bob")')
    await bobCard.locator('text=Add Book').click()
    await page.fill('input[name="title"]', 'Charlotte\'s Web')
    await page.fill('input[name="author"]', 'E.B. White')
    await page.click('button[type="submit"]')
    
    // Verify book counts
    await expect(aliceCard.locator('text=1 books read')).toBeVisible()
    await expect(bobCard.locator('text=1 books read')).toBeVisible()
  })

  test('user can generate and download a report', async ({ page }) => {
    // Add child and book first
    await page.click('text=Add Your First Child')
    await page.fill('input[name="name"]', 'Alice')
    await page.fill('input[name="age"]', '8')
    await page.click('button[type="submit"]')
    
    await page.click('text=Add Book')
    await page.fill('input[name="title"]', 'The Cat in the Hat')
    await page.fill('input[name="author"]', 'Dr. Seuss')
    await page.click('button[type="submit"]')
    
    // Generate report
    await page.click('text=Generate Report')
    
    // Should see report modal
    await expect(page.locator('text=Reading Report')).toBeVisible()
    await expect(page.locator('text=Alice (Age 8)')).toBeVisible()
    await expect(page.locator('text=1 books read')).toBeVisible()
    await expect(page.locator('text=The Cat in the Hat')).toBeVisible()
    await expect(page.locator('text=Dr. Seuss')).toBeVisible()
    
    // Test download button (we can't actually test file download in browser tests easily)
    await expect(page.locator('text=Download CSV')).toBeVisible()
  })

  test('user can share child data with others', async ({ page }) => {
    // Add child first
    await page.click('text=Add Your First Child')
    await page.fill('input[name="name"]', 'Alice')
    await page.fill('input[name="age"]', '8')
    await page.click('button[type="submit"]')
    
    // Click share button
    await page.click('text=Share')
    
    // Should see share modal
    await expect(page.locator('text=Share Alice')).toBeVisible()
    
    // Fill invitation form
    await page.fill('input[name="email"]', 'friend@example.com')
    await page.selectOption('select[name="permissionType"]', 'EDITOR')
    
    // Send invitation
    await page.click('text=Send Invitation')
    
    // Should see success message
    await expect(page.locator('text=Invitation sent successfully!')).toBeVisible()
  })
})