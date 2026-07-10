// 用户模型
package model

import (
	"time"

	"travel-server/pkg/snowflake"

	"gorm.io/gorm"
)

// User 用户表
type User struct {
	ID            string    `gorm:"primaryKey" json:"id"`
	OpenID        string    `gorm:"uniqueIndex;size:64" json:"openid"` // 微信 openid
	UnionID       string    `gorm:"size:64" json:"unionid"`            // 微信 unionid
	Nickname      string    `gorm:"size:50" json:"nickname"`           // 昵称
	AvatarURL     string    `gorm:"size:500" json:"avatarUrl"`         // 头像链接
	Role          int       `gorm:"default:0" json:"role"`             // 0普通 1领队 2管理员
	FollowCount   int       `gorm:"default:0" json:"followCount"`      // 关注数
	FollowerCount int       `gorm:"default:0" json:"followerCount"`    // 粉丝数
	CreatedAt     time.Time `json:"createdAt"`
}

// BeforeCreate GORM 钩子：创建前自动生成 8 位雪花短 ID
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = snowflake.GenerateShortID(8)
	}
	return nil
}
