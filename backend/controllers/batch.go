package controllers

import (
	"mcloud/services"

	"github.com/gin-gonic/gin"
)

type BatchController struct {
	service *services.HostService
}

func NewBatchController() *BatchController {
	return &BatchController{
		service: services.NewHostService(),
	}
}

func (ctrl *BatchController) BatchCreate(c *gin.Context) {
	var req services.BatchCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 40001, "参数校验失败: "+err.Error())
		return
	}

	resp, err := ctrl.service.BatchCreate(req)
	if err != nil {
		Error(c, 50001, "批量创建失败: "+err.Error())
		return
	}

	Success(c, resp)
}

func (ctrl *BatchController) BatchUpdate(c *gin.Context) {
	var req services.BatchUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 40001, "参数校验失败: "+err.Error())
		return
	}

	if err := ctrl.service.BatchUpdate(req); err != nil {
		Error(c, 50001, "批量更新失败: "+err.Error())
		return
	}

	Success(c, nil)
}
