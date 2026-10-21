package services

// 数据备份（T-038）：8 张业务表统一导出(xlsx)/导入(upsert)
//
// sheet 顺序 = 导入依赖顺序：后表通过自然键引用前表
// （人员/公网IP/云资源/网段 → 主机 → 申请信息 → 零信任 → 端口映射）。

// sheetSpec 单个工作表定义：Name 为 xlsx sheet 名（中文，导入按名匹配），Headers 为列头（导入按名匹配列）
type sheetSpec struct {
	Name    string
	Headers []string
}

// dataExchangeSheets 8 表定义，顺序即导出 sheet 顺序与导入处理顺序；
// T-063 起 sheet 名改为中文（人员/公网IP/云资源/网段/主机/主机申请信息/零信任/端口映射），导入按中文名匹配（不再兼容旧英文名备份）
var dataExchangeSheets = []sheetSpec{
	{Name: "人员", Headers: []string{"姓名", "联系方式", "单位"}},
	{Name: "公网IP", Headers: []string{"公网IP", "运营商", "出口位置", "备注"}},
	{Name: "云资源", Headers: []string{"区域", "物理CPU", "VCPU", "内存", "存储", "裸金属", "GPU卡数", "对象存储"}},
	{Name: "网段", Headers: []string{"网段CIDR"}},
	{Name: "主机", Headers: []string{
		"区域", "实例ID", "主机名称", "内网IP", "资产类型", "操作系统",
		"CPU核数", "CPU架构", "内存", "对象存储", "系统盘", "数据盘",
		"环境类型", "是否数据库服务器", "是否云外主机", "状态", "开放端口", "标签",
		"人员姓名", "人员联系方式",
	}},
	{Name: "主机申请信息", Headers: []string{
		"内网IP", "申请单位", "申请人", "申请人联系方式", "所属项目",
		"申请理由", "申请配置", "申请时间", "备注",
	}},
	{Name: "零信任", Headers: []string{
		"申请单位", "账户名", "账号姓名", "申请人", "账户联系方式", "接入地区",
		"内网IP", "tcp端口", "udp端口", "系统名称", "申请时间", "备注",
	}},
	{Name: "端口映射", Headers: []string{
		"公网IP", "内网IP", "外网端口", "内网端口", "外网协议", "内网协议", "域名", "备注",
	}},
}

// RowError 导入行级错误（定位到 sheet + Excel 行号）
type RowError struct {
	Sheet   string `json:"sheet"`
	Row     int    `json:"row"`
	Message string `json:"message"`
}

// SheetStat 单 sheet 导入统计（created+updated+skipped = 该表数据行数）
type SheetStat struct {
	Sheet   string `json:"sheet"`
	Rows    int    `json:"rows"`
	Created int    `json:"created"`
	Updated int    `json:"updated"`
	Skipped int    `json:"skipped"`
}

// ImportReport 导入结果：committed=false 表示存在行级错误、已整体回滚（库数据不变）
type ImportReport struct {
	Committed bool        `json:"committed"`
	Sheets    []SheetStat `json:"sheets"`
	Errors    []RowError  `json:"errors"`
}

// DataExchangeService 8 表统一导出/导入
type DataExchangeService struct{}

func NewDataExchangeService() *DataExchangeService {
	return &DataExchangeService{}
}
