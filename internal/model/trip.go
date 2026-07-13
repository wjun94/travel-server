package model

import (
	"time"
	"travel-server/pkg/snowflake"

	"gorm.io/gorm"
)

// Trip 行程表 — 执行计划型内容，按时间线组织
type Trip struct {
	ID            string         `gorm:"primaryKey" json:"id"`
	UserID        string         `gorm:"size:191" json:"userId"`                        // 创建者
	GuideID       *string        `json:"guideId"`                                       // 关联攻略（可为空）
	Title         string         `gorm:"size:200" json:"title"`                         // 行程标题
	CoverImage    string         `gorm:"size:500" json:"coverImage"`                    // 封面图
	Countries     []string       `gorm:"type:text;serializer:json" json:"countries"`    // 国家列表
	Provinces     []string       `gorm:"type:text;serializer:json" json:"provinces"`    // 省份列表
	Cities        []string       `gorm:"type:text;serializer:json" json:"cities"`       // 目的地城市列表
	Destinations  []string       `gorm:"type:text;serializer:json" json:"destinations"` // 完整目的地列表（冗余字段，便于搜索）
	TotalBudget   float64        `gorm:"type:decimal(10,2)" json:"totalBudget"`         // 总预算
	IsOverseas    int            `gorm:"default:0" json:"isOverseas"`                   // 境内境外：0国内 1境外
	Summary       string         `gorm:"type:text" json:"summary"`                      // 行程备注
	ViewCount     int            `gorm:"default:0" json:"viewCount"`                    // 浏览量
	LikeCount     int            `gorm:"default:0" json:"likeCount"`                    // 点赞数
	FavoriteCount int            `gorm:"default:0" json:"favoriteCount"`                // 收藏数
	Status        int            `gorm:"default:1" json:"status"`                       // 状态：1草稿 2已发布 3已归档完结
	IsPublic      int            `gorm:"default:0" json:"isPublic"`                     // 是否公开（0私密 1公开）
	CreatedAt     time.Time      `json:"createdAt"`
	UpdatedAt     time.Time      `json:"updatedAt"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"` // 软删除

	Days    []TripDay    `gorm:"foreignKey:TripID" json:"days,omitempty"`    // 行程日列表
	Members []TripMember `gorm:"foreignKey:TripID" json:"members,omitempty"` // 同行者列表
}

// BeforeCreate GORM 钩子
func (t *Trip) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = snowflake.GenerateID()
	}
	return nil
}

// TripMember 同行者表
type TripMember struct {
	ID        string    `gorm:"primaryKey" json:"id"`
	TripID    string    `gorm:"size:191" json:"tripId"`             // 所属行程
	UserID    *string   `gorm:"size:191" json:"userId"`             // 关联用户（可为空，支持非注册用户）
	Name      string    `gorm:"size:50" json:"name"`                // 姓名
	Role      string    `gorm:"size:20;default:viewer" json:"role"` // 角色：owner/editor/viewer
	CreatedAt time.Time `json:"createdAt"`
}

// BeforeCreate GORM 钩子
func (tm *TripMember) BeforeCreate(tx *gorm.DB) error {
	if tm.ID == "" {
		tm.ID = snowflake.GenerateID()
	}
	return nil
}
