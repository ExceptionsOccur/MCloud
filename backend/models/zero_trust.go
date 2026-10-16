package models

import "time"

// ZeroTrust 零信任台账记录；接入目标存于 zero_trust_targets 子表（T-052），
// Targets 为派生字段（gorm:"-"，读时组装 "host_id:port,..." 保 API 兼容）；
// T-061 起去公网IP列、接入地区独立存储（下拉候选值取 public_ips.exit_location 去重）；
// T-065 起接入地区支持多个（TEXT，逗号分隔多值）；
// T-063 新增 account_name_zh（账号姓名，account_name 账户名的补充）；
// T-065 起申请时间只记录到日期（VARCHAR(10) 严格 YYYY-MM-DD，对齐 host_applications.apply_time）
type ZeroTrust struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	ApplyUnit     string    `json:"apply_unit" gorm:"type:varchar(128);not null"`
	AccountName   string    `json:"account_name" gorm:"type:varchar(64);not null"`
	AccountNameZh string    `json:"account_name_zh" gorm:"type:varchar(64)"`
	Applicant     string    `json:"applicant" gorm:"type:varchar(64)"`
	Contact       string    `json:"contact" gorm:"type:varchar(64)"`
	ExitLocation  string    `json:"exit_location" gorm:"type:text"`
	Targets       string    `json:"targets" gorm:"-"`
	SystemName    string    `json:"system_name" gorm:"type:varchar(128)"`
	ApplyTime     string    `json:"apply_time" gorm:"type:varchar(10);not null"`
	Remark        string    `json:"remark" gorm:"type:text"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// TableName 显式指定表名，避免 GORM 复数化生成异常
func (ZeroTrust) TableName() string {
	return "zero_trusts"
}
