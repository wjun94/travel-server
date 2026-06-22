package model

import "time"

// Partner 搭子组队信息
type Partner struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	UserID         uint       `json:"user_id"`               // 发起人 ID（官方搭子为 0）
	Type           int        `gorm:"default:0" json:"type"` // 0用户发起 1官方活动
	Destination    string     `json:"destination"`
	StartDate      *time.Time `json:"start_date"`
	Days           int        `json:"days"`
	Requirement    string     `gorm:"type:text" json:"requirement"` // 要求
	MaxMembers     int        `json:"max_members"`                  // 最大人数
	CurrentMembers int        `json:"current_members"`              // 当前人数
	Price          float64    `json:"price"`                        // 官方活动价格
	Status         int        `gorm:"default:0" json:"status"`      // 0招募中 1满员 2取消
	CreatedAt      time.Time  `json:"created_at"`
}

// PartnerApplication 搭子申请记录
type PartnerApplication struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	PartnerID   uint      `json:"partner_id"`
	ApplicantID uint      `json:"applicant_id"`            // 申请人 ID
	Message     string    `json:"message"`                 // 申请留言
	Status      int       `gorm:"default:0" json:"status"` // 0待审核 1同意 2拒绝
	CreatedAt   time.Time `json:"created_at"`
}
