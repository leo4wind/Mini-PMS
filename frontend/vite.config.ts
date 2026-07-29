import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { fileURLToPath, URL } from 'node:url'
import http from 'node:http'

// 复用到后端的 TCP 连接，减少 Vite→API 反复建连
const apiAgent = new http.Agent({
  keepAlive: true,
  maxSockets: 10,
  keepAliveMsecs: 30_000,
})

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  server: {
    // 固定 IPv4，避免 Windows 上 localhost→::1 回退导致建连抖动
    host: '127.0.0.1',
    port: 5173,
    hmr: {
      host: '127.0.0.1',
    },
    proxy: {
      '/api': {
        target: 'http://127.0.0.1:8088',
        changeOrigin: true,
        agent: apiAgent,
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
