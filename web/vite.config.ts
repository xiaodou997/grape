import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'

// https://vite.dev/config/
export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src'),
    },
  },
  server: {
    port: 3000,
    proxy: {
      '/-/api': {
        target: 'http://localhost:4873',
        changeOrigin: true,
      },
      // Proxy npm registry requests to Grape
      '^/(?!@vite|@fs|node_modules|src)': {
        target: 'http://localhost:4873',
        changeOrigin: true,
      },
    },
  },
  build: {
    outDir: 'dist',
    emptyOutDir: true,
  },
})