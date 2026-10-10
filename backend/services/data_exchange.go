package services

// 数据备份（T-038）：8 张业务表统一导出(xlsx)/导入(upsert)
//
// sheet 顺序 = 导入依赖顺序：后表通过自然键引用前表
// （人员/公网IP/云资源/网段 → 主机 → 申请信息 → 零信任 → 端口映射）。

// sheetSpec 单个工作表定义：Name 为 xlsx sheet 名（机器稳定，导入按名匹配），Headers 为列头（导入按名匹配列）
type sheetSpec struct {
	Name    string
	Headers []string
}

// dataExchangeSheets 8 表定义，顺序即导出 sheet 顺序与导入处理顺序
var dataExchangeSheets = []sheetSpec{
	{Name: "persons", Headers: []string{"姓名", "联系方式", "单位"}},
	{Name: "public_ips", Headers: []string{"公网IP", "运营商", "出口位置", "备注"}},
	{Name: "cloud_resources", Headers: []string{"区域", "物理CPU", "VCPU", "内存", "存储", "裸金属", "GPU卡数", "对象存储"}},
	{Name: "ip_subnets", Headers: []string{"网段CIDR"}},
	{Name: "hosts", Headers: []string{
		"区域", "实例ID", "主机名称", "内网IP", "资产类型", "操作系统",
		"CPU核数", "CPU架构", "内存", "磁盘", "系统盘", "数据盘",
		"环境类型", "是否数据库服务器", "状态", "开放端口", "标签",
		"人员姓名", "人员联系方式",
	}},
	{Name: "host_applications", Headers: []string{
		"内网IP", "申请单位", "申请人", "申请人联系方式", "所属项目",
		"申请理由", "申请配置", "申请时间", "对象存储大小", "备注",
	}},
	{Name: "zero_trusts", Headers: []string{
		"申请单位", "账户名", "联系方式", "公网IP", "接入目标", "系统名称", "申请时间", "备注",
	}},
	{Name: "port_mappings", Headers: []string{
		"公网IP", "内网IP", "外网端口", "内网端口", "域名", "备注",
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
