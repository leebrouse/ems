package model

import (
	"time"
)

// RefreshToken 表示用户刷新令牌
type RefreshToken struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	UserID    int64     `gorm:"not null;index"`
	Token     string    `gorm:"size:512;not null;uniqueIndex"`
	ExpiresAt time.Time `gorm:"not null"`
	Revoked   bool      `gorm:"default:false"`
	CreatedAt time.Time
}

// TableName 指定刷新令牌表名
func (RefreshToken) TableName() string {
	return "refresh_tokens"
}
