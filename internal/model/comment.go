package model

import "time"

// Comment 评论表
type Comment struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `json:"userId"`                     // 评论者
	TargetType string    `gorm:"size:20" json:"targetType"`  // 目标类型：guide/trip
	TargetID   uint      `json:"targetId"`                   // 目标ID
	ParentID   *uint     `json:"parentId"`                   // 父评论ID（支持回复）
	Content    string    `gorm:"type:text" json:"content"`   // 评论内容
	LikeCount  int       `gorm:"default:0" json:"likeCount"` // 点赞数
	CreatedAt  time.Time `json:"createdAt"`
}
