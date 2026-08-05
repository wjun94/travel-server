package model

import (
	"time"

	"travel-server/pkg/snowflake"

	"gorm.io/gorm"
)

// Message 消息（私聊或系统通知）
type Message struct {
	ID                string    `gorm:"primaryKey" json:"id"`
	FromUserID        string    `gorm:"size:191" json:"fromUserId"` // 发送者 ID
	ToUserID          string    `gorm:"size:191" json:"toUserId"`   // 接收者 ID
	Content           string    `gorm:"type:text" json:"content"`
	Type              int       `gorm:"default:1" json:"type"`   // 1私聊 2系统通知
	IsRead            int       `gorm:"default:0" json:"isRead"` // 0未读 1已读
	DeletedBySender   int       `gorm:"default:0" json:"-"`      // 发送者已删除（清空/删除会话时置1）
	DeletedByReceiver int       `gorm:"default:0" json:"-"`      // 接收者已删除（清空/删除会话时置1）
	CreatedAt         time.Time `json:"createdAt"`
}

// ChatSession 私聊会话（用于会话列表展示与清空/删除控制）
type ChatSession struct {
	ID        string         `gorm:"primaryKey;size:64" json:"id"`
	UserID    string         `gorm:"size:191;uniqueIndex:uk_user_peer" json:"userId"` // 所属用户
	PeerID    string         `gorm:"size:191;uniqueIndex:uk_user_peer" json:"peerId"` // 对方用户ID
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"` // 软删除（删除会话时标记）
}

// BeforeCreate GORM 钩子
func (m *Message) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = snowflake.GenerateID()
	}
	return nil
}

// BeforeCreate GORM 钩子
func (cs *ChatSession) BeforeCreate(tx *gorm.DB) error {
	if cs.ID == "" {
		cs.ID = snowflake.GenerateID()
	}
	return nil
}
