package models

import "time"

type PublicIP struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	IP           string    `json:"ip" gorm:"type:varchar(45);not null;uniqueIndex"`
	ISP          string    `json:"isp" gorm:"type:varchar(128)"`
	ExitLocation string    `json:"exit_location" gorm:"type:varchar(128)"`
	Remark       string    `json:"remark" gorm:"type:text"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (PublicIP) TableName() string {
	return "public_ips"
}
