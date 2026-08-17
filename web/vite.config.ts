import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

const reviewApi = 'http://localhost:8080'

export default defineConfig(({ command }) => ({
  base: './',
  plugins: [vue({ script: { defineModel: true } })],
  server: {
    strictPort: true,
    proxy: { '/api': { target: reviewApi, changeOrigin: true } },
  },
  build: {
    target: 'es2022',
    sourcemap: command === 'serve',
    rollupOptions: { output: { manualChunks: { reviewCore: ['vue', 'pinia', 'vue-router'] } } },
  },
}))
