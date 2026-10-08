package models

import "time"

type Host struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	Region     string    `json:"region" gorm:"type:varchar(64);not null"`
	InstanceID string    `json:"instance_id" gorm:"type:varchar(128)"`
	Name       string    `json:"name" gorm:"type:varchar(128);not null"`
	PrivateIP  string    `json:"private_ip" gorm:"type:varchar(45);not null;uniqueIndex"`
	IPMapped   bool      `json:"ip_mapped" gorm:"default:false"`
	AssetType  string    `json:"asset_type" gorm:"type:varchar(32)"`
	OS         string    `json:"os" gorm:"type:varchar(64)"`
	CPU        int       `json:"cpu" gorm:"type:integer"`
	CPUArch    string    `json:"cpu_arch" gorm:"type:varchar(16)"`
	Memory     int       `json:"memory" gorm:"type:integer"`
	Disk       int       `json:"disk" gorm:"type:integer"`
	SystemDisk int       `json:"system_disk" gorm:"type:integer"`
	DataDisk   int       `json:"data_disk" gorm:"type:integer"`
	EnvType    string    `json:"env_type" gorm:"type:varchar(16)"`
	IsDBServer bool      `json:"is_db_server" gorm:"default:false"`
	Status     string    `json:"status" gorm:"type:varchar(32)"`
	OpenPorts  string    `json:"open_ports" gorm:"type:text"`
	Tags       string    `json:"tags" gorm:"type:text"`
	PersonID   *uint     `json:"person_id" gorm:"index"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`

	Application *HostApplication `json:"application,omitempty" gorm:"foreignKey:HostID"`
	Person      *Person          `json:"person,omitempty" gorm:"foreignKey:PersonID"`
	ZeroTrusts  []ZeroTrust      `json:"zero_trusts,omitempty" gorm:"foreignKey:HostID"`
	Domains     []Domain         `json:"domains,omitempty" gorm:"foreignKey:HostID"`
}
