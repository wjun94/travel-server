package model

import "time"

// SectionType 板块类型常量
const (
	SectionTransport  = "transport"  // 🚄 交通指南
	SectionHotel      = "hotel"      // 🏨 住宿推荐
	SectionAttraction = "attraction" // 🏞️ 景点推荐
	SectionFood       = "food"       // 🍜 美食地图
	SectionShopping   = "shopping"   // 🛍️ 购物攻略
	SectionTips       = "tips"       // ⚠️ 避坑提醒
	SectionCustom     = "custom"     // 📌 自定义
)

// ValidSectionTypes 合法的板块类型集合
var ValidSectionTypes = map[string]bool{
	SectionTransport:  true,
	SectionHotel:      true,
	SectionAttraction: true,
	SectionFood:       true,
	SectionShopping:   true,
	SectionTips:       true,
	SectionCustom:     true,
}

// GuideSection 攻略板块表 — 攻略内的分类内容板块（交通、住宿、美食、景点、避坑等）
type GuideSection struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	GuideID     uint      `json:"guideId"`                    // 所属攻略
	SectionType string    `gorm:"size:30" json:"sectionType"` // 板块类型：overview/transport/hotel/food/attraction/itinerary/budget/tips/custom
	Title       string    `gorm:"size:100" json:"title"`      // 板块标题
	Content     string    `gorm:"type:text" json:"content"`   // 板块内容（富文本/Markdown）
	SortOrder   int       `json:"sortOrder"`                  // 排序序号
	CreatedAt   time.Time `json:"createdAt"`
}
