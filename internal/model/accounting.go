package model

import (
	"time"

	"travel-server/pkg/snowflake"
)

// Accounting 旅行记账
type Accounting struct {
	ID            string    `gorm:"primaryKey" json:"id"`
	TripID        string    `json:"tripId"`                        // 关联的行程 ID
	UserID        string    `json:"userId"`                        // 记账用户
	Category      string    `gorm:"size:50" json:"category"`       // 分类：交通/餐饮/住宿/其他
	Amount        float64   `json:"amount"`                        // 金额
	Note          string    `gorm:"size:200" json:"note"`          // 备注
	TransactionID string    `gorm:"size:100" json:"transactionId"` // 微信支付单号
	ConsumedAt    time.Time `json:"consumedAt"`                    // 消费时间
	CreatedAt     time.Time `json:"createdAt"`
}

// BeforeCreate GORM 钩子
func (a *Accounting) BeforeCreate() error {
	if a.ID == "" {
		a.ID = snowflake.GenerateID()
	}
	return nil
}
