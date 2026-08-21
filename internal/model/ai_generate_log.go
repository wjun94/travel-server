package model

import (
	"time"

	"gorm.io/gorm"

	"travel-server/pkg/snowflake"
)

// AiGenerateLog AI生成统计日志（小程序点击AI生成时记录一次）
type AiGenerateLog struct {
	ID        string    `gorm:"primaryKey" json:"id"`
	UserID    string    `gorm:"size:191;index" json:"userId"` // 触发用户
	Type      string    `gorm:"size:20;index" json:"type"`    // trip 行程 / partner 搭子
	Status    int       `gorm:"default:0" json:"status"`      // 0已点击 1生成成功 2生成失败
	CreatedAt time.Time `json:"createdAt"`
}

// BeforeCreate GORM 钩子
func (l *AiGenerateLog) BeforeCreate(tx *gorm.DB) error {
	if l.ID == "" {
		l.ID = snowflake.GenerateID()
	}
	return nil
}
