package services

import (
	"encoding/csv"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"mcloud/database"
	"mcloud/models"
	"mcloud/utils"
)

var (
	// ErrCSVFileSuffix 上传文件后缀非 .csv（结构性错误，controller 映射 40001）
	ErrCSVFileSuffix = errors.New("仅支持.csv文件")
	// ErrCSVDecode 文件编码识别失败（结构性错误，controller 映射 40001）
	ErrCSVDecode = errors.New("文件编码识别失败")
	// ErrCSVParse CSV 解析失败（结构性错误，controller 映射 40001；具体原因经 wrap 携带）
	ErrCSVParse = errors.New("CSV解析失败")
)

// CSVImportResult CSV 导入结果计数（对外响应结构；字段序对齐历史 gin.H 字典序，保证响应字节不变）
type CSVImportResult struct {
	Errors  int `json:"errors"`
	Skipped int `json:"skipped"`
	Success int `json:"success"`
}

// parseCSVLine 按 CSV 规则解析单行（支持双引号包裹含逗号的字段）
func parseCSVLine(line string) ([]string, error) {
	reader := csv.NewReader(strings.NewReader(line))
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1
	return reader.Read()
}

// atoiOrZero 解析 CSV 数字列，非数字或超范围时返回 0（保持既有导入容错语义）
func atoiOrZero(s string) int {
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return 0
	}
	return n
}

func (s *HostService) ParseCSVRowToCreateHost(row []string) (CreateHostRequest, error) {
	if len(row) < 26 {
		return CreateHostRequest{}, fmt.Errorf("列数不足，至少26列，实际%d列", len(row))
	}

	cpu := atoiOrZero(row[7])
	memory := atoiOrZero(row[9])
	systemDisk := atoiOrZero(row[10])
	dataDisk := atoiOrZero(row[11])
	// 磁盘为 27 列新文件可选尾列（T-050）；旧 26 列文件无该列，保持 0
	disk := 0
	if len(row) >= 27 {
		disk = atoiOrZero(row[26])
	}

	isDB := strings.TrimSpace(row[13]) == "是" || strings.TrimSpace(row[13]) == "true" || strings.TrimSpace(row[13]) == "1"

	return CreateHostRequest{
		Region:            strings.TrimSpace(row[0]),
		InstanceID:        strings.TrimSpace(row[1]),
		Name:              strings.TrimSpace(row[2]),
		PrivateIP:         strings.TrimSpace(row[3]),
		AssetType:         strings.TrimSpace(row[5]),
		OS:                strings.TrimSpace(row[6]),
		CPU:               cpu,
		CPUArch:           strings.TrimSpace(row[8]),
		Memory:            memory,
		Disk:              disk,
		SystemDisk:        systemDisk,
		DataDisk:          dataDisk,
		EnvType:           strings.TrimSpace(row[12]),
		IsDBServer:        &isDB,
		Status:            strings.TrimSpace(row[14]),
		OpenPorts:         strings.TrimSpace(row[15]),
		Tags:              strings.TrimSpace(row[16]),
		ApplyUnit:         strings.TrimSpace(row[17]),
		Applicant:         strings.TrimSpace(row[18]),
		ApplicantContact:  strings.TrimSpace(row[19]),
		Project:           strings.TrimSpace(row[20]),
		ApplyReason:       strings.TrimSpace(row[21]),
		ApplyConfig:       strings.TrimSpace(row[22]),
		ApplyTime:         strings.TrimSpace(row[23]),
		ObjectStorageSize: strings.TrimSpace(row[24]),
		Remark:            strings.TrimSpace(row[25]),
	}, nil
}

// ImportCSV CSV 导入全流程：后缀校验 → 编码识别 → 解析 → 逐行创建
// 行为口径：跳过表头；列数不足或解析失败计入 errors；内网IP已存在计入 skipped
func (s *HostService) ImportCSV(filename string, data []byte, op Operator, requestID string) (*CSVImportResult, error) {
	if !strings.HasSuffix(filename, ".csv") {
		return nil, ErrCSVFileSuffix
	}

	content, err := utils.DetectAndDecode(data)
	if err != nil {
		return nil, ErrCSVDecode
	}

	records, err := utils.ParseCSV(content)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCSVParse, err)
	}

	result := &CSVImportResult{}
	for i, row := range records {
		if i == 0 {
			continue // skip header
		}
		if len(row) < 26 {
			result.Errors++
			continue
		}

		req, err := s.ParseCSVRowToCreateHost(row)
		if err != nil {
			result.Errors++
			continue
		}

		if _, err := s.Create(req, op, requestID); err != nil {
			if strings.Contains(err.Error(), "已存在") {
				result.Skipped++
			} else {
				result.Errors++
			}
		} else {
			result.Success++
		}
	}

	NewAuditService().Record(op, "import_csv", "host", nil,
		DetailDiff{After: map[string]interface{}{
			"filename": filename, "success": result.Success,
			"skipped": result.Skipped, "errors": result.Errors,
		}}, requestID)
	return result, nil
}

func (s *HostService) ExportToCSVRows(hosts []models.Host) [][]string {
	var rows [][]string
	for _, h := range hosts {
		isDB := "否"
		if h.IsDBServer {
			isDB = "是"
		}
		ipMapped := "否"
		if h.IPMapped {
			ipMapped = "是"
		}
		row := []string{
			h.Region, h.InstanceID, h.Name, h.PrivateIP, ipMapped,
			h.AssetType, h.OS, strconv.Itoa(h.CPU), h.CPUArch, strconv.Itoa(h.Memory),
			strconv.Itoa(h.SystemDisk), strconv.Itoa(h.DataDisk), h.EnvType, isDB,
			h.Status, h.OpenPorts, h.Tags,
		}
		if h.Application != nil {
			row = append(row,
				h.Application.ApplyUnit, h.Application.Applicant, h.Application.ApplicantContact,
				h.Application.Project, h.Application.ApplyReason, h.Application.ApplyConfig,
				h.Application.ApplyTime, h.Application.ObjectStorageSize, h.Application.Remark,
			)
		} else {
			row = append(row, "", "", "", "", "", "", "", "", "")
		}
		row = append(row, strconv.Itoa(h.Disk))
		rows = append(rows, row)
	}
	return rows
}

// ExportCSV 生成全量主机 CSV 文本（含 BOM 与表头）
func (s *HostService) ExportCSV() (string, error) {
	var hosts []models.Host
	if err := database.DB.Preload("Application").Find(&hosts).Error; err != nil {
		return "", err
	}
	return utils.BuildCSVOutput(utils.CSVHeaders, s.ExportToCSVRows(hosts))
}

// CSVTemplate 生成导入模板 CSV 文本（仅表头，含 BOM）
func (s *HostService) CSVTemplate() (string, error) {
	return utils.BuildCSVOutput(utils.CSVHeaders, nil)
}
