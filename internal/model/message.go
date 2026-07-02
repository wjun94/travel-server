package model

import (
	"time"

	"travel-server/pkg/snowflake"

	"gorm.io/gorm"
)

// Message 消息（私聊或系统通知）
type Message struct {
	ID         string    `gorm:"primaryKey" json:"id"`
	FromUserID string    `gorm:"size:191" json:"fromUserId"` // 发送者 ID
	ToUserID   string    `gorm:"size:191" json:"toUserId"`   // 接收者 ID
	Content    string    `gorm:"type:text" json:"content"`
	Type       int       `gorm:"default:1" json:"type"`   // 1私聊 2系统通知
	IsRead     int       `gorm:"default:0" json:"isRead"` // 0未读 1已读
	CreatedAt  time.Time `json:"createdAt"`
}

// BeforeCreate GORM 钩子
func (m *Message) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = snowflake.GenerateID()
	}
	return nil
}
