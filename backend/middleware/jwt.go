package middleware

import (
	"errors"
	"net/http"
	"strings"

	"mcloud/config"
	"mcloud/controllers"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// ValidateToken 校验 JWT token 串并返回 user_id。
// HTTP `Authorization: Bearer` 与 WS query token 共用同一校验（T-022）。
func ValidateToken(tokenString string) (uint, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(config.AppConfig.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return 0, errors.New("无效的认证Token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("无效的Token信息")
	}
	userID, ok := claims["user_id"].(float64)
	if !ok {
		return 0, errors.New("无效的Token信息")
	}
	return uint(userID), nil
}

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			controllers.Error(c, 40101, "未提供认证信息")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			controllers.Error(c, 40101, "认证格式错误")
			c.Abort()
			return
		}

		userID, err := ValidateToken(parts[1])
		if err != nil {
			controllers.Error(c, 40101, err.Error())
			c.Abort()
			return
		}

		c.Set("user_id", userID)
		c.Next()
	}
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
