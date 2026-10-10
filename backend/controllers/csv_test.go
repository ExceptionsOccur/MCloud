package controllers

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func newImportTestEngine() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/api/import", NewCSVController().Import)
	return r
}

func buildImportRequest(t *testing.T, filename string, content []byte) *http.Request {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fw, err := w.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("构造 multipart 文件字段失败: %v", err)
	}
	if _, err := fw.Write(content); err != nil {
		t.Fatalf("写入 multipart 内容失败: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("关闭 multipart writer 失败: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/import", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	return req
}

func decodeImportResp(t *testing.T, body []byte) (int, string) {
	t.Helper()
	var resp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("解析响应失败: %v (%s)", err, body)
	}
	return resp.Code, resp.Message
}

// TestImportSizeLimit 超过 16MB 的导入请求返回 40001 且信息可读（T-044）
func TestImportSizeLimit(t *testing.T) {
	r := newImportTestEngine()

	over := bytes.Repeat([]byte("a"), 17<<20) // 17MB > 16MB 上限
	w := httptest.NewRecorder()
	r.ServeHTTP(w, buildImportRequest(t, "hosts.csv", over))

	if w.Code != http.StatusBadRequest {
		t.Fatalf("HTTP status = %d, want 400", w.Code)
	}
	code, msg := decodeImportResp(t, w.Body.Bytes())
	if code != 40001 {
		t.Fatalf("business code = %d, want 40001", code)
	}
	if !strings.Contains(msg, "16MB") {
		t.Fatalf("超限提示不可读: %q", msg)
	}
}

// TestImportUnderLimitReachesBusinessValidation 未超限请求正常穿透大小闸门进入业务校验（T-044）
func TestImportUnderLimitReachesBusinessValidation(t *testing.T) {
	r := newImportTestEngine()

	// 非 .csv 后缀：能读满请求体并到达 service 后缀校验，说明大小限制未拦截正常请求
	w := httptest.NewRecorder()
	r.ServeHTTP(w, buildImportRequest(t, "hosts.txt", []byte("not-a-csv")))

	if w.Code != http.StatusBadRequest {
		t.Fatalf("HTTP status = %d, want 400", w.Code)
	}
	code, msg := decodeImportResp(t, w.Body.Bytes())
	if code != 40001 {
		t.Fatalf("business code = %d, want 40001", code)
	}
	if !strings.Contains(msg, ".csv") {
		t.Fatalf("应返回后缀校验错误, got %q", msg)
	}
}

// TestImportMissingField 缺少 file 字段仍返回 40001（T-044 回归原行为）
func TestImportMissingField(t *testing.T) {
	r := newImportTestEngine()

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/import",
		strings.NewReader("--x\r\nContent-Disposition: form-data\r\n\r\nok\r\n--x--\r\n"))
	req.Header.Set("Content-Type", "multipart/form-data; boundary=x")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("HTTP status = %d, want 400", w.Code)
	}
	code, _ := decodeImportResp(t, w.Body.Bytes())
	if code != 40001 {
		t.Fatalf("business code = %d, want 40001", code)
	}
}
