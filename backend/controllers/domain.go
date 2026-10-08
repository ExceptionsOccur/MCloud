package controllers

import (
	"errors"
	"strconv"

	"mcloud/services"

	"github.com/gin-gonic/gin"
)

type DomainController struct {
	service *services.DomainService
}

func NewDomainController() *DomainController {
	return &DomainController{
		service: services.NewDomainService(),
	}
}

func (ctrl *DomainController) List(c *gin.Context) {
	data, err := ctrl.service.List(c.Query("keyword"))
	if err != nil {
		Error(c, 50001, "查询失败: "+err.Error())
		return
	}
	Success(c, data)
}

// BatchCreateText 域名批量添加（文本粘贴）
func (ctrl *DomainController) BatchCreateText(c *gin.Context) {
	var req struct {
		Text string `json:"text" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 40001, "参数校验失败: "+err.Error())
		return
	}
	resp, err := ctrl.service.BatchCreateText(req.Text)
	if err != nil {
		Error(c, 50001, "批量创建失败: "+err.Error())
		return
	}
	Success(c, resp)
}

func (ctrl *DomainController) Create(c *gin.Context) {
	var req services.DomainRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 40001, "参数校验失败: "+err.Error())
		return
	}
	id, err := ctrl.service.Create(req)
	if err != nil {
		Error(c, 40001, err.Error())
		return
	}
	Success(c, gin.H{"id": id})
}

func (ctrl *DomainController) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 40001, "无效的ID")
		return
	}
	var req services.DomainRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 40001, "参数校验失败: "+err.Error())
		return
	}
	if err := ctrl.service.Update(uint(id), req); err != nil {
		if errors.Is(err, services.ErrDomainNotFound) {
			Error(c, 40401, err.Error())
			return
		}
		Error(c, 40001, err.Error())
		return
	}
	Success(c, nil)
}

func (ctrl *DomainController) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 40001, "无效的ID")
		return
	}
	if err := ctrl.service.Delete(uint(id)); err != nil {
		if errors.Is(err, services.ErrDomainNotFound) {
			Error(c, 40401, err.Error())
			return
		}
		Error(c, 40001, err.Error())
		return
	}
	Success(c, nil)
}
