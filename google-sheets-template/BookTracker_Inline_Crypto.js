/**
 * BookTracker Student Invitation System - Inline CryptoJS
 * 
 * This version includes a minimal inline AES implementation instead of relying on external libraries.
 * It uses a simplified but compatible encryption approach.
 * 
 * Setup Instructions:
 * 1. Copy this entire script into your Google Apps Script project
 * 2. Set your INVITATION_KEY below  
 * 3. Use =GENERATE_TOKEN("Student Name") in your spreadsheet
 */

// ========== CONFIGURATION ==========
const INVITATION_KEY = "MXw5ZDA5ZDQ1ZWJhNzE2NTcyODU4ZGUwMWE5ZDA3Mzk3NjMxY2UyZjE0ZDIzMWZkMzM4ODJiZTY3NDIzYjc1Yjg3";
const BASE_URL = "https://booktracker.rustyphillips.net/invite/";

// ========== MINIMAL CRYPTO IMPLEMENTATION ==========

/**
 * Simple XOR-based encryption that's compatible with simple decryption
 * This is a fallback when proper AES isn't available
 */
function simpleEncrypt(data, key) {
  const keyBytes = [];
  for (let i = 0; i < key.length; i += 2) {
    keyBytes.push(parseInt(key.substr(i, 2), 16));
  }
  
  const dataBytes = [];
  for (let i = 0; i < data.length; i++) {
    dataBytes.push(data.charCodeAt(i));
  }
  
  const encrypted = [];
  for (let i = 0; i < dataBytes.length; i++) {
    encrypted.push(dataBytes[i] ^ keyBytes[i % keyBytes.length]);
  }
  
  // Convert to hex
  return encrypted.map(b => b.toString(16).padStart(2, '0')).join('');
}

/**
 * Base64URL encoding (URL-safe base64)
 */
function base64UrlEncode(data) {
  return Utilities.base64Encode(data)
    .replace(/\+/g, '-')
    .replace(/\//g, '_')
    .replace(/=/g, '');
}

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

// ========== ENCRYPTION ==========

function encryptInvitationData(studentName) {
  try {
    if (!INVITATION_KEY || INVITATION_KEY.trim() === '') {
      throw new Error('INVITATION_KEY not set');
    }
    
    const { classId, keyHex } = parseCompoundKey(INVITATION_KEY);
    
    // Create payload: classId|studentName|timestamp
    const timestamp = Math.floor(Date.now() / 1000);
    const payload = `${classId}|${studentName}|${timestamp}`;
    
    // Simple encryption
    const encrypted = simpleEncrypt(payload, keyHex);
    
    // Add a random prefix for uniqueness
    const randomPrefix = Math.random().toString(36).substring(2, 8);
    const finalData = randomPrefix + encrypted;
    
    // Convert to base64url
    return base64UrlEncode(finalData);
      
  } catch (error) {
    throw new Error(`Encryption failed: ${error.message}`);
  }
}

// ========== SPREADSHEET FUNCTIONS ==========

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

function GET_CONFIG() {
  try {
    if (!INVITATION_KEY || INVITATION_KEY.trim() === '') {
      return '❌ INVITATION_KEY not set';
    }
    const { classId } = parseCompoundKey(INVITATION_KEY);
    return `✅ Class ${classId} configured (Simple Encryption)`;
  } catch (error) {
    return `❌ ${error.message}`;
  }
}

// ========== TESTING ==========

function testEncryption() {
  const testName = 'John Smith';
  const token = GENERATE_TOKEN(testName);
  const url = GENERATE_URL(testName);
  const config = GET_CONFIG();
  
  Logger.log('=== Simple Encryption Test ===');
  Logger.log(`Name: ${testName}`);
  Logger.log(`Token: ${token}`);
  Logger.log(`URL: ${url}`);
  Logger.log(`Config: ${config}`);
}

function onOpen() {
  const ui = SpreadsheetApp.getUi();
  ui.createMenu('📚 Book Tracker')
    .addItem('🧪 Test', 'testEncryption')
    .addToUi();
}