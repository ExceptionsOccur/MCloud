package controllers

import (
	"mcloud/services"

	"github.com/gin-gonic/gin"
)

type AuditLogController struct {
	service *services.AuditService
}

func NewAuditLogController() *AuditLogController {
	return &AuditLogController{
		service: services.NewAuditService(),
	}
}

func (ctrl *AuditLogController) List(c *gin.Context) {
	var req services.ListAuditLogRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		Error(c, 40001, "参数校验失败: "+err.Error())
		return
	}

	resp, err := ctrl.service.List(req)
	if err != nil {
		Error(c, 50001, "查询失败: "+err.Error())
		return
	}

	Success(c, resp)
}
