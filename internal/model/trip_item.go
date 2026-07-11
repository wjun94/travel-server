package model

import (
	"time"

	"travel-server/pkg/snowflake"

	"gorm.io/gorm"
)

// TripItem 每日行程项表 — 某一天的具体活动项，与攻略行程项字段一致
// 类型常量参见 guide_section.go 的 SectionType / ValidSectionTypes
type TripItem struct {
	ID              string    `gorm:"primaryKey" json:"id"`
	TripDayID       string    `gorm:"size:191;index" json:"tripDayId"` // 所属天的ID
	SectionType     string    `gorm:"size:30" json:"sectionType"`      // 板块类型
	Title           string    `gorm:"size:200" json:"title"`           // 活动标题
	Description     string    `gorm:"type:text" json:"description"`    // 活动描述/备注
	StartTime       string    `gorm:"size:50" json:"startTime"`        // 开始时间（直接保存前端字符串）
	EndTime         string    `gorm:"size:50" json:"endTime"`          // 结束时间（直接保存前端字符串）
	Latitude        *float64  `gorm:"type:decimal(10,7)" json:"latitude"`
	Longitude       *float64  `gorm:"type:decimal(10,7)" json:"longitude"`
	Address         string    `gorm:"size:255" json:"address"`
	Images          []string  `gorm:"type:text;serializer:json" json:"images"` // 图片URL列表（JSON数组，最多9张）
	NeedReservation bool      `gorm:"default:false" json:"needReservation"`    // 是否需要预约/购票
	TicketChannel   string    `gorm:"size:50" json:"ticketChannel"`            // 购票渠道：公众号/小程序/线下
	TicketPrice     *float64  `gorm:"type:decimal(10,2)" json:"ticketPrice"`   // 票价，nil=未填写，0=免费，>0=付费
	TransportMode   string    `gorm:"size:30" json:"transportMode"`            // 交通方式（仅transport类型使用）
	StartPoint      string    `gorm:"size:255" json:"startPoint"`              // 起点名称（仅transport类型使用）
	EndPoint        string    `gorm:"size:255" json:"endPoint"`                // 终点名称（仅transport类型使用）
	StartLat        *float64  `gorm:"type:decimal(10,7)" json:"startLat"`      // 起点纬度
	StartLng        *float64  `gorm:"type:decimal(10,7)" json:"startLng"`      // 起点经度
	EndLat          *float64  `gorm:"type:decimal(10,7)" json:"endLat"`        // 终点纬度
	EndLng          *float64  `gorm:"type:decimal(10,7)" json:"endLng"`        // 终点经度
	CreatedAt       time.Time `json:"createdAt"`
}

// BeforeCreate GORM 钩子
func (ti *TripItem) BeforeCreate(tx *gorm.DB) error {
	if ti.ID == "" {
		ti.ID = snowflake.GenerateID()
	}
	return nil
}
