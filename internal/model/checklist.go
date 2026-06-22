package model

import "time"

// Checklist 备忘清单
type Checklist struct {
	ID         uint            `gorm:"primaryKey" json:"id"`
	UserID     uint            `json:"user_id"`
	Name       string          `gorm:"size:100" json:"name"`                // 清单名称
	IsTemplate int             `gorm:"default:0" json:"is_template"`        // 0否 1官方模板
	TripID     uint            `json:"trip_id"`                             // 关联行程
	Items      []ChecklistItem `gorm:"foreignKey:ChecklistID" json:"items"` // 清单项
	CreatedAt  time.Time       `json:"created_at"`
}

// ChecklistItem 备忘清单条目
type ChecklistItem struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	ChecklistID uint   `json:"checklist_id"`
	Text        string `gorm:"size:200" json:"text"`
	Checked     int    `gorm:"default:0" json:"checked"` // 0未勾选 1已勾选
}
