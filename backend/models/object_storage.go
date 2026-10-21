package models

import "time"

// ObjectStorage 对象存储台账（T-088）：外挂存储，不专属某台主机，可被多台主机共享访问
type ObjectStorage struct {
	ID      uint   `json:"id" gorm:"primaryKey"`
	Bucket  string `json:"bucket" gorm:"type:varchar(128);not null;uniqueIndex"`
	Purpose string `json:"purpose" gorm:"type:text;not null"`
	SizeGB  int    `json:"size_gb" gorm:"column:size_gb;type:integer;not null;default:0"`
	// AccessKey/SecretKeyEnc 访问存储桶的凭证：AccessKey 明文，SecretKey 以 AES-GCM 密文存储（T-088）
	AccessKey    string    `json:"access_key" gorm:"column:access_key;type:varchar(128);not null;default:''"`
	SecretKeyEnc string    `json:"-" gorm:"column:secret_key_enc;type:text;not null;default:''"`
	Remark       string    `json:"remark" gorm:"type:text"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	Hosts []ObjectStorageHost `json:"hosts" gorm:"foreignKey:ObjectStorageID"`
}

func (ObjectStorage) TableName() string {
	return "object_storages"
}
