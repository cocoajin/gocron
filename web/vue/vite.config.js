import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
export default defineConfig({ plugins: [vue()], base: '/public/', server: { port: 8080, proxy: { '/api': 'http://127.0.0.1:5920' } }, build: { outDir: 'dist', assetsDir: 'static', sourcemap: false } })
