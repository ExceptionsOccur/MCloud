package database

import (
	"crypto/sha256"
	"fmt"
	"log"

	"mcloud/config"
	"mcloud/migrations"
	"mcloud/models"

	"github.com/pressly/goose/v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Connect() {
	// 前置依赖：调用方必须先执行 config.Load()（.env 加载在那里完成）
	cfg := config.AppConfig
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode,
	)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}

	log.Println("数据库连接成功")
}

func Migrate() {
	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatalf("获取数据库连接失败: %v", err)
	}

	if err = goose.SetDialect("postgres"); err != nil {
		log.Fatalf("goose 设置方言失败: %v", err)
	}
	goose.SetBaseFS(migrations.FS)
	if err = goose.Up(sqlDB, "."); err != nil {
		log.Fatalf("goose 迁移失败: %v", err)
	}

	// AutoMigrate 兜底：SQL 未覆盖的模型变更仍可由启动时补齐
	err = DB.AutoMigrate(
		&models.User{},
		&models.Person{},
		&models.Host{},
		&models.HostApplication{},
		&models.CloudResource{},
		&models.IPSubnet{},
		&models.ZeroTrust{},
		&models.PortMapping{},
		&models.PublicIP{},
		&models.AuditLog{},
	)
	if err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	seedAdmin()
	log.Println("数据库迁移完成")
}

func seedAdmin() {
	var count int64
	DB.Model(&models.User{}).Where("username = ?", "admin").Count(&count)
	if count == 0 {
		salt := "a1b2c3d4e5f6a7b8"
		password := "Pass4MCloud"
		hash := sha256.Sum256([]byte(salt + password))
		storedHash := fmt.Sprintf("%s$%x", salt, hash)

		admin := models.User{
			Username:     "admin",
			PasswordHash: storedHash,
		}
		DB.Create(&admin)
		log.Println("管理员账户已创建: admin / Pass4MCloud")
	}
}
