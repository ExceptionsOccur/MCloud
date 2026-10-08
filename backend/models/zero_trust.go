package models

import "time"

type ZeroTrust struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	ApplyUnit   string    `json:"apply_unit" gorm:"type:varchar(128);not null"`
	AccountName string    `json:"account_name" gorm:"type:varchar(64);not null"`
	Contact     string    `json:"contact" gorm:"type:varchar(64)"`
	HostID      uint      `json:"host_id" gorm:"index;not null"`
	Port        int       `json:"port" gorm:"not null"`
	ApplyTime   time.Time `json:"apply_time" gorm:"not null"`
	Remark      string    `json:"remark" gorm:"type:text"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`

	Host *Host `json:"host,omitempty" gorm:"foreignKey:HostID"`
}

// TableName 显式指定表名，避免 GORM 复数化生成异常
func (ZeroTrust) TableName() string {
	return "zero_trusts"
}
