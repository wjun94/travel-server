package model

// Role 后台角色
type Role struct {
	ID          string `gorm:"primaryKey" json:"id"`
	Name        string `gorm:"uniqueIndex;size:64;not null" json:"name"`
	Description string `gorm:"size:255" json:"description"`
	Permissions string `gorm:"type:text" json:"permissions"` // 权限列表JSON，如["dashboard","users_manage","posts_manage"]
}
