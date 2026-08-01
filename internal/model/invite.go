package model

import (
	"time"

	"travel-server/pkg/snowflake"

	"gorm.io/gorm"
)

// InviteRecord 邀请记录表 — 新用户通过邀请码注册成功即产生一条记录
type InviteRecord struct {
	ID        string    `gorm:"primaryKey" json:"id"`
	InviterID string    `gorm:"size:191;index" json:"inviterId"`       // 邀请人ID
	InviteeID string    `gorm:"size:191;uniqueIndex" json:"inviteeId"` // 被邀请人ID（唯一，防重复绑定）
	CreatedAt time.Time `json:"createdAt"`
}

// BeforeCreate GORM 钩子
func (ir *InviteRecord) BeforeCreate(tx *gorm.DB) error {
	if ir.ID == "" {
		ir.ID = snowflake.GenerateID()
	}
	return nil
}
