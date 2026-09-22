package models

import "time"

type IPSubnet struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CIDR      string    `json:"cidr" gorm:"column:cidr;type:varchar(32);not null;uniqueIndex"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}
