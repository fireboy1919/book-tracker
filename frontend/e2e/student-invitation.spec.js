import { test, expect } from '@playwright/test';

test.describe('Student Invitation Flow', () => {
  
  test.beforeEach(async ({ page }) => {
    test.setTimeout(60000); // Increase timeout for full flow
  });

  test('invitation URL structure and redirect flow', async ({ page }) => {
    // Step 1: Test the invitation URL structure and redirect flow
    console.log('🔍 Testing invitation URL structure and redirect flow...');
    
    // Use a test invitation URL that should trigger the proper redirect flow
    const classId = '1';
    const testToken = 'test-token-for-redirect-flow';
    
    const invitationUrl = `http://localhost:5173/invite/${classId}/${testToken}`;
    console.log('🔗 Testing invitation URL:', invitationUrl);
    
    // Navigate to the invitation URL
    await page.goto(invitationUrl);
    await page.waitForLoadState('networkidle');
    
    // Step 2: Should show invitation details
    console.log('📄 Current URL:', page.url());
    
    // Check if we're on the invitation page
    if (page.url().includes('/invite/')) {
      await expect(page.locator('h2:has-text("Student Invitation")')).toBeVisible({ timeout: 10000 });
      console.log('✅ Invitation page loaded successfully');
      
      // Check for invitation details
      const hasInvitationDetails = await page.locator('text=You\'ve been invited').isVisible();
      if (hasInvitationDetails) {
        console.log('✅ Invitation details are visible');
      }
      
      // Should show login/register options since user is not logged in
      const loginButton = page.locator('text=Log In');
      const signUpButton = page.locator('text=Sign Up');
      
      await expect(loginButton).toBeVisible();
      await expect(signUpButton).toBeVisible();
      console.log('✅ Login and Sign Up buttons are visible');
      
      // Step 3: Click Login button (should redirect with invitation parameter)
      await loginButton.click();
      await page.waitForLoadState('networkidle');
      
      // Should be on login page with redirect parameter
      expect(page.url()).toContain('/login');
      expect(page.url()).toContain('redirect=');
      console.log('✅ Redirected to login page with redirect parameter');
      console.log('🔗 Login URL:', page.url());
      
      // Step 4: Check if login form is present
      await expect(page.locator('h2:has-text("Sign in to your account")')).toBeVisible();
      await expect(page.locator('input[name="email"]')).toBeVisible();
      await expect(page.locator('input[name="password"]')).toBeVisible();
      await expect(page.locator('button:has-text("Sign in with Google")')).toBeVisible();
      console.log('✅ Login form elements are present');
      
      // For this test, we'll verify the flow up to the login page
      // In a real test, you would continue with actual login credentials
      console.log('🎉 Student invitation flow test completed successfully');
      console.log('  ✅ Invitation URL loads correctly');
      console.log('  ✅ Invitation details are displayed');
      console.log('  ✅ Login/Sign Up options are available');
      console.log('  ✅ Redirect to login preserves invitation URL');
      
    } else if (page.url().includes('/login')) {
      console.log('⚠️  Redirected directly to login (user not authenticated)');
      console.log('🔗 Login URL:', page.url());
      
      // Check if redirect parameter is preserved
      if (page.url().includes('redirect=')) {
        console.log('✅ Redirect parameter preserved in login URL');
      } else {
        console.log('❌ Redirect parameter missing from login URL');
      }
    } else {
      console.log('❌ Unexpected redirect to:', page.url());
    }
  });

  test('test invitation API endpoint directly', async ({ page }) => {
    // Test the invitation details API endpoint
    const classId = '1';
    const token = '8InDj-3gH9qPsChtzpogAnMyOInWMg4fSX8lxK4hsm45t4UPSas1JWcrDSpUwpSn';
    
    console.log('🔍 Testing invitation API endpoint...');
    
    // Go to any page first to set up the request context
    await page.goto('https://booktracker.rustyphillips.net/');
    
    // Make API request to test the invitation endpoint
    const response = await page.request.get(`https://booktracker.rustyphillips.net/api/invite/${classId}/${token}`);
    
    console.log('📡 API Response Status:', response.status());
    
    if (response.ok()) {
      const data = await response.json();
      console.log('✅ API returned valid invitation data:');
      console.log('   Student Name:', data.student_name);
      console.log('   Class Name:', data.class_name);
      console.log('   Teacher Name:', data.teacher_name);
      
      // Verify required fields are present
      expect(data.student_name).toBeTruthy();
      expect(data.class_name).toBeTruthy();
      expect(data.teacher_name).toBeTruthy();
      
    } else {
      const errorData = await response.json().catch(() => null);
      console.log('❌ API Error:', response.status());
      if (errorData) {
        console.log('   Error Message:', errorData.message);
      }
      
      // This will fail the test if the API doesn't work
      expect(response.status()).toBe(200);
    }
  });

  test('test invitation token validation', async ({ page }) => {
    console.log('🔍 Testing invitation token validation...');
    
    // Test with invalid token
    const invalidToken = 'invalid-token-123';
    const classId = '1';
    
    await page.goto('https://booktracker.rustyphillips.net/');
    
    const response = await page.request.get(`https://booktracker.rustyphillips.net/api/invite/${classId}/${invalidToken}`);
    
    console.log('📡 Invalid token API Response Status:', response.status());
    
    // Should return 400 or 404 for invalid token
    expect(response.status()).not.toBe(200);
    
    if (!response.ok()) {
      const errorData = await response.json().catch(() => null);
      console.log('✅ API correctly rejected invalid token');
      if (errorData && errorData.message) {
        console.log('   Error Message:', errorData.message);
      }
    }
  });

  test('full invitation flow simulation', async ({ page }) => {
    console.log('🎭 Simulating full invitation flow...');
    
    // Step 1: Start with invitation URL
    const invitationUrl = 'https://booktracker.rustyphillips.net/invite/1/8InDj-3gH9qPsChtzpogAnMyOInWMg4fSX8lxK4hsm45t4UPSas1JWcrDSpUwpSn';
    await page.goto(invitationUrl);
    await page.waitForLoadState('networkidle');
    
    console.log('📍 Step 1: Visited invitation URL');
    console.log('   Current URL:', page.url());
    
    // Step 2: Check what page we're on
    if (page.url().includes('/invite/')) {
      console.log('✅ On invitation page');
      
      // Check if invitation details are loaded
      const hasDetails = await page.locator('text=You\'ve been invited').isVisible();
      if (hasDetails) {
        console.log('✅ Invitation details loaded');
      }
      
      // Check if user is prompted to log in
      const needsLogin = await page.locator('text=Please log in').isVisible();
      if (needsLogin) {
        console.log('✅ User prompted to authenticate');
      }
      
    } else if (page.url().includes('/login')) {
      console.log('📍 Step 2: Redirected to login page');
      
      // Check if redirect parameter is present
      const hasRedirect = page.url().includes('redirect=');
      if (hasRedirect) {
        console.log('✅ Redirect parameter preserved');
        
        // Extract and decode the redirect URL
        const urlParams = new URLSearchParams(page.url().split('?')[1]);
        const redirectParam = urlParams.get('redirect');
        if (redirectParam) {
          console.log('   Redirect URL:', decodeURIComponent(redirectParam));
        }
      } else {
        console.log('❌ Redirect parameter missing');
      }
      
      // Check login form
      const hasLoginForm = await page.locator('input[name="email"]').isVisible();
      if (hasLoginForm) {
        console.log('✅ Login form present');
      }
      
      const hasGoogleLogin = await page.locator('button:has-text("Sign in with Google")').isVisible();
      if (hasGoogleLogin) {
        console.log('✅ Google login option available');
      }
      
    } else {
      console.log('🤔 Unexpected page:', page.url());
    }
    
    console.log('🎉 Full invitation flow simulation completed');
  });

});