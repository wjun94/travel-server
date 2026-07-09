package model

import (
	"time"

	"travel-server/pkg/snowflake"

	"gorm.io/gorm"
)

// BrowseHistory 浏览历史
type BrowseHistory struct {
	ID         string    `gorm:"primaryKey" json:"id"`
	UserID     string    `gorm:"size:191;index" json:"userId"`
	TargetID   string    `gorm:"size:191" json:"targetId"`                // 浏览对象ID（如攻略ID）
	TargetType string    `gorm:"size:30;default:guide" json:"targetType"` // 浏览类型：guide/trip
	Title      string    `gorm:"size:200" json:"title"`                   // 快照标题
	CoverImage string    `gorm:"size:255" json:"coverImage"`              // 快照封面
	CreatedAt  time.Time `json:"createdAt"`
}

// BeforeCreate GORM 钩子
func (b *BrowseHistory) BeforeCreate(tx *gorm.DB) error {
	if b.ID == "" {
		b.ID = snowflake.GenerateID()
	}
	return nil
}
