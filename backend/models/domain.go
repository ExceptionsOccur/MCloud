package models

import "time"

type Domain struct {
	ID        uint       `json:"id" gorm:"primaryKey"`
	Domain    string     `json:"domain" gorm:"type:varchar(255);not null;uniqueIndex"`
	PublicIP  string     `json:"public_ip" gorm:"type:varchar(45)"`
	Provider  string     `json:"provider" gorm:"type:varchar(128)"`
	ExpiresAt *time.Time `json:"expires_at"`
	Remark    string     `json:"remark" gorm:"type:text"`
	CreatedAt time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

func (Domain) TableName() string {
	return "domains"
}
