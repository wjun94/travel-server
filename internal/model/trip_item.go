package model

import "time"

// TripItem 行程项表 — 行程日下的具体安排项，行程的最小执行单元
type TripItem struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	TripDayID    uint      `json:"tripDayId"`                           // 所属行程日
	SortOrder    int       `json:"sortOrder"`                           // 排序序号
	StartTime    string    `gorm:"type:time" json:"startTime"`          // 开始时间
	EndTime      string    `gorm:"type:time" json:"endTime"`            // 结束时间
	ItemType     string    `gorm:"size:30" json:"itemType"`             // 类型：transport/attraction/meal/hotel/free
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
