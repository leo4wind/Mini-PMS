import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { fileURLToPath, URL } from 'node:url'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://127.0.0.1:8088',
        changeOrigin: true,
        configure(proxy) {
          proxy.on('error', (err, _req, res) => {
            console.error('[vite proxy]', err.message)
            // http-proxy 的 res 可能是 ServerResponse 或 Socket
            if (res && 'writeHead' in res && !res.headersSent) {
              res.writeHead(502, { 'Content-Type': 'application/json; charset=utf-8' })
              res.end(
                JSON.stringify({
                  code: 502,
                  message: '后端不可用，请确认 API 已启动（默认 :8088）',
                  detail: err.message,
                }),
              )
            }
          })
        },
      },
    },
  },
})
