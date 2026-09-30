package models

import "time"

type Person struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"type:varchar(64);not null"`
	Contact   string    `json:"contact" gorm:"type:varchar(64)"`
	Unit      string    `json:"unit" gorm:"type:varchar(128)"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// TableName 显式指定表名，避免 GORM 复数化生成 people
func (Person) TableName() string {
	return "persons"
}
