/**
 * BookTracker Student Invitation System - Simplified Version
 * 
 * This version uses a web service call to generate tokens instead of local encryption.
 * This is more reliable but requires internet connection.
 * 
 * Setup Instructions:
 * 1. Paste your invitation key in cell B1
 * 2. Use formulas: =GENERATE_URL_SIMPLE(A2), =GENERATE_TOKEN_SIMPLE(A2)
 */

const BASE_URL = "https://booktracker.rustyphillips.net/invite/";
const TOKEN_SERVICE_URL = "https://booktracker.rustyphillips.net/api/generate-invitation-token";

function parseCompoundKey(compoundKey) {
  try {
    const decoded = Utilities.base64Decode(compoundKey);
    const decodedString = Utilities.newBlob(decoded).getDataAsString();
    const parts = decodedString.split('|');
    if (parts.length !== 2) throw new Error('Invalid key format');
    return { classId: parseInt(parts[0]), keyHex: parts[1] };
  } catch (error) {
    throw new Error(`Key parse failed: ${error.message}`);
  }
}

/**
 * Generate invitation token using web service call
 * Usage: =GENERATE_TOKEN_SIMPLE(A2)
 */
function GENERATE_TOKEN_SIMPLE(studentName) {
  if (!studentName || studentName.toString().trim() === '') return '';
  
  try {
    const sheet = SpreadsheetApp.getActiveSheet();
    const invitationKey = sheet.getRange('B1').getValue();
    if (!invitationKey) return 'ERROR: No key in B1';
    
    const { classId, keyHex } = parseCompoundKey(invitationKey.toString());
    
    // Make web service call to generate token
    const payload = {
      studentName: studentName.toString().trim(),
      classId: classId,
      keyHex: keyHex
    };
    
    const options = {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      payload: JSON.stringify(payload)
    };
    
    const response = UrlFetchApp.fetch(TOKEN_SERVICE_URL, options);
    const result = JSON.parse(response.getContentText());
    
    if (result.error) {
      return `ERROR: ${result.error}`;
    }
    
    return result.token;
  } catch (error) {
    return `ERROR: ${error.message}`;
  }
}

/**
 * Generate full invitation URL using web service
 * Usage: =GENERATE_URL_SIMPLE(A2)
 */
function GENERATE_URL_SIMPLE(studentName) {
  if (!studentName || studentName.toString().trim() === '') return '';
  
  try {
    const sheet = SpreadsheetApp.getActiveSheet();
    const invitationKey = sheet.getRange('B1').getValue();
    if (!invitationKey) return 'ERROR: No key in B1';
    
    const { classId } = parseCompoundKey(invitationKey.toString());
    const token = GENERATE_TOKEN_SIMPLE(studentName);
    if (token.startsWith && token.startsWith('ERROR:')) return token;
    
    return `${BASE_URL}${classId}/${token}`;
  } catch (error) {
    return `ERROR: ${error.message}`;
  }
}

/**
 * Test token generation using web service
 * Usage: =TEST_TOKEN_SIMPLE(A2)
 */
function TEST_TOKEN_SIMPLE(studentName) {
  if (!studentName || studentName.toString().trim() === '') return '';
  
  try {
    const sheet = SpreadsheetApp.getActiveSheet();
    const invitationKey = sheet.getRange('B1').getValue();
    if (!invitationKey) return '❌ No key in B1';
    
    const { classId } = parseCompoundKey(invitationKey.toString());
    const token = GENERATE_TOKEN_SIMPLE(studentName);
    if (token.startsWith && token.startsWith('ERROR:')) return token;
    
    return `✅ Class ${classId} (${token.length} chars)`;
  } catch (error) {
    return `❌ ${error.message}`;
  }
}