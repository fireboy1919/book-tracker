import { test, expect } from '@playwright/test';

// Helper function to create a test user and login
async function loginAsTestUser(page) {
  // Navigate to the app
  await page.goto('/');
  
  // Wait for page to load
  await page.waitForLoadState('networkidle');
  
  // If redirected to register, create a test user first
  const currentUrl = page.url();
  if (currentUrl.includes('/register')) {
    await page.fill('input[name="firstName"]', 'Test');
    await page.fill('input[name="lastName"]', 'User'); 
    await page.fill('input[name="email"]', 'test@example.com');
    await page.fill('input[name="password"]', 'password123');
    await page.fill('input[name="confirmPassword"]', 'password123');
    await page.click('button[type="submit"]');
    
    // Wait for redirect to login page or dashboard
    await page.waitForURL(/.*\/(login|dashboard)/, { timeout: 10000 });
  }
  
  // If still not on dashboard, try to login
  if (!page.url().includes('/dashboard')) {
    // Ensure we're on login page
    if (!page.url().includes('/login')) {
      await page.goto('/login');
    }
    
    // Login with test credentials
    await page.fill('input[name="email"]', 'test@example.com');
    await page.fill('input[name="password"]', 'password123');
    await page.click('button[type="submit"]');
    
    // Wait for redirect to dashboard
    await page.waitForURL(/.*\/dashboard/, { timeout: 10000 });
  }
  
  // Verify we're on dashboard
  await expect(page).toHaveURL(/.*\/dashboard/);
}

// Helper function to add a test child
async function addTestChild(page, childName = 'Test Child', grade = '3rd') {
  // Click Add Child button
  await page.click('button:has-text("Add Child")');
  
  // Fill child form
  await page.fill('input[name="firstName"]', childName.split(' ')[0]);
  await page.fill('input[name="lastName"]', childName.split(' ')[1] || 'Child');
  await page.fill('input[name="grade"]', grade);
  
  // Submit form
  await page.click('button[type="submit"]');
  
  // Wait for child to appear
  await expect(page.locator(`text=${childName}`)).toBeVisible();
}

test.describe('Book Functionality Tests', () => {
  
  test.beforeEach(async ({ page }) => {
    // Reset database before each test
    await page.request.delete('http://localhost:8080/api/test/reset-db');
    
    // Login and add a test child
    await loginAsTestUser(page);
    await addTestChild(page);
  });

  test('user can add a book using ISBN lookup', async ({ page }) => {
    // Click "Add Book" for the test child
    await page.click('button:has-text("Add Book")').first();
    
    // Verify modal opened
    await expect(page.locator('text=Add Book for Test Child')).toBeVisible();
    
    // Check that Title and Author fields are editable by default
    const titleField = page.locator('input[name="title"]');
    const authorField = page.locator('input[name="author"]'); 
    await expect(titleField).not.toHaveAttribute('readonly');
    await expect(authorField).not.toHaveAttribute('readonly');
    
    // Enter ISBN
    await page.fill('input[name="isbn"]', '9780439708180');
    
    // Click Lookup button
    await page.click('button:has-text("Lookup")');
    
    // Wait for book details to populate
    await expect(titleField).toHaveValue(/Harry Potter/);
    await expect(authorField).toHaveValue(/Rowling/);
    
    // Verify fields become read-only after lookup
    await expect(titleField).toHaveAttribute('readonly');
    await expect(authorField).toHaveAttribute('readonly');
    
    // Submit the book
    await page.click('button:has-text("Add Book")');
    
    // Verify modal closes and book appears in list
    await expect(page.locator('text=Add Book for Test Child')).not.toBeVisible();
    await expect(page.locator('text=1 books this month')).toBeVisible();
  });

  test('user can add a custom book manually', async ({ page }) => {
    // Click "Add Book" for the test child
    await page.click('button:has-text("Add Book")').first();
    
    // Verify modal opened
    await expect(page.locator('text=Add Book for Test Child')).toBeVisible();
    
    // Enter custom book details directly (without ISBN)
    await page.fill('input[name="title"]', 'The Great Gatsby');
    await page.fill('input[name="author"]', 'F. Scott Fitzgerald');
    
    // Verify fields remain editable (not read-only)
    const titleField = page.locator('input[name="title"]');
    const authorField = page.locator('input[name="author"]');
    await expect(titleField).not.toHaveAttribute('readonly');
    await expect(authorField).not.toHaveAttribute('readonly');
    
    // Submit the custom book
    await page.click('button:has-text("Add Book")');
    
    // Verify modal closes and book appears in list
    await expect(page.locator('text=Add Book for Test Child')).not.toBeVisible();
    await expect(page.locator('text=1 books this month')).toBeVisible();
  });

  test('user can view all books for a child', async ({ page }) => {
    // First add a book
    await page.click('button:has-text("Add Book")').first();
    await page.fill('input[name="title"]', 'Test Book');
    await page.fill('input[name="author"]', 'Test Author');
    await page.click('button:has-text("Add Book")');
    
    // Wait for book count to update
    await expect(page.locator('text=1 books this month')).toBeVisible();
    
    // Click "View All" to open full screen view
    await page.click('button:has-text("View All")').first();
    
    // Verify full screen view opened
    await expect(page.locator('text=Test Child\'s Books')).toBeVisible();
    await expect(page.locator('text=Test Book')).toBeVisible();
    await expect(page.locator('text=by Test Author')).toBeVisible();
  });

  test('user can add multiple books and see proper counts', async ({ page }) => {
    // Add first book (ISBN)
    await page.click('button:has-text("Add Book")').first();
    await page.fill('input[name="isbn"]', '9780439708180');
    await page.click('button:has-text("Lookup")');
    await page.waitForTimeout(1000); // Wait for API response
    await page.click('button:has-text("Add Book")');
    
    // Wait for first book to be added
    await expect(page.locator('text=1 books this month')).toBeVisible();
    
    // Add second book (custom)
    await page.click('button:has-text("Add Book")').first();
    await page.fill('input[name="title"]', 'Custom Book');
    await page.fill('input[name="author"]', 'Custom Author');
    await page.click('button:has-text("Add Book")');
    
    // Verify both books are counted
    await expect(page.locator('text=2 books this month')).toBeVisible();
  });

  test('user can use "Finished Today" button', async ({ page }) => {
    // Click "Add Book" for the test child
    await page.click('button:has-text("Add Book")').first();
    
    // Enter book details
    await page.fill('input[name="title"]', 'Quick Read');
    await page.fill('input[name="author"]', 'Fast Author');
    
    // Use "Finished Today" button instead of "Add Book"
    await page.click('button:has-text("Finished Today")');
    
    // Verify book was added with today's date
    await expect(page.locator('text=1 books this month')).toBeVisible();
  });

  test('ISBN lookup shows proper validation states', async ({ page }) => {
    // Click "Add Book" for the test child
    await page.click('button:has-text("Add Book")').first();
    
    // Verify Lookup button is disabled initially
    await expect(page.locator('button:has-text("Lookup")')).toBeDisabled();
    
    // Enter ISBN
    await page.fill('input[name="isbn"]', '9780439708180');
    
    // Verify Lookup button is now enabled
    await expect(page.locator('button:has-text("Lookup")')).not.toBeDisabled();
    
    // Test with invalid ISBN
    await page.fill('input[name="isbn"]', '1234567890');
    await page.click('button:has-text("Lookup")');
    
    // Should show error message
    await expect(page.locator('text=Book not found')).toBeVisible();
  });

  test('form validation works correctly', async ({ page }) => {
    // Click "Add Book" for the test child
    await page.click('button:has-text("Add Book")').first();
    
    // Verify "Add Book" button is disabled initially (no title/author)
    await expect(page.locator('form button:has-text("Add Book")')).toBeDisabled();
    
    // Fill only title
    await page.fill('input[name="title"]', 'Test Title');
    await expect(page.locator('form button:has-text("Add Book")')).toBeDisabled();
    
    // Fill author as well
    await page.fill('input[name="author"]', 'Test Author');
    await expect(page.locator('form button:has-text("Add Book")')).not.toBeDisabled();
  });

  test('user can navigate between months in full view', async ({ page }) => {
    // Add a book
    await page.click('button:has-text("Add Book")').first();
    await page.fill('input[name="title"]', 'Month Test Book');
    await page.fill('input[name="author"]', 'Month Author');
    await page.click('button:has-text("Add Book")');
    
    // Open full view
    await page.click('button:has-text("View All")').first();
    
    // Check current month shows the book
    await expect(page.locator('text=Month Test Book')).toBeVisible();
    
    // Navigate to previous month
    await page.click('button').first(); // Left arrow
    
    // Should show no books for previous month
    await expect(page.locator('text=0 books read')).toBeVisible();
    
    // Navigate back to current month
    await page.click('button').nth(1); // Right arrow
    
    // Should show the book again
    await expect(page.locator('text=Month Test Book')).toBeVisible();
  });

  test('duplicate book detection works', async ({ page }) => {
    // Add first book
    await page.click('button:has-text("Add Book")').first();
    await page.fill('input[name="title"]', 'Duplicate Test');
    await page.fill('input[name="author"]', 'Duplicate Author');
    await page.click('button:has-text("Add Book")');
    
    // Try to add the same book again
    await page.click('button:has-text("Add Book")').first();
    await page.fill('input[name="title"]', 'Duplicate Test');
    await page.fill('input[name="author"]', 'Duplicate Author');
    
    // Should show duplicate warning
    await expect(page.locator('text=Duplicate Book Detected')).toBeVisible();
    
    // Add Book button should be disabled
    await expect(page.locator('form button:has-text("Add Book")')).toBeDisabled();
  });

  test('modal z-index displays correctly above full view', async ({ page }) => {
    // Open full view first
    await page.click('button:has-text("View All")').first();
    
    // Click "Add First Book" from within the full view
    await page.click('button:has-text("Add First Book")');
    
    // Verify modal appears above the full view (not hidden behind it)
    await expect(page.locator('text=Add Book for Test Child')).toBeVisible();
    
    // Verify we can interact with the modal
    await page.fill('input[name="title"]', 'Z-Index Test');
    await expect(page.locator('input[name="title"]')).toHaveValue('Z-Index Test');
  });

  test('user can close modal without adding book', async ({ page }) => {
    // Open modal
    await page.click('button:has-text("Add Book")').first();
    await expect(page.locator('text=Add Book for Test Child')).toBeVisible();
    
    // Close with Cancel button
    await page.click('button:has-text("Cancel")');
    await expect(page.locator('text=Add Book for Test Child')).not.toBeVisible();
    
    // Open modal again
    await page.click('button:has-text("Add Book")').first();
    
    // Close with X button
    await page.click('.text-gray-400'); // X button
    await expect(page.locator('text=Add Book for Test Child')).not.toBeVisible();
  });

});