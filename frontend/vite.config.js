import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { nodePolyfills } from 'vite-plugin-node-polyfills'

export default defineConfig({
  // officecrypto-tool（数据备份 xlsx 加解密，T-074）在浏览器端依赖 Node 的 buffer / crypto / stream。
  // crypto 必须做真实 polyfill：Vite 默认把它替换为空模块，调用 randomBytes 会直接抛错。
  plugins: [
    vue(),
    nodePolyfills({
      include: ['buffer', 'crypto', 'stream', 'util', 'process'],
      globals: { Buffer: true, global: true, process: true }
    })
  ],
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://backend:5677',
        changeOrigin: true,
        ws: true
      }
    }
  }
})
