/**
 * BookTracker Student Invitation System - cCryptoGS Implementation
 * 
 * This version uses the cCryptoGS library which is specifically designed for Google Apps Script.
 * 
 * Setup Instructions:
 * 1. Add cCryptoGS library to your Google Apps Script project:
 *    - Go to Libraries in the Apps Script editor
 *    - Add library ID: 1IEkpeS8hsMSVLRdCMprij996zG6ek9UvGwcCJao_hlDMlgbWWvJpONrs
 *    - Select the latest version, set identifier as "cCryptoGS", and save
 * 2. Set your INVITATION_KEY below (paste from teacher dashboard)
 * 3. Use =GENERATE_TOKEN("Student Name") in your spreadsheet
 */

// ========== CONFIGURATION ==========
const INVITATION_KEY = "MXw5ZDA5ZDQ1ZWJhNzE2NTcyODU4ZGUwMWE5ZDA3Mzk3NjMxY2UyZjE0ZDIzMWZkMzM4ODJiZTY3NDIzYjc1Yjg3";
const BASE_URL = "https://booktracker.rustyphillips.net/invite/";

// ========== KEY PARSING ==========

function parseCompoundKey(compoundKey) {
  try {
    const decoded = Utilities.base64Decode(compoundKey);
    const decodedString = Utilities.newBlob(decoded).getDataAsString();
    
    const parts = decodedString.split('|');
    if (parts.length !== 2) {
      throw new Error('Invalid compound key format');
    }
    
    const classId = parseInt(parts[0]);
    const keyHex = parts[1];
    
    if (isNaN(classId)) {
      throw new Error('Invalid class ID in compound key');
    }
    
    return { classId, keyHex };
  } catch (error) {
    throw new Error(`Failed to parse compound key: ${error.message}`);
  }
}

/**
 * Converts hex string to byte array for cCryptoGS
 */
function hexToBytes(hex) {
  const bytes = [];
  for (let i = 0; i < hex.length; i += 2) {
    bytes.push(parseInt(hex.substr(i, 2), 16));
  }
  return bytes;
}

// ========== ENCRYPTION USING cCryptoGS ==========

/**
 * Encrypts student invitation data using cCryptoGS library
 */
function encryptInvitationData(studentName) {
  try {
    if (!INVITATION_KEY || INVITATION_KEY.trim() === '') {
      throw new Error('INVITATION_KEY not set');
    }
    
    // Parse compound key
    const { classId, keyHex } = parseCompoundKey(INVITATION_KEY);
    
    // Create payload: classId|studentName|timestamp
    const timestamp = Math.floor(Date.now() / 1000);
    const payload = `${classId}|${studentName}|${timestamp}`;
    
    // Convert hex key to bytes
    const keyBytes = hexToBytes(keyHex);
    
    // Generate random IV (16 bytes)
    const iv = [];
    for (let i = 0; i < 16; i++) {
      iv.push(Math.floor(Math.random() * 256));
    }
    
    // Convert payload to bytes
    const payloadBytes = Utilities.newBlob(payload).getBytes();
    
    // Encrypt using cCryptoGS AES-CBC
    const encrypted = cCryptoGS.CryptoJS.AES.encrypt(
      cCryptoGS.CryptoJS.lib.WordArray.create(payloadBytes),
      cCryptoGS.CryptoJS.lib.WordArray.create(keyBytes),
      {
        iv: cCryptoGS.CryptoJS.lib.WordArray.create(iv),
        mode: cCryptoGS.CryptoJS.mode.CBC,
        padding: cCryptoGS.CryptoJS.pad.Pkcs7
      }
    );
    
    // Combine IV + encrypted data
    const combined = iv.concat(Array.from(encrypted.ciphertext.words.flatMap(word => [
      (word >>> 24) & 0xff,
      (word >>> 16) & 0xff, 
      (word >>> 8) & 0xff,
      word & 0xff
    ])));
    
    // Convert to base64url
    const base64 = Utilities.base64Encode(combined);
    return base64
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
    return encryptInvitationData(studentName.toString().trim());
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
    const { classId } = parseCompoundKey(INVITATION_KEY);
    const token = encryptInvitationData(studentName.toString().trim());
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
    return `✅ Class ${classId} configured (cCryptoGS)`;
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
 * Test the encryption system
 */
function testEncryption() {
  try {
    Logger.log('=== cCryptoGS Encryption Test ===');
    Logger.log(`Library available: ${typeof cCryptoGS !== 'undefined'}`);
    
    const testName = 'John Smith';
    const token = GENERATE_TOKEN(testName);
    const url = GENERATE_URL(testName);
    const config = GET_CONFIG();
    
    Logger.log(`Student Name: ${testName}`);
    Logger.log(`Generated Token: ${token}`);
    Logger.log(`Token Length: ${token.length} characters`);
    Logger.log(`Generated URL: ${url}`);
    Logger.log(`Configuration: ${config}`);
    
    // Test with various names
    const testNames = ['Sarah Johnson', 'Mike Brown', 'Emma Davis'];
    Logger.log('\n=== Multiple Name Tests ===');
    testNames.forEach(name => {
      const result = TEST_TOKEN(name);
      Logger.log(`${name} → ${result}`);
    });
    
  } catch (error) {
    Logger.log(`Test failed: ${error.message}`);
    Logger.log('Make sure cCryptoGS library is properly added');
  }
}

/**
 * Setup spreadsheet template
 */
function setupSpreadsheet() {
  const sheet = SpreadsheetApp.getActiveSheet();
  
  // Clear and setup
  sheet.clear();
  
  // Headers
  sheet.getRange(1, 1).setValue('Student Name');
  sheet.getRange(1, 2).setValue('Parent Email');
  sheet.getRange(1, 3).setValue('Invitation URL');
  sheet.getRange(1, 4).setValue('Test');
  
  // Format headers
  const headerRange = sheet.getRange(1, 1, 1, 4);
  headerRange.setFontWeight('bold');
  headerRange.setBackground('#4a90e2');
  headerRange.setFontColor('white');
  
  // Sample data
  sheet.getRange(2, 1).setValue('John Smith');
  sheet.getRange(2, 2).setValue('parent@example.com');
  sheet.getRange(2, 3).setFormula('=IF(A2<>"", GENERATE_URL(A2), "")');
  sheet.getRange(2, 4).setFormula('=IF(A2<>"", TEST_TOKEN(A2), "")');
  
  // Column widths
  sheet.setColumnWidth(1, 150);
  sheet.setColumnWidth(2, 200);
  sheet.setColumnWidth(3, 300);
  sheet.setColumnWidth(4, 150);
  
  // Config display
  sheet.getRange(4, 1).setValue('Configuration:');
  sheet.getRange(4, 1).setFontWeight('bold');
  sheet.getRange(4, 2).setFormula('=GET_CONFIG()');
  
  Logger.log('Spreadsheet setup complete!');
}

// ========== MENU ==========

function onOpen() {
  const ui = SpreadsheetApp.getUi();
  ui.createMenu('📚 Book Tracker')
    .addItem('🔧 Setup Spreadsheet', 'setupSpreadsheet')
    .addItem('🧪 Test Encryption', 'testEncryption')
    .addItem('ℹ️ Show Config', 'showConfig')
    .addToUi();
}

function showConfig() {
  const config = GET_CONFIG();
  SpreadsheetApp.getUi().alert('Configuration', config, SpreadsheetApp.getUi().ButtonSet.OK);
}