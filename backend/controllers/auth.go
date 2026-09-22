package controllers

import (
	"mcloud/services"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	service *services.AuthService
}

func NewAuthController() *AuthController {
	return &AuthController{
		service: services.NewAuthService(),
	}
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (ctrl *AuthController) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 40001, "参数校验失败: "+err.Error())
		return
	}

	resp, user, err := ctrl.service.Login(req.Username, req.Password)
	if err != nil {
		Error(c, 40101, err.Error())
		return
	}

	Success(c, gin.H{
		"access_token": resp.AccessToken,
		"expires_in":   resp.ExpiresIn,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
		},
	})
}

func (ctrl *AuthController) Logout(c *gin.Context) {
	Success(c, nil)
}

func (ctrl *AuthController) Me(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		Error(c, 40101, "未登录")
		return
	}

	user, err := ctrl.service.GetUserByID(userID.(uint))
	if err != nil {
		Error(c, 40401, "用户不存在")
		return
	}

	Success(c, gin.H{
		"id":       user.ID,
		"username": user.Username,
	})
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

func (ctrl *AuthController) ChangePassword(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		Error(c, 40101, "未登录")
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 40001, "参数校验失败: "+err.Error())
		return
	}

	if err := ctrl.service.ChangePassword(userID.(uint), req.OldPassword, req.NewPassword); err != nil {
		Error(c, 40001, err.Error())
		return
	}

	Success(c, nil)
}
