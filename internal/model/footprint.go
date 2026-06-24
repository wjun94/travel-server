package model

import "time"

// Footprint 用户足迹（点亮城市）
type Footprint struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `json:"userId"`
	City      string    `gorm:"size:50" json:"city"`
	Province  string    `gorm:"size:50" json:"province"`
	Lat       float64   `json:"lat"`        // 纬度
	Lng       float64   `json:"lng"`        // 经度
	VisitedAt time.Time `json:"visitedAt"` // 到访时间
}
