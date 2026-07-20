package model

import (
	"time"

	"travel-server/pkg/snowflake"

	"gorm.io/gorm"
)

// Partner 搭子组队信息
type Partner struct {
	ID     string `gorm:"primaryKey;size:64" json:"id"`
	UserID string `gorm:"size:191;index" json:"userId"`           // 发起人ID
	TripID string `gorm:"size:64;index;default:''" json:"tripId"` // 绑定行程ID，无行程为空
	Type   int    `gorm:"default:0;comment:0用户搭子 1官方活动" json:"type"`
	// 基础展示
	Title       string  `gorm:"size:255" json:"title"`         // 搭子招募标题
	Cover       string  `gorm:"size:512" json:"cover"`         // 封面图
	CityCode    string  `gorm:"size:64;index" json:"cityCode"` // 城市编码，用于同城筛选
	CityName    string  `gorm:"size:64" json:"cityName"`       // 城市名称
	Destination string  `gorm:"size:255" json:"destination"`   // 目的地详细文本
	Longitude   float64 `json:"longitude"`                     // 经度 同城匹配
	Latitude    float64 `json:"latitude"`                      // 纬度
	// 出行时间
	StartDate *time.Time `gorm:"index" json:"startDate"` // 出发日期
	EndDate   *time.Time `json:"endDate"`                // 结束日期
	Days      int        `json:"days"`                   // 出行天数（冗余，方便展示）
	// 出行分类标签
	TravelTags string `gorm:"type:text" json:"travelTags"` // 逗号分隔标签：自驾,徒步,美食,亲子
	// 招募文案
	Desc        string `gorm:"type:text" json:"desc"`        // 行程简述、出行计划
	Requirement string `gorm:"type:text" json:"requirement"` // 人员要求：年龄、性别、预算等
	// 人员配置
	MaxMembers     int `json:"maxMembers"`                      // 最多招募人数
	CurrentMembers int `gorm:"default:0" json:"currentMembers"` // 当前已通过人数
	GenderLimit    int `gorm:"default:0;comment:0不限 1仅男生 2仅女生" json:"genderLimit"`
	MinAge         int `gorm:"default:0" json:"minAge"`  // 最小年龄限制
	MaxAge         int `gorm:"default:99" json:"maxAge"` // 最大年龄限制
	// 费用（拆分区分用户/官方）
	BudgetPerPerson int     `json:"budgetPerPerson"`                // 人均预算（用户搭子用）
	OfficialPrice   float64 `gorm:"default:0" json:"officialPrice"` // 官方活动定价，type=1生效
	// 状态体系扩充
	Status   int `gorm:"default:0;index;comment:0招募中 1满员 2取消 3已过期 4行程结束" json:"status"`
	IsPublic int `gorm:"default:1;comment:0仅自己可见 1公开招募" json:"isPublic"`
	// 运营、统计字段
	ViewCount  int `gorm:"default:0" json:"viewCount"`  // 浏览量
	SortWeight int `gorm:"default:0" json:"sortWeight"` // 运营权重，首页置顶
	// 时间与软删
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"` // 软删除
}

// BeforeCreate GORM 钩子
func (p *Partner) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = snowflake.GenerateID()
	}
	return nil
}

// PartnerApplication 搭子申请记录
type PartnerApplication struct {
	ID        string         `gorm:"primaryKey;size:64" json:"id"`
	PartnerID string         `gorm:"size:64;index" json:"partnerId"` // 关联搭子ID
	UserID    string         `gorm:"size:191;index" json:"userId"`   // 报名用户
	Status    int            `gorm:"default:0;comment:0待审核 1通过 2拒绝 3主动退出" json:"status"`
	Remark    string         `gorm:"size:512" json:"remark"` // 报名留言
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// BeforeCreate GORM 钩子
func (pa *PartnerApplication) BeforeCreate(tx *gorm.DB) error {
	if pa.ID == "" {
		pa.ID = snowflake.GenerateID()
	}
	return nil
}
