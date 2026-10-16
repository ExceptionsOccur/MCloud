package controllers

import (
	"mcloud/services"

	"github.com/gin-gonic/gin"
)

type AIController struct {
	settingService *services.AISettingService
}

func NewAIController() *AIController {
	return &AIController{
		settingService: services.NewAISettingService(),
	}
}

// GetSettings 返回 AI 接入配置（api_key 脱敏）
func (ctrl *AIController) GetSettings(c *gin.Context) {
	data, err := ctrl.settingService.Get()
	if err != nil {
		Error(c, 50001, "读取 AI 配置失败: "+err.Error())
		return
	}
	Success(c, data)
}

// SaveSettings 保存 AI 接入配置（api_key 为空则保留原值；变更写审计，Key 不入审计明文）
func (ctrl *AIController) SaveSettings(c *gin.Context) {
	var req services.SaveAISettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 40001, "参数校验失败: "+err.Error())
		return
	}
	if err := ctrl.settingService.Save(req); err != nil {
		Error(c, 40001, err.Error())
		return
	}

	// 审计：仅记录非敏感字段，api_key 以「是否更新」布尔代替明文
	op, requestID := auditContext(c)
	rowID := uint(1)
	services.NewAuditService().Record(op, "update", "ai_setting", &rowID, services.DetailDiff{
		After: gin.H{
			"api_url":     req.APIURL,
			"model":       req.Model,
			"api_key_set": req.APIKey != "",
		},
	}, requestID)

	Success(c, nil)
}
