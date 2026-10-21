#!/usr/bin/env bash
# 校验前端打包产物中 Node polyfill 的真实可用性（T-076）
#
# 背景：lint / vitest / vite build 都不覆盖「运行时 polyfill」。
# T-074 的 xlsx 加密在浏览器里曾因 Buffer 未注入、crypto 被替换为空模块而崩溃，
# 而门禁全绿。本脚本用无全局 Buffer 的沙箱执行产物，验证加解密链路。
#
# 用法：bash scripts/check_frontend_bundle.sh
set -uo pipefail
export LC_ALL=C.UTF-8

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
FE="$ROOT/frontend"
TMPDIR_E2E="$(mktemp -d)"
trap 'rm -rf "$TMPDIR_E2E"' EXIT

fail() { echo "ERROR: $*" >&2; exit 1; }

mapfile -t built < <(ls "$FE"/dist/assets/index-*.js 2>/dev/null)
[[ ${#built[@]} -gt 0 ]] || fail "缺少打包产物，请先 cd frontend && npm run build"

# 1) 产物静态检查：空 crypto 桩出现即说明 Node crypto 未被 polyfill
main_js="$(ls "$FE"/dist/assets/index-*.js | head -1)"
grep -q "__viteBrowserExternal" "$main_js" && fail "打包产物中 crypto 被替换为空模块（需要 vite-plugin-node-polyfills）"
grep -q "globalThis.Buffer=globalThis.Buffer||" "$main_js" || fail "打包产物缺少 Buffer 全局注入"

# 2) 运行时检查：以独立 IIFE 打包 polyfills + xlsxCrypto，在无 Buffer 的沙箱内跑加解密
cat > "$FE/e2e-entry.js" <<'ENTRY'
import './src/polyfills'
import { encryptXlsx, decryptXlsx, isEncryptedXlsx } from './src/utils/xlsxCrypto'

globalThis.__e2e = async function run() {
  const plain = new Uint8Array([0x50, 0x4b, 0x03, 0x04, 0x68, 0x69])
  const enc = encryptXlsx(plain, 'pw-123456')
  const dec = await decryptXlsx(enc, 'pw-123456')
  return {
    bufferType: typeof globalThis.Buffer,
    encrypted: isEncryptedXlsx(enc),
    same: JSON.stringify([...dec]) === JSON.stringify([...plain])
  }
}
ENTRY

cat > "$FE/vite.e2e.config.mjs" <<'CONF'
import { defineConfig } from 'vite'
import { nodePolyfills } from 'vite-plugin-node-polyfills'

export default defineConfig({
  plugins: [nodePolyfills({ include: ['buffer', 'crypto', 'stream', 'util', 'process'], globals: { Buffer: true, global: true, process: true } })],
  build: {
    outDir: 'dist-e2e',
    lib: { entry: 'e2e-entry.js', formats: ['iife'], name: 'E2E' },
    minify: false,
    rollupOptions: { output: { inlineDynamicImports: true } }
  }
})
CONF

runner="$TMPDIR_E2E/run.mjs"
cat > "$runner" <<'RUN'
import fs from 'node:fs'
import vm from 'node:vm'

const bundle = process.argv[2]
const code = fs.readFileSync(bundle, 'utf8')
// 干净沙箱：不提供全局 Buffer / process / require，仅模拟浏览器 WebCrypto
const sandbox = { console }
sandbox.globalThis = sandbox
sandbox.crypto = {
  getRandomValues: (arr) => {
    for (let i = 0; i < arr.length; i++) arr[i] = Math.floor(Math.random() * 256)
    return arr
  },
  subtle: {}
}
const ctx = vm.createContext(sandbox)
vm.runInContext(code, ctx, { filename: 'iife.js' })
const res = await vm.runInContext('__e2e()', ctx)
console.log('bundle e2e result:', res)
if (res.bufferType !== 'function' || !res.encrypted || !res.same) {
  console.error('ERROR: 浏览器环境加解密不可用')
  process.exit(1)
}
RUN

(
  cd "$FE" || exit 1
  npx vite build -c vite.e2e.config.mjs >/dev/null 2>&1 || exit 1
  node "$runner" "$FE/dist-e2e/mcloud-frontend.iife.js"
)
status=$?
rm -f "$FE/e2e-entry.js" "$FE/vite.e2e.config.mjs"
rm -rf "$FE/dist-e2e"
[[ $status -eq 0 ]] || fail "打包产物在无全局 Buffer 的浏览器环境中加解密失败"
echo "check_frontend_bundle: 0 errors"
