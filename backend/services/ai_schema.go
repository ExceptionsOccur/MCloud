package services

import "strings"

// AI 申请表单提取 schema（服务端权威定义，T-068）。
// 大模型先按文档内容识别 doc_type，再按对应字段清单提取；字段名严格对齐各模块 CreateRequest，
// 前端确认后可直接调用现有 Create 接口入库。

// 文档类型枚举
const (
	DocTypeHostApply      = "host_apply"       // 主机申请
	DocTypeZeroTrustApply = "zero_trust_apply" // 零信任申请
)

// AIDocType 文档类型定义
type AIDocType struct {
	Key    string    // 枚举值
	Label  string    // 中文名
	Fields []AIField // 可提取字段
}

// AIField 字段定义：key 为 CreateRequest JSON 字段名，label 为中文提示
type AIField struct {
	Key   string
	Label string
	// Hint 供提示词参考的取值/格式说明（可选）
	Hint string
}

// hostApplyFields 主机申请字段（对齐 services.CreateHostRequest）
var hostApplyFields = []AIField{
	{Key: "region", Label: "区域", Hint: "如 region-a"},
	{Key: "name", Label: "主机名", Hint: "必填"},
	{Key: "private_ip", Label: "内网IP", Hint: "必填，IPv4"},
	{Key: "instance_id", Label: "实例ID"},
	{Key: "asset_type", Label: "资产类型", Hint: "虚拟机/裸金属服务器"},
	{Key: "os", Label: "操作系统"},
	{Key: "cpu", Label: "CPU核数", Hint: "整数"},
	{Key: "cpu_arch", Label: "CPU架构", Hint: "C86/X86/ARM"},
	{Key: "memory", Label: "内存(GB)", Hint: "整数"},
	{Key: "object_storage", Label: "对象存储(GB)", Hint: "整数"},
	{Key: "system_disk", Label: "系统盘(GB)", Hint: "整数"},
	{Key: "data_disk", Label: "数据盘(GB)", Hint: "整数"},
	{Key: "env_type", Label: "环境类型", Hint: "测试/生产"},
	{Key: "is_db_server", Label: "是否数据库服务器", Hint: "true/false"},
	{Key: "is_external", Label: "是否云外主机", Hint: "默认 false（云内）；true 表示云外服务器"},
	{Key: "status", Label: "状态", Hint: "运行中/已停止/已关机/待确认"},
	{Key: "open_ports", Label: "开放端口"},
	{Key: "tags", Label: "标签"},
	{Key: "apply_unit", Label: "申请单位"},
	{Key: "applicant", Label: "申请人"},
	{Key: "applicant_contact", Label: "申请人联系方式"},
	{Key: "project", Label: "项目名称"},
	{Key: "apply_reason", Label: "申请理由"},
	{Key: "apply_config", Label: "申请配置"},
	{Key: "apply_time", Label: "申请时间", Hint: "严格 YYYY-MM-DD，如 2026-10-16"},
	{Key: "object_storage_size", Label: "对象存储大小"},
	{Key: "remark", Label: "备注"},
}

// zeroTrustApplyFields 零信任申请字段（对齐 services.ZeroTrustRequest）
var zeroTrustApplyFields = []AIField{
	{Key: "apply_unit", Label: "申请单位", Hint: "必填"},
	{Key: "account_name", Label: "账户名", Hint: "必填"},
	{Key: "account_name_zh", Label: "账号姓名"},
	{Key: "applicant", Label: "申请人"},
	{Key: "contact", Label: "账户联系方式"},
	{Key: "exit_location", Label: "接入地区", Hint: "多值用英文逗号分隔"},
	{Key: "targets", Label: "接入目标", Hint: `数组，每项 {"host_ip":"内网IP或主机名","ports":"端口表达式,如22或1000-10005","protocol":"tcp或udp"}`},
	{Key: "system_name", Label: "系统名称"},
	{Key: "apply_time", Label: "申请时间", Hint: "严格 YYYY-MM-DD"},
	{Key: "remark", Label: "备注"},
}

// aiDocTypes 全部支持类型
var aiDocTypes = []AIDocType{
	{Key: DocTypeHostApply, Label: "主机申请", Fields: hostApplyFields},
	{Key: DocTypeZeroTrustApply, Label: "零信任申请", Fields: zeroTrustApplyFields},
}

// docTypeByKey 按枚举取类型定义
func docTypeByKey(key string) (AIDocType, bool) {
	for _, t := range aiDocTypes {
		if t.Key == key {
			return t, true
		}
	}
	return AIDocType{}, false
}

// buildExtractionPrompt 构造分类 + 提取提示词（注入字段清单与文档文本）
func buildExtractionPrompt(text string) string {
	var sb strings.Builder
	sb.WriteString("你是资产管理系统 MCloud 的申请表单信息提取助手。请根据文档内容：\n")
	sb.WriteString("1. 判断文档属于以下哪种类型（doc_type）：\n")
	for _, t := range aiDocTypes {
		sb.WriteString("   - " + t.Key + "（" + t.Label + "）\n")
	}
	sb.WriteString("   若都不匹配，doc_type 返回 \"unknown\"。\n")
	sb.WriteString("2. 仅提取该类型下列出的字段（fields 为对象；未在文档中出现的字段不要臆造，省略或留空）。\n")
	sb.WriteString("3. 严格返回 JSON，结构为：{\"doc_type\": \"...\", \"fields\": {...}, \"warnings\": [\"...\"]}\n")
	sb.WriteString("   warnings 用于说明缺失的必填字段、格式异常或不确定项；无则空数组。\n\n")

	sb.WriteString("各类型字段清单（key: 中文名（提示））：\n")
	for _, t := range aiDocTypes {
		sb.WriteString("- " + t.Key + "（" + t.Label + "）：\n")
		for _, f := range t.Fields {
			line := "    " + f.Key + ": " + f.Label
			if f.Hint != "" {
				line += "（" + f.Hint + "）"
			}
			sb.WriteString(line + "\n")
		}
	}

	sb.WriteString("\n=== 文档内容开始 ===\n")
	sb.WriteString(text)
	sb.WriteString("\n=== 文档内容结束 ===")
	return sb.String()
}
