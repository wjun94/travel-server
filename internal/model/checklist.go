package model

import (
	"time"

	"travel-server/pkg/snowflake"

	"gorm.io/gorm"
)

// Checklist 备忘清单
type Checklist struct {
	ID         string          `gorm:"primaryKey" json:"id"`
	UserID     string          `gorm:"size:191" json:"userId"`
	Name       string          `gorm:"size:100" json:"name"`                // 清单名称
	IsTemplate int             `gorm:"default:0" json:"isTemplate"`         // 0否 1官方模板
	TripID     string          `gorm:"size:191" json:"tripId"`              // 关联行程
	Items      []ChecklistItem `gorm:"foreignKey:ChecklistID" json:"items"` // 清单项
	CreatedAt  time.Time       `json:"createdAt"`
}

// BeforeCreate GORM 钩子
func (cl *Checklist) BeforeCreate(tx *gorm.DB) error {
	if cl.ID == "" {
		cl.ID = snowflake.GenerateID()
	}
	return nil
}

// ChecklistItem 备忘清单条目
type ChecklistItem struct {
	ID          string `gorm:"primaryKey" json:"id"`
	ChecklistID string `gorm:"size:191" json:"checklistId"`
	Text        string `gorm:"size:200" json:"text"`
	Checked     int    `gorm:"default:0" json:"checked"` // 0未勾选 1已勾选
}

// BeforeCreate GORM 钩子
func (ci *ChecklistItem) BeforeCreate(tx *gorm.DB) error {
	if ci.ID == "" {
		ci.ID = snowflake.GenerateID()
	}
	return nil
}
