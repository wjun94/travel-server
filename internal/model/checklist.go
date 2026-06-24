package model

import "time"

// Checklist 备忘清单
type Checklist struct {
	ID         uint            `gorm:"primaryKey" json:"id"`
	UserID     uint            `json:"userId"`
	Name       string          `gorm:"size:100" json:"name"`                // 清单名称
	IsTemplate int             `gorm:"default:0" json:"isTemplate"`        // 0否 1官方模板
	TripID     uint            `json:"tripId"`                             // 关联行程
	Items      []ChecklistItem `gorm:"foreignKey:ChecklistID" json:"items"` // 清单项
	CreatedAt  time.Time       `json:"createdAt"`
}

// ChecklistItem 备忘清单条目
type ChecklistItem struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	ChecklistID uint   `json:"checklistId"`
	Text        string `gorm:"size:200" json:"text"`
	Checked     int    `gorm:"default:0" json:"checked"` // 0未勾选 1已勾选
}
