package model

import "time"

// SectionType 板块类型常量
const (
	SectionOverview   = "overview"   // 📋 概览/总览
	SectionTransport  = "transport"  // ✈️ 交通指南
	SectionHotel      = "hotel"      // 🏨 住宿推荐
	SectionFood       = "food"       // 🍜 美食推荐
	SectionAttraction = "attraction" // 🏔️ 景点攻略
	SectionItinerary  = "itinerary"  // 📅 日程安排
	SectionBudget     = "budget"     // 💰 花费明细
	SectionTips       = "tips"       // ⚠️ 避坑/注意事项
	SectionCustom     = "custom"     // 📝 自定义板块
)

// GuideSection 攻略板块表 — 攻略内的分类内容板块（交通、住宿、美食、景点、避坑等）
type GuideSection struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	GuideID     uint      `json:"guideId"`                        // 所属攻略
	SectionType string    `gorm:"size:30" json:"sectionType"`     // 板块类型：overview/transport/hotel/food/attraction/itinerary/budget/tips/custom
	Title       string    `gorm:"size:100" json:"title"`          // 板块标题
	Content     string    `gorm:"type:text" json:"content"`       // 板块内容（富文本/Markdown）
	SortOrder   int       `json:"sortOrder"`                      // 排序序号
	CreatedAt   time.Time `json:"createdAt"`
}
