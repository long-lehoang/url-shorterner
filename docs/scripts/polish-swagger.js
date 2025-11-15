#!/usr/bin/env node

/**
 * Polish Swagger JSON by removing Go-specific extensions and cleaning up the spec
 */

const fs = require('fs');
const path = require('path');

const inputFile = process.argv[2];
const outputFile = process.argv[3] || inputFile;

if (!inputFile) {
  console.error('Usage: polish-swagger.js <input-file> [output-file]');
  process.exit(1);
}

console.log(`Polishing Swagger JSON: ${inputFile}`);

let swagger = JSON.parse(fs.readFileSync(inputFile, 'utf8'));

// Remove Go-specific extensions
function removeExtensions(obj) {
  if (typeof obj !== 'object' || obj === null) {
    return obj;
  }
  
  if (Array.isArray(obj)) {
    return obj.map(removeExtensions);
  }
  
  const cleaned = {};
  for (const [key, value] of Object.entries(obj)) {
    // Skip Go-specific extensions
    if (key.startsWith('x-go-') || key.startsWith('x-goName') || key.startsWith('x-goType')) {
      continue;
    }
    cleaned[key] = removeExtensions(value);
  }
  
  return cleaned;
}

swagger = removeExtensions(swagger);

// Clean up empty arrays and objects
function cleanEmpty(obj) {
  if (typeof obj !== 'object' || obj === null) {
    return obj;
  }
  
  if (Array.isArray(obj)) {
    return obj.map(cleanEmpty).filter(item => item !== null && item !== undefined);
  }
  
  const cleaned = {};
  for (const [key, value] of Object.entries(obj)) {
    const cleanedValue = cleanEmpty(value);
    if (cleanedValue !== null && cleanedValue !== undefined) {
      if (Array.isArray(cleanedValue) && cleanedValue.length === 0) {
        continue;
      }
      if (typeof cleanedValue === 'object' && Object.keys(cleanedValue).length === 0) {
        continue;
      }
      cleaned[key] = cleanedValue;
    }
  }
  
  return Object.keys(cleaned).length > 0 ? cleaned : undefined;
}

swagger = cleanEmpty(swagger);

// Ensure required fields
if (!swagger.info) {
  swagger.info = {
    title: 'API',
    version: '1.0.0'
  };
}

if (!swagger.paths) {
  swagger.paths = {};
}

if (!swagger.definitions) {
  swagger.definitions = {};
}

fs.writeFileSync(outputFile, JSON.stringify(swagger, null, 2));
console.log(`✓ Polished Swagger JSON written to: ${outputFile}`);

