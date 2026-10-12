package models

import "time"

// AuditLog 审计日志（只增不删；无 UNIQUE 约束、无 users 外键）
// NOT NULL 口径与迁移 20261010000001_create_audit_logs.sql 一致（T-053）；resource_id/detail 保持可空
type AuditLog struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	OperatorID   uint      `json:"operator_id" gorm:"index;not null"`
	OperatorName string    `json:"operator_name" gorm:"type:varchar(64);not null"`
	Action       string    `json:"action" gorm:"type:varchar(32);index;not null"`
	ResourceType string    `json:"resource_type" gorm:"type:varchar(32);index;not null"`
	ResourceID   *uint     `json:"resource_id" gorm:"index"`
	Detail       []byte    `json:"detail" gorm:"type:jsonb"`
	RequestID    string    `json:"request_id" gorm:"type:varchar(64);index;not null"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime;not null"`
}
