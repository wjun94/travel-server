package model

import (
	"time"

	"travel-server/pkg/snowflake"

	"gorm.io/gorm"
)

// Follow 关注表
type Follow struct {
	ID         string    `gorm:"primaryKey" json:"id"`
	UserID     string    `gorm:"size:191;uniqueIndex:idx_follow" json:"userId"`     // 被关注者
	FollowerID string    `gorm:"size:191;uniqueIndex:idx_follow" json:"followerId"` // 关注者
	Status     int       `gorm:"default:0" json:"status"`                           // 0:正常 1:拉黑
	CreatedAt  time.Time `json:"createdAt"`
}

// BeforeCreate GORM 钩子
func (f *Follow) BeforeCreate(tx *gorm.DB) error {
	if f.ID == "" {
		f.ID = snowflake.GenerateID()
	}
	return nil
}
