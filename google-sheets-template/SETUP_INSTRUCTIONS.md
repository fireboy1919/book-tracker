# BookTracker Student Invitation System - Google Sheets Template

This template allows teachers to generate secure invitation tokens for students directly in Google Sheets, enabling bulk parent invitations with class-specific security.

## Features

- **Class-Specific Security**: Each class has its own encryption key to prevent cross-class token misuse
- **Compact Tokens**: 72-76 character tokens that work in URLs and emails
- **Automated Generation**: Use simple formulas to generate tokens and URLs automatically
- **Validation**: Built-in name validation to prevent errors
- **Bulk Processing**: Process entire student lists at once

## Quick Setup

### 1. Create a New Google Sheet

1. Open [Google Sheets](https://sheets.google.com)
2. Create a new spreadsheet
3. Name it something like "Class Invitations - 3rd Grade"

### 2. Add the Script

1. Go to `Extensions` → `Apps Script`
2. Delete the default `myFunction()` code
3. Copy and paste the entire contents of `BookTracker_GoogleSheets.js`
4. **IMPORTANT**: Update the `CLASS_ID` variable with your actual class ID:
   ```javascript
   const CLASS_ID = 123; // Replace 123 with your class ID
   ```
5. Save the script (Ctrl+S or Cmd+S)

### 3. Set Up the Template

1. Go back to your Google Sheet
2. In the Apps Script editor, run the `setupInvitationTemplate()` function:
   - Click the function dropdown and select `setupInvitationTemplate`
   - Click the ▶️ Run button
   - Grant permissions when prompted
3. The sheet will automatically be formatted with headers and formulas

## Using the Template

### Column Layout

| Column A | Column B | Column C | Column D | Column E |
|----------|----------|----------|----------|----------|
| Student Name | Parent Email | Invitation Token | Invitation URL | Validation |
| John Smith | parent@email.com | Auto-generated | Auto-generated | Auto-validated |

### Adding Students

1. **Enter student names** in Column A (e.g., "John Smith", "Sarah Johnson")
2. **Enter parent emails** in Column B (e.g., "parent@example.com")
3. **Tokens and URLs** will generate automatically in Columns C and D
4. **Validation status** will appear in Column E

### Formula References

The template uses these formulas (automatically added):

- **Column C**: `=IF(A2<>"", GENERATE_TOKEN(A2), "")` - Generates the invitation token
- **Column D**: `=IF(A2<>"", GENERATE_URL(A2), "")` - Generates the complete invitation URL  
- **Column E**: `=IF(A2<>"", VALIDATE_NAME(A2), "")` - Validates the student name

## Available Functions

### Core Functions

- `GENERATE_TOKEN(studentName)` - Creates a secure invitation token
- `GENERATE_URL(studentName)` - Creates a complete invitation URL
- `VALIDATE_NAME(studentName)` - Validates student name format
- `CLASS_INFO()` - Shows current class configuration

### Batch Operations

- `generateTokensForRange()` - Generate tokens for selected range
- `setupInvitationTemplate()` - Set up a new invitation spreadsheet

### Testing

- `testEncryption()` - Test the system with sample data (check Apps Script logs)

## Security Features

### Class Isolation
- Each class uses a unique encryption key derived from the class ID
- Tokens from one class cannot be used to join a different class
- Teachers can safely share spreadsheets without cross-contamination

### Token Expiration
- Tokens automatically expire after 30 days
- Timestamp is embedded in each token for validation

### Name Validation
- Prevents empty or invalid names
- Checks for appropriate length (2-50 characters)
- Allows letters, spaces, hyphens, and apostrophes only

## Usage Workflow

### 1. Prepare Student List
```
Column A (Student Names):
John Smith
Sarah Johnson  
Mike Brown
Emma Davis
```

### 2. Add Parent Emails
```
Column B (Parent Emails):
john.parent@email.com
sarah.parent@email.com
mike.parent@email.com
emma.parent@email.com
```

### 3. Copy Invitation URLs
The URLs in Column D will look like:
```
https://booktracker.app/invite/k_zSCm2wGeXHwXNWk5T7c5VKkcB1MVrbIIzQnHkb2bsuJ4VxRp8IPhhUzNdvYCKEKkqXkhk=
```

### 4. Send to Parents
You can:
- Copy URLs individually and send via email
- Use mail merge tools to send bulk emails
- Export the list for use with external email systems

## Troubleshooting

### Common Issues

**"ERROR: Encryption failed"**
- Check that CLASS_ID is set correctly
- Ensure student name is not empty
- Verify the script has proper permissions

**Empty tokens/URLs**
- Student name in Column A might be empty
- Formula might not be copied to all rows
- Check that formulas reference the correct cells

**"Name required" or "Invalid characters"**
- Student name is empty or contains invalid characters
- Only letters, spaces, hyphens, and apostrophes are allowed
- Name must be 2-50 characters long

### Debugging Steps

1. **Test the system**: Run `testEncryption()` in Apps Script and check the logs
2. **Check CLASS_ID**: Run `CLASS_INFO()` formula to verify configuration
3. **Validate names**: Use `VALIDATE_NAME()` to check problematic entries
4. **Check permissions**: Ensure the script has permission to run functions

## Advanced Usage

### Custom Base URL
To change the invitation URL base (e.g., for custom domains):
```javascript
const BASE_URL = "https://your-domain.com/invite/";
```

### Multiple Classes
For teachers with multiple classes:
1. Create separate spreadsheets for each class
2. Set different CLASS_ID values in each script
3. This ensures proper security isolation

### Integration with Mail Merge
The generated URLs work perfectly with mail merge tools:
1. Export columns A, B, and D to CSV
2. Use with Gmail Mail Merge or other tools
3. Include the invitation URL in your email template

## Security Notes

- Never share your CLASS_ID publicly
- Each class should use its own spreadsheet
- Tokens expire automatically for security
- The encryption system prevents token tampering
- Cross-class token usage is automatically blocked

## Support

If you encounter issues:
1. Check the troubleshooting section above
2. Verify your CLASS_ID is correct and the class exists
3. Test with the built-in testing functions
4. Ensure all formulas are properly copied to new rows