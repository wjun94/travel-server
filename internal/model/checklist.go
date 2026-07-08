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

// ==================== 系统预置分类 ====================

// ChecklistCategory 系统预置的备忘清单分类（如"证件&手续类"）
type ChecklistCategory struct {
	ID        string                  `gorm:"primaryKey" json:"id"`
	Name      string                  `gorm:"size:50" json:"name"`   // 分类名称
	Type      int                     `gorm:"default:0" json:"type"` // 0基础分类 1场景分类
	SortOrder int                     `gorm:"default:0" json:"sortOrder"`
	Items     []ChecklistCategoryItem `gorm:"foreignKey:CategoryID" json:"items,omitempty"`
	CreatedAt time.Time               `json:"createdAt"`
}

// ChecklistCategoryItem 系统预置分类下的条目
type ChecklistCategoryItem struct {
	ID         string `gorm:"primaryKey" json:"id"`
	CategoryID string `gorm:"size:191;index" json:"categoryId"`
	Text       string `gorm:"size:200" json:"text"`
}

// BeforeCreate GORM 钩子
func (cc *ChecklistCategory) BeforeCreate(tx *gorm.DB) error {
	if cc.ID == "" {
		cc.ID = snowflake.GenerateID()
	}
	return nil
}

// BeforeCreate GORM 钩子
func (cci *ChecklistCategoryItem) BeforeCreate(tx *gorm.DB) error {
	if cci.ID == "" {
		cci.ID = snowflake.GenerateID()
	}
	return nil
}
