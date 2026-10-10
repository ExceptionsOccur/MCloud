package config

import (
	"errors"
	"io/fs"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost             string
	DBPort             string
	DBUser             string
	DBPassword         string
	DBName             string
	DBSSLMode          string
	JWTSecret          string
	JWTExpiresHrs      int
	ServerPort         string
	ServerMode         string
	CORSAllowedOrigins []string
}

var AppConfig *Config

func Load() {
	// .env 为可选文件：容器/CI 环境由 docker-compose 注入环境变量，缺失属正常情况
	if err := godotenv.Load(); err != nil && !errors.Is(err, fs.ErrNotExist) {
		log.Printf("加载 .env 失败: %v", err)
	}

	appConfig := &Config{
		DBHost:             getEnv("DB_HOST", "localhost"),
		DBPort:             getEnv("DB_PORT", "5432"),
		DBUser:             getEnv("DB_USER", "postgres"),
		DBPassword:         getEnv("DB_PASSWORD", ""),
		DBName:             getEnv("DB_NAME", "mcloud"),
		DBSSLMode:          getEnv("DB_SSLMODE", "disable"),
		JWTSecret:          getEnv("JWT_SECRET", "default-secret-change-me"),
		JWTExpiresHrs:      getEnvInt("JWT_EXPIRES_HOURS", 24),
		ServerPort:         getEnv("SERVER_PORT", "5677"),
		ServerMode:         getEnv("SERVER_MODE", "debug"),
		CORSAllowedOrigins: splitCSV(getEnv("CORS_ALLOWED_ORIGINS", "")),
	}

	AppConfig = appConfig
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return fallback
}

// splitCSV 拆分逗号分隔的配置值（如 CORS 白名单），去除空白项（T-043）
func splitCSV(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}
