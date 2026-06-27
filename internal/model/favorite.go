package model

import (
	"time"

	"travel-server/pkg/snowflake"
)

// Favorite 收藏表
type Favorite struct {
	ID         string    `gorm:"primaryKey" json:"id"`
	UserID     string    `json:"userId"`                                               // 用户ID
	TargetType string    `gorm:"size:20;uniqueIndex:uk_user_target" json:"targetType"` // 收藏类型：guide/trip
	TargetID   string    `gorm:"uniqueIndex:uk_user_target" json:"targetId"`           // 收藏对象ID
	CreatedAt  time.Time `json:"createdAt"`
}

// BeforeCreate GORM 钩子
func (f *Favorite) BeforeCreate() error {
	if f.ID == "" {
		f.ID = snowflake.GenerateID()
	}
	return nil
}
