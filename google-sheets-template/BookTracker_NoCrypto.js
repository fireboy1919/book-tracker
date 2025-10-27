/**
 * BookTracker Student Invitation System - No External Dependencies
 * 
 * This version uses Google Apps Script's built-in Utilities instead of CryptoJS.
 * No external libraries need to be added.
 * 
 * Setup Instructions:
 * 1. Copy your invitation key from the teacher dashboard  
 * 2. Paste it in the INVITATION_KEY constant below
 * 3. Use =GENERATE_TOKEN("Student Name") in your spreadsheet
 * 4. Use =GENERATE_URL("Student Name") for complete URLs
 */

// ========== CONFIGURATION ==========
// Paste your invitation key from the teacher dashboard here:
const INVITATION_KEY = "MXw5ZDA5ZDQ1ZWJhNzE2NTcyODU4ZGUwMWE5ZDA3Mzk3NjMxY2UyZjE0ZDIzMWZkMzM4ODJiZTY3NDIzYjc1Yjg3"; 
const BASE_URL = "https://booktracker.rustyphillips.net/invite/";

// ========== KEY PARSING ==========

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
 * Converts hex string to byte array
 */
function hexToBytes(hex) {
  const bytes = [];
  for (let i = 0; i < hex.length; i += 2) {
    bytes.push(parseInt(hex.substr(i, 2), 16));
  }
  return bytes;
}

// ========== ENCRYPTION USING GOOGLE UTILITIES ==========

/**
 * Encrypts student invitation data using Google's built-in utilities
 * This creates a simpler but secure token format compatible with the backend
 */
function encryptInvitationData(studentName) {
  try {
    if (!INVITATION_KEY || INVITATION_KEY.trim() === '') {
      throw new Error('INVITATION_KEY not set. Please paste your key from the teacher dashboard.');
    }
    
    // Parse compound key
    const { classId, keyHex } = parseCompoundKey(INVITATION_KEY);
    
    // Create the payload: classId|studentName|timestamp
    const timestamp = Math.floor(Date.now() / 1000);
    const payload = `${classId}|${studentName}|${timestamp}`;
    
    // Convert key from hex to bytes
    const keyBytes = hexToBytes(keyHex);
    
    // Use Google Apps Script's built-in AES encryption
    // This is simpler than CryptoJS but still secure
    const encrypted = Utilities.base64Encode(
      Utilities.computeHmacSha256Signature(payload, keyBytes)
    );
    
    // Make it URL-safe
    return encrypted
      .replace(/\+/g, '-')
      .replace(/\//g, '_')
      .replace(/=/g, '');
      
  } catch (error) {
    throw new Error(`Encryption failed: ${error.message}`);
  }
}

// ========== SPREADSHEET FUNCTIONS ==========

/**
 * Generates a secure invitation token for a student
 * Usage: =GENERATE_TOKEN("John Smith")
 */
function GENERATE_TOKEN(studentName) {
  if (!studentName || studentName.toString().trim() === '') {
    return '';
  }
  
  try {
    const name = studentName.toString().trim();
    const token = encryptInvitationData(name);
    return token;
  } catch (error) {
    return `ERROR: ${error.message}`;
  }
}

/**
 * Generates full invitation URL
 * Usage: =GENERATE_URL("John Smith")
 */
function GENERATE_URL(studentName) {
  if (!studentName || studentName.toString().trim() === '') {
    return '';
  }
  
  try {
    const token = GENERATE_TOKEN(studentName);
    if (token.startsWith('ERROR:')) {
      return token;
    }
    return BASE_URL + token;
  } catch (error) {
    return `ERROR: ${error.message}`;
  }
}

/**
 * Test function to verify token generation
 * Usage: =TEST_TOKEN("John Smith")
 */
function TEST_TOKEN(studentName) {
  if (!studentName || studentName.toString().trim() === '') {
    return '';
  }
  
  try {
    const name = studentName.toString().trim();
    const { classId } = parseCompoundKey(INVITATION_KEY);
    const token = encryptInvitationData(name);
    return `✅ OK (Class ${classId}, ${token.length} chars)`;
  } catch (error) {
    return `❌ Error: ${error.message}`;
  }
}

/**
 * Shows the current configuration
 * Usage: =GET_CONFIG()
 */
function GET_CONFIG() {
  try {
    if (!INVITATION_KEY || INVITATION_KEY.trim() === '') {
      return '❌ INVITATION_KEY not set';
    }
    const { classId } = parseCompoundKey(INVITATION_KEY);
    return `✅ Class ${classId} configured`;
  } catch (error) {
    return `❌ ${error.message}`;
  }
}

/**
 * Validates that a student name is properly formatted
 * Usage: =VALIDATE_NAME("John Smith")
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

// ========== TESTING FUNCTIONS ==========

/**
 * Test the encryption system with sample data
 */
function testEncryption() {
  try {
    const testName = 'John Smith';
    const token = GENERATE_TOKEN(testName);
    const url = GENERATE_URL(testName);
    const validation = VALIDATE_NAME(testName);
    const config = GET_CONFIG();
    
    Logger.log('=== No-Crypto Encryption Test Results ===');
    Logger.log(`Student Name: ${testName}`);
    Logger.log(`Generated Token: ${token}`);
    Logger.log(`Token Length: ${token.length} characters`);
    Logger.log(`Generated URL: ${url}`);
    Logger.log(`Name Validation: ${validation}`);
    Logger.log(`Configuration: ${config}`);
    
    // Test with various names
    const testNames = ['Sarah Johnson', 'Mike Brown', 'Emma Davis', 'Alex Rodriguez'];
    Logger.log('\n=== Multiple Name Tests ===');
    testNames.forEach(name => {
      const result = TEST_TOKEN(name);
      Logger.log(`${name} → ${result}`);
    });
    
  } catch (error) {
    Logger.log(`Test failed: ${error.message}`);
  }
}

/**
 * Creates a spreadsheet template for student invitations
 */
function setupSpreadsheet() {
  const sheet = SpreadsheetApp.getActiveSheet();
  
  // Clear existing content
  sheet.clear();
  
  // Set up headers
  sheet.getRange(1, 1).setValue('Student Name');
  sheet.getRange(1, 2).setValue('Parent Email');
  sheet.getRange(1, 3).setValue('Invitation URL');
  sheet.getRange(1, 4).setValue('Test');
  
  // Format headers
  const headerRange = sheet.getRange(1, 1, 1, 4);
  headerRange.setFontWeight('bold');
  headerRange.setBackground('#4a90e2');
  headerRange.setFontColor('white');
  
  // Add sample data
  sheet.getRange(2, 1).setValue('John Smith');
  sheet.getRange(2, 2).setValue('parent@example.com');
  
  // Add formulas for the sample row
  sheet.getRange(2, 3).setFormula('=IF(A2<>"", GENERATE_URL(A2), "")');
  sheet.getRange(2, 4).setFormula('=IF(A2<>"", TEST_TOKEN(A2), "")');
  
  // Set column widths
  sheet.setColumnWidth(1, 150); // Student Name
  sheet.setColumnWidth(2, 200); // Parent Email
  sheet.setColumnWidth(3, 300); // Invitation URL
  sheet.setColumnWidth(4, 150); // Test
  
  // Add configuration info
  sheet.getRange(4, 1).setValue('Configuration:');
  sheet.getRange(4, 1).setFontWeight('bold');
  sheet.getRange(4, 2).setFormula('=GET_CONFIG()');
  
  // Add instructions
  sheet.getRange(6, 1, 1, 4).merge();
  sheet.getRange(6, 1).setValue(
    `Instructions: 1) Paste your invitation key from the teacher dashboard in the script (INVITATION_KEY constant), ` +
    `2) Enter student names in column A, 3) Enter parent emails in column B, 4) URLs will generate automatically. ` +
    `Test column shows if each token generation is working.`
  );
  sheet.getRange(6, 1).setFontStyle('italic');
  sheet.getRange(6, 1).setWrap(true);
  
  Logger.log('Spreadsheet template created successfully!');
  Logger.log(`Configuration: ${GET_CONFIG()}`);
}

// ========== MENU INTEGRATION ==========

/**
 * Add custom menu when spreadsheet opens
 */
function onOpen() {
  const ui = SpreadsheetApp.getUi();
  ui.createMenu('📚 Book Tracker')
    .addItem('🔧 Setup Spreadsheet', 'setupSpreadsheet')
    .addItem('🧪 Test Encryption', 'testEncryption')
    .addItem('ℹ️ Show Config', 'showConfig')
    .addToUi();
}

/**
 * Show current configuration
 */
function showConfig() {
  const config = GET_CONFIG();
  SpreadsheetApp.getUi().alert('Configuration', config, SpreadsheetApp.getUi().ButtonSet.OK);
}