package model

import "time"

// Complaint 用户投诉
// 被投诉对象类型 target_type：user用户 guide攻略 trip行程 partner搭子 comment评论 other其他
type Complaint struct {
	ID         string     `gorm:"primaryKey" json:"id"`
	UserID     string     `gorm:"size:191;index" json:"userId"`    // 投诉人ID
	TargetType string     `gorm:"size:20;index" json:"targetType"` // 被投诉对象类型
	TargetID   string     `gorm:"size:191" json:"targetId"`        // 被投诉对象ID（other 时可为空）
	Reason     string     `gorm:"size:50" json:"reason"`           // 投诉原因
	Content    string     `gorm:"size:1000" json:"content"`        // 详细描述
	Images     string     `gorm:"size:2000" json:"images"`         // 图片URL（JSON数组字符串，最多9张）
	Status     int        `gorm:"default:0;index" json:"status"`   // 0待处理 1已处理 2已驳回
	HandleNote string     `gorm:"size:500" json:"handleNote"`      // 处理备注
	Reply      string     `gorm:"size:500" json:"reply"`           // 后台回复（小程序可见）
	HandledAt  *time.Time `json:"handledAt"`                       // 处理时间
	CreatedAt  time.Time  `json:"createdAt"`                       // 提交时间
}
