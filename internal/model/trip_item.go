package model

import (
	"time"

	"travel-server/pkg/snowflake"

	"gorm.io/gorm"
)

// 行程项类型常量
//
//	类型值       中文名   图标   说明                   典型场景
//	transport   交通    🚄    城市间/城市内的交通移动   高铁、飞机、打车、租车、地铁、轮渡
//	attraction  景点    🏞️    游览参观类活动           景区、博物馆、公园、古镇、演出
//	meal        餐饮    🍜    用餐类活动              早餐、午餐、晚餐、下午茶、小吃探店
//	hotel       住宿    🏨    过夜休息类活动           酒店入住、民宿、青旅、露营
//	free        自由活动 🎒    无固定安排的自由时间      逛街购物、休闲散步、自由探索、泡温泉
const (
	ItemTransport  = "transport"  // 🚄 交通
	ItemAttraction = "attraction" // 🏞️ 景点
	ItemMeal       = "meal"       // 🍜 餐饮
	ItemHotel      = "hotel"      // 🏨 住宿
	ItemFree       = "free"       // 🎒 自由活动
)

// ValidItemTypes 合法的行程项类型集合
var ValidItemTypes = map[string]bool{
	ItemTransport:  true,
	ItemAttraction: true,
	ItemMeal:       true,
	ItemHotel:      true,
	ItemFree:       true,
}

// TripItem 行程项表 — 行程日下的具体安排项，行程的最小执行单元
type TripItem struct {
	ID           string    `gorm:"primaryKey" json:"id"`
	TripDayID    string    `gorm:"size:191" json:"tripDayId"`           // 所属行程日
	SortOrder    int       `json:"sortOrder"`                           // 排序序号
	StartTime    string    `gorm:"type:time" json:"startTime"`          // 开始时间
	EndTime      string    `gorm:"type:time" json:"endTime"`            // 结束时间
	ItemType     string    `gorm:"size:30" json:"itemType"`             // 行程项类型（见上方常量）
	Title        string    `gorm:"size:200" json:"title"`               // 标题
	Description  string    `gorm:"type:text" json:"description"`        // 详细描述
	LocationName string    `gorm:"size:200" json:"locationName"`        // 地点名称
	Latitude     float64   `gorm:"type:decimal(10,7)" json:"latitude"`  // 纬度
	Longitude    float64   `gorm:"type:decimal(10,7)" json:"longitude"` // 经度
	Cost         float64   `gorm:"type:decimal(10,2)" json:"cost"`      // 预估花费
	BookingRef   string    `gorm:"size:200" json:"bookingRef"`          // 预订单号（机票/酒店等）
	Status       int       `gorm:"default:0" json:"status"`             // 状态：0待确认/1已确认/2已完成
	CreatedAt    time.Time `json:"createdAt"`
}

// BeforeCreate GORM 钩子
func (ti *TripItem) BeforeCreate(tx *gorm.DB) error {
	if ti.ID == "" {
		ti.ID = snowflake.GenerateID()
	}
	return nil
}
