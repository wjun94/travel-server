package model

import "time"

// Message 消息（私聊或系统通知）
type Message struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	FromUserID uint      `json:"from_user_id"` // 发送者 ID
	ToUserID   uint      `json:"to_user_id"`   // 接收者 ID
	Content    string    `gorm:"type:text" json:"content"`
	Type       int       `gorm:"default:1" json:"type"`    // 1私聊 2系统通知
	IsRead     int       `gorm:"default:0" json:"is_read"` // 0未读 1已读
	CreatedAt  time.Time `json:"created_at"`
}
