// 系统消息模型（管理后台发送的消息记录）
package model

import (
	"time"

	"travel-server/pkg/snowflake"

	"gorm.io/gorm"
)

// SysMessage 系统消息记录表：管理后台创建，按接收人群生成 Notification（type=4）下发
type SysMessage struct {
	ID            string     `gorm:"primaryKey" json:"id"`
	Title         string     `gorm:"size:100" json:"title"`          // 消息标题
	Content       string     `gorm:"type:text" json:"content"`       // 消息内容
	LinkURL       string     `gorm:"size:500" json:"linkUrl"`        // 跳转链接（可空）
	TargetType    string     `gorm:"size:20" json:"targetType"`      // 接收人群：all全部 users指定用户 group用户分组
	TargetUserIDs string     `gorm:"type:text" json:"targetUserIds"` // 指定用户ID列表（JSON数组，targetType=users 时）
	TargetGroup   string     `gorm:"size:20" json:"targetGroup"`     // 用户分组：normal普通 leader领队 admin管理员（targetType=group 时）
	Status        int        `gorm:"default:0" json:"status"`        // 发送状态：0待发送 1已发送 2已取消
	SendTime      time.Time  `json:"sendTime"`                       // 计划发送时间（立即发送=创建时间）
	SentAt        *time.Time `json:"sentAt"`                         // 实际发送时间
	SentCount     int        `gorm:"default:0" json:"sentCount"`     // 实际送达人数
	OperatorID    string     `gorm:"size:191" json:"operatorId"`     // 操作管理员ID
	CreatedAt     time.Time  `json:"createdAt"`
}

// BeforeCreate GORM 钩子：创建前自动生成雪花短 ID
func (s *SysMessage) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = snowflake.GenerateShortID(8)
	}
	return nil
}
