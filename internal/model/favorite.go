package model

import "time"

// Favorite 收藏表
type Favorite struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `json:"userId"`                                               // 用户ID
	TargetType string    `gorm:"size:20;uniqueIndex:uk_user_target" json:"targetType"` // 收藏类型：guide/trip
	TargetID   uint      `gorm:"uniqueIndex:uk_user_target" json:"targetId"`           // 收藏对象ID
	CreatedAt  time.Time `json:"createdAt"`
}
