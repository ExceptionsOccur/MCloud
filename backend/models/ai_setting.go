package models

import "time"

// AISetting AI 接入全局配置（单条：id 恒为 1）；api_key 以 AES-GCM 密文落库，不外泄
type AISetting struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	APIURL    string    `json:"api_url" gorm:"type:varchar(255)"`
	APIKeyEnc string    `json:"-" gorm:"column:api_key_enc;type:text"`
	Model     string    `json:"model" gorm:"type:varchar(128)"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
