# Book Tracker - Google Sheets Student Invitation Template

This template allows teachers to bulk-generate invitation links for students to join their Book Tracker reading class. It uses CryptoJS for local encryption - no API calls needed!

## Quick Start Guide

### Step 1: Get Your Teacher Invitation Data

1. Log into Book Tracker as a teacher
2. Go to your class dashboard
3. Click "Get Invitation Data for Google Sheets" button
4. Copy the three values: Teacher ID, Class ID, and Invitation Key

### Step 2: Copy the Book Tracker Template

1. **Open the template**: [Book Tracker Student Invitation Template](https://docs.google.com/spreadsheets/d/YOUR_TEMPLATE_ID/copy) *(Replace with actual template link)*
2. **Click "Make a copy"** - this will create your own version with everything pre-configured
3. The template already has:
   - ✅ CryptoJS library installed
   - ✅ All sheets set up (Config, Students, Email Template)  
   - ✅ Book Tracker menu ready to use
   - ✅ Sample data and instructions

### Step 3: Configure Your Template

1. Click **Book Tracker → Setup Sheets** from the menu
2. Go to the **Config** sheet
3. Paste your values:
   - **B1**: Your Teacher ID
   - **B2**: Your Class ID  
   - **B3**: Your Invitation Key

### Step 4: Add Your Students

1. Go to the **Students** sheet
2. Add your students' information:
   - **Column A**: Student Name (e.g., "John Smith")
   - **Column B**: Grade (e.g., "3rd Grade")
   - **Column C**: Parent Email (for mail merge)

### Step 4.5: Test Your Setup

1. Click **Book Tracker → Test Configuration** from the menu
2. If successful, you'll see a confirmation with sample token
3. If there are errors, follow the troubleshooting steps below

### Step 5: Generate Invitation Tokens

1. Click **Book Tracker → Generate Tokens** from the menu
2. The system will automatically fill columns D and E with:
   - **Column D**: Unique invitation token for each student (encrypted locally!)
   - **Column E**: Full invitation URL

### Step 6: Send Invitations via Mail Merge

1. Click **Book Tracker → Create Email Template** from the menu
2. Install a mail merge add-on like "Yet Another Mail Merge" or "Mail Merge with Attachments"
3. Use the email template from the **Email Template** sheet
4. Set up mail merge using the **Students** sheet as data source:
   - Map `{{STUDENT_NAME}}` to the "Student Name" column
   - Map `{{INVITATION_URL}}` to the "Invitation URL" column
5. Send your personalized invitations!

## How It Works

### Stateless Invitation System
- Each invitation token contains encrypted student information (name, grade, class)
- Tokens are generated using your unique teacher encryption key
- No data is stored on the server - everything is in the encrypted token
- Tokens expire after 30 days for security

### Parent Registration Process
1. Parent clicks the invitation link
2. If not logged in, they create an account or log in
3. System decrypts the token to get student information
4. If student already exists in parent's account, assigns to class
5. If student doesn't exist, creates new student and assigns to class

## Troubleshooting

### "Authorization required" error
1. When first running the script, Google will ask for permissions
2. Click "Review Permissions"
3. Choose your Google account
4. Click "Advanced" → "Go to Book Tracker Invitation Generator (unsafe)"
5. Click "Allow"

### "API Error" when generating tokens
- Check that your Teacher ID, Class ID, and Invitation Key are correct
- Ensure you're connected to the internet
- Verify the Book Tracker API URL is correct in the script

### Students not appearing in class
- Check that the parent used the exact invitation link
- Verify the token hasn't expired (30 day limit)
- Confirm the parent created their account with the correct email

## Advanced Features

### Custom API URL
If you're using a self-hosted instance of Book Tracker, update the `BOOKTRACKER_API_URL` in the script configuration.

### Batch Processing
The template can handle hundreds of students at once. For very large classes (500+), consider processing in smaller batches.

### Token Regeneration
You can regenerate tokens at any time by clicking "Generate Tokens" again. This is useful if you need to extend expiration dates or if some tokens were compromised.

## Security Notes

- Keep your Invitation Key private - anyone with it can create invitations for your class
- Tokens expire after 30 days for security
- Each token can only be used once per parent account
- If you suspect your Invitation Key is compromised, contact support to generate a new one

## Support

- For technical issues with the Google Sheets template, check the script's error messages
- For Book Tracker app issues, contact your system administrator
- For general questions about the invitation process, refer to the Book Tracker documentation