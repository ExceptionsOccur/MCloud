package services

import (
	"os"
	"path/filepath"
	"testing"

	"mcloud/config"
)

// useAIUploadDir 注入测试用上传目录（T-068）
func useAIUploadDir(t *testing.T, dir string) {
	t.Helper()
	prev := config.AppConfig
	config.AppConfig = &config.Config{AIUploadDir: dir}
	t.Cleanup(func() { config.AppConfig = prev })
}

// TestFileStoreSaveAndArchive 暂存后归档：incoming 移入 archived，token 失效（T-068）
func TestFileStoreSaveAndArchive(t *testing.T) {
	dir := t.TempDir()
	useAIUploadDir(t, dir)
	store := NewAIFileStore()

	token, err := store.SaveIncoming([]byte("hello"), "申请单.docx")
	if err != nil {
		t.Fatalf("SaveIncoming 失败: %v", err)
	}
	if !isValidToken(token) {
		t.Fatalf("token 格式非法: %q", token)
	}

	// incoming 存在
	incoming, err := store.findIncoming(token)
	if err != nil {
		t.Fatalf("findIncoming 失败: %v", err)
	}
	if _, statErr := os.Stat(incoming); statErr != nil {
		t.Fatalf("incoming 文件不存在: %v", statErr)
	}

	rel, err := store.Archive(token)
	if err != nil {
		t.Fatalf("Archive 失败: %v", err)
	}
	if filepath.Ext(rel) != ".docx" {
		t.Fatalf("归档路径应保留扩展名: %q", rel)
	}
	archived := filepath.Join(dir, filepath.FromSlash(rel))
	if _, err := os.Stat(archived); err != nil {
		t.Fatalf("归档文件不存在: %v", err)
	}
	// token 归档后失效
	if _, err := store.findIncoming(token); err != ErrFileTokenInvalid {
		t.Fatalf("归档后 findIncoming 应返回 ErrFileTokenInvalid，实际 %v", err)
	}
}

// TestFileStorePathTraversal 路径穿越文件名被清洗（T-068）
func TestFileStorePathTraversal(t *testing.T) {
	dir := t.TempDir()
	useAIUploadDir(t, dir)
	store := NewAIFileStore()

	token, err := store.SaveIncoming([]byte("x"), "../../etc/passwd")
	if err != nil {
		t.Fatalf("SaveIncoming 失败: %v", err)
	}
	incoming, err := store.findIncoming(token)
	if err != nil {
		t.Fatalf("findIncoming 失败: %v", err)
	}
	// 必须落在 incoming 目录内
	absIncoming, _ := filepath.Abs(incoming)
	absBase, _ := filepath.Abs(filepath.Join(dir, aiIncomingDir))
	if !pathHasPrefix(absIncoming, absBase) {
		t.Fatalf("文件逃逸出 incoming 目录: %s", absIncoming)
	}
	if filepath.Base(incoming) != token+"__passwd" {
		t.Fatalf("文件名未按预期清洗: %s", filepath.Base(incoming))
	}
}

// TestFileStoreInvalidToken 非法 token（穿越/格式错）被拒（T-068）
func TestFileStoreInvalidToken(t *testing.T) {
	dir := t.TempDir()
	useAIUploadDir(t, dir)
	store := NewAIFileStore()

	for _, tok := range []string{"", "../evil", "ZZZZ", "short"} {
		if _, err := store.Archive(tok); err != ErrFileTokenInvalid {
			t.Fatalf("token %q 期望 ErrFileTokenInvalid，实际 %v", tok, err)
		}
	}
}

// pathHasPrefix rel 安全前缀判断
func pathHasPrefix(p, prefix string) bool {
	rel, err := filepath.Rel(prefix, p)
	if err != nil {
		return false
	}
	return rel != ".." && !hasDotDotPrefix(rel)
}

func hasDotDotPrefix(rel string) bool {
	return len(rel) >= 3 && rel[0] == '.' && rel[1] == '.' && (rel[2] == '/' || rel[2] == filepath.Separator)
}
