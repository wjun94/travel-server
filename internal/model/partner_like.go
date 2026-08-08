package model

import (
	"time"

	"travel-server/pkg/snowflake"

	"gorm.io/gorm"
)

// PartnerLike 搭子点赞表（与收藏分离，独立记录点赞关系）
type PartnerLike struct {
	ID        string    `gorm:"primaryKey" json:"id"`
	UserID    string    `gorm:"size:191;uniqueIndex:uk_user_partner" json:"userId"`
	PartnerID string    `gorm:"size:191;uniqueIndex:uk_user_partner" json:"partnerId"`
	CreatedAt time.Time `json:"createdAt"`
}

// BeforeCreate GORM 钩子
func (pl *PartnerLike) BeforeCreate(tx *gorm.DB) error {
	if pl.ID == "" {
		pl.ID = snowflake.GenerateID()
	}
	return nil
}
