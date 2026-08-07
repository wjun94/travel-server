package model

import (
	"time"
	"travel-server/pkg/snowflake"

	"gorm.io/gorm"
)

// Guide 攻略表 — 每日行程型内容，包含目的地综合信息与每日行程安排
type Guide struct {
	ID              string         `gorm:"primaryKey" json:"id"`
	UserID          string         `gorm:"size:191" json:"userId"`              // 作者
	Title           string         `gorm:"size:200" json:"title"`               // 标题
	CoverImage      string         `gorm:"size:255" json:"coverImage"`          // 封面图
	Destination     string         `gorm:"size:100" json:"destination"`         // 目的地
	Summary         string         `gorm:"size:500" json:"summary"`             // 摘要
	BudgetMin       *float64       `gorm:"type:decimal(10,2)" json:"budgetMin"` // 预算下限
	BudgetMax       *float64       `gorm:"type:decimal(10,2)" json:"budgetMax"` // 预算上限
	BestSeason      string         `gorm:"size:50" json:"bestSeason"`           // 最佳季节
	RecommendedDays *int           `json:"recommendedDays"`                     // 建议天数
	Tags            string         `gorm:"size:500" json:"tags"`                // 标签（JSON数组，如 ["亲子游","穷游"]）
	Difficulty      string         `gorm:"size:20" json:"difficulty"`           // 难度：轻松/适中/挑战
	CrowdType       string         `gorm:"size:100" json:"crowdType"`           // 适合人群：情侣/家庭/独行/朋友
	IsOriginal      int            `gorm:"default:1" json:"isOriginal"`         // 是否原创
	IsOverseas      int            `gorm:"default:0;index" json:"isOverseas"`   // 境内境外：0国内 1境外
	ViewCount       int            `gorm:"default:0" json:"viewCount"`          // 浏览量
	LikeCount       int            `gorm:"default:0" json:"likeCount"`          // 点赞数
	Status          int            `gorm:"default:0" json:"status"`             // 状态：0草稿/1已发布/2下架
	CreatedAt       time.Time      `json:"createdAt"`
	UpdatedAt       time.Time      `json:"updatedAt"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"` // 软删除
}

// BeforeCreate GORM 钩子
func (g *Guide) BeforeCreate(tx *gorm.DB) error {
	if g.ID == "" {
		g.ID = snowflake.GenerateID()
	}
	return nil
}
