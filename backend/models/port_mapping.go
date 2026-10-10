package models

import "time"

type PortMapping struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	PublicIP      string    `json:"public_ip" gorm:"type:varchar(45);not null"`
	HostID        uint      `json:"host_id" gorm:"index;not null"`
	ExternalPorts string    `json:"external_ports" gorm:"type:text;not null"`
	InternalPorts string    `json:"internal_ports" gorm:"type:text;not null"`
	Domain        string    `json:"domain" gorm:"type:varchar(255);uniqueIndex:idx_port_mappings_domain"`
	ISP           string    `json:"isp" gorm:"type:varchar(128)"`
	ExitLocation  string    `json:"exit_location" gorm:"type:varchar(128)"`
	Remark        string    `json:"remark" gorm:"type:text"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	Host *Host `json:"host,omitempty" gorm:"foreignKey:HostID"`
}

func (PortMapping) TableName() string {
	return "port_mappings"
}
