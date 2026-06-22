package model

import "time"

// Post 攻略帖子
type Post struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `json:"user_id"`                      // 作者 ID
	Content    string    `gorm:"type:text" json:"content"`     // 图文内容（图片为 JSON 数组）
	Location   string    `gorm:"size:200" json:"location"`     // 定位信息
	City       string    `gorm:"size:50" json:"city"`          // 城市（用于足迹）
	Tags       string    `gorm:"type:text" json:"tags"`        // 标签（JSON 数组）
	Status     int       `gorm:"default:0" json:"status"`      // 0审核中 1已发布 2下架
	LikeCount  int       `gorm:"default:0" json:"like_count"`  // 点赞数
	ShareCount int       `gorm:"default:0" json:"share_count"` // 分享数
	CreatedAt  time.Time `json:"created_at"`
}
