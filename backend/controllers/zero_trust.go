package controllers

import (
	"errors"
	"strconv"

	"mcloud/services"

	"github.com/gin-gonic/gin"
)

type ZeroTrustController struct {
	service *services.ZeroTrustService
}

func NewZeroTrustController() *ZeroTrustController {
	return &ZeroTrustController{
		service: services.NewZeroTrustService(),
	}
}

func (ctrl *ZeroTrustController) List(c *gin.Context) {
	data, err := ctrl.service.List(c.Query("keyword"))
	if err != nil {
		Error(c, 50001, "查询失败: "+err.Error())
		return
	}
	Success(c, data)
}

// BatchCreateText 零信任批量添加（文本粘贴）
func (ctrl *ZeroTrustController) BatchCreateText(c *gin.Context) {
	var req struct {
		Text string `json:"text" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 40001, "参数校验失败: "+err.Error())
		return
	}
	op, requestID := auditContext(c)
	resp, err := ctrl.service.BatchCreateText(req.Text, op, requestID)
	if err != nil {
		Error(c, 50001, "批量创建失败: "+err.Error())
		return
	}
	Success(c, resp)
}

func (ctrl *ZeroTrustController) Create(c *gin.Context) {
	var req services.ZeroTrustRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 40001, "参数校验失败: "+err.Error())
		return
	}
	op, requestID := auditContext(c)
	id, err := ctrl.service.Create(req, op, requestID)
	if err != nil {
		if errors.Is(err, services.ErrHostNotFound) {
			Error(c, 40001, err.Error())
			return
		}
		Error(c, 40001, err.Error())
		return
	}
	Success(c, gin.H{"id": id})
}

func (ctrl *ZeroTrustController) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 40001, "无效的ID")
		return
	}
	var req services.ZeroTrustRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 40001, "参数校验失败: "+err.Error())
		return
	}
	op, requestID := auditContext(c)
	if err := ctrl.service.Update(uint(id), req, op, requestID); err != nil {
		if errors.Is(err, services.ErrZeroTrustNotFound) || errors.Is(err, services.ErrHostNotFound) {
			Error(c, 40401, err.Error())
			return
		}
		Error(c, 40001, err.Error())
		return
	}
	Success(c, nil)
}

func (ctrl *ZeroTrustController) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 40001, "无效的ID")
		return
	}
	op, requestID := auditContext(c)
	if err := ctrl.service.Delete(uint(id), op, requestID); err != nil {
		if errors.Is(err, services.ErrZeroTrustNotFound) {
			Error(c, 40401, err.Error())
			return
		}
		Error(c, 40001, err.Error())
		return
	}
	Success(c, nil)
}
