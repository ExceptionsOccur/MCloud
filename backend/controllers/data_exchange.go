package controllers

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"mcloud/services"

	"github.com/gin-gonic/gin"
)

// importMaxBytes 导入文件大小上限 20MB
const importMaxBytes = 20 << 20

// xlsxMIME xlsx 下载响应 Content-Type
const xlsxMIME = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"

type DataExchangeController struct {
	service *services.DataExchangeService
}

func NewDataExchangeController() *DataExchangeController {
	return &DataExchangeController{service: services.NewDataExchangeService()}
}

// Export 导出 8 张业务表为单个 xlsx（GET /api/export/all）
func (ctrl *DataExchangeController) Export(c *gin.Context) {
	data, err := ctrl.service.Export()
	if err != nil {
		Error(c, 50001, "导出失败: "+err.Error())
		return
	}

	c.Header("Content-Disposition", `attachment; filename="`+services.ExportFilename()+`"`)
	c.Data(http.StatusOK, xlsxMIME, data)
}

// Import 导入 xlsx 并 upsert（POST /api/import/all，multipart 字段 file）
func (ctrl *DataExchangeController) Import(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, importMaxBytes)
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		if strings.Contains(err.Error(), "request body too large") {
			Error(c, 40001, "文件大小超过 20MB 限制")
		} else {
			Error(c, 40001, "请上传 .xlsx 文件（字段名 file）")
		}
		return
	}
	defer file.Close()

	if !strings.HasSuffix(strings.ToLower(header.Filename), ".xlsx") {
		Error(c, 40001, "仅支持 .xlsx 文件")
		return
	}

	data, err := io.ReadAll(file)
	if err != nil {
		Error(c, 50001, "读取文件失败")
		return
	}

	op, requestID := auditContext(c)
	report, err := ctrl.service.Import(data, op, requestID)
	if err != nil {
		// 结构性错误（非 xlsx/缺 sheet/缺列/表头重复）→ 40001；执行失败 → 50001
		if errors.Is(err, services.ErrImportStructure) {
			Error(c, 40001, err.Error())
		} else {
			Error(c, 50001, "导入失败: "+err.Error())
		}
		return
	}
	// committed=false 且 errors 非空：行级错误已整体回滚，报告承载逐行明细
	Success(c, report)
}
