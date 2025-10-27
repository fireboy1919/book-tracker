/**
 * BookTracker Student Invitation System - CryptoJS Implementation
 * 
 * This version uses the CryptoJS library for proper AES encryption compatible with the Go backend.
 * 
 * Setup Instructions:
 * 1. Add CryptoJS library to your Google Apps Script project:
 *    - Go to Libraries in the Apps Script editor
 *    - Add library ID: 1rBxx0InQElBgZfBqiPVCEg-jfA3RAnel1p2ivE4S-q2CWtLBB-bnR9IE
 *    - Select the latest version and save
 * 2. Set your CLASS_ID below
 * 3. Use the functions in your spreadsheet
 */

// ========== CONFIGURATION ==========
const CLASS_ID = 123; // Replace with your actual class ID
const BASE_URL = "https://booktracker.app/invite/"; // Base URL for invitation links

// ========== CRYPTO IMPLEMENTATION ==========

/**
 * Generates a class-specific encryption key using SHA256
 * This matches the Go backend implementation exactly
 */
function getClassEncryptionKey(classId) {
  const keyString = `booktracker-class-${classId}-invitation-key-v1`;
  return CryptoJS.SHA256(keyString);
}

/**
 * Encrypts student invitation data using AES-CBC encryption
 * This creates tokens compatible with the Go backend decryption
 */
function encryptInvitationData(classId, studentName) {
  try {
    // Create compact format: "classId|studentName|timestamp"
    const timestamp = Math.floor(Date.now() / 1000);
    const compactData = `${classId}|${studentName}|${timestamp}`;
    
    // Get class-specific key (matches Go backend key generation)
    const key = getClassEncryptionKey(classId);
    
    // Generate random IV (16 bytes for AES-CBC)
    const iv = CryptoJS.lib.WordArray.random(128/8);
    
    // Encrypt using AES-CBC with PKCS7 padding
    const encrypted = CryptoJS.AES.encrypt(compactData, key, {
      iv: iv,
      mode: CryptoJS.mode.CBC,
      padding: CryptoJS.pad.Pkcs7
    });
    
    // Combine IV + encrypted data (this matches Go backend expectation)
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

/**
 * Test decryption function (for debugging only)
 * This helps verify that our encryption is working correctly
 */
function testDecryptInvitationData(token, classId) {
  try {
    // Convert from Base64URL back to regular base64
    let base64 = token.replace(/-/g, '+').replace(/_/g, '/');
    
    // Add padding if needed
    while (base64.length % 4) {
      base64 += '=';
    }
    
    // Parse the combined data
    const combined = CryptoJS.enc.Base64.parse(base64);
    
    // Extract IV (first 16 bytes) and ciphertext (rest)
    const iv = CryptoJS.lib.WordArray.create(combined.words.slice(0, 4));
    const ciphertext = CryptoJS.lib.WordArray.create(combined.words.slice(4));
    
    // Get class key
    const key = getClassEncryptionKey(classId);
    
    // Decrypt
    const decrypted = CryptoJS.AES.decrypt(
      {ciphertext: ciphertext}, 
      key, 
      {
        iv: iv,
        mode: CryptoJS.mode.CBC,
        padding: CryptoJS.pad.Pkcs7
      }
    );
    
    // Convert to string and parse
    const decryptedText = decrypted.toString(CryptoJS.enc.Utf8);
    const parts = decryptedText.split('|');
    
    if (parts.length !== 3) {
      throw new Error('Invalid decrypted format');
    }
    
    return {
      classId: parseInt(parts[0]),
      studentName: parts[1],
      timestamp: parseInt(parts[2]),
      decryptedText: decryptedText
    };
    
  } catch (error) {
    throw new Error(`Decryption failed: ${error.message}`);
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
    const token = encryptInvitationData(CLASS_ID, studentName.toString().trim());
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
    const token = encryptInvitationData(CLASS_ID, studentName.toString().trim());
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
  return `Class ID: ${CLASS_ID}, Base URL: ${BASE_URL}`;
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

/**
 * Tests the encryption and decryption process
 * Usage: =TEST_CRYPTO(A2) - for debugging only
 */
function TEST_CRYPTO(studentName) {
  if (!studentName || studentName.toString().trim() === '') {
    return 'No name provided';
  }
  
  try {
    const name = studentName.toString().trim();
    const token = encryptInvitationData(CLASS_ID, name);
    const decrypted = testDecryptInvitationData(token, CLASS_ID);
    
    if (decrypted.studentName === name && decrypted.classId === CLASS_ID) {
      return `✅ OK (${token.length} chars)`;
    } else {
      return `❌ Mismatch: ${decrypted.studentName} vs ${name}`;
    }
  } catch (error) {
    return `❌ Error: ${error.message}`;
  }
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
  sheet.getRange(1, 6).setValue('Test');
  
  // Format headers
  const headerRange = sheet.getRange(1, 1, 1, 6);
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
  sheet.getRange(2, 6).setFormula('=IF(A2<>"", TEST_CRYPTO(A2), "")');
  
  // Set column widths
  sheet.setColumnWidth(1, 150); // Student Name
  sheet.setColumnWidth(2, 200); // Parent Email
  sheet.setColumnWidth(3, 250); // Invitation Token
  sheet.setColumnWidth(4, 300); // Invitation URL
  sheet.setColumnWidth(5, 120); // Validation
  sheet.setColumnWidth(6, 150); // Test
  
  // Add instructions
  sheet.getRange(4, 1, 1, 6).merge();
  sheet.getRange(4, 1).setValue(
    `Instructions: 1) Add CryptoJS library, 2) Set CLASS_ID in script, 3) Enter student names in column A, ` +
    `4) Enter parent emails in column B, 5) Tokens and URLs will generate automatically. Test column verifies encryption works.`
  );
  sheet.getRange(4, 1).setFontStyle('italic');
  sheet.getRange(4, 1).setWrap(true);
  sheet.getRange(4, 1).setVerticalAlignment('top');
  
  Logger.log('Invitation template created successfully!');
  Logger.log(`Current CLASS_ID: ${CLASS_ID}`);
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
    
    Logger.log('=== CryptoJS Encryption Test Results ===');
    Logger.log(`Student Name: ${testName}`);
    Logger.log(`Generated Token: ${token}`);
    Logger.log(`Token Length: ${token.length} characters`);
    Logger.log(`Generated URL: ${url}`);
    Logger.log(`Name Validation: ${validation}`);
    Logger.log(`Class Info: ${CLASS_INFO()}`);
    
    // Test encryption/decryption roundtrip
    try {
      const decrypted = testDecryptInvitationData(token, CLASS_ID);
      Logger.log(`Decryption Test: ${decrypted.studentName} (Class ${decrypted.classId})`);
      Logger.log(`Timestamp: ${decrypted.timestamp} (${new Date(decrypted.timestamp * 1000)})`);
      
      if (decrypted.studentName === testName && decrypted.classId === CLASS_ID) {
        Logger.log('✅ Encryption/Decryption working correctly!');
      } else {
        Logger.log('❌ Encryption/Decryption mismatch!');
      }
    } catch (decryptError) {
      Logger.log(`❌ Decryption test failed: ${decryptError.message}`);
    }
    
    // Test with various names
    const testNames = ['Sarah Johnson', 'Mike Brown', 'Emma Davis', 'Alex Rodriguez'];
    Logger.log('\n=== Multiple Name Tests ===');
    testNames.forEach(name => {
      const token = GENERATE_TOKEN(name);
      const testResult = TEST_CRYPTO(name);
      Logger.log(`${name} → ${token.length} chars → ${testResult}`);
    });
    
  } catch (error) {
    Logger.log(`Test failed: ${error.message}`);
    Logger.log('Make sure CryptoJS library is properly added to this project');
  }
}

/**
 * Test with wrong class ID to verify security
 */
function testSecurity() {
  try {
    const testName = 'Security Test';
    const correctClassId = CLASS_ID;
    const wrongClassId = 999;
    
    Logger.log('=== Security Test ===');
    
    // Generate token with correct class
    const token = encryptInvitationData(correctClassId, testName);
    Logger.log(`Token generated with Class ${correctClassId}: ${token}`);
    
    // Try to decrypt with correct class (should work)
    try {
      const correctDecrypt = testDecryptInvitationData(token, correctClassId);
      Logger.log(`✅ Correct class decryption: ${correctDecrypt.studentName}`);
    } catch (error) {
      Logger.log(`❌ Correct class decryption failed: ${error.message}`);
    }
    
    // Try to decrypt with wrong class (should fail)
    try {
      const wrongDecrypt = testDecryptInvitationData(token, wrongClassId);
      Logger.log(`❌ Security breach: Wrong class decryption succeeded: ${wrongDecrypt.studentName}`);
    } catch (error) {
      Logger.log(`✅ Security working: Wrong class decryption properly failed`);
    }
    
  } catch (error) {
    Logger.log(`Security test failed: ${error.message}`);
  }
}