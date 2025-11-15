#!/usr/bin/env node

/**
 * Polish OpenAPI JSON by sorting schema properties and applying x-order annotations
 */

const fs = require('fs');
const path = require('path');

const inputFile = process.argv[2];
const outputFile = process.argv[3] || inputFile;

if (!inputFile) {
  console.error('Usage: polish-openapi.js <input-file> [output-file]');
  process.exit(1);
}

console.log(`Polishing OpenAPI JSON: ${inputFile}`);

let openapi = JSON.parse(fs.readFileSync(inputFile, 'utf8'));

// Sort schema properties based on x-order annotation
function sortSchemaProperties(schema) {
  if (!schema || typeof schema !== 'object') {
    return schema;
  }
  
  if (schema.properties) {
    const sorted = {};
    const entries = Object.entries(schema.properties);
    
    // Sort by x-order if present, otherwise maintain order
    entries.sort((a, b) => {
      const orderA = a[1]['x-order'] ?? 999;
      const orderB = b[1]['x-order'] ?? 999;
      return orderA - orderB;
    });
    
    entries.forEach(([key, value]) => {
      sorted[key] = sortSchemaProperties(value);
    });
    
    schema.properties = sorted;
  }
  
  // Recursively process nested schemas
  if (schema.allOf) {
    schema.allOf = schema.allOf.map(sortSchemaProperties);
  }
  if (schema.oneOf) {
    schema.oneOf = schema.oneOf.map(sortSchemaProperties);
  }
  if (schema.anyOf) {
    schema.anyOf = schema.anyOf.map(sortSchemaProperties);
  }
  if (schema.items) {
    schema.items = sortSchemaProperties(schema.items);
  }
  
  return schema;
}

// Process all components/schemas
if (openapi.components && openapi.components.schemas) {
  for (const [name, schema] of Object.entries(openapi.components.schemas)) {
    openapi.components.schemas[name] = sortSchemaProperties(schema);
  }
}

// Process request/response schemas in paths
if (openapi.paths) {
  for (const [path, methods] of Object.entries(openapi.paths)) {
    for (const [method, operation] of Object.entries(methods)) {
      if (operation.requestBody && operation.requestBody.content) {
        for (const [contentType, content] of Object.entries(operation.requestBody.content)) {
          if (content.schema) {
            content.schema = sortSchemaProperties(content.schema);
          }
        }
      }
      if (operation.responses) {
        for (const [status, response] of Object.entries(operation.responses)) {
          if (response.content) {
            for (const [contentType, content] of Object.entries(response.content)) {
              if (content.schema) {
                content.schema = sortSchemaProperties(content.schema);
              }
            }
          }
        }
      }
    }
  }
}

fs.writeFileSync(outputFile, JSON.stringify(openapi, null, 2));
console.log(`✓ Polished OpenAPI JSON written to: ${outputFile}`);

