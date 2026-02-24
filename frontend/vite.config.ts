import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueDevTools from 'vite-plugin-vue-devtools'

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    vue(),
    vueDevTools(),
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    },
  },
  server: {
    proxy: {
      '/api/v1/auth': { target: 'http://localhost:8080', changeOrigin: true },
      '/api/v1/users': { target: 'http://localhost:8081', changeOrigin: true },
      '/api/v1/roles': { target: 'http://localhost:8081', changeOrigin: true },
      '/api/v1/warehouses': { target: 'http://localhost:8082', changeOrigin: true },
      '/api/v1/items': { target: 'http://localhost:8082', changeOrigin: true },
      '/api/v1/alerts': { target: 'http://localhost:8082', changeOrigin: true },
      '/api/v1/requests': { target: 'http://localhost:8083', changeOrigin: true },
      '/api/v1/shipments': { target: 'http://localhost:8083', changeOrigin: true },
      '/api/v1/stats': { target: 'http://localhost:8084', changeOrigin: true },
    }
  }
})
