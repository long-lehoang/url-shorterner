#!/usr/bin/env node

/**
 * Apply JSON patch files to OpenAPI spec
 */

const fs = require('fs');
const path = require('path');

const inputFile = process.argv[2];
const patchesDir = process.argv[3] || path.join(__dirname, '../patches');
const outputFile = process.argv[4] || inputFile;

if (!inputFile) {
  console.error('Usage: apply-patches.js <input-file> [patches-dir] [output-file]');
  process.exit(1);
}

console.log(`Applying patches to: ${inputFile}`);

// Load dependencies (installed via npm install -g in Docker)
let yaml, applyPatch;
try {
  yaml = require('js-yaml');
  applyPatch = require('fast-json-patch').applyPatch;
} catch (e) {
  console.error('Error: Missing dependencies. These should be installed in Docker container.');
  console.error('Required: js-yaml, fast-json-patch');
  process.exit(1);
}

let spec = JSON.parse(fs.readFileSync(inputFile, 'utf8'));

// Read and apply all patch files
const patchFiles = fs.readdirSync(patchesDir)
  .filter(file => file.endsWith('.patch.yaml') || file.endsWith('.patch.json') || file.endsWith('.patch.yml'))
  .sort();

for (const patchFile of patchFiles) {
  const patchPath = path.join(patchesDir, patchFile);
  console.log(`  Applying patch: ${patchFile}`);
  
  let patches;
  if (patchFile.endsWith('.yaml') || patchFile.endsWith('.yml')) {
    patches = yaml.load(fs.readFileSync(patchPath, 'utf8'));
  } else {
    patches = JSON.parse(fs.readFileSync(patchPath, 'utf8'));
  }
  
  try {
    const result = applyPatch(spec, patches, true, false);
    if (result.length > 0) {
      spec = result[0].newDocument;
    }
  } catch (error) {
    console.error(`  Error applying patch ${patchFile}:`, error.message);
    // Try to continue with other patches
    if (error.name === 'OPERATION_PATH_UNRESOLVABLE') {
      console.error(`  Warning: Path does not exist. Trying to use 'add' operation instead...`);
      // Convert replace operations to add operations for missing paths
      const modifiedPatches = patches.map(patch => {
        if (patch.op === 'replace' && error.operation && error.operation.path === patch.path) {
          return { ...patch, op: 'add' };
        }
        return patch;
      });
      try {
        const retryResult = applyPatch(spec, modifiedPatches, true, false);
        if (retryResult.length > 0) {
          spec = retryResult[0].newDocument;
          console.log(`  ✓ Retried with 'add' operation`);
        }
      } catch (retryError) {
        console.error(`  Failed to apply patch even with 'add' operation:`, retryError.message);
        throw retryError;
      }
    } else {
      throw error;
    }
  }
}

fs.writeFileSync(outputFile, JSON.stringify(spec, null, 2));
console.log(`✓ Patches applied. Output written to: ${outputFile}`);
