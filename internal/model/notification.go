package model

import (
	"time"

	"travel-server/pkg/snowflake"

	"gorm.io/gorm"
)

// Notification 通知记录
type Notification struct {
	ID         string    `gorm:"primaryKey" json:"id"`
	UserID     string    `gorm:"size:191;index" json:"userId"` // 接收者
	FromUserID string    `gorm:"size:191" json:"fromUserId"`   // 触发者（谁点的赞/谁关注的/谁申请的）
	Type       int       `gorm:"default:1" json:"type"`        // 1搭子申请 2攻略点赞 3新增关注 4系统通知 5评论点赞
	RelatedID  string    `gorm:"size:191" json:"relatedId"`    // 关联的单据ID（搭子申请/点赞/关注 记录ID）
	IsRead     int       `gorm:"default:0" json:"isRead"`      // 0未读 1已读
	Content    string    `gorm:"type:text" json:"content"`     // 通知内容简述
	Title      string    `gorm:"size:100" json:"title"`        // 标题（type=4 系统通知用）
	LinkURL    string    `gorm:"size:500" json:"linkUrl"`      // 跳转链接（type=4 系统通知用，可空）
	CreatedAt  time.Time `json:"createdAt"`
}

// BeforeCreate GORM 钩子
func (n *Notification) BeforeCreate(tx *gorm.DB) error {
	if n.ID == "" {
		n.ID = snowflake.GenerateID()
	}
	return nil
}
