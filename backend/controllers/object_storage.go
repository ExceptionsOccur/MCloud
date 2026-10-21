package controllers

import (
	"errors"
	"strconv"

	"mcloud/services"

	"github.com/gin-gonic/gin"
)

type ObjectStorageController struct {
	service *services.ObjectStorageService
}

func NewObjectStorageController() *ObjectStorageController {
	return &ObjectStorageController{
		service: services.NewObjectStorageService(),
	}
}

func (ctrl *ObjectStorageController) List(c *gin.Context) {
	data, err := ctrl.service.List(c.Query("keyword"))
	if err != nil {
		Error(c, 50001, "查询失败: "+err.Error())
		return
	}
	Success(c, data)
}

func (ctrl *ObjectStorageController) Create(c *gin.Context) {
	var req services.ObjectStorageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 40001, "参数校验失败: "+err.Error())
		return
	}
	op, requestID := auditContext(c)
	id, err := ctrl.service.Create(req, op, requestID)
	if err != nil {
		if errors.Is(err, services.ErrObjectStorageBucketExists) {
			Error(c, 40901, err.Error())
			return
		}
		Error(c, 40001, err.Error())
		return
	}
	Success(c, gin.H{"id": id})
}

func (ctrl *ObjectStorageController) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 40001, "无效的ID")
		return
	}
	var req services.ObjectStorageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 40001, "参数校验失败: "+err.Error())
		return
	}
	op, requestID := auditContext(c)
	if err := ctrl.service.Update(uint(id), req, op, requestID); err != nil {
		if errors.Is(err, services.ErrObjectStorageNotFound) {
			Error(c, 40401, err.Error())
			return
		}
		if errors.Is(err, services.ErrObjectStorageBucketExists) {
			Error(c, 40901, err.Error())
			return
		}
		Error(c, 40001, err.Error())
		return
	}
	Success(c, nil)
}

func (ctrl *ObjectStorageController) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, 40001, "无效的ID")
		return
	}
	op, requestID := auditContext(c)
	if err := ctrl.service.Delete(uint(id), op, requestID); err != nil {
		if errors.Is(err, services.ErrObjectStorageNotFound) {
			Error(c, 40401, err.Error())
			return
		}
		Error(c, 40001, err.Error())
		return
	}
	Success(c, nil)
}
