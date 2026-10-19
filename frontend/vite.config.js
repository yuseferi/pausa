import { defineConfig } from 'vitest/config'
import vue from '@vitejs/plugin-vue'

// Shared config for both `vite build` and `vitest`. The `test` block is
// ignored by the Vite build; Vitest reads it.
export default defineConfig({
  plugins: [vue()],
  build: {
    outDir: 'dist',
  },
  server: {
    strictPort: true,
  },
  test: {
    environment: 'jsdom',
    include: ['src/**/*.{test,spec}.{js,vue}'],
    setupFiles: ['./vitest.setup.js'],
  },
})
