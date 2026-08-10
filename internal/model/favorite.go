package model

import (
	"time"

	"travel-server/pkg/snowflake"

	"gorm.io/gorm"
)

// Favorite 收藏/点赞表（Action 区分：1点赞 2收藏，同一用户对同一对象可同时点赞与收藏）
type Favorite struct {
	ID         string    `gorm:"primaryKey" json:"id"`
	UserID     string    `gorm:"size:191;uniqueIndex:uk_user_target_action" json:"userId"`    // 用户ID
	TargetType string    `gorm:"size:20;uniqueIndex:uk_user_target_action" json:"targetType"` // 收藏类型：guide/trip/partner
	TargetID   string    `gorm:"size:191;uniqueIndex:uk_user_target_action" json:"targetId"`  // 收藏对象ID
	Action     int       `gorm:"default:2;uniqueIndex:uk_user_target_action" json:"action"`   // 1点赞 2收藏
	CreatedAt  time.Time `json:"createdAt"`
}

// BeforeCreate GORM 钩子
func (f *Favorite) BeforeCreate(tx *gorm.DB) error {
	if f.ID == "" {
		f.ID = snowflake.GenerateID()
	}
	return nil
}
