import { describe, it, expect } from 'vitest'
import { Buffer } from 'buffer'
import { isEncryptedXlsx, encryptXlsx, decryptXlsx } from '../xlsxCrypto.js'

globalThis.Buffer = globalThis.Buffer || Buffer

// 最小 zip 结构（xlsx 为 zip 包，测试只需可被 Office 加密容器包裹的字节）
function fakeXlsx(payload = 'sheet-bytes') {
  return new Uint8Array(Buffer.from('PK\x03\x04' + payload, 'binary'))
}

describe('xlsxCrypto（T-074）', () => {
  it('加密后为 Office 加密文件，且可用同一密码解回原字节', async () => {
    const plain = fakeXlsx('roundtrip')
    const encrypted = encryptXlsx(plain, 'S3cret!pass')

    expect(encrypted).toBeInstanceOf(Uint8Array)
    expect(isEncryptedXlsx(encrypted)).toBe(true)
    expect(isEncryptedXlsx(plain)).toBe(false)

    const decrypted = await decryptXlsx(encrypted, 'S3cret!pass')
    expect(Buffer.from(decrypted).equals(Buffer.from(plain))).toBe(true)
  })

  it('未加密文件原样返回（明文 xlsx 可直接导入）', async () => {
    const plain = fakeXlsx('plain')
    expect(Buffer.from(await decryptXlsx(plain, '')).equals(Buffer.from(plain))).toBe(true)
    expect(Buffer.from(await decryptXlsx(plain, 'anything')).equals(Buffer.from(plain))).toBe(true)
  })

  it('缺少密码拒绝加密；加密文件无/错密码拒绝解密', async () => {
    const plain = fakeXlsx()
    expect(() => encryptXlsx(plain, '')).toThrow()

    const encrypted = encryptXlsx(plain, 'right-pass')
    await expect(decryptXlsx(encrypted, '')).rejects.toThrow()
    await expect(decryptXlsx(encrypted, 'wrong-pass')).rejects.toThrow('密码错误')
  })

  it('非二进制输入报错', () => {
    expect(() => encryptXlsx('not-binary', 'pw-123')).toThrow()
  })
})
