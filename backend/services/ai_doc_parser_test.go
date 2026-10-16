package services

import (
	"archive/zip"
	"bytes"
	"testing"

	"github.com/xuri/excelize/v2"
)

// buildDocx 构造含段落与表格的最小 docx（word/document.xml）
func buildDocx(t *testing.T, documentXML string) []byte {
	t.Helper()
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)
	f, err := zw.Create("word/document.xml")
	if err != nil {
		t.Fatalf("创建 zip 条目失败: %v", err)
	}
	if _, err := f.Write([]byte(documentXML)); err != nil {
		t.Fatalf("写入 document.xml 失败: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("关闭 zip 失败: %v", err)
	}
	return buf.Bytes()
}

const docxNS = `xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"`

// TestParseDocxParagraphsAndTables docx 段落与表格文本均可提取（T-068）
func TestParseDocxParagraphsAndTables(t *testing.T) {
	doc := `<?xml version="1.0"?><w:document ` + docxNS + `><w:body>` +
		`<w:p><w:r><w:t>主机申请单</w:t></w:r></w:p>` +
		`<w:tbl>` +
		`<w:tr><w:tc><w:p><w:r><w:t>主机名</w:t></w:r></w:p></w:tc><w:tc><w:p><w:r><w:t>web-01</w:t></w:r></w:p></w:tc></w:tr>` +
		`<w:tr><w:tc><w:p><w:r><w:t>内网IP</w:t></w:r></w:p></w:tc><w:tc><w:p><w:r><w:t>10.0.0.5</w:t></w:r></w:p></w:tc></w:tr>` +
		`</w:tbl>` +
		`</w:body></w:document>`

	text, err := NewAIDocParser().Parse("apply.docx", buildDocx(t, doc))
	if err != nil {
		t.Fatalf("解析 docx 失败: %v", err)
	}
	for _, want := range []string{"主机申请单", "web-01", "10.0.0.5", "| 主机名 | web-01 |"} {
		if !bytes.Contains([]byte(text), []byte(want)) {
			t.Fatalf("docx 文本缺少 %q，实际:\n%s", want, text)
		}
	}
}

// TestParseDocxInvalid 非法 docx 返回解析失败（T-068）
func TestParseDocxInvalid(t *testing.T) {
	if _, err := NewAIDocParser().Parse("bad.docx", []byte("not a zip")); err == nil {
		t.Fatal("非法 docx 应解析失败")
	}
}

// TestParseXlsxMultipleSheets xlsx 多 sheet 文本转换（T-068）
func TestParseXlsxMultipleSheets(t *testing.T) {
	f := excelize.NewFile()
	defer f.Close()
	idx, err := f.NewSheet("主机申请")
	if err != nil {
		t.Fatalf("新建 sheet 失败: %v", err)
	}
	f.SetActiveSheet(idx)
	_ = f.SetCellValue("主机申请", "A1", "字段")
	_ = f.SetCellValue("主机申请", "B1", "值")
	_ = f.SetCellValue("主机申请", "A2", "主机名")
	_ = f.SetCellValue("主机申请", "B2", "db-01")
	buf, err := f.WriteToBuffer()
	if err != nil {
		t.Fatalf("生成 xlsx 失败: %v", err)
	}

	text, err := NewAIDocParser().Parse("apply.xlsx", buf.Bytes())
	if err != nil {
		t.Fatalf("解析 xlsx 失败: %v", err)
	}
	for _, want := range []string{"工作表: 主机申请", "字段 | 值", "db-01"} {
		if !bytes.Contains([]byte(text), []byte(want)) {
			t.Fatalf("xlsx 文本缺少 %q，实际:\n%s", want, text)
		}
	}
}

// TestParseUnsupportedType 不支持的扩展名（T-068）
func TestParseUnsupportedType(t *testing.T) {
	if _, err := NewAIDocParser().Parse("a.pdf", []byte("x")); err != ErrUnsupportedFileType {
		t.Fatalf("期望 ErrUnsupportedFileType，实际 %v", err)
	}
}

// TestBuildExtractionPromptContainsFields 提示词注入字段清单（T-068）
func TestBuildExtractionPromptContainsFields(t *testing.T) {
	prompt := buildExtractionPrompt("测试文档内容")
	for _, want := range []string{DocTypeHostApply, DocTypeZeroTrustApply, "private_ip", "account_name", "测试文档内容"} {
		if !bytes.Contains([]byte(prompt), []byte(want)) {
			t.Fatalf("提示词缺少 %q", want)
		}
	}
}

// TestStripCodeFence 去除代码块包裹（T-068）
func TestStripCodeFence(t *testing.T) {
	in := "```json\n{\"doc_type\":\"host_apply\"}\n```"
	got := stripCodeFence(in)
	if got != `{"doc_type":"host_apply"}` {
		t.Fatalf("stripCodeFence 结果不符: %q", got)
	}
}
