package model

import (
	"time"

	"travel-server/pkg/snowflake"
)

// Comment 评论表
type Comment struct {
	ID         string    `gorm:"primaryKey" json:"id"`
	UserID     string    `json:"userId"`                     // 评论者
	TargetType string    `gorm:"size:20" json:"targetType"`  // 目标类型：guide/trip
	TargetID   string    `json:"targetId"`                   // 目标ID
	ParentID   *string   `json:"parentId"`                   // 父评论ID（支持回复）
	Content    string    `gorm:"type:text" json:"content"`   // 评论内容
	LikeCount  int       `gorm:"default:0" json:"likeCount"` // 点赞数
	CreatedAt  time.Time `json:"createdAt"`
}

// BeforeCreate GORM 钩子
func (c *Comment) BeforeCreate() error {
	if c.ID == "" {
		c.ID = snowflake.GenerateID()
	}
	return nil
}
