package model

import (
	"time"

	"travel-server/pkg/snowflake"
)

// Partner 搭子组队信息
type Partner struct {
	ID             string     `gorm:"primaryKey" json:"id"`
	UserID         string     `gorm:"size:191" json:"userId"` // 发起人 ID（官方搭子为空）
	Type           int        `gorm:"default:0" json:"type"`  // 0用户发起 1官方活动
	Destination    string     `json:"destination"`
	StartDate      *time.Time `json:"startDate"`
	Days           int        `json:"days"`
	Requirement    string     `gorm:"type:text" json:"requirement"` // 要求
	MaxMembers     int        `json:"maxMembers"`                   // 最大人数
	CurrentMembers int        `json:"currentMembers"`               // 当前人数
	Price          float64    `json:"price"`                        // 官方活动价格
	Status         int        `gorm:"default:0" json:"status"`      // 0招募中 1满员 2取消
	CreatedAt      time.Time  `json:"createdAt"`
}

// BeforeCreate GORM 钩子
func (p *Partner) BeforeCreate() error {
	if p.ID == "" {
		p.ID = snowflake.GenerateID()
	}
	return nil
}

// PartnerApplication 搭子申请记录
type PartnerApplication struct {
	ID          string    `gorm:"primaryKey" json:"id"`
	PartnerID   string    `gorm:"size:191" json:"partnerId"`
	ApplicantID string    `gorm:"size:191" json:"applicantId"` // 申请人 ID
	Message     string    `json:"message"`                     // 申请留言
	Status      int       `gorm:"default:0" json:"status"`     // 0待审核 1同意 2拒绝
	CreatedAt   time.Time `json:"createdAt"`
}

// BeforeCreate GORM 钩子
func (pa *PartnerApplication) BeforeCreate() error {
	if pa.ID == "" {
		pa.ID = snowflake.GenerateID()
	}
	return nil
}
