/**
 * BookTracker Compound Key Student Invitation System
 * 
 * This version uses a compound key approach where the class ID is encoded
 * with the encryption key, so teachers don't need to know their class ID.
 * 
 * Setup Instructions:
 * 1. Add CryptoJS library to your Google Apps Script project:
 *    - Go to Libraries in the Apps Script editor
 *    - Add library ID: 1rBxx0InQElBgZfBqiPVCEg-jfA3RAnel1p2ivE4S-q2CWtLBB-bnR9IE
 *    - Select the latest version and save
 * 2. Get your compound key from the teacher dashboard
 * 3. Set COMPOUND_KEY below
 * 4. Use the functions in your spreadsheet
 */

// ========== CONFIGURATION ==========
// Get this from your teacher dashboard - format: "classId|encryptionKey" (base64 encoded)
const COMPOUND_KEY = ""; // Teachers will paste this here
const BASE_URL = "https://booktracker.rustyphillips.net/invite/";

// ========== KEY PARSING ==========

/**
 * Parses the compound key to extract class ID and encryption key
 */
function parseCompoundKey(compoundKey) {
  if (!compoundKey || compoundKey.trim() === '') {
    throw new Error('COMPOUND_KEY not set. Please get your compound key from the teacher dashboard.');
  }
  
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
 * Encrypts student invitation data using the compound key
 */
function encryptInvitationData(studentName) {
  try {
    // Parse compound key
    const { classId, keyHex } = parseCompoundKey(COMPOUND_KEY);
    
    // Create compact format: "classId|studentName|timestamp"
    const timestamp = Math.floor(Date.now() / 1000);
    const compactData = `${classId}|${studentName}|${timestamp}`;
    
    // Convert hex key to WordArray
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
    const token = encryptInvitationData(name);
    
    // Parse compound key to show class info
    const { classId } = parseCompoundKey(COMPOUND_KEY);
    
    return `✅ OK (Class ${classId}, ${token.length} chars)`;
  } catch (error) {
    return `❌ Error: ${error.message}`;
  }
}

/**
 * Shows the current class ID from the compound key
 * Usage: =GET_CLASS_ID()
 */
function GET_CLASS_ID() {
  try {
    const { classId } = parseCompoundKey(COMPOUND_KEY);
    return classId;
  } catch (error) {
    return `ERROR: ${error.message}`;
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