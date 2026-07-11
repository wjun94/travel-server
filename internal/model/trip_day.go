package model

import (
	"time"

	"travel-server/pkg/snowflake"

	"gorm.io/gorm"
)

// TripDay 行程日表 — 代表行程中的一天
type TripDay struct {
	ID        string    `gorm:"primaryKey" json:"id"`
	TripID    string    `gorm:"size:191;index" json:"tripId"` // 所属行程
	DayNumber int       `gorm:"not null" json:"dayNumber"`    // 第几天 (1,2,3...)
	Date      string    `gorm:"size:50" json:"date"`          // 具体日期（直接保存前端字符串）
	Title     string    `gorm:"size:100" json:"title"`        // 标题（如"第一天：出发"）
	CreatedAt time.Time `json:"createdAt"`

	Items []TripItem `gorm:"foreignKey:TripDayID" json:"items,omitempty"` // 当天行程项列表
}

// BeforeCreate GORM 钩子
func (td *TripDay) BeforeCreate(tx *gorm.DB) error {
	if td.ID == "" {
		td.ID = snowflake.GenerateID()
	}
	return nil
}
