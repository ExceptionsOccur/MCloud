package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"mcloud/config"
)

// AIFileStore AI 上传文件的暂存与归档（T-068）。
// 目录结构：<AI_UPLOAD_DIR>/incoming/<token>__<原文件名> 与 <AI_UPLOAD_DIR>/archived/<YYYY-MM-DD>/<时间戳>_<原文件名>。
// token 为随机 hex，仅允许 [0-9a-f]，作为服务端文件定位键，避免暴露/篡改真实路径（防路径穿越）。

const (
	aiIncomingDir = "incoming"
	aiArchivedDir = "archived"
)

// ErrFileTokenInvalid token 非法（格式错误/不存在/已归档）
var ErrFileTokenInvalid = errors.New("文件令牌无效或已过期")

// ErrArchiveFailed 归档失败
var ErrArchiveFailed = errors.New("文件归档失败")

type AIFileStore struct{}

func NewAIFileStore() *AIFileStore {
	return &AIFileStore{}
}

// baseDir 返回配置的上传根目录（缺省 ./data/ai_uploads）
func (s *AIFileStore) baseDir() string {
	dir := config.AppConfig.AIUploadDir
	if strings.TrimSpace(dir) == "" {
		dir = "./data/ai_uploads"
	}
	return dir
}

// incomingDir / archivedDir 返回子目录绝对路径并确保存在
func (s *AIFileStore) incomingDir() (string, error) {
	dir := filepath.Join(s.baseDir(), aiIncomingDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

// sanitizeFilename 去除路径分隔与穿越片段，仅保留基础名（防目录穿越）
func sanitizeFilename(name string) string {
	name = filepath.Base(strings.TrimSpace(name))
	name = strings.ReplaceAll(name, "..", "_")
	if name == "" || name == "." || name == string(filepath.Separator) {
		name = "upload"
	}
	return name
}

// isValidToken token 仅允许 32 位小写 hex
func isValidToken(token string) bool {
	if len(token) != 32 {
		return false
	}
	for _, r := range token {
		if (r < '0' || r > '9') && (r < 'a' || r > 'f') {
			return false
		}
	}
	return true
}

// SaveIncoming 将上传文件写入 incoming 目录，返回 token（文件名内嵌 token 便于归档定位）
func (s *AIFileStore) SaveIncoming(data []byte, filename string) (string, error) {
	dir, err := s.incomingDir()
	if err != nil {
		return "", err
	}

	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token := hex.EncodeToString(b)

	safeName := sanitizeFilename(filename)
	target := filepath.Join(dir, token+"__"+safeName)
	if err := os.WriteFile(target, data, 0o600); err != nil {
		return "", err
	}
	return token, nil
}

// findIncoming 按 token 定位 incoming 目录中对应文件（token 严格校验，防穿越）
func (s *AIFileStore) findIncoming(token string) (string, error) {
	if !isValidToken(token) {
		return "", ErrFileTokenInvalid
	}
	dir, err := s.incomingDir()
	if err != nil {
		return "", err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", ErrFileTokenInvalid
	}
	prefix := token + "__"
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.HasPrefix(e.Name(), prefix) {
			return filepath.Join(dir, e.Name()), nil
		}
	}
	return "", ErrFileTokenInvalid
}

// Archive 将 token 对应文件从 incoming 移入 archived/<YYYY-MM-DD>/，返回归档后的相对路径。
// 移动后 token 失效（incoming 中不再存在）。
func (s *AIFileStore) Archive(token string) (string, error) {
	src, err := s.findIncoming(token)
	if err != nil {
		return "", err
	}

	day := time.Now().Format("2006-01-02")
	destDir := filepath.Join(s.baseDir(), aiArchivedDir, day)
	if mkErr := os.MkdirAll(destDir, 0o755); mkErr != nil {
		return "", ErrArchiveFailed
	}

	base := filepath.Base(src)
	// 去掉 token 前缀，保留时间戳前缀防重名
	cleanName := strings.TrimPrefix(base, token+"__")
	dest := filepath.Join(destDir, fmt.Sprintf("%s_%s", time.Now().Format("20060102150405"), cleanName))

	if renameErr := os.Rename(src, dest); renameErr != nil {
		// 跨设备时回退为复制+删除
		if copyErr := copyFile(src, dest); copyErr != nil {
			return "", ErrArchiveFailed
		}
		if removeErr := os.Remove(src); removeErr != nil {
			return "", ErrArchiveFailed
		}
	}

	rel, err := filepath.Rel(s.baseDir(), dest)
	if err != nil {
		rel = dest
	}
	return filepath.ToSlash(rel), nil
}

// copyFile 复制文件（归档跨设备回退用）
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o600)
}
