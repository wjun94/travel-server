package model

import (
	"time"

	"travel-server/pkg/snowflake"
)

// SectionType 每日行程项类型常量
const (
	SectionTransport  = "transport"  // 🚄 交通
	SectionHotel      = "hotel"      // 🏨 住宿
	SectionAttraction = "attraction" // 🏞️ 景点
	SectionFood       = "food"       // 🍜 美食
	SectionShopping   = "shopping"   // 🛍️ 购物
	SectionTips       = "tips"       // ⚠️ 避坑
	SectionCustom     = "custom"     // 📌 自定义
)

// ValidSectionTypes 合法的行程项类型集合
var ValidSectionTypes = map[string]bool{
	SectionTransport:  true,
	SectionHotel:      true,
	SectionAttraction: true,
	SectionFood:       true,
	SectionShopping:   true,
	SectionTips:       true,
	SectionCustom:     true,
}

// GuideSection 攻略每日行程表 — 代表攻略中的一天
// DayNumber 自动从1开始，无需手动排序
type GuideSection struct {
	ID        string     `gorm:"primaryKey" json:"id"`
	GuideID   string     `gorm:"size:191;index" json:"guideId"` // 所属攻略
	DayNumber int        `gorm:"not null" json:"dayNumber"`     // 第几天 (1,2,3...)
	Date      *time.Time `json:"date"`                          // 当天的日期（可选）
	Title     string     `gorm:"size:100" json:"title"`         // 标题（如"第一天：出发"）
	CreatedAt time.Time  `json:"createdAt"`

	Items []GuideDayItem `gorm:"foreignKey:DayID" json:"items,omitempty"` // 当天行程项列表
}

// BeforeCreate GORM 钩子
func (gs *GuideSection) BeforeCreate() error {
	if gs.ID == "" {
		gs.ID = snowflake.GenerateID()
	}
	return nil
}

// GuideDayItem 每日行程项表 — 某一天的具体活动项
type GuideDayItem struct {
	ID              string     `gorm:"primaryKey" json:"id"`
	DayID           string     `gorm:"size:191;index" json:"dayId"`  // 所属天的ID
	SectionType     string     `gorm:"size:30" json:"sectionType"`   // 板块类型
	Title           string     `gorm:"size:200" json:"title"`        // 活动标题
	Description     string     `gorm:"type:text" json:"description"` // 活动描述
	StartTime       *time.Time `json:"startTime"`                    // 开始时间
	EndTime         *time.Time `json:"endTime"`                      // 结束时间
	Latitude        *float64   `gorm:"type:decimal(10,7)" json:"latitude"`
	Longitude       *float64   `gorm:"type:decimal(10,7)" json:"longitude"`
	Address         string     `gorm:"size:255" json:"address"`
	Images          string     `gorm:"type:text" json:"images"`               // 图片URL列表（JSON数组，最多9张）
	NeedReservation bool       `gorm:"default:false" json:"needReservation"`  // 是否需要预约/购票
	TicketChannel   string     `gorm:"size:50" json:"ticketChannel"`          // 购票渠道：公众号/小程序/线下
	TicketPrice     *float64   `gorm:"type:decimal(10,2)" json:"ticketPrice"` // 票价，nil=未填写，0=免费，>0=付费
	CreatedAt       time.Time  `json:"createdAt"`
}

// BeforeCreate GORM 钩子
func (gdi *GuideDayItem) BeforeCreate() error {
	if gdi.ID == "" {
		gdi.ID = snowflake.GenerateID()
	}
	return nil
}
