package model

import (
	"time"

	"travel-server/pkg/snowflake"

	"gorm.io/gorm"
)

// Favorite 收藏表
type Favorite struct {
	ID         string    `gorm:"primaryKey" json:"id"`
	UserID     string    `gorm:"size:191;uniqueIndex:uk_user_target" json:"userId"`    // 用户ID
	TargetType string    `gorm:"size:20;uniqueIndex:uk_user_target" json:"targetType"` // 收藏类型：guide/trip/partner
	TargetID   string    `gorm:"size:191;uniqueIndex:uk_user_target" json:"targetId"`  // 收藏对象ID
	CreatedAt  time.Time `json:"createdAt"`
}

// BeforeCreate GORM 钩子
func (f *Favorite) BeforeCreate(tx *gorm.DB) error {
	if f.ID == "" {
		f.ID = snowflake.GenerateID()
	}
	return nil
}
