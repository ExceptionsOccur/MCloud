package services

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"mcloud/config"
)

// useAIConfig 同时注入加密密钥与上传目录（T-068）
func useAIConfig(t *testing.T, secret, uploadDir string) {
	t.Helper()
	prev := config.AppConfig
	config.AppConfig = &config.Config{AISecretKey: secret, AIUploadDir: uploadDir}
	t.Cleanup(func() { config.AppConfig = prev })
}

// fakeLLMServer 返回固定 content 的 OpenAI 兼容 /chat/completions
func fakeLLMServer(t *testing.T, content string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			http.NotFound(w, r)
			return
		}
		resp := map[string]interface{}{
			"choices": []map[string]interface{}{
				{"message": map[string]string{"content": content}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
}

// TestExtractOrchestration 端到端编排：解析→LLM→归一化+文件暂存（T-068）
func TestExtractOrchestration(t *testing.T) {
	newTestDB(t)
	useAIConfig(t, "svc-test-secret", t.TempDir())

	llmContent := `{"doc_type":"host_apply","fields":{"name":"web-01","private_ip":"10.0.0.5"},"warnings":["缺少申请时间"]}`
	srv := fakeLLMServer(t, llmContent)
	defer srv.Close()

	if err := NewAISettingService().Save(SaveAISettingRequest{
		APIURL: srv.URL, APIKey: "sk-x", Model: "gpt-4o-mini",
	}); err != nil {
		t.Fatalf("存储 AI 配置失败: %v", err)
	}

	svc := NewAIExtractService()
	doc := `<?xml version="1.0"?><w:document ` + docxNS + `><w:body><w:p><w:r><w:t>主机申请</w:t></w:r></w:p></w:body></w:document>`
	result, err := svc.Extract("apply.docx", buildDocx(t, doc))
	if err != nil {
		t.Fatalf("Extract 失败: %v", err)
	}
	if result.DocType != DocTypeHostApply || result.DocTypeName != "主机申请" {
		t.Fatalf("doc_type 不符: %+v", result)
	}
	if result.Fields["name"] != "web-01" || result.Fields["private_ip"] != "10.0.0.5" {
		t.Fatalf("fields 不符: %+v", result.Fields)
	}
	if len(result.Warnings) != 1 {
		t.Fatalf("warnings 不符: %+v", result.Warnings)
	}
	if !isValidToken(result.FileToken) {
		t.Fatalf("file_token 非法: %q", result.FileToken)
	}
}

// TestExtractUnknownType unknown 类型加告警且不报错（T-068）
func TestExtractUnknownType(t *testing.T) {
	newTestDB(t)
	useAIConfig(t, "svc-test-secret", t.TempDir())

	srv := fakeLLMServer(t, `{"doc_type":"unknown","fields":{}}`)
	defer srv.Close()

	if err := NewAISettingService().Save(SaveAISettingRequest{
		APIURL: srv.URL, APIKey: "sk-x", Model: "m",
	}); err != nil {
		t.Fatalf("存储 AI 配置失败: %v", err)
	}

	doc := `<?xml version="1.0"?><w:document ` + docxNS + `><w:body><w:p><w:r><w:t>some</w:t></w:r></w:p></w:body></w:document>`
	result, err := NewAIExtractService().Extract("x.docx", buildDocx(t, doc))
	if err != nil {
		t.Fatalf("Extract 失败: %v", err)
	}
	if result.DocType != "unknown" || len(result.Warnings) == 0 {
		t.Fatalf("期望 unknown + 告警: %+v", result)
	}
}

// TestExtractNotConfigured 未配置 API 地址返回 ErrLLMNotConfigured（T-068）
func TestExtractNotConfigured(t *testing.T) {
	newTestDB(t)
	useAIConfig(t, "svc-test-secret", t.TempDir())

	doc := `<?xml version="1.0"?><w:document ` + docxNS + `><w:body><w:p><w:r><w:t>x</w:t></w:r></w:p></w:body></w:document>`
	if _, err := NewAIExtractService().Extract("x.docx", buildDocx(t, doc)); err != ErrLLMNotConfigured {
		t.Fatalf("期望 ErrLLMNotConfigured，实际 %v", err)
	}
}
