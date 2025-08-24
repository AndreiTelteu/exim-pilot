import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  build: {
    outDir: 'dist',
    assetsDir: 'assets',
    sourcemap: false,
    minify: 'terser',
    rollupOptions: {
      output: {
        // Optimize chunk splitting for better caching
        manualChunks: {
          vendor: ['react', 'react-dom'],
          charts: ['echarts', 'echarts-for-react'],
          router: ['react-router-dom'],
          utils: ['react-window', 'react-window-infinite-loader'],
        },
        // Use consistent naming for better caching
        entryFileNames: 'assets/[name]-[hash].js',
        chunkFileNames: 'assets/[name]-[hash].js',
        assetFileNames: 'assets/[name]-[hash].[ext]',
      },
    },
    // Optimize for production embedding
    chunkSizeWarningLimit: 1000,
    assetsInlineLimit: 4096, // Inline small assets as base64
    reportCompressedSize: true,
    // Enable tree shaking and other optimizations
    target: 'es2020',
    cssCodeSplit: true,
  },
  server: {
    port: 3000,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/ws': {
        target: 'ws://localhost:8080',
        ws: true,
      },
    },
  },
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  // Optimize dependencies
  optimizeDeps: {
    include: ['react', 'react-dom', 'echarts', 'echarts-for-react'],
  },
})
