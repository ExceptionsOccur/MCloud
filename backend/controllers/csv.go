package controllers

import (
	"io"
	"net/http"
	"strings"

	"mcloud/database"
	"mcloud/models"
	"mcloud/services"
	"mcloud/utils"

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

	if !strings.HasSuffix(header.Filename, ".csv") {
		Error(c, 40001, "仅支持.csv文件")
		return
	}

	data, err := io.ReadAll(file)
	if err != nil {
		Error(c, 50001, "读取文件失败")
		return
	}

	content, err := utils.DetectAndDecode(data)
	if err != nil {
		Error(c, 40001, "文件编码识别失败")
		return
	}

	records, err := utils.ParseCSV(content)
	if err != nil {
		Error(c, 40001, "CSV解析失败: "+err.Error())
		return
	}

	success := 0
	skipped := 0
	errors := 0

	for i, row := range records {
		if i == 0 {
			continue // skip header
		}
		if len(row) < 26 {
			errors++
			continue
		}

		req, err := ctrl.hostService.ParseCSVRowToCreateHost(row)
		if err != nil {
			errors++
			continue
		}

		_, err = ctrl.hostService.Create(req)
		if err != nil {
			if strings.Contains(err.Error(), "已存在") {
				skipped++
			} else {
				errors++
			}
		} else {
			success++
		}
	}

	Success(c, gin.H{
		"success": success,
		"skipped": skipped,
		"errors":  errors,
	})
}

func (ctrl *CSVController) Export(c *gin.Context) {
	var hosts []models.Host
	database.DB.Preload("Application").Find(&hosts)

	rows := ctrl.hostService.ExportToCSVRows(hosts)
	csvContent, err := utils.BuildCSVOutput(utils.CSVHeaders, rows)
	if err != nil {
		Error(c, 50001, "生成CSV导出失败")
		return
	}

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=hosts_export.csv")
	c.String(http.StatusOK, csvContent)
}

func (ctrl *CSVController) Template(c *gin.Context) {
	csvContent, err := utils.BuildCSVOutput(utils.CSVHeaders, nil)
	if err != nil {
		Error(c, 50001, "生成CSV模板失败")
		return
	}
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=import_template.csv")
	c.String(http.StatusOK, csvContent)
}
