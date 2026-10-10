package models

import "time"

// AuditLog 审计日志（只增不删；无 UNIQUE 约束、无 users 外键）
type AuditLog struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	OperatorID   uint      `json:"operator_id" gorm:"index"`
	OperatorName string    `json:"operator_name" gorm:"type:varchar(64)"`
	Action       string    `json:"action" gorm:"type:varchar(32);index"`
	ResourceType string    `json:"resource_type" gorm:"type:varchar(32);index"`
	ResourceID   *uint     `json:"resource_id" gorm:"index"`
	Detail       []byte    `json:"detail" gorm:"type:jsonb"`
	RequestID    string    `json:"request_id" gorm:"type:varchar(64);index"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
}
