package model

import "time"

// AdminUser 后台管理用户
type AdminUser struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"uniqueIndex;size:64;not null" json:"username"`
	PasswordHash string    `gorm:"size:255;not null" json:"-"` // 不对外暴露
	RoleID       uint      `json:"role_id"`
	Role         Role      `gorm:"foreignKey:RoleID" json:"role"` // 关联角色
	Status       int       `gorm:"default:1" json:"status"`       // 1启用 0禁用
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"-"`
}
