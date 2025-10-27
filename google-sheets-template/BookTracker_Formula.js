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
    
    const keyBytes = hexToBytes(keyHex);
    const iv = [];
    for (let i = 0; i < 16; i++) {
      iv.push(Math.floor(Math.random() * 256));
    }
    
    const payloadBytes = Utilities.newBlob(payload).getBytes();
    
    const encrypted = cCryptoGS.CryptoJS.AES.encrypt(
      cCryptoGS.CryptoJS.lib.WordArray.create(payloadBytes),
      cCryptoGS.CryptoJS.lib.WordArray.create(keyBytes),
      {
        iv: cCryptoGS.CryptoJS.lib.WordArray.create(iv),
        mode: cCryptoGS.CryptoJS.mode.CBC,
        padding: cCryptoGS.CryptoJS.pad.Pkcs7
      }
    );
    
    const combined = iv.concat(Array.from(encrypted.ciphertext.words.flatMap(word => [
      (word >>> 24) & 0xff, (word >>> 16) & 0xff, (word >>> 8) & 0xff, word & 0xff
    ])));
    
    // Convert to base64 and make URL-safe
    let base64 = Utilities.base64Encode(combined);
    // Ensure URL-safe encoding
    base64 = base64.replace(/\+/g, '-').replace(/\//g, '_');
    // Remove padding for URL safety
    base64 = base64.replace(/=/g, '');
    return base64;
      
  } catch (error) {
    throw new Error(`Encryption failed: ${error.message}`);
  }
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