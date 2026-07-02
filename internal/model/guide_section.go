package model

import (
	"time"

	"travel-server/pkg/snowflake"

	"gorm.io/gorm"
)

// SectionType 每日行程项类型常量
const (
	SectionTransport  = "transport"  // 🚄 交通
	SectionHotel      = "hotel"      // 🏨 住宿
	SectionAttraction = "attraction" // 🏞️ 打卡地
	SectionFood       = "food"       // 🍜 美食
	SectionShopping   = "shopping"   // 🛍️ 购物
	SectionTips       = "tips"       // ⚠️ 避坑
)

// ValidSectionTypes 合法的行程项类型集合
var ValidSectionTypes = map[string]bool{
	SectionTransport:  true,
	SectionHotel:      true,
	SectionAttraction: true,
	SectionFood:       true,
	SectionShopping:   true,
	SectionTips:       true,
}

// TransportMode 交通方式常量
const (
	TransportTrain   = "train"   // 火车
	TransportBus     = "bus"     // 汽车
	TransportSubway  = "subway"  // 地铁
	TransportPlane   = "plane"   // 飞机
	TransportShip    = "ship"    // 轮船
	TransportCityBus = "citybus" // 公交
	TransportWalk    = "walk"    // 步行
	TransportBike    = "bike"    // 骑车
)

// ValidTransportModes 合法的交通方式集合
var ValidTransportModes = map[string]bool{
	TransportTrain:   true,
	TransportBus:     true,
	TransportSubway:  true,
	TransportPlane:   true,
	TransportShip:    true,
	TransportCityBus: true,
	TransportWalk:    true,
	TransportBike:    true,
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
func (gs *GuideSection) BeforeCreate(tx *gorm.DB) error {
	if gs.ID == "" {
		gs.ID = snowflake.GenerateID()
	}
	return nil
}

// GuideDayItem 每日行程项表 — 某一天的具体活动项
type GuideDayItem struct {
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
	Images          string    `gorm:"type:text" json:"images"`               // 图片URL列表（JSON数组，最多9张）
	NeedReservation bool      `gorm:"default:false" json:"needReservation"`  // 是否需要预约/购票
	TicketChannel   string    `gorm:"size:50" json:"ticketChannel"`          // 购票渠道：公众号/小程序/线下
	TicketPrice     *float64  `gorm:"type:decimal(10,2)" json:"ticketPrice"` // 票价，nil=未填写，0=免费，>0=付费
	TransportMode   string    `gorm:"size:30" json:"transportMode"`          // 交通方式（仅transport类型使用）
	StartPoint      string    `gorm:"size:255" json:"startPoint"`            // 起点名称（仅transport类型使用）
	EndPoint        string    `gorm:"size:255" json:"endPoint"`              // 终点名称（仅transport类型使用）
	StartLat        *float64  `gorm:"type:decimal(10,7)" json:"startLat"`    // 起点纬度
	StartLng        *float64  `gorm:"type:decimal(10,7)" json:"startLng"`    // 起点经度
	EndLat          *float64  `gorm:"type:decimal(10,7)" json:"endLat"`      // 终点纬度
	EndLng          *float64  `gorm:"type:decimal(10,7)" json:"endLng"`      // 终点经度
	CreatedAt       time.Time `json:"createdAt"`
}

// BeforeCreate GORM 钩子
func (gdi *GuideDayItem) BeforeCreate(tx *gorm.DB) error {
	if gdi.ID == "" {
		gdi.ID = snowflake.GenerateID()
	}
	return nil
}
