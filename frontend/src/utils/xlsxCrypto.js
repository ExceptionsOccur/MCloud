import officeCrypto from 'officecrypto-tool'

// 数据备份 xlsx 加密 / 解密（T-074）
//
// 使用 ECMA-376 标准 Office 加密（Agile）生成「加密 Excel」：
// 用户仍可用 Excel/WPS 输入密码打开、另存修改后再上传，而不是自定义加密容器。
// 浏览器端 Buffer 由 main.js 注入（npm 包 buffer 的全局 polyfill）。

const PASSWORD_MAX_LENGTH = 255

// asBuffer 统一把 ArrayBuffer / TypedArray / Buffer 转成 Buffer
function asBuffer(input) {
  if (Buffer.isBuffer(input)) return input
  if (ArrayBuffer.isView(input) || input instanceof ArrayBuffer) return Buffer.from(input)
  throw new Error('输入必须为二进制数据')
}

// OLE 复合文档 magic（Office 加密文件的容器头）
const OLE_MAGIC = [0xd0, 0xcf, 0x11, 0xe0, 0xa1, 0xb1, 0x1a, 0xe1]

// isEncryptedXlsx 判断是否为 Office 加密的 Excel。
// 加密文件是 OLE 复合文档（d0cf11e0...），普通 xlsx 是 zip（50 4b 03 04）；
// 直接调用 officeCrypto.isEncrypted 遇到普通 zip 会抛错，故此处先按 magic 判断。
export function isEncryptedXlsx(input) {
  const buf = asBuffer(input)
  if (buf.length < OLE_MAGIC.length) return false
  const magic = OLE_MAGIC.every((b, i) => buf[i] === b)
  if (!magic) return false
  try {
    return officeCrypto.isEncrypted(buf)
  } catch {
    return false
  }
}

// encryptXlsx 用密码加密 xlsx 字节，返回加密后的 xlsx 字节（ArrayBuffer 视图）
export function encryptXlsx(input, password) {
  if (!password) throw new Error('导出必须设置密码')
  if (password.length > PASSWORD_MAX_LENGTH) {
    throw new Error(`密码长度不能超过 ${PASSWORD_MAX_LENGTH} 位`)
  }
  const encrypted = officeCrypto.encrypt(asBuffer(input), { password })
  return toUint8Array(encrypted)
}

// decryptXlsx 用密码解开加密 xlsx，返回明文 xlsx 字节；密码错误抛错。
// 注意：Office Agile 加密的解密 API 返回 Promise，此处统一 await，调用方直接 await 本函数。
export async function decryptXlsx(input, password) {
  const data = asBuffer(input)
  if (!isEncryptedXlsx(data)) return toUint8Array(data)
  if (!password) throw new Error('该文件已加密，请输入打开密码')
  let decrypted
  try {
    decrypted = await officeCrypto.decrypt(data, { password })
  } catch (err) {
    throw new Error(err?.message === 'The password is incorrect' ? '密码错误' : err?.message || '解密失败')
  }
  return toUint8Array(decrypted)
}

// toUint8Array 输出与 Blob 兼容的字节视图
function toUint8Array(data) {
  const buf = asBuffer(data)
  return new Uint8Array(buf.buffer, buf.byteOffset, buf.byteLength)
}
