package services

import (
	"strings"
	"testing"
	"time"

	"mcloud/database"
	"mcloud/models"
	"mcloud/utils"
)

func TestLogin(t *testing.T) {
	newTestDB(t)
	newTestConfig(t)
	s := NewAuthService()

	tests := []struct {
		name       string
		prepare    func(t *testing.T) (username, password string)
		wantErrSub string // "" 表示登录成功
	}{
		{
			name: "正确密码签发token",
			prepare: func(t *testing.T) (string, string) {
				seedUser(t, "alice", "Pass4Test")
				return "alice", "Pass4Test"
			},
		},
		{
			name: "用户不存在",
			prepare: func(t *testing.T) (string, string) {
				return "ghost", "any"
			},
			wantErrSub: "用户名或密码错误",
		},
		{
			name: "错误密码计入失败次数",
			prepare: func(t *testing.T) (string, string) {
				u := seedUser(t, "bob", "right-pass")
				if err := setFailedAttempts(u.ID, utils.MaxFailedAttempts-1); err != nil {
					t.Fatalf("预置失败次数失败: %v", err)
				}
				return "bob", "wrong-pass"
			},
			wantErrSub: "用户名或密码错误",
		},
		{
			name: "已锁定账户拒绝",
			prepare: func(t *testing.T) (string, string) {
				u := seedUser(t, "carol", "right-pass")
				if err := setLockedUntil(u.ID, time.Now().Add(time.Minute)); err != nil {
					t.Fatalf("预置锁定失败: %v", err)
				}
				return "carol", "right-pass"
			},
			wantErrSub: "账户已锁定",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			username, password := tt.prepare(t)
			resp, user, err := s.Login(username, password)
			if tt.wantErrSub != "" {
				if err == nil {
					t.Fatal("期望错误，实际成功")
				}
				if !strings.Contains(err.Error(), tt.wantErrSub) {
					t.Fatalf("错误信息应含 %q，实际 %q", tt.wantErrSub, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("Login 失败: %v", err)
			}
			if resp.AccessToken == "" {
				t.Fatal("token 为空")
			}
			if resp.ExpiresIn <= 0 {
				t.Fatalf("expires_in 非法: %d", resp.ExpiresIn)
			}
			if user.Username != username {
				t.Fatalf("返回用户不符: %s", user.Username)
			}
			if user.LastLogin == nil {
				t.Fatal("LastLogin 未更新")
			}
			if user.FailedAttempts != 0 {
				t.Fatalf("FailedAttempts 未清零: %d", user.FailedAttempts)
			}
		})
	}
}

func TestLoginLockout(t *testing.T) {
	newTestDB(t)
	newTestConfig(t)
	s := NewAuthService()
	seedUser(t, "dave", "correct-horse")

	for i := 0; i < utils.MaxFailedAttempts; i++ {
		if _, _, err := s.Login("dave", "wrong"); err == nil {
			t.Fatalf("第 %d 次错误密码应失败", i+1)
		}
	}
	var locked models.User
	if err := database.DB.Where("username = ?", "dave").First(&locked).Error; err != nil {
		t.Fatalf("回查用户失败: %v", err)
	}
	if locked.FailedAttempts < utils.MaxFailedAttempts {
		t.Fatalf("失败次数未累计: %d", locked.FailedAttempts)
	}
	if locked.LockedUntil == nil || !locked.LockedUntil.After(time.Now()) {
		t.Fatal("锁定时间未生效")
	}
	if _, _, err := s.Login("dave", "correct-horse"); err == nil || !strings.Contains(err.Error(), "账户已锁定") {
		t.Fatalf("锁定期内正确密码也应拒绝，实际 %v", err)
	}
}

func TestLoginResetsFailedCount(t *testing.T) {
	newTestDB(t)
	newTestConfig(t)
	s := NewAuthService()
	u := seedUser(t, "erin", "good-pass")
	if err := setFailedAttempts(u.ID, 2); err != nil {
		t.Fatalf("预置失败次数失败: %v", err)
	}
	_, user, err := s.Login("erin", "good-pass")
	if err != nil {
		t.Fatalf("Login 失败: %v", err)
	}
	if user.FailedAttempts != 0 || user.LockedUntil != nil {
		t.Fatalf("成功登录未清零: attempts=%d locked=%v", user.FailedAttempts, user.LockedUntil)
	}
}

// TestLoginUpgradesLegacyHash 旧 SHA-256 用户登录成功后平滑升级为 bcrypt（T-009 验收）
func TestLoginUpgradesLegacyHash(t *testing.T) {
	newTestDB(t)
	newTestConfig(t)
	s := NewAuthService()

	salt := utils.GenerateSalt()
	legacy := salt + "$" + utils.HashPassword("legacy-pass", salt)
	u := models.User{Username: "olduser", PasswordHash: legacy}
	if err := database.DB.Create(&u).Error; err != nil {
		t.Fatalf("种子旧格式用户失败: %v", err)
	}

	if _, _, err := s.Login("olduser", "legacy-pass"); err != nil {
		t.Fatalf("旧格式登录失败: %v", err)
	}
	var after models.User
	if err := database.DB.Where("username = ?", "olduser").First(&after).Error; err != nil {
		t.Fatalf("回查用户失败: %v", err)
	}
	if utils.IsLegacyHash(after.PasswordHash) {
		t.Fatalf("登录后未升级为 bcrypt: %q", after.PasswordHash)
	}
	if !utils.VerifyPassword("legacy-pass", after.PasswordHash) {
		t.Fatal("升级后密码校验失败")
	}
	// 升级后走 bcrypt 路径再次登录
	if _, _, err := s.Login("olduser", "legacy-pass"); err != nil {
		t.Fatalf("升级后登录失败: %v", err)
	}
}
