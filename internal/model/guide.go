package model

import (
	"time"

	"gorm.io/gorm"
)

// Guide 攻略表 — 决策参考型内容，包含目的地综合信息
type Guide struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	UserID          uint           `json:"userId"`                        // 作者
	Title           string         `gorm:"size:200" json:"title"`         // 标题
	CoverImage      string         `gorm:"size:255" json:"coverImage"`    // 封面图
	Destination     string         `gorm:"size:100" json:"destination"`   // 目的地
	Summary         string         `gorm:"size:500" json:"summary"`       // 摘要
	BudgetMin       float64        `gorm:"type:decimal(10,2)" json:"budgetMin"` // 预算下限
	BudgetMax       float64        `gorm:"type:decimal(10,2)" json:"budgetMax"` // 预算上限
	BestSeason      string         `gorm:"size:50" json:"bestSeason"`     // 最佳季节
	RecommendedDays int            `json:"recommendedDays"`               // 建议天数
	ViewCount       int            `gorm:"default:0" json:"viewCount"`    // 浏览量
	LikeCount       int            `gorm:"default:0" json:"likeCount"`    // 点赞数
	Status          int            `gorm:"default:0" json:"status"`       // 状态：0草稿/1已发布/2下架
	CreatedAt       time.Time      `json:"createdAt"`
	UpdatedAt       time.Time      `json:"updatedAt"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`               // 软删除
}
