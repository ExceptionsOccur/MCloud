package services

import (
	"strings"
	"testing"
	"time"

	"mcloud/config"
	"mcloud/database"
	"mcloud/models"
	"mcloud/utils"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

// newTestDB 为每个测试创建独立 sqlite 内存库并注入 database.DB 全局变量；
// AutoMigrate 全量模型，口径与 database/postgres.go 的 Migrate() 一致。
func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	return openTestDB(t, t.Name())
}

// openTestDB 按指定库名打开 sqlite 内存库并注入 database.DB；
// 不同库名 = 不同内存库（回灌测试先开源库、再开空目标库）
func openTestDB(t *testing.T, name string) *gorm.DB {
	t.Helper()
	dsn := "file:" + strings.ReplaceAll(name, "/", "_") + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("打开 sqlite 内存库失败: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("获取底层连接失败: %v", err)
	}
	sqlDB.SetMaxIdleConns(4)
	if err := db.AutoMigrate(
		&models.User{},
		&models.Person{},
		&models.Host{},
		&models.HostApplication{},
		&models.CloudResource{},
		&models.IPSubnet{},
		&models.ZeroTrust{},
		&models.ZeroTrustTarget{},
		&models.PortMapping{},
		&models.PublicIP{},
		&models.AuditLog{},
		&models.AISetting{},
		&models.ObjectStorage{},
		&models.ObjectStorageHost{},
	); err != nil {
		t.Fatalf("AutoMigrate 失败: %v", err)
	}
	// sqlite 驱动仅按精确 decltype timestamp/datetime/date 解析时间；
	// 模型列 timestamptz 在测试库重建为 datetime，避免非空时间扫描失败
	for _, col := range []string{"locked_until", "last_login"} {
		if err := db.Exec("ALTER TABLE users DROP COLUMN " + col).Error; err != nil {
			t.Fatalf("重建 users.%s 失败: %v", col, err)
		}
		if err := db.Exec("ALTER TABLE users ADD COLUMN " + col + " datetime").Error; err != nil {
			t.Fatalf("重建 users.%s 失败: %v", col, err)
		}
	}
	prev := database.DB
	database.DB = db
	t.Cleanup(func() { database.DB = prev })
	registerILIKEDowngrade(t, db)
	return db
}

// registerILIKEDowngrade sqlite 无 ILIKE 运算符：查询前把 where 原生 SQL 中的 ILIKE
// 降级为 LIKE（sqlite LIKE 对 ASCII 大小写不敏感，搜索口径与生产 Postgres ILIKE 一致）
func registerILIKEDowngrade(t *testing.T, db *gorm.DB) {
	t.Helper()
	err := db.Callback().Query().Before("gorm:query").Register("test:ilike_downgrade", func(db *gorm.DB) {
		w, ok := db.Statement.Clauses["WHERE"]
		if !ok {
			return
		}
		wc, ok := w.Expression.(clause.Where)
		if !ok {
			return
		}
		changed := false
		for i, e := range wc.Exprs {
			v, ok := e.(clause.Expr)
			if !ok || !strings.Contains(v.SQL, "ILIKE") {
				continue
			}
			wc.Exprs[i] = clause.Expr{SQL: strings.ReplaceAll(v.SQL, "ILIKE", "LIKE"), Vars: v.Vars}
			changed = true
		}
		if changed {
			w.Expression = wc
			db.Statement.Clauses["WHERE"] = w
		}
	})
	if err != nil {
		t.Fatalf("注册 ILIKE 降级回调失败: %v", err)
	}
}

// newTestConfig 注入 JWT 测试配置（Login 依赖 config.AppConfig 非空）
func newTestConfig(t *testing.T) {
	t.Helper()
	prev := config.AppConfig
	config.AppConfig = &config.Config{JWTSecret: "test-secret", JWTExpiresHrs: 1}
	t.Cleanup(func() { config.AppConfig = prev })
}

var testOp = Operator{ID: 1, Name: "tester"}

func seedUser(t *testing.T, username, plain string) models.User {
	t.Helper()
	u := models.User{Username: username, PasswordHash: utils.FormatPasswordHash(plain)}
	if err := database.DB.Create(&u).Error; err != nil {
		t.Fatalf("种子用户失败: %v", err)
	}
	return u
}

func mustCreateHost(t *testing.T, s *HostService, ip string) uint {
	t.Helper()
	req := CreateHostRequest{
		Region:    "华东",
		Name:      "host-" + ip,
		PrivateIP: ip,
		AssetType: "虚拟机",
		EnvType:   "生产",
		Status:    "运行中",
		ApplyTime: "2026-01-02",
	}
	id, err := s.Create(req, testOp, NewRequestID())
	if err != nil {
		t.Fatalf("种子主机失败: %v", err)
	}
	return id
}

func setFailedAttempts(userID uint, n int) error {
	return database.DB.Model(&models.User{}).Where("id = ?", userID).Update("failed_attempts", n).Error
}

func setLockedUntil(userID uint, tm time.Time) error {
	return database.DB.Model(&models.User{}).Where("id = ?", userID).Update("locked_until", tm).Error
}
