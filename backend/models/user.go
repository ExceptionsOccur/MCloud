package models

import "time"

type User struct {
	ID             uint       `json:"id" gorm:"primaryKey"`
	Username       string     `json:"username" gorm:"type:varchar(64);not null;uniqueIndex"`
	PasswordHash   string     `json:"-" gorm:"type:varchar(256);not null"`
	FailedAttempts int        `json:"-" gorm:"default:0"`
	LockedUntil    *time.Time `json:"-" gorm:"type:timestamptz"`
	CreatedAt      time.Time  `json:"created_at" gorm:"autoCreateTime"`
	LastLogin      *time.Time `json:"last_login" gorm:"type:timestamptz"`
}
