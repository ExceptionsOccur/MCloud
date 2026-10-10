package utils

import (
	"strings"
	"testing"
)

func TestFormatPasswordHashBcrypt(t *testing.T) {
	tests := []struct {
		name     string
		password string
	}{
		{name: "普通密码", password: "Pass4MCloud"},
		{name: "含特殊字符", password: "p@ss$w0rd!#"},
		{name: "中文密码", password: "密码123"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stored := FormatPasswordHash(tt.password)
			if !strings.HasPrefix(stored, "$2") {
				t.Fatalf("应为 bcrypt 格式（$2 前缀），实际 %q", stored)
			}
			if IsLegacyHash(stored) {
				t.Fatal("bcrypt 格式不应判为旧格式")
			}
			if !VerifyPassword(tt.password, stored) {
				t.Fatal("正确密码校验失败")
			}
			if VerifyPassword(tt.password+"x", stored) {
				t.Fatal("错误密码不应通过")
			}
		})
	}
}

func TestFormatPasswordHashEmptyPassword(t *testing.T) {
	stored := FormatPasswordHash("")
	if !strings.HasPrefix(stored, "$2") {
		t.Fatalf("空密码也应生成 bcrypt 串，实际 %q", stored)
	}
	if VerifyPassword("x", stored) {
		t.Fatal("错误密码不应通过")
	}
}

// TestVerifyPasswordLegacySHA256 旧 SHA-256 salt$hash 存量格式仍可校验（T-009 平滑迁移前提）
func TestVerifyPasswordLegacySHA256(t *testing.T) {
	tests := []struct {
		name     string
		password string
	}{
		{name: "常规密码", password: "legacy-pass-123"},
		{name: "默认管理员密码", password: "Pass4MCloud"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			salt := GenerateSalt()
			legacy := salt + "$" + HashPassword(tt.password, salt)
			if !IsLegacyHash(legacy) {
				t.Fatal("salt$hash 应判为旧格式")
			}
			if !VerifyPassword(tt.password, legacy) {
				t.Fatal("旧格式正确密码校验失败")
			}
			if VerifyPassword(tt.password+"x", legacy) {
				t.Fatal("旧格式错误密码不应通过")
			}
		})
	}
}

func TestVerifyPasswordInvalidStored(t *testing.T) {
	tests := []struct {
		name   string
		stored string
	}{
		{name: "空串拒绝", stored: ""},
		{name: "无盐裸哈希拒绝", stored: "d7a8fbb307d7809469ca9abcb0082e4f8d5651e46d3cdb762d02d0bf37c9e592"},
		{name: "损坏的 bcrypt 串拒绝", stored: "$2a$10$notavalidhash"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if VerifyPassword("any", tt.stored) {
				t.Fatalf("非法存储格式不应通过: %q", tt.stored)
			}
		})
	}
}

func TestGenerateSaltUnique(t *testing.T) {
	a, b := GenerateSalt(), GenerateSalt()
	if a == "" || b == "" {
		t.Fatal("盐不应为空")
	}
	if a == b {
		t.Fatal("两次生成的盐应不同")
	}
	if len(a) != 32 {
		t.Fatalf("盐长度应为 32 hex 字符，实际 %d", len(a))
	}
}
