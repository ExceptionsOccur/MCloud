package controllers

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"mcloud/services"

	"github.com/gin-gonic/gin"
)

// aiExtractMaxBytes AI 申请表单上传大小上限 20MB
const aiExtractMaxBytes = 20 << 20

type AIController struct {
	settingService *services.AISettingService
	extractService *services.AIExtractService
}

func NewAIController() *AIController {
	return &AIController{
		settingService: services.NewAISettingService(),
		extractService: services.NewAIExtractService(),
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

// Extract 上传申请表单并调用大模型提取（POST /api/ai/extract，multipart 字段 file）
func (ctrl *AIController) Extract(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, aiExtractMaxBytes)
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		if strings.Contains(err.Error(), "request body too large") {
			Error(c, 40001, "文件大小超过 20MB 限制")
		} else {
			Error(c, 40001, "请上传 .docx 或 .xlsx 文件（字段名 file）")
		}
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		if strings.Contains(err.Error(), "request body too large") {
			Error(c, 40001, "文件大小超过 20MB 限制")
		} else {
			Error(c, 50001, "读取文件失败")
		}
		return
	}

	result, err := ctrl.extractService.Extract(header.Filename, data)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrLLMNotConfigured):
			Error(c, 40001, err.Error())
		case errors.Is(err, services.ErrUnsupportedFileType),
			errors.Is(err, services.ErrDocParseFailed),
			errors.Is(err, services.ErrLLMRequestFailed):
			Error(c, 40001, err.Error())
		default:
			Error(c, 50001, "提取失败: "+err.Error())
		}
		return
	}
	Success(c, result)
}

// Archive 归档已成功入库的源文件（POST /api/ai/archive，body {file_token}）
func (ctrl *AIController) Archive(c *gin.Context) {
	var req struct {
		FileToken string `json:"file_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 40001, "参数校验失败: "+err.Error())
		return
	}

	path, err := ctrl.extractService.Archive(req.FileToken)
	if err != nil {
		if errors.Is(err, services.ErrFileTokenInvalid) {
			Error(c, 40001, err.Error())
			return
		}
		Error(c, 50001, err.Error())
		return
	}
	Success(c, gin.H{"archived_path": path})
}
