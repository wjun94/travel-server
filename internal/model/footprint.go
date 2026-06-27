package model

import (
	"time"

	"travel-server/pkg/snowflake"
)

// Footprint 用户足迹（点亮城市）
type Footprint struct {
	ID        string    `gorm:"primaryKey" json:"id"`
	UserID    string    `gorm:"size:191" json:"userId"`
	City      string    `gorm:"size:50" json:"city"`
	Province  string    `gorm:"size:50" json:"province"`
	Lat       float64   `json:"lat"`       // 纬度
	Lng       float64   `json:"lng"`       // 经度
	VisitedAt time.Time `json:"visitedAt"` // 到访时间
}

// BeforeCreate GORM 钩子
func (f *Footprint) BeforeCreate() error {
	if f.ID == "" {
		f.ID = snowflake.GenerateID()
	}
	return nil
}
