package services

import (
	"testing"

	"mcloud/config"
)

// useAISecretKey 注入测试用 AI_SECRET_KEY 配置（T-067）
func useAISecretKey(t *testing.T, secret string) {
	t.Helper()
	prev := config.AppConfig
	config.AppConfig = &config.Config{AISecretKey: secret}
	t.Cleanup(func() { config.AppConfig = prev })
}

// TestAISettingMaskedAndPreserveKey 保存后回显脱敏、空 key 保留原值（T-067）
func TestAISettingMaskedAndPreserveKey(t *testing.T) {
	newTestDB(t)
	useAISecretKey(t, "svc-test-secret")
	svc := NewAISettingService()

	// 首次保存（含 key）
	if err := svc.Save(SaveAISettingRequest{
		APIURL: "https://api.example.com/v1",
		APIKey: "sk-abcdef1234",
		Model:  "gpt-4o-mini",
	}); err != nil {
		t.Fatalf("Save 失败: %v", err)
	}

	got, err := svc.Get()
	if err != nil {
		t.Fatalf("Get 失败: %v", err)
	}
	if !got.APIKeySet || got.APIKeyMasked != "****1234" {
		t.Fatalf("期望回显脱敏且已设置，实际 set=%v masked=%q", got.APIKeySet, got.APIKeyMasked)
	}
	if got.APIURL != "https://api.example.com/v1" || got.Model != "gpt-4o-mini" {
		t.Fatalf("配置回读不一致: %+v", got)
	}

	// 二次保存：空 key 应保留原值
	if saveErr := svc.Save(SaveAISettingRequest{
		APIURL: "https://api.example.com/v2",
		APIKey: "",
		Model:  "gpt-4o",
	}); saveErr != nil {
		t.Fatalf("二次 Save 失败: %v", saveErr)
	}

	got2, err := svc.Get()
	if err != nil {
		t.Fatalf("二次 Get 失败: %v", err)
	}
	if !got2.APIKeySet || got2.APIKeyMasked != "****1234" {
		t.Fatalf("空 key 应保留原值，实际 set=%v masked=%q", got2.APIKeySet, got2.APIKeyMasked)
	}
	if got2.APIURL != "https://api.example.com/v2" || got2.Model != "gpt-4o" {
		t.Fatalf("URL/模型应更新: %+v", got2)
	}

	// 内部明文可解密（GetDecrypted）
	_, plain, model, err := svc.GetDecrypted()
	if err != nil {
		t.Fatalf("GetDecrypted 失败: %v", err)
	}
	if plain != "sk-abcdef1234" || model != "gpt-4o" {
		t.Fatalf("明文解密不一致: key=%q model=%q", plain, model)
	}
}

// TestAISettingEmptyBeforeSave 未配置时 Get 返回空且 key 未设置（T-067）
func TestAISettingEmptyBeforeSave(t *testing.T) {
	newTestDB(t)
	useAISecretKey(t, "svc-test-secret")
	svc := NewAISettingService()

	got, err := svc.Get()
	if err != nil {
		t.Fatalf("Get 失败: %v", err)
	}
	if got.APIKeySet || got.APIKeyMasked != "" || got.APIURL != "" {
		t.Fatalf("空配置期望全空，实际 %+v", got)
	}
}

// TestAISettingValidate 必填校验失败（T-067）
func TestAISettingValidate(t *testing.T) {
	newTestDB(t)
	useAISecretKey(t, "svc-test-secret")
	svc := NewAISettingService()

	if err := svc.Save(SaveAISettingRequest{APIURL: "", Model: "m"}); err == nil {
		t.Fatal("空 API 地址应报错")
	}
	if err := svc.Save(SaveAISettingRequest{APIURL: "https://x", Model: ""}); err == nil {
		t.Fatal("空模型应报错")
	}
}
