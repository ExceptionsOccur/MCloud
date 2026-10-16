package services

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/xuri/excelize/v2"
)

// AIDocParser 申请表单文档解析：把 docx/xlsx 转为纯文本供大模型提取（T-068）。
// docx 用标准库自解析（archive/zip + encoding/xml），xlsx 复用 excelize。

// ErrUnsupportedFileType 不支持的文档类型
var ErrUnsupportedFileType = errors.New("仅支持 .docx 或 .xlsx 文件")

// ErrDocParseFailed 文档解析失败
var ErrDocParseFailed = errors.New("文档解析失败")

// maxExtractedChars 提取文本上限，防止超长提示（超出按截断）
const maxExtractedChars = 24000

type AIDocParser struct{}

func NewAIDocParser() *AIDocParser {
	return &AIDocParser{}
}

// Parse 按扩展名分派解析，返回纯文本
func (p *AIDocParser) Parse(filename string, data []byte) (string, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".docx":
		return p.parseDocx(data)
	case ".xlsx":
		return p.parseXlsx(data)
	default:
		return "", ErrUnsupportedFileType
	}
}

// ---- docx 解析 ----

// docx XML 结构（仅取需要的节点）
type docxDocument struct {
	Body docxBody `xml:"body"`
}

type docxBody struct {
	Paragraphs []docxParagraph `xml:"p"`
	Tables     []docxTable     `xml:"tbl"`
}

type docxParagraph struct {
	Runs []docxRun `xml:"r"`
}

type docxRun struct {
	Texts []string `xml:"t"`
}

type docxTable struct {
	Rows []docxTableRow `xml:"tr"`
}

type docxTableRow struct {
	Cells []docxTableCell `xml:"tc"`
}

type docxTableCell struct {
	Paragraphs []docxParagraph `xml:"p"`
}

// paraText 拼接段落内所有 run 文本
func paraText(p docxParagraph) string {
	var sb strings.Builder
	for _, r := range p.Runs {
		for _, t := range r.Texts {
			sb.WriteString(t)
		}
	}
	return strings.TrimSpace(sb.String())
}

// parseDocx 从 word/document.xml 提取段落与表格文本（表格转为 markdown 行）
func (p *AIDocParser) parseDocx(data []byte) (string, error) {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", ErrDocParseFailed
	}

	var docFile *zip.File
	for _, f := range zr.File {
		if f.Name == "word/document.xml" {
			docFile = f
			break
		}
	}
	if docFile == nil {
		return "", ErrDocParseFailed
	}

	rc, err := docFile.Open()
	if err != nil {
		return "", ErrDocParseFailed
	}
	defer rc.Close()

	raw, err := io.ReadAll(rc)
	if err != nil {
		return "", ErrDocParseFailed
	}

	var doc docxDocument
	if err := xml.Unmarshal(raw, &doc); err != nil {
		return "", ErrDocParseFailed
	}

	var sb strings.Builder
	// 段落
	for _, para := range doc.Body.Paragraphs {
		if txt := paraText(para); txt != "" {
			sb.WriteString(txt)
			sb.WriteString("\n")
		}
	}
	// 表格：每行以 | 分隔单元格，单元格内多段落以空格连接
	for _, tbl := range doc.Body.Tables {
		for _, row := range tbl.Rows {
			cells := make([]string, 0, len(row.Cells))
			for _, cell := range row.Cells {
				var parts []string
				for _, para := range cell.Paragraphs {
					if t := paraText(para); t != "" {
						parts = append(parts, t)
					}
				}
				cells = append(cells, strings.Join(parts, " "))
			}
			sb.WriteString("| ")
			sb.WriteString(strings.Join(cells, " | "))
			sb.WriteString(" |\n")
		}
	}

	out := strings.TrimSpace(sb.String())
	if out == "" {
		return "", ErrDocParseFailed
	}
	return truncateText(out), nil
}

// ---- xlsx 解析 ----

// parseXlsx 逐 sheet 转文本（每个非空 sheet 加标题行）
func (p *AIDocParser) parseXlsx(data []byte) (string, error) {
	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		return "", ErrDocParseFailed
	}
	defer f.Close()

	var sb strings.Builder
	for _, sheet := range f.GetSheetList() {
		rows, err := f.GetRows(sheet)
		if err != nil {
			continue
		}
		var rowsText strings.Builder
		for _, row := range rows {
			hasContent := false
			for _, c := range row {
				if strings.TrimSpace(c) != "" {
					hasContent = true
					break
				}
			}
			if !hasContent {
				continue
			}
			rowsText.WriteString(strings.Join(row, " | "))
			rowsText.WriteString("\n")
		}
		if rowsText.Len() == 0 {
			continue
		}
		fmt.Fprintf(&sb, "### 工作表: %s\n", sheet)
		sb.WriteString(rowsText.String())
	}

	out := strings.TrimSpace(sb.String())
	if out == "" {
		return "", ErrDocParseFailed
	}
	return truncateText(out), nil
}

// truncateText 截断超长文本
func truncateText(s string) string {
	runes := []rune(s)
	if len(runes) <= maxExtractedChars {
		return s
	}
	return string(runes[:maxExtractedChars])
}
