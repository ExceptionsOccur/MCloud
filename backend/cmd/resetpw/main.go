package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"os"

	"mcloud/config"
	"mcloud/database"
	"mcloud/models"
	"mcloud/utils"
)

const (
	upperSet   = "ABCDEFGHJKLMNPQRSTUVWXYZ"
	lowerSet   = "abcdefghijkmnpqrstuvwxyz"
	digitSet   = "23456789"
	specialSet = "!@#$%^&*"
	allSet     = upperSet + lowerSet + digitSet + specialSet
)

func randInt(max int) int {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		log.Fatalf("随机数生成失败: %v", err)
	}
	return int(n.Int64())
}

func randChar(set string) byte {
	return set[randInt(len(set))]
}

// generatePassword 生成指定长度的高熵随机密码，保证包含大小写字母、数字、特殊字符
func generatePassword(length int) string {
	chars := []byte{
		randChar(upperSet),
		randChar(lowerSet),
		randChar(digitSet),
		randChar(specialSet),
	}
	for len(chars) < length {
		chars = append(chars, randChar(allSet))
	}
	for i := len(chars) - 1; i > 0; i-- {
		j := randInt(i + 1)
		chars[i], chars[j] = chars[j], chars[i]
	}
	return string(chars)
}

func main() {
	config.Load()
	database.Connect()

	password := ""
	if len(os.Args) > 1 {
		password = os.Args[1]
	} else {
		password = generatePassword(10)
	}
	hash := utils.FormatPasswordHash(password)

	result := database.DB.Model(&models.User{}).
		Where("username = ?", "admin").
		Updates(map[string]interface{}{
			"password_hash":   hash,
			"failed_attempts": 0,
			"locked_until":    nil,
		})

	if result.Error != nil {
		log.Fatalf("更新失败: %v", result.Error)
	}
	if result.RowsAffected == 0 {
		log.Fatal("未找到 admin 用户")
	}

	fmt.Println("========================================")
	fmt.Println("  admin 密码已重置")
	fmt.Println("========================================")
	fmt.Printf("  用户名: admin\n")
	fmt.Printf("  新密码: %s\n", password)
	fmt.Println("========================================")
}
