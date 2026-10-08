package models

import "time"

// ZeroTrust 零信任台账记录，targets 为 "host_id:port" 逗号分隔多组配对；public_ip 关联公网IP资源池
type ZeroTrust struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	ApplyUnit   string    `json:"apply_unit" gorm:"type:varchar(128);not null"`
	AccountName string    `json:"account_name" gorm:"type:varchar(64);not null"`
	Contact     string    `json:"contact" gorm:"type:varchar(64)"`
	PublicIP    string    `json:"public_ip" gorm:"type:varchar(45)"`
	Targets     string    `json:"targets" gorm:"type:text;not null"`
	SystemName  string    `json:"system_name" gorm:"type:varchar(128)"`
	ApplyTime   time.Time `json:"apply_time" gorm:"not null"`
	Remark      string    `json:"remark" gorm:"type:text"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// TableName 显式指定表名，避免 GORM 复数化生成异常
func (ZeroTrust) TableName() string {
	return "zero_trusts"
}
