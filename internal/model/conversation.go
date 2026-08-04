package model

import (
	"time"

	"gorm.io/gorm"

	"travel-server/pkg/snowflake"
)

// Conversation 群聊会话（搭子群聊，一个搭子对应一个群聊）
type Conversation struct {
	ID        string         `gorm:"primaryKey;size:64" json:"id"`
	PartnerID string         `gorm:"size:64;uniqueIndex" json:"partnerId"` // 关联搭子ID
	Name      string         `gorm:"size:255" json:"name"`                 // 群聊名称（默认取搭子标题）
	OwnerID   string         `gorm:"size:191;index" json:"ownerId"`        // 群主（搭子创建者）
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// BeforeCreate GORM 钩子
func (c *Conversation) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = snowflake.GenerateID()
	}
	return nil
}

// ConversationMember 群聊成员（被踢/退出时软删）
type ConversationMember struct {
	ID             string         `gorm:"primaryKey;size:64" json:"id"`
	ConversationID string         `gorm:"size:64;index" json:"conversationId"` // 群聊ID
	UserID         string         `gorm:"size:191;index" json:"userId"`        // 成员用户ID
	JoinedAt       time.Time      `json:"joinedAt"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

// BeforeCreate GORM 钩子
func (cm *ConversationMember) BeforeCreate(tx *gorm.DB) error {
	if cm.ID == "" {
		cm.ID = snowflake.GenerateID()
	}
	return nil
}

// ConversationMessage 群聊消息（独立于私聊 Message 表）
type ConversationMessage struct {
	ID             string    `gorm:"primaryKey;size:64" json:"id"`
	ConversationID string    `gorm:"size:64;index" json:"conversationId"` // 群聊ID
	FromUserID     string    `gorm:"size:191" json:"fromUserId"`          // 发送者ID
	Content        string    `gorm:"type:text" json:"content"`            // 消息内容
	CreatedAt      time.Time `json:"createdAt"`
}

// BeforeCreate GORM 钩子
func (cm *ConversationMessage) BeforeCreate(tx *gorm.DB) error {
	if cm.ID == "" {
		cm.ID = snowflake.GenerateID()
	}
	return nil
}
