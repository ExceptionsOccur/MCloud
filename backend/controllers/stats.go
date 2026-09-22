package controllers

import (
	"mcloud/services"

	"github.com/gin-gonic/gin"
)

type StatsController struct {
	service *services.StatsService
}

func NewStatsController() *StatsController {
	return &StatsController{
		service: services.NewStatsService(),
	}
}

func (ctrl *StatsController) IPUsage(c *gin.Context) {
	data, err := ctrl.service.GetIPUsage()
	if err != nil {
		Error(c, 50001, "查询失败: "+err.Error())
		return
	}
	Success(c, data)
}

func (ctrl *StatsController) Probe(c *gin.Context) {
	var req services.ProbeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 40001, "参数校验失败: "+err.Error())
		return
	}

	data, err := ctrl.service.Probe(req)
	if err != nil {
		Error(c, 50001, "探测失败: "+err.Error())
		return
	}
	Success(c, data)
}

func (ctrl *StatsController) BusinessStats(c *gin.Context) {
	data, err := ctrl.service.GetBusinessStats()
	if err != nil {
		Error(c, 50001, "查询失败: "+err.Error())
		return
	}
	Success(c, data)
}
