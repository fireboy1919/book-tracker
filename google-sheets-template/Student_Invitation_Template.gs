/**
 * Book Tracker - Student Invitation Google Sheets Template
 * 
 * This Google Apps Script provides functions to generate invitation tokens
 * for bulk student enrollment in the Book Tracker application using CryptoJS
 * for local encryption (no API calls needed).
 * 
 * Setup Instructions:
 * 1. Get your teacher invitation data from the Book Tracker app
 * 2. Paste your teacher_id, class_id, and invitation_key in the Config sheet
 * 3. Add student names and grades in the Students sheet
 * 4. Use the generateTokens() function to create invitation URLs
 * 
 * Required Library:
 * - Add CryptoJS library: Script ID: 1LOcZic54URqKixmjViMaXzQ7xz1QxvF7Yb4NWFJfCCJ3cH_8mNOhASx1
 */

// Configuration - Replace with your actual values
const CONFIG = {
  BASE_URL: 'https://booktracker.app',  // Replace with your actual URL
  TEACHER_ID: null,     // Will be loaded from Config sheet
  CLASS_ID: null,       // Will be loaded from Config sheet
  INVITATION_KEY: null  // Will be loaded from Config sheet
};

/**
 * Load configuration from the Config sheet
 */
function loadConfig() {
  const sheet = SpreadsheetApp.getActiveSpreadsheet().getSheetByName('Config');
  if (!sheet) {
    throw new Error('Config sheet not found. Please create a Config sheet with your teacher data.');
  }
  
  const teacherId = sheet.getRange('B1').getValue();
  const classId = sheet.getRange('B2').getValue();
  const invitationKey = sheet.getRange('B3').getValue();
  
  if (!teacherId || !classId || !invitationKey) {
    throw new Error('Please fill in all config values: Teacher ID, Class ID, and Invitation Key');
  }
  
  CONFIG.TEACHER_ID = teacherId;
  CONFIG.CLASS_ID = classId;
  CONFIG.INVITATION_KEY = invitationKey;
}

/**
 * Generate invitation tokens for all students in the Students sheet
 */
function generateTokens() {
  try {
    loadConfig();
    
    const sheet = SpreadsheetApp.getActiveSpreadsheet().getSheetByName('Students');
    if (!sheet) {
      throw new Error('Students sheet not found. Please create a Students sheet.');
    }
    
    const lastRow = sheet.getLastRow();
    if (lastRow < 2) {
      SpreadsheetApp.getUi().alert('No students found. Please add students to the Students sheet.');
      return;
    }
    
    // Get student data (starting from row 2, assuming row 1 has headers)
    const range = sheet.getRange(2, 1, lastRow - 1, 4); // Columns A-D
    const values = range.getValues();
    
    // Generate tokens for each student
    for (let i = 0; i < values.length; i++) {
      const row = values[i];
      const studentName = row[0];
      const grade = row[1];
      const parentEmail = row[2];
      
      if (studentName && grade) {
        const token = generateTokenForStudent(studentName, grade);
        const invitationUrl = `${CONFIG.BASE_URL}/invite/${token}`;
        
        // Update the sheet with the token and URL
        sheet.getRange(i + 2, 4).setValue(token); // Column D
        sheet.getRange(i + 2, 5).setValue(invitationUrl); // Column E
      }
    }
    
    SpreadsheetApp.getUi().alert('Tokens generated successfully!');
    
  } catch (error) {
    SpreadsheetApp.getUi().alert('Error: ' + error.message);
  }
}

/**
 * Generate a token for a specific student using CryptoJS encryption
 */
function generateTokenForStudent(studentName, grade) {
  // Create the payload object
  const payload = {
    teacher_id: CONFIG.TEACHER_ID,
    class_id: CONFIG.CLASS_ID,
    student_name: studentName,
    grade: grade,
    timestamp: Math.floor(Date.now() / 1000) // Current timestamp in seconds
  };
  
  try {
    // Convert payload to JSON
    const jsonData = JSON.stringify(payload);
    
    // Convert the base64 teacher key to a WordArray
    const keyBytes = Utilities.base64DecodeWebSafe(CONFIG.INVITATION_KEY);
    const key = CryptoJS.lib.WordArray.create(keyBytes);
    
    // Generate a random IV (12 bytes for GCM)
    const iv = CryptoJS.lib.WordArray.random(12);
    
    // Encrypt the data using AES-GCM
    const encrypted = CryptoJS.AES.encrypt(jsonData, key, {
      iv: iv,
      mode: CryptoJS.mode.GCM,
      padding: CryptoJS.pad.NoPadding
    });
    
    // Combine IV + ciphertext + auth tag
    const combined = iv.concat(encrypted.ciphertext).concat(encrypted.authTag || CryptoJS.lib.WordArray.create());
    
    // Convert to base64 for URL safety
    return Utilities.base64EncodeWebSafe(combined.toString(CryptoJS.enc.Base64));
    
  } catch (error) {
    throw new Error(`Failed to encrypt token for ${studentName}: ${error.message}`);
  }
}

/**
 * Alternative simpler implementation using AES-CBC (if GCM doesn't work)
 */
function generateTokenForStudentCBC(studentName, grade) {
  const payload = {
    teacher_id: CONFIG.TEACHER_ID,
    class_id: CONFIG.CLASS_ID,
    student_name: studentName,
    grade: grade,
    timestamp: Math.floor(Date.now() / 1000)
  };
  
  try {
    // Convert payload to JSON
    const jsonData = JSON.stringify(payload);
    
    // Convert the base64 teacher key to bytes and then to WordArray
    const keyBytes = Utilities.base64DecodeWebSafe(CONFIG.INVITATION_KEY);
    const key = CryptoJS.lib.WordArray.create(keyBytes);
    
    // Generate a random IV (16 bytes for CBC)
    const iv = CryptoJS.lib.WordArray.random(16);
    
    // Encrypt the data using AES-CBC
    const encrypted = CryptoJS.AES.encrypt(jsonData, key, {
      iv: iv,
      mode: CryptoJS.mode.CBC,
      padding: CryptoJS.pad.Pkcs7
    });
    
    // Combine IV + ciphertext
    const combined = iv.concat(encrypted.ciphertext);
    
    // Convert to base64 for URL safety
    return Utilities.base64EncodeWebSafe(combined.toString(CryptoJS.enc.Base64));
    
  } catch (error) {
    throw new Error(`Failed to encrypt token for ${studentName}: ${error.message}`);
  }
}

/**
 * Create the necessary sheets with proper headers
 */
function setupSheets() {
  const spreadsheet = SpreadsheetApp.getActiveSpreadsheet();
  
  // Create Config sheet
  let configSheet = spreadsheet.getSheetByName('Config');
  if (!configSheet) {
    configSheet = spreadsheet.insertSheet('Config');
  }
  
  // Set up Config sheet headers and structure
  configSheet.getRange('A1').setValue('Teacher ID:');
  configSheet.getRange('A2').setValue('Class ID:');
  configSheet.getRange('A3').setValue('Invitation Key:');
  configSheet.getRange('A5').setValue('Instructions:');
  configSheet.getRange('A6').setValue('1. Get your teacher data from Book Tracker app');
  configSheet.getRange('A7').setValue('2. Paste Teacher ID in B1');
  configSheet.getRange('A8').setValue('3. Paste Class ID in B2');
  configSheet.getRange('A9').setValue('4. Paste Invitation Key in B3');
  configSheet.getRange('A10').setValue('5. Add students in Students sheet');
  configSheet.getRange('A11').setValue('6. Run generateTokens() function');
  
  // Create Students sheet
  let studentsSheet = spreadsheet.getSheetByName('Students');
  if (!studentsSheet) {
    studentsSheet = spreadsheet.insertSheet('Students');
  }
  
  // Set up Students sheet headers
  studentsSheet.getRange('A1').setValue('Student Name');
  studentsSheet.getRange('B1').setValue('Grade');
  studentsSheet.getRange('C1').setValue('Parent Email');
  studentsSheet.getRange('D1').setValue('Invitation Token');
  studentsSheet.getRange('E1').setValue('Invitation URL');
  
  // Add sample data
  studentsSheet.getRange('A2').setValue('John Smith');
  studentsSheet.getRange('B2').setValue('3rd Grade');
  studentsSheet.getRange('C2').setValue('parent@example.com');
  
  // Format headers
  const configHeaders = configSheet.getRange('A1:A3');
  configHeaders.setFontWeight('bold');
  
  const studentHeaders = studentsSheet.getRange('A1:E1');
  studentHeaders.setFontWeight('bold');
  studentHeaders.setBackground('#E8F0FE');
  
  SpreadsheetApp.getUi().alert('Sheets setup complete! Please fill in your teacher data in the Config sheet.');
}

/**
 * Create a mail merge template for sending invitations
 */
function createMailMergeTemplate() {
  const spreadsheet = SpreadsheetApp.getActiveSpreadsheet();
  
  let templateSheet = spreadsheet.getSheetByName('Email Template');
  if (!templateSheet) {
    templateSheet = spreadsheet.insertSheet('Email Template');
  }
  
  // Email template
  const emailTemplate = `Subject: Invitation to Join {{STUDENT_NAME}}'s Reading Class

Dear Parent/Guardian,

You're invited to join your child {{STUDENT_NAME}}'s reading class in the Book Tracker application!

🔗 Click this link to get started: {{INVITATION_URL}}

What is Book Tracker?
Book Tracker helps students and families track reading progress and achieve reading goals together. Your child's teacher has set up a class where you can:

✓ Log books your child reads
✓ Track reading progress toward class goals
✓ See how your child is doing compared to class objectives
✓ Celebrate reading achievements

Getting Started:
1. Click the invitation link above
2. Create your parent account (or log in if you already have one)
3. The system will automatically add {{STUDENT_NAME}} to your account and assign them to the class

Need Help?
If you have any questions about using Book Tracker, please contact your child's teacher or visit our help center.

Happy Reading!
The Book Tracker Team`;

  templateSheet.getRange('A1').setValue('Email Template (Copy this text for your mail merge):');
  templateSheet.getRange('A3').setValue(emailTemplate);
  
  // Instructions
  templateSheet.getRange('A25').setValue('Mail Merge Instructions:');
  templateSheet.getRange('A26').setValue('1. Use Gmail\'s mail merge add-on (like "Yet Another Mail Merge")');
  templateSheet.getRange('A27').setValue('2. Copy the email template above');
  templateSheet.getRange('A28').setValue('3. Use the Students sheet as your data source');
  templateSheet.getRange('A29').setValue('4. Map {{STUDENT_NAME}} to Student Name column');
  templateSheet.getRange('A30').setValue('5. Map {{INVITATION_URL}} to Invitation URL column');
  
  SpreadsheetApp.getUi().alert('Email template created! Check the "Email Template" sheet.');
}

/**
 * Add menu to the Google Sheets UI
 */
function onOpen() {
  const ui = SpreadsheetApp.getUi();
  ui.createMenu('Book Tracker')
    .addItem('📚 Setup Sheets', 'setupSheets')
    .addSeparator()
    .addItem('🔧 Setup CryptoJS Library', 'setupCryptoJS') 
    .addItem('✅ Test Configuration', 'testConfiguration')
    .addSeparator()
    .addItem('🎯 Generate Tokens', 'generateTokens')
    .addSeparator()
    .addItem('📧 Create Email Template', 'createMailMergeTemplate')
    .addItem('🚀 Setup Mail Merge', 'setupMailMerge')
    .addItem('📬 Create Gmail Drafts', 'createGmailDrafts')
    .addToUi();
}

/**
 * Test function to validate configuration and encryption
 */
function testConfiguration() {
  try {
    loadConfig();
    
    // Test encryption with a sample student
    const testToken = generateTokenForStudent("Test Student", "Test Grade");
    
    SpreadsheetApp.getUi().alert(`✅ Configuration and encryption working!

Teacher ID: ${CONFIG.TEACHER_ID}
Class ID: ${CONFIG.CLASS_ID}
Key Length: ${CONFIG.INVITATION_KEY ? CONFIG.INVITATION_KEY.length : 0} characters

Sample token generated: ${testToken.substring(0, 20)}...

You can now use generateTokens() to create tokens for all students.`);
  } catch (error) {
    SpreadsheetApp.getUi().alert('❌ Configuration Error: ' + error.message + '\n\nMake sure you have:\n1. Added CryptoJS library\n2. Filled in Config sheet values\n3. Correct base64 invitation key');
  }
}

/**
 * Setup function to add CryptoJS library (instructions)
 */
function setupCryptoJS() {
  const instructions = `To add the CryptoJS library:

1. In Apps Script, click "Libraries" in the left sidebar
2. Click "+ Add a library"
3. Enter this Script ID: 1LOcZic54URqKixmjViMaXzQ7xz1QxvF7Yb4NWFJfCCJ3cH_8mNOhASx1
4. Click "Look up"
5. Select the latest version
6. Set Identifier to "CryptoJS"
7. Click "Save"

Then run testConfiguration() to verify everything works.`;

  SpreadsheetApp.getUi().alert(instructions);
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
 * Create draft emails directly in Gmail (requires Gmail API setup)
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