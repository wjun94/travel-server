package model

import (
	"time"

	"gorm.io/gorm"
)

// Trip 行程表 — 执行计划型内容，按时间线组织
type Trip struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	UserID      uint           `json:"userId"`                                // 创建者
	GuideID     *uint          `json:"guideId"`                               // 关联攻略（可为空）
	Title       string         `gorm:"size:200" json:"title"`                 // 行程标题
	Destination string         `gorm:"size:100" json:"destination"`           // 目的地
	StartDate   string         `gorm:"type:date" json:"startDate"`            // 出发日期
	EndDate     string         `gorm:"type:date" json:"endDate"`              // 结束日期
	TotalBudget float64        `gorm:"type:decimal(10,2)" json:"totalBudget"` // 总预算
	Status      int            `gorm:"default:0" json:"status"`               // 状态：0计划中/1进行中/2已完成
	IsPublic    int            `gorm:"default:0" json:"isPublic"`             // 是否公开
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"` // 软删除

	Days    []TripDay    `gorm:"foreignKey:TripID" json:"days,omitempty"`    // 行程日列表
	Members []TripMember `gorm:"foreignKey:TripID" json:"members,omitempty"` // 同行者列表
}

// TripMember 同行者表
type TripMember struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	TripID    uint      `json:"tripId"`                             // 所属行程
	UserID    *uint     `json:"userId"`                             // 关联用户（可为空，支持非注册用户）
	Name      string    `gorm:"size:50" json:"name"`                // 姓名
	Role      string    `gorm:"size:20;default:viewer" json:"role"` // 角色：owner/editor/viewer
	CreatedAt time.Time `json:"createdAt"`
}
