package models

// ZeroTrustTarget 零信任接入目标子表（T-052）：zero_trusts.targets 拆出的 1NF 配对行，
// 一对「主机:端口」一行；双外键 RESTRICT（禁止删除被引用主机/台账记录）
type ZeroTrustTarget struct {
	ID          uint `json:"id" gorm:"primaryKey"`
	ZeroTrustID uint `json:"zero_trust_id" gorm:"index;not null"`
	HostID      uint `json:"host_id" gorm:"index;not null"`
	Port        int  `json:"port" gorm:"type:integer;not null"`

	ZeroTrust *ZeroTrust `json:"-" gorm:"foreignKey:ZeroTrustID;constraint:OnDelete:RESTRICT"`
	Host      *Host      `json:"-" gorm:"foreignKey:HostID;constraint:OnDelete:RESTRICT"`
}

// TableName 显式指定表名，避免 GORM 复数化生成异常
func (ZeroTrustTarget) TableName() string {
	return "zero_trust_targets"
}
