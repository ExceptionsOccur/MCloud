package services

import (
	"errors"
	"fmt"
	"time"

	"mcloud/config"
	"mcloud/database"
	"mcloud/models"
	"mcloud/utils"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type AuthService struct{}

func NewAuthService() *AuthService {
	return &AuthService{}
}

type LoginResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

func (s *AuthService) Login(username, password string) (*LoginResponse, *models.User, error) {
	var user models.User
	result := database.DB.Where("username = ?", username).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil, errors.New("用户名或密码错误")
		}
		return nil, nil, result.Error
	}

	if user.LockedUntil != nil && user.LockedUntil.After(time.Now()) {
		return nil, nil, errors.New("账户已锁定，请稍后再试")
	}

	if !utils.VerifyPassword(password, user.PasswordHash) {
		user.FailedAttempts++
		if user.FailedAttempts >= utils.MaxFailedAttempts {
			lockUntil := time.Now().Add(time.Duration(utils.LockoutMinutes) * time.Minute)
			user.LockedUntil = &lockUntil
		}
		database.DB.Save(&user)
		return nil, nil, errors.New("用户名或密码错误")
	}

	user.FailedAttempts = 0
	user.LockedUntil = nil
	now := time.Now()
	user.LastLogin = &now
	database.DB.Save(&user)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(time.Duration(config.AppConfig.JWTExpiresHrs) * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString([]byte(config.AppConfig.JWTSecret))
	if err != nil {
		return nil, nil, err
	}

	return &LoginResponse{
		AccessToken: tokenString,
		ExpiresIn:   config.AppConfig.JWTExpiresHrs * 3600,
	}, &user, nil
}

func (s *AuthService) GetUserByID(id uint) (*models.User, error) {
	var user models.User
	result := database.DB.First(&user, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func (s *AuthService) ChangePassword(userID uint, oldPassword, newPassword string) error {
	var user models.User
	result := database.DB.First(&user, userID)
	if result.Error != nil {
		return result.Error
	}

	if !utils.VerifyPassword(oldPassword, user.PasswordHash) {
		return errors.New("原密码错误")
	}

	salt := utils.GenerateSalt()
	hash := utils.HashPassword(newPassword, salt)
	user.PasswordHash = fmt.Sprintf("%s$%s", salt, hash)
	database.DB.Save(&user)
	return nil
}
