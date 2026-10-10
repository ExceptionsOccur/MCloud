package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"mcloud/config"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const testJWTSecret = "unit-test-secret"

func setupTestConfig(t *testing.T) {
	t.Helper()
	prev := config.AppConfig
	config.AppConfig = &config.Config{JWTSecret: testJWTSecret}
	t.Cleanup(func() { config.AppConfig = prev })
}

func signToken(t *testing.T, claims jwt.MapClaims, secret string) string {
	t.Helper()
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("签发测试 token 失败: %v", err)
	}
	return token
}

func TestValidateToken(t *testing.T) {
	setupTestConfig(t)
	now := time.Now()

	tests := []struct {
		name    string
		token   func(t *testing.T) string
		wantID  uint
		wantErr string
	}{
		{
			name: "有效token返回user_id",
			token: func(t *testing.T) string {
				return signToken(t, jwt.MapClaims{
					"user_id": float64(7),
					"exp":     now.Add(time.Hour).Unix(),
				}, testJWTSecret)
			},
			wantID: 7,
		},
		{
			name: "过期token拒绝",
			token: func(t *testing.T) string {
				return signToken(t, jwt.MapClaims{
					"user_id": float64(1),
					"exp":     now.Add(-time.Hour).Unix(),
				}, testJWTSecret)
			},
			wantErr: "无效的认证Token",
		},
		{
			name: "伪造签名拒绝",
			token: func(t *testing.T) string {
				return signToken(t, jwt.MapClaims{
					"user_id": float64(1),
					"exp":     now.Add(time.Hour).Unix(),
				}, "attacker-secret")
			},
			wantErr: "无效的认证Token",
		},
		{
			name:    "非法字符串拒绝",
			token:   func(t *testing.T) string { return "not-a-jwt" },
			wantErr: "无效的认证Token",
		},
		{
			name: "缺少user_id拒绝",
			token: func(t *testing.T) string {
				return signToken(t, jwt.MapClaims{
					"exp": now.Add(time.Hour).Unix(),
				}, testJWTSecret)
			},
			wantErr: "无效的Token信息",
		},
		{
			name: "非HMAC算法拒绝",
			token: func(t *testing.T) string {
				tok, err := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.MapClaims{
					"user_id": float64(1),
					"exp":     now.Add(time.Hour).Unix(),
				}).SignedString(jwt.UnsafeAllowNoneSignatureType)
				if err != nil {
					t.Fatalf("签发 none 算法测试 token 失败: %v", err)
				}
				return tok
			},
			wantErr: "无效的认证Token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userID, err := ValidateToken(tt.token(t))
			if tt.wantErr != "" {
				if err == nil || err.Error() != tt.wantErr {
					t.Fatalf("ValidateToken() error = %v, want %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("ValidateToken() unexpected error: %v", err)
			}
			if userID != tt.wantID {
				t.Fatalf("ValidateToken() userID = %d, want %d", userID, tt.wantID)
			}
		})
	}
}

// TestCORSMiddleware 白名单 Origin 回显 CORS 头（含 credentials/预检），非白名单不回显（T-043）
func TestCORSMiddleware(t *testing.T) {
	prev := config.AppConfig
	config.AppConfig = &config.Config{CORSAllowedOrigins: []string{"http://localhost:5173"}}
	t.Cleanup(func() { config.AppConfig = prev })

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(CORSMiddleware())
	r.GET("/ping", func(c *gin.Context) { c.Status(http.StatusOK) })

	const allowed = "http://localhost:5173"

	tests := []struct {
		name    string
		method  string
		origin  string
		wantACA bool
	}{
		{"白名单GET回显", http.MethodGet, allowed, true},
		{"非白名单GET不回显", http.MethodGet, "http://evil.example", false},
		{"无Origin不受影响", http.MethodGet, "", false},
		{"白名单预检回显", http.MethodOptions, allowed, true},
		{"非白名单预检不回显", http.MethodOptions, "http://evil.example", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(tt.method, "/ping", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			r.ServeHTTP(w, req)

			acao := w.Header().Get("Access-Control-Allow-Origin")
			if tt.wantACA {
				if acao != allowed {
					t.Fatalf("Allow-Origin = %q, want %q", acao, allowed)
				}
				if got := w.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
					t.Fatalf("Allow-Credentials = %q, want true", got)
				}
			} else if acao != "" {
				t.Fatalf("非白名单不应回显 Allow-Origin, got %q", acao)
			}

			if tt.method == http.MethodOptions && w.Code != http.StatusNoContent {
				t.Fatalf("预检状态码 = %d, want 204", w.Code)
			}
			if tt.method == http.MethodGet && w.Code != http.StatusOK {
				t.Fatalf("业务请求状态码 = %d, want 200", w.Code)
			}
		})
	}
}

// TestCORSMiddlewareEmptyDefault 未配置白名单（默认空）不下发任何 CORS 头（T-043）
func TestCORSMiddlewareEmptyDefault(t *testing.T) {
	prev := config.AppConfig
	config.AppConfig = &config.Config{}
	t.Cleanup(func() { config.AppConfig = prev })

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(CORSMiddleware())
	r.GET("/ping", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	r.ServeHTTP(w, req)

	if acao := w.Header().Get("Access-Control-Allow-Origin"); acao != "" {
		t.Fatalf("默认空白名单不应下发 Allow-Origin, got %q", acao)
	}
	if w.Code != http.StatusOK {
		t.Fatalf("同源无 Origin 语义的请求状态码 = %d, want 200", w.Code)
	}
}
