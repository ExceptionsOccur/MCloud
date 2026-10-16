package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"log"
	"strings"
)

// AI API Key 加密：AES-256-GCM，密钥由环境变量 AI_SECRET_KEY 经 SHA-256 派生（定长 32 字节）。
// 未配置 AI_SECRET_KEY 时降级明文存储并告警（不阻断功能，T-067）。

// deriveAESKey 将任意长度密钥经 SHA-256 派生为 32 字节 AES-256 密钥
func deriveAESKey(secret string) [32]byte {
	return sha256.Sum256([]byte(secret))
}

// deriveAESKeySlice 返回可切片的派生密钥副本
func deriveAESKeySlice(secret string) []byte {
	k := deriveAESKey(secret)
	return k[:]
}

// EncryptAPIKey 加密明文 API Key，返回 base64(nonce||ciphertext)。
// 未配置密钥时返回明文与 nil（降级），由调用方记录告警。
func EncryptAPIKey(plain, secret string) (string, error) {
	if plain == "" {
		return "", nil
	}
	if secret == "" {
		return plain, nil
	}

	block, err := aes.NewCipher(deriveAESKeySlice(secret))
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	sealed := gcm.Seal(nonce, nonce, []byte(plain), nil)
	return base64.StdEncoding.EncodeToString(sealed), nil
}

// DecryptAPIKey 解密 EncryptAPIKey 的产物。
// 未配置密钥时按明文原样返回（与加密降级对称）；密文非法返回错误。
func DecryptAPIKey(enc, secret string) (string, error) {
	if enc == "" {
		return "", nil
	}
	if secret == "" {
		return enc, nil
	}

	raw, err := base64.StdEncoding.DecodeString(enc)
	if err != nil {
		return "", errors.New("API Key 密文格式非法")
	}

	block, err := aes.NewCipher(deriveAESKeySlice(secret))
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(raw) < gcm.NonceSize() {
		return "", errors.New("API Key 密文长度非法")
	}

	nonce, ciphertext := raw[:gcm.NonceSize()], raw[gcm.NonceSize():]
	plain, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", errors.New("API Key 解密失败")
	}
	return string(plain), nil
}

// MaskAPIKey 脱敏展示：保留末四位，其余以 * 代替；空串返回空。
func MaskAPIKey(plain string) string {
	if plain == "" {
		return ""
	}
	if len(plain) <= 4 {
		return strings.Repeat("*", len(plain))
	}
	return "****" + plain[len(plain)-4:]
}

// EncryptAPIKeyWithWarning 供 service 使用的封装：无密钥时记录一次告警后降级明文
func EncryptAPIKeyWithWarning(plain, secret string) (string, error) {
	if secret == "" {
		log.Println("[warn] 未配置 AI_SECRET_KEY，API Key 将以明文存储")
	}
	return EncryptAPIKey(plain, secret)
}
