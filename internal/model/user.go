// 用户模型
package model

import "time"

// User 用户表
type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	OpenID    string    `gorm:"uniqueIndex;size:64" json:"openid"` // 微信 openid
	UnionID   string    `gorm:"size:64" json:"unionid"`            // 微信 unionid
	Nickname  string    `gorm:"size:50" json:"nickname"`           // 昵称
	AvatarURL string    `gorm:"size:500" json:"avatarUrl"`        // 头像链接
	Role      int       `gorm:"default:0" json:"role"`             // 0普通 1领队 2管理员
	CreatedAt time.Time `json:"createdAt"`
}
