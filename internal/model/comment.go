package model

import (
	"time"

	"travel-server/pkg/snowflake"

	"gorm.io/gorm"
)

// Comment 评论表
type Comment struct {
	ID         string    `gorm:"primaryKey" json:"id"`
	UserID     string    `gorm:"size:191" json:"userId"`     // 评论者
	User       User      `gorm:"foreignKey:UserID" json:"-"` // 关联用户（Preload 用）
	TargetType string    `gorm:"size:20" json:"targetType"`  // 目标类型：guide/trip
	TargetID   string    `gorm:"size:191" json:"targetId"`   // 目标ID
	ParentID   *string   `gorm:"size:191" json:"parentId"`   // 父评论ID（支持回复）
	Content    string    `gorm:"type:text" json:"content"`   // 评论内容
	LikeCount  int       `gorm:"default:0" json:"likeCount"` // 点赞数
	CreatedAt  time.Time `json:"createdAt"`
}

// BeforeCreate GORM 钩子
func (c *Comment) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = snowflake.GenerateID()
	}
	return nil
}
