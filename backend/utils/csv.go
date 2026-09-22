package utils

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"strings"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

var CSVHeaders = []string{
	"区域", "实例ID", "主机名称", "内网IP", "公网IP",
	"资产类型", "操作系统", "CPU核数", "CPU架构", "内存(GB)",
	"系统盘(GB)", "数据盘(GB)", "环境类型", "是否数据库服务器",
	"状态", "开放端口", "标签", "申请单位", "申请人", "申请人联系方式",
	"所属项目", "申请理由", "申请配置", "申请时间", "对象存储大小", "备注",
}

func DetectAndDecode(data []byte) (string, error) {
	if utf8Valid(data) {
		return string(data), nil
	}

	// Try GBK
	if decoded, err := decodeWith(data, simplifiedchinese.GBK.NewDecoder()); err == nil {
		return decoded, nil
	}

	// Try GB18030
	if decoded, err := decodeWith(data, simplifiedchinese.GB18030.NewDecoder()); err == nil {
		return decoded, nil
	}

	return string(data), fmt.Errorf("unable to detect encoding")
}

func utf8Valid(data []byte) bool {
	for i, b := range data {
		if b <= 0x7F {
			continue
		}
		if b < 0xC2 || b > 0xF4 {
			return false
		}
		_ = i
	}
	return true
}

func decodeWith(data []byte, transformer transform.Transformer) (string, error) {
	decoded, err := io.ReadAll(transform.NewReader(bytes.NewReader(data), transformer))
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

func ParseCSV(content string) ([][]string, error) {
	reader := csv.NewReader(strings.NewReader(content))
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1

	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(records) < 2 {
		return nil, fmt.Errorf("CSV文件为空或只有表头")
	}
	return records, nil
}

func BuildCSVOutput(headers []string, rows [][]string) string {
	var buf bytes.Buffer
	// Write BOM for Excel UTF-8 compatibility
	buf.WriteString("\ufeff")

	writer := csv.NewWriter(&buf)
	writer.Write(headers)
	for _, row := range rows {
		writer.Write(row)
	}
	writer.Flush()
	return buf.String()
}
