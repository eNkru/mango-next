#!/usr/bin/env node
/** Fail if declared React build outputs are missing under go/web/public/react. */
import { access } from 'node:fs/promises';
import path from 'node:path';

const root = path.resolve(import.meta.dirname, '..');
const required = [
  'go/web/public/react/assets/main.js',
  'go/web/public/react/assets/main.css',
];

const missing = [];
for (const rel of required) {
  try {
    await access(path.join(root, rel));
  } catch {
    missing.push(rel);
  }
}

if (missing.length) {
  console.error('Missing React build outputs:\n' + missing.map((m) => `  - ${m}`).join('\n'));
  console.error('Run `npm run build` before make/check/go build.');
  process.exit(1);
}

console.log('React build outputs present.');
