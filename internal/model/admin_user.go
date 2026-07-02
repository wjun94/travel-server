package model

import (
	"time"

	"travel-server/pkg/snowflake"

	"gorm.io/gorm"
)

// AdminUser 后台管理用户
type AdminUser struct {
	ID           string    `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"uniqueIndex;size:64;not null" json:"username"`
	PasswordHash string    `gorm:"size:255;not null" json:"-"` // 不对外暴露
	RoleID       string    `gorm:"size:191" json:"roleId"`
	Role         Role      `gorm:"foreignKey:RoleID" json:"role"` // 关联角色
	Status       int       `gorm:"default:1" json:"status"`       // 1启用 0禁用
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"-"`
}

// BeforeCreate GORM 钩子：创建前自动生成 6 位雪花短 ID
func (a *AdminUser) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = snowflake.GenerateShortID(6)
	}
	return nil
}
