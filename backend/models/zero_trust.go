package models

import "time"

// ZeroTrust 零信任台账记录；接入目标存于 zero_trust_targets 子表（T-052），
// Targets 为派生字段（gorm:"-"，读时组装 "host_id:port,..." 保 API 兼容）；
// T-061 起去公网IP列、接入地区独立存储（下拉候选值取 public_ips.exit_location 去重）
type ZeroTrust struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	ApplyUnit    string    `json:"apply_unit" gorm:"type:varchar(128);not null"`
	AccountName  string    `json:"account_name" gorm:"type:varchar(64);not null"`
	Applicant    string    `json:"applicant" gorm:"type:varchar(64)"`
	Contact      string    `json:"contact" gorm:"type:varchar(64)"`
	ExitLocation string    `json:"exit_location" gorm:"type:varchar(128)"`
	Targets      string    `json:"targets" gorm:"-"`
	SystemName   string    `json:"system_name" gorm:"type:varchar(128)"`
	ApplyTime    time.Time `json:"apply_time" gorm:"not null"`
	Remark       string    `json:"remark" gorm:"type:text"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// TableName 显式指定表名，避免 GORM 复数化生成异常
func (ZeroTrust) TableName() string {
	return "zero_trusts"
}
