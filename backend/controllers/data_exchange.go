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

// backupMIME 加密备份包下载响应 Content-Type
const backupMIME = "application/octet-stream"

type DataExchangeController struct {
	service *services.DataExchangeService
}

func NewDataExchangeController() *DataExchangeController {
	return &DataExchangeController{service: services.NewDataExchangeService()}
}

// Export 导出 8 张业务表并经密码加密（GET /api/export/all?password=xxx，T-072）
//
// 导出必须设置密码：xlsx 字节 → AES-256-GCM 加密包（.mcbak），导入时凭同一密码解开。
func (ctrl *DataExchangeController) Export(c *gin.Context) {
	password := c.Query("password")
	if password == "" {
		Error(c, 40001, "导出必须设置密码")
		return
	}

	data, err := ctrl.service.Export()
	if err != nil {
		Error(c, 50001, "导出失败: "+err.Error())
		return
	}

	encrypted, err := services.EncryptBackup(data, password)
	if err != nil {
		if errors.Is(err, services.ErrBackupPasswordEmpty) {
			Error(c, 40001, err.Error())
			return
		}
		Error(c, 50001, "加密失败: "+err.Error())
		return
	}

	filename := services.ExportFilename() + services.EncryptBackupExt()
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.Data(http.StatusOK, backupMIME, encrypted)
}

// Import 导入备份并 upsert（POST /api/import/all，multipart 字段 file，可选字段 password）
func (ctrl *DataExchangeController) Import(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, importMaxBytes)
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		if strings.Contains(err.Error(), "request body too large") {
			Error(c, 40001, "文件大小超过 20MB 限制")
		} else {
			Error(c, 40001, "请上传备份文件（字段名 file）")
		}
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		Error(c, 50001, "读取文件失败")
		return
	}

	// 加密备份凭密码解开；明文 xlsx 可直接导入、无需密码（T-072）
	isXlsx := strings.HasSuffix(strings.ToLower(header.Filename), ".xlsx")
	if !services.IsEncryptedBackup(data) && !isXlsx {
		Error(c, 40001, "仅支持 .mcbak 加密备份或 .xlsx 文件")
		return
	}
	op, requestID := auditContext(c)
	report, err := ctrl.service.ImportWithPassword(data, c.PostForm("password"), op, requestID)
	if err != nil {
		// 密码问题/结构性错误（非 xlsx/缺 sheet/缺列/表头重复）→ 40001；执行失败 → 50001
		if errors.Is(err, services.ErrImportStructure) ||
			errors.Is(err, services.ErrBackupPasswordRequired) ||
			errors.Is(err, services.ErrBackupPasswordWrong) {
			Error(c, 40001, err.Error())
		} else {
			Error(c, 50001, "导入失败: "+err.Error())
		}
		return
	}
	// committed=false 且 errors 非空：行级错误已整体回滚，报告承载逐行明细
	Success(c, report)
}
