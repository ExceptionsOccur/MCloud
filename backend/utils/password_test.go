package utils

import (
	"strings"
	"testing"
)

func TestHashPasswordRoundtrip(t *testing.T) {
	tests := []struct {
		name     string
		password string
	}{
		{name: "普通密码", password: "Pass4MCloud"},
		{name: "含特殊字符", password: "p@ss$w0rd!#"},
		{name: "中文密码", password: "密码123"},
		{name: "空密码", password: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stored := FormatPasswordHash(tt.password)
			if !strings.Contains(stored, "$") {
				t.Fatalf("存储格式应为 salt$hash，实际 %q", stored)
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

func TestVerifyPasswordLegacyFormats(t *testing.T) {
	tests := []struct {
		name   string
		stored string
	}{
		{name: "无盐裸哈希拒绝", stored: "d7a8fbb307d7809469ca9abcb0082e4f8d5651e46d3cdb762d02d0bf37c9e592"},
		{name: "空串拒绝", stored: ""},
		{name: "盐缺失格式拒绝", stored: "$onlyhashpart"},
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

func TestParsePasswordHash(t *testing.T) {
	tests := []struct {
		name     string
		stored   string
		wantSalt string
		wantHash string
	}{
		{name: "标准 salt$hash", stored: "abc$def", wantSalt: "abc", wantHash: "def"},
		{name: "无盐", stored: "barehash", wantSalt: "", wantHash: "barehash"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			salt, hash := ParsePasswordHash(tt.stored)
			if salt != tt.wantSalt || hash != tt.wantHash {
				t.Fatalf("got (%q,%q)，期望 (%q,%q)", salt, hash, tt.wantSalt, tt.wantHash)
			}
		})
	}
}
