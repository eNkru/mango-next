import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import path from 'node:path';

// Fixed asset names keep the Go HTML shell simple. Runtime BaseURL is injected
// by Go; Vite emits relative module URLs under /react/.
export default defineConfig({
  root: path.resolve(__dirname, 'frontend'),
  base: './',
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, 'frontend/src'),
    },
  },
  // Dev-only: npm run dev + Go on :9000. Build/embed path is unchanged.
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
    rollupOptions: {
      input: path.resolve(__dirname, 'frontend/index.html'),
      output: {
        entryFileNames: 'assets/main.js',
        chunkFileNames: 'assets/[name].js',
        assetFileNames: (info) => {
          if (info.name && info.name.endsWith('.css')) {
            return 'assets/main.css';
          }
          return 'assets/[name][extname]';
        },
      },
    },
  },
});
