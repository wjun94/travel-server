package model

import (
	"time"

	"travel-server/pkg/snowflake"

	"gorm.io/gorm"
)

// PartnerDay 搭子行程日表 — 代表搭子行程安排中的一天（参考行程 TripDay 结构）
type PartnerDay struct {
	ID        string           `gorm:"primaryKey" json:"id"`
	PartnerID string           `gorm:"size:191;index" json:"partnerId"` // 所属搭子
	DayNumber int              `gorm:"not null" json:"dayNumber"`       // 第几天 (1,2,3...)
	Date      string           `gorm:"size:50" json:"date"`             // 具体日期（直接保存前端字符串）
	Title     string           `gorm:"size:100" json:"title"`           // 标题（如"第一天：出发"）
	CreatedAt time.Time        `json:"createdAt"`
	Items     []PartnerDayItem `gorm:"foreignKey:DayID" json:"items,omitempty"` // 当天行程项列表
}

// BeforeCreate GORM 钩子
func (pd *PartnerDay) BeforeCreate(tx *gorm.DB) error {
	if pd.ID == "" {
		pd.ID = snowflake.GenerateID()
	}
	return nil
}

// PartnerDayItem 搭子每日行程项表 — 某一天的具体活动项，与行程项字段一致
type PartnerDayItem struct {
	ID              string    `gorm:"primaryKey" json:"id"`
	DayID           string    `gorm:"size:191;index" json:"dayId"`  // 所属天的ID
	SectionType     string    `gorm:"size:30" json:"sectionType"`   // 板块类型
	Title           string    `gorm:"size:200" json:"title"`        // 活动标题
	Description     string    `gorm:"type:text" json:"description"` // 活动描述/备注
	StartTime       string    `gorm:"size:50" json:"startTime"`     // 开始时间（直接保存前端字符串）
	EndTime         string    `gorm:"size:50" json:"endTime"`       // 结束时间（直接保存前端字符串）
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
func (pdi *PartnerDayItem) BeforeCreate(tx *gorm.DB) error {
	if pdi.ID == "" {
		pdi.ID = snowflake.GenerateID()
	}
	return nil
}
