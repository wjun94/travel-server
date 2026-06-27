package model

import "time"

// TripDay 行程日表
type TripDay struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	TripID    uint      `json:"tripId"`                // 所属行程
	DayNumber int       `json:"dayNumber"`             // 第几天（1, 2, 3...）
	Date      string    `gorm:"type:date" json:"date"` // 具体日期
	Note      string    `gorm:"type:text" json:"note"` // 当天备注
	CreatedAt time.Time `json:"createdAt"`

	Items []TripItem `gorm:"foreignKey:TripDayID" json:"items,omitempty"` // 行程项列表
}
