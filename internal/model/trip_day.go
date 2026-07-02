package model

import (
	"time"

	"travel-server/pkg/snowflake"

	"gorm.io/gorm"
)

// TripDay 行程日表
type TripDay struct {
	ID        string    `gorm:"primaryKey" json:"id"`
	TripID    string    `gorm:"size:191" json:"tripId"` // 所属行程
	DayNumber int       `json:"dayNumber"`              // 第几天（1, 2, 3...）
	Date      string    `gorm:"type:date" json:"date"`  // 具体日期
	Note      string    `gorm:"type:text" json:"note"`  // 当天备注
	CreatedAt time.Time `json:"createdAt"`

	Items []TripItem `gorm:"foreignKey:TripDayID" json:"items,omitempty"` // 行程项列表
}

// BeforeCreate GORM 钩子
func (td *TripDay) BeforeCreate(tx *gorm.DB) error {
	if td.ID == "" {
		td.ID = snowflake.GenerateID()
	}
	return nil
}
