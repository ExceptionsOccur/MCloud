package models

import "time"

type Host struct {
	ID         uint   `json:"id" gorm:"primaryKey"`
	Region     string `json:"region" gorm:"type:varchar(64);not null"`
	InstanceID string `json:"instance_id" gorm:"type:varchar(128)"`
	Name       string `json:"name" gorm:"type:varchar(128);not null"`
	PrivateIP  string `json:"private_ip" gorm:"type:varchar(45);not null;uniqueIndex"`
	IPMapped   bool   `json:"ip_mapped" gorm:"default:false"`
	IsExternal bool   `json:"is_external" gorm:"default:false"`
	AssetType  string `json:"asset_type" gorm:"type:varchar(32)"`
	OS         string `json:"os" gorm:"type:varchar(64)"`
	CPU        int    `json:"cpu" gorm:"type:integer"`
	CPUArch    string `json:"cpu_arch" gorm:"type:varchar(16)"`
	Memory     int    `json:"memory" gorm:"type:integer"`
	// HasObjectStorage 派生字段（T-088）：由对象存储台账配对自动重算，主机表单/导入不接受手填
	HasObjectStorage bool      `json:"has_object_storage" gorm:"column:has_object_storage;default:false"`
	SystemDisk       int       `json:"system_disk" gorm:"type:integer"`
	DataDisk         int       `json:"data_disk" gorm:"type:integer"`
	EnvType          string    `json:"env_type" gorm:"type:varchar(16)"`
	IsDBServer       bool      `json:"is_db_server" gorm:"default:false"`
	Status           string    `json:"status" gorm:"type:varchar(32)"`
	Tags             string    `json:"tags" gorm:"type:text"`
	PersonID         *uint     `json:"person_id" gorm:"index"`
	CreatedAt        time.Time `json:"created_at" gorm:"autoCreateTime"`

	Application  *HostApplication `json:"application,omitempty" gorm:"foreignKey:HostID"`
	Person       *Person          `json:"person,omitempty" gorm:"foreignKey:PersonID"`
	ZeroTrusts   []ZeroTrust      `json:"zero_trusts,omitempty" gorm:"-"`
	PortMappings []PortMapping    `json:"port_mappings,omitempty" gorm:"foreignKey:HostID"`
	// ObjectStorageLinks 该主机可访问的对象存储配对行（T-088；只读关联，删除主机侧重校验本表引用）
	ObjectStorageLinks []ObjectStorageHost `json:"-" gorm:"foreignKey:HostID"`
}
