/**
 * BookTracker Student Invitation System - Formula-Based
 * 
 * This version provides simple functions that can be used directly in spreadsheet formulas.
 * Put your invitation key in cell B1, then use formulas like =GENERATE_URL(A2) for each student.
 * 
 * Setup Instructions:
 * 1. Add cCryptoGS library: 1IEkpeS8hsMSVLRdCMprij996zG6ek9UvGwcCJao_hlDMlgbWWvJpONrs
 * 2. Paste your invitation key in cell B1
 * 3. Use formulas: =GENERATE_URL(A2), =GENERATE_TOKEN(A2), etc.
 */

const BASE_URL = "https://booktracker.rustyphillips.net/invite/";

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

function hexToBytes(hex) {
  const bytes = [];
  for (let i = 0; i < hex.length; i += 2) {
    bytes.push(parseInt(hex.substr(i, 2), 16));
  }
  return bytes;
}

function encryptData(studentName, invitationKey) {
  try {
    const { classId, keyHex } = parseCompoundKey(invitationKey);
    const timestamp = Math.floor(Date.now() / 1000);
    const payload = `${studentName}|${timestamp}`;
    
    // Try using hex strings directly instead of byte arrays
    const keyHexString = cCryptoGS.CryptoJS.enc.Hex.parse(keyHex);
    const payloadUtf8 = cCryptoGS.CryptoJS.enc.Utf8.parse(payload);
    
    // Generate random IV as hex
    let ivHex = '';
    for (let i = 0; i < 32; i++) { // 32 hex chars = 16 bytes
      ivHex += Math.floor(Math.random() * 16).toString(16);
    }
    const ivWordArray = cCryptoGS.CryptoJS.enc.Hex.parse(ivHex);
    
    // Use CBC mode (compatible with backend CBC decryption)
    const encrypted = cCryptoGS.CryptoJS.AES.encrypt(
      payloadUtf8,
      keyHexString,
      {
        iv: ivWordArray,
        mode: cCryptoGS.CryptoJS.mode.CBC,
        padding: cCryptoGS.CryptoJS.pad.Pkcs7
      }
    );
    
    // Extract IV and ciphertext as hex
    const ivBytes = hexToBytes(ivHex);
    const ciphertextBytes = wordArrayToBytes(encrypted.ciphertext);
    const combined = ivBytes.concat(ciphertextBytes);
    
    // Convert to base64 and make URL-safe
    let base64 = Utilities.base64Encode(combined);
    base64 = base64.replace(/\+/g, '-').replace(/\//g, '_');
    base64 = base64.replace(/=/g, '');
    return base64;
      
  } catch (error) {
    throw new Error(`Encryption failed: ${error.message}`);
  }
}

// Helper function to convert CryptoJS WordArray to byte array
function wordArrayToBytes(wordArray) {
  if (!wordArray || !wordArray.words) return [];
  
  const bytes = [];
  for (let i = 0; i < wordArray.words.length; i++) {
    const word = wordArray.words[i];
    bytes.push((word >>> 24) & 0xff);
    bytes.push((word >>> 16) & 0xff);
    bytes.push((word >>> 8) & 0xff);
    bytes.push(word & 0xff);
  }
  
  // Trim to actual byte length
  return bytes.slice(0, wordArray.sigBytes || bytes.length);
}

/**
 * Generate invitation token using key from B1
 * Usage: =GENERATE_TOKEN(A2)
 */
function GENERATE_TOKEN(studentName) {
  if (!studentName || studentName.toString().trim() === '') return '';
  
  try {
    const sheet = SpreadsheetApp.getActiveSheet();
    const invitationKey = sheet.getRange('B1').getValue();
    if (!invitationKey) return 'ERROR: No key in B1';
    
    return encryptData(studentName.toString().trim(), invitationKey.toString());
  } catch (error) {
    return `ERROR: ${error.message}`;
  }
}

/**
 * Generate full invitation URL using key from B1
 * Usage: =GENERATE_URL(A2)
 */
function GENERATE_URL(studentName) {
  if (!studentName || studentName.toString().trim() === '') return '';
  
  try {
    const sheet = SpreadsheetApp.getActiveSheet();
    const invitationKey = sheet.getRange('B1').getValue();
    if (!invitationKey) return 'ERROR: No key in B1';
    
    const { classId } = parseCompoundKey(invitationKey.toString());
    const token = encryptData(studentName.toString().trim(), invitationKey.toString());
    if (token.startsWith && token.startsWith('ERROR:')) return token;
    
    return `${BASE_URL}${classId}/${token}`;
  } catch (error) {
    return `ERROR: ${error.message}`;
  }
}

/**
 * Test token generation
 * Usage: =TEST_TOKEN(A2)
 */
function TEST_TOKEN(studentName) {
  if (!studentName || studentName.toString().trim() === '') return '';
  
  try {
    const sheet = SpreadsheetApp.getActiveSheet();
    const invitationKey = sheet.getRange('B1').getValue();
    if (!invitationKey) return '❌ No key in B1';
    
    const { classId } = parseCompoundKey(invitationKey.toString());
    const token = encryptData(studentName.toString().trim(), invitationKey.toString());
    return `✅ Class ${classId} (${token.length} chars)`;
  } catch (error) {
    return `❌ ${error.message}`;
  }
}