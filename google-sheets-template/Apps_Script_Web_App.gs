/**
 * Google Apps Script Web App for Book Tracker
 * This creates personalized copies of the invitation template
 * 
 * Deploy as Web App with:
 * - Execute as: Me
 * - Who has access: Anyone
 */

// Template spreadsheet ID (your master template)
const TEMPLATE_ID = 'YOUR_TEMPLATE_SPREADSHEET_ID';

/**
 * Main function called when the web app is accessed
 */
function doGet(e) {
  const params = e.parameter;
  
  // Get teacher data from URL parameters
  const teacherId = params.teacherId;
  const classId = params.classId;
  const invitationKey = params.invitationKey;
  const className = params.className;
  
  if (!teacherId || !classId || !invitationKey) {
    return HtmlService.createHtmlOutput(`
      <html>
        <body style="font-family: Arial, sans-serif; padding: 20px;">
          <h2>❌ Missing Parameters</h2>
          <p>Please access this link from your Book Tracker dashboard.</p>
          <button onclick="window.close()">Close Window</button>
        </body>
      </html>
    `);
  }
  
  try {
    // Create a copy of the template
    const template = DriveApp.getFileById(TEMPLATE_ID);
    const copyName = `Book Tracker Invitations - ${className} - ${new Date().toLocaleDateString()}`;
    const newSheet = template.makeCopy(copyName);
    
    // Open the new spreadsheet
    const spreadsheet = SpreadsheetApp.openById(newSheet.getId());
    
    // Fill in the teacher's configuration
    setupTeacherConfig(spreadsheet, teacherId, classId, invitationKey, className);
    
    // Set up mail merge template
    setupMailMergeTemplate(spreadsheet, className);
    
    // Return success page with link to new sheet
    const sheetUrl = spreadsheet.getUrl();
    
    return HtmlService.createHtmlOutput(`
      <html>
        <head>
          <style>
            body { 
              font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
              max-width: 600px; 
              margin: 0 auto; 
              padding: 40px 20px;
              background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
              color: white;
            }
            .container {
              background: white;
              color: #333;
              padding: 30px;
              border-radius: 12px;
              box-shadow: 0 10px 30px rgba(0,0,0,0.2);
              text-align: center;
            }
            .success-icon { font-size: 48px; margin-bottom: 20px; }
            .button {
              background: #4CAF50;
              color: white;
              padding: 15px 30px;
              border: none;
              border-radius: 6px;
              font-size: 16px;
              cursor: pointer;
              text-decoration: none;
              display: inline-block;
              margin: 10px 5px;
              transition: background 0.3s;
            }
            .button:hover { background: #45a049; }
            .button.secondary {
              background: #2196F3;
            }
            .button.secondary:hover { background: #1976D2; }
            .instructions {
              background: #f8f9fa;
              padding: 15px;
              border-radius: 6px;
              margin: 20px 0;
              text-align: left;
            }
          </style>
        </head>
        <body>
          <div class="container">
            <div class="success-icon">🎉</div>
            <h1>Your Google Sheets Template is Ready!</h1>
            <p>We've created a personalized copy of the Book Tracker invitation template with your class information already filled in.</p>
            
            <div class="instructions">
              <h3>🎯 Complete Mail Merge Setup - Next Steps:</h3>
              <ol>
                <li><strong>Add Students:</strong> Click "Open My Template" → Go to "Students" sheet → Add names, grades, and parent emails</li>
                <li><strong>Generate Tokens:</strong> Click "Book Tracker → Generate Tokens" (creates unique invitation links)</li>
                <li><strong>Send Emails:</strong> Use the pre-made email template in the "Email Template" sheet with Gmail mail merge</li>
                <li><strong>Done!</strong> Parents receive personalized invitations with one-click student setup</li>
              </ol>
              
              <div style="background: #e8f5e8; padding: 10px; border-radius: 4px; margin-top: 10px;">
                <strong>✨ Everything is pre-configured:</strong> Email template, mail merge instructions, and encryption - no technical setup needed!
              </div>
            </div>
            
            <a href="${sheetUrl}" target="_blank" class="button">
              📊 Open My Template
            </a>
            
            <br>
            
            <button onclick="window.close()" class="button secondary">
              ✅ Close This Window
            </button>
            
            <div style="margin-top: 30px; font-size: 14px; color: #666;">
              <p><strong>Class:</strong> ${className}</p>
              <p><strong>Template Name:</strong> ${copyName}</p>
              <p>You can find this sheet in your Google Drive anytime.</p>
            </div>
          </div>
          
          <script>
            // Auto-close after 30 seconds if user doesn't interact
            setTimeout(function() {
              if (confirm('Auto-close this window?')) {
                window.close();
              }
            }, 30000);
          </script>
        </body>
      </html>
    `);
    
  } catch (error) {
    return HtmlService.createHtmlOutput(`
      <html>
        <body style="font-family: Arial, sans-serif; padding: 20px;">
          <h2>❌ Error Creating Template</h2>
          <p><strong>Error:</strong> ${error.message}</p>
          <p>Please try again or contact support if the problem persists.</p>
          <button onclick="window.close()">Close Window</button>
        </body>
      </html>
    `);
  }
}

/**
 * Fill in the teacher's configuration in the new spreadsheet
 */
function setupTeacherConfig(spreadsheet, teacherId, classId, invitationKey, className) {
  // Get or create Config sheet
  let configSheet = spreadsheet.getSheetByName('Config');
  if (!configSheet) {
    configSheet = spreadsheet.insertSheet('Config');
  }
  
  // Set up headers and values
  configSheet.getRange('A1').setValue('Teacher ID:');
  configSheet.getRange('B1').setValue(teacherId);
  
  configSheet.getRange('A2').setValue('Class ID:');
  configSheet.getRange('B2').setValue(classId);
  
  configSheet.getRange('A3').setValue('Invitation Key:');
  configSheet.getRange('B3').setValue(invitationKey);
  
  configSheet.getRange('A4').setValue('Class Name:');
  configSheet.getRange('B4').setValue(className);
  
  configSheet.getRange('A6').setValue('✅ Configuration Complete!');
  configSheet.getRange('A7').setValue('Add your students in the "Students" sheet, then use "Book Tracker → Generate Tokens"');
  
  // Style the config sheet
  const headerRange = configSheet.getRange('A1:A4');
  headerRange.setFontWeight('bold');
  
  const valueRange = configSheet.getRange('B1:B4');
  valueRange.setBackground('#e8f5e8');
  
  // Ensure Students sheet exists with headers
  let studentsSheet = spreadsheet.getSheetByName('Students');
  if (!studentsSheet) {
    studentsSheet = spreadsheet.insertSheet('Students');
  }
  
  // Set up Students sheet headers if empty
  if (studentsSheet.getRange('A1').getValue() === '') {
    studentsSheet.getRange('A1:E1').setValues([['Student Name', 'Grade', 'Parent Email', 'Invitation Token', 'Invitation URL']]);
    studentsSheet.getRange('A1:E1').setFontWeight('bold').setBackground('#4285f4').setFontColor('white');
    
    // Add sample row
    studentsSheet.getRange('A2:C2').setValues([['John Smith', '3rd Grade', 'parent@example.com']]);
  }
}

/**
 * Set up mail merge template and functionality
 */
function setupMailMergeTemplate(spreadsheet, className) {
  // Create Email Template sheet
  let emailSheet = spreadsheet.getSheetByName('Email Template');
  if (!emailSheet) {
    emailSheet = spreadsheet.insertSheet('Email Template');
  }
  
  // Email template content
  const emailTemplate = `Subject: Join ${className}'s Reading Adventure on Book Tracker!

Dear {{Parent Email}} family,

You're invited to join {{Student Name}}'s reading class in Book Tracker! 📚

🔗 **Get Started Here:** {{Invitation URL}}

**What is Book Tracker?**
Book Tracker helps students and families track reading progress together. Your child's teacher has set up a class where you can:

✅ Log books {{Student Name}} reads at home and school
✅ Track progress toward class reading goals  
✅ See {{Student Name}}'s reading achievements
✅ Support your child's love of reading

**Getting Started (Takes 2 minutes):**
1. Click the link above
2. Create your parent account (or sign in if you have one)
3. {{Student Name}} will automatically be added to your account and assigned to ${className}

**Questions?**
If you need help with Book Tracker, please don't hesitate to reach out!

Happy Reading! 📖
${className} Teacher`;

  // Set up template
  emailSheet.getRange('A1').setValue('📧 Email Template for Mail Merge');
  emailSheet.getRange('A1').setFontSize(14).setFontWeight('bold');
  
  emailSheet.getRange('A3').setValue('Copy the template below and use it with Gmail Mail Merge:');
  emailSheet.getRange('A5').setValue(emailTemplate);
  emailSheet.getRange('A5').setWrap(true);
  
  // Instructions
  emailSheet.getRange('A20').setValue('📋 Mail Merge Instructions:');
  emailSheet.getRange('A20').setFontSize(12).setFontWeight('bold');
  
  const instructions = [
    '1. First, generate tokens using "Book Tracker → Generate Tokens"',
    '2. Install "Yet Another Mail Merge" from Google Workspace Marketplace',
    '3. In Gmail, create a new email and paste the template above',
    '4. Use Yet Another Mail Merge with the "Students" sheet as data source',
    '5. Map the fields: {{Student Name}} → Student Name, {{Parent Email}} → Parent Email, {{Invitation URL}} → Invitation URL',
    '6. Send your personalized invitations!',
    '',
    'Alternative: Use any other mail merge tool like Mail Merge with Attachments, etc.'
  ];
  
  for (let i = 0; i < instructions.length; i++) {
    emailSheet.getRange(22 + i, 1).setValue(instructions[i]);
  }
  
  // Style the sheet
  emailSheet.setColumnWidth(1, 800);
  emailSheet.getRange('A5').setBackground('#f8f9fa');
  
  // Add mail merge function button (custom menu will handle this)
  emailSheet.getRange('A35').setValue('🚀 Quick Mail Merge Setup');
  emailSheet.getRange('A35').setFontWeight('bold').setFontSize(12);
  emailSheet.getRange('A36').setValue('Use "Book Tracker → Setup Mail Merge" from the menu for guided setup');
}

/**
 * Set up automated mail merge (called from menu)
 */
function setupMailMerge() {
  const ui = SpreadsheetApp.getUi();
  
  try {
    // Check if tokens are generated
    const studentsSheet = SpreadsheetApp.getActiveSpreadsheet().getSheetByName('Students');
    if (!studentsSheet) {
      throw new Error('Students sheet not found');
    }
    
    const lastRow = studentsSheet.getLastRow();
    if (lastRow < 2) {
      throw new Error('No students found. Please add students first.');
    }
    
    // Check if tokens are generated
    const tokensGenerated = studentsSheet.getRange(2, 5).getValue();
    if (!tokensGenerated) {
      const generateFirst = ui.alert(
        'Generate Tokens First', 
        'You need to generate invitation tokens before setting up mail merge. Generate them now?',
        ui.ButtonSet.YES_NO
      );
      
      if (generateFirst === ui.Button.YES) {
        generateTokens();
      } else {
        return;
      }
    }
    
    // Instructions for mail merge setup
    const instructions = `📧 Mail Merge Setup Complete!

Your spreadsheet is ready for mail merge. Here's what to do next:

🔗 **Install Mail Merge Add-on:**
1. Go to Extensions → Add-ons → Get add-ons
2. Search for "Yet Another Mail Merge" 
3. Install it (it's free!)

📝 **Create Your Email:**
1. Go to Gmail and compose a new email
2. Copy the email template from the "Email Template" sheet
3. Use these placeholders:
   • {{Student Name}} - will be replaced with each student's name
   • {{Parent Email}} - will be replaced with parent email  
   • {{Invitation URL}} - will be replaced with unique invitation link

🚀 **Send Mail Merge:**
1. In Gmail, click "Yet Another Mail Merge" from the add-ons menu
2. Select this spreadsheet as data source
3. Choose the "Students" sheet
4. Map the fields and send!

✅ **Ready to go!** Your personalized invitations will be sent automatically.`;

    ui.alert('Mail Merge Ready!', instructions, ui.ButtonSet.OK);
    
  } catch (error) {
    ui.alert('Setup Error', error.message, ui.ButtonSet.OK);
  }
}

/**
 * Alternative: Create draft emails directly in Gmail (requires Gmail API setup)
 */
function createGmailDrafts() {
  const ui = SpreadsheetApp.getUi();
  
  try {
    loadConfig();
    
    const studentsSheet = SpreadsheetApp.getActiveSpreadsheet().getSheetByName('Students');
    const emailSheet = SpreadsheetApp.getActiveSpreadsheet().getSheetByName('Email Template');
    
    if (!studentsSheet || !emailSheet) {
      throw new Error('Required sheets not found');
    }
    
    // Get email template
    const template = emailSheet.getRange('A5').getValue();
    
    // Get student data
    const lastRow = studentsSheet.getLastRow();
    const studentData = studentsSheet.getRange(2, 1, lastRow - 1, 5).getValues();
    
    let draftsCreated = 0;
    
    // Create draft for each student
    for (let i = 0; i < studentData.length; i++) {
      const [studentName, grade, parentEmail, token, invitationURL] = studentData[i];
      
      if (studentName && parentEmail && invitationURL) {
        // Replace placeholders
        let emailContent = template
          .replace(/\{\{Student Name\}\}/g, studentName)
          .replace(/\{\{Parent Email\}\}/g, parentEmail)
          .replace(/\{\{Invitation URL\}\}/g, invitationURL);
        
        // Extract subject line
        const subjectMatch = emailContent.match(/Subject: (.+)/);
        const subject = subjectMatch ? subjectMatch[1] : `Join ${studentName}'s Reading Class!`;
        const body = emailContent.replace(/Subject: .+\n\n/, '');
        
        // Create Gmail draft
        GmailApp.createDraft(
          parentEmail,
          subject,
          body
        );
        
        draftsCreated++;
      }
    }
    
    ui.alert('Success!', `Created ${draftsCreated} email drafts in your Gmail. Check your Drafts folder to review and send them.`, ui.ButtonSet.OK);
    
  } catch (error) {
    ui.alert('Error', `Failed to create Gmail drafts: ${error.message}\n\nYou may need to authorize Gmail access first.`, ui.ButtonSet.OK);
  }
}

/**
 * Test function for development
 */
function testWebApp() {
  const testParams = {
    parameter: {
      teacherId: '1',
      classId: '1', 
      invitationKey: 'test-key-123',
      className: 'Mrs. Johnson\'s 3rd Grade'
    }
  };
  
  const result = doGet(testParams);
  Logger.log(result.getContent());
}