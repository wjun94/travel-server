package model

import (
	"travel-server/pkg/snowflake"

	"gorm.io/gorm"
)

// Role 后台角色
type Role struct {
	ID          string `gorm:"primaryKey" json:"id"`
	Name        string `gorm:"uniqueIndex;size:64;not null" json:"name"`
	Description string `gorm:"size:255" json:"description"`
	Permissions string `gorm:"type:text" json:"permissions"` // 权限列表JSON，如["dashboard","users_manage","posts_manage"]
}

// BeforeCreate GORM 钩子：创建前自动生成雪花 ID
func (r *Role) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = snowflake.GenerateID()
	}
	return nil
}
