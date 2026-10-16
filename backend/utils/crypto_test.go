package utils

import "testing"

// TestEncryptDecryptAPIKeyRoundTrip AES-GCM 加解密往返 + 密文非明文（T-067）
func TestEncryptDecryptAPIKeyRoundTrip(t *testing.T) {
	const secret = "unit-test-secret"
	plain := "sk-abcdef1234567890"

	enc, err := EncryptAPIKey(plain, secret)
	if err != nil {
		t.Fatalf("加密失败: %v", err)
	}
	if enc == plain || enc == "" {
		t.Fatalf("密文不应等于明文或为空: %q", enc)
	}

	got, err := DecryptAPIKey(enc, secret)
	if err != nil {
		t.Fatalf("解密失败: %v", err)
	}
	if got != plain {
		t.Fatalf("往返不一致: 期望 %q 实际 %q", plain, got)
	}
}

// TestEncryptAPIKeyNoSecret 未配置密钥时降级明文（与解密对称）（T-067）
func TestEncryptAPIKeyNoSecret(t *testing.T) {
	plain := "sk-plain"
	enc, err := EncryptAPIKey(plain, "")
	if err != nil {
		t.Fatalf("降级加密失败: %v", err)
	}
	if enc != plain {
		t.Fatalf("无密钥应原样返回，实际 %q", enc)
	}
	got, err := DecryptAPIKey(enc, "")
	if err != nil || got != plain {
		t.Fatalf("无密钥解密应原样返回: got=%q err=%v", got, err)
	}
}

// TestDecryptAPIKeyWrongSecret 错误密钥应解密失败（T-067）
func TestDecryptAPIKeyWrongSecret(t *testing.T) {
	enc, err := EncryptAPIKey("sk-secret-value", "right-key")
	if err != nil {
		t.Fatalf("加密失败: %v", err)
	}
	if _, err := DecryptAPIKey(enc, "wrong-key"); err == nil {
		t.Fatal("错误密钥应返回解密失败")
	}
}

// TestMaskAPIKey 脱敏仅保留末四位（T-067）
func TestMaskAPIKey(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"", ""},
		{"abc", "***"},
		{"sk-abcdef1234", "****1234"},
	}
	for _, c := range cases {
		if got := MaskAPIKey(c.in); got != c.want {
			t.Fatalf("MaskAPIKey(%q)=%q 期望 %q", c.in, got, c.want)
		}
	}
}
