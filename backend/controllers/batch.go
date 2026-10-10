package controllers

import (
	"errors"

	"mcloud/services"
	"mcloud/utils"

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

	op, requestID := auditContext(c)
	resp, err := ctrl.service.BatchCreate(req, op, requestID)
	if err != nil {
		Error(c, 50001, "批量创建失败: "+err.Error())
		return
	}

	Success(c, resp)
}

// BatchCreateText 纯文本批量添加：每行一条记录，逗号分隔
func (ctrl *BatchController) BatchCreateText(c *gin.Context) {
	var req services.BatchCreateTextRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 40001, "参数校验失败: "+err.Error())
		return
	}

	op, requestID := auditContext(c)
	resp, err := ctrl.service.BatchCreateFromText(req.Text, op, requestID)
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

	op, requestID := auditContext(c)
	if err := ctrl.service.BatchUpdate(req, op, requestID); err != nil {
		if errors.Is(err, services.ErrPersonNotFound) || errors.Is(err, utils.ErrInvalidDate) {
			Error(c, 40001, err.Error())
			return
		}
		Error(c, 50001, "批量更新失败: "+err.Error())
		return
	}

	Success(c, nil)
}
