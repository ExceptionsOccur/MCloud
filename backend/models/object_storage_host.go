package models

// ObjectStorageHost 对象存储 ↔ 主机配对行（T-088）：一个存储桶可被多台主机共享
type ObjectStorageHost struct {
	ID              uint `json:"id" gorm:"primaryKey"`
	ObjectStorageID uint `json:"object_storage_id" gorm:"index;not null;uniqueIndex:idx_object_storage_hosts_pair"`
	HostID          uint `json:"host_id" gorm:"index;not null;uniqueIndex:idx_object_storage_hosts_pair"`

	Host *Host `json:"host,omitempty" gorm:"foreignKey:HostID"`
}

func (ObjectStorageHost) TableName() string {
	return "object_storage_hosts"
}
