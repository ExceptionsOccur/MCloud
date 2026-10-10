package controllers

import (
	"errors"
	"io"
	"net/http"

	"mcloud/services"

	"github.com/gin-gonic/gin"
)

type CSVController struct {
	hostService *services.HostService
}

func NewCSVController() *CSVController {
	return &CSVController{
		hostService: services.NewHostService(),
	}
}

func (ctrl *CSVController) Import(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		Error(c, 40001, "请上传CSV文件")
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		Error(c, 50001, "读取文件失败")
		return
	}

	result, err := ctrl.hostService.ImportCSV(header.Filename, data)
	if err != nil {
		// 结构性错误（后缀/编码/解析）→ 40001；其余执行失败 → 50001
		if errors.Is(err, services.ErrCSVFileSuffix) ||
			errors.Is(err, services.ErrCSVDecode) ||
			errors.Is(err, services.ErrCSVParse) {
			Error(c, 40001, err.Error())
			return
		}
		Error(c, 50001, "导入失败: "+err.Error())
		return
	}

	Success(c, result)
}

func (ctrl *CSVController) Export(c *gin.Context) {
	csvContent, err := ctrl.hostService.ExportCSV()
	if err != nil {
		Error(c, 50001, "生成CSV导出失败")
		return
	}

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=hosts_export.csv")
	c.String(http.StatusOK, csvContent)
}

func (ctrl *CSVController) Template(c *gin.Context) {
	csvContent, err := ctrl.hostService.CSVTemplate()
	if err != nil {
		Error(c, 50001, "生成CSV模板失败")
		return
	}

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=import_template.csv")
	c.String(http.StatusOK, csvContent)
}
