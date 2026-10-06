package models

import "time"

type CloudResource struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	Region        string    `json:"region" gorm:"type:varchar(64);not null;uniqueIndex"`
	PhysicalCPU   int       `json:"physical_cpu" gorm:"type:integer;default:0"`
	Vcpu          int       `json:"vcpu" gorm:"type:integer;default:0"`
	Memory        int       `json:"memory" gorm:"type:integer;default:0"`
	Storage       int       `json:"storage" gorm:"type:integer;default:0"`
	BareMetal     int       `json:"bare_metal" gorm:"type:integer;default:0"`
	GpuCardCount  int       `json:"gpu_card_count" gorm:"type:integer;default:0"`
	ObjectStorage int       `json:"object_storage" gorm:"type:integer;default:0"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
}
