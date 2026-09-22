package controllers

import (
	"mcloud/services"

	"github.com/gin-gonic/gin"
)

type CloudResourceController struct {
	service *services.CloudResourceService
}

func NewCloudResourceController() *CloudResourceController {
	return &CloudResourceController{
		service: services.NewCloudResourceService(),
	}
}

func (ctrl *CloudResourceController) List(c *gin.Context) {
	resources, err := ctrl.service.List()
	if err != nil {
		Error(c, 50001, "查询失败: "+err.Error())
		return
	}
	Success(c, resources)
}

func (ctrl *CloudResourceController) Update(c *gin.Context) {
	var req services.UpdateCloudResourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 40001, "参数校验失败: "+err.Error())
		return
	}

	if err := ctrl.service.Update(req); err != nil {
		Error(c, 50001, "更新失败: "+err.Error())
		return
	}

	Success(c, nil)
}
