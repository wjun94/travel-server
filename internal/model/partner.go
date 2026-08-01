package model

import (
	"time"

	"travel-server/pkg/snowflake"

	"gorm.io/gorm"
)

// Partner 搭子组队信息
type Partner struct {
	ID       string `gorm:"primaryKey;size:64" json:"id"`
	UserID   string `gorm:"size:191;index" json:"userId"`           // 发起人ID
	TripID   string `gorm:"size:64;index;default:''" json:"tripId"` // 关联行程ID
	Type     int    `gorm:"default:0" json:"type"`                  // 0用户搭子 1官方活动
	Category string `gorm:"size:50;default:''" json:"category"`     // 活动分类：旅游/美食/运动/学习/探店/看展/桌游
	// 基础展示
	Title        string  `gorm:"size:255" json:"title"`             // 招募标题
	Cover        string  `gorm:"size:512" json:"cover"`             // 封面图
	Images       string  `gorm:"type:text" json:"images"`           // 多图JSON数组
	Destination  string  `gorm:"size:255;index" json:"destination"` // 目的地
	Longitude    float64 `json:"longitude"`                         // 经度（同城匹配）
	Latitude     float64 `json:"latitude"`                          // 纬度
	Address      string  `gorm:"size:500" json:"address"`           // 详细地址
	LocationType int     `gorm:"default:0" json:"locationType"`     // 0线下 1线上
	OnlineLink   string  `gorm:"size:500" json:"onlineLink"`        // 线上链接
	// 出行时间
	StartDate *time.Time `gorm:"index" json:"startDate"` // 出发日期
	EndDate   *time.Time `json:"endDate"`                // 结束日期
	Days      int        `json:"days"`                   // 出行天数
	// 标签（多选）
	TravelTags string `gorm:"type:text" json:"travelTags"` // 逗号分隔，如：自驾,徒步,美食（兼容旧版）
	Tags       string `gorm:"type:text" json:"tags"`       // 多选标签JSON数组，如：["徒步","拍照","美食"]
	// 招募文案
	Desc        string `gorm:"type:text" json:"desc"`        // 行程简述
	RichDesc    string `gorm:"type:text" json:"richDesc"`    // 详细介绍（富文本）
	Requirement string `gorm:"type:text" json:"requirement"` // 人员要求
	// 人员配置
	MaxMembers     int `json:"maxMembers"`                      // 招募上限
	MinMembers     int `gorm:"default:0" json:"minMembers"`     // 最小成团人数
	CurrentMembers int `gorm:"default:0" json:"currentMembers"` // 已通过人数
	GenderLimit    int `gorm:"default:0" json:"genderLimit"`    // 0不限 1仅男生 2仅女生
	MaleCount      int `gorm:"default:0" json:"maleCount"`      // 男生需求数
	FemaleCount    int `gorm:"default:0" json:"femaleCount"`    // 女生需求数
	MinAge         int `gorm:"default:0" json:"minAge"`         // 年龄下限
	MaxAge         int `gorm:"default:99" json:"maxAge"`        // 年龄上限
	// 费用
	FeeMode         int     `gorm:"default:0" json:"feeMode"`       // 0免费 1AA 2组织者全包 3人均固定预算
	BudgetPerPerson int     `json:"budgetPerPerson"`                // 人均预算
	OfficialPrice   float64 `gorm:"default:0" json:"officialPrice"` // 官方活动定价
	FeeInclude      string  `gorm:"type:text" json:"feeInclude"`    // 费用包含
	FeeExclude      string  `gorm:"type:text" json:"feeExclude"`    // 费用不含
	EstTotal        int     `json:"estTotal"`                       // 预估总价
	// 隐私管控
	Visibility   int `gorm:"default:0" json:"visibility"`   // 0全部可见 1同城可见 2好友可见
	JoinMode     int `gorm:"default:0" json:"joinMode"`     // 0需审核 1直接加入
	AutoClose    int `gorm:"default:1" json:"autoClose"`    // 满员自动关闭：0否 1是
	AllowShare   int `gorm:"default:1" json:"allowShare"`   // 允许转发：0否 1是
	AllowCollect int `gorm:"default:1" json:"allowCollect"` // 允许收藏：0否 1是
	IsDraft      int `gorm:"default:0" json:"isDraft"`      // 0已发布 1草稿
	IsAI         int `gorm:"default:0" json:"isAI"`         // 是否AI生成：0手动 1AI
	// 状态
	Status   int `gorm:"default:0;index" json:"status"` // 0招募中 1满员 2取消 3已过期 4行程结束
	IsPublic int `gorm:"default:1" json:"isPublic"`     // 0仅自己可见 1公开招募
	// 运营统计
	ViewCount     int `gorm:"default:0" json:"viewCount"`     // 浏览量
	LikeCount     int `gorm:"default:0" json:"likeCount"`     // 点赞数
	FavoriteCount int `gorm:"default:0" json:"favoriteCount"` // 收藏数
	CommentCount  int `gorm:"default:0" json:"commentCount"`  // 评论数
	SortWeight    int `gorm:"default:0" json:"sortWeight"`    // 排序权重
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
