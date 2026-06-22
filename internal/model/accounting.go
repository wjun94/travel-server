package model

import "time"

// Accounting 旅行记账
type Accounting struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	TripID        uint      `json:"trip_id"`                        // 关联的行程 ID
	UserID        uint      `json:"user_id"`                        // 记账用户
	Category      string    `gorm:"size:50" json:"category"`        // 分类：交通/餐饮/住宿/其他
	Amount        float64   `json:"amount"`                         // 金额
	Note          string    `gorm:"size:200" json:"note"`           // 备注
	TransactionID string    `gorm:"size:100" json:"transaction_id"` // 微信支付单号
	ConsumedAt    time.Time `json:"consumed_at"`                    // 消费时间
	CreatedAt     time.Time `json:"created_at"`
}
