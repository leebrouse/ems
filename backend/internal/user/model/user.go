package model

import (
	"time"
)

// User 用户模型
type User struct {
	ID           int64  `gorm:"primaryKey;autoIncrement"`
	Username     string `gorm:"size:64;not null;uniqueIndex"`
	PasswordHash string `gorm:"size:255;not null"`
	IsActive     bool   `gorm:"default:true"`
	CreatedAt    time.Time
	UpdatedAt    time.Time

	Roles []Role `gorm:"many2many:user_roles;"`
}

// Role 角色模型
type Role struct {
	ID          int64  `gorm:"primaryKey;autoIncrement"`
	Name        string `gorm:"size:64;not null;uniqueIndex"`
	Description string `gorm:"size:255"`
}

// TableName 指定用户表名
func (User) TableName() string {
	return "users"
}
