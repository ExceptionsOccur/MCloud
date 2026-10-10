package routes

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"mcloud/config"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
)

const testJWTSecret = "ws-route-test-secret"

func newTestEngine(t *testing.T) *gin.Engine {
	t.Helper()
	prev := config.AppConfig
	config.AppConfig = &config.Config{JWTSecret: testJWTSecret}
	t.Cleanup(func() { config.AppConfig = prev })

	gin.SetMode(gin.TestMode)
	r := gin.New()
	SetupRoutes(r)
	return r
}

func signTestToken(t *testing.T) string {
	t.Helper()
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": float64(1),
		"exp":     time.Now().Add(time.Hour).Unix(),
	}).SignedString([]byte(testJWTSecret))
	if err != nil {
		t.Fatalf("签发测试 token 失败: %v", err)
	}
	return token
}

func decodeBody(t *testing.T, body []byte) (int, string) {
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

// TestWSProbeRejectsBadToken 无效/缺失 token 无法建立连接（升级前 40101）
func TestWSProbeRejectsBadToken(t *testing.T) {
	r := newTestEngine(t)

	tests := []struct {
		name  string
		query string
	}{
		{"缺失token", ""},
		{"非空但伪造", "token=not-a-jwt"},
		{"错误签名", "token=eyJhbGciOiJIUzI1NiJ9.eyJ1c2VyX2lkIjoxfQ.invalidsig"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/api/ws/probe?"+tt.query, nil)
			r.ServeHTTP(w, req)

			if w.Code != http.StatusUnauthorized {
				t.Fatalf("HTTP status = %d, want 401", w.Code)
			}
			code, _ := decodeBody(t, w.Body.Bytes())
			if code != 40101 {
				t.Fatalf("business code = %d, want 40101", code)
			}
		})
	}
}

// TestWSProbeAcceptsValidToken 有效 token 正常升级为 WebSocket 连接
func TestWSProbeAcceptsValidToken(t *testing.T) {
	r := newTestEngine(t)
	srv := httptest.NewServer(r)
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") +
		"/api/ws/probe?token=" + url.QueryEscape(signTestToken(t))
	conn, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		if resp != nil {
			t.Fatalf("有效 token 建连失败: %v (HTTP %d)", err, resp.StatusCode)
		}
		t.Fatalf("有效 token 建连失败: %v", err)
	}
	conn.Close()
}
