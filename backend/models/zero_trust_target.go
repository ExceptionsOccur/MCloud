package models

// ZeroTrustTarget 零信任接入目标子表（T-052，T-056 端口列改表达式，T-065 加协议）：
// zero_trusts.targets 拆出的 1NF 配对行，一对「主机:端口表达式」一行
// （ports 存单端口/范围/列表，如 `22` 或 `1000-10005,22`；protocol 为该目标行单协议 tcp/udp，缺省 tcp）；
// 双外键 RESTRICT（禁止删除被引用主机/台账记录）
type ZeroTrustTarget struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	ZeroTrustID uint   `json:"zero_trust_id" gorm:"index;not null"`
	HostID      uint   `json:"host_id" gorm:"index;not null"`
	Ports       string `json:"ports" gorm:"type:text;not null"`
	Protocol    string `json:"protocol" gorm:"type:varchar(8);not null;default:tcp"`

	ZeroTrust *ZeroTrust `json:"-" gorm:"foreignKey:ZeroTrustID;constraint:OnDelete:RESTRICT"`
	Host      *Host      `json:"-" gorm:"foreignKey:HostID;constraint:OnDelete:RESTRICT"`
}

// TableName 显式指定表名，避免 GORM 复数化生成异常
func (ZeroTrustTarget) TableName() string {
	return "zero_trust_targets"
}
