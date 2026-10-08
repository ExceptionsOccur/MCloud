package controllers

import (
	"errors"
	"strconv"

	"mcloud/services"

	"github.com/gin-gonic/gin"
)

type PublicIPController struct {
	service *services.PublicIPService
}

func NewPublicIPController() *PublicIPController {
	return &PublicIPController{
		service: services.NewPublicIPService(),
	}
}

func (ctrl *PublicIPController) List(c *gin.Context) {
	data, err := ctrl.service.List(c.Query("keyword"))
	if err != nil {
		Error(c, 50001, "查询失败: "+err.Error())
		return
	}
	Success(c, data)
}

func (ctrl *PublicIPController) Create(c *gin.Context) {
	var req services.PublicIPRequest
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

func (ctrl *PublicIPController) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 40001, "无效的ID")
		return
	}
	var req services.PublicIPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 40001, "参数校验失败: "+err.Error())
		return
	}
	if err := ctrl.service.Update(uint(id), req); err != nil {
		if errors.Is(err, services.ErrPublicIPNotFound) {
			Error(c, 40401, err.Error())
			return
		}
		Error(c, 40001, err.Error())
		return
	}
	Success(c, nil)
}

func (ctrl *PublicIPController) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 40001, "无效的ID")
		return
	}
	if err := ctrl.service.Delete(uint(id)); err != nil {
		if errors.Is(err, services.ErrPublicIPNotFound) {
			Error(c, 40401, err.Error())
			return
		}
		Error(c, 40001, err.Error())
		return
	}
	Success(c, nil)
}
