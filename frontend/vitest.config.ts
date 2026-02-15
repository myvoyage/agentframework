import { defineConfig } from 'vitest/config'
import vue from '@vitejs/plugin-vue'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [vue()],
  test: {
    globals: true,
    environment: 'happy-dom',
    include: ['**/*.test.{ts,tsx,vue}'],
    exclude: ['node_modules', 'dist'],
    setupFiles: ['./setupTests.ts'],
    coverage: {
      provider: 'v8',
      reporter: ['text', 'json', 'html'],
      include: ['src/**/*.{vue,ts,tsx}'],
      exclude: ['node_modules', 'dist', 'src/main.ts', 'src/router/**/*.ts']
    }
  },
})