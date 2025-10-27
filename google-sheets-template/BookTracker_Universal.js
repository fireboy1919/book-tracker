/**
 * BookTracker Universal Student Invitation System
 * 
 * This version generates tokens that work for ANY class without requiring teachers
 * to know their class ID. The backend will automatically match the student to the correct class.
 * 
 * Setup Instructions:
 * 1. Add CryptoJS library to your Google Apps Script project:
 *    - Go to Libraries in the Apps Script editor
 *    - Add library ID: 1rBxx0InQElBgZfBqiPVCEg-jfA3RAnel1p2ivE4S-q2CWtLBB-bnR9IE
 *    - Select the latest version and save
 * 2. Use the functions in your spreadsheet - no configuration needed!
 */

// ========== CONFIGURATION ==========
const BASE_URL = "https://booktracker.rustyphillips.net/invite/"; // Updated to your domain

// ========== UNIVERSAL TOKEN GENERATION ==========

/**
 * Generates a universal encryption key for class-agnostic tokens
 */
function getUniversalEncryptionKey() {
  const keyString = `booktracker-universal-invitation-key-v1`;
  return CryptoJS.SHA256(keyString);
}

/**
 * Encrypts student invitation data using a universal key
 * The backend will match the student name to the appropriate class
 */
function encryptUniversalInvitationData(studentName) {
  try {
    // Create compact format: "0|studentName|timestamp" (0 = universal class ID)
    const timestamp = Math.floor(Date.now() / 1000);
    const compactData = `0|${studentName}|${timestamp}`;
    
    // Get universal key
    const key = getUniversalEncryptionKey();
    
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

// ========== SPREADSHEET FUNCTIONS ==========

/**
 * Main function to generate invitation tokens
 * Usage: =GENERATE_TOKEN("John Smith")
 */
function GENERATE_TOKEN(studentName) {
  if (!studentName || studentName.toString().trim() === '') {
    return '';
  }
  
  try {
    const name = studentName.toString().trim();
    const token = encryptUniversalInvitationData(name);
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
    const token = encryptUniversalInvitationData(name);
    return `✅ OK (${token.length} chars)`;
  } catch (error) {
    return `❌ Error: ${error.message}`;
  }
}

// ========== BATCH OPERATIONS ==========

/**
 * Generates tokens for a range of student names
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
 * Generates full URLs for a range of student names
 */
function generateURLsForRange() {
  const sheet = SpreadsheetApp.getActiveSheet();
  const range = sheet.getActiveRange();
  const values = range.getValues();
  
  const results = values.map(row => {
    if (row[0] && row[0].toString().trim() !== '') {
      try {
        return [GENERATE_URL(row[0])];
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