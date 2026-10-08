package controllers

import (
	"errors"
	"strconv"

	"mcloud/services"

	"github.com/gin-gonic/gin"
)

type HostController struct {
	service *services.HostService
}

func NewHostController() *HostController {
	return &HostController{
		service: services.NewHostService(),
	}
}

func (ctrl *HostController) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		Error(c, 40001, "无效的ID")
		return
	}

	host, err := ctrl.service.GetByID(uint(id))
	if err != nil {
		Error(c, 40401, "主机不存在")
		return
	}

	Success(c, host)
}

func (ctrl *HostController) Create(c *gin.Context) {
	var req services.CreateHostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 40001, "参数校验失败: "+err.Error())
		return
	}

	id, err := ctrl.service.Create(req)
	if err != nil {
		if errors.Is(err, services.ErrPersonNotFound) {
			Error(c, 40001, err.Error())
			return
		}
		Error(c, 40901, err.Error())
		return
	}

	Success(c, gin.H{"id": id})
}

func (ctrl *HostController) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		Error(c, 40001, "无效的ID")
		return
	}

	var req services.UpdateHostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 40001, "参数校验失败: "+err.Error())
		return
	}

	if err := ctrl.service.Update(uint(id), req); err != nil {
		if errors.Is(err, services.ErrPersonNotFound) {
			Error(c, 40001, err.Error())
			return
		}
		Error(c, 40901, err.Error())
		return
	}

	Success(c, nil)
}

func (ctrl *HostController) Filter(c *gin.Context) {
	var req services.FilterHostRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		Error(c, 40001, "参数校验失败: "+err.Error())
		return
	}

	resp, err := ctrl.service.Filter(req)
	if err != nil {
		Error(c, 50001, "查询失败: "+err.Error())
		return
	}

	Success(c, resp)
}

func (ctrl *HostController) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		Error(c, 40001, "无效的ID")
		return
	}

	if err := ctrl.service.Delete(uint(id)); err != nil {
		if errors.Is(err, services.ErrHostReferenced) || errors.Is(err, services.ErrHostReferencedByMapping) {
			Error(c, 40901, err.Error())
			return
		}
		Error(c, 40401, "主机不存在")
		return
	}

	Success(c, nil)
}

func (ctrl *HostController) ListRegions(c *gin.Context) {
	regions, err := ctrl.service.ListRegions()
	if err != nil {
		Error(c, 50001, "查询失败: "+err.Error())
		return
	}
	Success(c, regions)
}
