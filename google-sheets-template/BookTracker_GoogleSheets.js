/**
 * BookTracker Student Invitation System - Google Sheets Implementation
 * 
 * This script allows teachers to generate secure invitation tokens directly in Google Sheets.
 * Teachers paste their invitation key from the frontend into a spreadsheet cell.
 * 
 * Instructions:
 * 1. Add CryptoJS library to your Google Apps Script project (see setup instructions below)
 * 2. Create a cell named "INVITATION_KEY" and paste your invitation key from the teacher dashboard
 * 3. Use =GENERATE_TOKEN(A2) formula where A2 contains the student name
 * 4. The token will be generated automatically and can be used for parent invitations
 * 
 * CryptoJS Library Setup:
 * - Go to Libraries in Apps Script editor
 * - Add library ID: 1rBxx0InQElBgZfBqiPVCEg-jfA3RAnel1p2ivE4S-q2CWtLBB-bnR9IE
 * - Select latest version, identifier: "CryptoJS", save
 */

// ========== CONFIGURATION ==========
const BASE_URL = "https://booktracker.rustyphillips.net/invite/"; // Base URL for invitation links

// ========== KEY MANAGEMENT ==========

/**
 * Gets the invitation key from the spreadsheet
 */
function getInvitationKeyFromSheet() {
  const sheet = SpreadsheetApp.getActiveSheet();
  
  // Try to find a cell named "INVITATION_KEY"
  const namedRanges = SpreadsheetApp.getActiveSpreadsheet().getNamedRanges();
  for (let range of namedRanges) {
    if (range.getName() === "INVITATION_KEY") {
      const value = range.getRange().getValue();
      if (value && value.toString().trim() !== '') {
        return value.toString().trim();
      }
    }
  }
  
  // If no named range, look for a cell containing "INVITATION_KEY" in column A and get the value from column B
  const lastRow = sheet.getLastRow();
  for (let i = 1; i <= lastRow; i++) {
    const cellA = sheet.getRange(i, 1).getValue();
    if (cellA && cellA.toString().toLowerCase().includes("invitation") && cellA.toString().toLowerCase().includes("key")) {
      const cellB = sheet.getRange(i, 2).getValue();
      if (cellB && cellB.toString().trim() !== '') {
        return cellB.toString().trim();
      }
    }
  }
  
  throw new Error('INVITATION_KEY not found. Please create a named range "INVITATION_KEY" or put "Invitation Key:" in column A and your key in column B.');
}

/**
 * Parses the compound key to extract class ID and encryption key
 */
function parseCompoundKey(compoundKey) {
  try {
    // Decode the base64 compound key
    const decoded = Utilities.base64Decode(compoundKey);
    const decodedString = Utilities.newBlob(decoded).getDataAsString();
    
    // Split on pipe
    const parts = decodedString.split('|');
    if (parts.length !== 2) {
      throw new Error('Invalid compound key format');
    }
    
    const classId = parseInt(parts[0]);
    const keyHex = parts[1];
    
    if (isNaN(classId)) {
      throw new Error('Invalid class ID in compound key');
    }
    
    return {
      classId: classId,
      keyHex: keyHex
    };
  } catch (error) {
    throw new Error(`Failed to parse compound key: ${error.message}`);
  }
}

/**
 * Converts hex string to CryptoJS WordArray
 */
function hexToWordArray(hex) {
  const words = [];
  for (let i = 0; i < hex.length; i += 8) {
    words.push(parseInt(hex.substr(i, 8), 16));
  }
  return CryptoJS.lib.WordArray.create(words, hex.length / 2);
}

// ========== ENCRYPTION ==========

/**
 * Encrypts student invitation data using real AES-CBC encryption with CryptoJS
 */
function encryptInvitationData(studentName) {
  try {
    // Get compound key from spreadsheet
    const compoundKey = getInvitationKeyFromSheet();
    
    // Parse compound key to get class ID and encryption key
    const { classId, keyHex } = parseCompoundKey(compoundKey);
    
    // Create compact format: "studentName|timestamp" (class ID comes from URL)
    const timestamp = Math.floor(Date.now() / 1000);
    const compactData = `${studentName}|${timestamp}`;
    
    // Convert hex key to CryptoJS WordArray
    const key = hexToWordArray(keyHex);
    
    // Generate random IV (16 bytes for AES-CBC)
    const iv = CryptoJS.lib.WordArray.random(128/8);
    
    // Encrypt using AES-CBC with PKCS7 padding
    const encrypted = CryptoJS.AES.encrypt(compactData, key, {
      iv: iv,
      mode: CryptoJS.mode.CBC,
      padding: CryptoJS.pad.Pkcs7
    });
    
    // Combine IV + encrypted data
    const combined = iv.concat(encrypted.ciphertext);
    
    // Convert to Base64URL (URL-safe base64)
    return combined.toString(CryptoJS.enc.Base64)
      .replace(/\+/g, '-')
      .replace(/\//g, '_')
      .replace(/=/g, '');
      
  } catch (error) {
    throw new Error(`Encryption failed: ${error.message}`);
  }
}


// ========== GOOGLE SHEETS FUNCTIONS ==========

/**
 * Generates a secure invitation token for a student
 * Usage: =GENERATE_TOKEN(A2) where A2 contains the student name
 */
function GENERATE_TOKEN(studentName) {
  if (!studentName || studentName.toString().trim() === '') {
    return '';
  }
  
  try {
    const token = encryptInvitationData(studentName.toString().trim());
    return token;
  } catch (error) {
    return `ERROR: ${error.message}`;
  }
}

/**
 * Generates a complete invitation URL for a student
 * Usage: =GENERATE_URL(A2) where A2 contains the student name
 */
function GENERATE_URL(studentName) {
  if (!studentName || studentName.toString().trim() === '') {
    return '';
  }
  
  try {
    const token = encryptInvitationData(studentName.toString().trim());
    return `${BASE_URL}${token}`;
  } catch (error) {
    return `ERROR: ${error.message}`;
  }
}

/**
 * Gets information about the current class configuration
 * Usage: =CLASS_INFO()
 */
function CLASS_INFO() {
  try {
    const compoundKey = getInvitationKeyFromSheet();
    const { classId } = parseCompoundKey(compoundKey);
    return `Class ID: ${classId}, Base URL: ${BASE_URL}`;
  } catch (error) {
    return `ERROR: ${error.message}`;
  }
}

/**
 * Validates that a student name is properly formatted
 * Usage: =VALIDATE_NAME(A2)
 */
function VALIDATE_NAME(studentName) {
  if (!studentName || studentName.toString().trim() === '') {
    return 'Name required';
  }
  
  const name = studentName.toString().trim();
  if (name.length < 2) {
    return 'Name too short';
  }
  
  if (name.length > 50) {
    return 'Name too long';
  }
  
  // Check for valid characters (letters, spaces, hyphens, apostrophes)
  if (!/^[a-zA-Z\s\-'\.]+$/.test(name)) {
    return 'Invalid characters';
  }
  
  return 'Valid';
}

// ========== BATCH OPERATIONS ==========

/**
 * Generates tokens for a range of student names
 * Usage: Select a range and run this function from the script editor
 */
function generateTokensForRange() {
  const sheet = SpreadsheetApp.getActiveSheet();
  const range = sheet.getActiveRange();
  const values = range.getValues();
  
  const results = values.map(row => {
    if (row[0] && row[0].toString().trim() !== '') {
      try {
        return [GENERATE_TOKEN(row[0])];
      } catch (error) {
        return [`ERROR: ${error.message}`];
      }
    }
    return [''];
  });
  
  // Write results to the column next to the selected range
  const outputRange = sheet.getRange(
    range.getRow(), 
    range.getColumn() + 1, 
    results.length, 
    1
  );
  outputRange.setValues(results);
}

/**
 * Creates a complete invitation spreadsheet template
 * Run this function to set up a new spreadsheet for student invitations
 */
function setupInvitationTemplate() {
  const sheet = SpreadsheetApp.getActiveSheet();
  
  // Clear existing content
  sheet.clear();
  
  // Set up headers
  sheet.getRange(1, 1).setValue('Student Name');
  sheet.getRange(1, 2).setValue('Parent Email');
  sheet.getRange(1, 3).setValue('Invitation Token');
  sheet.getRange(1, 4).setValue('Invitation URL');
  sheet.getRange(1, 5).setValue('Validation');
  
  // Format headers
  const headerRange = sheet.getRange(1, 1, 1, 5);
  headerRange.setFontWeight('bold');
  headerRange.setBackground('#4a90e2');
  headerRange.setFontColor('white');
  
  // Add sample data
  sheet.getRange(2, 1).setValue('John Smith');
  sheet.getRange(2, 2).setValue('parent@example.com');
  
  // Add formulas for the sample row
  sheet.getRange(2, 3).setFormula('=IF(A2<>"", GENERATE_TOKEN(A2), "")');
  sheet.getRange(2, 4).setFormula('=IF(A2<>"", GENERATE_URL(A2), "")');
  sheet.getRange(2, 5).setFormula('=IF(A2<>"", VALIDATE_NAME(A2), "")');
  
  // Set column widths
  sheet.setColumnWidth(1, 150); // Student Name
  sheet.setColumnWidth(2, 200); // Parent Email
  sheet.setColumnWidth(3, 250); // Invitation Token
  sheet.setColumnWidth(4, 300); // Invitation URL
  sheet.setColumnWidth(5, 120); // Validation
  
  // Add instructions
  sheet.getRange(4, 1, 1, 5).merge();
  sheet.getRange(4, 1).setValue(
    `Instructions: 1) Paste your invitation key from the teacher dashboard in the key cell, ` +
    `2) Enter student names in column A, 3) Enter parent emails in column B, 4) Tokens and URLs will generate automatically`
  );
  sheet.getRange(4, 1).setFontStyle('italic');
  sheet.getRange(4, 1).setWrap(true);
  
  Logger.log('Invitation template created successfully!');
  try {
    const classInfo = CLASS_INFO();
    Logger.log(`Class Info: ${classInfo}`);
  } catch (error) {
    Logger.log(`Class Info: Not configured yet - ${error.message}`);
  }
  Logger.log(`Current BASE_URL: ${BASE_URL}`);
}

// ========== TESTING FUNCTIONS ==========

/**
 * Test the encryption system with sample data
 * Run this from the script editor to verify everything works
 */
function testEncryption() {
  try {
    const testName = 'John Smith';
    const token = GENERATE_TOKEN(testName);
    const url = GENERATE_URL(testName);
    const validation = VALIDATE_NAME(testName);
    
    Logger.log('=== Encryption Test Results ===');
    Logger.log(`Student Name: ${testName}`);
    Logger.log(`Generated Token: ${token}`);
    Logger.log(`Token Length: ${token.length} characters`);
    Logger.log(`Generated URL: ${url}`);
    Logger.log(`Name Validation: ${validation}`);
    Logger.log(`Class Info: ${CLASS_INFO()}`);
    
    // Test with various names
    const testNames = ['Sarah Johnson', 'Mike Brown', 'Emma Davis', 'Alex Rodriguez'];
    Logger.log('\n=== Multiple Name Tests ===');
    testNames.forEach(name => {
      const token = GENERATE_TOKEN(name);
      Logger.log(`${name} → ${token.length} chars → ${token.substring(0, 20)}...`);
    });
    
  } catch (error) {
    Logger.log(`Test failed: ${error.message}`);
  }
}

// ========== TEACHER SETUP FUNCTIONS ==========

/**
 * Complete setup function for teachers - creates the spreadsheet structure and email template
 * Run this first, then paste your invitation key from the teacher dashboard
 */
function SETUP_INVITATION_KEY() {
  const spreadsheet = SpreadsheetApp.getActiveSpreadsheet();
  
  // Setup main sheet
  const sheet = SpreadsheetApp.getActiveSheet();
  sheet.setName('Student Invitations');
  
  // Create a clear section for the invitation key
  sheet.getRange('A1').setValue('Invitation Key:');
  sheet.getRange('A1').setFontWeight('bold');
  sheet.getRange('A1').setBackground('#E8F0FE');
  
  sheet.getRange('B1').setValue('Paste your invitation key here from the teacher dashboard');
  sheet.getRange('B1').setBackground('#FFF3E0');
  
  // Create a named range for easy access
  const keyRange = sheet.getRange('B1');
  spreadsheet.setNamedRange('INVITATION_KEY', keyRange);
  
  // Add student headers below
  sheet.getRange('A3').setValue('Student Name');
  sheet.getRange('B3').setValue('Parent Email');
  sheet.getRange('C3').setValue('Invitation URL');
  
  // Format headers
  const headers = sheet.getRange('A3:C3');
  headers.setFontWeight('bold');
  headers.setBackground('#E8F0FE');
  
  // Add sample student
  sheet.getRange('A4').setValue('John Smith');
  sheet.getRange('B4').setValue('parent@example.com');
  
  // Add formulas that will work once the key is pasted
  sheet.getRange('C4').setFormula('=IF(A4<>"", GENERATE_URL(A4), "")');
  
  // Create email template sheet
  let templateSheet = spreadsheet.getSheetByName('Email Template');
  if (!templateSheet) {
    templateSheet = spreadsheet.insertSheet('Email Template');
  }
  
  // Clear and setup email template
  templateSheet.clear();
  templateSheet.getRange('A1').setValue('📧 Email Template for Mail Merge');
  templateSheet.getRange('A1').setFontWeight('bold');
  templateSheet.getRange('A1').setFontSize(14);
  templateSheet.getRange('A1').setBackground('#E8F0FE');
  
  const emailTemplate = `Subject: Join {{Student Name}}'s Reading Class!

Dear Parent/Guardian,

You're invited to join your child {{Student Name}}'s reading class in Book Tracker!

🔗 Click here to get started: {{Invitation URL}}

What is Book Tracker?
Book Tracker helps students and families track reading progress together. In your child's class, you can:

✓ Log books your child reads
✓ Track reading progress toward class goals
✓ See your child's achievements
✓ Connect with your child's teacher

Getting Started:
1. Click the invitation link above
2. Create your parent account
3. {{Student Name}} will be automatically added to your account

Need Help?
Contact your child's teacher with any questions.

Happy Reading!
The Book Tracker Team`;

  templateSheet.getRange('A3').setValue(emailTemplate);
  templateSheet.getRange('A3').setWrap(true);
  templateSheet.setColumnWidth(1, 600);
  
  // Add mail merge instructions
  templateSheet.getRange('A25').setValue('📬 Mail Merge Instructions:');
  templateSheet.getRange('A25').setFontWeight('bold');
  templateSheet.getRange('A25').setBackground('#FFF3E0');
  
  templateSheet.getRange('A26').setValue('1. Copy the email template above');
  templateSheet.getRange('A27').setValue('2. Use Gmail mail merge add-on (like "Yet Another Mail Merge")');
  templateSheet.getRange('A28').setValue('3. Replace {{Student Name}} with student names from the main sheet');
  templateSheet.getRange('A29').setValue('4. Replace {{Invitation URL}} with invitation URLs from the main sheet');
  templateSheet.getRange('A30').setValue('5. Send personalized emails to all parents at once');
  
  SpreadsheetApp.getUi().alert(
    'Setup Complete!', 
    'Spreadsheet is ready! Two sheets created:\n\n' +
    '• "Student Invitations" - paste your invitation key in cell B1, then add student names\n' +
    '• "Email Template" - ready-to-use email template for mail merge\n\n' +
    'Once you paste your key, invitation URLs will generate automatically!',
    SpreadsheetApp.getUi().ButtonSet.OK
  );
}

// ========== MENU SYSTEM ==========

/**
 * Add menu to the Google Sheets UI when the sheet opens
 */
function onOpen() {
  const ui = SpreadsheetApp.getUi();
  ui.createMenu('📚 Book Tracker')
    .addItem('🔧 Setup Spreadsheet', 'SETUP_INVITATION_KEY')
    .addItem('🎯 Generate All URLs', 'generateAllTokens')
    .addToUi();
}

/**
 * Test function that teachers can run to verify everything is working
 */
function testTokenGeneration() {
  try {
    const testName = 'Test Student';
    const token = GENERATE_TOKEN(testName);
    const url = GENERATE_URL(testName);
    const classInfo = CLASS_INFO();
    
    SpreadsheetApp.getUi().alert(
      '✅ Test Successful!',
      `Configuration is working correctly!\n\n` +
      `Class Info: ${classInfo}\n` +
      `Test Token: ${token.substring(0, 20)}...\n` +
      `Token Length: ${token.length} characters\n` +
      `Test URL: ${url.substring(0, 50)}...`,
      SpreadsheetApp.getUi().ButtonSet.OK
    );
  } catch (error) {
    SpreadsheetApp.getUi().alert(
      '❌ Test Failed',
      `Error: ${error.message}\n\n` +
      `Make sure you have:\n` +
      `1. Added the CryptoJS library\n` +
      `2. Pasted your invitation key in cell B1\n` +
      `3. The key is from your teacher dashboard`,
      SpreadsheetApp.getUi().ButtonSet.OK
    );
  }
}

/**
 * Show class information from the compound key
 */
function showClassInfo() {
  try {
    const classInfo = CLASS_INFO();
    SpreadsheetApp.getUi().alert('Class Information', classInfo, SpreadsheetApp.getUi().ButtonSet.OK);
  } catch (error) {
    SpreadsheetApp.getUi().alert('Error', error.message, SpreadsheetApp.getUi().ButtonSet.OK);
  }
}

/**
 * Generate tokens for all students in the sheet
 */
function generateAllTokens() {
  try {
    const sheet = SpreadsheetApp.getActiveSheet();
    const lastRow = sheet.getLastRow();
    let studentsProcessed = 0;
    
    // Look for student names starting from row 4 (after headers)
    for (let row = 4; row <= lastRow; row++) {
      const studentName = sheet.getRange(row, 1).getValue();
      if (studentName && studentName.toString().trim() !== '') {
        const name = studentName.toString().trim();
        
        // Generate URL (no token column anymore)
        const url = GENERATE_URL(name);
        
        // Update the sheet (column C is URL)
        sheet.getRange(row, 3).setValue(url);
        
        studentsProcessed++;
      }
    }
    
    SpreadsheetApp.getUi().alert(
      '✅ Tokens Generated!',
      `Successfully generated tokens for ${studentsProcessed} students.`,
      SpreadsheetApp.getUi().ButtonSet.OK
    );
  } catch (error) {
    SpreadsheetApp.getUi().alert(
      '❌ Generation Failed',
      `Error: ${error.message}`,
      SpreadsheetApp.getUi().ButtonSet.OK
    );
  }
}

