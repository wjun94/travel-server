package model

import "time"

// Recommendation 后台推荐内容（如 TOP 民宿）
type Recommendation struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Title     string    `json:"title"`
	Cover     string    `json:"cover"` // 封面图
	City      string    `gorm:"size:50" json:"city"`
	Type      string    `gorm:"size:50" json:"type"` // house / activity
	Link      string    `json:"link"`                // 跳转链接
	StartTime time.Time `json:"startTime"`          // 展示开始时间
	EndTime   time.Time `json:"endTime"`            // 展示结束时间
	CreatedAt time.Time `json:"createdAt"`
}
