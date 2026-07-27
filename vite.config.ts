import { defineConfig, type Plugin } from 'vite';
import react from '@vitejs/plugin-react';
import path from 'node:path';
import { copyFileSync, existsSync, mkdirSync } from 'node:fs';

// Vite writes the manifest to <outDir>/.vite/manifest.json, but Go's
// `//go:embed public/*` skips dotfile directories, so .vite/ is not embedded.
// This plugin copies it to <outDir>/manifest.json (embeddable as react/manifest.json)
// after the build writes it. Runs in closeBundle so the file exists by then.
function copyManifestToOutDir(outDir: string): Plugin {
  return {
    name: 'mango-copy-manifest',
    apply: 'build',
    closeBundle() {
      const src = path.join(outDir, '.vite', 'manifest.json');
      if (!existsSync(src)) return;
      const destDir = path.join(outDir);
      mkdirSync(destDir, { recursive: true });
      copyFileSync(src, path.join(destDir, 'manifest.json'));
    },
  };
}

// Content-hashed asset names: the Go HTML shell resolves them at runtime from
// the Vite manifest (manifest.json), so a Mango upgrade can never serve a
// stale main.js / main.css from a browser or CDN cache. Runtime BaseURL is
// injected by Go; Vite emits relative module URLs under /react/.
export default defineConfig({
  root: path.resolve(__dirname, 'frontend'),
  base: './',
  plugins: [react(), copyManifestToOutDir(path.resolve(__dirname, 'go/web/public/react'))],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, 'frontend/src'),
    },
  },
  // Dev-only: npm run dev + Go on :9000. Build/embed path is unchanged.
  // Vite already SPA-fallbacks unknown paths to index.html; pageId from URL in boot.ts.
  server: {
    proxy: {
      '/api': { target: 'http://127.0.0.1:9000', changeOrigin: true },
      '/img': { target: 'http://127.0.0.1:9000', changeOrigin: true },
    },
  },
  build: {
    outDir: path.resolve(__dirname, 'go/web/public/react'),
    emptyOutDir: true,
    assetsDir: 'assets',
    // Emit manifest.json mapping logical entry → hashed filename. The Go shell
    // reads it (web embed) and injects the hashed URLs into react-shell.tmpl.
    manifest: true,
    rollupOptions: {
      input: path.resolve(__dirname, 'frontend/index.html'),
      output: {
        entryFileNames: 'assets/[name]-[hash].js',
        chunkFileNames: 'assets/[name]-[hash].js',
        assetFileNames: 'assets/[name]-[hash][extname]',
      },
    },
  },
});
