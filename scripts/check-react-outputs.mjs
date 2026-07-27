#!/usr/bin/env node
/**
 * Fail if the Vite build outputs are missing under go/web/public/react.
 *
 * Assets are content-hashed (vite.config.ts: entryFileNames/chunkFileNames/
 * assetFileNames use [hash]), so we cannot assert fixed filenames. Instead we
 * verify the Vite manifest exists and that the entry JS + CSS files it points
 * to are present on disk. This catches both "forgot to build" and a broken
 * manifest/copy step.
 */
import { access, readFile } from 'node:fs/promises';
import path from 'node:path';

const root = path.resolve(import.meta.dirname, '..');
const reactDir = path.join(root, 'go/web/public/react');
const manifestPath = path.join(reactDir, 'manifest.json');

async function exists(p) {
  try {
    await access(p);
    return true;
  } catch {
    return false;
  }
}

const problems = [];

// 1. Manifest must exist (Vite emits it; vite.config.ts copies it out of .vite/
//    so go:embed public/* picks it up — dotfile dirs are skipped by embed).
if (!(await exists(manifestPath))) {
  problems.push('go/web/public/react/manifest.json (run `npm run build`)');
} else {
  let manifest;
  try {
    manifest = JSON.parse(await readFile(manifestPath, 'utf8'));
  } catch (err) {
    problems.push(`manifest.json is not valid JSON: ${err.message}`);
    manifest = {};
  }

  // 2. The HTML entry ("index.html") carries the entry JS + css[] the Go shell
  //    injects. Verify those files exist on disk.
  const entry = manifest['index.html'];
  if (!entry) {
    problems.push('manifest.json has no "index.html" entry');
  } else {
    const checks = [];
    if (entry.file) checks.push(entry.file);
    if (Array.isArray(entry.css)) checks.push(...entry.css);
    for (const rel of checks) {
      if (!(await exists(path.join(reactDir, rel)))) {
        problems.push(`go/web/public/react/${rel} (referenced by manifest)`);
      }
    }
  }
}

if (problems.length) {
  console.error('Missing or broken React build outputs:\n' + problems.map((m) => `  - ${m}`).join('\n'));
  console.error('Run `npm run build` before make/check/go build.');
  process.exit(1);
}

console.log('React build outputs present.');