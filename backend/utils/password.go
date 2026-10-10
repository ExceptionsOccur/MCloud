package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

const (
	MaxFailedAttempts = 5
	LockoutMinutes    = 15
)

// bcryptCost bcrypt 计算成本（10 为默认强度）
const bcryptCost = 10

// FormatPasswordHash 生成 bcrypt 存储串（T-009 起新格式，`$2a$…`）
func FormatPasswordHash(password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return ""
	}
	return string(hash)
}

// IsLegacyHash 判断存储值是否为旧 SHA-256 `salt$hash` 格式（bcrypt 均以 `$2` 前缀）
func IsLegacyHash(stored string) bool {
	return !strings.HasPrefix(stored, "$2")
}

// VerifyPassword 校验密码：bcrypt 格式优先；旧 SHA-256 `salt$hash` 平滑兼容
// （登录成功后由 auth_service 升级为 bcrypt，T-009）
func VerifyPassword(password, stored string) bool {
	if !IsLegacyHash(stored) {
		return bcrypt.CompareHashAndPassword([]byte(stored), []byte(password)) == nil
	}
	salt, expectedHash := ParsePasswordHash(stored)
	if salt == "" {
		return false
	}
	return HashPassword(password, salt) == expectedHash
}

// HashPassword 旧格式哈希：SHA-256(salt + password)（仅存量校验与测试使用）
func HashPassword(password, salt string) string {
	hash := sha256.Sum256([]byte(salt + password))
	return hex.EncodeToString(hash[:])
}

func GenerateSalt() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func ParsePasswordHash(stored string) (salt, hash string) {
	parts := strings.SplitN(stored, "$", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return "", stored
}
