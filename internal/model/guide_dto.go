package model

import (
	"fmt"
	"strings"
	"time"
)

// ==================== 创建攻略 ====================

// CreateGuideReq 创建攻略请求 — 包含攻略基本信息，可选每日行程
// 不传 Days 时自动创建第1天空天
type CreateGuideReq struct {
	Title           string   `json:"title" binding:"required"`
	CoverImage      string   `json:"coverImage" binding:"required"`
	Destination     string   `json:"destination" binding:"required"`
	Summary         string   `json:"summary"`
	BudgetMin       *float64 `json:"budgetMin"`
	BudgetMax       *float64 `json:"budgetMax"`
	BestSeason      string   `json:"bestSeason"`
	RecommendedDays *int     `json:"recommendedDays"`
	Tags            string   `json:"tags"`
	Difficulty      string   `json:"difficulty"`
	CrowdType       string   `json:"crowdType"`
	IsOriginal      int      `json:"isOriginal"`
	Status          int      `json:"status"` // 0草稿 / 1已发布
	Days            []DayReq `json:"days"`   // 可选的每日行程（不传则自动创建第1天）
}

// DayReq 创建每日行程请求
type DayReq struct {
	Date  *time.Time   `json:"date"`  // 当天的日期（可选）
	Title string       `json:"title"` // 标题（如"第一天：出发"）
	Items []DayItemReq `json:"items"` // 当天的行程项
}

// DayItemReq 创建行程项请求
type DayItemReq struct {
	SectionType     string   `json:"sectionType" binding:"required"`
	Title           string   `json:"title" binding:"required"`
	Description     string   `json:"description"` // 活动描述/备注
	StartTime       string   `json:"startTime"`   // 开始时间（直接保存前端字符串）
	EndTime         string   `json:"endTime"`     // 结束时间（直接保存前端字符串）
	Latitude        *float64 `json:"latitude"`
	Longitude       *float64 `json:"longitude"`
	Address         string   `json:"address"`
	Images          []string `json:"images"`          // 图片URL数组，最多9张
	NeedReservation bool     `json:"needReservation"` // 是否需要预约/购票
	TicketChannel   string   `json:"ticketChannel"`   // 购票渠道：公众号/小程序/线下
	TicketPrice     *float64 `json:"ticketPrice"`     // 票价，nil=未填写，0=免费，>0=付费
	TransportMode   string   `json:"transportMode"`   // 交通方式（仅transport类型使用）
	StartPoint      string   `json:"startPoint"`      // 起点名称（仅transport类型使用）
	EndPoint        string   `json:"endPoint"`        // 终点名称（仅transport类型使用）
	StartLat        *float64 `json:"startLat"`        // 起点纬度
	StartLng        *float64 `json:"startLng"`        // 起点经度
	EndLat          *float64 `json:"endLat"`          // 终点纬度
	EndLng          *float64 `json:"endLng"`          // 终点经度
}

// ==================== 更新攻略 ====================

// UpdateGuideReq 更新攻略请求（所有字段可选）
type UpdateGuideReq struct {
	Title           *string  `json:"title"`
	CoverImage      *string  `json:"coverImage"`
	Destination     *string  `json:"destination"`
	Summary         *string  `json:"summary"`
	BudgetMin       *float64 `json:"budgetMin"`
	BudgetMax       *float64 `json:"budgetMax"`
	BestSeason      *string  `json:"bestSeason"`
	RecommendedDays *int     `json:"recommendedDays"`
	Tags            *string  `json:"tags"`
	Difficulty      *string  `json:"difficulty"`
	CrowdType       *string  `json:"crowdType"`
	IsOriginal      *int     `json:"isOriginal"`
	Status          *int     `json:"status"`
}

// ==================== 校验逻辑 ====================

// ValidateCreateGuideReq 校验创建攻略请求，返回 nil 表示通过
func ValidateCreateGuideReq(req *CreateGuideReq) *ValidationError {
	// title: 不能为空，长度 5~200
	if len(strings.TrimSpace(req.Title)) < 5 || len(req.Title) > 200 {
		return &ValidationError{Field: "title", Msg: "标题长度需在5~200字符之间"}
	}
	// cover_image: 不能为空
	if strings.TrimSpace(req.CoverImage) == "" {
		return &ValidationError{Field: "coverImage", Msg: "封面图不能为空"}
	}
	// destination: 不能为空
	if strings.TrimSpace(req.Destination) == "" {
		return &ValidationError{Field: "destination", Msg: "目的地不能为空"}
	}
	// summary: 非必填，填写时长度不超过150
	s := strings.TrimSpace(req.Summary)
	if s != "" && len(s) > 150 {
		return &ValidationError{Field: "summary", Msg: "摘要长度不能超过150字符"}
	}
	// budget_min / budget_max: 若填写，必须为正数，且 min ≤ max
	if req.BudgetMin != nil && req.BudgetMax != nil {
		if *req.BudgetMin <= 0 || *req.BudgetMax <= 0 {
			return &ValidationError{Field: "budgetMin", Msg: "预算必须为正数"}
		}
		if *req.BudgetMin > *req.BudgetMax {
			return &ValidationError{Field: "budgetMin", Msg: "预算下限不能大于上限"}
		}
	}
	// recommended_days: 若填写，必须 ≥ 1
	if req.RecommendedDays != nil && *req.RecommendedDays < 1 {
		return &ValidationError{Field: "recommendedDays", Msg: "建议天数至少为1"}
	}
	// status: 只能是 0 或 1
	if req.Status != 0 && req.Status != 1 {
		return &ValidationError{Field: "status", Msg: "状态值无效"}
	}
	// days: 如果有传，校验每项
	for i, d := range req.Days {
		for j, it := range d.Items {
			if !ValidSectionTypes[it.SectionType] {
				return &ValidationError{Field: fmt.Sprintf("days[%d].items[%d].sectionType", i, j),
					Msg: fmt.Sprintf("无效的板块类型: %s", it.SectionType)}
			}
			if strings.TrimSpace(it.Title) == "" {
				return &ValidationError{Field: fmt.Sprintf("days[%d].items[%d].title", i, j), Msg: "行程项标题不能为空"}
			}
			// 交通类型需校验交通方式
			if it.SectionType == "transport" && !ValidTransportModes[it.TransportMode] {
				return &ValidationError{Field: fmt.Sprintf("days[%d].items[%d].transportMode", i, j),
					Msg: fmt.Sprintf("无效的交通方式: %s", it.TransportMode)}
			}
		}
	}
	return nil
}

// ==================== 类型 ====================

// ValidationError 字段级校验错误
type ValidationError struct {
	Field string `json:"field"`
	Msg   string `json:"msg"`
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Msg)
}

// GuideFeedItem 攻略瀑布流返回项
type GuideFeedItem struct {
	ID              string    `json:"id"`
	UserID          string    `json:"userId"`
	Title           string    `json:"title"`
	CoverImage      string    `json:"coverImage"`
	Destination     string    `json:"destination"`
	Summary         string    `json:"summary"`
	BudgetMin       *float64  `json:"budgetMin"`
	BudgetMax       *float64  `json:"budgetMax"`
	BestSeason      string    `json:"bestSeason"`
	RecommendedDays *int      `json:"recommendedDays"`
	Tags            string    `json:"tags"`
	Difficulty      string    `json:"difficulty"`
	CrowdType       string    `json:"crowdType"`
	IsOriginal      int       `json:"isOriginal"`
	ViewCount       int       `json:"viewCount"`
	LikeCount       int       `json:"likeCount"`
	Status          int       `json:"status"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
	AuthorName      string    `json:"authorName"`
	AuthorAvatar    string    `json:"authorAvatar"`
	IsLiked         bool      `json:"isLiked"`
}
