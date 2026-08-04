package model

import (
	"time"

	"travel-server/pkg/snowflake"

	"gorm.io/gorm"
)

// Accounting 旅行记账（可绑定行程/攻略/搭子）
type Accounting struct {
	ID            string    `gorm:"primaryKey" json:"id"`
	TargetType    string    `gorm:"size:20;index" json:"targetType"` // 绑定类型：trip行程 guide攻略 partner搭子
	TargetID      string    `gorm:"size:191;index" json:"targetId"`  // 绑定目标ID
	UserID        string    `gorm:"size:191" json:"userId"`          // 记账用户
	Category      string    `gorm:"size:50" json:"category"`         // 分类：交通/餐饮/住宿/门票/购物/其他
	Amount        float64   `json:"amount"`                          // 金额
	Note          string    `gorm:"size:200" json:"note"`            // 备注
	TransactionID string    `gorm:"size:100" json:"transactionId"`   // 微信支付单号
	ConsumedAt    time.Time `json:"consumedAt"`                      // 消费时间
	CreatedAt     time.Time `json:"createdAt"`
}

// BeforeCreate GORM 钩子
func (a *Accounting) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = snowflake.GenerateID()
	}
	return nil
}
