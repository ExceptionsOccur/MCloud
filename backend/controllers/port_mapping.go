package controllers

import (
	"errors"
	"strconv"

	"mcloud/services"

	"github.com/gin-gonic/gin"
)

type PortMappingController struct {
	service *services.PortMappingService
}

func NewPortMappingController() *PortMappingController {
	return &PortMappingController{
		service: services.NewPortMappingService(),
	}
}

func (ctrl *PortMappingController) List(c *gin.Context) {
	data, err := ctrl.service.List(c.Query("keyword"))
	if err != nil {
		Error(c, 50001, "查询失败: "+err.Error())
		return
	}
	Success(c, data)
}

func (ctrl *PortMappingController) Create(c *gin.Context) {
	var req services.PortMappingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 40001, "参数校验失败: "+err.Error())
		return
	}
	op, requestID := auditContext(c)
	id, err := ctrl.service.Create(req, op, requestID)
	if err != nil {
		if errors.Is(err, services.ErrMappingHostNotFound) {
			Error(c, 40001, err.Error())
			return
		}
		if errors.Is(err, services.ErrPortMappingExists) {
			Error(c, 40901, err.Error())
			return
		}
		Error(c, 40001, err.Error())
		return
	}
	Success(c, gin.H{"id": id})
}

func (ctrl *PortMappingController) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 40001, "无效的ID")
		return
	}
	var req services.PortMappingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 40001, "参数校验失败: "+err.Error())
		return
	}
	op, requestID := auditContext(c)
	if err := ctrl.service.Update(uint(id), req, op, requestID); err != nil {
		if errors.Is(err, services.ErrPortMappingNotFound) {
			Error(c, 40401, err.Error())
			return
		}
		if errors.Is(err, services.ErrPortMappingExists) {
			Error(c, 40901, err.Error())
			return
		}
		Error(c, 40001, err.Error())
		return
	}
	Success(c, nil)
}

func (ctrl *PortMappingController) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 40001, "无效的ID")
		return
	}
	op, requestID := auditContext(c)
	if err := ctrl.service.Delete(uint(id), op, requestID); err != nil {
		if errors.Is(err, services.ErrPortMappingNotFound) {
			Error(c, 40401, err.Error())
			return
		}
		Error(c, 40001, err.Error())
		return
	}
	Success(c, nil)
}

func (ctrl *PortMappingController) BatchCreateText(c *gin.Context) {
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
