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

// ResourceHistory 查询某资产的变更记录（resource_type + resource_id），供资产详情时间线使用。
func (ctrl *AuditLogController) ResourceHistory(c *gin.Context) {
	var req services.ResourceHistoryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		Error(c, 40001, "参数校验失败: "+err.Error())
		return
	}

	items, err := ctrl.service.ListByResource(req)
	if err != nil {
		Error(c, 40001, "查询失败: "+err.Error())
		return
	}

	Success(c, items)
}
