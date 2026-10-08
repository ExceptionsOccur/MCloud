package models

import "time"

type Domain struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Domain    string    `json:"domain" gorm:"type:varchar(255);not null;uniqueIndex"`
	PublicIP  string    `json:"public_ip" gorm:"type:varchar(45)"`
	ISP       string    `json:"isp" gorm:"type:varchar(128)"`
	HostID    uint      `json:"host_id" gorm:"index"`
	HostPort  int       `json:"host_port"`
	Remark    string    `json:"remark" gorm:"type:text"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	Host *Host `json:"host,omitempty" gorm:"foreignKey:HostID"`
}

func (Domain) TableName() string {
	return "domains"
}
