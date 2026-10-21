// 浏览器端全局 polyfill（必须在任何依赖 Buffer 的库之前加载）
//
// officecrypto-tool / cfb / crypto-js（数据备份 xlsx 加解密，T-074）在浏览器中依赖
// Node 的全局 Buffer；这些库在模块求值时就会判断 `typeof Buffer`，因此注入必须发生在
// 主 bundle 之前，不能写在 main.js 的应用逻辑里（同一 chunk 内顺序无法保证）。

import { Buffer } from 'buffer'

globalThis.Buffer = globalThis.Buffer || Buffer
